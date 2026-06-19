# vision

## mission
hermes desktop is a native go application that provides a desktop experience for the hermes agent — a self-improving ai agent with multi-platform messaging support, memory, skills, and tool integrations.

## core principles
1. **desktop-first launcher**: the go app wraps the python hermes dashboard, providing process management, browser integration, and a native desktop entry point
2. **progressive go porting**: start with a launcher, gradually port subsystems to go (agent, gateway, memory, skills)
3. **zero-loss merges**: always sync with upstream, preserve local progress, handle conflicts intelligently
4. **documentation-driven development**: every feature documented in roadmap, todo, and handoff files

## architecture
```
hermes-desktop.exe (go)
  ├── spawns → python hermes dashboard (--tui, port 9120)
  ├── opens → browser to dashboard url
  ├── manages → process lifecycle (start, wait, shutdown)
  └── initializes → go subsystems (agent, gateway, mcp, memory, scheduler, skill)
```

## user experience goals
- one-click desktop launch
- automatic dashboard detection (skip spawning if already running)
- clean shutdown on ctrl+c
- browser auto-open on launch
- ready for future go-native features

## success metrics
- [x] binary builds without errors
- [x] launcher detects existing dashboard
- [x] launcher opens browser
- [x] clean process termination
- [ ] go agent loop implemented (future)
- [ ] go skill system implemented (future)
- [ ] go memory persistence implemented (future)
## The Ultimate Agentic Coding Harness
We are turning this into the ULTIMATE AGENTIC CODING HARNESS. We will rebuild/port the entire system identically in 5 languages: TypeScript, Rust, Go, C#, and Java. It will include all features and functionalities from over 30 leading CLI tools, and we will feature parity with Amp, Auggie, Claude Code, Codebuff, Codemachine, Codex, Copilot CLI, Crush, Factory Droid, Gemini CLI, Goose CLI, Grok Build, Kilo Code CLI, Kimi CLI, Mistral Vibe CLI, Opencode, Qwen Code CLI, Warp CLI, Trae CLI. We will be the best CLI/TUI/WebUI tool.
