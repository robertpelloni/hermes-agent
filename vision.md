# vision

## mission
hermes desktop is a native go application that provides a desktop experience for the hermes agent — a self-improving ai agent with multi-platform messaging support, memory, skills, and tool integrations.

## core principles
1. **desktop-first launcher**: the go app wraps the python hermes dashboard, providing process management, browser integration, and a native desktop entry point
2. **progressive go porting**: start with a launcher, gradually port subsystems to go (agent, gateway, memory, skills)
3. **zero-loss merges**: always sync with upstream, preserve local progress, handle conflicts intelligently
4. **documentation-driven development**: every feature documented in roadmap, todo, and handoff files

## architecture
```
hermes-desktop.exe (go)
  ├── spawns → python hermes dashboard (--tui, port 9120)
  ├── opens → browser to dashboard url
  ├── manages → process lifecycle (start, wait, shutdown)
  └── initializes → go subsystems (agent, gateway, mcp, memory, scheduler, skill, sessionruntime, modelregistry, plugin, adapters)
```

## user experience goals
- one-click desktop launch
- automatic dashboard detection (skip spawning if already running)
- clean shutdown on ctrl+c
- browser auto-open on launch
- ready for future go-native features

## success metrics
- [x] binary builds without errors
- [x] launcher detects existing dashboard
- [x] launcher opens browser
- [x] clean process termination
- [x] go agent loop implemented (streaming, compaction, and tool execution)
- [x] go skill system implemented (frontmatter & docstring discovery in skills/)
- [x] go memory persistence implemented (SQLite ~/.hermes/memory.db)
- [x] go task scheduler implemented (cron and interval triggers)
- [x] go mcp server implemented (JSON-RPC 2.0 tools/list and tools/call)
- [x] native electron system tray with server lifecycle and clean quit