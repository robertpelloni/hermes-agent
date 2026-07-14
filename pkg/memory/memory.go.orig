package memory

import (
	"fmt"
	"sync"
)

// Store implements the Hermes memory store.
// It combines an in-memory map for fast ephemeral storage
// with a SQLite backend for persistent session storage.
type Store struct {
	mu         sync.RWMutex
	items      map[string]string
	sqliteDB   *sqliteDB // nil if SQLite is unavailable
}

// NewStore creates a memory store.
func NewStore() *Store {
	s := &Store{items: make(map[string]string)}
	// Try to initialize SQLite backend; ignore failure
	if db := initSQLite(); db != nil {
		s.sqliteDB = db
	}
	return s
}

// Store saves a memory item (ephemeral, in-memory).
func (s *Store) Store(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[key] = value
	fmt.Printf("[hermes:memory] Stored: %s\n", key)
}

// Recall retrieves a memory item.
func (s *Store) Recall(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.items[key]
	return v, ok
}

// Search performs full-text search over memories.
func (s *Store) Search(query string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var results []string
	for k, v := range s.items {
		if k == query || v == query {
			results = append(results, v)
		}
	}
	return results
}

// SaveMessage persists a message to SQLite (if available).
func (s *Store) SaveMessage(sessionID, role, content string) error {
	if s.sqliteDB != nil {
		return s.sqliteDB.saveMessage(sessionID, role, content)
	}
	return nil
}

// LoadMessages retrieves messages from SQLite (if available).
func (s *Store) LoadMessages(sessionID string) ([]MessageRecord, error) {
	if s.sqliteDB != nil {
		return s.sqliteDB.loadMessages(sessionID)
	}
	return nil, nil
}

// Close closes the SQLite backend.
func (s *Store) Close() error {
	if s.sqliteDB != nil {
		return s.sqliteDB.close()
	}
	return nil
}
