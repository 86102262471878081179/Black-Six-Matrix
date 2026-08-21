package deepseekengine

import (
	"fmt"
	"sync"
	"time"
)

// ReasoningCore implements the Mixture of Experts reasoning engine
type ReasoningCore struct {
	mu           sync.RWMutex
	Experts      map[string]*Expert
	ReasoningLog []ReasoningResult
}

// Expert represents a specialized reasoning expert
type Expert struct {
	ID          string
	Name        string
	Specialty   string
	Confidence  float64
	Enabled     bool
	LastUsed    time.Time
}

// ReasoningResult represents the output of a reasoning operation
type ReasoningResult struct {
	ID         string
	Timestamp  time.Time
	Query      string
	ExpertsUsed []string
	Conclusion string
	Confidence float64
}

// NewReasoningCore creates a new reasoning core
func NewReasoningCore() *ReasoningCore {
	return &ReasoningCore{
		Experts:      make(map[string]*Expert),
		ReasoningLog: []ReasoningResult{},
	}
}

// RegisterExpert registers a new reasoning expert
func (rc *ReasoningCore) RegisterExpert(expert *Expert) error {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if _, exists := rc.Experts[expert.ID]; exists {
		return fmt.Errorf("expert %s already registered", expert.ID)
	}

	expert.LastUsed = time.Now()
	rc.Experts[expert.ID] = expert
	return nil
}

// Reason performs reasoning using the Mixture of Experts approach
func (rc *ReasoningCore) Reason(query string) (*ReasoningResult, error) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if len(rc.Experts) == 0 {
		return nil, fmt.Errorf("no experts registered")
	}

	// Select enabled experts
	var enabledExperts []string
	totalConfidence := 0.0

	for id, expert := range rc.Experts {
		if expert.Enabled {
			enabledExperts = append(enabledExperts, id)
			totalConfidence += expert.Confidence
		}
	}

	if len(enabledExperts) == 0 {
		return nil, fmt.Errorf("no enabled experts")
	}

	// Calculate average confidence
	avgConfidence := totalConfidence / float64(len(enabledExperts))

	// Generate conclusion
	conclusion := fmt.Sprintf("Reasoning complete: %d experts consulted. Query: %s", len(enabledExperts), query)

	result := &ReasoningResult{
		ID:          generateID(),
		Timestamp:   time.Now(),
		Query:       query,
		ExpertsUsed: enabledExperts,
		Conclusion:  conclusion,
		Confidence:  avgConfidence,
	}

	rc.ReasoningLog = append(rc.ReasoningLog, *result)

	return result, nil
}

// GetExpertsForSpecialty returns all experts for a given specialty
func (rc *ReasoningCore) GetExpertsForSpecialty(specialty string) []Expert {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	var experts []Expert
	for _, expert := range rc.Experts {
		if expert.Specialty == specialty && expert.Enabled {
			experts = append(experts, *expert)
		}
	}
	return experts
}

// UpdateExpertConfidence updates an expert's confidence level
func (rc *ReasoningCore) UpdateExpertConfidence(expertID string, confidence float64) error {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	expert, exists := rc.Experts[expertID]
	if !exists {
		return fmt.Errorf("expert not found")
	}

	if confidence < 0 || confidence > 1 {
		return fmt.Errorf("confidence must be between 0 and 1")
	}

	expert.Confidence = confidence
	expert.LastUsed = time.Now()
	return nil
}

// GetReasoningLog returns the reasoning history
func (rc *ReasoningCore) GetReasoningLog() []ReasoningResult {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	return rc.ReasoningLog
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
