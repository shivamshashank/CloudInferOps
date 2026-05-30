#!/usr/bin/env bash

set -euo pipefail

# Bash text formatting colors
RED='\033[31m'
GREEN='\033[32m'
YELLOW='\033[33m'
BLUE='\033[34m'
CYAN='\033[36m'
BOLD='\033[1m'
RESET='\033[0m'

# Prefixes
INFO="  🔵 ${CYAN}[INFO]${RESET} "
SUCCESS="  🟢 ${GREEN}[OK]${RESET} "
WARN="  🟡 ${YELLOW}[WARN]${RESET} "
ERROR="  🔴 ${RED}[ERROR]${RESET} "
READY="  🚀 ${GREEN}${BOLD}[READY]${RESET} "

echo -e "${BOLD}StackPulse Installer${RESET}"
echo -e "---------------------"

# 1. System environment detection
echo -e "${INFO}Detecting system environment..."
OS_NAME=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH_NAME=$(uname -m)

case "${OS_NAME}" in
    linux)
        OS="linux"
        ;;
    darwin)
        echo -e "${ERROR}StackPulse no longer supports macOS (darwin) natively."
        echo -e "${INFO}To test locally on macOS, please run inside a Linux VM (e.g., using Multipass)."
        exit 1
        ;;
    *)
        echo -e "${ERROR}Unsupported operating system: ${OS_NAME}"
        echo -e "${ERROR}StackPulse only supports Linux."
        exit 1
        ;;
esac

case "${ARCH_NAME}" in
    x86_64|amd64)
        ARCH="amd64"
        ;;
    arm64|aarch64)
        ARCH="arm64"
        ;;
    *)
        echo -e "${ERROR}Unsupported CPU architecture: ${ARCH_NAME}"
        echo -e "${ERROR}StackPulse supports x86_64 (amd64) and arm64."
        exit 1
        ;;
esac

echo -e "${SUCCESS}System matches: ${OS}/${ARCH}"

# 2. Retrieve version
echo -e "${INFO}Fetching latest StackPulse version..."
LATEST_TAG=$(curl -s "https://api.github.com/repos/shivamshashank/StackPulse/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' || true)

if [ -z "${LATEST_TAG}" ]; then
    VERSION="v0.1.0"
    echo -e "${WARN}Failed to fetch latest version from GitHub API. Falling back to default: ${VERSION}"
else
    VERSION="${LATEST_TAG}"
    echo -e "${SUCCESS}Latest release detected: ${VERSION}"
fi

# 3. Download binary
BINARY_NAME="stackpulse-${OS}-${ARCH}"
DOWNLOAD_URL="https://github.com/shivamshashank/StackPulse/releases/download/${VERSION}/${BINARY_NAME}"
TEMP_DIR=$(mktemp -d)
TEMP_BIN="${TEMP_DIR}/stackpulse"

echo -e "${INFO}Downloading binary from: ${DOWNLOAD_URL}"
# Try downloading the compiled binary
if curl -sSLf -o "${TEMP_BIN}" "${DOWNLOAD_URL}"; then
    echo -e "${SUCCESS}Successfully downloaded StackPulse release ${VERSION}."
else
    # Fallback/Development mock helper for local installations before releases are tagged
    echo -e "${WARN}Release asset not found on GitHub yet (project is in pre-release stage)."
    echo -e "${INFO}Constructing fallback binary builder..."
    if command -v go &>/dev/null; then
        echo -e "${INFO}Compiling local binary via 'go build'..."
        # If in repository context, build locally
        if [ -f "cmd/stackpulse/main.go" ]; then
            COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
            BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || date +%Y-%m-%dT%H:%M:%SZ)
            PKG="github.com/shivamshashank/StackPulse/internal/cli"
            go build -ldflags="-s -w -X ${PKG}.Version=${VERSION} -X ${PKG}.Commit=${COMMIT} -X ${PKG}.BuildDate=${BUILD_DATE}" -o "${TEMP_BIN}" cmd/stackpulse/main.go
            echo -e "${SUCCESS}Compiled local StackPulse binary successfully."
        else
            echo -e "${ERROR}Cannot install: No release asset exists yet, and not in StackPulse repository source folder."
            rm -rf "${TEMP_DIR}"
            exit 1
        fi
    else
        echo -e "${ERROR}Cannot install: Download failed and 'go' compiler is not installed to compile locally."
        rm -rf "${TEMP_DIR}"
        exit 1
    fi
fi

# 4. Install binary to /usr/local/bin
INSTALL_PATH="/usr/local/bin/stackpulse"
echo -e "${INFO}Installing to ${INSTALL_PATH}..."

# Check write permissions, use sudo if needed
if [ -w "/usr/local/bin" ]; then
    mv "${TEMP_BIN}" "${INSTALL_PATH}"
else
    echo -e "${INFO}Requesting administrator permissions to copy to /usr/local/bin..."
    sudo mv "${TEMP_BIN}" "${INSTALL_PATH}"
fi

chmod +x "${INSTALL_PATH}"
rm -rf "${TEMP_DIR}"

echo -e "${SUCCESS}StackPulse binary installed to ${INSTALL_PATH} successfully."

# 5. Verify run
echo -e "${INFO}Verifying installation..."
INSTALLED_VER=$("${INSTALL_PATH}" version | grep "version:" | awk '{print $NF}' || echo "unknown")

echo -e "\n${READY}${BOLD}Successfully installed StackPulse!${RESET}"
echo -e "  🚀 Installed version: ${INSTALLED_VER}"
echo -e "  🩺 Run '${BOLD}stackpulse doctor${RESET}' to test your system environment."
