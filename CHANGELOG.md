# changelog

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