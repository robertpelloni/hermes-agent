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

## phase 3: go subsystem expansion (in progress)
- [x] tool parity catalog and provider routing
- [x] repo map generation and assimilation tracking
- [ ] agent loop: implement full autonomous reasoning loop in go
- [ ] memory store: persistent sqlite or file-based memory
- [ ] skill loader: discover and load .py skills from skills/
- [ ] gateway: implement telegram/discord cli platforms
- [ ] mcp server: full model context protocol implementation
- [ ] scheduler: cron-like job execution

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
