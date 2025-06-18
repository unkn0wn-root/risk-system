// Package auth provides HTTP and gRPC authentication middleware for securing endpoints.
package auth

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthMiddleware provides authentication functionality for both HTTP and gRPC services.
// It wraps a JWTManager to handle token validation and user context enrichment.
type AuthMiddleware struct {
	jwtManager *JWTManager
}

// NewAuthMiddleware creates a new authentication middleware instance.
// It requires a configured JWTManager for token operations.
func NewAuthMiddleware(jwtManager *JWTManager) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager: jwtManager,
	}
}

// HTTPMiddleware provides JWT authentication for HTTP requests.
// It validates tokens, enriches the request context with user data, and handles public endpoints.
func (a *AuthMiddleware) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip authentication for health checks and public endpoints
		if a.isPublicEndpoint(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		token := a.extractTokenFromHTTP(r)
		if token == "" {
			a.unauthorizedHTTP(w, "Missing authorization token")
			return
		}

		claims, err := a.jwtManager.ValidateToken(token)
		if err != nil {
			a.unauthorizedHTTP(w, "Invalid token: "+err.Error())
			return
		}

		// Add user info to request context
		ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "user_email", claims.Email)
		ctx = context.WithValue(ctx, "user_roles", claims.Roles)
		ctx = context.WithValue(ctx, "claims", claims)
		ctx = context.WithValue(ctx, "jwt_token", token)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole creates an HTTP middleware that restricts access to users with specific roles.
// It should be used after the main HTTPMiddleware to enforce role-based authorization.
func (a *AuthMiddleware) RequireRole(roles ...UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value("claims").(*Claims)
			if !ok {
				a.forbiddenHTTP(w, "Authentication required")
				return
			}

			if !claims.HasAnyRole(roles...) {
				a.forbiddenHTTP(w, "Insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GRPCUnaryInterceptor provides JWT authentication for gRPC unary method calls.
// It validates tokens, enriches the context with user data, and skips auth for public methods.
func (a *AuthMiddleware) GRPCUnaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	// Skip authentication for health checks and internal calls
	if a.isPublicGRPCMethod(info.FullMethod) {
		return handler(ctx, req)
	}

	token, err := a.extractTokenFromGRPC(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Missing authorization token: %v", err)
	}

	claims, err := a.jwtManager.ValidateToken(token)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Invalid token: %v", err)
	}

	// Add user info to gRPC context
	ctx = context.WithValue(ctx, "user_id", claims.UserID)
	ctx = context.WithValue(ctx, "user_email", claims.Email)
	ctx = context.WithValue(ctx, "user_roles", claims.Roles)
	ctx = context.WithValue(ctx, "claims", claims)

	return handler(ctx, req)
}

// GRPCRequireRole creates a gRPC interceptor that enforces role-based access control.
// It should be chained after the main authentication interceptor.
func (a *AuthMiddleware) GRPCRequireRole(roles ...UserRole) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		claims, ok := ctx.Value("claims").(*Claims)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "Authentication required")
		}

		if !claims.HasAnyRole(roles...) {
			return nil, status.Errorf(codes.PermissionDenied, "Insufficient permissions")
		}

		return handler(ctx, req)
	}
}

// extractTokenFromHTTP extracts JWT token from HTTP request headers or query parameters.
// It supports both Authorization header (Bearer token) and query parameter formats.
func (a *AuthMiddleware) extractTokenFromHTTP(r *http.Request) string {
	// Check Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return parts[1]
		}
	}

	// Check query parameter (for websockets, etc.)
	return r.URL.Query().Get("token")
}

// extractTokenFromGRPC extracts JWT token from gRPC metadata.
// It looks for the authorization header in the incoming metadata.
func (a *AuthMiddleware) extractTokenFromGRPC(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("missing metadata")
	}

	tokens := md["authorization"]
	if len(tokens) == 0 {
		return "", fmt.Errorf("missing authorization header")
	}

	token := tokens[0]
	if !strings.HasPrefix(strings.ToLower(token), "bearer ") {
		return "", fmt.Errorf("invalid authorization header format")
	}

	return token[7:], nil // Remove "Bearer " prefix
}

// isPublicEndpoint determines if an HTTP endpoint should skip authentication.
// Public endpoints include health checks and authentication-related endpoints.
func (a *AuthMiddleware) isPublicEndpoint(path string) bool {
	publicPaths := []string{
		"/",
		"/api/v1/health",
		"/api/v1/auth/login",
		"/api/v1/auth/register",
		"/api/v1/auth/refresh",
	}

	for _, publicPath := range publicPaths {
		if path == publicPath {
			return true
		}
	}
	return false
}

// isPublicGRPCMethod determines if a gRPC method should skip authentication.
// Public methods include health checks and user registration/login endpoints.
func (a *AuthMiddleware) isPublicGRPCMethod(method string) bool {
	publicMethods := []string{
		"/grpc.health.v1.Health/Check",
		"/user.UserService/Login",
		"/user.UserService/Register",
	}

	for _, publicMethod := range publicMethods {
		if method == publicMethod {
			return true
		}
	}
	return false
}

// unauthorizedHTTP sends a 401 Unauthorized response with the given message.
// It sets appropriate headers and logs the unauthorized access attempt.
func (a *AuthMiddleware) unauthorizedHTTP(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"error": "` + message + `"}`))
	log.Printf("🔒 Unauthorized access: %s", message)
}

// forbiddenHTTP sends a 403 Forbidden response with the given message.
// It's used when authentication succeeds but authorization fails due to insufficient permissions.
func (a *AuthMiddleware) forbiddenHTTP(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	w.Write([]byte(`{"error": "` + message + `"}`))
	log.Printf("🚫 Forbidden access: %s", message)
}
