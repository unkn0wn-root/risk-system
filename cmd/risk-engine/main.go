// Package main implements the risk engine service entry point.
// This service evaluates user data against configurable risk rules and provides analytics.
package main

import (
	"log"
	"net"
	"regexp"
	"time"

	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"user-risk-system/cmd/risk-engine/handlers"
	"user-risk-system/cmd/risk-engine/models"
	"user-risk-system/cmd/risk-engine/repository"
	"user-risk-system/cmd/risk-engine/services"
	"user-risk-system/pkg/config"
	"user-risk-system/pkg/logger"
	pb_risk "user-risk-system/pkg/proto/risk"
)

// riskConfig holds the configuration specific to the risk engine service.
// It includes database connection details and server port information.
type riskConfig struct {
	DatabaseURL string
	Port        string
}

// maskPassword obscures password information in database URLs for secure logging.
// It replaces the password parameter value with asterisks to prevent credential exposure.
func maskPassword(databaseURL string) string {
	re := regexp.MustCompile(`password=([^&\s]+)`)
	return re.ReplaceAllString(databaseURL, "password=***")
}

// setupDatabase initializes the PostgreSQL database connection with optimal settings.
// It configures the connection pool, tests connectivity, and runs auto-migration for risk models.
func setupDatabase(databaseURL string, logger *logger.Logger) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Test the connection
	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	// Run auto-migration
	logger.Info("Running database auto-migration...")
	if err := models.AutoMigrate(db); err != nil {
		return nil, err
	}
	logger.Info("Database auto-migration completed successfully")

	return db, nil
}

// main initializes and starts the risk engine service with gRPC endpoints.
// It sets up database connections, repositories, services, and handlers for risk evaluation.
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	rcfg := &riskConfig{
		DatabaseURL: cfg.RiskDatabaseURL,
		Port:        ":" + cfg.Port,
	}

	logConfig := logger.LogConfig{
		Level:       "info",
		Format:      "json",
		ServiceName: cfg.ServiceName,
		Environment: cfg.Environment,
	}

	rl := logger.New(logConfig)

	// setup database
	db, err := setupDatabase(rcfg.DatabaseURL, rl)
	if err != nil {
		rl.Fatalf("Failed to setup database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		rl.Fatalf("Failed to get underlying SQL DB: %v", err)
	}
	defer sqlDB.Close()

	rl.Info("Risk engine configuration",
		"database_url", maskPassword(rcfg.DatabaseURL),
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

	rl.Info("Risk service starting", "port", rcfg.Port)
	if err := s.Serve(lis); err != nil {
		rl.Fatalf("Failed to serve: %v", err)
	}
}
