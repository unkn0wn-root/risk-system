// Package services implements the core business logic for risk evaluation.
// It provides the risk engine for rule-based assessment and caching mechanisms.
package services

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
	"user-risk-system/cmd/risk-engine/models"
	"user-risk-system/cmd/risk-engine/repository"
	"user-risk-system/pkg/logger"
	pb_risk "user-risk-system/pkg/proto/risk"

	"github.com/google/uuid"
)

// RiskEngine orchestrates risk evaluation against configurable rules.
// It provides caching, rule evaluation, and scoring mechanisms for user data assessment.
type RiskEngine struct {
	riskRepo   *repository.RiskRepository
	logger     *logger.Logger
	ruleCache  map[string][]models.RiskRule // Cache rules by category
	cacheTime  time.Time
	cacheTTL   time.Duration
	cacheMutex sync.RWMutex
}

// NewRiskEngine creates a new risk engine with repository and logger dependencies.
// It initializes the rule cache with a 5-minute TTL for optimal performance.
func NewRiskEngine(riskRepo *repository.RiskRepository, logger *logger.Logger) *RiskEngine {
	return &RiskEngine{
		riskRepo:  riskRepo,
		logger:    logger,
		ruleCache: make(map[string][]models.RiskRule),
		cacheTTL:  5 * time.Minute, // Cache rules for 5 minutes
	}
}

// CheckRisk evaluates user data against all active risk rules.
// It returns a comprehensive risk assessment with flags, scores, and matched rules.
func (re *RiskEngine) CheckRisk(ctx context.Context, req *pb_risk.RiskCheckRequest) (*models.RiskCheckResult, error) {
	result := &models.RiskCheckResult{
		CheckID:      generateCheckID(),
		UserID:       req.UserId,
		IsRisky:      false,
		RiskLevel:    "MINIMAL",
		TotalScore:   0,
		Reason:       "No risk factors detected",
		Flags:        []models.RiskCheckFlag{},
		MatchedRules: []models.RiskCheckRuleMatch{},
		CheckedAt:    time.Now().UTC(),
	}

	// Refresh rules cache if needed
	if err := re.refreshRulesCache(ctx); err != nil {
		re.logger.ErrorCtx(ctx, "Failed to refresh rules cache", err)
		return result, fmt.Errorf("failed to refresh rules cache: %w", err)
	}

	var flagStrings []string
	var matchedRules []models.RiskRule

	// Check email risks
	emailScore, emailFlags, emailRules := re.checkEmailRisk(ctx, req.Email)
	result.TotalScore += emailScore
	flagStrings = append(flagStrings, emailFlags...)
	matchedRules = append(matchedRules, emailRules...)

	// Check name risks
	nameScore, nameFlags, nameRules := re.checkNameRisk(ctx, req.FirstName, req.LastName)
	result.TotalScore += nameScore
	flagStrings = append(flagStrings, nameFlags...)
	matchedRules = append(matchedRules, nameRules...)

	// Check phone risks
	phoneScore, phoneFlags, phoneRules := re.checkPhoneRisk(ctx, req.Phone)
	result.TotalScore += phoneScore
	flagStrings = append(flagStrings, phoneFlags...)
	matchedRules = append(matchedRules, phoneRules...)

	// Determine risk level based on total score
	result.RiskLevel, result.IsRisky = re.calculateRiskLevel(result.TotalScore)

	for _, flagStr := range flagStrings {
		result.Flags = append(result.Flags, models.RiskCheckFlag{
			CheckID: result.CheckID,
			Flag:    flagStr,
		})
	}

	if len(matchedRules) > 0 {
		reasons := make([]string, 0, len(matchedRules))
		for _, rule := range matchedRules {
			adjustedScore := int(float64(rule.Score) * rule.Confidence)

			result.MatchedRules = append(result.MatchedRules, models.RiskCheckRuleMatch{
				CheckID:    result.CheckID,
				RuleID:     rule.ID,
				RuleName:   rule.Name,
				ScoreAdded: adjustedScore,
			})

			// Build reason string
			reasons = append(reasons, fmt.Sprintf("%s (score: %d)", rule.Name, adjustedScore))
		}
		result.Reason = strings.Join(reasons, "; ")
	}

	re.logger.InfoCtx(ctx, "Risk check completed",
		"user_id", req.UserId,
		"email", maskEmail(req.Email),
		"total_score", result.TotalScore,
		"risk_level", result.RiskLevel,
		"is_risky", result.IsRisky,
		"matched_rules", len(matchedRules),
		"flags", strings.Join(flagStrings, ","),
	)

	return result, nil
}

