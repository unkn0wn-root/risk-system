package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"user-risk-system/pkg/auth"
	"user-risk-system/pkg/errors"
	pb_user "user-risk-system/pkg/proto/user"
	"user-risk-system/pkg/validator"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	userClient pb_user.UserServiceClient
	jwtManager *auth.JWTManager
}

// NewAuthHandler creates a new authentication handler with user service client and JWT manager
func NewAuthHandler(userClient pb_user.UserServiceClient, jwtManager *auth.JWTManager) *AuthHandler {
	return &AuthHandler{
		userClient: userClient,
		jwtManager: jwtManager,
	}
}

// LoginRequest represents the request payload for user login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// RegisterRequest represents the request payload for user registration
type RegisterRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
	Phone     string `json:"phone"`
}

// AuthResponse represents the response payload for authentication endpoints
type AuthResponse struct {
	User         *UserResponse `json:"user"`
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time     `json:"expires_at"`
}

// RefreshTokenRequest represents the request payload for token refresh
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// Login authenticates a user and returns JWT token
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.ErrInvalidJSON.SendJSON(w)
		return
	}

	v := validator.New()
	v.Required("email", req.Email).
		Email("email", req.Email).
		Required("password", req.Password).
		MinLength("password", req.Password, 6)

	if !v.IsValid() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":             "Validation failed",
			"validation_errors": v.Errors(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Call user service to authenticate
	grpcReq := &pb_user.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	grpcResp, err := h.userClient.Login(ctx, grpcReq)
	if err != nil {
		errors.ErrAuthenticationFailed.SendJSON(w)
		return
	}

	if grpcResp.Error != "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": grpcResp.Error})
		return
	}

	// Generate JWT token
	token, err := h.jwtManager.GenerateToken(
		grpcResp.User.Id,
		grpcResp.User.Email,
		grpcResp.User.Roles,
	)
	if err != nil {
		errors.ErrInternalServerError.WithMessage("Failed to generate token").SendJSON(w)
		return
	}

	user := &UserResponse{
		ID:         grpcResp.User.Id,
		Email:      grpcResp.User.Email,
		FirstName:  grpcResp.User.FirstName,
		LastName:   grpcResp.User.LastName,
		Phone:      grpcResp.User.Phone,
		Roles:      grpcResp.User.Roles,
		IsActive:   grpcResp.User.IsActive,
		IsVerified: grpcResp.User.IsVerified,
		CreatedAt:  grpcResp.User.CreatedAt.AsTime(),
	}

	response := AuthResponse{
		User:        user,
		AccessToken: token,
		ExpiresAt:   time.Now().Add(24 * time.Hour), // Should match JWT expiry
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Register creates a new user account and returns JWT token for the user
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.ErrInvalidJSON.SendJSON(w)
		return
	}

	v := validator.New()
	v.Required("email", req.Email).
		Email("email", req.Email).
		Required("password", req.Password).
		MinLength("password", req.Password, 8).
		Required("first_name", req.FirstName).
		MinLength("first_name", req.FirstName, 2).
		Required("last_name", req.LastName).
		MinLength("last_name", req.LastName, 2).
		Phone("phone", req.Phone) // Phone is optional but validated if provided

	if !v.IsValid() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":             "Validation failed",
			"validation_errors": v.Errors(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Call user service to register
	grpcReq := &pb_user.RegisterRequest{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
	}

	grpcResp, err := h.userClient.Register(ctx, grpcReq)
	if err != nil {
		errors.NewAppError("REGISTRATION_FAILED", "Registration failed", "").SendJSON(w)
		return
	}

	if grpcResp.Error != "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": grpcResp.Error})
		return
	}

	token, err := h.jwtManager.GenerateToken(
		grpcResp.User.Id,
		grpcResp.User.Email,
		grpcResp.User.Roles,
	)
	if err != nil {
		errors.ErrInvalidToken.SendJSON(w)
		return
	}

	user := &UserResponse{
		ID:         grpcResp.User.Id,
		Email:      grpcResp.User.Email,
		FirstName:  grpcResp.User.FirstName,
		LastName:   grpcResp.User.LastName,
		Phone:      grpcResp.User.Phone,
		Roles:      grpcResp.User.Roles,
		IsActive:   grpcResp.User.IsActive,
		IsVerified: grpcResp.User.IsVerified,
		CreatedAt:  grpcResp.User.CreatedAt.AsTime(),
	}

	response := AuthResponse{
		User:        user,
		AccessToken: token,
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// RefreshToken generates a new access token from a valid refresh token
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.ErrInvalidJSON.SendJSON(w)
		return
	}

	newToken, err := h.jwtManager.RefreshToken(req.RefreshToken)
	if err != nil {
		errors.ErrInvalidToken.SendJSON(w)
		return
	}

	response := map[string]interface{}{
		"access_token": newToken,
		"expires_at":   time.Now().Add(24 * time.Hour),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetProfile retrieves the authenticated user's profile information
func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	// User info is already in context from middleware
	userID := r.Context().Value("user_id").(string)
	userRoles := r.Context().Value("user_roles").([]string)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	grpcReq := &pb_user.GetUserRequest{Id: userID}
	grpcResp, err := h.userClient.GetUser(ctx, grpcReq)
	if err != nil {
		errors.ErrInternalServerError.WithMessage("Could not get user").SendJSON(w)
		return
	}

	if grpcResp.Error != "" {
		errors.ErrInternalServerError.WithMessage(grpcResp.Error).SendJSON(w)
		return
	}

	user := &UserResponse{
		ID:         grpcResp.User.Id,
		Email:      grpcResp.User.Email,
		FirstName:  grpcResp.User.FirstName,
		LastName:   grpcResp.User.LastName,
		Phone:      grpcResp.User.Phone,
		Roles:      userRoles,
		IsActive:   grpcResp.User.IsActive,
		IsVerified: grpcResp.User.IsVerified,
		CreatedAt:  grpcResp.User.CreatedAt.AsTime(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
