@echo off
setlocal enabledelayedexpansion
title Hermes Agent

:: =============================================================================
:: Hermes Agent — Multi-Harness Launcher & Build Automation
:: Module:  github.com/robertpelloni/hermes-agent
:: Entry:   cmd/hermes/main.go & Python / Web Dashboards
:: =============================================================================

cd /d "%~dp0"

set "BINARY=hermes-desktop.exe"
set "ENTRY=./cmd/hermes"

:: ─── Parse command ──────────────────────────────────────────
set "CMD=%1"
if "%CMD%"=="" set "CMD=run"
if /i "%CMD%"=="run"        goto :run
if /i "%CMD%"=="build"      goto :build
if /i "%CMD%"=="build-all"  goto :build_all
if /i "%CMD%"=="web"        goto :web
if /i "%CMD%"=="tui"        goto :tui
if /i "%CMD%"=="gateway"    goto :gateway
if /i "%CMD%"=="cli"        goto :cli
if /i "%CMD%"=="foundation" goto :foundation
if /i "%CMD%"=="test"       goto :test
if /i "%CMD%"=="clean"      goto :clean
if /i "%CMD%"=="help"       goto :help
echo Unknown command: %CMD%
goto :help

:: ─── Build Go ───────────────────────────────────────────────
:build
echo.
echo  [Hermes Agent] Building Go binary (%BINARY%)...
go mod download
if errorlevel 1 ( echo  [FAIL] Dependency download & exit /b 1 )
go build -buildvcs=false -ldflags="-s -w" -o %BINARY% %ENTRY%
if errorlevel 1 ( echo  [FAIL] Build failed & exit /b 1 )
for %%f in (%BINARY%) do echo  [OK]   %%~zf bytes
goto :end

:: ─── Build All (Go + Web) ───────────────────────────────────
:build_all
call :build
if errorlevel 1 exit /b 1
echo.
echo  [Hermes Agent] Building Web UI assets...
cd web
call npm run build
cd ..
if errorlevel 1 ( echo  [FAIL] Web build failed & exit /b 1 )
echo  [OK] All harnesses built successfully.
goto :end

:: ─── Web Dashboard ──────────────────────────────────────────
:web
echo  [Hermes Agent] Starting Web Dashboard...
if exist %BINARY% (
    %BINARY% --dashboard
) else (
    python -m hermes_cli.main web
)
goto :end

:: ─── Run Default ────────────────────────────────────────────
:run
if not exist %BINARY% call :build
if errorlevel 1 exit /b 1
echo.
echo  [Hermes Agent] Starting...
echo  Model: %HERMES_MODEL%
echo  Provider: %HERMES_PROVIDER%
echo.
%BINARY%
goto :end

:: ─── TUI Mode ───────────────────────────────────────────────
:tui
if not exist %BINARY% call :build
if errorlevel 1 exit /b 1
echo  [Hermes Agent] Interactive TUI
%BINARY% --tui
goto :end

:: ─── Gateway Mode ───────────────────────────────────────────
:gateway
if not exist %BINARY% call :build
if errorlevel 1 exit /b 1
echo  [Hermes Agent] Gateway only
%BINARY% --gateway
goto :end

:: ─── CLI Mode ───────────────────────────────────────────────
:cli
if not exist %BINARY% call :build
if errorlevel 1 exit /b 1
echo  [Hermes Agent] Running Python CLI
%BINARY% --cli
goto :end

:: ─── Foundation Mode ────────────────────────────────────────
:foundation
if not exist %BINARY% call :build
if errorlevel 1 exit /b 1
echo  [Hermes Agent] Foundation Mode
%BINARY% --foundation
goto :end

:: ─── Test ───────────────────────────────────────────────────
:test
echo  [Hermes Agent] Running tests...
go test ./pkg/... ./cmd/... -v -count=1 -timeout 120s
cd web
call npm test
cd ..
goto :end

:: ─── Clean ──────────────────────────────────────────────────
:clean
del /q %BINARY% 2>nul
go clean
echo  [Hermes Agent] Cleaned.
goto :end

:: ─── Help ───────────────────────────────────────────────────
:help
echo.
echo  Hermes Agent - Usage: start.bat [command]
echo.
echo  Commands:
echo    run         Build and run agent (default)
echo    build       Build Go binary only
echo    build-all   Build Go binary and Web UI distribution
echo    web         Start Python/Go web dashboard
echo    tui         Interactive TUI mode
echo    gateway     Start gateway listeners only
echo    cli         Run Python hermes CLI
echo    foundation  Show foundation integration features
echo    test        Run Go and Web test suites
echo    clean       Remove binary
echo    help        Show this help
echo.
goto :end

:end
endlocal
