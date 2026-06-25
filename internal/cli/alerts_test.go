package cli

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestAlertsCommandWiring(t *testing.T) {
	if alertsCmd.Use != "alerts" {
		t.Errorf("expected command Use 'alerts', got '%s'", alertsCmd.Use)
	}
}

func TestTestCmdRun(t *testing.T) {
	// Execute the command to ensure it prints standard output correctly without panicking
	// This simulates a user running `cloudinfer alerts test`
	testCmd.Run(&cobra.Command{}, []string{})
}
