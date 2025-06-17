package repository

import (
	"fmt"
	"time"
	"user-risk-system/cmd/risk-engine/models"

	"gorm.io/gorm"
)

type RiskRepository struct {
	db *gorm.DB
}

func NewRiskRepository(db *gorm.DB) *RiskRepository {
	return &RiskRepository{db: db}
}

func (r *RiskRepository) GetActiveRules() ([]models.RiskRule, error) {
	var rules []models.RiskRule

	result := r.db.Where("is_active = ? AND (expires_at IS NULL OR expires_at > ?)",
		true, time.Now()).
		Order("score DESC").
		Find(&rules)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to query risk rules: %w", result.Error)
	}

	return rules, nil
}

func (r *RiskRepository) GetRulesByCategory(category string) ([]models.RiskRule, error) {
	var rules []models.RiskRule

	result := r.db.Where("category = ? AND is_active = ? AND (expires_at IS NULL OR expires_at > ?)",
		category, true, time.Now()).
		Order("score DESC").
		Find(&rules)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to query risk rules by category: %w", result.Error)
	}

	return rules, nil
}

func (r *RiskRepository) CreateRule(rule *models.RiskRule) error {
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()

	result := r.db.Create(rule)
	if result.Error != nil {
		return fmt.Errorf("failed to create risk rule: %w", result.Error)
	}

	return nil
}

func (r *RiskRepository) UpdateRule(rule *models.RiskRule) error {
	rule.UpdatedAt = time.Now()

	result := r.db.Save(rule)
	if result.Error != nil {
		return fmt.Errorf("failed to update risk rule: %w", result.Error)
	}

	return nil
}

func (r *RiskRepository) GetRuleByID(id string) (*models.RiskRule, error) {
	var rule models.RiskRule

	result := r.db.Where("id = ?", id).First(&rule)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("risk rule not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get risk rule: %w", result.Error)
	}

	return &rule, nil
}

func (r *RiskRepository) DeleteRule(id string) error {
	result := r.db.Delete(&models.RiskRule{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete risk rule: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("risk rule not found: %s", id)
	}

	return nil
}

func (r *RiskRepository) DeactivateRule(id string) error {
	result := r.db.Model(&models.RiskRule{}).
		Where("id = ?", id).
		Update("is_active", false)

	if result.Error != nil {
		return fmt.Errorf("failed to deactivate risk rule: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("risk rule not found: %s", id)
	}

	return nil
}
