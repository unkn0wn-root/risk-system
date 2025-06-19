package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

	"user-risk-system/cmd/notification/handlers"
	"user-risk-system/cmd/notification/templates"
	"user-risk-system/pkg/config"
	"user-risk-system/pkg/health"
	"user-risk-system/pkg/logger"
	"user-risk-system/pkg/messaging"
	pb_notification "user-risk-system/proto/notification"
)

// main initializes and starts the notification service with both gRPC and message queue consumers.
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	logConfig := logger.LogConfig{
		Level:       "info",
		Format:      "json",
		ServiceName: cfg.ServiceName,
		Environment: cfg.Environment,
	}
	nl := logger.New(logConfig)

	nl.Info("Starting Notification Service...")
	nl.Info("Email Provider: %s", cfg.EmailProvider)
	nl.Info("SMS Provider: %s", cfg.SMSProvider)
	nl.Info("Push Provider: %s", cfg.PushProvider)

	rabbitMQ, err := messaging.NewRabbitMQ(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitMQ.Close()

	queues := []string{"user.created", "risk.detected", "notifications"}
	for _, queue := range queues {
		if err := rabbitMQ.DeclareQueue(queue); err != nil {
			nl.Fatalf("Failed to declare queue %s: %v", queue, err)
		}
	}

	templ := templates.NewEmailTemplateManager(cfg.TemplatesDirectoryPath)

	// Create notification handler
	notificationHandler := handlers.NewNotificationHandler(rabbitMQ, cfg, templ, nl)

	// Start message consumers for asynchronous processing
	notificationHandler.StartMessageConsumer()

	// Create gRPC server for synchronous processing
	lis, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		nl.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb_notification.RegisterNotificationServiceServer(s, notificationHandler)

	// Health service
	health.RegisterHealthServiceWithDefaults(s, "notification.NotificationService")

	go func() {
		nl.Info("Notification service starting on port %s...", cfg.Port)
		if err := s.Serve(lis); err != nil {
			nl.Fatalf("Failed to serve: %v", err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	nl.Warn("Shutting down notification service...")
	s.GracefulStop()
}
