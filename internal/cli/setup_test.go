package cli

import (
	"runtime"
	"testing"
)

func TestSetupCommandWiring(t *testing.T) {
	// Verify command name and description wiring
	if setupCmd.Use != "setup" {
		t.Errorf("expected command Use 'setup', got '%s'", setupCmd.Use)
	}

	if k8sCmd.Use != "k8s" {
		t.Errorf("expected subcommand Use 'k8s', got '%s'", k8sCmd.Use)
	}

	// Verify defaults of flags
	typeFlag := k8sCmd.Flag("type")
	if typeFlag == nil {
		t.Fatal("expected flag 'type' to be wired")
	}
	if typeFlag.DefValue != "k3s" {
		t.Errorf("expected default type value 'k3s', got '%s'", typeFlag.DefValue)
	}

	yesFlag := k8sCmd.Flag("yes")
	if yesFlag == nil {
		t.Fatal("expected flag 'yes' to be wired")
	}
	if yesFlag.DefValue != "false" {
		t.Errorf("expected default yes value 'false', got '%s'", yesFlag.DefValue)
	}
}

func TestSetupOSRestrictions(t *testing.T) {
	// Set mock home directory so config doesn't pollute user environment
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Since we are running tests on macOS, the command should fail OS restrictions if run.
	if runtime.GOOS == "darwin" {
		// Mock config variables so execution doesn't block on loading
		setupType = "k3s"
		setupYes = true // bypass prompts

		err := k8sCmd.RunE(k8sCmd, []string{})
		if err == nil {
			t.Error("expected setup k8s to fail on macOS due to OS platform restriction, but it succeeded")
		}

		if err.Error() == "" || !stringsContains(err.Error(), "unsupported operating system for k3s") {
			t.Errorf("expected unsupported operating system error on macOS, got: %v", err)
		}
	}
}

func TestSetupMinikubeOSRestrictionBypass(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	if runtime.GOOS == "darwin" {
		setupType = "minikube"
		setupYes = true

		err := k8sCmd.RunE(k8sCmd, []string{})
		if err == nil {
			t.Error("expected setup k8s --type minikube to fail on missing minikube dependency, but it succeeded")
		}

		// It should fail on missing minikube binary, NOT the macOS safeguard!
		if err.Error() == "" || !stringsContains(err.Error(), "minikube dependency missing") {
			t.Errorf("expected minikube dependency missing error, got: %v", err)
		}
	}
}

func stringsContains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || s[0:len(substr)] == substr || stringsContains(s[1:], substr))
}
