package score

import (
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"
)

func TestCalculateFindsReliabilityIssues(t *testing.T) {
	engine := Engine{
		Namespace: "prod",
		Run: fakeRunner(map[string]string{
			"get pods -n prod -o json": `{
				"items":[{
					"metadata":{"name":"api-1","namespace":"prod"},
					"spec":{"containers":[{"name":"api","image":"repo/api:latest","securityContext":{"privileged":true}}]},
					"status":{
						"phase":"Running",
						"containerStatuses":[{"name":"api","restartCount":12,"state":{"waiting":{"reason":"CrashLoopBackOff"}}}]
					}
				}]
			}`,
			"get deployments -n prod -o json": `{
				"items":[{
					"metadata":{"name":"api","namespace":"prod"},
					"spec":{"template":{"spec":{"containers":[{"name":"api","image":"repo/api:latest","resources":{},"securityContext":{"privileged":true}}]}}},
					"status":{"replicas":3,"readyReplicas":1,"unavailableReplicas":2}
				}]
			}`,
			"get services -n prod -o json": `{
				"items":[{"metadata":{"name":"api","namespace":"prod"},"spec":{"type":"LoadBalancer"}}]
			}`,
			"get prometheusrules.monitoring.coreos.com -n prod -o json": `{"items":[]}`,
		}),
	}

	report, err := engine.Calculate()
	if err != nil {
		t.Fatal(err)
	}

	if report.Score != 0 {
		t.Fatalf("expected score 0 after capped deductions, got %d with issues %#v", report.Score, report.Issues)
	}
	if report.Summary.PodsChecked != 1 || report.Summary.DeploymentsChecked != 1 || report.Summary.ServicesChecked != 1 {
		t.Fatalf("unexpected summary: %#v", report.Summary)
	}
	assertIssueContains(t, report, "High restart rate")
	assertIssueContains(t, report, "missing readiness and liveness probe")
	assertIssueContains(t, report, "missing resource requests and limits")
	assertIssueContains(t, report, "No SLO coverage")
	assertIssueContains(t, report, "privileged")
	assertIssueContains(t, report, "latest image tag")
	assertIssueContains(t, report, "publicly exposed")
}

func TestCalculateHealthyNamespace(t *testing.T) {
	engine := Engine{
		Namespace: "prod",
		Run: fakeRunner(map[string]string{
			"get pods -n prod -o json": `{
				"items":[{
					"metadata":{"name":"api-1","namespace":"prod"},
					"spec":{"containers":[{"name":"api","image":"repo/api:v1.2.3","resources":{"requests":{"cpu":"100m","memory":"128Mi"},"limits":{"cpu":"500m","memory":"512Mi"}}}]},
					"status":{"phase":"Running","containerStatuses":[{"name":"api","restartCount":0}]}
				}]
			}`,
			"get deployments -n prod -o json": `{
				"items":[{
					"metadata":{"name":"api","namespace":"prod"},
					"spec":{"template":{"spec":{"containers":[{
						"name":"api",
						"image":"repo/api:v1.2.3",
						"readinessProbe":{"httpGet":{}},
						"livenessProbe":{"httpGet":{}},
						"resources":{"requests":{"cpu":"100m","memory":"128Mi"},"limits":{"cpu":"500m","memory":"512Mi"}}
					}]}}},
					"status":{"replicas":2,"readyReplicas":2,"unavailableReplicas":0}
				}]
			}`,
			"get services -n prod -o json": `{
				"items":[{"metadata":{"name":"api","namespace":"prod"},"spec":{"type":"ClusterIP"}}]
			}`,
			"get prometheusrules.monitoring.coreos.com -n prod -o json": `{
				"items":[{"metadata":{"name":"api-slo","namespace":"prod"}}]
			}`,
		}),
	}

	report, err := engine.Calculate()
	if err != nil {
		t.Fatal(err)
	}

	if report.Score != 100 {
		t.Fatalf("expected score 100, got %d", report.Score)
	}
	if len(report.Issues) != 0 {
		t.Fatalf("expected no issues, got %#v", report.Issues)
	}
}

func TestReportJSON(t *testing.T) {
	report := Report{
		Score:     82,
		MaxScore:  100,
		Namespace: "prod",
		Issues: []Finding{{
			Category:  "probe_coverage",
			Severity:  "medium",
			Message:   "Missing readiness probes",
			Deduction: 7,
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
	if decoded.Score != 82 || decoded.Issues[0].Category != "probe_coverage" {
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
	return func(args ...string) (string, string, error) {
		key := strings.Join(args, " ")
		out, ok := outputs[key]
		if !ok {
			return "", "unexpected command: " + key, errors.New("unexpected command")
		}
		return out, "", nil
	}
}

func assertIssueContains(t *testing.T, report Report, want string) {
	t.Helper()
	if !slices.ContainsFunc(report.Issues, func(f Finding) bool {
		return strings.Contains(f.Message, want)
	}) {
		t.Fatalf("expected issue containing %q in %#v", want, report.Issues)
	}
}
