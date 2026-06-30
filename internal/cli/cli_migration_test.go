package cli

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestMigrationCommandsExist(t *testing.T) {
	if deployCmd == nil {
		t.Error("deployCmd was not initialized")
	}
	if deployPlatformCmd == nil {
		t.Error("deployPlatformCmd was not initialized")
	}
	if deployObservabilityCmd == nil {
		t.Error("deployObservabilityCmd was not initialized")
	}
	if deployInferenceCmd == nil {
		t.Error("deployInferenceCmd was not initialized")
	}
	if modelsCmd == nil {
		t.Error("modelsCmd was not initialized")
	}
	if modelsListCmd == nil {
		t.Error("modelsListCmd was not initialized")
	}
	if modelsPullCmd == nil {
		t.Error("modelsPullCmd was not initialized")
	}
	if benchmarkCmd == nil {
		t.Error("benchmarkCmd was not initialized")
	}
	if benchmarkRunCmd == nil {
		t.Error("benchmarkRunCmd was not initialized")
	}
	if benchmarkReportCmd == nil {
		t.Error("benchmarkReportCmd was not initialized")
	}
}

func captureStdout(f func() error) (string, error) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := f()

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String(), err
}

func TestDeployInferenceCmdDryRun(t *testing.T) {
	RootCmd.SetArgs([]string{"deploy", "inference", "--dry-run", "--provider", "ollama", "--model", "llama3"})
	output, err := captureStdout(func() error {
		return RootCmd.Execute()
	})
	if err != nil {
		t.Fatalf("Failed to execute deploy inference: %v", err)
	}

	if !strings.Contains(output, "[DRY-RUN]") {
		t.Errorf("Expected dry-run output, got: %s", output)
	}
	if !strings.Contains(output, "Provider: ollama") || !strings.Contains(output, "Model: llama3") {
		t.Errorf("Expected provider and model in dry-run output, got: %s", output)
	}
}

func TestModelsListFallback(t *testing.T) {
	RootCmd.SetArgs([]string{"models", "list"})
	output, err := captureStdout(func() error {
		return RootCmd.Execute()
	})
	if err != nil {
		t.Fatalf("Failed to execute models list: %v", err)
	}

	// Since no active gateway/ollama is running in test, it should fallback to static catalog
	if !strings.Contains(output, "No active gateway or local Ollama daemon detected") {
		t.Errorf("Expected fallback notice, got: %s", output)
	}
	if !strings.Contains(output, "llama3") || !strings.Contains(output, "mistral") {
		t.Errorf("Expected fallback models catalog, got: %s", output)
	}
}

func TestBenchmarkRunCmd(t *testing.T) {
	RootCmd.SetArgs([]string{"benchmark", "run", "--model", "mistral", "--requests", "50", "--concurrency", "5", "--provider", "vllm"})
	output, err := captureStdout(func() error {
		return RootCmd.Execute()
	})
	if err != nil {
		t.Fatalf("Failed to execute benchmark run: %v", err)
	}

	if !strings.Contains(output, "Running inference benchmark against model") || !strings.Contains(output, "mistral") {
		t.Errorf("Expected model mistral in output, got: %s", output)
	}
	if !strings.Contains(output, "Target Provider: vllm") {
		t.Errorf("Expected provider vllm in output, got: %s", output)
	}
}

func TestDeployObservabilityAliasDeprecated(t *testing.T) {
	if deployObservabilityCmd.Deprecated == "" {
		t.Error("Expected deployObservabilityCmd to be marked as deprecated")
	}
}
