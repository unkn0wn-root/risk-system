package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"user-risk-system/api-gateway/handlers"
	"user-risk-system/api-gateway/middleware"
	"user-risk-system/pkg/auth"
	"user-risk-system/pkg/config"
	"user-risk-system/pkg/logger"
	pb_risk "user-risk-system/pkg/proto/risk"
	pb_user "user-risk-system/pkg/proto/user"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	logConfig := logger.LogConfig{
		Level:       cfg.LogLevel,
		Format:      "json",
		ServiceName: "api-gateway",
		Environment: cfg.Environment,
	}

	appLogger := logger.New(logConfig)
	appLogger.Info("Starting API Gateway",
		"port", cfg.Port,
		"environment", cfg.Environment,
		"jwt_issuer", cfg.JWTIssuer,
	)

	jwtManager := auth.NewJWTManager(cfg.JWTSecret, cfg.JWTDuration, cfg.JWTIssuer)
	authMiddleware := auth.NewAuthMiddleware(jwtManager)

	// gRPC connection with interceptor to user service
	userConn, err := auth.NewAuthenticatedGRPCConnection(cfg.UserServiceURL)
	if err != nil {
		appLogger.Fatalf(
			"Failed to connect to user service", err,
			"service_url", cfg.UserServiceURL,
		)
	}
	defer userConn.Close()

	// gRPC connection to risk service
	riskConn, err := auth.NewAuthenticatedGRPCConnection(cfg.RiskServiceURL)
	if err != nil {
		appLogger.Fatalf(
			"Failed to connect to risk service", err,
			"service_url", cfg.RiskServiceURL,
		)
	}
	defer riskConn.Close()

	userClient := pb_user.NewUserServiceClient(userConn)
	riskClient := pb_risk.NewRiskServiceClient(riskConn)
	riskAdminClient := pb_risk.NewRiskAdminServiceClient(riskConn)

	userHandler := handlers.NewUserHandler(userClient)
	riskHandler := handlers.NewRiskHandler(riskClient, riskAdminClient)
	authHandler := handlers.NewAuthHandler(userClient, jwtManager)
	swaggerHandler := handlers.NewSwaggerHandler()

	r := chi.NewRouter()

	r.Use(middleware.LoggingMiddleware)
	r.Use(middleware.CORSMiddleware)

	// API Documentation routes
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/api/docs", http.StatusMovedPermanently)
	})
	r.Get("/api/docs", swaggerHandler.GetSwaggerUI)
	r.Get("/api/docs/openapi.json", swaggerHandler.GetOpenAPISpec)

	r.Route("/api/v1", func(r chi.Router) {
		// Public routes (no authentication required)
		r.Get("/health", userHandler.HealthCheck)

		// Authentication routes (public)
		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", authHandler.Login)
			r.Post("/register", authHandler.Register)
			r.Post("/refresh", authHandler.RefreshToken)
		})

		// Protected routes group (authentication required)
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.HTTPMiddleware)

			// User profile routes
			r.Get("/profile", authHandler.GetProfile)

			// User management routes
			r.Route("/users", func(r chi.Router) {
				// Admin only routes
				r.With(authMiddleware.RequireRole(auth.RoleAdmin)).Post("/", userHandler.CreateUser)
				r.With(authMiddleware.RequireRole(auth.RoleAdmin)).Get("/", userHandler.ListUsers)

				// User can access their own data, admin can access any
				r.Get("/{id}", userHandler.GetUser)
				r.Put("/{id}", userHandler.UpdateUser)
			})

			// Risk management routes
			r.Route("/risk", func(r chi.Router) {
				// Risk checking - authenticated users can check risk
				r.Post("/check", riskHandler.CheckRisk)

				// Admin only risk rule management
				r.With(authMiddleware.RequireRole(auth.RoleAdmin)).Post("/rules", riskHandler.CreateRiskRule)
				r.With(authMiddleware.RequireRole(auth.RoleAdmin)).Get("/rules", riskHandler.ListRiskRules)
				r.With(authMiddleware.RequireRole(auth.RoleAdmin)).Put("/rules/{id}", riskHandler.UpdateRiskRule)
				r.With(authMiddleware.RequireRole(auth.RoleAdmin)).Delete("/rules/{id}", riskHandler.DeleteRiskRule)
			})
		})
	})

	port := cfg.Port
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		appLogger.Info("API Gateway listening",
			"port", port,
			"endpoints", []string{
				"GET /health",
				"POST /api/v1/auth/login",
				"POST /api/v1/auth/register",
				"GET /api/v1/profile",
				"GET /api/v1/users",
				"POST /api/v1/risk/check",
				"POST /api/v1/risk/rules",
			},
		)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Error("Server failed to start", err)
			os.Exit(1)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	appLogger.Info("Shutting down API Gateway...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		appLogger.Fatalf("Server forced to shutdown: %v", err)
	} else {
		appLogger.Info("API Gateway shutdown complete")
	}
}
