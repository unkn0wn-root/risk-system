package providers

import (
	"fmt"
	"log"

	"github.com/twilio/twilio-go"
	api "github.com/twilio/twilio-go/rest/api/v2010"
)

// TwilioProvider implements the SMSProvider interface using Twilio's API.
// It handles authentication and SMS formatting for the Twilio service.
type TwilioProvider struct {
	client     *twilio.RestClient
	fromNumber string
}

// NewTwilioProvider creates a new Twilio SMS provider with the given credentials.
// Returns nil if credentials are not properly configured, allowing fallback to simulation.
func NewTwilioProvider(accountSid, authToken, fromNumber string) *TwilioProvider {
	if accountSid == "" || authToken == "" {
		log.Printf("Twilio credentials not configured, will fall back to simulation")
		return nil
	}

	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: accountSid,
		Password: authToken,
	})

	return &TwilioProvider{
		client:     client,
		fromNumber: fromNumber,
	}
}

// SendSMS sends an SMS message using the Twilio API.
// It validates the client configuration and handles API errors.
func (p *TwilioProvider) SendSMS(to, message string) error {
	if p.client == nil {
		return fmt.Errorf("Twilio client not configured")
	}

	params := &api.CreateMessageParams{}
	params.SetTo(to)
	params.SetFrom(p.fromNumber)
	params.SetBody(message)

	resp, err := p.client.Api.CreateMessage(params)
	if err != nil {
		return fmt.Errorf("failed to send SMS via Twilio: %w", err)
	}

	log.Printf("[TWILIO] SMS sent successfully to %s (SID: %s)", to, *resp.Sid)
	return nil
}

// GetProviderName returns the name of this SMS provider for logging and identification.
func (p *TwilioProvider) GetProviderName() string {
	return "TWILIO"
}
