package doctor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDoctorReportAndRun(t *testing.T) {
	// Prepend a mock kubectl to PATH to make K8s cluster check run successfully or predictably
	tmpDir := t.TempDir()
	mockBinDir := filepath.Join(tmpDir, "bin")
	if err := os.MkdirAll(mockBinDir, 0755); err != nil {
		t.Fatalf("failed to create mock bin dir: %v", err)
	}

	mockKubectlPath := filepath.Join(mockBinDir, "kubectl")
	mockKubectlContent := `#!/bin/sh
exit 0
`
	if err := os.WriteFile(mockKubectlPath, []byte(mockKubectlContent), 0755); err != nil { //nolint:gosec
		t.Fatalf("failed to write mock kubectl: %v", err)
	}

	oldPath := os.Getenv("PATH")
	t.Setenv("PATH", mockBinDir+":"+oldPath)

	// Run aggregated diagnostics
	report := RunDoctor()
	if report == nil {
		t.Fatal("expected non-nil Report")
	}

	// Capture stdout of Print()
	report.Print()

	// Test print variations manually by mutating status
	report.HasErrors = true
	report.Print()

	report.HasErrors = false
	report.HasK8s = true
	report.Print()

	report.HasK8s = false
	report.Results = []CheckResult{
		{Name: "CheckOK", Status: StatusOK, Message: "ok"},
		{Name: "CheckWarn", Status: StatusWarn, Message: "warn"},
		{Name: "CheckError", Status: StatusError, Message: "error"},
		{Name: "CheckInfo", Status: StatusInfo, Message: "info"},
	}
	report.Print()
}
