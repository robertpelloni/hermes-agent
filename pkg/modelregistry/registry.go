package modelregistry

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/robertpelloni/hermes-agent/pkg/ai"
)

// ProviderInfo describes a provider and how to authenticate with it.
type ProviderInfo struct {
	Name          string   `json:"name"`
	EnvVars       []string `json:"envVars"`
	Label         string   `json:"label"`
	DocsURL       string   `json:"docsUrl"`
	OAuthSupported bool    `json:"oauthSupported"`
}

// ModelRegistry manages available models and API key resolution.
type ModelRegistry struct {
	mu        sync.RWMutex
	providers map[string]ProviderInfo
	models    []ai.ModelInfo
}

// NewModelRegistry creates a registry with default providers and models.
func NewModelRegistry() *ModelRegistry {
	r := &ModelRegistry{
		providers: make(map[string]ProviderInfo),
		models:    defaultModels(),
	}

	// Register known providers
	for _, p := range []ProviderInfo{
		{Name: "openai", EnvVars: []string{"OPENAI_API_KEY"}, Label: "OpenAI", DocsURL: "https://platform.openai.com/api-keys"},
		{Name: "anthropic", EnvVars: []string{"ANTHROPIC_API_KEY"}, Label: "Anthropic", DocsURL: "https://console.anthropic.com/settings/keys"},
		{Name: "google", EnvVars: []string{"GEMINI_API_KEY", "GOOGLE_API_KEY"}, Label: "Google", DocsURL: "https://aistudio.google.com/apikey"},
		{Name: "groq", EnvVars: []string{"GROQ_API_KEY"}, Label: "Groq", DocsURL: "https://console.groq.com/keys"},
		{Name: "xai", EnvVars: []string{"XAI_API_KEY"}, Label: "xAI", DocsURL: "https://console.x.ai/"},
		{Name: "mistral", EnvVars: []string{"MISTRAL_API_KEY"}, Label: "Mistral", DocsURL: "https://console.mistral.ai/"},
		{Name: "openrouter", EnvVars: []string{"OPENROUTER_API_KEY"}, Label: "OpenRouter", DocsURL: "https://openrouter.ai/keys"},
		{Name: "cerebras", EnvVars: []string{"CEREBRAS_API_KEY"}, Label: "Cerebras", DocsURL: "https://console.cerebras.ai/"},
		{Name: "deepseek", EnvVars: []string{"DEEPSEEK_API_KEY"}, Label: "DeepSeek", DocsURL: "https://platform.deepseek.com/"},
		{Name: "minimax", EnvVars: []string{"MINIMAX_API_KEY"}, Label: "MiniMax", DocsURL: "https://platform.minimaxi.com/"},
	} {
		r.RegisterProvider(p)
	}

	return r
}

// RegisterProvider adds a provider to the registry.
func (r *ModelRegistry) RegisterProvider(info ProviderInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[info.Name] = info
}

// GetProviderInfo returns information about a provider.
func (r *ModelRegistry) GetProviderInfo(name string) (ProviderInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	info, ok := r.providers[name]
	return info, ok
}

// HasAPIKey checks if an API key is available for a provider.
func (r *ModelRegistry) HasAPIKey(providerName string) bool {
	info, ok := r.GetProviderInfo(providerName)
	if !ok {
		return false
	}
	for _, envVar := range info.EnvVars {
		if os.Getenv(envVar) != "" {
			return true
		}
	}
	return false
}

// GetAPIKey returns the API key for a provider.
func (r *ModelRegistry) GetAPIKey(providerName string) (string, error) {
	info, ok := r.GetProviderInfo(providerName)
	if !ok {
		return "", fmt.Errorf("unknown provider: %s", providerName)
	}
	for _, envVar := range info.EnvVars {
		if key := os.Getenv(envVar); key != "" {
			return key, nil
		}
	}
	return "", fmt.Errorf("no API key found for %s. Set one of: %s", info.Label, strings.Join(info.EnvVars, ", "))
}

