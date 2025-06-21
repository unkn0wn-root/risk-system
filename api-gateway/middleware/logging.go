package middleware

import (
	"log"
	"net/http"
	"time"
	"user-risk-system/pkg/logger"
)

// responseWriter wraps http.ResponseWriter to capture status codes
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

type LoggerMiddlewareConfig struct {
	Log            *logger.Logger
	SkipPaths      []string
	AllowedOrigins []string
}

// WriteHeader captures and stores the HTTP status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// LoggingMiddleware logs HTTP request details including method, path, status, and duration
func NewLoggingMiddleware(config LoggerMiddlewareConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, path := range config.SkipPaths {
				if r.URL.Path == path {
					next.ServeHTTP(w, r)
					return
				}
			}
			start := time.Now()
			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(rw, r)

			duration := time.Since(start)
			log.Printf("%s %s %d %v", r.Method, r.URL.Path, rw.statusCode, duration)
		})
	}
}

// CORSMiddleware handles Cross-Origin Resource Sharing headers
func CORSMiddleware(config LoggerMiddlewareConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			allowed := false

			for _, allowedOrigin := range config.AllowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
					allowed = true
					break
				}
			}

			if !allowed && len(config.AllowedOrigins) > 0 {
				w.Header().Set("Access-Control-Allow-Origin", config.AllowedOrigins[0])
			}

			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
