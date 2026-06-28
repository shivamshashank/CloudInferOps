package cli

import (
	"os/user"
	"testing"
)

func TestCommandsInitialization(t *testing.T) {
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

func TestHasRootPrivilegesRecognizesSudoUID(t *testing.T) {
	got := hasRootPrivilegesForTest(
		func() int { return 1000 },
		func() int { return 1000 },
		func(key string) string {
			if key == "SUDO_UID" {
				return "0"
			}
			return ""
		},
		func() (*user.User, error) { return &user.User{Uid: "1000"}, nil },
	)

	if !got {
		t.Fatal("expected sudo env with SUDO_UID=0 to be treated as elevated privileges")
	}
}
