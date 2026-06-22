package security

import (
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"
)

func TestScanFindsSecurityIssues(t *testing.T) {
	engine := Engine{
		Namespace: "prod",
		Run: fakeRunner(map[string]string{
			"kubectl get pods -n prod -o json": `{
				"items":[{
					"metadata":{"name":"api-1","namespace":"prod"},
					"spec":{"containers":[{"name":"api","image":"repo/api:latest","securityContext":{"privileged":true}}]}
				}]
			}`,
			"kubectl get deployments -n prod -o json": `{
				"items":[{
					"metadata":{"name":"api","namespace":"prod"},
					"spec":{"template":{"spec":{"containers":[{"name":"api","image":"repo/api:latest","resources":{},"securityContext":{"privileged":true}}]}}}
				}]
			}`,
			"kubectl get services -n prod -o json": `{
				"items":[{"metadata":{"name":"api","namespace":"prod"},"spec":{"type":"LoadBalancer"}}]
			}`,
			"trivy k8s --namespace prod --severity HIGH,CRITICAL --format json all": `{
				"Results":[{"Target":"repo/api:latest","Vulnerabilities":[
					{"VulnerabilityID":"CVE-2026-0001","PkgName":"openssl","Severity":"CRITICAL","Title":"test critical vuln"},
					{"VulnerabilityID":"CVE-2026-0002","PkgName":"curl","Severity":"MEDIUM","Title":"ignored medium vuln"}
				]}]
			}`,
			"kube-score score --namespace prod": "CRITICAL: Deployment has no network policy\nWARNING: Container has no pod disruption budget",
		}),
	}

	report, err := engine.Scan()
	if err != nil {
		t.Fatal(err)
	}

	if report.Summary.PodsChecked != 1 || report.Summary.DeploymentsChecked != 1 || report.Summary.ServicesChecked != 1 {
		t.Fatalf("unexpected summary: %#v", report.Summary)
	}
	if report.Integrations.Trivy.Status != "ok" || report.Integrations.KubeScore.Status != "ok" {
		t.Fatalf("expected integrations ok, got %#v", report.Integrations)
	}
	assertFinding(t, report, "privileged_container")
	assertFinding(t, report, "missing_resource_limits")
	assertFinding(t, report, "latest_image_tag")
	assertFinding(t, report, "public_exposure")
	assertFinding(t, report, "high_severity_vulnerability")
	assertFinding(t, report, "kube_score")
	if report.Summary.CriticalFindings != 1 {
		t.Fatalf("expected one critical finding, got %#v", report.Summary)
	}
}

func TestScanOptionalIntegrationsUnavailable(t *testing.T) {
	engine := Engine{
		Namespace: "prod",
		Run: fakeRunner(map[string]string{
			"kubectl get pods -n prod -o json":        `{"items":[]}`,
			"kubectl get deployments -n prod -o json": `{"items":[]}`,
			"kubectl get services -n prod -o json":    `{"items":[]}`,
		}),
	}

	report, err := engine.Scan()
	if err != nil {
		t.Fatal(err)
	}

	if len(report.Findings) != 0 {
		t.Fatalf("expected no findings, got %#v", report.Findings)
	}
	if report.Integrations.Trivy.Status != "unavailable" {
		t.Fatalf("expected trivy unavailable, got %#v", report.Integrations.Trivy)
	}
	if report.Integrations.KubeScore.Status != "unavailable" {
		t.Fatalf("expected kube-score unavailable, got %#v", report.Integrations.KubeScore)
	}
}

func TestReportJSON(t *testing.T) {
	report := Report{
		Namespace: "prod",
		Findings: []Finding{{
			Check:          "privileged_container",
			Severity:       "high",
			Resource:       "deployment/prod/api",
			Message:        "Container api is running privileged",
			Recommendation: "Disable privileged mode.",
		}},
	}

	out, err := report.JSON()
	if err != nil {
		t.Fatal(err)
	}

	var decoded Report
	if err := json.Unmarshal(out, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Namespace != "prod" || decoded.Findings[0].Check != "privileged_container" {
		t.Fatalf("unexpected JSON report: %s", string(out))
	}
}

func TestUsesLatestTag(t *testing.T) {
	for _, image := range []string{"nginx", "repo/api:latest", "registry:5000/repo/api:latest"} {
		if !usesLatestTag(image) {
			t.Fatalf("expected %q to be latest/mutable", image)
		}
	}
	for _, image := range []string{"nginx:1.25", "registry:5000/repo/api:v1"} {
		if usesLatestTag(image) {
			t.Fatalf("expected %q to be pinned", image)
		}
	}
}

func fakeRunner(outputs map[string]string) Runner {
	return func(name string, args ...string) (string, string, error) {
		key := strings.TrimSpace(name + " " + strings.Join(args, " "))
		out, ok := outputs[key]
		if !ok {
			return "", "unexpected command: " + key, errors.New("unexpected command")
		}
		return out, "", nil
	}
}

func assertFinding(t *testing.T, report Report, check string) {
	t.Helper()
	if !slices.ContainsFunc(report.Findings, func(f Finding) bool {
		return f.Check == check
	}) {
		t.Fatalf("expected finding %q in %#v", check, report.Findings)
	}
}
