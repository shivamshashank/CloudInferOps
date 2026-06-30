#!/usr/bin/env bash

set -euo pipefail

HOOK_DIR=".git/hooks"
PRE_COMMIT_FILE="${HOOK_DIR}/pre-commit"

echo "Setting up Git pre-commit hook..."

mkdir -p "${HOOK_DIR}"

cat << 'EOF' > "${PRE_COMMIT_FILE}"
#!/usr/bin/env bash

# Add common paths for GUI Git clients that don't load the user's PATH
export PATH=$PATH:/usr/local/go/bin:/usr/local/bin:/opt/homebrew/bin:$HOME/go/bin

echo "Running go fmt..."
if ! go fmt ./...; then
    echo "🔴 go fmt failed! Please format your code."
    exit 1
fi

echo "Running go vet..."
if ! go vet ./...; then
    echo "🔴 go vet failed! Please fix the issues."
    exit 1
fi

if command -v golangci-lint &> /dev/null; then
    echo "Running golangci-lint..."
    if ! golangci-lint run; then
        echo "🔴 golangci-lint failed! Please fix the linting errors before committing."
        exit 1
    fi
else
    echo "🟡 golangci-lint not installed locally. Skipping lint check. (Run 'brew install golangci-lint' to install)"
fi
EOF

chmod +x "${PRE_COMMIT_FILE}"
echo "✅ Pre-commit hook installed successfully at ${PRE_COMMIT_FILE}!"
