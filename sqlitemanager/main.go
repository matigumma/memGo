package sqlitemanager

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

// SQLiteManager - Corresponds to the Python SQLiteManager class
type SQLiteManager struct {
	db *sql.DB
}

// NewSQLiteManager creates a new SQLiteManager instance
func NewSQLiteManager(dbPath string) (*SQLiteManager, error) {
	if dbPath == "" {
		dbPath = ":memory:"
	}
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	sm := &SQLiteManager{db: db}
	if err := sm.migrateHistoryTable(); err != nil {
		// Non-fatal, but log it
		log.Printf("Error migrating history table: %v", err)
	}
	if err := sm.createHistoryTable(); err != nil {
		return nil, err
	}
	return sm, nil
}

func (sm *SQLiteManager) migrateHistoryTable() error {
	// Check if the table exists
	row := sm.db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='history'")
	var tableName string
	if err := row.Scan(&tableName); err != nil {
		if err == sql.ErrNoRows {
			return nil // Table doesn't exist, nothing to migrate
		}
		return fmt.Errorf("failed to check for history table: %w", err)
	}

	// Get the current schema
	rows, err := sm.db.Query("PRAGMA table_info(history)")
	if err != nil {
		return fmt.Errorf("failed to get history table schema: %w", err)
	}
	defer rows.Close()

	currentSchema := make(map[string]string)
	for rows.Next() {
		var cid int
		var name, dataType, notnull, pk string
		var dfltValue interface{}
		if err := rows.Scan(&cid, &name, &dataType, &notnull, &dfltValue, &pk); err != nil {
			return fmt.Errorf("failed to scan table info: %w", err)
		}
		currentSchema[name] = dataType
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("error reading table info rows: %w", err)
	}

	// Define the expected schema
	expectedSchema := map[string]string{
		"id":         "TEXT",
		"memory_id":  "TEXT",
		"old_memory": "TEXT",
		"new_memory": "TEXT",
		"new_value":  "TEXT",
		"event":      "TEXT",
		"created_at": "DATETIME",
		"updated_at": "DATETIME",
		"is_deleted": "INTEGER",
	}

	// Check if schemas are the same
	if !sm.schemaEquals(currentSchema, expectedSchema) {
		// Rename the old table
		_, err = sm.db.Exec("ALTER TABLE history RENAME TO old_history")
		if err != nil {
			return fmt.Errorf("failed to rename history table: %w", err)
		}

		// Create the new table
		if err := sm.createHistoryTable(); err != nil {
			return err
		}

		// Copy data from the old table to the new table
		_, err = sm.db.Exec(`
			INSERT INTO history (id, memory_id, old_memory, new_memory, new_value, event, created_at, updated_at, is_deleted)
			SELECT id, memory_id, prev_value, new_value, new_value, event, timestamp, timestamp, is_deleted
			FROM old_history
		`)
		if err != nil {
			return fmt.Errorf("failed to copy data during migration: %w", err)
		}

		// Drop the old table
		_, err = sm.db.Exec("DROP TABLE old_history")
		if err != nil {
			return fmt.Errorf("failed to drop old_history table: %w", err)
		}
	}

	return nil
}

func (sm *SQLiteManager) schemaEquals(s1, s2 map[string]string) bool {
	if len(s1) != len(s2) {
		return false
	}
	for k, v := range s1 {
		if s2[k] != v {
			return false
		}
	}
	return true
}

func (sm *SQLiteManager) createHistoryTable() error {
	_, err := sm.db.Exec(`
		CREATE TABLE IF NOT EXISTS history (
			id TEXT PRIMARY KEY,
			memory_id TEXT,
			old_memory TEXT,
			new_memory TEXT,
			new_value TEXT,
			event TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			is_deleted INTEGER
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create history table: %w", err)
	}
	return nil
}

// AddHistory adds a new history record
func (sm *SQLiteManager) AddHistory(memoryID string, oldMemory *string, newMemory string, event string, createdAt *string, updatedAt *string, isDeleted int) error {
	_, err := sm.db.Exec(`
		INSERT INTO history (id, memory_id, old_memory, new_memory, event, created_at, updated_at, is_deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, uuid.New().String(), memoryID, oldMemory, newMemory, event, createdAt, updatedAt, isDeleted)
	if err != nil {
		return fmt.Errorf("failed to insert history: %w", err)
	}
	return nil
}

// GetHistory retrieves the history for a given memory ID
func (sm *SQLiteManager) GetHistory(memoryID string) ([]map[string]interface{}, error) {
	rows, err := sm.db.Query(`
		SELECT id, memory_id, old_memory, new_memory, event, created_at, updated_at
		FROM history
		WHERE memory_id = ?
		ORDER BY updated_at ASC
	`, memoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to query history: %w", err)
	}
	defer rows.Close()

	var history []map[string]interface{}
	for rows.Next() {
		var id, memID, oldMem, newMem, evt string
		var createdAt, updatedAt *string
		if err := rows.Scan(&id, &memID, &oldMem, &newMem, &evt, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan history row: %w", err)
		}
		history = append(history, map[string]interface{}{
			"id":         id,
			"memory_id":  memID,
			"old_memory": oldMem,
			"new_memory": newMem,
			"event":      evt,
			"created_at": createdAt,
			"updated_at": updatedAt,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading history rows: %w", err)
	}
	return history, nil
}

// Reset drops the history table
func (sm *SQLiteManager) Reset() error {
	_, err := sm.db.Exec("DROP TABLE IF EXISTS history")
	if err != nil {
		return fmt.Errorf("failed to drop history table: %w", err)
	}
	return nil
}
