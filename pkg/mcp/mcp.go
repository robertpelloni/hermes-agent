package mcp

import "fmt"

// Server implements an MCP (Model Context Protocol) server.
type Server struct {
	name    string
	version string
	tools   map[string]ToolHandler
}

// ToolHandler processes an MCP tool call.
type ToolHandler func(params map[string]interface{}) (interface{}, error)

// NewServer creates a new MCP server.
func NewServer(name, version string) *Server {
	return &Server{
		name:    name,
		version: version,
		tools:   make(map[string]ToolHandler),
	}
}

// RegisterTool adds a tool to the server.
func (s *Server) RegisterTool(name string, handler ToolHandler) {
	s.tools[name] = handler
	fmt.Printf("[hermes:mcp] Tool registered: %s\n", name)
}

// Start begins serving MCP requests.
func (s *Server) Start() error {
	fmt.Printf("[hermes:mcp] Server %s v%s started\n", s.name, s.version)
	return nil
}
