// Package models defines event structures for inter-service communication and message publishing.
package models

import "time"

// Event type constants for identifying different types of system events.
const (
	EventUserCreated  = "user.created"  // Fired when a new user account is created
	EventRiskDetected = "risk.detected" // Fired when risk assessment detects potential issues
)

// UserCreatedEvent represents the event data published when a new user is created.
// It contains essential user information for downstream services like notifications and analytics.
type UserCreatedEvent struct {
	UserID    string    `json:"user_id"`    // Unique user identifier
	Email     string    `json:"email"`     // User's email address
	FirstName string    `json:"first_name"` // User's first name
	LastName  string    `json:"last_name"`  // User's last name
	Phone     string    `json:"phone"`     // User's phone number
	CreatedAt time.Time `json:"created_at"` // Timestamp when user was created
}

// RiskDetectedEvent represents the event data published when risk assessment identifies potential issues.
// It includes risk level, reasons, and specific flags for automated and manual review processes.
type RiskDetectedEvent struct {
	UserID     string    `json:"user_id"`     // Unique user identifier associated with the risk
	Email      string    `json:"email"`      // User's email address for notification purposes
	RiskLevel  string    `json:"risk_level"`  // Risk severity level (low, medium, high, critical)
	Reason     string    `json:"reason"`     // Primary reason for risk detection
	Flags      []string  `json:"flags"`      // Specific risk flags that were triggered
	DetectedAt time.Time `json:"detected_at"` // Timestamp when risk was detected
}
