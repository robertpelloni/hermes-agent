package memory

import "fmt"

// Save persists a value to the SQLite store for a given session.
func (s *Store) Save(sessionID, key, value string) error {
	if s.sqliteDB != nil {
		_, err := s.sqliteDB.db.Exec(`INSERT INTO session_state (session_id, key, value, created_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP)
			ON CONFLICT(session_id, key) DO UPDATE SET value=excluded.value, created_at=CURRENT_TIMESTAMP`,
			sessionID, key, value)
		if err == nil {
			fmt.Printf("[hermes:memory] Saved state for %s: %s\n", sessionID, key)
		}
		return err
	}
	// Fallback to in-memory if no DB
	s.Store(fmt.Sprintf("%s:%s", sessionID, key), value)
	return nil
}

// Load retrieves a value from the SQLite store for a given session.
func (s *Store) Load(sessionID, key string) (string, error) {
	if s.sqliteDB != nil {
		var value string
		err := s.sqliteDB.db.QueryRow(`SELECT value FROM session_state WHERE session_id = ? AND key = ?`, sessionID, key).Scan(&value)
		if err != nil {
			return "", err
		}
		return value, nil
	}
	// Fallback to in-memory
	val, ok := s.Recall(fmt.Sprintf("%s:%s", sessionID, key))
	if !ok {
		return "", fmt.Errorf("not found in memory")
	}
	return val, nil
}

// Prune removes session state older than a certain duration (e.g. for cleanup).
func (s *Store) Prune(sessionID string) error {
	if s.sqliteDB != nil {
		_, err := s.sqliteDB.db.Exec(`DELETE FROM session_state WHERE session_id = ?`, sessionID)
		if err == nil {
			fmt.Printf("[hermes:memory] Pruned state for %s\n", sessionID)
		}
		return err
	}
	return nil
}
