package observability

import (
	"testing"
)

func TestProvisionDashboards_DryRun(t *testing.T) {
	// Dry run will execute all the dashboard JSON string generators but won't invoke kubectl
	err := ProvisionDashboards("test-ns", true)
	if err != nil {
		t.Errorf("expected no error on dry-run, got %v", err)
	}
}
