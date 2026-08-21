package runtimegovernance

import (
	"fmt"
	"regexp"
	"time"
)

// RuleValidator validates rules against conditions
type RuleValidator struct {
	ValidationCache map[string]*ValidationResult
	MaxCacheSize    int
}

// ValidationResult stores validation results
type ValidationResult struct {
	RuleID    string
	Valid     bool
	Timestamp time.Time
	Reason    string
}

// NewRuleValidator creates a new rule validator
func NewRuleValidator() *RuleValidator {
	return &RuleValidator{
		ValidationCache: make(map[string]*ValidationResult),
		MaxCacheSize:    10000,
	}
}

// ValidateRule validates a single rule
func (rv *RuleValidator) ValidateRule(rule *Rule, context map[string]interface{}) (bool, error) {
	cacheKey := fmt.Sprintf("%s-%v", rule.ID, context)

	if result, exists := rv.ValidationCache[cacheKey]; exists {
		if time.Since(result.Timestamp) < 5*time.Minute {
			return result.Valid, nil
		}
	}

	valid, reason := rv.evaluateCondition(rule.Condition, context)

	result := &ValidationResult{
		RuleID:    rule.ID,
		Valid:     valid,
		Timestamp: time.Now(),
		Reason:    reason,
	}

	rv.ValidationCache[cacheKey] = result

	// Cleanup old cache entries if needed
	if len(rv.ValidationCache) > rv.MaxCacheSize {
		rv.cleanupCache()
	}

	return valid, nil
}

// evaluateCondition evaluates a condition against context
func (rv *RuleValidator) evaluateCondition(condition string, context map[string]interface{}) (bool, string) {
	// Simple pattern matching - in production, use a more sophisticated evaluator
	if condition == "*" {
		return true, "Wildcard match"
	}

	// Check for regex pattern
	if matched, err := regexp.MatchString(condition, fmt.Sprintf("%v", context)); err == nil {
		return matched, fmt.Sprintf("Pattern %s matched", condition)
	}

	return false, fmt.Sprintf("Condition %s not met", condition)
}

// ValidateRuleset validates a complete set of rules
func (rv *RuleValidator) ValidateRuleset(rules []Rule, context map[string]interface{}) ([]ValidationResult, error) {
	results := []ValidationResult{}

	for _, rule := range rules {
		valid, _ := rv.ValidateRule(&rule, context)
		results = append(results, ValidationResult{
			RuleID:    rule.ID,
			Valid:     valid,
			Timestamp: time.Now(),
		})
	}

	return results, nil
}

// cleanupCache removes oldest cache entries
func (rv *RuleValidator) cleanupCache() {
	oldestKey := ""
	var oldestTime time.Time

	for key, result := range rv.ValidationCache {
		if oldestTime.IsZero() || result.Timestamp.Before(oldestTime) {
			oldestKey = key
			oldestTime = result.Timestamp
		}
	}

	if oldestKey != "" {
		delete(rv.ValidationCache, oldestKey)
	}
}
