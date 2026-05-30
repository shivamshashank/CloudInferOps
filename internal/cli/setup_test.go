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

	// Since we are running tests on non-Linux, the command should fail OS restrictions if run.
	if runtime.GOOS != "linux" {
		// Mock config variables so execution doesn't block on loading
		setupType = "k3s"
		setupYes = true // bypass prompts

		err := k8sCmd.RunE(k8sCmd, []string{})
		if err == nil {
			t.Error("expected setup k8s to fail on non-Linux due to OS platform restriction, but it succeeded")
		}

		if err.Error() == "" || !stringsContains(err.Error(), "setup is only supported on Linux") {
			t.Errorf("expected setup is only supported on Linux error, got: %v", err)
		}
	}
}

func stringsContains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || s[0:len(substr)] == substr || stringsContains(s[1:], substr))
}
