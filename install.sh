#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

echo "Installing dependencies..."
npm install

echo "Building gemini-runner..."
npm run build

echo "Linking globally..."
npm link

echo "Done! Run 'gemini-runner help' to verify."
