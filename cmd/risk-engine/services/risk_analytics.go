// Package services implements analytics and reporting for risk engine operations.
// It provides statistical analysis, trend data, and historical risk assessment storage.
package services

import (
	"context"
	"fmt"
	"time"
	"user-risk-system/cmd/risk-engine/models"
	"user-risk-system/pkg/logger"

	"gorm.io/gorm"
)

// RiskAnalytics provides statistical analysis and reporting for risk assessments.
// It stores risk check results and generates analytics data for monitoring and reporting.
type RiskAnalytics struct {
	db     *gorm.DB
	logger *logger.Logger
}

// NewRiskAnalytics creates a new analytics service with database and logger dependencies.
func NewRiskAnalytics(db *gorm.DB, logger *logger.Logger) *RiskAnalytics {
	return &RiskAnalytics{
		db:     db,
		logger: logger,
	}
}

// RiskStats represents aggregated risk assessment statistics.
// It includes counts, rates, scores, and trend data for reporting dashboards.
type RiskStats struct {
	TotalChecks  int64        `json:"total_checks"`
	RiskyUsers   int64        `json:"risky_users"`
	RiskRate     float64      `json:"risk_rate"`
	AvgRiskScore float64      `json:"avg_risk_score"`
	TopFlags     []FlagCount  `json:"top_flags"`
	TrendData    []TrendPoint `json:"trend_data"`
}

// FlagCount represents the frequency of specific risk flags.
// It tracks how often particular risk indicators are triggered.
type FlagCount struct {
	Flag  string `json:"flag"`
	Count int64  `json:"count"`
}

// TrendPoint represents risk assessment data for a specific date.
// It contains daily counts and totals for trend analysis.
type TrendPoint struct {
	Date       time.Time `json:"date"`
	RiskCount  int64     `json:"risk_count"`
	TotalCount int64     `json:"total_count"`
}

// GetRiskStats computes comprehensive risk statistics for the specified number of days.
// It includes total checks, risk rates, average scores, top flags, and trend data.
func (ra *RiskAnalytics) GetRiskStats(ctx context.Context, days int) (*RiskStats, error) {
	stats := &RiskStats{}
	since := time.Now().AddDate(0, 0, -days)

	// Get total checks and risky users
	var result struct {
		TotalChecks  int64   `gorm:"column:total_checks"`
		RiskyUsers   int64   `gorm:"column:risky_users"`
		AvgRiskScore float64 `gorm:"column:avg_risk_score"`
	}

	err := ra.db.WithContext(ctx).Model(&models.RiskCheckResult{}).
		Select(`
			COUNT(*) as total_checks,
			COUNT(CASE WHEN is_risky = true THEN 1 END) as risky_users,
			COALESCE(AVG(total_score), 0) as avg_risk_score
		`).
		Where("checked_at >= ?", since).
		Scan(&result).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get basic stats: %w", err)
	}

	stats.TotalChecks = result.TotalChecks
	stats.RiskyUsers = result.RiskyUsers
	stats.AvgRiskScore = result.AvgRiskScore

	if stats.TotalChecks > 0 {
		stats.RiskRate = float64(stats.RiskyUsers) / float64(stats.TotalChecks)
	}

	// Get top flags
	var flagResults []struct {
		Flag  string `gorm:"column:flag"`
		Count int64  `gorm:"column:count"`
	}

	err = ra.db.WithContext(ctx).
		Table("risk_check_flags rcf").
		Select("rcf.flag, COUNT(*) as count").
		Joins("JOIN risk_check_results rcr ON rcf.check_id = rcr.check_id").
		Where("rcr.checked_at >= ?", since).
		Group("rcf.flag").
		Order("count DESC").
		Limit(10).
		Scan(&flagResults).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get flag stats: %w", err)
	}

	for _, flag := range flagResults {
		stats.TopFlags = append(stats.TopFlags, FlagCount{
			Flag:  flag.Flag,
			Count: flag.Count,
		})
	}

	// Get trend data
	var trendResults []struct {
		Date       time.Time `gorm:"column:date"`
		RiskCount  int64     `gorm:"column:risk_count"`
		TotalCount int64     `gorm:"column:total_count"`
	}

	err = ra.db.WithContext(ctx).Model(&models.RiskCheckResult{}).
		Select(`
			DATE(checked_at) as date,
			COUNT(CASE WHEN is_risky = true THEN 1 END) as risk_count,
			COUNT(*) as total_count
		`).
		Where("checked_at >= ?", since).
		Group("DATE(checked_at)").
		Order("date").
		Scan(&trendResults).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get trend data: %w", err)
	}

	for _, trend := range trendResults {
		stats.TrendData = append(stats.TrendData, TrendPoint{
			Date:       trend.Date,
			RiskCount:  trend.RiskCount,
			TotalCount: trend.TotalCount,
		})
	}

	return stats, nil
}

