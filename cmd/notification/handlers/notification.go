package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	notification_models "user-risk-system/cmd/notification/models"
	"user-risk-system/cmd/notification/providers"
	"user-risk-system/cmd/notification/templates"
	"user-risk-system/pkg/config"
	"user-risk-system/pkg/logger"
	"user-risk-system/pkg/messaging"
	"user-risk-system/pkg/models"
	pb_notification "user-risk-system/proto/notification"
)

// NotificationHandler orchestrates notification delivery across multiple channels and providers.
// implements both gRPC services and message queue consumers for flexible notification processing.
type NotificationHandler struct {
	pb_notification.UnimplementedNotificationServiceServer
	messageQueue    *messaging.RabbitMQ
	config          *config.Config
	emailProvider   providers.EmailProvider
	smsProvider     providers.SMSProvider
	pushProvider    providers.PushProvider
	templateManager *templates.EmailTemplateManager
	logger          *logger.Logger
}

// NewNotificationHandler creates a new notification handler with the provided dependencies.
// initializes all notification providers based on configuration settings.
func NewNotificationHandler(
	messageQueue *messaging.RabbitMQ,
	cfg *config.Config,
	templateManager *templates.EmailTemplateManager,
	appLogger *logger.Logger,
) *NotificationHandler {
	handler := &NotificationHandler{
		messageQueue:    messageQueue,
		config:          cfg,
		templateManager: templateManager,
		logger:          appLogger,
	}

	handler.initializeProviders()
	return handler
}

// initializeProviders configures email, SMS, and push notification providers based on config.
// falls back to simulation providers when real providers are not properly configured.
func (h *NotificationHandler) initializeProviders() {
	// Email Provider
	switch h.config.EmailProvider {
	case "SENDGRID":
		if h.config.SendGridAPIKey != "" {
			h.emailProvider = providers.NewSendGridProvider(
				h.config.SendGridAPIKey,
				h.config.SendGridFromEmail,
				h.config.SendGridFromName,
			)
			h.logger.Info("Email provider initialized: SendGrid")
		} else {
			h.logger.Warn("SendGrid API key not configured, falling back to simulation")
			h.emailProvider = providers.NewSimulateEmailProvider()
		}
	default:
		h.emailProvider = providers.NewSimulateEmailProvider()
		h.logger.Info("Email provider initialized: Simulate")
	}

	// SMS Provider
	switch h.config.SMSProvider {
	case "TWILIO":
		if h.config.TwilioAccountSID != "" && h.config.TwilioAuthToken != "" {
			twilioProvider := providers.NewTwilioProvider(
				h.config.TwilioAccountSID,
				h.config.TwilioAuthToken,
				h.config.TwilioFromNumber,
			)
			if twilioProvider != nil {
				h.smsProvider = twilioProvider
				h.logger.Info("SMS provider initialized: Twilio")
			} else {
				h.smsProvider = providers.NewSimulateSMSProvider()
				h.logger.Warn("Twilio not configured properly, using simulation")
			}
		} else {
			h.smsProvider = providers.NewSimulateSMSProvider()
			h.logger.Warn("Twilio credentials not configured, using simulation")
		}
	default:
		h.smsProvider = providers.NewSimulateSMSProvider()
		h.logger.Info("SMS provider initialized: Simulate")
	}

	// Push Provider (always simulate for now)
	h.pushProvider = providers.NewSimulatePushProvider()
	h.logger.Info("Push provider: Simulate")
}

