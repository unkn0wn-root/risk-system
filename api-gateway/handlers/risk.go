package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"user-risk-system/pkg/errors"
	"user-risk-system/pkg/validator"
	pb_risk "user-risk-system/proto/risk"
)

// RiskHandler manages risk assessment and rule administration endpoints
type RiskHandler struct {
	riskClient      pb_risk.RiskServiceClient
	riskAdminClient pb_risk.RiskAdminServiceClient
}

// NewRiskHandler creates a new risk handler with risk service clients
func NewRiskHandler(riskClient pb_risk.RiskServiceClient, riskAdminClient pb_risk.RiskAdminServiceClient) *RiskHandler {
	return &RiskHandler{
		riskClient:      riskClient,
		riskAdminClient: riskAdminClient,
	}
}

// CreateRiskRuleRequest represents the payload for creating a new risk rule
type CreateRiskRuleRequest struct {
	Name          string  `json:"name" validate:"required"`
	Type          string  `json:"type" validate:"required"`
	Category      string  `json:"category" validate:"required"`
	Value         string  `json:"value" validate:"required"`
	Score         int32   `json:"score" validate:"required,min=1,max=1000"`
	IsActive      bool    `json:"is_active"`
	Confidence    float64 `json:"confidence" validate:"min=0,max=1"`
	ExpiresInDays int32   `json:"expires_in_days"`
}

// CreateRiskRuleResponse represents the response for risk rule creation
type CreateRiskRuleResponse struct {
	RuleID  string `json:"rule_id,omitempty"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// CheckRiskRequest represents the payload for risk assessment
type CheckRiskRequest struct {
	UserID    string `json:"user_id" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
	Phone     string `json:"phone"`
}

// CheckRiskResponse represents the response for risk assessment
type CheckRiskResponse struct {
	UserID    string   `json:"user_id"`
	IsRisky   bool     `json:"is_risky"`
	RiskLevel string   `json:"risk_level"`
	Reason    string   `json:"reason"`
	Flags     []string `json:"flags"`
	Error     string   `json:"error,omitempty"`
}

// UpdateRiskRuleRequest represents the payload for updating an existing risk rule
type UpdateRiskRuleRequest struct {
	Name          string  `json:"name" validate:"required"`
	Type          string  `json:"type" validate:"required"`
	Category      string  `json:"category" validate:"required"`
	Value         string  `json:"value" validate:"required"`
	Score         int32   `json:"score" validate:"required,min=1,max=1000"`
	IsActive      bool    `json:"is_active"`
	Confidence    float64 `json:"confidence" validate:"min=0,max=1"`
	ExpiresInDays int32   `json:"expires_in_days"`
}

