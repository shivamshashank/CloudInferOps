package utils

import (
	"os"
	"testing"
)

func TestGetRealHomeDir(t *testing.T) {
	// 1. Standard execution
	home, err := GetRealHomeDir()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	expectedHome, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get os.UserHomeDir: %v", err)
	}

	// Under non-sudo, non-root execution, GetRealHomeDir must equal os.UserHomeDir
	if os.Getuid() != 0 {
		if home != expectedHome {
			t.Errorf("expected home %q, got %q", expectedHome, home)
		}
	}
}

func TestGetLocalIP(t *testing.T) {
	ip := GetLocalIP()
	if ip == "" {
		t.Error("expected non-empty IP address")
	}
}
