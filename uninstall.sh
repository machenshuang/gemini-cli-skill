#!/usr/bin/env bash
set -euo pipefail

echo "Stopping daemon (if running)..."
gemini-runner daemon stop 2>/dev/null || true

echo "Unlinking gemini-runner..."
npm unlink -g gemini-runner

echo "Done! gemini-runner has been removed."
