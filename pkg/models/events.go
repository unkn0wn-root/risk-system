package models

import "time"

const (
	EventUserCreated  = "user.created"
	EventRiskDetected = "risk.detected"
)

type UserCreatedEvent struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Phone     string    `json:"phone"`
	CreatedAt time.Time `json:"created_at"`
}

type RiskDetectedEvent struct {
	UserID     string    `json:"user_id"`
	Email      string    `json:"email"`
	RiskLevel  string    `json:"risk_level"`
	Reason     string    `json:"reason"`
	Flags      []string  `json:"flags"`
	DetectedAt time.Time `json:"detected_at"`
}
