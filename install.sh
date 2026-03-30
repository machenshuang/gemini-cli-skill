#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

echo "Installing dependencies..."
npm install

echo "Building cli-agent..."
npm run build

echo "Linking globally..."
npm link

echo "Installing Claude Code skills..."
mkdir -p ~/.claude/skills/gemini
mkdir -p ~/.claude/skills/kimi
cp skills/gemini/SKILL.md ~/.claude/skills/gemini/SKILL.md
cp skills/kimi/SKILL.md ~/.claude/skills/kimi/SKILL.md

echo "Done! Run 'cli-agent help' to verify."
