# changelog

## 1.0.0-build.30 (2026-11-05)

### added
- Unit test suite for Go Compat Subsystem (`pkg/compat/compat_test.go`) covering tool parity catalog registration, cross-harness lookup, and summary generation.

## 1.0.0-build.29 (2026-11-05)

### added
- Unit test suite for Go LSP Client subsystem (`pkg/lspclient/client_test.go`) covering client singleton initialization, safe closing, and lifecycle protection against post-shutdown calls.

## 1.0.0-build.28 (2026-11-05)

### added
- Unit test suite for Go Subcmd subsystem (`pkg/subcmd/run_binary_test.go`) covering external binary execution, combined stdout/stderr capture, and context timeout cancellation.

## 1.0.0-build.27 (2026-11-05)

### added
- Unit test suite for Go AI Helpers subsystem (`pkg/ai/helpers_test.go`) covering text message creation, tool call/result serialization, error formatting, and token count estimation.

## 1.0.0-build.26 (2026-11-05)

### added
- Unit test suite for Go Compaction subsystem (`pkg/compaction/compaction_test.go`) covering token estimation, threshold triggers, and lossy context summarization.

## 1.0.0-build.25 (2026-11-05)

### added
- Unit test suite for Go Assimilation Subsystem (`pkg/assimilation/inventory_test.go`) validating toolchain inventory and formatted catalog output.

## 1.0.0-build.24 (2026-11-05)

### added
- Updated `vision.md` architecture diagram tree to enumerate all initialized Go subsystems (`agent, gateway, mcp, memory, scheduler, skill, sessionruntime, modelregistry, plugin, adapters`).

## 1.0.0-build.23 (2026-11-05)

### added
- Added `EventThinking` streaming event type to `pkg/agent/events.go` and wired it into `pkg/agent/agent.go` to emit structured reasoning deltas during LLM streaming.

## 1.0.0-build.22 (2026-11-05)

### added
- Unit test suite for Go Plugin subsystem (`pkg/plugin/manager_test.go`) covering plugin discovery, registration, lifecycle hooks (OnLoad/OnUnload), tool registration/unregistration, and status reporting.

## 1.0.0-build.21 (2026-11-05)

### added
- Multi-turn conversational memory injection in Go Agent (`buildInitialMessagesWithSession` in `pkg/agent/agent.go`), allowing past session history to provide context across sequential turns.
- Multi-turn conversation persistence test in `pkg/agent/agent_e2e_test.go`.

## 1.0.0-build.20 (2026-11-05)

### added
- Unit test suite `pkg/orchestration/planner_test.go` verifying execution plan derivation, keyword-based repository mapping, and formatted plan output.

## 1.0.0-build.19 (2026-11-05)

### added
- Unit test suite `internal/tui/commands_test.go` covering slash command registry, alias resolution, prefix stripping, and help/status handlers.

## 1.0.0-build.18 (2026-11-05)

### added
- Updated `vision.md` metrics and architectural alignment reflecting completed Go agent loop, skills system, SQLite memory persistence, task scheduler, MCP server, and Electron system tray.

## 1.0.0-build.17 (2026-11-05)

### added
- Environment variable configuration support for `HERMES_MODEL`, `HERMES_PROVIDER`, and `HERMES_BASE_URL` in Go `agent.DefaultConfig()` (`pkg/agent/agent.go`).
- Unit test coverage `TestDefaultConfigWithEnvOverrides` in `pkg/agent/agent_e2e_test.go`.

## 1.0.0-build.16 (2026-11-05)

### added
- Documented `pkg/adapters` (provider route resolution and MCP discovery) in `docs/design/go-subsystems-architecture.md`.
- Added "Future Integration Points & Native Desktop Expansion" architecture roadmap covering Electron IPC channels, Pi-Mono streaming alignment, and self-improving skill synthesis.

## 1.0.0-build.15 (2026-11-05)

### added
- Unit test suite for `pkg/adapters` (`pkg/adapters/adapters_test.go`) covering provider route selection, cost preferences, execution plan formatting, and MCP configuration discovery.

## 1.0.0-build.14 (2026-11-05)

### added
- Detailed architecture specifications for `pkg/agent`, `pkg/sessionruntime`, and `pkg/modelregistry` in `docs/design/go-subsystems-architecture.md`.

## 1.0.0-build.13 (2026-11-05)

### added
- Streaming agent loop unit tests (`TestAgentStreamingLoop`) in `pkg/agent/agent_e2e_test.go` verifying delta event propagation, outcome delivery, and done events.
- Agent session persistence verification (`TestAgentSessionPersistence`) in `pkg/agent/agent_e2e_test.go` ensuring conversation turns are recorded into `sessionruntime.Runtime`.

## 1.0.0-build.12 (2026-11-05)

### added
- Session Runtime unit test suite (`pkg/sessionruntime/sessionruntime_test.go`) covering session lifecycle, message appending, search, forking, and context truncation.
- Model Registry unit test suite (`pkg/modelregistry/registry_test.go`) testing provider registration, key resolution, and model lookup.
- Integrated `sessionruntime.Runtime` into Go `Agent` (`pkg/agent/agent.go`) to automatically record conversational turns per session.
- Subsystem initialization in `cmd/hermes/main.go` wires `sessionruntime.NewRuntime` directly to the `Agent`.

