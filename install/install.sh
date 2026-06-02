#!/bin/sh
# git-wtm installer script
# Usage: curl -fsSL https://raw.githubusercontent.com/aryanpnd/git-wtm/main/install/install.sh | sh

set -e

REPO="aryanpnd/git-wtm"
BINARY="git-wtm"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

case "$OS" in
  linux|darwin) ;;
  *) echo "Unsupported OS: $OS. Use Scoop on Windows."; exit 1 ;;
esac

# Get latest version
VERSION=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed -E 's/.*"v([^"]+)".*/\1/')

if [ -z "$VERSION" ]; then
  echo "Error: Could not determine latest version"
  exit 1
fi

# Download
URL="https://github.com/$REPO/releases/download/v${VERSION}/${BINARY}_${OS}_${ARCH}.tar.gz"
echo "Downloading git-wtm v${VERSION} for ${OS}/${ARCH}..."

TMP=$(mktemp -d)
curl -fsSL "$URL" -o "$TMP/git-wtm.tar.gz"
tar -xzf "$TMP/git-wtm.tar.gz" -C "$TMP"

# Install
INSTALL_DIR="/usr/local/bin"
if [ ! -w "$INSTALL_DIR" ]; then
  echo "Installing to $INSTALL_DIR (requires sudo)..."
  sudo mv "$TMP/$BINARY" "$INSTALL_DIR/$BINARY"
else
  mv "$TMP/$BINARY" "$INSTALL_DIR/$BINARY"
fi

chmod +x "$INSTALL_DIR/$BINARY"
rm -rf "$TMP"

echo "git-wtm v${VERSION} installed successfully!"
echo "Run 'git wtm' in any git repository to get started."
