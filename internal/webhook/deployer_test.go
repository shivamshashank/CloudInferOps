package webhook

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shivamshashank/CloudInferOps/internal/config"
)

func TestDeployWebhookHandlerDryRun(t *testing.T) {
	config.GlobalConfig = config.DefaultConfig()
	err := DeployWebhookHandler(true)
	if err != nil {
		t.Errorf("expected no error under dry-run, got: %v", err)
	}
}

func TestDeployWebhookHandlerMockedPATH(t *testing.T) {
	// Create mock bin directory
	tmpDir := t.TempDir()
	mockBinDir := filepath.Join(tmpDir, "bin")
	if err := os.MkdirAll(mockBinDir, 0755); err != nil {
		t.Fatalf("failed to create mock bin dir: %v", err)
	}

	// Write mock 'sh'
	mockShPath := filepath.Join(mockBinDir, "sh")
	mockShContent := `#!/bin/sh
if echo "$*" | grep -q "kubectl apply"; then
    exit 0
fi
exit 0
`
	if err := os.WriteFile(mockShPath, []byte(mockShContent), 0755); err != nil {
		t.Fatalf("failed to write mock sh: %v", err)
	}

	// Prepend mock bin dir to PATH
	oldPath := os.Getenv("PATH")
	t.Setenv("PATH", mockBinDir+":"+oldPath)

	config.GlobalConfig = config.DefaultConfig()
	err := DeployWebhookHandler(false)
	if err != nil {
		t.Errorf("expected no error with mocked PATH, got: %v", err)
	}

	// Test failing case
	mockShFailContent := `#!/bin/sh
echo "mock kubectl apply failed" >&2
exit 1
`
	if err := os.WriteFile(mockShPath, []byte(mockShFailContent), 0755); err != nil {
		t.Fatalf("failed to rewrite mock sh: %v", err)
	}

	err = DeployWebhookHandler(false)
	if err == nil {
		t.Error("expected error under failing kubectl apply, got nil")
	} else if !strings.Contains(err.Error(), "mock kubectl apply failed") {
		t.Errorf("expected error containing 'mock kubectl apply failed', got: %v", err)
	}
}
