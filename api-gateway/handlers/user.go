package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"user-risk-system/pkg/errors"
	"user-risk-system/pkg/validator"
	pb_user "user-risk-system/proto/user"
)

// UserHandler manages user-related HTTP endpoints
type UserHandler struct {
	userClient pb_user.UserServiceClient
}

// NewUserHandler creates a new user handler with user service client
func NewUserHandler(userClient pb_user.UserServiceClient) *UserHandler {
	return &UserHandler{
		userClient: userClient,
	}
}

// CreateUserRequest represents the payload for creating a new user
type CreateUserRequest struct {
	Email     string `json:"email" validate:"required,email"`
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
	Phone     string `json:"phone"`
}

// CreateUserResponse represents the response for user creation
type CreateUserResponse struct {
	User  *UserResponse `json:"user,omitempty"`
	Error string        `json:"error,omitempty"`
}

// GetUserResponse represents the response for user retrieval
type GetUserResponse struct {
	User  *UserResponse `json:"user,omitempty"`
	Error string        `json:"error,omitempty"`
}

// UserResponse represents the standard user data response structure
type UserResponse struct {
	ID         string    `json:"id"`
	Email      string    `json:"email"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Phone      string    `json:"phone"`
	Roles      []string  `json:"roles"`
	IsActive   bool      `json:"is_active"`
	IsVerified bool      `json:"is_verified"`
	CreatedAt  time.Time `json:"created_at"`
}

// CreateUser creates a new user account (admin only)
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.ErrInvalidJSON.SendJSON(w)
		return
	}

	v := validator.New()
	v.Required("email", req.Email).
		Email("email", req.Email).
		Required("first_name", req.FirstName).
		MinLength("first_name", req.FirstName, 2).
		Required("last_name", req.LastName).
		MinLength("last_name", req.LastName, 2).
		Phone("phone", req.Phone) // Phone validation only if provided (not required)

	if !v.IsValid() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":             "Validation failed",
			"validation_errors": v.Errors(),
		})
		return
	}

	// Call user service via gRPC
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	grpcReq := &pb_user.CreateUserRequest{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
	}

	grpcResp, err := h.userClient.CreateUser(ctx, grpcReq)
	if err != nil {
		errors.ErrInternalServerError.WithMessage("Failed to create user").SendJSON(w)
		return
	}

	if grpcResp.Error != "" {
		response := CreateUserResponse{Error: grpcResp.Error}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Convert protobuf user to JSON response
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

	response := CreateUserResponse{User: user}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetUser retrieves a user by ID
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if userID == "" {
		errors.ErrMissingRequiredFileds.WithMessage("User ID is required").SendJSON(w)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	grpcReq := &pb_user.GetUserRequest{Id: userID}
	grpcResp, err := h.userClient.GetUser(ctx, grpcReq)
	if err != nil {
		errors.ErrInternalServerError.WithMessage("Failed to get user").SendJSON(w)
		return
	}

	if grpcResp.Error != "" {
		response := GetUserResponse{Error: grpcResp.Error}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
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

	response := GetUserResponse{User: user}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateUser modifies user profile information
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if userID == "" {
		errors.ErrMissingRequiredFileds.WithMessage("User ID is required").SendJSON(w)
		return
	}

	var updateReq struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Phone     string `json:"phone"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		errors.ErrInvalidJSON.SendJSON(w)
		return
	}

	v := validator.New()
	if updateReq.FirstName != "" {
		v.MinLength("first_name", updateReq.FirstName, 2)
	}
	if updateReq.LastName != "" {
		v.MinLength("last_name", updateReq.LastName, 2)
	}
	if updateReq.Phone != "" {
		v.Phone("phone", updateReq.Phone)
	}

	if !v.IsValid() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":             "Validation failed",
			"validation_errors": v.Errors(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	grpcReq := &pb_user.UpdateUserRequest{
		Id:        userID,
		FirstName: updateReq.FirstName,
		LastName:  updateReq.LastName,
		Phone:     updateReq.Phone,
	}

	grpcResp, err := h.userClient.UpdateUser(ctx, grpcReq)
	if err != nil {
		errors.ErrInternalServerError.WithMessage("Failed to update user").SendJSON(w)
		return
	}

	if grpcResp.Error != "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": grpcResp.Error})
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// ListUsers retrieves all users (admin only)
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	// todo: call a ListUsers gRPC method
	// For now, return a placeholder response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]string{
		"error": "ListUsers endpoint not yet implemented",
	})
}

// HealthCheck returns the service health status
func (h *UserHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status":    "healthy",
		"service":   "api-gateway",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
