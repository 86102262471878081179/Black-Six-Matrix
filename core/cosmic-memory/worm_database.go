package cosmicmemory

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// WORMDatabase implements a Write-Once-Read-Many database
type WORMDatabase struct {
	mu   sync.RWMutex
	db   *sql.DB
	path string
}

// WORMRecord represents a WORM database record
type WORMRecord struct {
	ID        int64
	Key       string
	Value     string
	Hash      string
	Timestamp int64
	Written   bool
}

// NewWORMDatabase creates a new WORM database
func NewWORMDatabase(dbPath string) (*WORMDatabase, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// Create WORM table
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS worm_records (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			key TEXT NOT NULL UNIQUE,
			value TEXT NOT NULL,
			hash TEXT NOT NULL UNIQUE,
			timestamp INTEGER NOT NULL,
			written BOOLEAN DEFAULT true,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`

	if _, err := db.Exec(createTableSQL); err != nil {
		return nil, fmt.Errorf("failed to create table: %v", err)
	}

	return &WORMDatabase{
		db:   db,
		path: dbPath,
	}, nil
}

// Write writes a record to the WORM database (Append-Only)
func (wd *WORMDatabase) Write(key, value string) (int64, error) {
	wd.mu.Lock()
	defer wd.mu.Unlock()

	// Check if key already exists
	var exists bool
	err := wd.db.QueryRow("SELECT COUNT(*) > 0 FROM worm_records WHERE key = ?", key).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to check existence: %v", err)
	}

	if exists {
		return 0, fmt.Errorf("key already exists in WORM database (write-once)")
	}

	// Generate hash
	hash := generateWORMHash(key, value)
	timestamp := time.Now().Unix()

	// Insert record
	insertSQL := `
		INSERT INTO worm_records (key, value, hash, timestamp, written)
		VALUES (?, ?, ?, ?, ?)
	`

	result, err := wd.db.Exec(insertSQL, key, value, hash, timestamp, true)
	if err != nil {
		return 0, fmt.Errorf("failed to write to WORM: %v", err)
	}

	return result.LastInsertRowid()
}

// Read reads a record from the WORM database
func (wd *WORMDatabase) Read(key string) (*WORMRecord, error) {
	wd.mu.RLock()
	defer wd.mu.RUnlock()

	var record WORMRecord
	query := "SELECT id, key, value, hash, timestamp, written FROM worm_records WHERE key = ?"

	err := wd.db.QueryRow(query, key).Scan(
		&record.ID,
		&record.Key,
		&record.Value,
		&record.Hash,
		&record.Timestamp,
		&record.Written,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to read from WORM: %v", err)
	}

	return &record, nil
}

// ReadAll reads all records from the WORM database
func (wd *WORMDatabase) ReadAll() ([]WORMRecord, error) {
	wd.mu.RLock()
	defer wd.mu.RUnlock()

	query := "SELECT id, key, value, hash, timestamp, written FROM worm_records ORDER BY id ASC"
	rows, err := wd.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query WORM: %v", err)
	}
	defer rows.Close()

	var records []WORMRecord
	for rows.Next() {
		var record WORMRecord
		if err := rows.Scan(&record.ID, &record.Key, &record.Value, &record.Hash, &record.Timestamp, &record.Written); err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
		records = append(records, record)
	}

	return records, nil
}

// VerifyIntegrity verifies the entire WORM database integrity
func (wd *WORMDatabase) VerifyIntegrity() (bool, error) {
	records, err := wd.ReadAll()
	if err != nil {
		return false, err
	}

	for _, record := range records {
		expectedHash := generateWORMHash(record.Key, record.Value)
		if expectedHash != record.Hash {
			return false, fmt.Errorf("integrity check failed for key %s", record.Key)
		}
	}

	return true, nil
}

// GetCount returns the total number of records
func (wd *WORMDatabase) GetCount() (int64, error) {
	wd.mu.RLock()
	defer wd.mu.RUnlock()

	var count int64
	err := wd.db.QueryRow("SELECT COUNT(*) FROM worm_records").Scan(&count)
	return count, err
}

// Close closes the database connection
func (wd *WORMDatabase) Close() error {
	return wd.db.Close()
}

// generateWORMHash generates a hash for a WORM record
func generateWORMHash(key, value string) string {
	data := fmt.Sprintf("%s-%s-%d", key, value, time.Now().Unix())
	var hash = fmt.Sprintf("%x", sha256(data))
	return hash
}

func sha256(data string) [32]byte {
	import "crypto/sha256"
	return sha256.Sum256([]byte(data))
}
