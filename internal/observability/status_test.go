package observability

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shivamshashank/StackPulse/internal/config"
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
			input:    "c3RhY2twdWxzZQ==",
			expected: "stackpulse",
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
    echo "stackpulse-prometheus-server-123 1/1 Running 0 1d"
    echo "stackpulse-prometheus-grafana-123 1/1 Running 0 1d"
    echo "stackpulse-loki-123 1/1 Running 0 1d"
    echo "stackpulse-tempo-123 1/1 Running 0 1d"
    echo "stackpulse-otel-123 1/1 Running 0 1d"
    echo "stackpulse-webhook-handler-123 1/1 Running 0 1d"
    exit 0
    ;;
  *"get secret stackpulse-prometheus-grafana"*)
    echo "YWRtaW4="
    exit 0
    ;;
  *"get svc stackpulse-ingress-nginx-controller"*)
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
	config.GlobalConfig.Alerts.Slack.Enabled = true

	// Capture output or just ensure it completes without error
	err := PrintStatus()
	if err != nil {
		t.Errorf("PrintStatus failed: %v", err)
	}
}

