package runtimegovernance

import (
	"fmt"
	"sync"
	"time"
)

// PolicyEngine enforces runtime policies
type PolicyEngine struct {
	mu              sync.RWMutex
	ActivePolicies  map[string]*Policy
	ViolationLog    []PolicyViolation
	EnforcementMode string // STRICT, WARN, REPORT
}

// PolicyViolation records policy violations
type PolicyViolation struct {
	ID        string
	Timestamp time.Time
	PolicyID  string
	Resource  string
	Action    string
	Reason    string
	Severity  string // CRITICAL, HIGH, MEDIUM, LOW
}

// NewPolicyEngine creates a new policy engine
func NewPolicyEngine(mode string) *PolicyEngine {
	return &PolicyEngine{
		ActivePolicies:  make(map[string]*Policy),
		ViolationLog:    []PolicyViolation{},
		EnforcementMode: mode,
	}
}

// EnforcePolicy enforces a policy on a request
func (pe *PolicyEngine) EnforcePolicy(request *PolicyRequest) (bool, error) {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	for _, policy := range pe.ActivePolicies {
		if !policy.Active {
			continue
		}

		if violates, reason := pe.checkViolation(policy, request); violates {
			violation := PolicyViolation{
				ID:        generateID(),
				Timestamp: time.Now(),
				PolicyID:  policy.ID,
				Resource:  request.Resource,
				Action:    request.Action,
				Reason:    reason,
				Severity:  "HIGH",
			}

			pe.mu.RUnlock()
			pe.logViolation(violation)
			pe.mu.RLock()

			if pe.EnforcementMode == "STRICT" {
				return false, fmt.Errorf("policy violation: %s", reason)
			}
		}
	}

	return true, nil
}

// checkViolation checks if a request violates a policy
func (pe *PolicyEngine) checkViolation(policy *Policy, request *PolicyRequest) (bool, string) {
	for _, rule := range policy.Rules {
		if rule.Action == "DENY" && rule.Condition == request.Resource {
			return true, fmt.Sprintf("Resource %s is denied by policy %s", request.Resource, policy.Name)
		}
	}
	return false, ""
}

// logViolation records a policy violation
func (pe *PolicyEngine) logViolation(violation PolicyViolation) {
	pe.mu.Lock()
	defer pe.mu.Unlock()
	pe.ViolationLog = append(pe.ViolationLog, violation)
}

// PolicyRequest represents a policy enforcement request
type PolicyRequest struct {
	Resource string
	Action   string
	Actor    string
	Context  map[string]interface{}
}

// RegisterPolicy registers a new policy
func (pe *PolicyEngine) RegisterPolicy(policy *Policy) error {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	if _, exists := pe.ActivePolicies[policy.ID]; exists {
		return fmt.Errorf("policy %s already exists", policy.ID)
	}

	pe.ActivePolicies[policy.ID] = policy
	return nil
}

// GetViolationLog returns all recorded violations
func (pe *PolicyEngine) GetViolationLog() []PolicyViolation {
	pe.mu.RLock()
	defer pe.mu.RUnlock()
	return pe.ViolationLog
}
