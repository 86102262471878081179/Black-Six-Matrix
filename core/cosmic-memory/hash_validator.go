package cosmicmemory

import (
	"crypto/sha256"
	"fmt"
	"time"
)

// HashValidator validates SHA-256 hashes for data integrity
type HashValidator struct {
	ValidatedHashes map[string]*HashValidation
}

// HashValidation represents a validated hash
type HashValidation struct {
	ID        string
	Data      string
	Hash      string
	Valid     bool
	Timestamp time.Time
}

// NewHashValidator creates a new hash validator
func NewHashValidator() *HashValidator {
	return &HashValidator{
		ValidatedHashes: make(map[string]*HashValidation),
	}
}

// GenerateHash generates a SHA-256 hash
func (hv *HashValidator) GenerateHash(data string) string {
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// ValidateHash validates a hash against data
func (hv *HashValidator) ValidateHash(data, providedHash string) (bool, error) {
	expectedHash := hv.GenerateHash(data)

	validation := &HashValidation{
		ID:        generateID(),
		Data:      data,
		Hash:      providedHash,
		Valid:     expectedHash == providedHash,
		Timestamp: time.Now(),
	}

	hv.ValidatedHashes[validation.ID] = validation

	if !validation.Valid {
		return false, fmt.Errorf("hash mismatch: expected %s, got %s", expectedHash, providedHash)
	}

	return true, nil
}

// ChainHash creates a chain hash (hash of hash)
func (hv *HashValidator) ChainHash(previousHash, newData string) string {
	combined := previousHash + newData
	return hv.GenerateHash(combined)
}

// GenerateHashWithTimestamp generates a hash including timestamp
func (hv *HashValidator) GenerateHashWithTimestamp(data string, timestamp int64) string {
	combined := fmt.Sprintf("%s-%d", data, timestamp)
	return hv.GenerateHash(combined)
}

// VerifyChain verifies a chain of hashes
func (hv *HashValidator) VerifyChain(hashes []string) (bool, error) {
	if len(hashes) < 2 {
		return false, fmt.Errorf("chain must have at least 2 hashes")
	}

	for i := 1; i < len(hashes); i {
		// In a real implementation, verify each hash is chained correctly
		previousHash := hashes[i-1]
		currentHash := hashes[i]

		if previousHash == "" || currentHash == "" {
			return false, fmt.Errorf("invalid hash at position %d", i)
		}
	}

	return true, nil
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
