package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Namespace != "observability" {
		t.Errorf("Expected namespace 'observability', got '%s'", cfg.Namespace)
	}
	if !cfg.Observability.Prometheus {
		t.Error("Expected Prometheus to be true by default")
	}
	if cfg.Alerts.Slack.WebhookUrlSecret != "stackpulse-slack-webhook" {
		t.Errorf("Expected default slack webhook secret, got '%s'", cfg.Alerts.Slack.WebhookUrlSecret)
	}
}

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err == nil {
		expanded := ExpandPath("~/testpath")
		expected := filepath.Join(home, "testpath")
		if expanded != expected {
			t.Errorf("Expected %s, got %s", expected, expanded)
		}
	}

	expanded := ExpandPath("/absolute/path")
	if expanded != "/absolute/path" {
		t.Errorf("Expected /absolute/path, got %s", expanded)
	}
}

func TestInitAndSaveConfig(t *testing.T) {
	// Isolate the config directory to a temporary folder
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)

	// 1. Test InitConfig (should create the file with defaults)
	err := InitConfig(true)
	if err != nil {
		t.Fatalf("InitConfig failed to create config: %v", err)
	}

	// 2. Modify and SaveConfig
	GlobalConfig.Namespace = "test-custom-namespace"
	err = SaveConfig()
	if err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// 3. Re-read and verify changes persisted
	if err := InitConfig(false); err != nil {
		t.Fatalf("InitConfig failed to read existing config: %v", err)
	}
	if GlobalConfig.Namespace != "test-custom-namespace" {
		t.Errorf("Expected modified namespace 'test-custom-namespace', got '%s'", GlobalConfig.Namespace)
	}
}

func TestGetConfigPath(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)

	path, err := GetConfigPath()
	if err != nil {
		t.Fatalf("expected no error from GetConfigPath, got: %v", err)
	}

	expected := filepath.Join(tempDir, ".stackpulse", "config.yaml")
	if path != expected {
		t.Errorf("expected path %q, got %q", expected, path)
	}
}
