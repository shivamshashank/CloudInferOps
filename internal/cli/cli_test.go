package cli

import (
	"testing"
)

func TestCommandsInitialization(t *testing.T) {
	if connectCmd == nil {
		t.Error("connectCmd was not initialized")
	}
	if uninstallCmd == nil {
		t.Error("uninstallCmd was not initialized")
	}
	if uninstallObservabilityCmd == nil {
		t.Error("uninstallObservabilityCmd was not initialized")
	}
	if uninstallAllCmd == nil {
		t.Error("uninstallAllCmd was not initialized")
	}
}