// SendNotification handles synchronous gRPC notification requests from other services.
// determines appropriate channels based on notification type and sends via all relevant providers.
func (h *NotificationHandler) SendNotification(ctx context.Context, req *pb_notification.SendNotificationRequest) (*pb_notification.SendNotificationResponse, error) {
	h.logger.InfoCtx(ctx, "Sending notification",
		"type", req.Type,
		"user_id", req.UserId,
		"email", req.Email,
	)

	notification := &notification_models.Notification{
		ID:        uuid.New().String(),
		UserID:    req.UserId,
		Type:      req.Type,
		Message:   req.Message,
		Email:     req.Email,
		Channel:   notification_models.ChannelEmail, // Default to email
		Status:    notification_models.NotificationStatusPending,
		CreatedAt: time.Now(),
	}

	channels := h.determineChannels(req.Type)
	success := true
	var errorMsg string

	for _, channel := range channels {
		notification.Channel = channel
		if err := h.sendNotificationByChannel(ctx, notification); err != nil {
			h.logger.ErrorCtx(ctx, "Failed to send notification", err,
				"channel", channel,
				"notification_id", notification.ID,
			)
			success = false
			errorMsg = err.Error()
		}
	}

	if success {
		now := time.Now()
		notification.Status = notification_models.NotificationStatusSent
		notification.SentAt = &now
		h.logger.InfoCtx(ctx, "Notification sent successfully",
			"notification_id", notification.ID,
			"channels", channels,
		)
	} else {
		notification.Status = notification_models.NotificationStatusFailed
		notification.Error = errorMsg
	}

	return &pb_notification.SendNotificationResponse{
		Success: success,
		Error:   errorMsg,
	}, nil
}

// determineChannels selects appropriate notification channels based on notification type.
// Critical notifications like risk alerts use multiple channels for redundancy.
func (h *NotificationHandler) determineChannels(notificationType string) []string {
	switch notificationType {
	case notification_models.NotificationTypeUserCreated:
		return []string{notification_models.ChannelEmail}
	case notification_models.NotificationTypeRiskDetected:
		// High priority - send via multiple channels
		return []string{
			notification_models.ChannelEmail,
			notification_models.ChannelSMS,
			notification_models.ChannelPush,
		}
	case notification_models.NotificationTypePasswordReset:
		return []string{notification_models.ChannelEmail, notification_models.ChannelSMS}
	case notification_models.NotificationTypeLoginAlert:
		return []string{notification_models.ChannelEmail, notification_models.ChannelPush}
	default:
		return []string{notification_models.ChannelEmail}
	}
}

// sendNotificationByChannel routes notifications to the appropriate provider based on channel type.
// acts as a dispatcher between channel types and their respective implementations.
func (h *NotificationHandler) sendNotificationByChannel(ctx context.Context, notification *notification_models.Notification) error {
	switch notification.Channel {
	case notification_models.ChannelEmail:
		return h.sendEmailNotification(ctx, notification)
	case notification_models.ChannelSMS:
		return h.sendSMSNotification(ctx, notification)
	case notification_models.ChannelPush:
		return h.sendPushNotification(ctx, notification)
	default:
		return fmt.Errorf("unsupported notification channel: %s", notification.Channel)
	}
}

// sendEmailNotification handles email delivery using configured email providers.
// selects appropriate templates and renders them with notification data.
func (h *NotificationHandler) sendEmailNotification(ctx context.Context, notification *notification_models.Notification) error {
	templateData := templates.EmailTemplateData{
		UserID:    notification.UserID,
		Email:     notification.Email,
		FirstName: extractFirstName(notification.Message),
	}

	var templateName string
	switch notification.Type {
	case notification_models.NotificationTypeUserCreated:
		templateName = "welcome"
	case notification_models.NotificationTypeRiskDetected:
		templateName = "risk_alert"
		templateData.Reason = notification.Message
		templateData.RiskLevel = "HIGH" // Should be extracted from message
	default:
		templateName = "welcome"
	}

	subject, htmlBody, err := h.templateManager.RenderTemplate(templateName, templateData)
	if err != nil {
		return err
	}

	notification.Provider = h.emailProvider.GetProviderName()

	err = h.emailProvider.SendEmail(notification.Email, subject, htmlBody, map[string]interface{}{
		"template": templateName,
		"user_id":  notification.UserID,
	})

	if err != nil {
		h.logger.ErrorCtx(ctx, "Email sending failed", err,
			"provider", notification.Provider,
			"template", templateName,
		)
		return err
	}

	h.logger.InfoCtx(ctx, "Email sent successfully",
		"provider", notification.Provider,
		"template", templateName,
		"subject", subject,
	)

	return nil
}

