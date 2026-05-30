package observability

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shivamshashank/StackPulse/internal/config"
)

func TestDeployObservabilityDryRun(t *testing.T) {
	// Enable Prometheus, Loki, Tempo, and OpenTelemetry in config
	config.GlobalConfig = config.DefaultConfig()
	config.GlobalConfig.Observability.Prometheus = true
	config.GlobalConfig.Observability.Grafana = true
	config.GlobalConfig.Observability.Loki = true
	config.GlobalConfig.Observability.Tempo = true
	config.GlobalConfig.Observability.OpenTelemetry = true

	// Deploy under dry-run
	err := DeployObservability(true)
	if err != nil {
		t.Errorf("expected no error under DeployObservability dry-run, got: %v", err)
	}
}

func TestDeployObservabilityMockedPATH(t *testing.T) {
	// Create mock bin directory
	tmpDir := t.TempDir()
	mockBinDir := filepath.Join(tmpDir, "bin")
	if err := os.MkdirAll(mockBinDir, 0755); err != nil {
		t.Fatalf("failed to create mock bin dir: %v", err)
	}

	// Write mock 'kubectl'
	mockKubectlPath := filepath.Join(mockBinDir, "kubectl")
	mockKubectlContent := `#!/bin/sh
case "$*" in
  *"get svc"*"-o json"*)
    echo '{"items": [{"metadata": {"name": "stackpulse-prometheus-grafana"}, "spec": {"ports": [{"port": 80}]}}, {"metadata": {"name": "stackpulse-prometheus-kube-prometheus"}, "spec": {"ports": [{"port": 9090}]}}, {"metadata": {"name": "stackpulse-prometheus-kube-alertmanager"}, "spec": {"ports": [{"port": 9093}]}}]}'
    exit 0
    ;;
  *)
    exit 0
    ;;
esac
`
	if err := os.WriteFile(mockKubectlPath, []byte(mockKubectlContent), 0755); err != nil {
		t.Fatalf("failed to write mock kubectl: %v", err)
	}

	// Write mock 'helm'
	mockHelmPath := filepath.Join(mockBinDir, "helm")
	mockHelmContent := `#!/bin/sh
exit 0
`
	if err := os.WriteFile(mockHelmPath, []byte(mockHelmContent), 0755); err != nil {
		t.Fatalf("failed to write mock helm: %v", err)
	}

	// Prepend mock bin dir to PATH
	oldPath := os.Getenv("PATH")
	t.Setenv("PATH", mockBinDir+":"+oldPath)

	config.GlobalConfig = config.DefaultConfig()
	config.GlobalConfig.Observability.Prometheus = true
	config.GlobalConfig.Observability.Grafana = true
	config.GlobalConfig.Observability.Loki = true
	config.GlobalConfig.Observability.Tempo = true
	config.GlobalConfig.Observability.OpenTelemetry = true

	// Deploy under non-dryrun with mocked tools
	err := DeployObservability(false)
	if err != nil {
		t.Errorf("expected no error under DeployObservability with mocks, got: %v", err)
	}
}
