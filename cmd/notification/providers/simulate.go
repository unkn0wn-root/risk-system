package providers

import (
	"fmt"
	"log"
	"math/rand"
	"time"
)

// SimulateEmailProvider simulates email sending
type SimulateEmailProvider struct{}

func NewSimulateEmailProvider() *SimulateEmailProvider {
	return &SimulateEmailProvider{}
}

func (p *SimulateEmailProvider) SendEmail(to, subject, body string, templateData map[string]interface{}) error {
	log.Printf("📧 [SIMULATE] Sending Email")
	log.Printf("   To: %s", to)
	log.Printf("   Subject: %s", subject)
	log.Printf("   Body: %s", body)
	if len(templateData) > 0 {
		log.Printf("   Template Data: %+v", templateData)
	}

	time.Sleep(time.Duration(100+rand.Intn(200)) * time.Millisecond)

	// (5% failure rate)
	if rand.Intn(100) < 5 {
		return fmt.Errorf("simulated email delivery failure")
	}

	log.Printf("   ✅ Email sent successfully (simulated)")
	return nil
}

func (p *SimulateEmailProvider) GetProviderName() string {
	return "SIMULATE_EMAIL"
}

// SimulateSMSProvider simulates SMS sending
type SimulateSMSProvider struct{}

func NewSimulateSMSProvider() *SimulateSMSProvider {
	return &SimulateSMSProvider{}
}

func (p *SimulateSMSProvider) SendSMS(to, message string) error {
	log.Printf("📱 [SIMULATE] Sending SMS")
	log.Printf("   To: %s", to)
	log.Printf("   Message: %s", message)

	time.Sleep(time.Duration(50+rand.Intn(100)) * time.Millisecond)

	// (3% failure rate)
	if rand.Intn(100) < 3 {
		return fmt.Errorf("simulated SMS delivery failure")
	}

	log.Printf("   ✅ SMS sent successfully (simulated)")
	return nil
}

func (p *SimulateSMSProvider) GetProviderName() string {
	return "SIMULATE_SMS"
}

// SimulatePushProvider simulates push notifications
type SimulatePushProvider struct{}

func NewSimulatePushProvider() *SimulatePushProvider {
	return &SimulatePushProvider{}
}

func (p *SimulatePushProvider) SendPush(userID, title, message string, data map[string]interface{}) error {
	log.Printf("🔔 [SIMULATE] Sending Push Notification")
	log.Printf("   User ID: %s", userID)
	log.Printf("   Title: %s", title)
	log.Printf("   Message: %s", message)
	if len(data) > 0 {
		log.Printf("   Data: %+v", data)
	}

	time.Sleep(time.Duration(30+rand.Intn(70)) * time.Millisecond)

	// (2% failure rate)
	if rand.Intn(100) < 2 {
		return fmt.Errorf("simulated push notification delivery failure")
	}

	log.Printf("   ✅ Push notification sent successfully (simulated)")
	return nil
}

func (p *SimulatePushProvider) GetProviderName() string {
	return "SIMULATE_PUSH"
}
