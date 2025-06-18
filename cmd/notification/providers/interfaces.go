// Package providers defines interfaces for different notification delivery methods.
// This allows for easy switching between different email, SMS, and push providers.
package providers

// EmailProvider defines the interface for sending email notifications.
// Implementations can use different email services like SendGrid, AWS SES, etc.
type EmailProvider interface {
	SendEmail(to, subject, body string, templateData map[string]interface{}) error
	GetProviderName() string
}

// SMSProvider defines the interface for sending SMS notifications.
// Implementations can use different SMS services like Twilio, AWS SNS, etc.
type SMSProvider interface {
	SendSMS(to, message string) error
	GetProviderName() string
}

// PushProvider defines the interface for sending push notifications.
// Implementations can use different push services like Firebase, APNs, etc.
type PushProvider interface {
	SendPush(userID, title, message string, data map[string]interface{}) error
	GetProviderName() string
}
