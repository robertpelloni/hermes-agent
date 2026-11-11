# deploy

## prerequisites
- go 1.24.3+
- python 3.12+ (with hermes-cli dependencies installed)
- git (for repository operations)

## build

### go desktop binary
```bash
cd C:/Users/hyper/workspace/hermes-agent
go build -buildvcs=false -ldflags="-s -w" -o hermes-desktop.exe ./cmd/hermes
```

### using start.bat
```cmd
start.bat build       # build binary only
start.bat build-all   # build Go binary and Web UI distribution
start.bat run         # build and run
start.bat web         # start web dashboard
```

### unified multi-harness build
```bash
./scripts/build_all.sh
```
This script compiles the Go desktop binary, runs Go unit tests, builds production web assets (`web/` -> `hermes_cli/web_dist/`), and typechecks the Electron desktop application.

## environment variables
| variable | default | description |
|----------|---------|-------------|
| HERMES_PORT | 9120 | dashboard port |
| HERMES_MODEL | free-llm | model name for agent |
| HERMES_PROVIDER | local-llm | provider name (e.g. openai, anthropic, local-llm) |
| HERMES_BASE_URL | http://127.0.0.1:4000/v1 | LLM server API base URL |
| HERMES_FREE_LLM_PORT | 4000 | fallback port for local development LLM server |
| HERMES_FREE_LLM_API_KEY | not-needed | API key for local LLM requests |
| HERMES_MCP_PORT | :9090 | port for background Model Context Protocol JSON-RPC server |
| HERMES_COMPACT_THRESHOLD | 8000 | token count threshold before lossy context compaction kicks in |
| HERMES_PROMPT_TEMPLATE | default | prompt template name loaded from pkg/prompttemplates |

## running

### direct (Go Launcher)
```bash
./hermes-desktop.exe
```

### Electron Desktop App
```bash
cd apps/desktop
npm run dev
```
Launches the native Electron desktop shell with the integrated system tray icon and background server lifecycle management.

### via start.bat
```cmd
start.bat        # build and run (default)
start.bat run    # explicit run
start.bat build  # build only
start.bat test   # run go tests
start.bat clean  # remove binary
```

## dashboard access
once running, open http://127.0.0.1:9120 in a browser.

## submodules
this project does not use git submodules.

## upstream sync
```bash
git fetch upstream
git merge upstream/main
# resolve conflicts if any
git push origin main
```