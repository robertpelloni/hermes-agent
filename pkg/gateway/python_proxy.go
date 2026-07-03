package gateway

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// PythonProxy represents a connection to the Python dashboard backend.
type PythonProxy struct {
	BaseURL string
	Client  *http.Client
}

// NewPythonProxy creates a new proxy to communicate with the python backend.
func NewPythonProxy() *PythonProxy {
	port := os.Getenv("HERMES_PORT")
	if port == "" {
		port = "9120"
	}
	return &PythonProxy{
		BaseURL: fmt.Sprintf("http://127.0.0.1:%s", port),
		Client:  &http.Client{},
	}
}

// Ping checks if the Python backend is alive.
func (p *PythonProxy) Ping() bool {
	resp, err := p.Client.Get(p.BaseURL + "/api/status")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// DelegateTask sends a complex task to the Python backend to be handled by its full tool suite.
func (p *PythonProxy) DelegateTask(taskID string, payload map[string]interface{}) (map[string]interface{}, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", p.BaseURL+"/api/delegate", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Task-ID", taskID)

	resp, err := p.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("python backend request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("python backend returned status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}
