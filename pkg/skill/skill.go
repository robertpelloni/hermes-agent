package skill

import (
	"fmt"
	"sync"
)

// Skill represents a learned capability.
type Skill struct {
	Name        string
	Description string
	Code        string
	UsageCount  int
}

// Repository manages agent skills.
type Repository struct {
	mu    sync.RWMutex
	skills map[string]*Skill
}

// NewRepository creates a new skill repository.
func NewRepository() *Repository {
	return &Repository{skills: make(map[string]*Skill)}
}

// Register adds a new skill.
func (r *Repository) Register(s *Skill) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.skills[s.Name] = s
	fmt.Printf("[hermes:skill] Registered: %s\n", s.Name)
}

// Lookup finds a skill by name.
func (r *Repository) Lookup(name string) (*Skill, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.skills[name]
	return s, ok
}

// List returns all skills.
func (r *Repository) List() []*Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*Skill
	for _, s := range r.skills {
		out = append(out, s)
	}
	return out
}
