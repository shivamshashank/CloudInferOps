package gitops

import (
	"strings"
	"testing"
)

func TestGitServerManifests(t *testing.T) {
	manifest := GitServerManifests("gitops-ns")

	if !strings.Contains(manifest, "namespace: gitops-ns") {
		t.Errorf("expected manifest to contain namespace, got: %s", manifest)
	}
	if !strings.Contains(manifest, "name: cloudinferops-git-server") {
		t.Errorf("expected manifest to contain cloudinferops-git-server deployment")
	}
}

func TestDeployGitServer_DryRun(t *testing.T) {
	err := DeployGitServer("gitops-ns", true)
	if err != nil {
		t.Errorf("expected no error on dry run, got %v", err)
	}
}
