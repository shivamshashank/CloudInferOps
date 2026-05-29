package cli

import (
	"strings"
	"testing"
)

func TestDeployCommandWiring(t *testing.T) {
	// Verify CLI command structures
	if deployCmd.Use != "deploy" {
		t.Errorf("expected command Use 'deploy', got '%s'", deployCmd.Use)
	}

	if observabilityCmd.Use != "observability" {
		t.Errorf("expected subcommand Use 'observability', got '%s'", observabilityCmd.Use)
	}

	// Verify dry-run flag
	dryRunFlag := observabilityCmd.Flag("dry-run")
	if dryRunFlag == nil {
		t.Fatal("expected flag 'dry-run' to be wired")
	}
	if dryRunFlag.DefValue != "false" {
		t.Errorf("expected default dry-run value 'false', got '%s'", dryRunFlag.DefValue)
	}
}

func TestDeployClusterPreCheckSafeguard(t *testing.T) {
	// Isolate tests using sandboxed HOME path
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Since there is no active K8s cluster in this environment, it should immediately fail with K8s unreachable error.
	err := observabilityCmd.RunE(observabilityCmd, []string{})
	if err == nil {
		t.Error("expected deploy observability to fail due to missing K8s cluster connectivity, but it succeeded")
	}

	if err.Error() == "" || !strings.Contains(err.Error(), "Kubernetes cluster unreachable") {
		t.Errorf("expected cluster unreachable safeguard error, got: %v", err)
	}
}
