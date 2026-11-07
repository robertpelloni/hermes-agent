# Go Subsystems Architecture Specification

## Overview

The Go implementation of Hermes Agent (`cmd/hermes`) serves as a high-performance desktop launcher and autonomous agent runtime that operates alongside the Python core (FastAPI dashboard) and Electron desktop shell.

This document specifies the package architecture, data storage, JSON-RPC endpoints, and lifecycle integration for the Go subsystems.

---

## 1. Subsystem Architecture & Package Layout

```
cmd/hermes/
  └── main.go                  # Main entry point & unified subsystem launcher

pkg/
  ├── agent/                   # Agent core conversation loops and tool dispatch
  │     ├── agent.go
  │     └── events.go
  ├── memory/                  # Ephemeral in-memory store and SQLite persistent backend
  │     ├── memory.go
  │     ├── sqlite_store.go
  │     └── graph.go
  ├── skill/                   # Autonomous skill repository and filesystem discovery
  │     └── skill.go
  ├── mcp/                     # Model Context Protocol JSON-RPC 2.0 server
  │     └── mcp.go
  ├── scheduler/               # Cron and interval task execution tick loop
  │     └── scheduler.go
  ├── gateway/                 # Multi-platform communication gateway and status probes
  │     └── gateway.go
  ├── adapters/                # LLM provider routing and MCP adapters
  ├── compat/                  # Cross-harness tool parity catalog
  ├── modelregistry/           # Model capabilities and cost metadata
  └── repomap/                 # Aider-style ranked repository context mapping
```

---

## 2. Package Specifications

### 2.1 `pkg/memory` — Dual-Layer Memory & SQLite Persistence
- **In-Memory Cache**: Fast concurrent key-value storage (`sync.RWMutex`) for ephemeral context and working memory.
- **SQLite Storage**: Backed by `modernc.org/sqlite` (pure Go, zero CGO dependencies) stored at `~/.hermes/memory.db`.
  - **Schema**:
    ```sql
    CREATE TABLE IF NOT EXISTS messages (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        session_id TEXT NOT NULL,
        role TEXT NOT NULL,
        content TEXT NOT NULL,
        created_at DATETIME NOT NULL
    );
    CREATE INDEX IF NOT EXISTS idx_messages_session ON messages(session_id);
    ```
  - **API**:
    - `Store(key, value string)`: Ephemeral storage.
    - `Recall(key string) (string, bool)`: Ephemeral retrieval.
    - `SaveMessage(sessionID, role, content string) error`: Persist conversation turn to SQLite.
    - `LoadMessages(sessionID string) ([]MessageRecord, error)`: Query historical conversation turns.
    - `ListSessions() ([]string, error)`: Retrieve all distinct session identifiers.
    - `CountMessages(sessionID string) (int, error)`: Return message counts for session telemetry.
    - `DeleteSession(sessionID string) error`: Remove conversation records.

### 2.2 `pkg/skill` — Automated Skill Discovery & Loading
- **Repository**: Thread-safe storage for agent capabilities and tools.
- **Dynamic Discovery (`LoadFromDir`)**:
  - Recursively crawls `skills/`.
  - **`SKILL.md` Parsing**: Extracts YAML frontmatter (`name:`, `description:`).
  - **Python `.py` Parsing**: Extracts module-level docstrings (`"""..."""`) as skill documentation.
- **API**:
  - `Register(s *Skill)`
  - `Lookup(name string) (*Skill, bool)`
  - `List() []*Skill`
  - `LoadFromDir(dir string) (int, error)`

### 2.3 `pkg/mcp` — Model Context Protocol (MCP) Server
- **Protocol**: JSON-RPC 2.0 over HTTP (`/mcp` endpoint, default `:9090`).
- **Methods**:
  - `tools/list`: Returns array of registered `ToolDefinition` objects (name, description, parameter schemas).
  - `tools/call`: Executes dynamic `ToolHandler` with provided parameters and returns JSON content.
- **API**:
  - `NewServer(name, version string) *Server`
  - `RegisterToolWithDefinition(def ToolDefinition, handler ToolHandler)`
  - `CallTool(name string, params map[string]interface{}) (interface{}, error)`
  - `ServeHTTP(w http.ResponseWriter, r *http.Request)`