// UpdateRiskRuleResponse represents the response for risk rule updates
type UpdateRiskRuleResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// CreateRiskRule creates a new risk rule (admin only)
func (h *RiskHandler) CreateRiskRule(w http.ResponseWriter, r *http.Request) {
	var req CreateRiskRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.ErrInvalidJSON.SendJSON(w)
		return
	}

	v := validator.New()
	v.Required("name", req.Name).
		Required("type", req.Type).
		Required("category", req.Category).
		Required("value", req.Value).
		Min("score", float64(req.Score), 1).
		Max("score", float64(req.Score), 1000)

	if req.Confidence != 0 {
		v.Min("confidence", req.Confidence, 0).Max("confidence", req.Confidence, 1)
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

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	grpcReq := &pb_risk.CreateRiskRuleRequest{
		Name:          req.Name,
		Type:          req.Type,
		Category:      req.Category,
		Value:         req.Value,
		Score:         req.Score,
		IsActive:      req.IsActive,
		Confidence:    req.Confidence,
		ExpiresInDays: req.ExpiresInDays,
	}

	grpcResp, err := h.riskAdminClient.CreateRiskRule(ctx, grpcReq)
	if err != nil {
		errors.ErrInternalServerError.WithMessage("Failed to create risk rule").WithDetails(err.Error()).SendJSON(w)
		return
	}

	response := CreateRiskRuleResponse{
		RuleID:  grpcResp.RuleId,
		Success: grpcResp.Success,
		Error:   grpcResp.Error,
	}

	statusCode := http.StatusCreated
	if grpcResp.Error != "" {
		statusCode = http.StatusBadRequest
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// UpdateRiskRule modifies an existing risk rule (admin only)
func (h *RiskHandler) UpdateRiskRule(w http.ResponseWriter, r *http.Request) {
	ruleID := chi.URLParam(r, "id")
	if ruleID == "" {
		errors.ErrMissingRequiredFileds.WithMessage("Rule ID is required").SendJSON(w)
		return
	}

	var req UpdateRiskRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.ErrInvalidJSON.SendJSON(w)
		return
	}

	v := validator.New()
	v.Required("name", req.Name).
		Required("type", req.Type).
		Required("category", req.Category).
		Required("value", req.Value).
		Min("score", float64(req.Score), 1).
		Max("score", float64(req.Score), 1000)

	if req.Confidence != 0 {
		v.Min("confidence", req.Confidence, 0).Max("confidence", req.Confidence, 1)
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

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	grpcReq := &pb_risk.UpdateRiskRuleRequest{
		RuleId:        ruleID,
		Name:          req.Name,
		Type:          req.Type,
		Category:      req.Category,
		Value:         req.Value,
		Score:         req.Score,
		IsActive:      req.IsActive,
		Confidence:    req.Confidence,
		ExpiresInDays: req.ExpiresInDays,
	}

	grpcResp, err := h.riskAdminClient.UpdateRiskRule(ctx, grpcReq)
	if err != nil {
		errors.ErrInternalServerError.WithMessage("Failed to update risk rule").SendJSON(w)
		return
	}

	response := UpdateRiskRuleResponse{
		Success: grpcResp.Success,
		Error:   grpcResp.Error,
	}

	statusCode := http.StatusOK
	if grpcResp.Error != "" {
		statusCode = http.StatusBadRequest
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// CheckRisk evaluates user data against risk rules
func (h *RiskHandler) CheckRisk(w http.ResponseWriter, r *http.Request) {
	var req CheckRiskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.ErrInvalidJSON.SendJSON(w)
		return
	}

	v := validator.New()
	v.Required("user_id", req.UserID).
		Required("email", req.Email).
		Email("email", req.Email).
		Required("first_name", req.FirstName).
		Required("last_name", req.LastName)

	if req.Phone != "" {
		v.Phone("phone", req.Phone)
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

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	grpcReq := &pb_risk.RiskCheckRequest{
		UserId:    req.UserID,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
	}

	grpcResp, err := h.riskClient.CheckRisk(ctx, grpcReq)
	if err != nil {
		response := CheckRiskResponse{
			Error: "Failed to check risk",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := CheckRiskResponse{
		UserID:    grpcResp.UserId,
		IsRisky:   grpcResp.IsRisky,
		RiskLevel: grpcResp.RiskLevel,
		Reason:    grpcResp.Reason,
		Flags:     grpcResp.Flags,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteRiskRule removes a risk rule by ID (admin only)
func (h *RiskHandler) DeleteRiskRule(w http.ResponseWriter, r *http.Request) {
	ruleID := chi.URLParam(r, "id")
	if ruleID == "" {
		errors.ErrMissingRequiredFileds.WithMessage("Rule ID is required").SendJSON(w)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	grpcReq := &pb_risk.DeleteRiskRuleRequest{
		RuleId: ruleID,
	}

	grpcResp, err := h.riskAdminClient.DeleteRiskRule(ctx, grpcReq)
	if err != nil {
		errors.ErrInternalServerError.WithMessage("Failed to delete risk rule").SendJSON(w)
		return
	}

	response := map[string]interface{}{
		"success": grpcResp.Success,
	}

	if grpcResp.Error != "" {
		response["error"] = grpcResp.Error
	}

	statusCode := http.StatusOK
	if grpcResp.Error != "" {
		statusCode = http.StatusBadRequest
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// ListRiskRules retrieves all active risk rules (admin only)
func (h *RiskHandler) ListRiskRules(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	grpcReq := &pb_risk.ListRiskRulesRequest{
		ActiveOnly: true,
		Page:       1,
		PageSize:   100,
	}

	grpcResp, err := h.riskAdminClient.ListRiskRules(ctx, grpcReq)
	if err != nil {
		errors.ErrInternalServerError.WithMessage("Failed to list risk rules").SendJSON(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(grpcResp)
}
