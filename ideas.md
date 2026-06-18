# ideas

## refactoring & re-architecture

### hybrid go-python bridge
instead of a full go rewrite, use the go app as a native desktop shell that communicates with the python agent via json-rpc (like the tui gateway does). this gives native window management + full python agent capabilities.

### electron-free desktop
the desktop app currently launches a browser window. consider embedding a webview (like webview/webview go library) so the dashboard renders inside a native window without requiring a full browser. this would feel more like a real desktop app.

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
## ultimate coding harness integrations
- comprehensive feature parity with 30+ top agentic tools (Claude Code, Warp, Replit, Aider, etc).
- multi-language architecture: ship 5 identical versions of the core engine in Rust, Go, Java, C#, TypeScript.
- Web browser extension integration: interface with ChatGPT, Gemini, Claude web interfaces to provide MCP server functionality, memory recording, web scraping.
