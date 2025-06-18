package providers

import (
	"fmt"
	"log"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// SendGridProvider implements the EmailProvider interface using SendGrid's API.
// It handles authentication and email formatting for the SendGrid service.
type SendGridProvider struct {
	apiKey    string
	fromEmail string
	fromName  string
}

// NewSendGridProvider creates a new SendGrid email provider with the given credentials.
// It requires an API key, sender email address, and sender name.
func NewSendGridProvider(apiKey, fromEmail, fromName string) *SendGridProvider {
	return &SendGridProvider{
		apiKey:    apiKey,
		fromEmail: fromEmail,
		fromName:  fromName,
	}
}

// SendEmail sends an email using the SendGrid API.
// It validates the API key and handles error responses from the SendGrid service.
func (p *SendGridProvider) SendEmail(to, subject, body string, templateData map[string]interface{}) error {
	if p.apiKey == "" {
		return fmt.Errorf("SendGrid API key not configured")
	}

	from := mail.NewEmail(p.fromName, p.fromEmail)
	toEmail := mail.NewEmail("", to)
	message := mail.NewSingleEmail(from, subject, toEmail, body, body)

	client := sendgrid.NewSendClient(p.apiKey)
	response, err := client.Send(message)
	if err != nil {
		return fmt.Errorf("failed to send email via SendGrid: %w", err)
	}

	if response.StatusCode >= 400 {
		return fmt.Errorf("SendGrid API error: %d - %s", response.StatusCode, response.Body)
	}

	log.Printf("[SENDGRID] Email sent successfully to %s", to)
	return nil
}

// GetProviderName returns the name of this email provider for logging and identification.
func (p *SendGridProvider) GetProviderName() string {
	return "SENDGRID"
}
