package observability

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupMocks(t *testing.T) {
	tmpDir := t.TempDir()
	mockBinDir := filepath.Join(tmpDir, "bin")
	if err := os.MkdirAll(mockBinDir, 0755); err != nil {
		t.Fatal(err)
	}

	// We mock kubectl to output an empty JSON list so 'get svc' parsing doesn't crash
	mockScript := `#!/bin/sh
echo '{"items":[]}' 
exit 0
`
	if err := os.WriteFile(filepath.Join(mockBinDir, "kubectl"), []byte(mockScript), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mockBinDir, "helm"), []byte(mockScript), 0755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("PATH", mockBinDir+":"+os.Getenv("PATH"))
}

func TestCreateThanosSecret_DryRun(t *testing.T) {
	err := createThanosSecret("test-ns", true)
	if err != nil {
		t.Errorf("expected no error on dry-run, got %v", err)
	}
}

func TestApplyObservabilityIngress_DryRun(t *testing.T) {
	err := applyObservabilityIngress("test-ns", true)
	if err != nil {
		t.Errorf("expected no error on dry-run, got %v", err)
	}
}

func TestApplyObservabilityIngress_LiveMock(t *testing.T) {
	setupMocks(t)
	oldMax := serviceDiscoveryMaxRetries
	oldDelay := serviceDiscoveryRetryDelay
	serviceDiscoveryMaxRetries = 2
	serviceDiscoveryRetryDelay = 1 * time.Millisecond
	defer func() {
		serviceDiscoveryMaxRetries = oldMax
		serviceDiscoveryRetryDelay = oldDelay
	}()

	err := applyObservabilityIngress("test-ns", false)
	if err != nil {
		t.Errorf("expected no error on live mock, got %v", err)
	}
}

func TestDeployViaArgoCD_DryRun(t *testing.T) {
	flags := []string{"--set", "foo=bar", "--set-string", "baz=qux"}
	err := deployViaArgoCD("test-app", "https://repo", "test-chart", "test-ns", flags, "", true)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestWaitForArgoCDApps_DryRun(t *testing.T) {
	waitForArgoCDApps("test-ns", true) // Should return immediately without error
}