// HasAnyAPIKey returns true if at least one provider has an API key configured.
func (r *ModelRegistry) HasAnyAPIKey() bool {
	for _, p := range r.GetAllProviders() {
		if r.HasAPIKey(p.Name) {
			return true
		}
	}
	return false
}

// GetAllProviders returns all registered providers.
func (r *ModelRegistry) GetAllProviders() []ProviderInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]ProviderInfo, 0, len(r.providers))
	for _, info := range r.providers {
		result = append(result, info)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result
}

// GetAvailableProviders returns providers that have API keys configured.
func (r *ModelRegistry) GetAvailableProviders() []ProviderInfo {
	all := r.GetAllProviders()
	var available []ProviderInfo
	for _, info := range all {
		if r.HasAPIKey(info.Name) {
			available = append(available, info)
		}
	}
	return available
}

// GetAllModels returns all known models.
func (r *ModelRegistry) GetAllModels() []ai.ModelInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]ai.ModelInfo, len(r.models))
	copy(result, r.models)
	return result
}

// GetAvailableModels returns models for providers that have API keys.
func (r *ModelRegistry) GetAvailableModels() []ai.ModelInfo {
	all := r.GetAllModels()
	var available []ai.ModelInfo
	for _, model := range all {
		if r.HasAPIKey(string(model.Provider)) {
			available = append(available, model)
		}
	}
	return available
}

// SearchModels searches for models matching a pattern.
func (r *ModelRegistry) SearchModels(pattern string) []ai.ModelInfo {
	pattern = strings.ToLower(pattern)
	all := r.GetAllModels()
	var result []ai.ModelInfo
	for _, model := range all {
		if strings.Contains(strings.ToLower(model.ID), pattern) ||
			strings.Contains(strings.ToLower(model.Name), pattern) ||
			strings.Contains(strings.ToLower(string(model.Provider)), pattern) {
			result = append(result, model)
		}
	}
	return result
}

// FindModelByID finds a model by its ID.
func (r *ModelRegistry) FindModelByID(id string) (ai.ModelInfo, bool) {
	all := r.GetAllModels()
	for _, model := range all {
		if model.ID == id {
			return model, true
		}
	}
	return ai.ModelInfo{}, false
}

// ResolveProvider determines the provider from a model ID.
func (r *ModelRegistry) ResolveProvider(modelID string) (string, error) {
	// Check if model ID contains a provider prefix (e.g. "openai/gpt-4o")
	if idx := strings.Index(modelID, "/"); idx > 0 {
		providerName := modelID[:idx]
		if _, ok := r.GetProviderInfo(providerName); ok {
			return providerName, nil
		}
	}

	// Look up model in registry
	if model, ok := r.FindModelByID(modelID); ok {
		return string(model.Provider), nil
	}

	// Infer from model name patterns
	id := strings.ToLower(modelID)
	switch {
	case strings.Contains(id, "gpt"), strings.Contains(id, "o1"), strings.Contains(id, "o3"), strings.Contains(id, "o4"):
		return "openai", nil
	case strings.Contains(id, "claude"):
		return "anthropic", nil
	case strings.Contains(id, "gemini"):
		return "google", nil
	case strings.Contains(id, "llama"), strings.Contains(id, "mixtral"):
		return "groq", nil
	case strings.Contains(id, "grok"):
		return "xai", nil
	default:
		return "", fmt.Errorf("unable to resolve provider for model: %s", modelID)
	}
}

// FormatProviderStatus returns a human-readable summary of provider status.
func (r *ModelRegistry) FormatProviderStatus() string {
	var out strings.Builder
	out.WriteString(fmt.Sprintf("%-25s %-15s %s\n", "PROVIDER", "API KEY", "MODELS"))
	out.WriteString(fmt.Sprintf("%-25s %-15s %s\n", strings.Repeat("-", 25), strings.Repeat("-", 15), strings.Repeat("-", 20)))

	for _, p := range r.GetAllProviders() {
		hasKey := r.HasAPIKey(p.Name)
		keyStatus := "✓ configured"
		if !hasKey {
			keyStatus = "✗ missing"
		}
		count := len(r.GetModelsForProvider(p.Name))
		out.WriteString(fmt.Sprintf("%-25s %-15s %d\n", p.Label, keyStatus, count))
	}
	return out.String()
}

