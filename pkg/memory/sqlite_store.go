package memory

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// MessageRecord represents a stored message.
type MessageRecord struct {
	ID        int64
	Role      string
	Content   string
	CreatedAt time.Time
}

// sqliteDB wraps a SQLite connection for persistent memory storage.
type sqliteDB struct {
	db *sql.DB
}

// initSQLite opens (or creates) the SQLite memory database.
func initSQLite() *sqliteDB {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".hermes", "memory.db")

	os.MkdirAll(filepath.Dir(path), 0755)
	db, err := sql.Open("sqlite", path+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil
	}

	s := &sqliteDB{db: db}
	if err := s.ensureSchema(); err != nil {
		db.Close()
		return nil
	}
	return s
}

func (s *sqliteDB) ensureSchema() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL,
			role TEXT NOT NULL,
			content TEXT NOT NULL,
			created_at DATETIME NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_messages_session ON messages(session_id);`,
	}
	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return fmt.Errorf("failed to create schema: %w", err)
		}
	}
	return nil
}

func (s *sqliteDB) saveMessage(sessionID, role, content string) error {
	_, err := s.db.Exec(`INSERT INTO messages (session_id, role, content, created_at) VALUES (?, ?, ?, ?)`,
		sessionID, role, content, time.Now().UTC())
	return err
}

func (s *sqliteDB) loadMessages(sessionID string) ([]MessageRecord, error) {
	rows, err := s.db.Query(`SELECT id, role, content, created_at FROM messages WHERE session_id = ? ORDER BY id ASC`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []MessageRecord
	for rows.Next() {
		var m MessageRecord
		if err := rows.Scan(&m.ID, &m.Role, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	return msgs, nil
}

func (s *sqliteDB) deleteSession(sessionID string) error {
	_, err := s.db.Exec(`DELETE FROM messages WHERE session_id = ?`, sessionID)
	return err
}

func (s *sqliteDB) close() error {
	return s.db.Close()
}
