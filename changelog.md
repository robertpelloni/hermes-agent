# changelog

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
