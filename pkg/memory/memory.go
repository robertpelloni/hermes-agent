package memory

import (
	"fmt"
	"sync"
)

// Store implements the Hermes memory store.
type Store struct {
	mu    sync.RWMutex
	items map[string]string
}

// NewStore creates a new memory store.
func NewStore() *Store {
	return &Store{items: make(map[string]string)}
}

// Store saves a memory item.
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
