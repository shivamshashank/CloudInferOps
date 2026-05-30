package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Namespace != "observability" {
		t.Errorf("expected default namespace 'observability', got '%s'", cfg.Namespace)
	}
	if cfg.Kubernetes.Type != "auto" {
		t.Errorf("expected default k8s type 'auto', got '%s'", cfg.Kubernetes.Type)
	}
	if cfg.Observability.Prometheus != true {
		t.Error("expected default Prometheus enabled to be true")
	}
	if cfg.Observability.LogCollector != "alloy" {
		t.Errorf("expected default log collector 'alloy', got '%s'", cfg.Observability.LogCollector)
	}
}

func TestExpandPath(t *testing.T) {
	// Set mock home
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	path := "~/test/kubeconfig"
	expanded := ExpandPath(path)

	expected := filepath.Clean(filepath.Join(tmpHome, "test/kubeconfig"))
	if expanded != expected {
		t.Errorf("expected expanded path '%s', got '%s'", expected, expanded)
	}

	// Should not expand relative or absolute path without ~
	normPath := "/etc/resolv.conf"
	if ExpandPath(normPath) != normPath {
		t.Errorf("expected no expansion for standard path, got '%s'", ExpandPath(normPath))
	}
}

func TestInitConfig(t *testing.T) {
	// Isolate tests using dynamic HOME env overriding
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Verify init fails if config is missing and we set createIfMissing = false
	err := InitConfig(false)
	if err == nil || !strings.Contains(err.Error(), "configuration file does not exist") {
		t.Errorf("expected missing config error, got %v", err)
	}

	// Initialize config and write defaults
	err = InitConfig(true)
	if err != nil {
		t.Fatalf("failed to initialize default config: %v", err)
	}

	// Verify the config directory and file were generated
	expectedPath := filepath.Join(tmpHome, ".stackpulse", "config.yaml")
	if _, statErr := os.Stat(expectedPath); os.IsNotExist(statErr) {
		t.Fatalf("expected config file to be written at '%s'", expectedPath)
	}

	// Verify GlobalConfig fields populated correctly
	if GlobalConfig.Namespace != "observability" {
		t.Errorf("expected global namespace 'observability', got '%s'", GlobalConfig.Namespace)
	}

	// Update GlobalConfig and verify saving works
	GlobalConfig.Namespace = "custom-obs"
	GlobalConfig.Observability.Prometheus = false
	err = SaveConfig()
	if err != nil {
		t.Fatalf("failed to save custom config: %v", err)
	}

	// Reinitialize and load from disk (without creating missing)
	GlobalConfig = Config{} // reset memory config
	err = InitConfig(false)
	if err != nil {
		t.Fatalf("failed to re-initialize and load written config: %v", err)
	}

	if GlobalConfig.Namespace != "custom-obs" {
		t.Errorf("expected loaded namespace 'custom-obs', got '%s'", GlobalConfig.Namespace)
	}
	if GlobalConfig.Observability.Prometheus != false {
		t.Error("expected loaded Prometheus setting to be false")
	}
}
