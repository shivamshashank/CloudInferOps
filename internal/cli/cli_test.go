package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLICommandsSimple(t *testing.T) {
	// 1. Test versionCmd
	versionCmd.Run(versionCmd, []string{})

	// 2. Test testCmd (alerts test)
	testCmd.Run(testCmd, []string{})
}

func TestCLICommandsWithPATHMock(t *testing.T) {
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
  *"cluster-info"*)
    echo "Kubernetes control plane is running"
    exit 0
    ;;
  *"get nodes"*)
    echo "node-1 Ready"
    exit 0
    ;;
  *"get storageclass"*)
    echo "standard"
    exit 0
    ;;
  *"get pods"*)
    echo "stackpulse-prometheus-server-123 1/1 Running"
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

	// Write mock 'helm'
	mockHelmPath := filepath.Join(mockBinDir, "helm")
	mockHelmContent := `#!/bin/sh
exit 0
`
	if err := os.WriteFile(mockHelmPath, []byte(mockHelmContent), 0755); err != nil {
		t.Fatalf("failed to write mock helm: %v", err)
	}

	// Prepend mock bin dir to PATH
	oldPath := os.Getenv("PATH")
	t.Setenv("PATH", mockBinDir+":"+oldPath)

	// Set temp HOME for config isolation
	t.Setenv("HOME", tmpDir)
	if err := os.MkdirAll(filepath.Join(tmpDir, ".stackpulse"), 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// 1. Test initCmd
	err := initCmd.RunE(initCmd, []string{})
	if err != nil {
		t.Errorf("initCmd failed: %v", err)
	}

	// Re-run initCmd to trigger "already initialized" path
	err = initCmd.RunE(initCmd, []string{})
	if err != nil {
		t.Errorf("initCmd re-run failed: %v", err)
	}

	// 2. Test doctorCmd
	doctorCmd.Run(doctorCmd, []string{})

	// 3. Test statusCmd (happy path)
	err = statusCmd.RunE(statusCmd, []string{})
	if err != nil {
		t.Errorf("statusCmd failed: %v", err)
	}

	// 4. Test connectCmd (happy path)
	connectBrowser = false // disable opening browser
	err = connectCmd.RunE(connectCmd, []string{})
	if err != nil {
		t.Errorf("connectCmd failed: %v", err)
	}

	// 5. Test deploy observabilityCmd (happy path)
	deployDryRun = true
	err = observabilityCmd.RunE(observabilityCmd, []string{})
	if err != nil {
		t.Errorf("observabilityCmd failed: %v", err)
	}

	// 6. Test configureCmd (alerts configure)
	// We need to pipe mock stdin to read credentials
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	r, w, pipeErr := os.Pipe()
	if pipeErr != nil {
		t.Fatalf("failed to create pipe: %v", pipeErr)
	}
	os.Stdin = r

	// Write mock inputs into pipe
	_, _ = w.Write([]byte("https://hooks.slack.com/services/mock\npd-key-123\n"))
	_ = w.Close()

	// Configure both slack and pagerduty
	configureSlack = true
	configurePagerDuty = true
	err = configureCmd.RunE(configureCmd, []string{})
	if err != nil {
		t.Errorf("configureCmd failed: %v", err)
	}
}

func TestRootExecute(t *testing.T) {
	oldArgs := os.Args
	os.Args = []string{"stackpulse"}
	defer func() { os.Args = oldArgs }()

	Execute()
}

func TestDeployObservabilityPrompts(t *testing.T) {
	// Create mock bin directory
	tmpDir := t.TempDir()
	mockBinDir := filepath.Join(tmpDir, "bin")
	if err := os.MkdirAll(mockBinDir, 0755); err != nil {
		t.Fatalf("failed to create mock bin dir: %v", err)
	}

	// Write a stateful 'kubectl' mock script
	stateFile := filepath.Join(tmpDir, "kubectl_state")
	mockKubectlPath := filepath.Join(mockBinDir, "kubectl")
	mockKubectlContent := `#!/bin/sh
if [ ! -f "` + stateFile + `" ]; then
    echo "called" > "` + stateFile + `"
    exit 1
fi
case "$*" in
  *"config current-context"*)
    echo "mock-context"
    exit 0
    ;;
  *"get svc"*"-o json"*)
    echo '{"items": [{"metadata": {"name": "stackpulse-prometheus-grafana"}, "spec": {"ports": [{"port": 80}]}}, {"metadata": {"name": "stackpulse-prometheus-kube-prometheus"}, "spec": {"ports": [{"port": 9090}]}}, {"metadata": {"name": "stackpulse-prometheus-kube-alertmanager"}, "spec": {"ports": [{"port": 9093}]}}]}'
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

	// Write mock scripts for other tools
	mocks := []string{"docker", "kind", "minikube", "helm", "sh"}
	for _, m := range mocks {
		path := filepath.Join(mockBinDir, m)
		content := `#!/bin/sh
exit 0
`
		if err := os.WriteFile(path, []byte(content), 0755); err != nil {
			t.Fatalf("failed to write mock %s: %v", m, err)
		}
	}

	// Prepend mock bin dir to PATH
	oldPath := os.Getenv("PATH")
	t.Setenv("PATH", mockBinDir+":"+oldPath)

	// Mock stdin choosing option 1 (kind)
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdin = r

	_, _ = w.Write([]byte("1\n"))
	_ = w.Close()

	// Deploy under non-dryrun to trigger prompts
	deployDryRun = false
	defer func() { deployDryRun = true }()

	// Trigger deploy observability
	err = observabilityCmd.RunE(observabilityCmd, []string{})
	if err != nil {
		t.Errorf("observabilityCmd with prompt option 1 failed: %v", err)
	}
}

func TestDeployObservabilityPromptCancel(t *testing.T) {
	tmpDir := t.TempDir()
	mockBinDir := filepath.Join(tmpDir, "bin")
	if err := os.MkdirAll(mockBinDir, 0755); err != nil {
		t.Fatalf("failed to create mock bin dir: %v", err)
	}

	// Kubectl always fails
	mockKubectlPath := filepath.Join(mockBinDir, "kubectl")
	mockKubectlContent := `#!/bin/sh
exit 1
`
	if err := os.WriteFile(mockKubectlPath, []byte(mockKubectlContent), 0755); err != nil {
		t.Fatalf("failed to write mock kubectl: %v", err)
	}

	oldPath := os.Getenv("PATH")
	t.Setenv("PATH", mockBinDir+":"+oldPath)

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdin = r

	_, _ = w.Write([]byte("4\n")) // choose 'no'
	_ = w.Close()

	deployDryRun = false
	defer func() { deployDryRun = true }()

	err = observabilityCmd.RunE(observabilityCmd, []string{})
	if err == nil {
		t.Error("expected cancel error from deploy, got nil")
	} else if !strings.Contains(err.Error(), "kubernetes cluster unreachable") {
		t.Errorf("expected unreachable error, got: %v", err)
	}
}
