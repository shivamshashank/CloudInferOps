package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestDiagnoseCommandWiring(t *testing.T) {
	if diagnoseCmd.Use != "diagnose" {
		t.Fatalf("expected diagnose command, got %q", diagnoseCmd.Use)
	}
	if err := validateDiagnoseCommands(); err != nil {
		t.Fatal(err)
	}
}

func TestDiagnosePodCommandRuns(t *testing.T) {
	oldNamespace := diagnoseNamespace
	diagnoseNamespace = "prod"
	t.Cleanup(func() {
		diagnoseNamespace = oldNamespace
	})

	tmpDir := t.TempDir()
	mockBinDir := filepath.Join(tmpDir, "bin")
	if err := os.MkdirAll(mockBinDir, 0755); err != nil {
		t.Fatal(err)
	}

	mockKubectlPath := filepath.Join(mockBinDir, "kubectl")
	mockKubectlContent := `#!/bin/sh
case "$*" in
  "get pod api -n prod -o json")
    printf '{"metadata":{"name":"api"},"status":{"phase":"Running","containerStatuses":[{"name":"api","restartCount":4,"state":{"waiting":{"reason":"CrashLoopBackOff","message":"back-off"}}}]}}'
    ;;
  "get events -n prod --field-selector involvedObject.name=api,involvedObject.kind=Pod -o json")
    printf '{"items":[{"reason":"BackOff","message":"Back-off restarting failed container","lastTimestamp":"2026-06-21T10:00:00Z"}]}'
    ;;
  "logs api -n prod --previous --tail=50")
    printf 'panic: missing DATABASE_URL'
    ;;
  *)
    echo "unexpected kubectl args: $*" >&2
    exit 1
    ;;
esac
`
	if err := os.WriteFile(mockKubectlPath, []byte(mockKubectlContent), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", mockBinDir+":"+os.Getenv("PATH"))

	if err := diagnosePodCmd.RunE(&cobra.Command{}, []string{"api"}); err != nil {
		t.Fatal(err)
	}
}

func TestResolveDiagnoseNamespaceUsesFlag(t *testing.T) {
	oldNamespace := diagnoseNamespace
	diagnoseNamespace = "payments"
	t.Cleanup(func() {
		diagnoseNamespace = oldNamespace
	})

	if got := resolveDiagnoseNamespace(); got != "payments" {
		t.Fatalf("expected namespace from flag, got %q", got)
	}
}
