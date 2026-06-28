#!/usr/bin/env bash
set -euo pipefail

RED='\033[31m'
GREEN='\033[32m'
CYAN='\033[36m'
BOLD='\033[1m'
RESET='\033[0m'

info(){ echo -e "🔵 ${CYAN}$*${RESET}"; }
ok(){ echo -e "🟢 ${GREEN}$*${RESET}"; }
err(){ echo -e "🔴 ${RED}$*${RESET}"; }

[ "$(uname -s)" = "Linux" ] || { err "Linux only."; exit 1; }

case "$(uname -m)" in
  x86_64|amd64) ARCH=amd64;;
  arm64|aarch64) ARCH=arm64;;
  *) err "Unsupported architecture."; exit 1;;
esac

BIN="cloudinferops-linux-${ARCH}"
BASE="https://github.com/shivamshashank/CloudInferOps/releases/latest/download"

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

info "Downloading binary..."
curl -fsSL "${BASE}/${BIN}" -o "${TMP}/cloudinferops"

# info "Downloading checksums..."
# curl -fsSL "${BASE}/checksums.txt" -o "${TMP}/checksums.txt"

# info "Verifying checksum..."
# (
# cd "${TMP}"
# grep " ${BIN}$" checksums.txt | sed "s#${BIN}#cloudinferops#" | sha256sum -c -
# )

chmod +x "${TMP}/cloudinferops"

DEST=/usr/local/bin/cloudinferops
info "Installing..."

if [ -w /usr/local/bin ]; then
  mv "${TMP}/cloudinferops" "$DEST"
else
  sudo mv "${TMP}/cloudinferops" "$DEST"
fi

chmod +x "$DEST"

ok "CloudInferOps installed."

echo
cloudinferops version || true

echo
echo "Next steps:"
echo "  sudo cloudinferops doctor"
echo "  sudo cloudinferops deploy observability"
echo "  sudo cloudinferops status"
