package inference

import (
	"testing"
)

func TestManifestGenerators(t *testing.T) {
	ns := NamespaceManifest()
	if ns == "" {
		t.Error("expected namespace manifest to be non-empty")
	}

	cfg := ModelConfigManifest("ollama", "llama3")
	if cfg == "" {
		t.Error("expected config manifest to be non-empty")
	}

	gw := GatewayDeploymentManifest()
	if gw == "" {
		t.Error("expected gateway deployment manifest to be non-empty")
	}
}

func TestDeployInference_DryRun(t *testing.T) {
	err := DeployInference("ollama", "llama3", true)
	if err != nil {
		t.Errorf("expected no error in dry-run, got: %v", err)
	}
}
