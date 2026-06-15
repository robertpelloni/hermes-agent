package mcp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// ToolHandler processes an MCP tool call.
type ToolHandler func(params map[string]interface{}) (interface{}, error)

// ToolDefinition describes a tool for the MCP list endpoint.
type ToolDefinition struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"`
}

// Server implements an MCP (Model Context Protocol) server.
type Server struct {
	mu          sync.RWMutex
	name        string
	version     string
	tools       map[string]ToolHandler
	definitions map[string]ToolDefinition
}

// NewServer creates a new MCP server.
func NewServer(name, version string) *Server {
	return &Server{
		name:        name,
		version:     version,
		tools:       make(map[string]ToolHandler),
		definitions: make(map[string]ToolDefinition),
	}
}

// RegisterTool adds a tool to the server.
func (s *Server) RegisterTool(name string, handler ToolHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tools[name] = handler
	fmt.Printf("[hermes:mcp] Tool registered: %s\n", name)
}

// RegisterToolWithDefinition adds a tool with a full JSON description.
func (s *Server) RegisterToolWithDefinition(def ToolDefinition, handler ToolHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tools[def.Name] = handler
	s.definitions[def.Name] = def
	fmt.Printf("[hermes:mcp] Tool registered: %s (%s)\n", def.Name, def.Description)
}

// ListTools returns all registered tool definitions.
func (s *Server) ListTools() []ToolDefinition {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]ToolDefinition, 0, len(s.definitions))
	for _, def := range s.definitions {
		result = append(result, def)
	}
	return result
}

// CallTool executes a tool by name.
func (s *Server) CallTool(name string, params map[string]interface{}) (interface{}, error) {
	s.mu.RLock()
	handler, ok := s.tools[name]
	s.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("tool not found: %s", name)
	}

	return handler(params)
}

// ServeHTTP serves MCP tools over HTTP as a JSON-RPC endpoint.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		JSONRPC string          `json:"jsonrpc"`
		Method  string          `json:"method"`
		Params  json.RawMessage `json:"params,omitempty"`
		ID      interface{}     `json:"id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"jsonrpc": "2.0",
			"error":   map[string]interface{}{"code": -32700, "message": "Parse error"},
			"id":      nil,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	switch req.Method {
	case "tools/list":
		tools := s.ListTools()
		json.NewEncoder(w).Encode(map[string]interface{}{
			"jsonrpc": "2.0",
			"result":  map[string]interface{}{"tools": tools},
			"id":      req.ID,
		})

	case "tools/call":
		var callParams struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments"`
		}
		json.Unmarshal(req.Params, &callParams)

		result, err := s.CallTool(callParams.Name, callParams.Arguments)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"jsonrpc": "2.0",
				"error":   map[string]interface{}{"code": -32000, "message": err.Error()},
				"id":      req.ID,
			})
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"jsonrpc": "2.0",
			"result":  map[string]interface{}{"content": result},
			"id":      req.ID,
		})

	default:
		json.NewEncoder(w).Encode(map[string]interface{}{
			"jsonrpc": "2.0",
			"error":   map[string]interface{}{"code": -32601, "message": "Method not found"},
			"id":      req.ID,
		})
	}
}

// Start begins serving MCP requests on a given address.
func (s *Server) Start(address string) error {
	if address == "" {
		address = ":9090"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", s.ServeHTTP)

	fmt.Printf("[hermes:mcp] Server %s v%s listening on %s\n", s.name, s.version, address)
	return http.ListenAndServe(address, mux)
}
