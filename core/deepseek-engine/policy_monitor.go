package deepseekengine

import (
	"fmt"
	"sync"
	"time"
)

// PolicyMonitor monitors policy violations in real-time
type PolicyMonitor struct {
	mu        sync.RWMutex
	Policies  map[string]*MonitoredPolicy
	Violations []PolicyViolationAlert
	AlertQueue chan *PolicyViolationAlert
}

// MonitoredPolicy represents a policy being monitored
type MonitoredPolicy struct {
	ID           string
	Name         string
	Thresholds   map[string]interface{}
	Active       bool
	LastViolated time.Time
}

// PolicyViolationAlert represents a policy violation alert
type PolicyViolationAlert struct {
	ID        string
	Timestamp time.Time
	PolicyID  string
	Severity  string // CRITICAL, HIGH, MEDIUM, LOW
	Message   string
	Resource  string
}

// NewPolicyMonitor creates a new policy monitor
func NewPolicyMonitor() *PolicyMonitor {
	return &PolicyMonitor{
		Policies:   make(map[string]*MonitoredPolicy),
		Violations: []PolicyViolationAlert{},
		AlertQueue: make(chan *PolicyViolationAlert, 1000),
	}
}

// MonitorPolicy adds a policy to the monitor
func (pm *PolicyMonitor) MonitorPolicy(policy *MonitoredPolicy) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if policy.ID == "" {
		return fmt.Errorf("policy ID cannot be empty")
	}

	policy.Active = true
	pm.Policies[policy.ID] = policy
	return nil
}

// ReportViolation reports a policy violation
func (pm *PolicyMonitor) ReportViolation(policyID, severity, message, resource string) error {
	pm.mu.RLock()
	policy, exists := pm.Policies[policyID]
	pm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("policy %s not found", policyID)
	}

	alert := &PolicyViolationAlert{
		ID:        generateID(),
		Timestamp: time.Now(),
		PolicyID:  policyID,
		Severity:  severity,
		Message:   message,
		Resource:  resource,
	}

	// Log violation
	pm.mu.Lock()
	policy.LastViolated = time.Now()
	pm.Violations = append(pm.Violations, *alert)
	pm.mu.Unlock()

	// Send to alert queue (non-blocking)
	select {
	case pm.AlertQueue <- alert:
	default:
		// Queue full, drop alert
	}

	return nil
}

// GetViolations returns all recorded violations
func (pm *PolicyMonitor) GetViolations() []PolicyViolationAlert {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.Violations
}

// GetViolationsBySeverity returns violations filtered by severity
func (pm *PolicyMonitor) GetViolationsBySeverity(severity string) []PolicyViolationAlert {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var filtered []PolicyViolationAlert
	for _, v := range pm.Violations {
		if v.Severity == severity {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

// GetCriticalViolations returns all critical violations
func (pm *PolicyMonitor) GetCriticalViolations() []PolicyViolationAlert {
	return pm.GetViolationsBySeverity("CRITICAL")
}

// ProcessAlerts processes alerts from the queue
func (pm *PolicyMonitor) ProcessAlerts(handler func(*PolicyViolationAlert)) {
	for alert := range pm.AlertQueue {
		handler(alert)
	}
}

// StopMonitoring stops monitoring a policy
func (pm *PolicyMonitor) StopMonitoring(policyID string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	policy, exists := pm.Policies[policyID]
	if !exists {
		return fmt.Errorf("policy not found")
	}

	policy.Active = false
	return nil
}
