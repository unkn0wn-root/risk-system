package main

import (
	"log"
	"net"

	"google.golang.org/grpc"
	"gorm.io/gorm"

	"user-risk-system/cmd/user/handlers"
	"user-risk-system/cmd/user/models"
	"user-risk-system/cmd/user/repository"
	"user-risk-system/pkg/auth"
	"user-risk-system/pkg/config"
	"user-risk-system/pkg/health"
	"user-risk-system/pkg/logger"
	"user-risk-system/pkg/messaging"
	"user-risk-system/pkg/utils"
	pb_notification "user-risk-system/proto/notification"
	pb_risk "user-risk-system/proto/risk"
	pb_user "user-risk-system/proto/user"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	logConfig := logger.LogConfig{
		Level:       "info",
		Format:      "json",
		ServiceName: cfg.ServiceName,
		Environment: cfg.Environment,
	}
	appLogger := logger.New(logConfig)

	// Database
	db, err := utils.SetupDatabase(cfg.DatabaseURL, &gorm.Config{}, cfg, appLogger)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run auto-migration for user models
	appLogger.Info("Running user database migration...")
	if err := models.AutoMigrate(db); err != nil {
		appLogger.Fatalf("Failed to run migration: %v", err)
	}
	appLogger.Info("User database migration completed successfully")

	sdb, err := db.DB()
	if err != nil {
		appLogger.Fatalf("Failed to get underlying SQL DB: %v", err)
	}
	defer sdb.Close()

	// gRPC client connections
	riskConn, err := grpc.Dial(cfg.RiskServiceURL, grpc.WithInsecure())
	if err != nil {
		appLogger.Fatalf("Failed to connect to risk service: %v", err)
	}
	defer riskConn.Close()

	notificationConn, err := grpc.Dial(cfg.NotificationServiceURL, grpc.WithInsecure())
	if err != nil {
		appLogger.Fatalf("Failed to connect to notification service: %v", err)
	}
	defer notificationConn.Close()

	// RabbitMQ connection
	rabbitMQ, err := messaging.NewRabbitMQ(cfg.RabbitMQURL)
	if err != nil {
		appLogger.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitMQ.Close()

	// Declare queues
	queues := []string{"user.created", "risk.detected", "notifications"}
	for _, queue := range queues {
		if err := rabbitMQ.DeclareQueue(queue); err != nil {
			appLogger.Fatalf("Failed to declare queue %s: %v", queue, err)
		}
	}

	// Create clients
	riskClient := pb_risk.NewRiskServiceClient(riskConn)
	notificationClient := pb_notification.NewNotificationServiceClient(notificationConn)

	// Create repository and handler
	userRepo := repository.NewUserRepository(db)
	userHandler := handlers.NewUserHandler(
		userRepo,
		riskClient,
		notificationClient,
		rabbitMQ,
		appLogger,
	)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		appLogger.Fatalf("Failed to listen: %v", err)
	}

	// JWT is enabled by default
	// if you want to explicitly disable it, you have to set REQUIRE_SERVICE_JWT_FORWARDING to false
	var s *grpc.Server
	if cfg.RequireServiceJWTForwarding {
		jwtManager := auth.NewJWTManager(cfg.JWTSecret, cfg.JWTDuration, cfg.JWTIssuer)
		authMiddleware := auth.NewAuthMiddleware(jwtManager)
		s = grpc.NewServer(
			grpc.UnaryInterceptor(authMiddleware.GRPCUnaryInterceptor),
		)
		appLogger.Info("gRPC JWT authentication enabled")
	} else {
		s = grpc.NewServer()
		appLogger.Warn("gRPC JWT authentication disabled")
	}

	pb_user.RegisterUserServiceServer(s, userHandler)

	health.RegisterHealthServiceWithDefaults(s, "user.UserService")

	appLogger.Info("User service starting on port 50051...")
	if err := s.Serve(lis); err != nil {
		appLogger.Fatalf("Failed to serve: %v", err)
	}
}
