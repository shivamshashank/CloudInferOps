package gitops

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateGitOpsRepo(t *testing.T) {
	tmpDir := t.TempDir()
	err := generateGitOpsRepo(tmpDir)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Verify that the directory structure and core files were created
	for _, file := range []string{"infra/Chart.yaml", "monitoring/Chart.yaml", "apps/sample-app.yaml"} {
		if _, err := os.Stat(filepath.Join(tmpDir, file)); os.IsNotExist(err) {
			t.Errorf("expected %s to exist", file)
		}
	}
}

func TestApplyGitOpsIngress_DryRun(t *testing.T) {
	err := applyGitOpsIngress("observability", true)
	if err != nil {
		t.Errorf("expected no error on dry-run, got %v", err)
	}
}

func TestDeployArgoCDApplications_DryRun(t *testing.T) {
	err := deployArgoCDApplications("observability", true)
	if err != nil {
		t.Errorf("expected no error on dry-run, got %v", err)
	}
}
