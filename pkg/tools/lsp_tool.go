package tools

import (
    "fmt"

    "github.com/robertpelloni/hermes-agent/pkg/lspclient"
    "github.com/robertpelloni/hermes-agent/pkg/toolregistry"
)

func init() {
    toolregistry.Global().Register(&toolregistry.Tool{
        Name:        "lsp",
        Description: "Query a Language Server (gopls) for code intelligence. Supports 'definition', 'diagnostics', and 'symbol' methods.",
        Category:    "code_intelligence",
        Parameters: map[string]any{
            "method":    map[string]string{"type": "string", "description": "Operation: 'definition', 'diagnostics', 'symbol'"},
            "uri":       map[string]string{"type": "string", "description": "File URI (e.g., file:///path/to/file.go)"},
            "line":      map[string]string{"type": "integer", "description": "Line number (0‑based, required for 'definition')"},
            "character": map[string]string{"type": "integer", "description": "Character offset (0‑based, required for 'definition')"},
            "query":     map[string]string{"type": "string", "description": "Search query (required for 'symbol')"},
        },
        Handler: lspHandler,
        Native:  true,
    })
}

func lspHandler(args map[string]any, ctx map[string]any) (any, error) {
    methodRaw, ok := args["method"]
    if !ok {
        return nil, fmt.Errorf("missing 'method' argument")
    }
    method, ok := methodRaw.(string)
    if !ok {
        return nil, fmt.Errorf("'method' must be a string")
    }

    client := lspclient.Get()

    switch method {
    case "definition":
        uriRaw, ok := args["uri"]
        if !ok {
            return nil, fmt.Errorf("missing 'uri' for definition")
        }
        uri, ok := uriRaw.(string)
        if !ok {
            return nil, fmt.Errorf("'uri' must be a string")
        }
        lineRaw, ok := args["line"]
        if !ok {
            return nil, fmt.Errorf("missing 'line' for definition")
        }
        line, ok := lineRaw.(int)
        if !ok {
            // Handle JSON float conversion
            if lf, ok := lineRaw.(float64); ok {
                line = int(lf)
            } else {
                return nil, fmt.Errorf("'line' must be an integer")
            }
        }
        charRaw, ok := args["character"]
        if !ok {
            return nil, fmt.Errorf("missing 'character' for definition")
        }
        character, ok := charRaw.(int)
        if !ok {
            if cf, ok := charRaw.(float64); ok {
                character = int(cf)
            } else {
                return nil, fmt.Errorf("'character' must be an integer")
            }
        }
        params := map[string]any{
            "textDocument": map[string]string{"uri": uri},
            "position":     map[string]int{"line": line, "character": character},
        }
        result, err := client.Call("textDocument/definition", params)
        if err != nil {
            return nil, err
        }
        return string(result), nil

    case "diagnostics":
        uriRaw, ok := args["uri"]
        if !ok {
            return nil, fmt.Errorf("missing 'uri' for diagnostics")
        }
        uri, ok := uriRaw.(string)
        if !ok {
            return nil, fmt.Errorf("'uri' must be a string")
        }
        // Request workspace diagnostics (gopls specific) or use textDocument/publishDiagnostics.
        params := map[string]any{
            "textDocument": map[string]string{"uri": uri},
        }
        result, err := client.Call("textDocument/diagnostic", params)
        if err != nil {
            return nil, err
        }
        return string(result), nil

    case "symbol":
        queryRaw, ok := args["query"]
        if !ok {
            return nil, fmt.Errorf("missing 'query' for symbol")
        }
        query, ok := queryRaw.(string)
        if !ok {
            return nil, fmt.Errorf("'query' must be a string")
        }
        params := map[string]any{"query": query}
        result, err := client.Call("workspace/symbol", params)
        if err != nil {
            return nil, err
        }
        return string(result), nil

    default:
        return nil, fmt.Errorf("unsupported LSP method: %s", method)
    }
}