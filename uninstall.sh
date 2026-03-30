#!/usr/bin/env bash
set -euo pipefail

echo "Stopping daemon (if running)..."
cli-agent daemon stop 2>/dev/null || true

echo "Unlinking gemini-runner..."
npm unlink -g cli-agent

echo "Done! cli-agent has been removed."
