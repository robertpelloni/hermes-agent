package skill

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Skill represents a learned capability or tool specification.
type Skill struct {
	Name        string
	Description string
	Code        string
	Path        string
	UsageCount  int
}

// Repository manages agent skills.
type Repository struct {
	mu     sync.RWMutex
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
	out := make([]*Skill, 0, len(r.skills))
	for _, s := range r.skills {
		out = append(out, s)
	}
	return out
}

// LoadFromDir recursively scans a directory for skill files (SKILL.md or .py files).
// It parses metadata blocks from SKILL.md or docstrings from .py files.
func (r *Repository) LoadFromDir(dir string) (int, error) {
	if _, err := os.Stat(dir); err != nil {
		return 0, err
	}

	loaded := 0
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}

		filename := d.Name()
		if filename == "SKILL.md" {
			s, err := parseSkillMd(path)
			if err == nil && s != nil && s.Name != "" {
				r.Register(s)
				loaded++
			}
		} else if strings.HasSuffix(filename, ".py") && !strings.HasPrefix(filename, "_") && !strings.HasPrefix(filename, "test_") {
			s, err := parseSkillPy(path)
			if err == nil && s != nil && s.Name != "" {
				r.Register(s)
				loaded++
			}
		}

		return nil
	})

	return loaded, err
}

func parseSkillMd(path string) (*Skill, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var inFrontmatter bool
	var name, description string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "---" {
			if !inFrontmatter {
				inFrontmatter = true
				continue
			} else {
				break
			}
		}
		if inFrontmatter {
			if strings.HasPrefix(line, "name:") {
				name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
				name = strings.Trim(name, `"'`)
			} else if strings.HasPrefix(line, "description:") {
				description = strings.TrimSpace(strings.TrimPrefix(line, "description:"))
				description = strings.Trim(description, `"'`)
			}
		}
	}

	if name == "" {
		parent := filepath.Base(filepath.Dir(path))
		if parent != "." && parent != "/" {
			name = parent
		} else {
			name = strings.TrimSuffix(filepath.Base(path), ".md")
		}
	}

	return &Skill{
		Name:        name,
		Description: description,
		Path:        path,
	}, nil
}

func parseSkillPy(path string) (*Skill, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	name := strings.TrimSuffix(filepath.Base(path), ".py")
	desc := fmt.Sprintf("Python skill: %s", name)

	text := string(content)
	if strings.Contains(text, `"""`) {
		parts := strings.Split(text, `"""`)
		if len(parts) >= 3 {
			doc := strings.TrimSpace(parts[1])
			if doc != "" {
				lines := strings.Split(doc, "\n")
				desc = strings.TrimSpace(lines[0])
			}
		}
	}

	return &Skill{
		Name:        name,
		Description: desc,
		Code:        text,
		Path:        path,
	}, nil
}
