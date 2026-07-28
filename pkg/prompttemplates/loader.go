package prompttemplates

import (
	"embed"
	"fmt"
	"strings"
)

//go:embed templates/*.txt
var templateFS embed.FS

// Load returns the system prompt text for the named template.
// If the template is not found, Load returns an error.
func Load(name string) (string, error) {
	data, err := templateFS.ReadFile(fmt.Sprintf("templates/%s.txt", name))
	if err != nil {
		return "", fmt.Errorf("prompt template %q not found: %w", name, err)
	}
	return strings.TrimSpace(string(data)), nil
}

// List returns all available template names.
func List() []string {
	entries, err := templateFS.ReadDir("templates")
	if err != nil {
		return nil
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		name := strings.TrimSuffix(e.Name(), ".txt")
		names = append(names, name)
	}
	return names
}

// Default returns the system prompt for the default template.
func Default() string {
	prompt, err := Load("default")
	if err != nil {
		return "You are Hermes, an intelligent AI assistant. Use tools when needed to help the user."
	}
	return prompt
}
