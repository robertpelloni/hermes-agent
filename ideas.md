# ideas

## realized architectures (current milestones)
- **Native System Tray with Server Lifecycle**: Built in `apps/desktop/electron/tray.ts`, enabling backend restart, browser launch, and clean process termination.
- **Local SQLite Memory Persistence**: Pure-Go SQLite store (`modernc.org/sqlite`) at `~/.hermes/memory.db` with session enumeration and message management.
- **Unified Multi-Harness Build**: `scripts/build_all.sh` building Go binaries, running Go test suites, compiling Web UI distribution, and typechecking Desktop.
- **Autonomous Go Reasoning Loop**: Implemented with streaming (`HandleMessageStream`), structured reasoning (`EventThinking`), and lossy context compaction.

## refactoring & re-architecture

### hybrid go-python direct ipc bridge
instead of local HTTP polling, connect Electron Desktop directly to the Go daemon over a local Unix domain socket (`/tmp/hermes.sock`) or Windows named pipe (`\\.\pipe\hermes`) using JSON-RPC 2.0. This cuts out network port overhead and gives native window management with zero HTTP roundtrips.

### embedded native webview
replace browser launching with an embedded webview (such as `webview/webview` Go bindings or Wails) to render the dashboard inside an ultra-lightweight native OS window (<15MB RAM) without requiring full Chromium or Electron runtimes.

### sqlite-vec semantic memory search
compile `sqlite-vec` or pure-Go vector embeddings directly into `pkg/memory` so that cross-session semantic search happens locally in microseconds without cloud embedding APIs.

### windmill-style chaining
take inspiration from windmill.dev — allow users to chain ai actions visually in a drag-and-drop ui. the go app could host a local web server for this visual editor.

## feature expansions

### offline-first agent
build a lightweight go-only agent that works fully offline (using local llms like llama.cpp or ollama). the python backend handles complex tasks, but basic conversation works without internet.

### agent as windows service
run hermes as a windows service (or linux systemd unit) that stays running in the background. the desktop app connects to it. this enables always-on capabilities like cron jobs and messaging.

### multi-profile tray management
system tray with profiles: switch between "work", "personal", "coding" profiles. each profile has different models, skills, and providers.

### local-first memory
implement memory persistence in go (sqlite or badgerdb) that syncs to the python backend when available. this makes memory work even when the python backend is down.

### voice input
integrate windows speech recognition or whisper locally in go for voice-to-text input to the agent.

## language & platform

### consider rust rewrite
if the go app grows large, rust would give better performance and smaller binary size. the trade-off is development speed (go is faster to write).

### wasm plugin system
allow skills to be compiled to wasm and run inside the go app. this would be more portable than python skills.

### mobile companion
a flutter or kotlin mobile app that connects to the desktop agent for push notifications and quick queries.

## radical ideas

### agent-as-file-explorer
hermes could replace file explorer entirely — "show me all pdfs from last week" becomes a natural language query instead of clicking through folders.

### self-improving codebase
hermes could modify its own source code by submitting pull requests to itself after being granted permission.

### multi-agent marketplace
a p2p marketplace where users can rent out their hermes agent's compute or skills to other users, creating a decentralized ai compute network.