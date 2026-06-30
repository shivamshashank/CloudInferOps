package diagnose

import (
	"bytes"
	"errors"
	"slices"
	"strings"
	"testing"
)

func TestDiagnosePodCrashLoopOOM(t *testing.T) {
	engine := Engine{
		Namespace: "payments",
		Run: fakeRunner(map[string]string{
			"get pod api-7f9 -n payments -o json": `{
				"metadata":{"name":"api-7f9"},
				"status":{
					"phase":"Running",
					"containerStatuses":[{
						"name":"api",
						"restartCount":15,
						"state":{"waiting":{"reason":"CrashLoopBackOff","message":"back-off restarting failed container"}},
						"lastState":{"terminated":{"reason":"OOMKilled","exitCode":137}}
					}]
				}
			}`,
			"get events -n payments --field-selector involvedObject.name=api-7f9,involvedObject.kind=Pod -o json": `{
				"items":[{"reason":"Killing","message":"Container api was OOMKilled","lastTimestamp":"2026-06-21T10:00:00Z"}]
			}`,
			"logs api-7f9 -n payments --previous --tail=50": "fatal: out of memory",
			"get pods -A -l app.kubernetes.io/name=alertmanager -o json": `{
				"items":[{"metadata":{"name":"alertmanager-0","namespace":"observability"},"status":{"phase":"Running"}}]
			}`,
			"logs alertmanager-0 -n observability --tail=30": "level=info msg=\"alert firing\" alert=PodCrashLooping",
		}),
	}

	report, err := engine.DiagnosePod("api-7f9")
	if err != nil {
		t.Fatal(err)
	}

	if report.Issue != "OOMKilled" {
		t.Fatalf("expected OOMKilled issue, got %q", report.Issue)
	}
	assertEvidenceContains(t, report, "Container api restarts: 15")
	assertEvidenceContains(t, report, "Last exit code: 137")
	assertEvidenceContains(t, report, "Event Killing: Container api was OOMKilled")
	assertEvidenceContains(t, report, "Previous log signal: fatal: out of memory")
	assertEvidenceContains(t, report, "Recent alert signal: level=info msg=\"alert firing\" alert=PodCrashLooping")
	if !strings.Contains(report.Recommendation, "Increase memory limits") {
		t.Fatalf("expected memory recommendation, got %q", report.Recommendation)
	}
}

func TestDiagnoseDeploymentUnavailable(t *testing.T) {
	engine := Engine{
		Namespace: "prod",
		Run: fakeRunner(map[string]string{
			"get deployment checkout -n prod -o json": `{
				"status":{
					"replicas":3,
					"readyReplicas":1,
					"unavailableReplicas":2,
					"conditions":[{"type":"Available","status":"False","reason":"MinimumReplicasUnavailable"}]
				},
				"spec":{"selector":{"matchLabels":{"app":"checkout","tier":"api"}}}
			}`,
			"get pods -n prod -l app=checkout,tier=api -o json": `{
				"items":[
					{"metadata":{"name":"checkout-a"},"status":{"phase":"Running","containerStatuses":[{"name":"api","restartCount":0}]}},
					{"metadata":{"name":"checkout-b"},"status":{"phase":"Pending","containerStatuses":[]}}
				]
			}`,
			"get events -n prod --field-selector involvedObject.name=checkout,involvedObject.kind=Deployment -o json": `{"items":[]}`,
		}),
	}

	report, err := engine.DiagnoseDeployment("checkout")
	if err != nil {
		t.Fatal(err)
	}

	if report.Issue != "DeploymentUnavailable" {
		t.Fatalf("expected DeploymentUnavailable, got %q", report.Issue)
	}
	assertEvidenceContains(t, report, "Ready replicas: 1/3")
	assertEvidenceContains(t, report, "Unhealthy pods: checkout-b(Pending)")
}

func TestDiagnoseServiceNoEndpoints(t *testing.T) {
	engine := Engine{
		Namespace: "prod",
		Run: fakeRunner(map[string]string{
			"get service checkout -n prod -o json": `{
				"spec":{"type":"ClusterIP","selector":{"app":"checkout"}}
			}`,
			"get endpoints checkout -n prod -o json": `{"subsets":[]}`,
			"get events -n prod --field-selector involvedObject.name=checkout,involvedObject.kind=Service -o json": `{"items":[]}`,
		}),
	}

	report, err := engine.DiagnoseService("checkout")
	if err != nil {
		t.Fatal(err)
	}

	if report.Issue != "ServiceNoEndpoints" {
		t.Fatalf("expected ServiceNoEndpoints, got %q", report.Issue)
	}
	assertEvidenceContains(t, report, "Ready endpoints: 0")
	if !strings.Contains(report.Recommendation, "match the service selector") {
		t.Fatalf("expected selector recommendation, got %q", report.Recommendation)
	}
}

func TestDiagnoseClusterNodeNotReady(t *testing.T) {
	engine := Engine{
		Run: fakeRunner(map[string]string{
			"get nodes -o json": `{
				"items":[
					{"metadata":{"name":"node-a"},"status":{"conditions":[{"type":"Ready","status":"True"}]}},
					{"metadata":{"name":"node-b"},"status":{"conditions":[{"type":"Ready","status":"False"},{"type":"MemoryPressure","status":"True"}]}}
				]
			}`,
			"get pods -A -o json": `{"items":[]}`,
			"get events -A --field-selector type=Warning -o json": `{
				"items":[{"reason":"NodeNotReady","message":"Node node-b status is now: NodeNotReady","lastTimestamp":"2026-06-21T11:00:00Z"}]
			}`,
		}),
	}

	report, err := engine.DiagnoseCluster()
	if err != nil {
		t.Fatal(err)
	}

	if report.Issue != "NodeNotReady" {
		t.Fatalf("expected NodeNotReady, got %q", report.Issue)
	}
	assertEvidenceContains(t, report, "Ready nodes: 1/2")
	assertEvidenceContains(t, report, "Not ready nodes: node-b")
	assertEvidenceContains(t, report, "Node node-b condition: MemoryPressure")
}

func TestReportPrint(t *testing.T) {
	report := Report{
		Target:         "Pod api",
		Issue:          "CrashLoopBackOff",
		Evidence:       []string{"Restart Count: 15"},
		Recommendation: "Inspect previous logs.",
	}

	var out bytes.Buffer
	report.Print(&out)

	for _, want := range []string{"Target:", "Pod api", "Issue:", "CrashLoopBackOff", "- Restart Count: 15", "Recommendation:"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out.String())
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

func assertEvidenceContains(t *testing.T, report Report, want string) {
	t.Helper()
	if !slices.Contains(report.Evidence, want) {
		t.Fatalf("expected evidence %q in %#v", want, report.Evidence)
	}
}