// refreshRulesCache updates the in-memory rule cache when expired.
// It loads rules by category to optimize evaluation performance.
func (re *RiskEngine) refreshRulesCache(ctx context.Context) error {
	re.cacheMutex.RLock()
	cacheExpired := time.Since(re.cacheTime) >= re.cacheTTL
	re.cacheMutex.RUnlock()

	if !cacheExpired {
		return nil
	}

	re.cacheMutex.Lock()
	defer re.cacheMutex.Unlock()

	re.logger.InfoCtx(ctx, "Refreshing risk rules cache")

	newCache := make(map[string][]models.RiskRule)

	// Load rules by category
	categories := []string{"EMAIL", "NAME", "PHONE"}
	for _, category := range categories {
		rules, err := re.riskRepo.GetRulesByCategory(category)
		if err != nil {
			return fmt.Errorf("failed to load %s rules: %w", category, err)
		}
		newCache[category] = rules
	}

	re.ruleCache = newCache
	re.cacheTime = time.Now()

	re.logger.InfoCtx(ctx, "Risk rules cache refreshed",
		"email_rules", len(re.ruleCache["EMAIL"]),
		"name_rules", len(re.ruleCache["NAME"]),
		"phone_rules", len(re.ruleCache["PHONE"]),
		"total_rules", len(re.ruleCache["EMAIL"])+len(re.ruleCache["NAME"])+len(re.ruleCache["PHONE"]),
	)

	return nil
}

// checkEmailRisk evaluates email addresses against email-specific risk rules.
// It returns the total score, flags, and matched rules for the email.
func (re *RiskEngine) checkEmailRisk(ctx context.Context, email string) (int, []string, []models.RiskRule) {
	var totalScore int
	var flags []string
	var matchedRules []models.RiskRule

	emailLower := strings.ToLower(strings.TrimSpace(email))

	re.cacheMutex.RLock()
	rules := make([]models.RiskRule, len(re.ruleCache["EMAIL"]))
	copy(rules, re.ruleCache["EMAIL"])
	re.cacheMutex.RUnlock()

	for _, rule := range rules {
		matched, err := re.evaluateEmailRule(rule, emailLower)
		if err != nil {
			re.logger.WarnCtx(ctx, "Failed to evaluate email rule",
				"rule_id", rule.ID,
				"error", err.Error())
			continue
		}

		if matched {
			// Apply confidence scoring
			adjustedScore := int(float64(rule.Score) * rule.Confidence)
			totalScore += adjustedScore
			flags = append(flags, fmt.Sprintf("EMAIL_%s", rule.Type))
			matchedRules = append(matchedRules, rule)

			re.logger.InfoCtx(ctx, "Email risk rule matched",
				"rule_id", rule.ID,
				"rule_name", rule.Name,
				"rule_type", rule.Type,
				"score_added", adjustedScore,
				"original_score", rule.Score,
				"confidence", rule.Confidence,
			)
		}
	}

	return totalScore, flags, matchedRules
}

// evaluateEmailRule determines if an email matches a specific risk rule.
// It supports blacklist, pattern matching, domain filtering, and containment checks.
func (re *RiskEngine) evaluateEmailRule(rule models.RiskRule, emailLower string) (bool, error) {
	switch rule.Type {
	case "EMAIL_BLACKLIST":
		return strings.EqualFold(emailLower, strings.ToLower(rule.Value)), nil
	case "PATTERN_MATCH":
		matched, err := regexp.MatchString(rule.Value, emailLower)
		if err != nil {
			return false, fmt.Errorf("invalid regex pattern: %w", err)
		}
		return matched, nil
	case "DOMAIN_BLACKLIST":
		domain := extractDomain(emailLower)
		return strings.EqualFold(domain, strings.ToLower(rule.Value)), nil
	case "CONTAINS":
		return strings.Contains(emailLower, strings.ToLower(rule.Value)), nil
	default:
		return false, fmt.Errorf("unknown email rule type: %s", rule.Type)
	}
}

