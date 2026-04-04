#!/usr/bin/env bash
set -e

BINARY_NAME="cli-agent"
INSTALL_DIR="/usr/local/bin"

# Stop daemon if running
if command -v "$BINARY_NAME" &>/dev/null; then
  echo "Stopping daemon (if running)..."
  "$BINARY_NAME" daemon stop 2>/dev/null || true
fi

# Remove binary
if [ -f "$INSTALL_DIR/$BINARY_NAME" ]; then
  echo "Removing $INSTALL_DIR/$BINARY_NAME..."
  sudo rm -f "$INSTALL_DIR/$BINARY_NAME"
fi

# Remove skills
CLAUDE_SKILLS_DIR="$HOME/.claude/skills"
for skill in kimi gemini; do
  if [ -d "$CLAUDE_SKILLS_DIR/$skill" ]; then
    echo "Removing skill: $skill"
    rm -rf "$CLAUDE_SKILLS_DIR/$skill"
  fi
done

echo "Uninstall complete."