// GetModelsForProvider returns all models for a specific provider.
func (r *ModelRegistry) GetModelsForProvider(provider string) []ai.ModelInfo {
	all := r.GetAllModels()
	var result []ai.ModelInfo
	for _, model := range all {
		if string(model.Provider) == provider {
			result = append(result, model)
		}
	}
	return result
}

// defaultModels returns the default set of known models.
func defaultModels() []ai.ModelInfo {
	return []ai.ModelInfo{
		// OpenAI models
		{ID: "gpt-4o-mini", Name: "GPT-4o Mini", Provider: ai.ProviderOpenAI, ProviderName: "OpenAI",
			Api: ai.ApiOpenAICompletions, ContextWindow: 128000, MaxOutput: 16384,
			Cost: ai.CostConfig{Input: 0.15, Output: 0.60, CacheRead: 0.075, CacheWrite: 0.15},
			Capabilities: ai.ModelCapabilities{Vision: true, Thinking: false, ToolUse: true, Streaming: true}},
		{ID: "gpt-4o", Name: "GPT-4o", Provider: ai.ProviderOpenAI, ProviderName: "OpenAI",
			Api: ai.ApiOpenAICompletions, ContextWindow: 128000, MaxOutput: 16384,
			Cost: ai.CostConfig{Input: 2.50, Output: 10.00, CacheRead: 1.25, CacheWrite: 2.50},
			Capabilities: ai.ModelCapabilities{Vision: true, Thinking: false, ToolUse: true, Streaming: true}},
		{ID: "gpt-5-mini", Name: "GPT-5 Mini", Provider: ai.ProviderOpenAI, ProviderName: "OpenAI",
			Api: ai.ApiOpenAICompletions, ContextWindow: 131072, MaxOutput: 32768,
			Cost: ai.CostConfig{Input: 0.33, Output: 1.33, CacheRead: 1.25, CacheWrite: 2.50},
			Capabilities: ai.ModelCapabilities{Vision: true, Thinking: true, ToolUse: true, Streaming: true}},
		{ID: "o4-mini", Name: "o4 Mini", Provider: ai.ProviderOpenAI, ProviderName: "OpenAI",
			Api: ai.ApiOpenAICompletions, ContextWindow: 200000, MaxOutput: 100000,
			Cost: ai.CostConfig{Input: 1.10, Output: 4.40, CacheRead: 1.25, CacheWrite: 2.50},
			Capabilities: ai.ModelCapabilities{Vision: true, Thinking: true, ToolUse: true, Streaming: true}},

		// Anthropic models
		{ID: "claude-sonnet-4-20250514", Name: "Claude Sonnet 4", Provider: ai.ProviderAnthropic, ProviderName: "Anthropic",
			Api: ai.ApiAnthropicMessages, ContextWindow: 200000, MaxOutput: 8192,
			Cost: ai.CostConfig{Input: 3.00, Output: 15.00, CacheRead: 0.30, CacheWrite: 3.75},
			Capabilities: ai.ModelCapabilities{Vision: true, Thinking: true, ToolUse: true, Streaming: true}},
		{ID: "claude-3-5-haiku-20241022", Name: "Claude 3.5 Haiku", Provider: ai.ProviderAnthropic, ProviderName: "Anthropic",
			Api: ai.ApiAnthropicMessages, ContextWindow: 200000, MaxOutput: 8192,
			Cost: ai.CostConfig{Input: 0.80, Output: 4.00, CacheRead: 0.08, CacheWrite: 1.00},
			Capabilities: ai.ModelCapabilities{Vision: true, Thinking: false, ToolUse: true, Streaming: true}},

		// Google models
		{ID: "gemini-2.5-flash", Name: "Gemini 2.5 Flash", Provider: ai.ProviderGoogle, ProviderName: "Google",
			Api: ai.ApiGoogleGenerativeAI, ContextWindow: 1048576, MaxOutput: 8192,
			Cost: ai.CostConfig{Input: 0.15, Output: 0.60, CacheRead: 0.025, CacheWrite: 0.15},
			Capabilities: ai.ModelCapabilities{Vision: true, Thinking: true, ToolUse: true, Streaming: true}},
		{ID: "gemini-2.5-pro", Name: "Gemini 2.5 Pro", Provider: ai.ProviderGoogle, ProviderName: "Google",
			Api: ai.ApiGoogleGenerativeAI, ContextWindow: 1048576, MaxOutput: 8192,
			Cost: ai.CostConfig{Input: 1.25, Output: 5.00, CacheRead: 0.25, CacheWrite: 1.25},
			Capabilities: ai.ModelCapabilities{Vision: true, Thinking: true, ToolUse: true, Streaming: true}},

		// Groq models
		{ID: "deepseek-r1-distill-llama-70b", Name: "DeepSeek R1 Distill Llama 70B", Provider: ai.ProviderGroq, ProviderName: "Groq",
			Api: ai.ApiOpenAICompletions, ContextWindow: 131072, MaxOutput: 16384,
			Cost: ai.CostConfig{Input: 0.75, Output: 0.99},
			Capabilities: ai.ModelCapabilities{Vision: false, Thinking: true, ToolUse: true, Streaming: true}},
		{ID: "groq-llama-3.3-70b-versatile", Name: "Llama 3.3 70B Versatile", Provider: ai.ProviderGroq, ProviderName: "Groq",
			Api: ai.ApiOpenAICompletions, ContextWindow: 131072, MaxOutput: 32768,
			Cost: ai.CostConfig{Input: 0.59, Output: 0.79},
			Capabilities: ai.ModelCapabilities{Vision: false, Thinking: false, ToolUse: true, Streaming: true}},

		// xAI models
		{ID: "grok-code-fast-1", Name: "Grok Code Fast 1", Provider: ai.ProviderXAI, ProviderName: "xAI",
			Api: ai.ApiOpenAICompletions, ContextWindow: 131072, MaxOutput: 16384,
			Cost: ai.CostConfig{Input: 0.50, Output: 0.50},
			Capabilities: ai.ModelCapabilities{Vision: false, Thinking: false, ToolUse: true, Streaming: true}},

		// OpenRouter models
		{ID: "openai/gpt-4o-mini", Name: "GPT-4o Mini (OpenRouter)", Provider: ai.ProviderOpenRouter, ProviderName: "OpenRouter",
			Api: ai.ApiOpenAICompletions, ContextWindow: 128000, MaxOutput: 16384,
			Cost: ai.CostConfig{Input: 0.15, Output: 0.60},
			Capabilities: ai.ModelCapabilities{Vision: true, Thinking: false, ToolUse: true, Streaming: true}},
		{ID: "openai/gpt-5-mini", Name: "GPT-5 Mini (OpenRouter)", Provider: ai.ProviderOpenRouter, ProviderName: "OpenRouter",
			Api: ai.ApiOpenAICompletions, ContextWindow: 131072, MaxOutput: 32768,
			Cost: ai.CostConfig{Input: 0.44, Output: 1.77},
			Capabilities: ai.ModelCapabilities{Vision: true, Thinking: true, ToolUse: true, Streaming: true}},

		// DeepSeek models
		{ID: "deepseek-chat", Name: "DeepSeek V3", Provider: ai.ProviderDeepSeek, ProviderName: "DeepSeek",
			Api: ai.ApiOpenAICompletions, ContextWindow: 131072, MaxOutput: 8192,
			Cost: ai.CostConfig{Input: 0.27, Output: 1.10},
			Capabilities: ai.ModelCapabilities{Vision: false, Thinking: false, ToolUse: true, Streaming: true}},
		{ID: "deepseek-reasoner", Name: "DeepSeek R1", Provider: ai.ProviderDeepSeek, ProviderName: "DeepSeek",
			Api: ai.ApiOpenAICompletions, ContextWindow: 131072, MaxOutput: 8192,
			Cost: ai.CostConfig{Input: 0.55, Output: 2.19},
			Capabilities: ai.ModelCapabilities{Vision: false, Thinking: true, ToolUse: true, Streaming: true}},
	}
}
