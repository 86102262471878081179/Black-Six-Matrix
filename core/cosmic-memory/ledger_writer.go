package cosmicmemory

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// LedgerWriter handles write operations to the WORM ledger
type LedgerWriter struct {
	mu   sync.Mutex
	db   *sql.DB
	path string
}

// LedgerEntry represents an entry in the ledger
type LedgerEntry struct {
	ID        string
	Command   string
	Target    string
	Timestamp int64
	DataHash  string
	Actor     string
	Status    string
	CreatedAt time.Time
}

// NewLedgerWriter creates a new ledger writer
func NewLedgerWriter(dbPath string) (*LedgerWriter, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// Create table if not exists
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS ledger (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			command TEXT NOT NULL,
			target TEXT NOT NULL,
			timestamp INTEGER NOT NULL,
			data_hash TEXT NOT NULL UNIQUE,
			actor TEXT,
			status TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`

	if _, err := db.Exec(createTableSQL); err != nil {
		return nil, fmt.Errorf("failed to create table: %v", err)
	}

	return &LedgerWriter{
		db:   db,
		path: dbPath,
	}, nil
}

// Write appends an entry to the WORM ledger (Append-Only)
func (lw *LedgerWriter) Write(entry *LedgerEntry) error {
	lw.mu.Lock()
	defer lw.mu.Unlock()

	// Calculate SHA256 hash of the data
	data := fmt.Sprintf("%s-%s-%d", entry.Command, entry.Target, entry.Timestamp)
	hash := sha256.Sum256([]byte(data))
	entry.DataHash = fmt.Sprintf("%x", hash)

	// Insert into database (append-only)
	insertSQL := `
		INSERT INTO ledger (command, target, timestamp, data_hash, actor, status)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := lw.db.Exec(insertSQL,
		entry.Command,
		entry.Target,
		entry.Timestamp,
		entry.DataHash,
		entry.Actor,
		entry.Status,
	)

	if err != nil {
		return fmt.Errorf("failed to write to ledger: %v", err)
	}

	id, err := result.LastInsertRowid()
	if err != nil {
		return fmt.Errorf("failed to get insert ID: %v", err)
	}

	entry.ID = fmt.Sprintf("%d", id)
	entry.CreatedAt = time.Now()

	return nil
}

// ReadAll reads all entries from the ledger
func (lw *LedgerWriter) ReadAll() ([]LedgerEntry, error) {
	lw.mu.Lock()
	defer lw.mu.Unlock()

	query := "SELECT command, target, timestamp, data_hash, actor, status FROM ledger ORDER BY id ASC"
	rows, err := lw.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query ledger: %v", err)
	}
	defer rows.Close()

	var entries []LedgerEntry
	for rows.Next() {
		var entry LedgerEntry
		if err := rows.Scan(&entry.Command, &entry.Target, &entry.Timestamp, &entry.DataHash, &entry.Actor, &entry.Status); err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// VerifyIntegrity verifies the integrity of the ledger
func (lw *LedgerWriter) VerifyIntegrity() (bool, error) {
	entries, err := lw.ReadAll()
	if err != nil {
		return false, err
	}

	// Verify each entry's hash
	for _, entry := range entries {
		data := fmt.Sprintf("%s-%s-%d", entry.Command, entry.Target, entry.Timestamp)
		hash := sha256.Sum256([]byte(data))
		expectedHash := fmt.Sprintf("%x", hash)

		if expectedHash != entry.DataHash {
			return false, fmt.Errorf("integrity check failed for entry %s", entry.ID)
		}
	}

	return true, nil
}

// Close closes the database connection
func (lw *LedgerWriter) Close() error {
	return lw.db.Close()
}
