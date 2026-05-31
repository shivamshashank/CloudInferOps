package gitops

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckGitInstalled(t *testing.T) {
	installed := CheckGitInstalled()
	t.Logf("Git installed: %t", installed)
}

func TestGitServerManifests(t *testing.T) {
	manifests := GitServerManifests("test-ns")
	if !contains(manifests, "namespace: test-ns") {
		t.Errorf("manifests should contain the specified namespace")
	}
	if !contains(manifests, "stackpulse-git-server") {
		t.Errorf("manifests should define stackpulse-git-server")
	}
}

func TestGenerateGitOpsRepo(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gitops-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	err = generateGitOpsRepo(tmpDir)
	if err != nil {
		t.Fatalf("generateGitOpsRepo failed: %v", err)
	}

	expectedDirs := []string{"prometheus", "loki", "tempo", "otel"}
	for _, dir := range expectedDirs {
		path := filepath.Join(tmpDir, dir)
		if fi, err := os.Stat(path); err != nil || !fi.IsDir() {
			t.Errorf("expected directory %s to be created", path)
		}

		chartPath := filepath.Join(path, "Chart.yaml")
		if _, err := os.Stat(chartPath); err != nil {
			t.Errorf("expected Chart.yaml to exist in %s", path)
		}

		valuesPath := filepath.Join(path, "values.yaml")
		if _, err := os.Stat(valuesPath); err != nil {
			t.Errorf("expected values.yaml to exist in %s", path)
		}
	}
}

func contains(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
