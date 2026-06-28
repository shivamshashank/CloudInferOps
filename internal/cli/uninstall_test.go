package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUninstallK8sCommand(t *testing.T) {
	// Create a temporary directory for mock binaries
	tmpDir := t.TempDir()
	mockBinDir := filepath.Join(tmpDir, "bin")
	if err := os.MkdirAll(mockBinDir, 0755); err != nil {
		t.Fatalf("Failed to create mock bin dir: %v", err)
	}

	// Create mock binaries that will be "uninstalled"
	for _, bin := range []string{"kubectl"} {
		path := filepath.Join(mockBinDir, bin)
		if err := os.WriteFile(path, []byte("#!/bin/sh\necho 'mock "+bin+"'"), 0755); err != nil {
			t.Fatalf("Failed to create mock binary %s: %v", bin, err)
		}
	}

	// Override PATH to use our mock binaries
	originalPath := os.Getenv("PATH")
	t.Setenv("PATH", mockBinDir+":"+originalPath)
	t.Setenv("HOME", tmpDir) // Set a mock home to avoid touching real user config

	// Mock the functions that perform the actual uninstallation to avoid side effects
	originalPerformUninstallBinaries := performUninstallBinaries
	defer func() {
		performUninstallBinaries = originalPerformUninstallBinaries
	}()

	var binariesCalled bool
	performUninstallBinaries = func(dryRun bool) error {
		binariesCalled = true
		return nil
	}

	// Set forceUninstall to true to bypass prompt and run the command's RunE directly
	forceUninstall = true
	defer func() { forceUninstall = false }() // Reset flag

	err := uninstallK8sCmd.RunE(uninstallK8sCmd, []string{})
	if err != nil {
		t.Fatalf("uninstall k8s command failed: %v", err)
	}

	if !binariesCalled {
		t.Error("expected performUninstallBinaries to be called, but it wasn't")
	}
}
