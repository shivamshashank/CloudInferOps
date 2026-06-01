package gitops

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shivamshashank/StackPulse/internal/config"
)

func TestGenerateGitOpsRepo(t *testing.T) {
	// Initialize default config for testing
	config.GlobalConfig = config.DefaultConfig()

	tmpDir := t.TempDir()
	repoDir := filepath.Join(tmpDir, "gitops-repo")

	// Generate the repository layout
	err := generateGitOpsRepo(repoDir)
	if err != nil {
		t.Fatalf("generateGitOpsRepo failed: %v", err)
	}

	// Verify expected files and folders were created
	expectedPaths := []string{
		"infra/Chart.yaml",
		"infra/values.yaml",
		"monitoring/Chart.yaml",
		"monitoring/values.yaml",
		"apps/sample-app.yaml",
	}

	for _, path := range expectedPaths {
		if _, err := os.Stat(filepath.Join(repoDir, path)); os.IsNotExist(err) {
			t.Errorf("expected generated file %s to exist, but it was not found", path)
		}
	}
}
