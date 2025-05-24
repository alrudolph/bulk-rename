#!/bin/bash

set -e

REPO="alrudolph/bulk-rename"
BIN_NAME="bulk-rename"
VERSION="${1:-latest}"

OS=$(uname | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    aarch64 | arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

if [ "$VERSION" = "latest" ]; then
    VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | cut -d'"' -f4)
fi

INSTALL_DIR="$HOME/.local/bin"
mkdir -p "$INSTALL_DIR"
curl -L "https://github.com/$REPO/releases/download/$VERSION/$BIN_NAME-$OS-$ARCH" -o "$INSTALL_DIR/$BIN_NAME"
chmod +x "$INSTALL_DIR/$BIN_NAME"

if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
    echo "export PATH=\"\$PATH:$INSTALL_DIR\"" >> "$HOME/.bashrc"
    echo "Added $INSTALL_DIR to PATH. Restart your shell or run 'source ~/.bashrc'"
fi

echo "$BIN_NAME installed successfully to $INSTALL_DIR"
