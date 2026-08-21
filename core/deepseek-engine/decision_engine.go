package deepseekengine

import (
	"fmt"
	"sync"
	"time"
)

// DecisionEngine makes autonomous decisions based on policies and reasoning
type DecisionEngine struct {
	mu               sync.RWMutex
	Decisions        map[string]*Decision
	DecisionHistory  []Decision
	MaxHistorySize   int
}

// Decision represents a single autonomous decision
type Decision struct {
	ID                string
	Timestamp         time.Time
	Query             string
	ReasoningResult   string
	Policy            string
	Action            string
	ConfidenceScore   float64
	Status            string // PENDING, APPROVED, EXECUTED, REJECTED
	ExecutedAt        time.Time
	ExecutionResult   string
}

// DecisionRequest represents a request for a decision
type DecisionRequest struct {
	Query       string
	Policy      string
	Context     map[string]interface{}
	Urgency     string // LOW, MEDIUM, HIGH, CRITICAL
	Timeout     time.Duration
}

// NewDecisionEngine creates a new decision engine
func NewDecisionEngine() *DecisionEngine {
	return &DecisionEngine{
		Decisions:       make(map[string]*Decision),
		DecisionHistory: []Decision{},
		MaxHistorySize:  10000,
	}
}

// MakeDecision makes an autonomous decision
func (de *DecisionEngine) MakeDecision(request *DecisionRequest) (*Decision, error) {
	de.mu.Lock()
	defer de.mu.Unlock()

	decision := &Decision{
		ID:              generateID(),
		Timestamp:       time.Now(),
		Query:           request.Query,
		Policy:          request.Policy,
		ConfidenceScore: 0.85,
		Status:          "PENDING",
	}

	// Evaluate based on policy and context
	decision.Action, decision.ReasoningResult = de.evaluateRequest(request)

	de.Decisions[decision.ID] = decision
	de.DecisionHistory = append(de.DecisionHistory, *decision)

	// Cleanup old history
	if len(de.DecisionHistory) > de.MaxHistorySize {
		de.DecisionHistory = de.DecisionHistory[1:]
	}

	return decision, nil
}

// ApproveDecision approves a pending decision
func (de *DecisionEngine) ApproveDecision(decisionID string) error {
	de.mu.Lock()
	defer de.mu.Unlock()

	decision, exists := de.Decisions[decisionID]
	if !exists {
		return fmt.Errorf("decision not found")
	}

	if decision.Status != "PENDING" {
		return fmt.Errorf("decision is not pending")
	}

	decision.Status = "APPROVED"
	return nil
}

// ExecuteDecision executes an approved decision
func (de *DecisionEngine) ExecuteDecision(decisionID string) error {
	de.mu.Lock()
	defer de.mu.Unlock()

	decision, exists := de.Decisions[decisionID]
	if !exists {
		return fmt.Errorf("decision not found")
	}

	if decision.Status != "APPROVED" {
		return fmt.Errorf("decision is not approved")
	}

	decision.Status = "EXECUTED"
	decision.ExecutedAt = time.Now()
	decision.ExecutionResult = fmt.Sprintf("Action %s executed", decision.Action)

	return nil
}

// RejectDecision rejects a pending decision
func (de *DecisionEngine) RejectDecision(decisionID, reason string) error {
	de.mu.Lock()
	defer de.mu.Unlock()

	decision, exists := de.Decisions[decisionID]
	if !exists {
		return fmt.Errorf("decision not found")
	}

	decision.Status = "REJECTED"
	decision.ExecutionResult = reason

	return nil
}

// evaluateRequest evaluates a decision request
func (de *DecisionEngine) evaluateRequest(request *DecisionRequest) (string, string) {
	// Simplified evaluation logic
	action := "ALLOW"
	reason := "Policy evaluation complete"

	if request.Urgency == "CRITICAL" {
		action = "EXPEDITE"
	}

	return action, reason
}

// GetDecision retrieves a specific decision
func (de *DecisionEngine) GetDecision(decisionID string) (*Decision, error) {
	de.mu.RLock()
	defer de.mu.RUnlock()

	decision, exists := de.Decisions[decisionID]
	if !exists {
		return nil, fmt.Errorf("decision not found")
	}

	return decision, nil
}

// GetDecisionHistory returns the decision history
func (de *DecisionEngine) GetDecisionHistory() []Decision {
	de.mu.RLock()
	defer de.mu.RUnlock()
	return de.DecisionHistory
}
