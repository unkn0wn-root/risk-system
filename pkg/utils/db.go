package utils

import (
	"user-risk-system/cmd/risk-engine/models"
	"user-risk-system/pkg/config"
	"user-risk-system/pkg/logger"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// setupDatabase initializes the PostgreSQL database connection with optimal settings.
// configures the connection pool, tests connectivity, and runs auto-migration for risk models.
func SetupDatabase(
	databaseURL string,
	gormConfig *gorm.Config,
	appConfig *config.Config,
	logger *logger.Logger,
) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), gormConfig)
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(appConfig.DatabaseMaxIdleConn)
	sqlDB.SetMaxOpenConns(appConfig.DatabaseMaxConns)
	sqlDB.SetConnMaxLifetime(appConfig.DatabaseConnLiftime)

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
