package cli

import (
	"strings"
	"testing"
)

func TestStatusCommandWiring(t *testing.T) {
	if statusCmd.Use != "status" {
		t.Errorf("expected command Use 'status', got '%s'", statusCmd.Use)
	}
}

func TestStatusClusterPreCheckSafeguard(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Since there is no active K8s cluster in this test environment, running statusCmd should fail immediately.
	err := statusCmd.RunE(statusCmd, []string{})
	if err == nil {
		t.Error("expected status command to fail due to missing K8s cluster connectivity, but it succeeded")
	}

	if err.Error() == "" || !strings.Contains(err.Error(), "Kubernetes cluster unreachable") {
		t.Errorf("expected cluster unreachable safeguard error, got: %v", err)
	}
}
