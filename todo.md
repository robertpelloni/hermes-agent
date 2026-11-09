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
- [x] expand Go architecture specs in `docs/design/go-subsystems-architecture.md`
- [x] add unit tests for `pkg/adapters`
- [x] document future integration points (IPC, Pi-Mono streaming, self-improving loop)
- [x] align vision.md success metrics with completed Go subsystems and system tray
- [x] update vision.md architecture diagram with full subsystem initialization list
- [x] add unit tests for `internal/tui/commands.go`

## near-term (next session)
- [ ] test foundation mode: START.BAT foundation
- [x] expand agent loop (pkg/agent/agent.go) with streaming response handling & tests
- [x] implement memory store with sqlite persistence (pkg/memory/memory.go)
- [x] add basic skill loader that discovers .py files in skills/
- [x] run go test ./pkg/... ./cmd/...
- [x] add unit tests for `pkg/orchestration/planner.go`
- [x] add multi-turn conversation memory context to agent loop
- [x] add unit tests for `pkg/plugin/manager.go`
- [x] add unit tests for `pkg/assimilation/inventory.go`
- [x] add unit tests for `pkg/compaction/compaction.go`
- [x] add unit tests for `pkg/ai/helpers.go`
- [x] add unit tests for `pkg/subcmd/run_binary.go`

## medium-term
- [x] implement gateway platform registration and lifecycle (pkg/gateway/gateway.go)
- [x] implement mcp server with basic tool registration and testing (pkg/mcp/mcp.go)
- [x] implement scheduler with cron expression parsing (pkg/scheduler/scheduler.go) and wire to main
- [x] add environment variable config: HERMES_MODEL, HERMES_PROVIDER, HERMES_BASE_URL
- [ ] integrate pi-ai multi-provider API from pi-mono
- [x] integrate pi-agent-core event streaming (EventThinking) from pi-mono

## features integrated from other harnesses
- [x] tool parity catalog (hyperharness compat package)
- [x] provider routing and repo mapping
- [x] native system tray icon and server lifecycle management
