package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shivamshashank/StackPulse/internal/config"

	"github.com/spf13/cobra"
)

func TestConnectCommand_NoK8s(t *testing.T) {
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

	err := connectCmd.RunE(&cobra.Command{}, []string{})
	if err == nil {
		t.Error("expected error when k8s is unreachable")
	}
}

func TestConnectCommand_MockSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	mockBinDir := filepath.Join(tmpDir, "bin")
	if err := os.MkdirAll(mockBinDir, 0755); err != nil {
		t.Fatal(err)
	}

	mockKubectlPath := filepath.Join(mockBinDir, "kubectl")
	mockKubectlContent := `#!/bin/sh
echo "mocked-secret-string"
exit 0
`
	if err := os.WriteFile(mockKubectlPath, []byte(mockKubectlContent), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", mockBinDir+":"+os.Getenv("PATH"))

	// Override the CLI flag so we don't accidentally pop open the CI server's web browser
	connectBrowser = false
	_ = connectCmd.RunE(&cobra.Command{}, []string{})
}

func TestConnectCommand_WithArgoCD(t *testing.T) {
	tmpDir := t.TempDir()
	mockBinDir := filepath.Join(tmpDir, "bin")
	if err := os.MkdirAll(mockBinDir, 0755); err != nil {
		t.Fatal(err)
	}

	// This mock simulates success for kubectl calls including ArgoCD and Grafana secrets
	mockKubectlPath := filepath.Join(mockBinDir, "kubectl")
	mockKubectlContent := `#!/bin/sh
if echo "$*" | grep -q "stackpulse-prometheus-grafana"; then
    echo "cGFzc3dvcmQ=" # "password" in base64
    exit 0
fi
if echo "$*" | grep -q "-o name"; then
    echo "secret/argocd-initial-admin-secret"
    exit 0
fi
if echo "$*" | grep -q "argocd-initial-admin-secret"; then
    echo "YXJnb3Bhc3M=" # "argopass" in base64
    exit 0
fi
echo "mocked-k8s-response"
exit 0
`
	if err := os.WriteFile(mockKubectlPath, []byte(mockKubectlContent), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", mockBinDir+":"+os.Getenv("PATH"))

	config.GlobalConfig = config.DefaultConfig()
	config.GlobalConfig.Observability.ArgoCD = true

	connectBrowser = false
	err := connectCmd.RunE(&cobra.Command{}, []string{})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestConnectCommand_BrowserExecution(t *testing.T) {
	tmpDir := t.TempDir()
	mockBinDir := filepath.Join(tmpDir, "bin")
	if err := os.MkdirAll(mockBinDir, 0755); err != nil {
		t.Fatal(err)
	}

	mockKubectlPath := filepath.Join(mockBinDir, "kubectl")
	if err := os.WriteFile(mockKubectlPath, []byte(`#!/bin/sh
echo "mock-context"
exit 0
`), 0755); err != nil {
		t.Fatal(err)
	}

	mockXdgOpenPath := filepath.Join(mockBinDir, "xdg-open")
	if err := os.WriteFile(mockXdgOpenPath, []byte(`#!/bin/sh
exit 0
`), 0755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("PATH", mockBinDir+":"+os.Getenv("PATH"))

	connectBrowser = true
	_ = connectCmd.RunE(&cobra.Command{}, []string{})
}
