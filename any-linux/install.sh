#!/usr/bin/env bash
# commit-ai
# Copyright (c) 2026 Jonathan Henrique Perozi Lourenço (jhowk14)
# Licensed under the MIT License

set -e

# ===============================================
# commit-ai Installer for Linux
# ===============================================

VERSION="1.2.0"
INSTALL_DIR="/usr/local/bin"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_FILE="$HOME/.commit-ai.conf"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  🤖 commit-ai v$VERSION - Linux Installer"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo

# Check dependencies
echo "📦 Checking dependencies..."
MISSING=""
for cmd in git jq curl; do
  if ! command -v "$cmd" &> /dev/null; then
    MISSING="$MISSING $cmd"
  fi
done

if [ -n "$MISSING" ]; then
  echo "❌ Missing dependencies:$MISSING"
  echo
  echo "Install them with:"
  echo "  Ubuntu/Debian: sudo apt install$MISSING"
  echo "  Fedora:        sudo dnf install$MISSING"
  echo "  Arch:          sudo pacman -S$MISSING"
  exit 1
fi
echo "✅ All dependencies found"
echo

# Install script
echo "📥 Installing commit-ai to $INSTALL_DIR..."
if [ -w "$INSTALL_DIR" ]; then
  cp "$SCRIPT_DIR/commit-ai.sh" "$INSTALL_DIR/commit-ai"
  chmod +x "$INSTALL_DIR/commit-ai"
else
  echo "⚠️  Requires sudo for system-wide installation"
  sudo cp "$SCRIPT_DIR/commit-ai.sh" "$INSTALL_DIR/commit-ai"
  sudo chmod +x "$INSTALL_DIR/commit-ai"
fi
echo "✅ Script installed to $INSTALL_DIR/commit-ai"
echo

# Run interactive setup
echo "🔧 Running initial configuration..."
echo
"$INSTALL_DIR/commit-ai" --setup

echo
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  ✅ Installation complete!"
echo "  Run 'commit-ai' to generate commit messages"
echo "  Run 'commit-ai --setup' to reconfigure"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
