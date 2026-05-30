package webhook

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestParseAlertmanagerPayload(t *testing.T) {
	// Mock JSON Alertmanager webhook request body
	payloadJSON := `{
		"receiver": "stackpulse-webhook",
		"status": "firing",
		"alerts": [
			{
				"status": "firing",
				"labels": {
					"alertname": "PodCrashLooping",
					"severity": "critical",
					"instance": "k8s-node-01",
					"pod": "nginx-deployment-abc123"
				},
				"annotations": {
					"summary": "Pod is repeatedly crashing",
					"description": "Pod nginx-deployment-abc123 has restarted 5 times in 10 minutes"
				},
				"startsAt": "2026-05-29T16:30:00Z",
				"fingerprint": "alert-fp-123"
			}
		],
		"commonLabels": {
			"alertname": "PodCrashLooping"
		},
		"commonAnnotations": {
			"summary": "Pod is repeatedly crashing"
		}
	}`

	var payload AlertmanagerPayload
	if err := json.Unmarshal([]byte(payloadJSON), &payload); err != nil {
		t.Fatalf("failed to unmarshal mock JSON: %v", err)
	}

	incidents := ParseAlertmanagerPayload(payload)

	if len(incidents) != 1 {
		t.Fatalf("expected 1 incident parsed, got %d", len(incidents))
	}

	inc := incidents[0]
	if inc.ID != "alert-fp-123" {
		t.Errorf("expected incident ID 'alert-fp-123', got '%s'", inc.ID)
	}
	if inc.AlertName != "PodCrashLooping" {
		t.Errorf("expected alertname 'PodCrashLooping', got '%s'", inc.AlertName)
	}
	if inc.Status != "firing" {
		t.Errorf("expected status 'firing', got '%s'", inc.Status)
	}
	if inc.Severity != "critical" {
		t.Errorf("expected severity 'critical', got '%s'", inc.Severity)
	}
	if inc.Instance != "k8s-node-01" {
		t.Errorf("expected instance 'k8s-node-01', got '%s'", inc.Instance)
	}
	if inc.Summary != "Pod is repeatedly crashing" {
		t.Errorf("expected summary 'Pod is repeatedly crashing', got '%s'", inc.Summary)
	}
	if inc.Description != "Pod nginx-deployment-abc123 has restarted 5 times in 10 minutes" {
		t.Errorf("expected description 'Pod nginx-deployment-abc123 has restarted 5 times in 10 minutes', got '%s'", inc.Description)
	}

	expectedTime, _ := time.Parse(time.RFC3339, "2026-05-29T16:30:00Z")
	if !inc.Timestamp.Equal(expectedTime) {
		t.Errorf("expected timestamp '%v', got '%v'", expectedTime, inc.Timestamp)
	}
}

func TestPDAndSlackPayloadHelpers(t *testing.T) {
	// Test uppercase helper
	if stringsToUpper("critical") != "CRITICAL" {
		t.Errorf("expected 'CRITICAL', got '%s'", stringsToUpper("critical"))
	}
	if stringsToUpper("warn-info") != "WARN-INFO" {
		t.Errorf("expected 'WARN-INFO', got '%s'", stringsToUpper("warn-info"))
	}

	// Test PagerDuty severity mapping helper
	if mapPDSeverity("critical") != "critical" {
		t.Errorf("expected critical mapping to remain critical, got '%s'", mapPDSeverity("critical"))
	}
	if mapPDSeverity("warning") != "warning" {
		t.Errorf("expected warning mapping to remain warning, got '%s'", mapPDSeverity("warning"))
	}
	if mapPDSeverity("unknown") != "error" {
		t.Errorf("expected unknown severity mapping to fallback to error, got '%s'", mapPDSeverity("unknown"))
	}
}

type mockTransport struct {
	roundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTripFunc(req)
}

func TestDispatchAlerts(t *testing.T) {
	// Intercept default HTTP client
	oldTransport := http.DefaultClient.Transport
	defer func() { http.DefaultClient.Transport = oldTransport }()

	var capturedSlackReq *http.Request
	var capturedPDReq *http.Request

	http.DefaultClient.Transport = &mockTransport{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			if strings.Contains(req.URL.String(), "slack.com") || strings.Contains(req.URL.String(), "localhost") {
				capturedSlackReq = req
			} else if strings.Contains(req.URL.String(), "pagerduty.com") {
				capturedPDReq = req
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"status":"ok"}`)),
			}, nil
		},
	}

	inc := Incident{
		ID:          "inc-123",
		AlertName:   "CPUUsageHigh",
		Status:      "firing",
		Severity:    "critical",
		Summary:     "CPU usage is high",
		Description: "CPU has been above 90% for 5 mins",
		Instance:    "server-01",
	}

	// 1. Dispatch without env variables set (should do nothing)
	t.Setenv("SLACK_WEBHOOK_URL", "")
	t.Setenv("PAGERDUTY_INTEGRATION_KEY", "")
	DispatchSlackAlert(inc)
	DispatchPagerDutyAlert(inc)
	if capturedSlackReq != nil || capturedPDReq != nil {
		t.Error("expected no requests dispatched when env variables are empty")
	}

	// 2. Dispatch with env variables set (firing)
	t.Setenv("SLACK_WEBHOOK_URL", "https://hooks.slack.com/services/test/webhook")
	t.Setenv("PAGERDUTY_INTEGRATION_KEY", "pd-integration-key-123")

	DispatchSlackAlert(inc)
	if capturedSlackReq == nil {
		t.Error("expected Slack request to be captured")
	} else if capturedSlackReq.URL.String() != "https://hooks.slack.com/services/test/webhook" {
		t.Errorf("unexpected Slack URL: %s", capturedSlackReq.URL.String())
	}

	DispatchPagerDutyAlert(inc)
	if capturedPDReq == nil {
		t.Error("expected PagerDuty request to be captured")
	} else if capturedPDReq.URL.String() != "https://events.pagerduty.com/v2/enqueue" {
		t.Errorf("unexpected PagerDuty URL: %s", capturedPDReq.URL.String())
	}

	// 3. Dispatch resolved status
	inc.Status = "resolved"
	capturedSlackReq = nil
	capturedPDReq = nil

	DispatchSlackAlert(inc)
	if capturedSlackReq == nil {
		t.Error("expected Slack resolved request to be captured")
	}

	DispatchPagerDutyAlert(inc)
	if capturedPDReq == nil {
		t.Error("expected PagerDuty resolved request to be captured")
	}
}
