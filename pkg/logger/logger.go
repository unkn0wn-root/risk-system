package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"
)

type Logger struct {
	*slog.Logger
}

type LogConfig struct {
	Level       string // debug, info, warn, error
	Format      string // json, text
	ServiceName string
	Environment string
}

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

// Info logs an info message
func (l *Logger) Info(msg string, args ...any) {
	l.Logger.Info(msg, args...)
}

func (l *Logger) InfoCtx(ctx context.Context, msg string, args ...any) {
	allArgs := append(l.extractContextFields(ctx), args...)
	l.Logger.InfoContext(ctx, msg, allArgs...)
}

func (l *Logger) Warn(msg string, args ...any) {
	l.Logger.Warn(msg, args...)
}

func (l *Logger) WarnCtx(ctx context.Context, msg string, args ...any) {
	allArgs := append(l.extractContextFields(ctx), args...)
	l.Logger.WarnContext(ctx, msg, allArgs...)
}

// Error logs an error message
func (l *Logger) Error(msg string, err error, args ...any) {
	allArgs := append([]any{"error", err}, args...)
	l.Logger.Error(msg, allArgs...)
}

func (l *Logger) ErrorCtx(ctx context.Context, msg string, err error, args ...any) {
	allArgs := append([]any{"error", err}, args...)
	l.Logger.ErrorContext(ctx, msg, append(l.extractContextFields(ctx), allArgs...)...)
}

// Fatalf logs a fatal error message with formatting and exits the program
func (l *Logger) Fatalf(format string, args ...any) {
	l.Logger.Error(fmt.Sprintf(format, args...))
	os.Exit(1)
}

// FatalfCtx logs a fatal error message with formatting, context fields, and exits the program
func (l *Logger) FatalfCtx(ctx context.Context, format string, args ...any) {
	contextFields := l.extractContextFields(ctx)
	l.Logger.ErrorContext(ctx, fmt.Sprintf(format, args...), contextFields...)
	os.Exit(1)
}

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
