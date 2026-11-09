# roadmap

## phase 1: desktop launcher & core harnesses (complete)
- [x] go binary builds
- [x] finds project root and python
- [x] launches python hermes dashboard as subprocess
- [x] detects if dashboard already running
- [x] opens browser automatically
- [x] handles clean shutdown on ctrl+c
- [x] initializes go subsystems (stubs)

## phase 2: dashboard consolidation & desktop controls (complete)
- [x] native system tray icon with quick actions and server shutdown (`apps/desktop/electron/tray.ts`)
- [x] consolidated overview & command center page (`web/src/pages/OverviewPage.tsx`)
- [x] categorized dashboard navigation (workspace, capabilities, automations, operations)
- [x] comprehensive tooltips and workflow guidance across web UI
- [x] unified multi-harness build automation script (`scripts/build_all.sh`)
- [x] non-destructive script archiving mechanism

## phase 3: go subsystem expansion (complete)
- [x] tool parity catalog and provider routing
- [x] repo map generation and assimilation tracking
- [x] agent loop: autonomous reasoning loop with streaming, EventThinking, context compaction, and session persistence
- [x] memory store: persistent SQLite storage (~/.hermes/memory.db) with message query and deletion
- [x] skill loader: recursive discovery of SKILL.md frontmatter and .py docstrings from skills/
- [x] gateway: dynamic platform registration, enumeration, and HTTP status probe
- [x] mcp server: JSON-RPC 2.0 tool endpoints (tools/list, tools/call)
- [x] scheduler: cron and interval task execution with background tick loop
- [x] session runtime: file-backed JSON session persistence and cross-session search
- [x] model registry: model capability search and API key resolution
- [x] plugin manager: dynamic discovery, lifecycle hooks, and tool registration

## phase 4: hybrid operation
- [ ] go app can run standalone (no python needed for basic tasks)
- [ ] go app delegates complex tasks to python backend
- [ ] bi-directional communication between go and python

## phase 5: native desktop features
- [x] system tray integration
- [x] native notifications
- [ ] global hotkey activation
- [ ] auto-start on boot
- [ ] settings gui (not just web dashboard)

## phase 6: distribution
- [ ] windows installer (msi)
- [ ] macOS app bundle
- [ ] linux appimage
