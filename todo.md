# todo

## immediate (completed this turn)
- [x] attach `memory.Store` and connect global tools to MCP server in `cmd/hermes/main.go`
- [x] integrate persistent `memory.Store` into Go Agent conversation loops (`pkg/agent/agent.go`)
- [x] add unit test suite `TestAgentMemoryStorePersistence` (`pkg/agent/agent_e2e_test.go`)
- [x] implement `RunStandaloneLoop` with interactive slash commands in Go Agent (`pkg/agent/standalone.go`)
- [x] add unit test suite `TestStandaloneLoopSlashCommands` (`pkg/agent/standalone_test.go`)
- [x] wire `RunStandaloneLoop` into `cmd/hermes/main.go`
- [x] implement `TurnMetrics` and `TurnMetricsListener` with grace turn injection in Go Agent (`pkg/agent/agent.go`)
- [x] implement `TurnListener` interface and lifecycle callbacks in Go Agent (`pkg/agent/agent.go`)
- [x] add unit test suite `TestAgentTurnListenerHooks` in `pkg/agent/agent_e2e_test.go`
- [x] author unified cross-harness configuration specification (`docs/config/specification.md`)
- [x] cross-reference configuration architecture across user guide and Go specs
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
- [x] add automatic session label generation to agent loop
- [x] add session persistence to streaming agent loop (HandleMessageStream)
- [x] add unit tests for `pkg/toolregistry/registry.go`
- [x] add unit tests for `pkg/repomap/repomap.go`
- [x] add unit tests for `pkg/plugin/manager.go`
- [x] add unit tests for `pkg/assimilation/inventory.go`
- [x] add unit tests for `pkg/compaction/compaction.go`
- [x] add unit tests for `pkg/ai/helpers.go`
- [x] add unit tests for `pkg/subcmd/run_binary.go`
- [x] add unit tests for `pkg/lspclient/client.go`
- [x] add unit tests for `pkg/compat/compat.go`
- [x] document Configuration Architecture & Runtime Environment in go-subsystems-architecture.md
- [x] update deploy.md with environment variables matrix and Electron desktop launch guide
- [x] document Desktop & Go Subsystems environment and system tray in website user guide
- [x] document realized milestones and aggressive proposals in ideas.md
- [x] mark completed Phase 3 subsystems in roadmap.md

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
