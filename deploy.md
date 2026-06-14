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
start.bat build   # build binary only
start.bat run     # build and run
```

## environment variables
| variable | default | description |
|----------|---------|-------------|
| HERMES_PORT | 9120 | dashboard port |
| HERMES_MODEL | (optional) | model name for agent |
| HERMES_PROVIDER | (optional) | provider name (e.g. openrouter) |

## running

### direct
```bash
./hermes-desktop.exe
```

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