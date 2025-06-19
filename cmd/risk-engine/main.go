package main

import (
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"gorm.io/gorm"

	"user-risk-system/cmd/risk-engine/handlers"
	"user-risk-system/cmd/risk-engine/repository"
	"user-risk-system/cmd/risk-engine/services"
	"user-risk-system/pkg/config"
	"user-risk-system/pkg/health"
	"user-risk-system/pkg/logger"
	"user-risk-system/pkg/utils"
	pb_risk "user-risk-system/proto/risk"
)

// riskConfig holds the configuration specific to the risk engine service.
type riskConfig struct {
	DatabaseURL string
	Port        string
}

// main initializes and starts the risk engine service with gRPC endpoints.
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	rcfg := &riskConfig{
		DatabaseURL: cfg.RiskDatabaseURL,
		Port:        ":" + cfg.Port,
	}

	// log
	logConfig := logger.LogConfig{
		Level:       "info",
		Format:      "json",
		ServiceName: cfg.ServiceName,
		Environment: cfg.Environment,
	}

	rl := logger.New(logConfig)

	// databse
	db, err := utils.SetupDatabase(rcfg.DatabaseURL, &gorm.Config{}, cfg, rl)
	if err != nil {
		rl.Fatalf("Failed to setup database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		rl.Fatalf("Failed to get underlying SQL DB: %v", err)
	}
	defer sqlDB.Close()

	rl.Info("Risk engine configuration",
		"database_url", utils.MaskPassword(rcfg.DatabaseURL),
		"port", rcfg.Port)

	// Initialize repositories
	riskRepo := repository.NewRiskRepository(db)

	// Initialize services
	riskEngine := services.NewRiskEngine(riskRepo, rl)
	riskAnalytics := services.NewRiskAnalytics(db, rl)

	// Initialize handlers
	riskHandler := handlers.NewRiskHandler(riskEngine, riskAnalytics, rl)
	riskAdminHandler := handlers.NewRiskAdminHandler(riskRepo, rl, riskEngine)

	// Create gRPC server
	lis, err := net.Listen("tcp", rcfg.Port)
	if err != nil {
		rl.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()

	// Register services
	pb_risk.RegisterRiskServiceServer(s, riskHandler)
	pb_risk.RegisterRiskAdminServiceServer(s, riskAdminHandler)

	// Health service
	healthConfig := health.Config{
		OverallStatus: grpc_health_v1.HealthCheckResponse_SERVING,
		Services: []health.ServiceHealth{
			{Name: "risk.RiskService", Status: grpc_health_v1.HealthCheckResponse_SERVING},
			{Name: "risk.RiskAdminService", Status: grpc_health_v1.HealthCheckResponse_SERVING},
		},
	}
	health.RegisterHealthService(s, healthConfig)

	rl.Info("Risk service starting", "port", rcfg.Port)
	if err := s.Serve(lis); err != nil {
		rl.Fatalf("Failed to serve: %v", err)
	}
}
