package config

import (
	"fmt"
	"time"
)

type Config struct {
	// Service configuration
	ServiceName string
	Port        string
	Environment string // dev, staging, prod
	LogLevel    string // debug, info, warn, error

	// Database
	DatabaseURL      string
	RiskDatabaseURL  string
	DatabaseMaxConns int
	DatabaseTimeout  time.Duration

	// JWT
	JWTSecret   string
	JWTDuration time.Duration
	JWTIssuer   string

	// External Services
	UserServiceURL         string
	RiskServiceURL         string
	NotificationServiceURL string
	RabbitMQURL            string

	// Email Configuration
	EmailProvider     string
	SendGridAPIKey    string
	SendGridFromEmail string
	SendGridFromName  string

	// SMS Configuration
	SMSProvider      string
	TwilioAccountSID string
	TwilioAuthToken  string
	TwilioFromNumber string
	PushProvider     string

	// Security
	RateLimitRequests int
	RateLimitWindow   time.Duration

	// Monitoring
	MetricsEnabled bool
	TracingEnabled bool

	TemplatesDirectoryPath string
}

func Load() (*Config, error) {
	config := &Config{
		// Using the generic Env function (Approach 1)
		ServiceName:     Env.String("SERVICE_NAME", "user-risk-system"),
		Port:            Env.String("PORT", "8080"),
		Environment:     Env.String("ENVIRONMENT", "development"),
		LogLevel:        Env.String("LOG_LEVEL", "info"),
		DatabaseTimeout: Env.Duration("DATABASE_TIMEOUT", 30*time.Second),
		JWTDuration:     Env.Duration("JWT_DURATION", 24*time.Hour),
		JWTIssuer:       Env.String("JWT_ISSUER", "user-risk-system"),

		// Required values
		DatabaseURL:     Env.String("DATABASE_URL", ""),
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
		DatabaseMaxConns:  Env.Int("DATABASE_MAX_CONNS", 25),
		RateLimitRequests: Env.Int("RATE_LIMIT_REQUESTS", 100),
		RateLimitWindow:   Env.Duration("RATE_LIMIT_WINDOW", time.Minute),
		MetricsEnabled:    Env.Bool("METRICS_ENABLED", false),
		TracingEnabled:    Env.Bool("TRACING_ENABLED", false),

		TemplatesDirectoryPath: Env.String("TEMPLATES_PATH", ""),
	}

	// Validate required fields
	if err := config.validate(); err != nil {
		return nil, err
	}

	return config, nil
}

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

func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}