// StoreRiskResult persists a risk check result with associated flags and rule matches.
// It uses database transactions to ensure data consistency across related tables.
func (ra *RiskAnalytics) StoreRiskResult(ctx context.Context, result *models.RiskCheckResult) error {
	tx := ra.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer tx.Rollback()

	if err := tx.Create(result).Error; err != nil {
		return fmt.Errorf("failed to create risk result: %w", err)
	}

	if len(result.Flags) > 0 {
		for i := range result.Flags {
			// Ensure CheckID is set for each flag
			result.Flags[i].CheckID = result.CheckID
		}

		if err := tx.Create(&result.Flags).Error; err != nil {
			ra.logger.ErrorCtx(ctx, "Failed to insert flags", err)
			return fmt.Errorf("failed to insert flags: %w", err)
		}
	}

	if len(result.MatchedRules) > 0 {
		for i := range result.MatchedRules {
			// Ensure CheckID is set for each rule match
			result.MatchedRules[i].CheckID = result.CheckID
		}

		if err := tx.Create(&result.MatchedRules).Error; err != nil {
			ra.logger.ErrorCtx(ctx, "Failed to insert rule matches", err)
			return fmt.Errorf("failed to insert rule matches: %w", err)
		}
	}

	return tx.Commit().Error
}

// GetRiskHistory retrieves historical risk assessments for a specific user.
// It includes associated flags and rule matches, ordered by most recent first.
func (ra *RiskAnalytics) GetRiskHistory(ctx context.Context, userID string, limit int) ([]models.RiskCheckResult, error) {
	var results []models.RiskCheckResult

	err := ra.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Preload("Flags").
		Preload("MatchedRules").
		Order("checked_at DESC").
		Limit(limit).
		Find(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get risk history: %w", err)
	}

	return results, nil
}

// GetRiskSummaryByDateRange gets aggregated risk data for a specific date range.
// It provides summary statistics for custom time periods defined by start and end dates.
func (ra *RiskAnalytics) GetRiskSummaryByDateRange(ctx context.Context, startDate, endDate time.Time) (*RiskStats, error) {
	return ra.getRiskStatsInRange(ctx, startDate, endDate)
}

// getRiskStatsInRange computes risk statistics for a custom date range.
// It's an internal helper method for date-bounded analytics queries.
func (ra *RiskAnalytics) getRiskStatsInRange(ctx context.Context, startDate, endDate time.Time) (*RiskStats, error) {
	stats := &RiskStats{}

	// Get basic stats for the date range
	var result struct {
		TotalChecks  int64   `gorm:"column:total_checks"`
		RiskyUsers   int64   `gorm:"column:risky_users"`
		AvgRiskScore float64 `gorm:"column:avg_risk_score"`
	}

	err := ra.db.WithContext(ctx).Model(&models.RiskCheckResult{}).
		Select(`
			COUNT(*) as total_checks,
			COUNT(CASE WHEN is_risky = true THEN 1 END) as risky_users,
			COALESCE(AVG(total_score), 0) as avg_risk_score
		`).
		Where("checked_at BETWEEN ? AND ?", startDate, endDate).
		Scan(&result).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get stats for date range: %w", err)
	}

	stats.TotalChecks = result.TotalChecks
	stats.RiskyUsers = result.RiskyUsers
	stats.AvgRiskScore = result.AvgRiskScore

	if stats.TotalChecks > 0 {
		stats.RiskRate = float64(stats.RiskyUsers) / float64(stats.TotalChecks)
	}

	return stats, nil
}
