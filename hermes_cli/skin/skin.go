package skin

import (
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"
    "strings"

    "gopkg.in/yaml.v3"
)

// Skin mirrors the structure used by the Python skin engine (minimal fields we need).
type Skin struct {
    Name   string `yaml:"name"`
    Colors struct {
        BannerBorder   string `yaml:"banner_border,omitempty"`
        BannerTitle    string `yaml:"banner_title,omitempty"`
        BannerAccent   string `yaml:"banner_accent,omitempty"`
        BannerDim      string `yaml:"banner_dim,omitempty"`
        BannerText     string `yaml:"banner_text,omitempty"`
        ResponseBorder string `yaml:"response_border,omitempty"`
    } `yaml:"colors,omitempty"`
    Spinner struct {
        ThinkingFaces []string `yaml:"thinking_faces,omitempty"`
        WaitingFaces  []string `yaml:"waiting_faces,omitempty"`
        Verbs        []string `yaml:"thinking_verbs,omitempty"`
        Wings        []string `yaml:"wings,omitempty"`
        Style        string   `yaml:"style,omitempty"`
    } `yaml:"spinner,omitempty"`
    ToolPrefix string `yaml:"tool_prefix,omitempty"`
    Branding   struct {
        AgentName      string `yaml:"agent_name,omitempty"`
        ResponseLabel  string `yaml:"response_label,omitempty"`
        PromptSymbol   string `yaml:"prompt_symbol,omitempty"`
        WelcomeMessage string `yaml:"welcome,omitempty"`
    } `yaml:"branding,omitempty"`
}

// builtinSkins holds a few minimal built‑in skins expressed as YAML strings.
var builtinSkins = map[string]string{
    "default": `name: default
colors:
  response_border: "#FFD700"
spinner:
  style: "#ffcc00"
  thinking_faces: ["⠋","⠙","⠹","⠸","⠼","⠴","⠦","⠧","⠇","⠏"]
  waiting_faces: []
  thinking_verbs: []
  wings: []
  verbs: []
tool_prefix: "▊"`,
    "mono": `name: mono
colors:
  response_border: "#CCCCCC"
spinner:
  style: "#AAAAAA"
  thinking_faces: ["|","/","-","\\"]
tool_prefix: "#"`,
    "slate": `name: slate
colors:
  response_border: "#5F9EA0"
spinner:
  style: "#5F9EA0"
  thinking_faces: ["⣾","⣽","⣻","⢿","⡿","⣟","⣯","⣷"]
tool_prefix: "•"`,
}

// Load returns a Skin by name. It first checks user‑installed skins under
// $HERMES_HOME/skins/<name>.yaml, then falls back to the built‑in collection.
func Load(name string) (*Skin, error) {
    name = strings.TrimSpace(name)
    // 1. User‑installed skin.
    if home := os.Getenv("HERMES_HOME"); home != "" {
        userPath := filepath.Join(home, "skins", name+".yaml")
        if data, err := ioutil.ReadFile(userPath); err == nil {
            var s Skin
            if err := yaml.Unmarshal(data, &s); err != nil {
                return nil, fmt.Errorf("invalid user skin %s: %w", userPath, err)
            }
            if s.Name == "" {
                s.Name = name
            }
            return &s, nil
        }
    }
    // 2. Built‑in skin.
    if yamlStr, ok := builtinSkins[name]; ok {
        var s Skin
        if err := yaml.Unmarshal([]byte(yamlStr), &s); err != nil {
            return nil, fmt.Errorf("builtin skin decode failed: %w", err)
        }
        return &s, nil
    }
    return nil, fmt.Errorf("skin %q not found", name)
}

// DefaultName returns the name of the default built‑in skin.
func DefaultName() string { return "default" }
