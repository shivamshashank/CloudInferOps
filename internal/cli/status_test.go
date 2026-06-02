package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestStatusCommand(t *testing.T) {
	tmpDir := t.TempDir()
	mockBinDir := filepath.Join(tmpDir, "bin")
	if err := os.MkdirAll(mockBinDir, 0755); err != nil {
		t.Fatal(err)
	}

	mockKubectlPath := filepath.Join(mockBinDir, "kubectl")
	mockKubectlContent := `#!/bin/sh
echo "mock-context"
exit 0
`
	if err := os.WriteFile(mockKubectlPath, []byte(mockKubectlContent), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", mockBinDir+":"+os.Getenv("PATH"))

	_ = statusCmd.RunE(&cobra.Command{}, []string{})
}

func TestStatusCommand_NoK8s(t *testing.T) {
	tmpDir := t.TempDir()
	mockBinDir := filepath.Join(tmpDir, "bin")
	if err := os.MkdirAll(mockBinDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Mock kubectl to fail, triggering the "Kubernetes cluster not detected" flow
	mockKubectlPath := filepath.Join(mockBinDir, "kubectl")
	mockKubectlContent := `#!/bin/sh
exit 1
`
	if err := os.WriteFile(mockKubectlPath, []byte(mockKubectlContent), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", mockBinDir+":"+os.Getenv("PATH"))

	err := statusCmd.RunE(&cobra.Command{}, []string{})
	if err == nil {
		t.Error("expected error when k8s is unreachable")
	}
}
