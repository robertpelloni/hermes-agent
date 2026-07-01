# roadmap

## phase 1: desktop launcher (complete)
- [x] go binary builds
- [x] finds project root and python
- [x] launches python hermes dashboard as subprocess
- [x] detects if dashboard already running
- [x] opens browser automatically
- [x] handles clean shutdown on ctrl+c
- [x] initializes go subsystems (stubs)

## phase 2: go subsystem expansion (next)
- [x] agent loop: implement full conversation handling in go
- [x] memory store: persistent sqlite or file-based memory
- [x] skill loader: discover and load .py skills from skills/
- [x] gateway: implement telegram/discord cli platforms
- [x] mcp server: full model context protocol implementation
- [x] scheduler: cron-like job execution

## phase 3: hybrid operation
- [ ] go app can run standalone (no python needed for basic tasks)
- [x] go app delegates complex tasks to python backend
- [x] bi-directional communication between go and python

## phase 4: native desktop features
- [x] system tray integration
- [x] native notifications
- [ ] global hotkey activation
- [ ] auto-start on boot
- [ ] settings gui (not just web dashboard)

## phase 5: distribution
- [ ] windows installer (msi)
- [ ] macOS app bundle
- [ ] linux appimage
- [ ] auto-update mechanism