# todo

## immediate (next session)
- [ ] test foundation mode: START.BAT foundation
- [ ] verify tool parity catalog shows all harnesses
- [ ] verify provider routing selects correct providers
- [ ] verify repo map generation works on project
- [ ] test go desktop binary on clean windows install
- [ ] verify browser opens to correct dashboard url
- [ ] verify clean shutdown terminates both go binary and python subprocess
- [ ] verify dashboard detection skips re-spawn when already running

## near-term (this week)
- [ ] implement agent loop (pkg/agent/agent.go) - basic conversation handling
- [ ] implement memory store with sqlite persistence (pkg/memory/memory.go)
- [ ] add basic skill loader that discovers .py files in skills/
- [ ] test go test ./pkg/... ./cmd/... passes

## medium-term
- [ ] implement gateway cli platform (pkg/gateway/gateway.go) - read from stdin, write to stdout
- [ ] implement mcp server with basic tool registration (pkg/mcp/mcp.go)
- [ ] implement scheduler with cron expression parsing (pkg/scheduler/scheduler.go)
- [ ] add environment variable config: HERMES_MODEL, HERMES_PROVIDER, HERMES_PORT
- [ ] integrate pi-ai multi-provider API from pi-mono
- [ ] integrate pi-agent-core event streaming from pi-mono

## optional/nice-to-have
- [ ] system tray icon for windows
- [ ] auto-reconnect if dashboard crashes
- [ ] logging to file instead of stdout
- [ ] config file (yaml or json) for desktop-specific settings
- [ ] graceful shutdown with timeout (kill subprocess after 10s)

## features integrated from other harnesses
- [x] tool parity catalog (hyperharness compat package)
- [x] provider routing (hyperharness foundation providers)
- [x] repo map generation (hyperharness foundation repomap)
- [x] mcp adapter system (hyperharness foundation adapters)
- [x] hypercode adapter (hyperharness foundation adapters)
- [ ] pi-ai unified multi-provider API (pi-mono packages/ai)
- [ ] pi-agent-core agent runtime (pi-mono packages/agent)
- [ ] pi-tui differential rendering (pi-mono packages/tui)
- [ ] pi-web-ui web components (pi-mono packages/web-ui)