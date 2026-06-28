package installer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallers(t *testing.T) {
	// Create mock bin directory
	tmpDir := t.TempDir()
	mockBinDir := filepath.Join(tmpDir, "bin")
	if err := os.MkdirAll(mockBinDir, 0755); err != nil {
		t.Fatalf("failed to create mock bin dir: %v", err)
	}

	// Write mock scripts
	mocks := map[string]string{
		"sh": `#!/bin/sh
case "$*" in
  *"curl -fsSL https://get.docker.com"*)
    echo "mock docker installer success"
    exit 0
    ;;
  *)
    exit 0
    ;;
esac
`,
		"sudo": `#!/bin/sh
exec "$@"
`,
		"mv": `#!/bin/sh
exit 0
`,
		"systemctl": `#!/bin/sh
exit 0
`,
		"docker": `#!/bin/sh
case "$*" in
  *"info"*)
    echo "Server Version: 20.10.0"
    exit 0
    ;;
esac
exit 0
`,
		"kubectl": `#!/bin/sh
exit 0
`,
	}

	for name, content := range mocks {
		path := filepath.Join(mockBinDir, name)
		if err := os.WriteFile(path, []byte(content), 0755); err != nil {
			t.Fatalf("failed to write mock %s: %v", name, err)
		}
	}

	// Prepend mock bin dir to PATH
	oldPath := os.Getenv("PATH")
	t.Setenv("PATH", mockBinDir+":"+oldPath)

	// 1. Test InstallDocker
	if err := InstallDocker(); err != nil {
		t.Errorf("InstallDocker failed: %v", err)
	}
}

func TestRunInstallCommandExecutesCommand(t *testing.T) {
	if err := runInstallCommand("echo runInstallCommand-ok"); err != nil {
		t.Fatalf("expected runInstallCommand to succeed, got: %v", err)
	}
}

func TestInstallersFailures(t *testing.T) {
	// Create mock bin directory for failures
	tmpDir := t.TempDir()
	mockBinDir := filepath.Join(tmpDir, "bin")
	if err := os.MkdirAll(mockBinDir, 0755); err != nil {
		t.Fatalf("failed to create mock bin dir: %v", err)
	}

	// All commands fail
	failScript := `#!/bin/sh
echo "mock installation execution failed" >&2
exit 1
`
	commands := []string{"sh", "sudo", "docker", "kubectl"}
	for _, cmd := range commands {
		path := filepath.Join(mockBinDir, cmd)
		if err := os.WriteFile(path, []byte(failScript), 0755); err != nil {
			t.Fatalf("failed to write failing mock %s: %v", cmd, err)
		}
	}

	// Prepend mock bin dir to PATH
	oldPath := os.Getenv("PATH")
	t.Setenv("PATH", mockBinDir+":"+oldPath)

	// Verify all return errors properly
	if err := InstallDocker(); err == nil {
		t.Error("expected InstallDocker to fail, got nil")
	} else if !strings.Contains(err.Error(), "mock installation execution failed") {
		t.Errorf("expected error containing 'mock installation execution failed', got: %v", err)
	}
}
