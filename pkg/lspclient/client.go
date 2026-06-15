package lspclient

import (
    "bufio"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "os/exec"
    "sync"
    "time"
)

// Client wraps a JSON‑RPC connection to an LSP server (e.g., gopls).
// It lazily starts the server on first use and shuts it down when Close is called.
type Client struct {
    mu       sync.Mutex
    cmd      *exec.Cmd
    enc      *json.Encoder
    dec      *json.Decoder
    wr       *bufio.Writer
    nextID   int
    shutDown bool
}

var (
    globalClient *Client
    clientOnce   sync.Once
)

// Get returns a singleton LSP client, creating it on first call.
func Get() *Client {
    clientOnce.Do(func() {
        globalClient = &Client{nextID: 1}
    })
    return globalClient
}

// ensureRunning starts the LSP server if it is not already running.
func (c *Client) ensureRunning() error {
    c.mu.Lock()
    defer c.mu.Unlock()
    if c.shutDown {
        return errors.New("lsp client is shut down")
    }
    if c.cmd != nil {
        // Already running.
        return nil
    }
    // Locate gopls binary – try PATH and fallback to "go" tool.
    path, err := exec.LookPath("gopls")
    if err != nil {
        // Try invoking via "go tool gopls" as a fallback.
        // Ensure "go" exists.
        goPath, goErr := exec.LookPath("go")
        if goErr != nil {
            return fmt.Errorf("gopls not found and go tool unavailable: %w", err)
        }
        path = goPath
        // Use arguments to run as a tool.
        c.cmd = exec.Command(path, "tool", "gopls")
    } else {
        c.cmd = exec.Command(path)
    }
    // Connect stdin/stdout via pipes.
    stdin, err := c.cmd.StdinPipe()
    if err != nil {
        return fmt.Errorf("failed to get stdin pipe: %w", err)
    }
    stdout, err := c.cmd.StdoutPipe()
    if err != nil {
        return fmt.Errorf("failed to get stdout pipe: %w", err)
    }
    // We also capture stderr for debugging.
    c.cmd.Stderr = nil
    // Start the process.
    if err := c.cmd.Start(); err != nil {
        return fmt.Errorf("failed to start LSP server: %w", err)
    }
    // Use buffered writer for encoding.
    c.wr = bufio.NewWriter(stdin)
    c.enc = json.NewEncoder(c.wr)
    // Decoder reads directly from stdout.
    c.dec = json.NewDecoder(bufio.NewReader(stdout))
    // No background flusher – we'll flush after each encode.
    return nil
}

// Request represents a JSON‑RPC request object.
type request struct {
    JSONRPC string      `json:"jsonrpc"`
    ID      int         `json:"id"`
    Method  string      `json:"method"`
    Params  interface{} `json:"params,omitempty"`
}

// response represents a JSON‑RPC response.
type response struct {
    JSONRPC string          `json:"jsonrpc"`
    ID      int             `json:"id"`
    Result  json.RawMessage `json:"result,omitempty"`
    Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
}

// Call sends a request to the LSP server and returns the raw result.
func (c *Client) Call(method string, params interface{}) (json.RawMessage, error) {
    if err := c.ensureRunning(); err != nil {
        return nil, err
    }
    c.mu.Lock()
    id := c.nextID
    c.nextID++
    req := request{JSONRPC: "2.0", ID: id, Method: method, Params: params}
    if err := c.enc.Encode(req); err != nil {
        c.mu.Unlock()
        return nil, fmt.Errorf("failed to encode request: %w", err)
    }
    // Flush the writer to ensure the data is sent.
    if err := c.wr.Flush(); err != nil {
        c.mu.Unlock()
        return nil, fmt.Errorf("failed to flush request: %w", err)
    }
    c.mu.Unlock()

    // Wait for response with matching ID.
    deadline := time.Now().Add(5 * time.Second)
    for {
        if time.Now().After(deadline) {
            return nil, fmt.Errorf("timeout waiting for LSP response to %s", method)
        }
        var resp response
        if err := c.dec.Decode(&resp); err != nil {
            if err == io.EOF {
                return nil, fmt.Errorf("LSP server closed connection")
            }
            // Continue looping on decode errors (might be partial message).
            continue
        }
        if resp.ID != id {
            // Not our response; ignore.
            continue
        }
        if resp.Error != nil {
            return nil, fmt.Errorf("LSP error %d: %s", resp.Error.Code, resp.Error.Message)
        }
        return resp.Result, nil
    }
}

// Close terminates the LSP server process.
func (c *Client) Close() error {
    c.mu.Lock()
    defer c.mu.Unlock()
    if c.shutDown {
        return nil
    }
    c.shutDown = true
    if c.cmd != nil && c.cmd.Process != nil {
        _ = c.cmd.Process.Kill()
        _, _ = c.cmd.Process.Wait()
    }
    return nil
}

// Helper to build LSP position parameter.
func buildPosition(line, character int) map[string]int {
    return map[string]int{"line": line, "character": character}
}

// Helper to build TextDocumentIdentifier.
func buildTextDocument(uri string) map[string]string {
    return map[string]string{"uri": uri}
}