// sendSMSNotification handles SMS delivery using configured SMS providers.
// formats messages appropriately for SMS length constraints.
func (h *NotificationHandler) sendSMSNotification(ctx context.Context, notification *notification_models.Notification) error {
	message := h.getSMSMessage(notification.Type, notification.Message)

	notification.Provider = h.smsProvider.GetProviderName()
	return h.smsProvider.SendSMS(notification.Phone, message)
}

// sendPushNotification handles push notification delivery using configured push providers.
// formats titles and messages with additional metadata for mobile apps.
func (h *NotificationHandler) sendPushNotification(ctx context.Context, notification *notification_models.Notification) error {
	title := h.getPushTitle(notification.Type)
	message := notification.Message

	data := map[string]interface{}{
		"type":    notification.Type,
		"user_id": notification.UserID,
	}

	notification.Provider = h.pushProvider.GetProviderName()
	return h.pushProvider.SendPush(notification.UserID, title, message, data)
}

// getEmailSubject generates email subject lines based on notification type.
// includes emojis and urgency indicators for better user experience.
func (h *NotificationHandler) getEmailSubject(notificationType string) string {
	switch notificationType {
	case notification_models.NotificationTypeUserCreated:
		return "Welcome to User Risk Management System!"
	case notification_models.NotificationTypeRiskDetected:
		return "🚨 Security Alert - Risk Detected"
	case notification_models.NotificationTypePasswordReset:
		return "Password Reset Request"
	case notification_models.NotificationTypeLoginAlert:
		return "🔐 New Login Alert"
	default:
		return "Notification from User Risk System"
	}
}

// getSMSMessage formats messages for SMS delivery with length constraints.
// truncates long messages and adds context-appropriate prefixes.
func (h *NotificationHandler) getSMSMessage(notificationType, message string) string {
	switch notificationType {
	case notification_models.NotificationTypeRiskDetected:
		return fmt.Sprintf("🚨 SECURITY ALERT: %s Please check your email for details.", message)
	case notification_models.NotificationTypePasswordReset:
		return fmt.Sprintf("Password reset requested. %s", message)
	default:
		// Truncate long messages for SMS
		if len(message) > 140 {
			return message[:137] + "..."
		}
		return message
	}
}

// getPushTitle generates concise titles for push notifications based on type.
// Titles are optimized for mobile notification display constraints.
func (h *NotificationHandler) getPushTitle(notificationType string) string {
	switch notificationType {
	case notification_models.NotificationTypeRiskDetected:
		return "Security Alert"
	case notification_models.NotificationTypeLoginAlert:
		return "New Login"
	default:
		return "Notification"
	}
}

// StartMessageConsumer initializes all message queue consumers for asynchronous processing.
func (h *NotificationHandler) StartMessageConsumer() {
	go func() {
		h.logger.Info("Starting user.created queue consumer...")
		err := h.messageQueue.Consume("user.created", h.handleUserCreatedEvent)
		if err != nil {
			h.logger.Error("Error consuming user.created queue", err)
		}
	}()

	// Consume risk detected events
	go func() {
		h.logger.Info("Starting risk.detected queue consumer...")
		err := h.messageQueue.Consume("risk.detected", h.handleRiskDetectedEvent)
		if err != nil {
			h.logger.Error("Error consuming risk.detected queue", err)
		}
	}()

	// Consume direct notification requests
	go func() {
		h.logger.Info("Starting notifications queue consumer...")
		err := h.messageQueue.Consume("notifications", h.handleNotificationEvent)
		if err != nil {
			h.logger.Error("Error consuming notifications queue", err)
		}
	}()
}

