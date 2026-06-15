package ai

// Api represents the type of API interface for a provider.
type Api string

const (
	ApiOpenAICompletions     Api = "openai-completions"
	ApiOpenAIResponses       Api = "openai-responses"
	ApiAnthropicMessages     Api = "anthropic-messages"
	ApiGoogleGenerativeAI    Api = "google-generative-ai"
	ApiMistralConversations  Api = "mistral-conversations"
	ApiBedrockConverseStream Api = "bedrock-converse-stream"
)

// Provider represents the service provider offering the model.
type Provider string

const (
	ProviderOpenAI      Provider = "openai"
	ProviderAnthropic   Provider = "anthropic"
	ProviderGoogle      Provider = "google"
	ProviderGroq        Provider = "groq"
	ProviderCerebras    Provider = "cerebras"
	ProviderXAI         Provider = "xai"
	ProviderMistral     Provider = "mistral"
	ProviderOpenRouter  Provider = "openrouter"
	ProviderGithubCopilot  Provider = "github-copilot"
	ProviderMinimax     Provider = "minimax"
	ProviderOllama      Provider = "ollama"
	ProviderDeepSeek    Provider = "deepseek"
	ProviderZAI         Provider = "zai"
	ProviderVertex      Provider = "google-vertex"
)

// ThinkingLevel specifies the amount of reasoning effort.
type ThinkingLevel string

const (
	ThinkingOff   ThinkingLevel = "off"
	ThinkingLow   ThinkingLevel = "low"
	ThinkingMedium ThinkingLevel = "medium"
	ThinkingHigh  ThinkingLevel = "high"
)

// CacheRetention specifies prompt cache duration.
type CacheRetention string

const (
	CacheRetentionNone  CacheRetention = "none"
	CacheRetentionShort CacheRetention = "short"
	CacheRetentionLong  CacheRetention = "long"
)

// ContentType indicates the type of a content block.
type ContentType string

const (
	ContentTypeText     ContentType = "text"
	ContentTypeThinking ContentType = "thinking"
	ContentTypeImage    ContentType = "image"
	ContentTypeToolCall ContentType = "toolCall"
)

// Content is a block of content within a message.
type Content struct {
	Type    ContentType `json:"type"`
	Text    string      `json:"text,omitempty"`
	Image   string      `json:"image,omitempty"`   // Base64-encoded image data
	MimeType string     `json:"mimeType,omitempty"` // e.g. "image/png"
	Thinking string     `json:"thinking,omitempty"`
	Redacted bool       `json:"redacted,omitempty"`

	// Tool call fields
	ToolCallID string         `json:"toolCallId,omitempty"`
	Name       string         `json:"name,omitempty"`
	Arguments  map[string]any `json:"arguments,omitempty"`
}

// MessageRole represents the role of a message sender.
type MessageRole string

const (
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
	RoleTool      MessageRole = "tool"
)

// Message represents a single message in a conversation.
type Message struct {
	Role       MessageRole `json:"role"`
	Content    []Content   `json:"content"`
	Timestamp  int64       `json:"timestamp"`
	ToolCallID string      `json:"toolCallId,omitempty"`
	ToolName   string      `json:"toolName,omitempty"`
	IsError    bool        `json:"isError,omitempty"`
}

// Usage tracks token consumption for an API call.
type Usage struct {
	Input       int       `json:"input"`
	Output      int       `json:"output"`
	CacheRead   int       `json:"cacheRead"`
	CacheWrite  int       `json:"cacheWrite"`
	TotalTokens int       `json:"totalTokens"`
	Cost        UsageCost `json:"cost"`
}

// UsageCost tracks the monetary cost of an API call.
type UsageCost struct {
	Input      float64 `json:"input"`
	Output     float64 `json:"output"`
	CacheRead  float64 `json:"cacheRead"`
	CacheWrite float64 `json:"cacheWrite"`
	Total      float64 `json:"total"`
}

// StopReason indicates why generation stopped.
type StopReason string

const (
	StopReasonStop    StopReason = "stop"
	StopReasonLength  StopReason = "length"
	StopReasonToolUse StopReason = "toolUse"
	StopReasonError   StopReason = "error"
	StopReasonAborted StopReason = "aborted"
)

// CostConfig defines pricing per million tokens for a model.
type CostConfig struct {
	Input      float64 `json:"input"`
	Output     float64 `json:"output"`
	CacheRead  float64 `json:"cacheRead"`
	CacheWrite float64 `json:"cacheWrite"`
}

// ModelCapabilities describes what a model can do.
type ModelCapabilities struct {
	Vision    bool `json:"vision"`
	Thinking  bool `json:"thinking"`
	ToolUse   bool `json:"toolUse"`
	Streaming bool `json:"streaming"`
}

// ModelInfo describes a specific AI model.
type ModelInfo struct {
	ID              string            `json:"id"`   // e.g. "gpt-4o-mini"
	Name            string            `json:"name"` // e.g. "GPT-4o Mini"
	Provider        Provider          `json:"provider"`
	ProviderName    string            `json:"providerName"` // human-readable
	Api             Api               `json:"api"`
	BaseURL         string            `json:"baseUrl,omitempty"`
	ContextWindow   int               `json:"contextWindow"`
	MaxOutput       int               `json:"maxOutput"`
	Cost            CostConfig        `json:"cost"`
	Capabilities    ModelCapabilities `json:"capabilities"`
	DeprecationDate string            `json:"deprecationDate,omitempty"`
}

// Context represents a serializable conversation context.
type Context struct {
	SystemPrompt string    `json:"systemPrompt,omitempty"`
	Messages     []Message `json:"messages"`
}

// ToolParameter describes a parameter of a tool definition.
type ToolParameter struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

// ToolDefinition describes a callable tool for the LLM.
type ToolDefinition struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  []ToolParameter `json:"parameters,omitempty"`
}

// StreamEvent represents a single event during streaming.
type StreamEvent struct {
	Type         string         `json:"type"` // start, text_delta, text_end, toolcall_start, toolcall_delta, toolcall_end, done, error
	Delta        string        `json:"delta,omitempty"`
	ContentIndex *int          `json:"contentIndex,omitempty"`
	ToolCall     *ToolCallData `json:"toolCall,omitempty"`
	Message      *Message      `json:"message,omitempty"`
	Reason       string        `json:"reason,omitempty"`
	Error        string        `json:"error,omitempty"`
}

// ToolCallData contains data about a tool call during streaming.
type ToolCallData struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

// StreamResult contains the final result of a streaming call.
type StreamResult struct {
	Message   Message    `json:"message"`
	Usage     Usage      `json:"usage"`
	StopReason StopReason `json:"stopReason"`
}
