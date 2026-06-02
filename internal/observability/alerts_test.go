package observability

import (
	"strings"
	"testing"
)

func TestProvisionAlertRules_DryRun(t *testing.T) {
	err := ProvisionAlertRules("test-ns", true)
	if err != nil {
		t.Errorf("expected no error on dry-run, got %v", err)
	}
}

func TestGetAlertRulesManifest(t *testing.T) {
	manifest := GetAlertRulesManifest("custom-namespace")

	// Verify the namespace was injected properly
	if !strings.Contains(manifest, "namespace: custom-namespace") {
		t.Errorf("expected manifest to contain the custom namespace, got:\n%s", manifest)
	}
}
