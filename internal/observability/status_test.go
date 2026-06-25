package observability

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shivamshashank/CloudInferOps/internal/config"
)

func TestDecodeBase64(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		err      bool
	}{
		{
			input:    "c3VwZXItc2VjcmV0LXBhc3N3b3JkLTEyMw==",
			expected: "super-secret-password-123",
			err:      false,
		},
		{
			input:    "Y2xvdWRpbmZlcg==",
			expected: "cloudinferops",
			err:      false,
		},
		{
			input:    "!!!invalidbase64!!!",
			expected: "",
			err:      true,
		},
	}

	for _, test := range tests {
		result, err := DecodeBase64(test.input)
		if test.err {
			if err == nil {
				t.Errorf("expected error decoding '%s', got nil", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("unexpected error decoding '%s': %v", test.input, err)
			}
			if result != test.expected {
				t.Errorf("expected decoded value '%s', got '%s'", test.expected, result)
			}
		}
	}
}

func TestPrintStatus(t *testing.T) {
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
  *"config current-context"*)
    echo "mock-context"
    exit 0
    ;;
  *"get pods"*)
    echo "cloudinferops-prometheus-server-123 1/1 Running 0 1d"
    echo "cloudinferops-prometheus-grafana-123 1/1 Running 0 1d"
    echo "cloudinferops-loki-123 1/1 Running 0 1d"
    echo "cloudinferops-tempo-123 1/1 Running 0 1d"
    echo "cloudinferops-otel-123 1/1 Running 0 1d"
    echo "cloudinferops-webhook-handler-123 1/1 Running 0 1d"
    echo "cloudinferops-victoria-metrics-0 1/1 Running 0 1d"
    echo "cloudinferops-pyroscope-0 1/1 Running 0 1d"
    echo "cloudinferops-thanos-store-0 1/1 Running 0 1d"
    echo "cloudinferops-blackbox-exporter-123 1/1 Running 0 1d"
    echo "cloudinferops-alertmanager-0 1/1 Running 0 1d"
    exit 0
    ;;
  *"get secret cloudinferops-prometheus-grafana"*)
    echo "YWRtaW4="
    exit 0
    ;;
  *"get svc cloudinferops-ingress-nginx-controller"*)
    echo "192.168.99.100"
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

	// Prepend mock bin dir to PATH
	oldPath := os.Getenv("PATH")
	t.Setenv("PATH", mockBinDir+":"+oldPath)

	// Configure mock settings
	config.GlobalConfig = config.DefaultConfig()
	config.GlobalConfig.Observability.Prometheus = true
	config.GlobalConfig.Observability.VictoriaMetrics = true
	config.GlobalConfig.Observability.Alertmanager = true
	config.GlobalConfig.Observability.BlackboxExporter = true
	config.GlobalConfig.Observability.Pyroscope = true
	config.GlobalConfig.Observability.Thanos = true
	config.GlobalConfig.Alerts.Slack.Enabled = true

	// Capture output or just ensure it completes without error
	err := PrintStatus()
	if err != nil {
		t.Errorf("PrintStatus failed: %v", err)
	}
}
