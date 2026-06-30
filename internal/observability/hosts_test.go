package observability

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/shivamshashank/CloudInferOps/internal/utils"
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
	// Isolate CloudInferOps config dir
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	if err := os.MkdirAll(filepath.Join(tmpDir, ".cloudinferops"), 0755); err != nil {
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
	data, err := os.ReadFile(mockHostsFile) //nolint:gosec
	if err != nil {
		t.Fatalf("failed to read mock hosts file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "192.168.99.100   grafana.local") {
		t.Errorf("expected updated mapping in hosts file, got:\n%s", content)
	}
}

func TestFetchIngressIPStrategies(t *testing.T) {
	// Temporarily set ingressIPRetryDelay to 1ms
	oldDelay := ingressIPRetryDelay
	ingressIPRetryDelay = 1 * time.Millisecond
	defer func() { ingressIPRetryDelay = oldDelay }()

	tmpDir := t.TempDir()
	mockBinDir := filepath.Join(tmpDir, "bin")
	if err := os.MkdirAll(mockBinDir, 0755); err != nil {
		t.Fatalf("failed to create mock bin dir: %v", err)
	}

	mockKubectlPath := filepath.Join(mockBinDir, "kubectl")
	oldPath := os.Getenv("PATH")

	// 1. Hostname strategy test
	mockKubectlContent := `#!/bin/sh
case "$*" in
  *"loadBalancer.ingress[0].ip"*)
    exit 1
    ;;
  *"loadBalancer.ingress[0].hostname"*)
    echo "my-loadbalancer.aws.com"
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

	t.Setenv("PATH", mockBinDir+":"+oldPath)
	ip, err := FetchIngressIP("observability", false)
	if err != nil {
		t.Fatalf("expected no error under hostname strategy, got: %v", err)
	}
	if ip != "my-loadbalancer.aws.com" {
		t.Errorf("expected my-loadbalancer.aws.com, got %s", ip)
	}

	// 2. ExternalIP strategy test
	mockKubectlContent = `#!/bin/sh
case "$*" in
  *"loadBalancer.ingress[0].ip"*)
    exit 1
    ;;
  *"loadBalancer.ingress[0].hostname"*)
    exit 1
    ;;
  *"externalIPs[0]"*)
    echo "192.168.1.15"
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

	ip, err = FetchIngressIP("observability", false)
	if err != nil {
		t.Fatalf("expected no error under externalIP strategy, got: %v", err)
	}
	if ip != "192.168.1.15" {
		t.Errorf("expected 192.168.1.15, got %s", ip)
	}

	// 3. Ingress strategy test
	mockKubectlContent = `#!/bin/sh
case "$*" in
  *"get svc"*)
    exit 1
    ;;
  *"get ingress"*)
    echo "10.0.5.25"
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

	ip, err = FetchIngressIP("observability", false)
	if err != nil {
		t.Fatalf("expected no error under ingress strategy, got: %v", err)
	}
	if ip != "10.0.5.25" {
		t.Errorf("expected 10.0.5.25, got %s", ip)
	}

	// 4. Fallback to host local IP strategy test
	mockKubectlContent = `#!/bin/sh
exit 1
`
	if err := os.WriteFile(mockKubectlPath, []byte(mockKubectlContent), 0755); err != nil {
		t.Fatalf("failed to write mock kubectl: %v", err)
	}

	ip, err = FetchIngressIP("observability", false)
	if err != nil {
		t.Fatalf("expected fallback to host local IP, got error: %v", err)
	}
	expectedIP := utils.GetLocalIP()
	if ip != expectedIP {
		t.Errorf("expected %s, got %s", expectedIP, ip)
	}
}
