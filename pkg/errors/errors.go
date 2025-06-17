package errors

import (
	"encoding/json"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func NewAppError(code, message, details string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

func (e *AppError) Error() string {
	return e.Message
}

// Predefined errors
var (
	ErrUserNotFound               = &AppError{Code: "USER_NOT_FOUND", Message: "User not found"}
	ErrInvalidPassword            = &AppError{Code: "INVALID_PASSWORD", Message: "Invalid password"}
	ErrEmailExists                = &AppError{Code: "EMAIL_EXISTS", Message: "Email already exists"}
	ErrInvalidToken               = &AppError{Code: "INVALID_TOKEN", Message: "Invalid or expired token"}
	ErrInvalidJSON                = &AppError{Code: "INVALID_JSON", Message: "Invalid JSON payload"}
	ErrInsufficientRole           = &AppError{Code: "INSUFFICIENT_ROLE", Message: "Insufficient permissions"}
	ErrRateLimitExceeded          = &AppError{Code: "RATE_LIMIT_EXCEEDED", Message: "Rate limit exceeded"}
	ErrUserInactive               = &AppError{Code: "USER_INACTIVE", Message: "Account is deactivated"}
	ErrPasswordHashFailed         = &AppError{Code: "PASSWORD_HASH_FAILED", Message: "Failed to process password"}
	ErrUserCreateFailed           = &AppError{Code: "USER_CREATE_FAILED", Message: "Failed to create user account"}
	ErrUserUpdateFailed           = &AppError{Code: "USER_UPDATE_FAILED", Message: "Failed to update user"}
	ErrRequiredUsernameOrPassword = &AppError{Code: "UNAME_OR_PASS_REQUIRED", Message: "Username/Password required"}
	ErrAuthenticationFailed       = &AppError{Code: "AUTHENTICATION_FAILED", Message: "Authentication failed"}
	ErrMissingRequiredFileds      = &AppError{Code: "MISSING_REQUIRED_FILEDS", Message: "Missing required fileds"}
	ErrInternalServerError        = &AppError{Code: "INTERNAL_SERVER_ERROR", Message: "Something went wrong"}
)

// HTTP status mapping
func (e *AppError) HTTPStatus() int {
	switch e.Code {
	case "USER_NOT_FOUND":
		return http.StatusNotFound
	case "INVALID_PASSWORD", "INVALID_TOKEN", "AUTHENTICATION_FAILED":
		return http.StatusUnauthorized
	case "EMAIL_EXISTS":
		return http.StatusConflict
	case "INSUFFICIENT_ROLE":
		return http.StatusForbidden
	case "RATE_LIMIT_EXCEEDED":
		return http.StatusTooManyRequests
	case "USER_INACTIVE":
		return http.StatusForbidden
	case "PASSWORD_HASH_FAILED", "INVALID_JSON", "UNAME_OR_PASS_REQUIRED", "MISSING_REQUIRED_FILEDS":
		return http.StatusBadRequest
	case "USER_CREATE_FAILED":
		return http.StatusInternalServerError
	case "USER_UPDATE_FAILED":
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

func (e *AppError) WithMessage(message string) *AppError {
	return &AppError{
		Code:    e.Code,
		Message: message,
		Details: e.Details,
	}
}

func (e *AppError) WithDetails(details string) *AppError {
	return &AppError{
		Code:    e.Code,
		Message: e.Message,
		Details: details,
	}
}

// gRPC status mapping
func (e *AppError) GRPCStatus() *status.Status {
	switch e.Code {
	case "USER_NOT_FOUND":
		return status.New(codes.NotFound, e.Message)
	case "INVALID_PASSWORD", "INVALID_TOKEN":
		return status.New(codes.Unauthenticated, e.Message)
	case "INSUFFICIENT_ROLE":
		return status.New(codes.PermissionDenied, e.Message)
	default:
		return status.New(codes.Internal, e.Message)
	}
}

// SendJSON sends the error as a JSON response
func (e *AppError) SendJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.HTTPStatus())
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   e.Message,
		"details": e.Details, // Only included if not empty
	})
}
