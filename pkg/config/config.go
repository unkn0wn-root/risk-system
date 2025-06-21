// Package config provides application configuration management with environment variable support.
package config

import (
	"fmt"
	"strings"
	"time"
)

// Config holds all application configuration settings loaded from environment variables.
// It includes service settings, database connections, external service URLs, and security parameters.
type Config struct {
	// Service configuration
	ServiceName    string   // Name of the service
	Port           string   // HTTP server port
	Environment    string   // Runtime environment (dev, staging, prod)
	LogLevel       string   // Logging level (debug, info, warn, error)
	AllowedOrigins []string // Allowed cors origins

	// Database
	DatabaseURL         string        // Primary database connection string
	RiskDatabaseURL     string        // Risk assessment database connection string
	DatabaseMaxConns    int           // Maximum database connections in pool
	DatabaseMaxIdleConn int           // Maximum idle connection
	DatabaseConnLiftime time.Duration // Database operation timeout

	// JWT
	JWTSecret   string        // Secret key for JWT token signing
	JWTDuration time.Duration // JWT token validity duration
	JWTIssuer   string        // JWT token issuer identifier

	// External Services
	UserServiceURL         string // User service gRPC endpoint
	RiskServiceURL         string // Risk assessment service gRPC endpoint
	NotificationServiceURL string // Notification service gRPC endpoint
	RabbitMQURL            string // RabbitMQ message broker connection string

	// Email Configuration
	EmailProvider     string // Email service provider (SENDGRID, SIMULATE)
	SendGridAPIKey    string // SendGrid API key for email delivery
	SendGridFromEmail string // Default sender email address
	SendGridFromName  string // Default sender name

	// SMS Configuration
	SMSProvider      string // SMS service provider (TWILIO, SIMULATE)
	TwilioAccountSID string // Twilio account SID for SMS
	TwilioAuthToken  string // Twilio authentication token
	TwilioFromNumber string // Twilio sender phone number
	PushProvider     string // Push notification provider

	// Security
	RateLimitRequests int           // Maximum requests per rate limit window
	RateLimitWindow   time.Duration // Rate limiting time window

	// Monitoring
	MetricsEnabled bool // Enable application metrics collection
	TracingEnabled bool // Enable distributed tracing

	// Service Communication
	RequireServiceJWTForwarding bool // Whether to enforce JWT authentication on service-to-service gRPC calls

	TemplatesDirectoryPath string // Path to notification templates directory
}

// Load creates and validates a new Config instance from environment variables.
// It applies default values where appropriate and validates required fields.
func Load() (*Config, error) {
	config := &Config{
		ServiceName: Env.String("SERVICE_NAME", "user-risk-system"),
		Port:        Env.String("PORT", "8080"),
		Environment: Env.String("ENVIRONMENT", "development"),
		LogLevel:    Env.String("LOG_LEVEL", "info"),
		JWTDuration: Env.Duration("JWT_DURATION", 24*time.Hour),
		JWTIssuer:   Env.String("JWT_ISSUER", "user-risk-system"),

		RiskDatabaseURL: Env.String("RISK_DATABASE_URL", "postgres://user:password@localhost/risk_db?sslmode=disable"),
		JWTSecret:       Env.String("JWT_SECRET", ""),
		RabbitMQURL:     Env.String("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),

		// Service URLs
		UserServiceURL:         Env.String("USER_SERVICE_URL", "localhost:50051"),
		RiskServiceURL:         Env.String("RISK_SERVICE_URL", "localhost:50052"),
		NotificationServiceURL: Env.String("NOTIFICATION_SERVICE_URL", "localhost:50053"),

		// External providers
		EmailProvider:     Env.String("EMAIL_PROVIDER", "SIMULATE"),
		SMSProvider:       Env.String("SMS_PROVIDER", "SIMULATE"),
		SendGridAPIKey:    Env.String("SENDGRID_API_KEY", ""),
		SendGridFromEmail: Env.String("SENDGRID_FROM_EMAIL", "noreply@example.com"),
		SendGridFromName:  Env.String("SENDGRID_FROM_NAME", "User Risk System"),
		TwilioAccountSID:  Env.String("TWILIO_ACCOUNT_SID", ""),
		TwilioAuthToken:   Env.String("TWILIO_AUTH_TOKEN", ""),
		TwilioFromNumber:  Env.String("TWILIO_FROM_NUMBER", ""),
		PushProvider:      Env.String("PUSH_PROVIDER", "SIMULATE"),

		// Security & Performance
		RateLimitRequests: Env.Int("RATE_LIMIT_REQUESTS", 100),
		RateLimitWindow:   Env.Duration("RATE_LIMIT_WINDOW", time.Minute),
		MetricsEnabled:    Env.Bool("METRICS_ENABLED", false),
		TracingEnabled:    Env.Bool("TRACING_ENABLED", false),

		// Service Communication - default to true unless explicitly disabled
		RequireServiceJWTForwarding: Env.Bool("REQUIRE_SERVICE_JWT_FORWARDING", true),

		// Database
		DatabaseURL:         Env.String("DATABASE_URL", ""),
		DatabaseConnLiftime: Env.Duration("DATABASE_CONN_LIFETIME", time.Hour),
		DatabaseMaxIdleConn: Env.Int("DB_MAX_IDLE", 10),
		DatabaseMaxConns:    Env.Int("DATABASE_MAX_CONNS", 25),

		// Common
		TemplatesDirectoryPath: Env.String("TEMPLATES_PATH", ""),
		AllowedOrigins:         strings.Split(Env.String("ALLOWED_CORS", "*"), ","),
	}

	// Validate required fields
	if err := config.validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// validate checks that required configuration values are present.
// It ensures security-critical settings like JWT secrets meet minimum requirements.
func (c *Config) validate() error {
	if c.Environment == "production" {
		if c.JWTSecret == "" {
			return fmt.Errorf("JWT_SECRET is required in production")
		}
		if len(c.JWTSecret) < 32 {
			return fmt.Errorf("JWT_SECRET must be at least 32 characters in production")
		}
		if c.DatabaseURL == "" {
			return fmt.Errorf("DATABASE_URL is required")
		}
	}
	return nil
}

// IsProduction returns true if the application is running in production.
func (c *Config) IsProduction() bool {
	return strings.ToLower(c.Environment) == "production"
}

// IsDevelopment returns true if the application is running in development.
func (c *Config) IsDevelopment() bool {
	return strings.ToLower(c.Environment) == "development"
}
