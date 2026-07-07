package ui

import (
	"strings"
	"testing"
)

func TestDeploymentManifest(t *testing.T) {
	manifest := DeploymentManifest("example/portal:test", true)
	for _, expected := range []string{"example/portal:test", "kind: Deployment", "kind: Service", "kind: Ingress", "kind: ClusterRole"} {
		if !strings.Contains(manifest, expected) {
			t.Fatalf("manifest missing %q", expected)
		}
	}
}

func TestDeployPortalDryRun(t *testing.T) {
	if err := DeployPortal("example/portal:test", false, true); err != nil {
		t.Fatalf("dry run failed: %v", err)
	}
}
