package tools

import (
    "context"
    "fmt"
    "time"

    "github.com/robertpelloni/hermes-agent/pkg/subcmd"
    "github.com/robertpelloni/hermes-agent/pkg/toolregistry"
)

func init() {
    toolregistry.Global().Register(&toolregistry.Tool{
        Name:        "adrenaline_chat",
        Description: "Thin wrapper for the Adrenaline CLI `chat` command (model‑agnostic).",
        Category:    "llm",
        Parameters: map[string]any{
            "prompt": map[string]string{"type": "string", "description": "User prompt"},
            "model":  map[string]string{"type": "string", "description": "Optional model identifier"},
        },
        Handler: adrenalineChatHandler,
        Native:  true,
    })
}

func adrenalineChatHandler(args map[string]any, ctx map[string]any) (any, error) {
    promptRaw, ok := args["prompt"]
    if !ok {
        return nil, fmt.Errorf("missing 'prompt'")
    }
    prompt, ok := promptRaw.(string)
    if !ok {
        return nil, fmt.Errorf("'prompt' must be a string")
    }
    // Build CLI arguments
    cliArgs := []string{"chat", prompt}
    if modelRaw, ok := args["model"]; ok {
        if model, ok := modelRaw.(string); ok && model != "" {
            cliArgs = append(cliArgs, "--model", model)
        }
    }
    out, err := subcmd.RunBinary(context.Background(), "adrenaline", cliArgs, 30*time.Second)
    if err != nil {
        return nil, fmt.Errorf("adrenaline error: %w – output: %s", err, out)
    }
    return out, nil
}
