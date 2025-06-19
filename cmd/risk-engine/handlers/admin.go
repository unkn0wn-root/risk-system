package handlers

import (
	"context"
	"time"
	"user-risk-system/cmd/risk-engine/models"
	"user-risk-system/cmd/risk-engine/repository"
	"user-risk-system/pkg/logger"
	pb_risk "user-risk-system/pkg/proto/risk"

	"github.com/google/uuid"
)

// RiskAdminHandler manages risk rules through administrative gRPC endpoints.
type RiskAdminHandler struct {
	pb_risk.UnimplementedRiskAdminServiceServer
	riskRepo   *repository.RiskRepository
	logger     *logger.Logger
	riskEngine RiskEngineService
}

type RiskEngineService interface {
	InvalidateCache()
}

// NewRiskAdminHandler creates a new administrative handler with repository, logger, and risk engine dependencies.
func NewRiskAdminHandler(riskRepo *repository.RiskRepository, logger *logger.Logger, riskEngine RiskEngineService) *RiskAdminHandler {
	return &RiskAdminHandler{
		riskRepo:   riskRepo,
		logger:     logger,
		riskEngine: riskEngine,
	}
}

// CreateRiskRule adds a new risk rule to the system via gRPC.
// validates the request and creates a rule with optional expiration.
func (h *RiskAdminHandler) CreateRiskRule(ctx context.Context, req *pb_risk.CreateRiskRuleRequest) (*pb_risk.CreateRiskRuleResponse, error) {
	rule := &models.RiskRule{
		ID:         uuid.New().String(),
		Name:       req.Name,
		Type:       req.Type,
		Category:   req.Category,
		Value:      req.Value,
		Score:      int(req.Score),
		IsActive:   req.IsActive,
		Source:     "MANUAL",
		Confidence: req.Confidence,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if req.ExpiresInDays > 0 {
		expiresAt := time.Now().AddDate(0, 0, int(req.ExpiresInDays))
		rule.ExpiresAt = &expiresAt
	}

	if err := h.riskRepo.CreateRule(rule); err != nil {
		h.logger.ErrorCtx(ctx, "Failed to create risk rule", err)
		return nil, err
	}

	// Invalidate cache to ensure new rule is immediately available
	h.riskEngine.InvalidateCache()

	h.logger.InfoCtx(ctx, "Risk rule created", "rule_id", rule.ID, "name", rule.Name)

	return &pb_risk.CreateRiskRuleResponse{
		RuleId:  rule.ID,
		Success: true,
	}, nil
}

// UpdateRiskRule modifies an existing risk rule via gRPC.
// updates all rule fields except ID and creation timestamp.
func (h *RiskAdminHandler) UpdateRiskRule(ctx context.Context, req *pb_risk.UpdateRiskRuleRequest) (*pb_risk.UpdateRiskRuleResponse, error) {
	rule := &models.RiskRule{
		ID:         req.RuleId,
		Name:       req.Name,
		Type:       req.Type,
		Category:   req.Category,
		Value:      req.Value,
		Score:      int(req.Score),
		IsActive:   req.IsActive,
		Source:     "MANUAL",
		Confidence: req.Confidence,
		UpdatedAt:  time.Now(),
	}

	if err := h.riskRepo.UpdateRule(rule); err != nil {
		h.logger.ErrorCtx(ctx, "Failed to update risk rule", err)
		return nil, err
	}

	h.riskEngine.InvalidateCache()

	h.logger.InfoCtx(ctx, "Risk rule updated", "rule_id", rule.ID)

	return &pb_risk.UpdateRiskRuleResponse{
		Success: true,
	}, nil
}

// ListRiskRules retrieves all active risk rules via gRPC.
// returns rules with their current configuration and metadata.
func (h *RiskAdminHandler) ListRiskRules(ctx context.Context, req *pb_risk.ListRiskRulesRequest) (*pb_risk.ListRiskRulesResponse, error) {
	rules, err := h.riskRepo.GetActiveRules()
	if err != nil {
		h.logger.ErrorCtx(ctx, "Failed to list risk rules", err)
		return nil, err
	}

	var pbRules []*pb_risk.RiskRule
	for _, rule := range rules {
		pbRule := &pb_risk.RiskRule{
			Id:         rule.ID,
			Name:       rule.Name,
			Type:       rule.Type,
			Category:   rule.Category,
			Value:      rule.Value,
			Score:      int32(rule.Score),
			IsActive:   rule.IsActive,
			Confidence: rule.Confidence,
			CreatedAt:  rule.CreatedAt.Unix(),
			UpdatedAt:  rule.UpdatedAt.Unix(),
		}
		if rule.ExpiresAt != nil {
			pbRule.ExpiresAt = rule.ExpiresAt.Unix()
		}
		pbRules = append(pbRules, pbRule)
	}

	return &pb_risk.ListRiskRulesResponse{Rules: pbRules}, nil
}

// DeleteRiskRule permanently removes a risk rule from the system via gRPC.
// performs a hard delete and returns an error if the rule doesn't exist.
func (h *RiskAdminHandler) DeleteRiskRule(ctx context.Context, req *pb_risk.DeleteRiskRuleRequest) (*pb_risk.DeleteRiskRuleResponse, error) {
	if err := h.riskRepo.DeleteRule(req.RuleId); err != nil {
		h.logger.ErrorCtx(ctx, "Failed to delete risk rule", err)
		return nil, err
	}

	h.riskEngine.InvalidateCache()

	h.logger.InfoCtx(ctx, "Risk rule deleted", "rule_id", req.RuleId)

	return &pb_risk.DeleteRiskRuleResponse{
		Success: true,
	}, nil
}
