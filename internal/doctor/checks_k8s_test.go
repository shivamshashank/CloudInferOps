package doctor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckTool(t *testing.T) {
	// 1. Success case: check an existing tool (like sh or go)
	res := CheckTool("sh", true)
	if res.Status != StatusOK {
		t.Errorf("expected StatusOK for sh, got %v", res.Status)
	}

	// 2. Failure case: check non-existing tool
	res = CheckTool("non-existing-tool-abc-xyz", true)
	if res.Status != StatusError {
		t.Errorf("expected StatusError for critical missing tool, got %v", res.Status)
	}

	res = CheckTool("non-existing-tool-abc-xyz", false)
	if res.Status != StatusWarn {
		t.Errorf("expected StatusWarn for non-critical missing tool, got %v", res.Status)
	}
}

func TestCheckK8sClusterHappyPath(t *testing.T) {
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
    echo "mock-k8s-context"
    exit 0
    ;;
  *"cluster-info"*)
    echo "Kubernetes control plane is running"
    exit 0
    ;;
  *"get nodes"*)
    echo "node-1 Ready"
    echo "node-2 Ready"
    exit 0
    ;;
  *"get storageclass"*)
    echo "standard (default)"
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

	results, detected := CheckK8sCluster()
	if !detected {
		t.Error("expected cluster to be detected")
	}

	hasContext := false
	hasNodes := false
	hasSC := false
	for _, res := range results {
		if strings.Contains(res.Name, "Context") && res.Status == StatusOK {
			hasContext = true
		}
		if strings.Contains(res.Name, "Nodes") && res.Status == StatusOK {
			hasNodes = true
		}
		if strings.Contains(res.Name, "StorageClass") && res.Status == StatusOK {
			hasSC = true
		}
	}

	if !hasContext {
		t.Error("expected Context OK check result")
	}
	if !hasNodes {
		t.Error("expected Nodes OK check result")
	}
	if !hasSC {
		t.Error("expected StorageClass OK check result")
	}
}

func TestCheckK8sClusterFailures(t *testing.T) {
	// 1. Case: kubectl doesn't exist in PATH
	oldPath := os.Getenv("PATH")
	t.Setenv("PATH", "") // clear PATH

	results, detected := CheckK8sCluster()
	if detected {
		t.Error("expected cluster to not be detected when kubectl is missing")
	}
	if len(results) != 1 || results[0].Status != StatusWarn {
		t.Errorf("expected 1 warning result, got %v", results)
	}

	// Restore PATH and create mock failing kubectl
	tmpDir := t.TempDir()
	mockBinDir := filepath.Join(tmpDir, "bin")
	if err := os.MkdirAll(mockBinDir, 0755); err != nil {
		t.Fatalf("failed to create mock bin dir: %v", err)
	}
	mockKubectlPath := filepath.Join(mockBinDir, "kubectl")
	mockKubectlContent := `#!/bin/sh
echo "mock kubectl failed" >&2
exit 1
`
	if err := os.WriteFile(mockKubectlPath, []byte(mockKubectlContent), 0755); err != nil {
		t.Fatalf("failed to write mock kubectl: %v", err)
	}

	t.Setenv("PATH", mockBinDir+":"+oldPath)

	_, detected = CheckK8sCluster()
	if detected {
		t.Error("expected cluster to not be detected when kubectl command fails")
	}
}
