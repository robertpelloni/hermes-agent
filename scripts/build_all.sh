#!/usr/bin/env bash
set -euo pipefail

# Hermes Agent — Unified Multi-Harness Build & Check Script
echo "=================================================="
echo "  Hermes Agent — Unified Multi-Harness Build"
echo "=================================================="

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

echo ">> 1. Building Go Harness (hermes-desktop)..."
if command -v go >/dev/null 2>&1; then
  go build -buildvcs=false -ldflags="-s -w" -o hermes-desktop ./cmd/hermes
  echo "   [OK] Go binary built successfully: $(ls -lh hermes-desktop | awk '{print $5}')"
else
  echo "   [WARN] Go not found in PATH, skipping Go build."
fi

echo ">> 2. Testing Go Subsystems..."
if command -v go >/dev/null 2>&1; then
  go test ./pkg/... ./cmd/... -count=1 -timeout 30s
  echo "   [OK] Go tests passed."
fi

echo ">> 3. Building Web UI Assets (web/)..."
if command -v npm >/dev/null 2>&1; then
  (cd web && npm run build)
  echo "   [OK] Web bundle built into hermes_cli/web_dist/."
else
  echo "   [WARN] npm not found in PATH, skipping Web build."
fi

echo ">> 4. Typechecking Desktop Electron App..."
if [ -d "apps/desktop" ] && command -v npm >/dev/null 2>&1; then
  (cd apps/desktop && npm run typecheck)
  echo "   [OK] Desktop typecheck passed."
fi

echo "=================================================="
echo "  Build Completed Successfully!"
echo "=================================================="
