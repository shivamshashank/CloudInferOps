package alerts

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateSecret(t *testing.T) {
	// Create mock bin directory
	tmpDir := t.TempDir()
	mockBinDir := filepath.Join(tmpDir, "bin")
	if err := os.MkdirAll(mockBinDir, 0755); err != nil {
		t.Fatalf("failed to create mock bin dir: %v", err)
	}

	// Write a mock 'sh' script
	mockShPath := filepath.Join(mockBinDir, "sh")
	mockShContent := `#!/bin/sh
if [ "$1" = "-c" ]; then
    if echo "$2" | grep -q "kubectl create secret"; then
        echo "mock success"
        exit 0
    else
        echo "mock unexpected: $2" >&2
        exit 2
    fi
fi
echo "mock sh called without -c" >&2
exit 1
`
	if err := os.WriteFile(mockShPath, []byte(mockShContent), 0755); err != nil {
		t.Fatalf("failed to write mock sh: %v", err)
	}

	// Save original PATH
	oldPath := os.Getenv("PATH")
	// Prepend mockBinDir to PATH
	t.Setenv("PATH", mockBinDir+":"+oldPath)

	// 1. Success case
	err := CreateSecret("test-secret", "key", "value", "test-namespace")
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	// 2. Failure case: Write a failing mock sh
	mockShFailContent := `#!/bin/sh
echo "mock kubernetes connection failed" >&2
exit 1
`
	if err := os.WriteFile(mockShPath, []byte(mockShFailContent), 0755); err != nil {
		t.Fatalf("failed to rewrite mock sh: %v", err)
	}

	err = CreateSecret("test-secret", "key", "value", "test-namespace")
	if err == nil {
		t.Error("expected error from failing sh execution, got nil")
	} else if !strings.Contains(err.Error(), "mock kubernetes connection failed") {
		t.Errorf("expected error message containing 'mock kubernetes connection failed', got: %v", err)
	}
}
