package observability

import (
	"testing"

	"github.com/shivamshashank/StackPulse/internal/config"
)

func TestDryRunDeployments(t *testing.T) {
	config.GlobalConfig = config.DefaultConfig()

	// Ensure dry-run functions return early without attempting to execute kubectl
	if err := createThanosSecret("test-ns", true); err != nil {
		t.Errorf("createThanosSecret dryRun failed: %v", err)
	}
	if err := applyObservabilityIngress("test-ns", true); err != nil {
		t.Errorf("applyObservabilityIngress dryRun failed: %v", err)
	}
}
