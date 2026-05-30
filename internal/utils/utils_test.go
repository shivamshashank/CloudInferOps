package utils

import (
	"os"
	"testing"
)

func TestGetRealHomeDir(t *testing.T) {
	// 1. Standard execution
	home, err := GetRealHomeDir()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	expectedHome, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get os.UserHomeDir: %v", err)
	}

	// Under non-sudo, non-root execution, GetRealHomeDir must equal os.UserHomeDir
	if os.Getuid() != 0 {
		if home != expectedHome {
			t.Errorf("expected home %q, got %q", expectedHome, home)
		}
	}
}

func TestGetLocalIP(t *testing.T) {
	ip := GetLocalIP()
	if ip == "" {
		t.Error("expected non-empty IP address")
	}
}

func TestExecCommand(t *testing.T) {
	stdout, stderr, err := ExecCommand("", "echo", "hello")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if stdout != "hello" {
		t.Errorf("expected stdout 'hello', got '%s'", stdout)
	}
	if stderr != "" {
		t.Errorf("expected empty stderr, got '%s'", stderr)
	}

	// Test with non-empty directory
	tmpDir := t.TempDir()
	stdout, _, err = ExecCommand(tmpDir, "pwd")
	if err == nil {
		// Just ensure it ran successfully
		if stdout == "" {
			t.Error("expected non-empty pwd output")
		}
	}
}

func TestExecCommandEnv(t *testing.T) {
	env := map[string]string{"TEST_ENV_VAR": "stackpulse-test"}
	// We run 'sh -c echo $TEST_ENV_VAR' or just a simple command that prints it.
	// On Mac/Linux, we can run 'printenv' or 'env' or 'sh'. Let's run printenv or env.
	stdout, _, err := ExecCommandEnv("", env, "sh", "-c", "echo $TEST_ENV_VAR")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if stdout != "stackpulse-test" {
		t.Errorf("expected 'stackpulse-test', got '%s'", stdout)
	}
}

func TestExecCommandStream(t *testing.T) {
	stdout, stderr, err := ExecCommandStream("", "echo", "streaming")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if stdout != "streaming" {
		t.Errorf("expected stdout 'streaming', got '%s'", stdout)
	}
	if stderr != "" {
		t.Errorf("expected empty stderr, got '%s'", stderr)
	}
}

func TestExecCommandInteractive(t *testing.T) {
	stdout, stderr, err := ExecCommandInteractive("", "echo", "interactive")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if stdout != "interactive" {
		t.Errorf("expected stdout 'interactive', got '%s'", stdout)
	}
	if stderr != "" {
		t.Errorf("expected empty stderr, got '%s'", stderr)
	}
}
