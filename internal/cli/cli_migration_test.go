package cli

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMigrationCommandsExist(t *testing.T) {
	if deployCmd == nil {
		t.Error("deployCmd was not initialized")
	}
	if deployObservabilityCmd == nil {
		t.Error("deployObservabilityCmd was not initialized")
	}
	if deployInferenceCmd == nil {
		t.Error("deployInferenceCmd was not initialized")
	}
	if deployWebhookHandlerCmd == nil {
		t.Error("deployWebhookHandlerCmd was not initialized")
	}
	if dashboardsCmd == nil {
		t.Error("dashboardsCmd was not initialized")
	}
	if dashboardsImportCmd == nil {
		t.Error("dashboardsImportCmd was not initialized")
	}
	if logsCmd == nil {
		t.Error("logsCmd was not initialized")
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

func TestBenchmarkRunCmdRealRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"eval_count": 25, "completion_tokens": 25, "usage": {"completion_tokens": 25}}`))
	}))
	defer server.Close()

	RootCmd.SetArgs([]string{"benchmark", "run", "--model", "mistral", "--requests", "5", "--concurrency", "2", "--provider", "vllm", "--test-url", server.URL})
	output, err := captureStdout(func() error {
		return RootCmd.Execute()
	})
	if err != nil {
		t.Fatalf("Failed to execute benchmark run: %v", err)
	}

	if !strings.Contains(output, "Running inference benchmark against model") {
		t.Errorf("Expected start message, got: %s", output)
	}
	if !strings.Contains(output, "Benchmark completed successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}
	if !strings.Contains(output, "Token Throughput") {
		t.Errorf("Expected stats in output, got: %s", output)
	}
}

func TestBenchmarkReportCmd(t *testing.T) {
	RootCmd.SetArgs([]string{"benchmark", "report"})
	output, err := captureStdout(func() error {
		return RootCmd.Execute()
	})
	if err != nil {
		t.Fatalf("Failed to execute benchmark report: %v", err)
	}

	if !strings.Contains(output, "Unified Benchmark Report") {
		t.Errorf("Expected report header, got: %s", output)
	}
	if !strings.Contains(output, "Model:") {
		t.Errorf("Expected Model field, got: %s", output)
	}
}

func TestModelsPullCmdRealRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "downloading manifest"}` + "\n"))
		_, _ = w.Write([]byte(`{"status": "downloading weights", "completed": 50, "total": 100}` + "\n"))
	}))
	defer server.Close()

	RootCmd.SetArgs([]string{"models", "pull", "llama3", "--test-url", server.URL})
	output, err := captureStdout(func() error {
		return RootCmd.Execute()
	})
	if err != nil {
		t.Fatalf("Failed to execute models pull: %v", err)
	}

	if !strings.Contains(output, "Pulling model") || !strings.Contains(output, "llama3") {
		t.Errorf("Expected pulling message, got: %s", output)
	}
	if !strings.Contains(output, "Status: downloading weights (50.00%)") {
		t.Errorf("Expected weights download progress, got: %s", output)
	}
	if !strings.Contains(output, "Successfully pulled model 'llama3'") {
		t.Errorf("Expected success message, got: %s", output)
	}
}

func TestDeployWebhookHandlerCmdDryRun(t *testing.T) {
	RootCmd.SetArgs([]string{"deploy", "webhook-handler", "--dry-run"})
	output, err := captureStdout(func() error {
		return RootCmd.Execute()
	})
	if err != nil {
		t.Fatalf("Failed to execute deploy webhook-handler: %v", err)
	}

	if !strings.Contains(output, "[DRY-RUN]") {
		t.Errorf("Expected dry-run output, got: %s", output)
	}
	if !strings.Contains(output, "cloudinferops-webhook-handler") {
		t.Errorf("Expected webhook service/deployment name in dry-run output, got: %s", output)
	}
}

func TestDashboardsImportCmdDryRun(t *testing.T) {
	RootCmd.SetArgs([]string{"dashboards", "import", "--dry-run"})
	output, err := captureStdout(func() error {
		return RootCmd.Execute()
	})
	if err != nil {
		t.Fatalf("Failed to execute dashboards import: %v", err)
	}

	if !strings.Contains(output, "[DRY-RUN]") {
		t.Errorf("Expected dry-run output, got: %s", output)
	}
	if !strings.Contains(output, "Would apply Grafana Dashboard ConfigMap") {
		t.Errorf("Expected ConfigMap in dry-run output, got: %s", output)
	}
}

func TestLogsCmdMissingKubectl(t *testing.T) {
	// Temporarily clear PATH to simulate missing kubectl
	oldPath := os.Getenv("PATH")
	if err := os.Setenv("PATH", ""); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}
	defer func() {
		if err := os.Setenv("PATH", oldPath); err != nil {
			t.Errorf("failed to restore PATH: %v", err)
		}
	}()

	RootCmd.SetArgs([]string{"logs", "--component", "grafana"})
	_, err := captureStdout(func() error {
		return RootCmd.Execute()
	})
	if err == nil {
		t.Fatal("Expected error due to missing kubectl, got nil")
	}
	if !strings.Contains(err.Error(), "kubectl is required") {
		t.Errorf("Expected error about kubectl, got: %v", err)
	}
}

func TestGitopsBootstrapCmdDryRun(t *testing.T) {
	t.Setenv("SUDO_USER", "root")

	tmpDir := t.TempDir()
	mockBinDir := filepath.Join(tmpDir, "bin")
	if err := os.MkdirAll(mockBinDir, 0755); err != nil {
		t.Fatal(err)
	}

	mockKubectlPath := filepath.Join(mockBinDir, "kubectl")
	mockKubectlContent := `#!/bin/sh
if [ "$1" = "config" ] && [ "$2" = "current-context" ]; then
    echo "mock-context"
elif [ "$1" = "get" ] && [ "$2" = "nodes" ]; then
    echo "node-01 Ready control-plane 2d v1.29.0"
else
    echo "mock kubectl output"
fi
exit 0
`
	if err := os.WriteFile(mockKubectlPath, []byte(mockKubectlContent), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", mockBinDir+":"+os.Getenv("PATH"))

	RootCmd.SetArgs([]string{"gitops", "bootstrap", "--dry-run"})
	output, err := captureStdout(func() error {
		return RootCmd.Execute()
	})
	if err != nil {
		t.Fatalf("Failed to execute gitops bootstrap: %v", err)
	}

	if !strings.Contains(output, "Starting CloudInferOps GitOps Bootstrap") {
		t.Errorf("Expected GitOps Bootstrap starting message, got: %s", output)
	}
	if !strings.Contains(output, "[DRY-RUN]") {
		t.Errorf("Expected dry-run notices in output, got: %s", output)
	}
}
