package runtimegovernance

import (
	"crypto/sha256"
	"fmt"
	"time"
)

// JudgmentEngine handles all policy decisions and judgments
type JudgmentEngine struct {
	Policies map[string]*Policy
	AuditLog []JudgmentRecord
}

// Policy represents a governance policy
type Policy struct {
	ID          string
	Name        string
	Rules       []Rule
	Active      bool
	Version     string
	CreatedAt   time.Time
	LastUpdated time.Time
}

// Rule represents a single policy rule
type Rule struct {
	ID        string
	Condition string
	Action    string
	Priority  int
}

// JudgmentRecord tracks all decisions made
type JudgmentRecord struct {
	ID          string
	Timestamp   time.Time
	Decision    string
	PolicyID    string
	RulesMatched []string
	Hash        string
}

// NewJudgmentEngine creates a new judgment engine
func NewJudgmentEngine() *JudgmentEngine {
	return &JudgmentEngine{
		Policies: make(map[string]*Policy),
		AuditLog: []JudgmentRecord{},
	}
}

// AddPolicy adds a new policy to the engine
func (je *JudgmentEngine) AddPolicy(policy *Policy) error {
	if policy.ID == "" {
		return fmt.Errorf("policy ID cannot be empty")
	}
	policy.CreatedAt = time.Now()
	policy.LastUpdated = time.Now()
	je.Policies[policy.ID] = policy
	return nil
}

// EvaluateRequest evaluates a request against all policies
func (je *JudgmentEngine) EvaluateRequest(request map[string]interface{}) (string, error) {
	decision := "DENIED"
	matchedRules := []string{}

	for policyID, policy := range je.Policies {
		if !policy.Active {
			continue
		}

		for _, rule := range policy.Rules {
			if evaluateRule(rule, request) {
				matchedRules = append(matchedRules, rule.ID)
				if rule.Action == "ALLOW" {
					decision = "ALLOWED"
				}
			}
		}
	}

	// Record judgment
	record := JudgmentRecord{
		ID:           generateID(),
		Timestamp:    time.Now(),
		Decision:     decision,
		RulesMatched: matchedRules,
	}
	record.Hash = calculateHash(record)

je.AuditLog = append(je.AuditLog, record)

	return decision, nil
}

// evaluateRule checks if a rule matches the request
func evaluateRule(rule Rule, request map[string]interface{}) bool {
	// Simplified evaluation - in production, use a more sophisticated evaluator
	if resource, ok := request["resource"]; ok {
		return fmt.Sprintf("%v", resource) == rule.Condition
	}
	return false
}

// calculateHash generates SHA256 hash of a judgment
func calculateHash(record JudgmentRecord) string {
	data := fmt.Sprintf("%s-%s-%s", record.ID, record.Timestamp.String(), record.Decision)
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// generateID creates a unique ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// GetAuditLog returns the complete audit trail
func (je *JudgmentEngine) GetAuditLog() []JudgmentRecord {
	return je.AuditLog
}
