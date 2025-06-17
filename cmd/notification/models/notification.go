package models

import "time"

type Notification struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	Type      string     `json:"type"`
	Message   string     `json:"message"`
	Email     string     `json:"email"`
	Phone     string     `json:"phone,omitempty"`
	Channel   string     `json:"channel"`            // EMAIL, SMS, PUSH, ALL
	Status    string     `json:"status"`             // PENDING, SENT, FAILED
	Provider  string     `json:"provider,omitempty"` // SIMULATE, SENDGRID, TWILIO, etc.
	SentAt    *time.Time `json:"sent_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	Error     string     `json:"error,omitempty"`
}

const (
	NotificationTypeUserCreated   = "USER_CREATED"
	NotificationTypeRiskDetected  = "RISK_DETECTED"
	NotificationTypePasswordReset = "PASSWORD_RESET"
	NotificationTypeLoginAlert    = "LOGIN_ALERT"

	NotificationStatusPending = "PENDING"
	NotificationStatusSent    = "SENT"
	NotificationStatusFailed  = "FAILED"

	// Notification channels
	ChannelEmail = "EMAIL"
	ChannelSMS   = "SMS"
	ChannelPush  = "PUSH"
	ChannelAll   = "ALL"

	// Providers
	ProviderSimulate = "SIMULATE"
	ProviderSendGrid = "SENDGRID"
	ProviderTwilio   = "TWILIO"
	ProviderFirebase = "FIREBASE"
)