## 1.0.0-build.11 (2026-11-05)

### added
- comprehensive architecture documentation specification for Go subsystems (`docs/design/go-subsystems-architecture.md`) detailing memory, skills, MCP, scheduler, and gateway layouts.

## 1.0.0-build.10 (2026-11-05)

### added
- HTTP status and health endpoint (`StatusHandler`) on Go Gateway (`pkg/gateway/gateway.go`) to expose registered communication platforms and gateway status.
- unit test coverage in `pkg/gateway/gateway_test.go`.

## 1.0.0-build.9 (2026-11-05)

### added
- dynamic platform registration (`RegisterPlatform`) and query (`Platforms`) in Go gateway subsystem (`pkg/gateway/gateway.go`).
- unit test coverage in `pkg/gateway/gateway_test.go` verifying platform lifecycle management and startup dispatch.

## 1.0.0-build.8 (2026-11-05)

### added
- updated `deploy.md` with instructions for `scripts/build_all.sh` and multi-harness commands.
- updated `memory.md` recording architecture of Go subsystems (SQLite memory, skill loader, scheduler) and Electron desktop system tray.

## 1.0.0-build.7 (2026-11-05)

### added
- connected Go skill discovery (`pkg/skill.LoadFromDir`) and task scheduler (`pkg/scheduler.New`) into the Go desktop initialization lifecycle in `cmd/hermes/main.go` with graceful shutdown handling.

## 1.0.0-build.6 (2026-11-05)

### added
- automated unit test coverage for Model Context Protocol (MCP) server (`pkg/mcp/mcp_test.go`) covering JSON-RPC tool listing, dynamic invocation, and result serialization.

## 1.0.0-build.5 (2026-11-05)

### added
- automated skill loader and directory crawler in Go (`pkg/skill/skill.go`): `LoadFromDir()`, supporting YAML frontmatter extraction from `SKILL.md` and module docstring parsing from `.py` files.
- unit test suite `pkg/skill/skill_test.go` verifying discovery, metadata extraction, and repository registration.

## 1.0.0-build.4 (2026-11-04)

### added
- persistent session query and lifecycle methods in Go memory package (`pkg/memory/sqlite_store.go` and `pkg/memory/memory.go`): `ListSessions()`, `CountMessages()`, and `DeleteSession()`.
- unit test suite `pkg/memory/sqlite_store_test.go` verifying SQLite schema creation, message inserts, counting, session enumeration, and session deletion.

## 1.0.0-build.3 (2026-11-04)

### added
- system tray icon and server management (`apps/desktop/electron/tray.ts`) with quick status indicators, main window toggling, external dashboard launcher, docs shortcuts, backend restart, and safe graceful server shutdown.
- vitest unit test suite for electron system tray actions (`apps/desktop/electron/tray.test.ts`).
- consolidated command center & overview page (`web/src/pages/OverviewPage.tsx`) unifying live telemetry, active model stats, persistent session counters, scheduled automations, quick workflows, and platform health.
- unit test coverage for the overview command center (`web/src/pages/OverviewPage.test.tsx`).
- unified multi-harness build & check script (`scripts/build_all.sh`).
- non-destructive archive mechanism for legacy scripts under `scripts/archive/`.

### changed
- redesigned dashboard sidebar navigation in `web/src/App.tsx` into four prioritized functional sections: Workspace, Capabilities & Intelligence, Automations & Connectors, and Operations & System.
- added informative hover tooltips across collapsed and expanded nav items to explain feature roles.
- updated `start.bat` with `build-all` and `web` targets.
- recompiled production web bundle into `hermes_cli/web_dist/`.

## 1.0.0-build.2 (2026-06-14)

### added
- tool parity catalog (`pkg/compat`) - exact-name contracts from Claude Code, Codex, Warp, Aider, pi, and 10+ other harnesses
- provider routing (`pkg/adapters`) - cost/preference-aware provider selection with MCP and HyperCode adapters
- repo map generation (`pkg/repomap`) - aider-style context condensation with ranked file relevance
- foundation mode (`--foundation`) - integrated features from pi-mono and hyperharness
- mode switching in START.BAT for foundation mode

### changed
- main.go imports and initializes all integration packages
- START.BAT updated with foundation command and comprehensive help
- hermes-desktop.exe rebuilt with new packages (6.0 MB)

## 1.0.0-build.1 (2026-06-13)

### added
- go desktop app (`hermes-desktop.exe`) - launches python hermes web dashboard with process management
- go packages: `pkg/agent`, `pkg/gateway`, `pkg/mcp`, `pkg/memory`, `pkg/scheduler`, `pkg/skill`
- documentation structure: `version.md`, `changelog.md`, `roadmap.md`, `todo.md`, `vision.md`, `memory.md`, `deploy.md`, `ideas.md`, `handoff.md`

### changed
- merged upstream/main (34 commits) - ssl ca guard, read_extract tool, desktop scroll-to-bottom button, test improvements

### fixed
- upstream merge conflicts resolved (tinker-atropos submodule deletion accepted)

### notes
- go desktop app functions as launcher for python hermes dashboard (port 9120)
- go packages are stub implementations ready for expansion
- dashboard already running on port 9120 detected by launcher
