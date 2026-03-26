#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo "Building gemini-runner..."
cd "$PROJECT_DIR"
npm run build

echo "Linking globally..."
npm link

echo "Done. Run 'gemini-runner help' to verify."
