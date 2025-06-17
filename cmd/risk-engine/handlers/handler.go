package handlers

import (
	"context"
	"user-risk-system/cmd/risk-engine/services"
	"user-risk-system/pkg/logger"
	pb_risk "user-risk-system/pkg/proto/risk"
)

type RiskHandler struct {
	pb_risk.UnimplementedRiskServiceServer
	riskEngine *services.RiskEngine    // Does the actual risk checking
	analytics  *services.RiskAnalytics // Stores results for reporting
	logger     *logger.Logger
}

func NewRiskHandler(
	riskEngine *services.RiskEngine,
	analytics *services.RiskAnalytics,
	logger *logger.Logger,
) *RiskHandler {
	return &RiskHandler{
		riskEngine: riskEngine,
		analytics:  analytics,
		logger:     logger,
	}
}

func (h *RiskHandler) CheckRisk(ctx context.Context, req *pb_risk.RiskCheckRequest) (*pb_risk.RiskCheckResponse, error) {
	ctx = context.WithValue(ctx, "user_id", req.UserId)
	ctx = context.WithValue(ctx, "user_email", req.Email)

	h.logger.InfoCtx(ctx, "Checking risk for user", "user_id", req.UserId, "email", req.Email)

	result, err := h.riskEngine.CheckRisk(ctx, req)
	if err != nil {
		h.logger.ErrorCtx(ctx, "Risk check failed", err)
		return nil, err
	}

	go func() {
		if err := h.analytics.StoreRiskResult(context.Background(), result); err != nil {
			h.logger.Error("Failed to store risk result for analytics", err)
		}
	}()

	flagStrings := make([]string, len(result.Flags))
	for i, flag := range result.Flags {
		flagStrings[i] = flag.Flag
	}

	response := &pb_risk.RiskCheckResponse{
		UserId:    result.UserID,
		IsRisky:   result.IsRisky,
		RiskLevel: result.RiskLevel,
		Reason:    result.Reason,
		Flags:     flagStrings,
	}

	if result.IsRisky {
		h.logger.InfoCtx(ctx, "RISK DETECTED for user",
			"user_id", req.UserId,
			"risk_level", result.RiskLevel,
			"reason", result.Reason,
			"flags", flagStrings,
		)
	} else {
		h.logger.InfoCtx(ctx, "No risk detected for user", "user_id", req.UserId)
	}

	return response, nil
}
