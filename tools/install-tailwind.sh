#!/usr/bin/env bash
# Download and verify the Tailwind CSS standalone CLI into .tools/
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
# shellcheck disable=SC1091
source "$ROOT/tools/versions.env"

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$OS" in
  linux)  PLATFORM=linux ;;
  darwin) PLATFORM=macos ;;
  *) echo "unsupported OS: $OS" >&2; exit 1 ;;
esac
case "$ARCH" in
  x86_64|amd64) ARCH_TAG=x64 ;;
  arm64|aarch64) ARCH_TAG=arm64 ;;
  *) echo "unsupported arch: $ARCH" >&2; exit 1 ;;
esac

ASSET="tailwindcss-${PLATFORM}-${ARCH_TAG}"
URL="https://github.com/tailwindlabs/tailwindcss/releases/download/v${TAILWIND_VERSION}/${ASSET}"
SUMS_URL="https://github.com/tailwindlabs/tailwindcss/releases/download/v${TAILWIND_VERSION}/sha256sums.txt"

DEST_DIR="$ROOT/.tools"
DEST="$DEST_DIR/tailwindcss"
mkdir -p "$DEST_DIR"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

echo "Downloading Tailwind CSS v${TAILWIND_VERSION} (${ASSET})..."
curl -fsSL "$URL" -o "$TMP/$ASSET"
curl -fsSL "$SUMS_URL" -o "$TMP/sha256sums.txt"

# Release checksums list paths like "./tailwindcss-linux-x64"
EXPECTED="$(awk -v f="$ASSET" '$2 == f || $2 == "./" f { print $1; exit }' "$TMP/sha256sums.txt")"
if [[ -z "$EXPECTED" ]]; then
  echo "checksum not found for $ASSET in sha256sums.txt" >&2
  cat "$TMP/sha256sums.txt" >&2
  exit 1
fi

ACTUAL="$(sha256sum "$TMP/$ASSET" | awk '{print $1}')"
if [[ "$ACTUAL" != "$EXPECTED" ]]; then
  echo "checksum mismatch for $ASSET" >&2
  echo "  expected: $EXPECTED" >&2
  echo "  actual:   $ACTUAL" >&2
  exit 1
fi

install -m 0755 "$TMP/$ASSET" "$DEST"
echo "Installed $DEST (sha256 $ACTUAL)"
