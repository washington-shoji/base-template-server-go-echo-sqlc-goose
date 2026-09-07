#!/usr/bin/env bash
# Download and verify htmx into web/static/js/htmx.min.js
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
# shellcheck disable=SC1091
source "$ROOT/tools/versions.env"

URL="https://unpkg.com/htmx.org@${HTMX_VERSION}/dist/htmx.min.js"
DEST="$ROOT/web/static/js/htmx.min.js"
mkdir -p "$(dirname "$DEST")"
TMP="$(mktemp)"
trap 'rm -f "$TMP"' EXIT

echo "Downloading htmx v${HTMX_VERSION}..."
curl -fsSL "$URL" -o "$TMP"

ACTUAL="$(sha256sum "$TMP" | awk '{print $1}')"
if [[ "${HTMX_SHA256}" != "PLACEHOLDER" && -n "${HTMX_SHA256}" ]]; then
  if [[ "$ACTUAL" != "$HTMX_SHA256" ]]; then
    echo "htmx checksum mismatch" >&2
    echo "  expected: $HTMX_SHA256" >&2
    echo "  actual:   $ACTUAL" >&2
    exit 1
  fi
else
  echo "HTMX_SHA256 not pinned yet; recorded actual: $ACTUAL"
fi

install -m 0644 "$TMP" "$DEST"
echo "Installed $DEST"
