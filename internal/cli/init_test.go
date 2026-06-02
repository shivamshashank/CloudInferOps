package cli

import (
	"testing"
)

func TestInitCommandWiring(t *testing.T) {
	if initCmd.Use != "init" {
		t.Errorf("expected command Use 'init', got '%s'", initCmd.Use)
	}
}
