package models

import (
	"time"

	"gorm.io/gorm"
)

type RiskRule struct {
	ID         string     `json:"id" gorm:"primaryKey;type:varchar(255)"`
	Name       string     `json:"name" gorm:"type:varchar(255);not null"`
	Type       string     `json:"type" gorm:"type:varchar(100);not null"`     // EMAIL_BLACKLIST, NAME_BLACKLIST, PATTERN_MATCH
	Category   string     `json:"category" gorm:"type:varchar(100);not null"` // EMAIL, NAME, PHONE
	Value      string     `json:"value" gorm:"type:text;not null"`            // The actual value or pattern
	Score      int        `json:"score" gorm:"not null"`                      // Risk score to add
	IsActive   bool       `json:"is_active" gorm:"default:true"`
	Source     string     `json:"source" gorm:"type:varchar(100);not null"`        // MANUAL, EXTERNAL_API, ML_MODEL
	Confidence float64    `json:"confidence" gorm:"type:decimal(3,2);default:1.0"` // 0.0 to 1.0
	CreatedAt  time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	ExpiresAt  *time.Time `json:"expires_at" gorm:"index"` // For temporary rules
}

func (RiskRule) TableName() string {
	return "risk_rules"
}

type RiskCheckResult struct {
	ID         uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	CheckID    string    `json:"check_id" gorm:"uniqueIndex;type:varchar(255);not null"`
	UserID     string    `json:"user_id" gorm:"type:varchar(255);not null;index"`
	IsRisky    bool      `json:"is_risky" gorm:"default:false;index"`
	RiskLevel  string    `json:"risk_level" gorm:"type:varchar(50)"` // LOW, MEDIUM, HIGH, CRITICAL
	TotalScore int       `json:"total_score" gorm:"default:0"`
	Reason     string    `json:"reason" gorm:"type:text"`
	CheckedAt  time.Time `json:"checked_at" gorm:"index"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	Flags        []RiskCheckFlag      `json:"flags" gorm:"foreignKey:CheckID;references:CheckID"`
	MatchedRules []RiskCheckRuleMatch `json:"matched_rules" gorm:"foreignKey:CheckID;references:CheckID"`
}

func (RiskCheckResult) TableName() string {
	return "risk_check_results"
}

type RiskCheckFlag struct {
	ID      uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	CheckID string `json:"check_id" gorm:"type:varchar(255);not null;index"`
	Flag    string `json:"flag" gorm:"type:varchar(255);not null"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (RiskCheckFlag) TableName() string {
	return "risk_check_flags"
}

type RiskCheckRuleMatch struct {
	ID         uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	CheckID    string `json:"check_id" gorm:"type:varchar(255);not null;index"`
	RuleID     string `json:"rule_id" gorm:"type:varchar(255);not null"`
	RuleName   string `json:"rule_name" gorm:"type:varchar(255);not null"`
	ScoreAdded int    `json:"score_added" gorm:"not null"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (RiskCheckRuleMatch) TableName() string {
	return "risk_check_rule_matches"
}

// AutoMigrate runs database migrations for all models
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&RiskRule{},
		&RiskCheckResult{},
		&RiskCheckFlag{},
		&RiskCheckRuleMatch{},
	)
}
