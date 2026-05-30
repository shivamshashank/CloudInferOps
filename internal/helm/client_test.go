package helm

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHelmOperationsDryRun(t *testing.T) {
	// Test happy paths under dry-run
	if err := AddRepo("test-repo", "https://example.com/charts", true); err != nil {
		t.Errorf("expected no error under dry-run AddRepo, got: %v", err)
	}

	if err := UpdateRepos(true); err != nil {
		t.Errorf("expected no error under dry-run UpdateRepos, got: %v", err)
	}

	if err := InstallRelease("test-release", "test-repo/test-chart", "default", []string{"--set", "key=value"}, true); err != nil {
		t.Errorf("expected no error under dry-run InstallRelease, got: %v", err)
	}
}

func TestHelmOperationsMockedPATH(t *testing.T) {
	// Create mock bin directory
	tmpDir := t.TempDir()
	mockBinDir := filepath.Join(tmpDir, "bin")
	if err := os.MkdirAll(mockBinDir, 0755); err != nil {
		t.Fatalf("failed to create mock bin dir: %v", err)
	}

	// Write mock 'helm' script
	mockHelmPath := filepath.Join(mockBinDir, "helm")
	mockHelmContent := `#!/bin/sh
# Echo arguments so we can verify if needed, and exit 0
echo "mock helm executed with: $*"
exit 0
`
	if err := os.WriteFile(mockHelmPath, []byte(mockHelmContent), 0755); err != nil {
		t.Fatalf("failed to write mock helm: %v", err)
	}

	// Prepend mock bin dir to PATH
	oldPath := os.Getenv("PATH")
	t.Setenv("PATH", mockBinDir+":"+oldPath)

	// 1. Happy path (exit 0)
	if err := AddRepo("test-repo", "https://example.com/charts", false); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	if err := UpdateRepos(false); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	if err := InstallRelease("test-release", "test-repo/test-chart", "default", []string{"--set", "key=value"}, false); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	// 2. Failure path (exit 1)
	mockHelmFailContent := `#!/bin/sh
echo "mock helm failed command" >&2
exit 1
`
	if err := os.WriteFile(mockHelmPath, []byte(mockHelmFailContent), 0755); err != nil {
		t.Fatalf("failed to rewrite mock helm: %v", err)
	}

	if err := AddRepo("test-repo", "https://example.com/charts", false); err == nil {
		t.Error("expected error from AddRepo, got nil")
	} else if !strings.Contains(err.Error(), "mock helm failed command") {
		t.Errorf("expected error containing 'mock helm failed command', got: %v", err)
	}

	if err := UpdateRepos(false); err == nil {
		t.Error("expected error from UpdateRepos, got nil")
	} else if !strings.Contains(err.Error(), "mock helm failed command") {
		t.Errorf("expected error containing 'mock helm failed command', got: %v", err)
	}

	if err := InstallRelease("test-release", "test-repo/test-chart", "default", nil, false); err == nil {
		t.Error("expected error from InstallRelease, got nil")
	} else if !strings.Contains(err.Error(), "mock helm failed command") {
		t.Errorf("expected error containing 'mock helm failed command', got: %v", err)
	}
}
