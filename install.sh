#!/usr/bin/env bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BINARY_NAME="cli-agent"
INSTALL_DIR="/usr/local/bin"

echo "Building $BINARY_NAME..."
cd "$SCRIPT_DIR"
go build -o "$BINARY_NAME" .

echo "Installing $BINARY_NAME to $INSTALL_DIR..."
sudo mv "$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"

# Install Claude Code skills
SKILLS_SRC="$SCRIPT_DIR/skills"
if [ -d "$SKILLS_SRC" ]; then
  CLAUDE_SKILLS_DIR="$HOME/.claude/skills"
  mkdir -p "$CLAUDE_SKILLS_DIR"
  cp -r "$SKILLS_SRC/"* "$CLAUDE_SKILLS_DIR/"
  echo "Skills installed to $CLAUDE_SKILLS_DIR"
fi

echo "Done! Run '$BINARY_NAME --help' to get started."
