# todo

## immediate (completed this turn)
- [x] implement native system tray with server lifecycle and quit hooks (`apps/desktop/electron/tray.ts`)
- [x] write tray unit tests and verify in vitest (`apps/desktop/electron/tray.test.ts`)
- [x] create consolidated Command Center / Overview page (`web/src/pages/OverviewPage.tsx`)
- [x] reorganize dashboard sidebar into prioritized sections with descriptive tooltips
- [x] create unified build and test runner (`scripts/build_all.sh`)
- [x] add non-destructive script archiving and update `start.bat`
- [x] compile web distribution into `hermes_cli/web_dist/`
- [x] add unit tests for `pkg/sessionruntime` and `pkg/modelregistry`
- [x] connect `sessionruntime.Runtime` to `agent.Agent` and `cmd/hermes/main.go` lifecycle

## near-term (next session)
- [ ] test foundation mode: START.BAT foundation
- [x] expand agent loop (pkg/agent/agent.go) with streaming response handling & tests
- [x] implement memory store with sqlite persistence (pkg/memory/memory.go)
- [x] add basic skill loader that discovers .py files in skills/
- [ ] run go test ./pkg/... ./cmd/...

## medium-term
- [x] implement gateway platform registration and lifecycle (pkg/gateway/gateway.go)
- [x] implement mcp server with basic tool registration and testing (pkg/mcp/mcp.go)
- [x] implement scheduler with cron expression parsing (pkg/scheduler/scheduler.go) and wire to main
- [ ] add environment variable config: HERMES_MODEL, HERMES_PROVIDER, HERMES_PORT
- [ ] integrate pi-ai multi-provider API from pi-mono
- [ ] integrate pi-agent-core event streaming from pi-mono

## features integrated from other harnesses
- [x] tool parity catalog (hyperharness compat package)
- [x] provider routing and repo mapping
- [x] native system tray icon and server lifecycle management
