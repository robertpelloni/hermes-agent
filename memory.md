# memory

## project observations

### architecture
- python hermes-agent is the core implementation (~12k loc in run_agent.py)
- go desktop app is a thin wrapper/launcher (currently ~150 loc in cmd/hermes/main.go)
- go packages are stubs: agent, gateway, mcp, memory, scheduler, skill

### build system
- python: uses uv/pip, venv in .venv/ or venv/
- go: go 1.24.3, builds to hermes-desktop.exe (6.1mb)
- tui: node/ink typescript, built with pnpm
- web dashboard: fastapi + react, built with vite

### key paths
- project root: C:/Users/hyper/workspace/hermes-agent/
- python: C:/Python314/python.exe (system python, not venv)
- dashboard: port 9120 (configurable via HERMES_PORT)
- go binary: ./hermes-desktop.exe

### upstream relationship
- fork: github.com/robertpelloni/hermes-agent
- upstream: github.com/NousResearch/hermes-agent
- merge strategy: regular syncs, preserve local go/ desktop work
- local untracked files: .pi/, cmd/, pkg/, go.mod, start.bat, hermes-desktop.exe

### design decisions
- go app launches python dashboard (not a full rewrite)
- go packages initialized as stubs for future expansion
- dashboard detection avoids double-spawning
- browser auto-open on launch
- clean shutdown on ctrl+c

### quirks
- "no caps files" = use lowercase filenames for new docs
- dashboard already running on 9120 from previous sessions
- go binary uses system python, not venv python
- start.bat updated to reference hermes-desktop.exe (was hermes-agent-go.exe)