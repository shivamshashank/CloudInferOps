package observability

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFetchIngressIP(t *testing.T) {
	// 1. Dry run
	ip, err := FetchIngressIP("observability", true)
	if err != nil {
		t.Fatalf("expected no error under dry-run, got: %v", err)
	}
	if ip != "127.0.0.1" {
		t.Errorf("expected 127.0.0.1, got %s", ip)
	}

	// 2. Non-dryrun with mocked kubectl
	tmpDir := t.TempDir()
	mockBinDir := filepath.Join(tmpDir, "bin")
	if err := os.MkdirAll(mockBinDir, 0755); err != nil {
		t.Fatalf("failed to create mock bin dir: %v", err)
	}

	mockKubectlPath := filepath.Join(mockBinDir, "kubectl")
	mockKubectlContent := `#!/bin/sh
echo "192.168.99.100"
exit 0
`
	if err := os.WriteFile(mockKubectlPath, []byte(mockKubectlContent), 0755); err != nil {
		t.Fatalf("failed to write mock kubectl: %v", err)
	}

	oldPath := os.Getenv("PATH")
	t.Setenv("PATH", mockBinDir+":"+oldPath)

	ip, err = FetchIngressIP("observability", false)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if ip != "192.168.99.100" {
		t.Errorf("expected 192.168.99.100, got %s", ip)
	}
}

func TestUpdateHostsFile(t *testing.T) {
	// Isolate StackPulse config dir
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	if err := os.MkdirAll(filepath.Join(tmpDir, ".stackpulse"), 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Create a mock hosts file
	mockHostsFile := filepath.Join(tmpDir, "hosts")
	initialContent := "127.0.0.1       localhost\n"
	if err := os.WriteFile(mockHostsFile, []byte(initialContent), 0644); err != nil {
		t.Fatalf("failed to write mock hosts file: %v", err)
	}

	// Set hostsFilePath variable to mock hosts file
	oldHostsFilePath := hostsFilePath
	hostsFilePath = mockHostsFile
	defer func() { hostsFilePath = oldHostsFilePath }()

	// Create mock cp command in PATH
	mockBinDir := filepath.Join(tmpDir, "bin")
	if err := os.MkdirAll(mockBinDir, 0755); err != nil {
		t.Fatalf("failed to create mock bin dir: %v", err)
	}

	mockCpPath := filepath.Join(mockBinDir, "cp")
	mockCpContent := `#!/bin/sh
# Real cp under mock path to simulate copying hosts.tmp to mock hosts file
exec /bin/cp "$1" "$2"
`
	// Wait, on Mac, /bin/cp exists, but it's safer to just do a simple copy in shell
	mockCpContent = `#!/bin/sh
cat "$1" > "$2"
exit 0
`
	if err := os.WriteFile(mockCpPath, []byte(mockCpContent), 0755); err != nil {
		t.Fatalf("failed to write mock cp: %v", err)
	}

	// Also write a mock 'sudo'
	mockSudoPath := filepath.Join(mockBinDir, "sudo")
	mockSudoContent := `#!/bin/sh
exec "$@"
`
	if err := os.WriteFile(mockSudoPath, []byte(mockSudoContent), 0755); err != nil {
		t.Fatalf("failed to write mock sudo: %v", err)
	}

	oldPath := os.Getenv("PATH")
	t.Setenv("PATH", mockBinDir+":"+oldPath)

	err := UpdateHostsFile("192.168.99.100", "grafana.local")
	if err != nil {
		t.Fatalf("UpdateHostsFile failed: %v", err)
	}

	// Read back and verify
	data, err := os.ReadFile(mockHostsFile)
	if err != nil {
		t.Fatalf("failed to read mock hosts file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "192.168.99.100   grafana.local") {
		t.Errorf("expected updated mapping in hosts file, got:\n%s", content)
	}
}
