package tools

import (
    "encoding/json"
    "fmt"

    "github.com/robertpelloni/hermes-agent/pkg/memory"
    "github.com/robertpelloni/hermes-agent/pkg/toolregistry"
)

// graphStore is a singleton used by the tool handler.
var graphStore = memory.NewGraphStore()

func init() {
    toolregistry.Global().Register(&toolregistry.Tool{
        Name:        "graph_query",
        Description: "Run a query against the memory graph. Returns JSON array of matching nodes.",
        Category:    "knowledge",
        Parameters: map[string]any{
            "query": map[string]string{"type": "string", "description": "Graph DSL query (e.g., MATCH (n) RETURN n)"},
        },
        Handler: graphQueryHandler,
        Native:  true,
    })
}

func graphQueryHandler(args map[string]any, ctx map[string]any) (any, error) {
    raw, ok := args["query"]
    if !ok {
        return nil, fmt.Errorf("missing query argument")
    }
    q, ok := raw.(string)
    if !ok {
        return nil, fmt.Errorf("query must be a string")
    }
    if graphStore == nil {
        return nil, fmt.Errorf("graph store not initialized")
    }
    nodes, err := graphStore.Query(q)
    if err != nil {
        return nil, err
    }
    // Return JSON string for consistency with other tool outputs.
    b, err := json.Marshal(nodes)
    if err != nil {
        return nil, err
    }
    return string(b), nil
}
