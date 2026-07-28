package sessionruntime

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/robertpelloni/hermes-agent/pkg/ai"
)

// Session represents a single conversation session.
type Session struct {
	ID        string       `json:"id"`
	CreatedAt time.Time    `json:"createdAt"`
	UpdatedAt time.Time    `json:"updatedAt"`
	Label     string       `json:"label,omitempty"`
	Messages  []ai.Message `json:"messages"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// Runtime manages sessions with optional file persistence.
type Runtime struct {
	mu          sync.RWMutex
	sessions    map[string]*Session
	sessionsDir string
	autoSave    bool
}

// NewRuntime creates a new session runtime.
func NewRuntime(sessionsDir string, autoSave bool) *Runtime {
	if sessionsDir == "" {
		home, _ := os.UserHomeDir()
		sessionsDir = filepath.Join(home, ".hermes", "sessions")
	}
	os.MkdirAll(sessionsDir, 0755)

	rt := &Runtime{
		sessions:    make(map[string]*Session),
		sessionsDir: sessionsDir,
		autoSave:    autoSave,
	}

	// Load existing sessions
	rt.loadSessions()

	return rt
}

func (r *Runtime) sessionPath(id string) string {
	return filepath.Join(r.sessionsDir, id+".json")
}

func (r *Runtime) loadSessions() {
	entries, err := os.ReadDir(r.sessionsDir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			id := strings.TrimSuffix(entry.Name(), ".json")
			data, err := os.ReadFile(r.sessionsDir + "/" + entry.Name())
			if err != nil {
				continue
			}
			var session Session
			if err := json.Unmarshal(data, &session); err == nil {
				r.sessions[id] = &session
			}
		}
	}
}

// Create creates a new session.
func (r *Runtime) Create(id string) (*Session, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.sessions[id]; exists {
		return nil, fmt.Errorf("session already exists: %s", id)
	}

	session := &Session{
		ID:        id,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Messages:  make([]ai.Message, 0),
		Metadata:  make(map[string]string),
	}

	r.sessions[id] = session

	if r.autoSave {
		r.saveSession(session)
	}

	return session, nil
}

// Get returns a session by ID.
func (r *Runtime) Get(id string) (*Session, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.sessions[id]
	return s, ok
}

// List returns all sessions sorted by updated time (most recent first).
func (r *Runtime) List() []*Session {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*Session, 0, len(r.sessions))
	for _, s := range r.sessions {
		result = append(result, s)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].UpdatedAt.After(result[j].UpdatedAt)
	})
	return result
}

// AppendMessage adds a message to a session.
func (r *Runtime) AppendMessage(sessionID string, msg ai.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	session, ok := r.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	session.Messages = append(session.Messages, msg)
	session.UpdatedAt = time.Now()

	if r.autoSave {
		r.saveSession(session)
	}

	return nil
}

// Truncate removes messages from a session to reduce context size.
func (r *Runtime) Truncate(sessionID string, keep int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	session, ok := r.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	if keep >= len(session.Messages) {
		return nil
	}

	// Always keep the first message (system prompt)
	newMessages := make([]ai.Message, 0, keep+1)
	newMessages = append(newMessages, session.Messages[0])
	if len(session.Messages)-keep > 1 {
		newMessages = append(newMessages, session.Messages[len(session.Messages)-keep:]...)
	} else {
		newMessages = append(newMessages, session.Messages[1:]...)
	}
	session.Messages = newMessages
	session.UpdatedAt = time.Now()

	if r.autoSave {
		r.saveSession(session)
	}

	return nil
}

// Fork creates a new session starting from the Nth message of an existing one.
func (r *Runtime) Fork(sourceID, newID string, msgCount int) (*Session, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	source, ok := r.sessions[sourceID]
	if !ok {
		return nil, fmt.Errorf("source session not found: %s", sourceID)
	}

	if msgCount <= 0 {
		msgCount = len(source.Messages)
	}

	msgs := make([]ai.Message, msgCount)
	copy(msgs, source.Messages[:msgCount])

	newSession := &Session{
		ID:        newID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Label:     fmt.Sprintf("forked from %s", sourceID),
		Messages:  msgs,
		Metadata: map[string]string{
			"forked_from": sourceID,
		},
	}

	r.sessions[newID] = newSession

	if r.autoSave {
		r.saveSession(newSession)
	}

	return newSession, nil
}

// Delete removes a session.
func (r *Runtime) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.sessions[id]; !ok {
		return fmt.Errorf("session not found: %s", id)
	}

	delete(r.sessions, id)

	// Remove file
	path := r.sessionPath(id)
	os.Remove(path)

	return nil
}

// Search finds messages matching a query across all sessions.
func (r *Runtime) Search(query string, limit int) []struct {
	SessionID string
	Message   ai.Message
} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	results := make([]struct {
		SessionID string
		Message   ai.Message
	}, 0)
	q := strings.ToLower(query)

	for sid, session := range r.sessions {
		for _, msg := range session.Messages {
			for _, content := range msg.Content {
				if strings.Contains(strings.ToLower(content.Text), q) {
					results = append(results, struct {
						SessionID string
						Message   ai.Message
					}{SessionID: sid, Message: msg})
					break
				}
			}
			if len(results) >= limit {
				return results
			}
		}
	}

	return results
}

func (r *Runtime) saveSession(session *Session) {
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return
	}
	path := r.sessionPath(session.ID)
	os.WriteFile(path, data, 0644)
}
