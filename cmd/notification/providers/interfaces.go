package providers

type EmailProvider interface {
	SendEmail(to, subject, body string, templateData map[string]interface{}) error
	GetProviderName() string
}

type SMSProvider interface {
	SendSMS(to, message string) error
	GetProviderName() string
}

type PushProvider interface {
	SendPush(userID, title, message string, data map[string]interface{}) error
	GetProviderName() string
}
