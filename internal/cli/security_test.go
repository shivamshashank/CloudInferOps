package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestSecurityCommandWiring(t *testing.T) {
	if securityCmd.Use != "security" {
		t.Fatalf("expected security command, got %q", securityCmd.Use)
	}
	if securityCmd.PersistentFlags().Lookup("json") == nil {
		t.Fatal("expected --json flag")
	}
	if securityCmd.PersistentFlags().Lookup("namespace") == nil {
		t.Fatal("expected --namespace flag")
	}
	if err := validateSecurityCommands(); err != nil {
		t.Fatal(err)
	}
}

func TestSecurityScanCommandRuns(t *testing.T) {
	oldNamespace := securityNamespace
	oldJSON := securityJSON
	securityNamespace = "prod"
	securityJSON = false
	t.Cleanup(func() {
		securityNamespace = oldNamespace
		securityJSON = oldJSON
	})

	setupSecurityTools(t)

	if err := securityScanCmd.RunE(&cobra.Command{}, nil); err != nil {
		t.Fatal(err)
	}
}

func TestSecurityReportCommandRunsJSON(t *testing.T) {
	oldNamespace := securityNamespace
	oldJSON := securityJSON
	securityNamespace = "prod"
	securityJSON = true
	t.Cleanup(func() {
		securityNamespace = oldNamespace
		securityJSON = oldJSON
	})

	setupSecurityTools(t)

	if err := securityReportCmd.RunE(&cobra.Command{}, nil); err != nil {
		t.Fatal(err)
	}
}

func TestResolveSecurityNamespaceUsesFlag(t *testing.T) {
	oldNamespace := securityNamespace
	securityNamespace = "payments"
	t.Cleanup(func() {
		securityNamespace = oldNamespace
	})

	if got := resolveSecurityNamespace(); got != "payments" {
		t.Fatalf("expected namespace from flag, got %q", got)
	}
}

func setupSecurityTools(t *testing.T) {
	t.Helper()

	tmpDir := t.TempDir()
	mockBinDir := filepath.Join(tmpDir, "bin")
	if err := os.MkdirAll(mockBinDir, 0755); err != nil {
		t.Fatal(err)
	}

	writeExecutable(t, filepath.Join(mockBinDir, "kubectl"), `#!/bin/sh
case "$*" in
  "get pods -n prod -o json")
    printf '{"items":[{"metadata":{"name":"api-1","namespace":"prod"},"spec":{"containers":[{"name":"api","image":"repo/api:v1"}]}}]}'
    ;;
  "get deployments -n prod -o json")
    printf '{"items":[{"metadata":{"name":"api","namespace":"prod"},"spec":{"template":{"spec":{"containers":[{"name":"api","image":"repo/api:v1","resources":{"requests":{"cpu":"100m","memory":"128Mi"},"limits":{"cpu":"500m","memory":"512Mi"}}}]}}}}]}'
    ;;
  "get services -n prod -o json")
    printf '{"items":[{"metadata":{"name":"api","namespace":"prod"},"spec":{"type":"ClusterIP"}}]}'
    ;;
  *)
    echo "unexpected kubectl args: $*" >&2
    exit 1
    ;;
esac
`)

	writeExecutable(t, filepath.Join(mockBinDir, "trivy"), `#!/bin/sh
printf '{"Results":[]}'
`)

	writeExecutable(t, filepath.Join(mockBinDir, "kube-score"), `#!/bin/sh
printf 'OK'
`)

	t.Setenv("PATH", mockBinDir+":"+os.Getenv("PATH"))
}

func writeExecutable(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0755); err != nil {
		t.Fatal(err)
	}
}
