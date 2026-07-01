package skill

import (
	"fmt"
	"sync"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// GlobalRepository is a singleton instance of Repository for convenient access.
var globalRepo *Repository
var initOnce sync.Once

// Global returns the singleton skill repository.
func Global() *Repository {
	initOnce.Do(func() {
		globalRepo = NewRepository()
	})
	return globalRepo
}

// ExtractDocstring reads the top-level docstring from a Python file.
func ExtractDocstring(content string) string {
	// Match triple double quotes
	reDouble := regexp.MustCompile(`(?s)^[\s]*"""(.*?)"""`)
	if match := reDouble.FindStringSubmatch(content); match != nil {
		return strings.TrimSpace(match[1])
	}

	// Match triple single quotes
	reSingle := regexp.MustCompile(`(?s)^[\s]*'''(.*?)'''`)
	if match := reSingle.FindStringSubmatch(content); match != nil {
		return strings.TrimSpace(match[1])
	}

	return ""
}

// DiscoverAndLoad scans a directory recursively for .py files and registers them as skills.
func (r *Repository) DiscoverAndLoad(root string) error {
	if _, err := os.Stat(root); os.IsNotExist(err) {
		return fmt.Errorf("skills directory not found: %s", root)
	}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // ignore unreadable dirs
		}

		if d.IsDir() {
			// Skip __pycache__ and other hidden dirs
			if strings.HasPrefix(d.Name(), ".") || d.Name() == "__pycache__" {
				return fs.SkipDir
			}
			return nil
		}

		if strings.HasSuffix(d.Name(), ".py") && d.Name() != "__init__.py" {
			// Found a python skill file
			contentBytes, readErr := os.ReadFile(path)
			if readErr != nil {
				return nil
			}

			content := string(contentBytes)
			desc := ExtractDocstring(content)
			if desc == "" {
				desc = "A dynamically loaded python skill."
			}

			// Skill name is the filename without .py
			name := strings.TrimSuffix(d.Name(), ".py")

			// Optional: if skills need unique namespace, could use relative path
			// rel, _ := filepath.Rel(root, path)
			// name = strings.ReplaceAll(strings.TrimSuffix(rel, ".py"), string(filepath.Separator), "_")

			skill := &Skill{
				Name:        name,
				Description: desc,
				Code:        content,
				UsageCount:  0,
			}
			r.Register(skill)
		}
		return nil
	})

	return err
}
