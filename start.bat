@echo off
setlocal enabledelayedexpansion
title Hermes Agent Go

:: ═══════════════════════════════════════════════════════════════
:: Hermes Agent Go - The Self-Improving AI Agent
:: Module:  github.com/robertpelloni/hermes-agent
:: Entry:    cmd/hermes/main.go
:: binary: hermes-desktop.exe
:: ═══════════════════════════════════════════════════════════════

cd /d "%~dp0"

set "BINARY=hermes-desktop.exe"
set "ENTRY=./cmd/hermes"

:: ─── Parse command ──────────────────────────────────────────
set "CMD=%1"
if "%CMD%"=="" set "CMD=run"
if /i "%CMD%"=="run"      goto :run
if /i "%CMD%"=="build"    goto :build
if /i "%CMD%"=="tui"      goto :tui
if /i "%CMD%"=="gateway"  goto :gateway
if /i "%CMD%"=="test"     goto :test
if /i "%CMD%"=="clean"    goto :clean
if /i "%CMD%"=="help"     goto :help
echo Unknown command: %CMD%
goto :help

:: ─── Build ──────────────────────────────────────────────────
:build
echo.
echo  [Hermes Agent Go] Building...
go mod download
if errorlevel 1 ( echo  [FAIL] Dependency download & exit /b 1 )
go build -buildvcs=false -ldflags="-s -w" -o %BINARY% %ENTRY%
if errorlevel 1 ( echo  [FAIL] Build failed & exit /b 1 )
for %%f in (%BINARY%) do echo  [OK]   %%~zf bytes
goto :end

:: ─── Run ────────────────────────────────────────────────────
:run
if not exist %BINARY% call :build
if errorlevel 1 exit /b 1
echo.
echo  [Hermes Agent Go] Starting...
echo  Model: %HERMES_MODEL%
echo  Provider: %HERMES_PROVIDER%
echo.
%BINARY%
goto :end

:: ─── TUI Mode ───────────────────────────────────────────────
:tui
if not exist %BINARY% call :build
if errorlevel 1 exit /b 1
echo  [Hermes Agent Go] Interactive TUI
%BINARY% --tui
goto :end

:: ─── Gateway Mode ───────────────────────────────────────────
:gateway
if not exist %BINARY% call :build
if errorlevel 1 exit /b 1
echo  [Hermes Agent Go] Gateway only (Telegram/Discord/CLI)
%BINARY% --gateway
goto :end

:: ─── Test ───────────────────────────────────────────────────
:test
echo  [Hermes Agent Go] Running tests...
go test ./pkg/... ./cmd/... -v -count=1 -timeout 120s
goto :end

:: ─── Clean ──────────────────────────────────────────────────
:clean
del /q %BINARY% 2>nul
go clean
echo  [Hermes Agent Go] Cleaned.
goto :end

:: ─── Help ───────────────────────────────────────────────────
:help
echo.
echo  Hermes Agent Go - Usage: start.bat [command]
echo.
echo  Commands:
echo    run       Build and run agent (default)
echo    build     Build binary only
echo    tui       Interactive TUI mode
echo    tui is now the same as run - desktop launches both dashboard and tui
echo    gateway   Start gateway listeners only
echo    test      Run tests
echo    clean     Remove binary
echo    help      Show this help
echo.
echo  Packages:
echo    pkg/agent      Agent core logic
echo    pkg/gateway    Telegram/Discord/CLI gateways
echo    pkg/mcp        Model Context Protocol
echo    pkg/memory     Persistent memory
echo    pkg/scheduler  Task scheduling
echo    pkg/skill      Skill system
echo.
echo  Environment:
echo    HERMES_MODEL      Default model name
echo    HERMES_PROVIDER   Default provider (openrouter)
echo    OPENAI_API_KEY    OpenAI API key
echo    ANTHROPIC_API_KEY Anthropic API key
echo.
goto :end

:end
endlocal