// checkNameRisk evaluates user names against name-specific risk rules.
// It checks first name, last name, and full name combinations.
func (re *RiskEngine) checkNameRisk(ctx context.Context, firstName, lastName string) (int, []string, []models.RiskRule) {
	var totalScore int
	var flags []string
	var matchedRules []models.RiskRule

	firstNameLower := strings.ToLower(strings.TrimSpace(firstName))
	lastNameLower := strings.ToLower(strings.TrimSpace(lastName))
	fullName := strings.TrimSpace(firstNameLower + " " + lastNameLower)

	re.cacheMutex.RLock()
	rules := make([]models.RiskRule, len(re.ruleCache["NAME"]))
	copy(rules, re.ruleCache["NAME"])
	re.cacheMutex.RUnlock()

	for _, rule := range rules {
		matched, err := re.evaluateNameRule(rule, firstNameLower, lastNameLower, fullName)
		if err != nil {
			re.logger.WarnCtx(ctx, "Failed to evaluate name rule",
				"rule_id", rule.ID,
				"error", err.Error())
			continue
		}

		if matched {
			adjustedScore := int(float64(rule.Score) * rule.Confidence)
			totalScore += adjustedScore
			flags = append(flags, fmt.Sprintf("NAME_%s", rule.Type))
			matchedRules = append(matchedRules, rule)

			re.logger.InfoCtx(ctx, "Name risk rule matched",
				"rule_id", rule.ID,
				"rule_name", rule.Name,
				"rule_type", rule.Type,
				"score_added", adjustedScore,
			)
		}
	}

	return totalScore, flags, matchedRules
}

// evaluateNameRule determines if a name matches a specific risk rule.
// It supports blacklists, pattern matching, and containment checks for names.
func (re *RiskEngine) evaluateNameRule(rule models.RiskRule, firstNameLower, lastNameLower, fullName string) (bool, error) {
	switch rule.Type {
	case "NAME_BLACKLIST":
		return strings.EqualFold(fullName, strings.ToLower(rule.Value)), nil
	case "PATTERN_MATCH":
		matched, err := regexp.MatchString(rule.Value, fullName)
		if err != nil {
			return false, fmt.Errorf("invalid regex pattern: %w", err)
		}
		return matched, nil
	case "CONTAINS":
		return strings.Contains(fullName, strings.ToLower(rule.Value)), nil
	case "FIRST_NAME_BLACKLIST":
		return strings.EqualFold(firstNameLower, strings.ToLower(rule.Value)), nil
	case "LAST_NAME_BLACKLIST":
		return strings.EqualFold(lastNameLower, strings.ToLower(rule.Value)), nil
	default:
		return false, fmt.Errorf("unknown name rule type: %s", rule.Type)
	}
}

// checkPhoneRisk evaluates phone numbers against phone-specific risk rules.
// It normalizes phone numbers and checks against various rule types.
func (re *RiskEngine) checkPhoneRisk(ctx context.Context, phone string) (int, []string, []models.RiskRule) {
	var totalScore int
	var flags []string
	var matchedRules []models.RiskRule

	// Normalize phone number (remove spaces, dashes, etc.)
	normalizedPhone := normalizePhoneNumber(phone)

	re.cacheMutex.RLock()
	rules := make([]models.RiskRule, len(re.ruleCache["PHONE"]))
	copy(rules, re.ruleCache["PHONE"])
	re.cacheMutex.RUnlock()

	for _, rule := range rules {
		matched, err := re.evaluatePhoneRule(rule, normalizedPhone)
		if err != nil {
			re.logger.WarnCtx(ctx, "Failed to evaluate phone rule",
				"rule_id", rule.ID,
				"error", err.Error())
			continue
		}

		if matched {
			adjustedScore := int(float64(rule.Score) * rule.Confidence)
			totalScore += adjustedScore
			flags = append(flags, fmt.Sprintf("PHONE_%s", rule.Type))
			matchedRules = append(matchedRules, rule)

			re.logger.InfoCtx(ctx, "Phone risk rule matched",
				"rule_id", rule.ID,
				"rule_name", rule.Name,
				"rule_type", rule.Type,
				"score_added", adjustedScore,
			)
		}
	}

	return totalScore, flags, matchedRules
}