### 2.4 `pkg/scheduler` — Autonomous Task Scheduling
- **Engine**: Background tick loop (`runLoop`) evaluating registered cron/interval triggers.
- **Schedule Formats**:
  - RFC3339 timestamps (one-shot).
  - Duration intervals (e.g. `30m`, `2h`, `1d`, `7d`).
  - Natural cadence (e.g. `every monday 9am`, `every 10m`).
- **API**:
  - `Schedule(name, schedule string, handler func()) error`
  - `PauseJob(name string) error` / `ResumeJob(name string) error`
  - `RemoveJob(name string) error`
  - `Start(ctx context.Context)` / `Stop()`

### 2.5 `pkg/gateway` — Communication Gateway & Health Probe
- **Platforms**: Extensible communication platform interface (`Platform`) supporting CLI, Discord, Telegram, and Slack.
- **HTTP Status Probe**: `StatusHandler` responds to `GET /status` with JSON summary of gateway state and registered platforms for health monitoring by outer processes.
- **API**:
  - `New(ag *agent.Agent) *Gateway`
  - `RegisterPlatform(p Platform)`
  - `Platforms() []string`
  - `StatusHandler(w http.ResponseWriter, r *http.Request)`
  - `Start(ctx context.Context) error` / `Stop()`

### 2.6 `pkg/agent` — Conversational Agent Loop & Streaming Dispatch
- **Execution Loop**:
  - Automatically loads system prompt from templates (`pkg/prompttemplates`).
  - Estimates token count and triggers context compaction (`maybeCompact`) when token limits are approached.
  - Intercepts structured and text-based tool calls (`detectTextToolCall`), dispatches execution to `pkg/toolregistry`, and appends tool results to the conversation.
- **Streaming Pipeline**:
  - `HandleMessageStream` streams `StreamEvent` structures (`text`, `tool_start`, `tool_done`, `outcome`, `done`, `error`) over a Go channel.
- **Session Persistence**:
  - Integrates with `pkg/sessionruntime.Runtime` via `SetSessionRuntime(rt)`.
  - Automatically records user input and assistant final responses per session.

### 2.7 `pkg/sessionruntime` — Session Lifecycle & Workspace Management
- **Persistence**: File-backed JSON session persistence stored in `~/.hermes/sessions/` (`autoSave` mode).
- **Features**:
  - `Create(id)` / `Get(id)` / `List()`
  - `AppendMessage(sessionID, msg)`
  - `Truncate(sessionID, keep)`: Reduces conversation context while preserving system prompt.
  - `Fork(sourceID, newID, count)`: Forks an existing session starting from message *N*.
  - `Search(query, limit)`: Substring message search across all active sessions.
  - `Delete(id)`: Removes memory and deletes the JSON session record from disk.

### 2.8 `pkg/modelregistry` — Model Capabilities & Provider Authentication
- **Provider Resolution**:
  - Catalogs major LLM providers (OpenAI, Anthropic, Google, Groq, xAI, Mistral, OpenRouter, Cerebras, DeepSeek, MiniMax).
  - Resolves active API keys from environment variable fallbacks (`GetAPIKey`, `HasAPIKey`, `HasAnyAPIKey`).
- **Model Metadata**:
  - Catalogs model capability flags (reasoning, vision, tool use, audio).
  - Provides model search (`SearchModels`) and tabular formatting (`FormatProviderStatus`).

---

## 3. Desktop Application Integration (`cmd/hermes/main.go`)

On launch, `cmd/hermes/main.go`:
1. Locates Python runtime and project root.
2. Initializes the persistent SQLite memory store (`pkg/memory`).
3. Discovers and indexes skills from `skills/` into `pkg/skill`.
4. Starts the background task scheduler (`pkg/scheduler`).
5. Launches the background MCP server on `:9090` (`pkg/mcp`).
6. Starts the multi-platform gateway (`pkg/gateway`).
7. Spawns or connects to the Python dashboard on port `9120` and opens the user's browser.
8. Manages graceful shutdown on `SIGINT`/`SIGTERM` to stop the scheduler, close SQLite connections, and terminate gateway platforms.
