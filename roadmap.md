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
- [ ] agent loop: implement full conversation handling in go
- [ ] memory store: persistent sqlite or file-based memory
- [ ] skill loader: discover and load .py skills from skills/
- [ ] gateway: implement telegram/discord cli platforms
- [ ] mcp server: full model context protocol implementation
- [ ] scheduler: cron-like job execution

## phase 3: hybrid operation
- [ ] go app can run standalone (no python needed for basic tasks)
- [ ] go app delegates complex tasks to python backend
- [ ] bi-directional communication between go and python

## phase 4: native desktop features
- [ ] system tray integration
- [ ] native notifications
- [ ] global hotkey activation
- [ ] auto-start on boot
- [ ] settings gui (not just web dashboard)

## phase 5: distribution
- [ ] windows installer (msi)
- [ ] macOS app bundle
- [ ] linux appimage
- [ ] auto-update mechanism