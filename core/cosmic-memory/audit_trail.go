package cosmicmemory

import (
	"fmt"
	"sync"
	"time"
)

// AuditTrail manages the immutable audit log
type AuditTrail struct {
	mu      sync.RWMutex
	Records []AuditRecord
	MaxSize int
}

// AuditRecord represents a single audit log entry
type AuditRecord struct {
	ID            string
	TransactionID string
	Action        string
	Actor         string
	Resource      string
	Status        string
	Result        string
	Timestamp     int64
	Hash          string
	CreatedAt     time.Time
}

// NewAuditTrail creates a new audit trail
func NewAuditTrail(maxSize int) *AuditTrail {
	return &AuditTrail{
		Records: []AuditRecord{},
		MaxSize: maxSize,
	}
}

// LogAction logs an action to the audit trail
func (at *AuditTrail) LogAction(action, actor, resource, status, result string) (string, error) {
	at.mu.Lock()
	defer at.mu.Unlock()

	record := AuditRecord{
		ID:            generateID(),
		TransactionID: generateTransactionID(),
		Action:        action,
		Actor:         actor,
		Resource:      resource,
		Status:        status,
		Result:        result,
		Timestamp:     time.Now().Unix(),
		CreatedAt:     time.Now(),
	}

	// Generate hash
	record.Hash = generateRecordHash(record)

	at.Records = append(at.Records, record)

	// Maintain max size (FIFO if exceeds)
	if len(at.Records) > at.MaxSize {
		at.Records = at.Records[1:]
	}

	return record.ID, nil
}

// GetRecords returns all audit records
func (at *AuditTrail) GetRecords() []AuditRecord {
	at.mu.RLock()
	defer at.mu.RUnlock()
	return at.Records
}

// GetRecordsByActor returns all records from a specific actor
func (at *AuditTrail) GetRecordsByActor(actor string) []AuditRecord {
	at.mu.RLock()
	defer at.mu.RUnlock()

	var filtered []AuditRecord
	for _, record := range at.Records {
		if record.Actor == actor {
			filtered = append(filtered, record)
		}
	}
	return filtered
}

// GetRecordsByResource returns all records for a specific resource
func (at *AuditTrail) GetRecordsByResource(resource string) []AuditRecord {
	at.mu.RLock()
	defer at.mu.RUnlock()

	var filtered []AuditRecord
	for _, record := range at.Records {
		if record.Resource == resource {
			filtered = append(filtered, record)
		}
	}
	return filtered
}

// GetRecordsByTimeRange returns records within a time range
func (at *AuditTrail) GetRecordsByTimeRange(startTime, endTime int64) []AuditRecord {
	at.mu.RLock()
	defer at.mu.RUnlock()

	var filtered []AuditRecord
	for _, record := range at.Records {
		if record.Timestamp >= startTime && record.Timestamp <= endTime {
			filtered = append(filtered, record)
		}
	}
	return filtered
}

// VerifyRecordIntegrity verifies a single record's hash
func (at *AuditTrail) VerifyRecordIntegrity(recordID string) (bool, error) {
	at.mu.RLock()
	defer at.mu.RUnlock()

	for _, record := range at.Records {
		if record.ID == recordID {
			expectedHash := generateRecordHash(record)
			return expectedHash == record.Hash, nil
		}
	}

	return false, fmt.Errorf("record not found")
}

// VerifyChainIntegrity verifies the entire chain integrity
func (at *AuditTrail) VerifyChainIntegrity() (bool, error) {
	at.mu.RLock()
	defer at.mu.RUnlock()

	for i, record := range at.Records {
		expectedHash := generateRecordHash(record)
		if expectedHash != record.Hash {
			return false, fmt.Errorf("integrity check failed at record %d", i)
		}
	}

	return true, nil
}

// generateRecordHash generates a hash for a record
func generateRecordHash(record AuditRecord) string {
	data := fmt.Sprintf("%s-%s-%s-%s-%d", record.ID, record.Action, record.Actor, record.Resource, record.Timestamp)
	return hashSHA256(data)
}

// generateTransactionID generates a unique transaction ID
func generateTransactionID() string {
	return fmt.Sprintf("TXN-%d", time.Now().UnixNano())
}

// generateID creates a unique ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// hashSHA256 creates a SHA256 hash
func hashSHA256(data string) string {
	import "crypto/sha256"
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}
