package tools

import (
    "bufio"
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"

    "github.com/robertpelloni/hermes-agent/pkg/toolregistry"
)

// OllamaRunNative sends a prompt directly to a local Ollama server via its HTTP API.
// It streams the response and returns the full concatenated text.
func init() {
    toolregistry.Global().Register(&toolregistry.Tool{
        Name:        "ollama_run_native",
        Description: "Run a prompt against a local Ollama model without invoking the external binary.",
        Category:    "llm",
        Parameters: map[string]any{
            "model":  map[string]string{"type": "string", "description": "Name of the Ollama model (e.g. llama3)"},
            "prompt": map[string]string{"type": "string", "description": "User prompt"},
        },
        Handler: ollamaRunNativeHandler,
        Native:  true,
    })
}

// request payload for Ollama's /api/chat endpoint (simplified for run)
type ollamaChatRequest struct {
    Model  string `json:"model"`
    Prompt string `json:"prompt"`
    Stream bool   `json:"stream"`
}

type ollamaChatChunk struct {
    Response string `json:"response"`
    Done     bool   `json:"done"`
}

func ollamaRunNativeHandler(args map[string]any, ctx map[string]any) (any, error) {
    modelRaw, ok := args["model"]
    if !ok {
        return nil, fmt.Errorf("missing 'model'")
    }
    model, ok := modelRaw.(string)
    if !ok || model == "" {
        return nil, fmt.Errorf("'model' must be a non‑empty string")
    }
    promptRaw, ok := args["prompt"]
    if !ok {
        return nil, fmt.Errorf("missing 'prompt'")
    }
    prompt, ok := promptRaw.(string)
    if !ok {
        return nil, fmt.Errorf("'prompt' must be a string")
    }

    // Build request JSON
    reqBody, err := json.Marshal(ollamaChatRequest{Model: model, Prompt: prompt, Stream: true})
    if err != nil {
        return nil, fmt.Errorf("json marshal: %w", err)
    }

    // Use a short‑lived context with timeout (30s default)
    cctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Ollama default address – can be overridden via env var if needed.
    baseURL := "http://127.0.0.1:11434"
    if env := "OLLAMA_HOST"; env != "" {
        // Not mandatory – placeholder for future extension.
    }
    endpoint := fmt.Sprintf("%s/api/chat", baseURL)
    request, err := http.NewRequestWithContext(cctx, http.MethodPost, endpoint, bytes.NewReader(reqBody))
    if err != nil {
        return nil, fmt.Errorf("create request: %w", err)
    }
    request.Header.Set("Content-Type", "application/json")

    resp, err := http.DefaultClient.Do(request)
    if err != nil {
        return nil, fmt.Errorf("ollama request failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        // Read body for a more helpful error message.
        b, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("ollama returned %d: %s", resp.StatusCode, string(b))
    }

    // Stream response line‑by‑line. Each line is a JSON object.
    scanner := bufio.NewScanner(resp.Body)
    var buf bytes.Buffer
    for scanner.Scan() {
        line := scanner.Bytes()
        if len(line) == 0 {
            continue
        }
        var chunk ollamaChatChunk
        if err := json.Unmarshal(line, &chunk); err != nil {
            // If we cannot unmarshal we treat it as raw text fallback.
            buf.Write(line)
            continue
        }
        buf.WriteString(chunk.Response)
        if chunk.Done {
            break
        }
    }
    if err := scanner.Err(); err != nil {
        return nil, fmt.Errorf("reading stream: %w", err)
    }
    return buf.String(), nil
}
