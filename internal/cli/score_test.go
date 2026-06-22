package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestScoreCommandWiring(t *testing.T) {
	if scoreCmd.Use != "score" {
		t.Fatalf("expected score command, got %q", scoreCmd.Use)
	}
	if scoreCmd.Flags().Lookup("json") == nil {
		t.Fatal("expected --json flag")
	}
	if scoreCmd.Flags().Lookup("namespace") == nil {
		t.Fatal("expected --namespace flag")
	}
}

func TestScoreCommandRunsHumanOutput(t *testing.T) {
	oldNamespace := scoreNamespace
	oldJSON := scoreJSON
	scoreNamespace = "prod"
	scoreJSON = false
	t.Cleanup(func() {
		scoreNamespace = oldNamespace
		scoreJSON = oldJSON
	})

	setupScoreKubectl(t)

	if err := scoreCmd.RunE(&cobra.Command{}, nil); err != nil {
		t.Fatal(err)
	}
}

func TestScoreCommandRunsJSONOutput(t *testing.T) {
	oldNamespace := scoreNamespace
	oldJSON := scoreJSON
	scoreNamespace = "prod"
	scoreJSON = true
	t.Cleanup(func() {
		scoreNamespace = oldNamespace
		scoreJSON = oldJSON
	})

	setupScoreKubectl(t)

	if err := scoreCmd.RunE(&cobra.Command{}, nil); err != nil {
		t.Fatal(err)
	}
}

func TestResolveScoreNamespaceUsesFlag(t *testing.T) {
	oldNamespace := scoreNamespace
	scoreNamespace = "payments"
	t.Cleanup(func() {
		scoreNamespace = oldNamespace
	})

	if got := resolveScoreNamespace(); got != "payments" {
		t.Fatalf("expected namespace from flag, got %q", got)
	}
}

func setupScoreKubectl(t *testing.T) {
	t.Helper()

	tmpDir := t.TempDir()
	mockBinDir := filepath.Join(tmpDir, "bin")
	if err := os.MkdirAll(mockBinDir, 0755); err != nil {
		t.Fatal(err)
	}

	mockKubectlPath := filepath.Join(mockBinDir, "kubectl")
	mockKubectlContent := `#!/bin/sh
case "$*" in
  "get pods -n prod -o json")
    printf '{"items":[{"metadata":{"name":"api-1","namespace":"prod"},"spec":{"containers":[{"name":"api","image":"repo/api:v1"}]},"status":{"phase":"Running","containerStatuses":[{"name":"api","restartCount":0}]}}]}'
    ;;
  "get deployments -n prod -o json")
    printf '{"items":[{"metadata":{"name":"api","namespace":"prod"},"spec":{"template":{"spec":{"containers":[{"name":"api","image":"repo/api:v1","readinessProbe":{"httpGet":{}},"livenessProbe":{"httpGet":{}},"resources":{"requests":{"cpu":"100m","memory":"128Mi"},"limits":{"cpu":"500m","memory":"512Mi"}}}]}}},"status":{"replicas":1,"readyReplicas":1,"unavailableReplicas":0}}]}'
    ;;
  "get services -n prod -o json")
    printf '{"items":[{"metadata":{"name":"api","namespace":"prod"},"spec":{"type":"ClusterIP"}}]}'
    ;;
  "get prometheusrules.monitoring.coreos.com -n prod -o json")
    printf '{"items":[{"metadata":{"name":"api-slo","namespace":"prod"}}]}'
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
}