// handleUserCreatedEvent processes user registration events from the message queue.
func (h *NotificationHandler) handleUserCreatedEvent(data []byte) error {
	var event models.UserCreatedEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return fmt.Errorf("failed to unmarshal user created event: %w", err)
	}

	h.logger.Info("Processing user created event",
		"user_id", event.UserID,
		"email", event.Email,
	)

	notification := &notification_models.Notification{
		ID:        uuid.New().String(),
		UserID:    event.UserID,
		Type:      notification_models.NotificationTypeUserCreated,
		Message:   fmt.Sprintf("Welcome %s %s! Your account has been created successfully.", event.FirstName, event.LastName),
		Email:     event.Email,
		Status:    notification_models.NotificationStatusPending,
		CreatedAt: time.Now(),
	}

	ctx := context.WithValue(context.Background(), "user_id", event.UserID)
	ctx = context.WithValue(ctx, "user_email", event.Email)

	if err := h.sendEmailNotification(ctx, notification); err != nil {
		h.logger.ErrorCtx(ctx, "Failed to send welcome email", err,
			"notification_id", notification.ID,
		)
		notification.Status = notification_models.NotificationStatusFailed
		notification.Error = err.Error()
		return err
	}

	now := time.Now()
	notification.Status = notification_models.NotificationStatusSent
	notification.SentAt = &now
	h.logger.InfoCtx(ctx, "Welcome email sent successfully",
		"notification_id", notification.ID,
	)

	return nil
}

// handleRiskDetectedEvent processes risk detection events from the message queue.
// sends urgent security alerts via multiple channels based on risk level.
func (h *NotificationHandler) handleRiskDetectedEvent(data []byte) error {
	var event models.RiskDetectedEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return fmt.Errorf("failed to unmarshal risk detected event: %w", err)
	}

	h.logger.Info("Processing risk detected event",
		"user_id", event.UserID,
		"risk_level", event.RiskLevel,
	)

	notification := &notification_models.Notification{
		ID:        uuid.New().String(),
		UserID:    event.UserID,
		Type:      notification_models.NotificationTypeRiskDetected,
		Message:   fmt.Sprintf("Risk Alert: %s (Level: %s, Flags: %s)", event.Reason, event.RiskLevel, strings.Join(event.Flags, ", ")),
		Email:     event.Email,
		Status:    notification_models.NotificationStatusPending,
		CreatedAt: time.Now(),
	}

	ctx := context.WithValue(context.Background(), "user_id", event.UserID)
	ctx = context.WithValue(ctx, "user_email", event.Email)

	channels := h.determineChannels(notification.Type)
	success := true

	for _, channel := range channels {
		notification.Channel = channel
		if err := h.sendNotificationByChannel(ctx, notification); err != nil {
			h.logger.ErrorCtx(ctx, "Failed to send notification", err,
				"channel", channel,
				"notification_id", notification.ID,
			)
			success = false
		}
	}

	if success {
		now := time.Now()
		notification.Status = notification_models.NotificationStatusSent
		notification.SentAt = &now
		h.logger.InfoCtx(ctx, "Risk alert notifications sent successfully",
			"notification_id", notification.ID,
			"channels", channels,
		)
	} else {
		notification.Status = notification_models.NotificationStatusFailed
	}

	return nil
}

// handleNotificationEvent processes direct notification requests from the message queue.
// handles generic notifications that don't fit into specific event categories.
func (h *NotificationHandler) handleNotificationEvent(data []byte) error {
	var notification notification_models.Notification
	if err := json.Unmarshal(data, &notification); err != nil {
		return fmt.Errorf("failed to unmarshal notification event: %w", err)
	}

	h.logger.Info("Processing direct notification",
		"notification_id", notification.ID,
		"type", notification.Type,
	)

	ctx := context.WithValue(context.Background(), "user_id", notification.UserID)
	if notification.Email != "" {
		ctx = context.WithValue(ctx, "user_email", notification.Email)
	}

	if err := h.sendNotificationByChannel(ctx, &notification); err != nil {
		h.logger.ErrorCtx(ctx, "Failed to send notification", err,
			"notification_id", notification.ID,
		)
		notification.Status = notification_models.NotificationStatusFailed
		notification.Error = err.Error()
		return err
	}

	now := time.Now()
	notification.Status = notification_models.NotificationStatusSent
	notification.SentAt = &now
	h.logger.InfoCtx(ctx, "Direct notification sent successfully",
		"notification_id", notification.ID,
	)

	return nil
}

// extractFirstName parses user names from welcome messages for personalization.
func extractFirstName(message string) string {
	if strings.Contains(message, "Welcome") {
		parts := strings.Split(message, " ")
		if len(parts) > 1 {
			return parts[1]
		}
	}
	return "User"
}
