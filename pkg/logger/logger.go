// Package logger provides structured logging capabilities with context-aware field extraction.
package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"
)

// Logger wraps the standard slog.Logger with additional context-aware logging methods.
type Logger struct {
	*slog.Logger
}

// LogConfig defines the configuration options for creating a new logger instance.
type LogConfig struct {
	Level       string // Logging level (debug, info, warn, error)
	Format      string // Output format (json, text)
	ServiceName string // Service name to include in log entries
	Environment string // Environment name to include in log entries
}

// New creates a new Logger instance with the specified configuration.
func New(config LogConfig) *Logger {
	var level slog.Level
	switch config.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Add custom formatting here
			if a.Key == slog.TimeKey {
				a.Value = slog.StringValue(time.Now().Format(time.RFC3339))
			}
			return a
		},
	}

	var handler slog.Handler
	if config.Format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	logger := slog.New(handler).With(
		"service", config.ServiceName,
		"environment", config.Environment,
	)

	return &Logger{Logger: logger}
}

// Info logs an informational message with optional key-value pairs.
func (l *Logger) Info(msg string, args ...any) {
	l.Logger.Info(msg, args...)
}

// InfoCtx logs an informational message with context-extracted fields and optional key-value pairs.
func (l *Logger) InfoCtx(ctx context.Context, msg string, args ...any) {
	allArgs := append(l.extractContextFields(ctx), args...)
	l.Logger.InfoContext(ctx, msg, allArgs...)
}

// Warn logs a warning message with optional key-value pairs.
func (l *Logger) Warn(msg string, args ...any) {
	l.Logger.Warn(msg, args...)
}

// WarnCtx logs a warning message with context-extracted fields and optional key-value pairs.
func (l *Logger) WarnCtx(ctx context.Context, msg string, args ...any) {
	allArgs := append(l.extractContextFields(ctx), args...)
	l.Logger.WarnContext(ctx, msg, allArgs...)
}

// Error logs an error message with the error object and optional key-value pairs.
func (l *Logger) Error(msg string, err error, args ...any) {
	allArgs := append([]any{"error", err}, args...)
	l.Logger.Error(msg, allArgs...)
}

// ErrorCtx logs an error message with context-extracted fields, error object, and optional key-value pairs.
func (l *Logger) ErrorCtx(ctx context.Context, msg string, err error, args ...any) {
	allArgs := append([]any{"error", err}, args...)
	l.Logger.ErrorContext(ctx, msg, append(l.extractContextFields(ctx), allArgs...)...)
}

// Fatalf logs a fatal error message with formatting and exits the program with status code 1.
func (l *Logger) Fatalf(format string, args ...any) {
	l.Logger.Error(fmt.Sprintf(format, args...))
	os.Exit(1)
}

// FatalfCtx logs a fatal error message with formatting and context fields, then exits the program.
func (l *Logger) FatalfCtx(ctx context.Context, format string, args ...any) {
	contextFields := l.extractContextFields(ctx)
	l.Logger.ErrorContext(ctx, fmt.Sprintf(format, args...), contextFields...)
	os.Exit(1)
}

// extractContextFields extracts relevant logging fields from the request context.
func (l *Logger) extractContextFields(ctx context.Context) []any {
	var fields []any

	if userID := ctx.Value("user_id"); userID != nil {
		fields = append(fields, "user_id", userID)
	}
	if userEmail := ctx.Value("user_email"); userEmail != nil {
		fields = append(fields, "user_email", userEmail)
	}

	if userRole := ctx.Value("user_role"); userRole != nil {
		fields = append(fields, "user_role", userRole)
	}

	if userRoles := ctx.Value("user_roles"); userRoles != nil {
		if roles, ok := userRoles.([]string); ok {
			fields = append(fields, "user_roles", roles)
		}
	}

	// Request fields
	if requestID := ctx.Value("request_id"); requestID != nil {
		fields = append(fields, "request_id", requestID)
	}
	if sessionID := ctx.Value("session_id"); sessionID != nil {
		fields = append(fields, "session_id", sessionID)
	}

	return fields
}
