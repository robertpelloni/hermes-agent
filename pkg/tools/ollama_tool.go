package tools

import (
    "context"
    "fmt"
    "time"

    "github.com/robertpelloni/hermes-agent/pkg/subcmd"
    "github.com/robertpelloni/hermes-agent/pkg/toolregistry"
)

func init() {
    // Run command
    toolregistry.Global().Register(&toolregistry.Tool{
        Name:        "ollama_run",
        Description: "Thin wrapper for the Ollama CLI `run` command.",
        Category:    "llm",
        Parameters: map[string]any{
            "model":  map[string]string{"type": "string", "description": "Model name"},
            "prompt": map[string]string{"type": "string", "description": "User prompt"},
            "json":   map[string]string{"type": "string", "description": "Return JSON (optional, set to 'true')"},
        },
        Handler: ollamaRunHandler,
        Native:  true,
    })
    // List command
    toolregistry.Global().Register(&toolregistry.Tool{
        Name:        "ollama_list",
        Description: "Thin wrapper for the Ollama CLI `list` command.",
        Category:    "utility",
        Parameters:  map[string]any{},
        Handler:     ollamaListHandler,
        Native:      true,
    })
}

func ollamaRunHandler(args map[string]any, ctx map[string]any) (any, error) {
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
    cliArgs := []string{"run", model, prompt}
    if jsonRaw, ok := args["json"]; ok {
        if jsonStr, ok := jsonRaw.(string); ok && jsonStr == "true" {
            cliArgs = append([]string{"run", "--json"}, cliArgs[1:]...)
        }
    }
    out, err := subcmd.RunBinary(context.Background(), "ollama", cliArgs, 30*time.Second)
    if err != nil {
        return nil, fmt.Errorf("ollama run error: %w – output: %s", err, out)
    }
    return out, nil
}

func ollamaListHandler(args map[string]any, ctx map[string]any) (any, error) {
    out, err := subcmd.RunBinary(context.Background(), "ollama", []string{"list"}, 15*time.Second)
    if err != nil {
        return nil, fmt.Errorf("ollama list error: %w – output: %s", err, out)
    }
    return out, nil
}
