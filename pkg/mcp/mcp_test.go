package mcp

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMCPServer(t *testing.T) {
	srv := NewServer("test-server", "1.0.0")

	def := ToolDefinition{
		Name:        "add",
		Description: "Adds two numbers",
		Parameters:  nil,
	}
	srv.RegisterToolWithDefinition(def, func(params map[string]interface{}) (interface{}, error) {
		a, _ := params["a"].(float64)
		b, _ := params["b"].(float64)
		return a + b, nil
	})

	// 1. List tools directly
	tools := srv.ListTools()
	if len(tools) != 1 || tools[0].Name != "add" {
		t.Fatalf("expected 1 tool 'add', got %+v", tools)
	}

	// 2. Call tool directly
	res, err := srv.CallTool("add", map[string]interface{}{"a": 10.0, "b": 5.0})
	if err != nil {
		t.Fatalf("CallTool failed: %v", err)
	}
	if res.(float64) != 15.0 {
		t.Errorf("expected 15.0, got %v", res)
	}

	// 3. Test HTTP JSON-RPC tools/list
	listBody := []byte(`{"jsonrpc":"2.0","method":"tools/list","id":1}`)
	req := httptest.NewRequest("POST", "/mcp", bytes.NewReader(listBody))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("tools/list status = %d, expected 200", w.Code)
	}

	var listResp struct {
		Result struct {
			Tools []ToolDefinition `json:"tools"`
		} `json:"result"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("unmarshal list response failed: %v", err)
	}
	if len(listResp.Result.Tools) != 1 {
		t.Fatalf("expected 1 tool in HTTP response, got %d", len(listResp.Result.Tools))
	}

	// 4. Test HTTP JSON-RPC tools/call
	callBody := []byte(`{"jsonrpc":"2.0","method":"tools/call","params":{"name":"add","arguments":{"a":20,"b":22}},"id":2}`)
	req = httptest.NewRequest("POST", "/mcp", bytes.NewReader(callBody))
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("tools/call status = %d, expected 200", w.Code)
	}

	var callResp struct {
		Result struct {
			Content float64 `json:"content"`
		} `json:"result"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &callResp); err != nil {
		t.Fatalf("unmarshal call response failed: %v", err)
	}
	if callResp.Result.Content != 42.0 {
		t.Errorf("expected content 42.0, got %v", callResp.Result.Content)
	}
}