// evaluatePhoneRule determines if a phone number matches a specific risk rule.
// It supports blacklists, pattern matching, and prefix-based checks.
func (re *RiskEngine) evaluatePhoneRule(rule models.RiskRule, normalizedPhone string) (bool, error) {
	switch rule.Type {
	case "PHONE_BLACKLIST":
		return normalizedPhone == normalizePhoneNumber(rule.Value), nil
	case "PATTERN_MATCH":
		matched, err := regexp.MatchString(rule.Value, normalizedPhone)
		if err != nil {
			return false, fmt.Errorf("invalid regex pattern: %w", err)
		}
		return matched, nil
	case "PREFIX_BLACKLIST":
		return strings.HasPrefix(normalizedPhone, normalizePhoneNumber(rule.Value)), nil
	default:
		return false, fmt.Errorf("unknown phone rule type: %s", rule.Type)
	}
}

// calculateRiskLevel determines risk level and risky status based on total score.
// It uses predefined thresholds to classify risk from MINIMAL to CRITICAL.
func (re *RiskEngine) calculateRiskLevel(totalScore int) (string, bool) {
	switch {
	case totalScore >= 100:
		return "CRITICAL", true
	case totalScore >= 80:
		return "HIGH", true
	case totalScore >= 40:
		return "MEDIUM", true
	case totalScore >= 20:
		return "LOW", false
	default:
		return "MINIMAL", false
	}
}

// GetCachedRules returns a copy of the currently cached rules for a category.
// It prevents external modification by returning a deep copy of cached rules.
func (re *RiskEngine) GetCachedRules(category string) []models.RiskRule {
	re.cacheMutex.RLock()
	defer re.cacheMutex.RUnlock()

	rules, exists := re.ruleCache[category]
	if !exists {
		return []models.RiskRule{}
	}

	// Return a copy to prevent external modification
	result := make([]models.RiskRule, len(rules))
	copy(result, rules)
	return result
}

// InvalidateCache forces a cache refresh on the next request.
// It resets the cache timestamp to trigger immediate rule reloading.
func (re *RiskEngine) InvalidateCache() {
	re.cacheMutex.Lock()
	defer re.cacheMutex.Unlock()

	re.cacheTime = time.Time{} // Reset to zero time to force refresh
}

// GetCacheStats returns information about the current cache state.
// It provides metrics about cache age, rule counts, and last update time.
func (re *RiskEngine) GetCacheStats() map[string]interface{} {
	re.cacheMutex.RLock()
	defer re.cacheMutex.RUnlock()

	return map[string]interface{}{
		"cache_age_seconds": time.Since(re.cacheTime).Seconds(),
		"cache_ttl_seconds": re.cacheTTL.Seconds(),
		"email_rules_count": len(re.ruleCache["EMAIL"]),
		"name_rules_count":  len(re.ruleCache["NAME"]),
		"phone_rules_count": len(re.ruleCache["PHONE"]),
		"last_updated":      re.cacheTime.Format(time.RFC3339),
	}
}

// extractDomain extracts the domain portion from an email address.
// It returns the domain part after the @ symbol, or empty string if invalid.
func extractDomain(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return ""
	}
	return strings.ToLower(parts[1])
}

// normalizePhoneNumber removes common formatting characters from phone numbers.
// It strips spaces, dashes, parentheses, plus signs, and dots for consistent comparison.
func normalizePhoneNumber(phone string) string {
	// Remove common phone number formatting
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	phone = strings.ReplaceAll(phone, "(", "")
	phone = strings.ReplaceAll(phone, ")", "")
	phone = strings.ReplaceAll(phone, "+", "")
	phone = strings.ReplaceAll(phone, ".", "")
	return strings.TrimSpace(phone)
}

// isValidEmail performs basic email format validation using regex.
// It checks for standard email structure but doesn't verify deliverability.
func isValidEmail(email string) bool {
	// Basic email validation regex
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// maskEmail obscures email addresses for secure logging.
// It shows only the first 2 characters of the username to protect privacy.
func maskEmail(email string) string {
	if email == "" {
		return ""
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "***"
	}

	username := parts[0]
	domain := parts[1]

	if len(username) <= 2 {
		return "***@" + domain
	}

	return username[:2] + "***@" + domain
}

// generateCheckID creates a unique identifier for risk check operations.
// It combines timestamp and UUID elements for guaranteed uniqueness.
func generateCheckID() string {
	return fmt.Sprintf("check_%d_%s", time.Now().Unix(), strings.ReplaceAll(uuid.New().String()[:8], "-", ""))
}
