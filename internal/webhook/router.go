package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// AlertmanagerPayload represents the incoming schema from Prometheus Alertmanager
type AlertmanagerPayload struct {
	Receiver          string            `json:"receiver"`
	Status            string            `json:"status"` // firing / resolved
	Alerts            []Alert           `json:"alerts"`
	GroupLabels       map[string]string `json:"groupLabels"`
	CommonLabels      map[string]string `json:"commonLabels"`
	CommonAnnotations map[string]string `json:"commonAnnotations"`
	ExternalURL       string            `json:"externalURL"`
}

type Alert struct {
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     time.Time         `json:"startsAt"`
	EndsAt       time.Time         `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
	Fingerprint  string            `json:"fingerprint"`
}

// Incident represents the internal parsed record stored in memory
type Incident struct {
	ID          string    `json:"id"`
	AlertName   string    `json:"alertname"`
	Status      string    `json:"status"`
	Severity    string    `json:"severity"`
	Summary     string    `json:"summary"`
	Description string    `json:"description"`
	Instance    string    `json:"instance"`
	Timestamp   time.Time `json:"timestamp"`
}

// ParseAlertmanagerPayload converts the payload into in-memory Incident models and dispatches external notifications.
func ParseAlertmanagerPayload(payload AlertmanagerPayload) []Incident {
	var incidents []Incident

	for _, alert := range payload.Alerts {
		incident := Incident{
			ID:          alert.Fingerprint,
			AlertName:   alert.Labels["alertname"],
			Status:      alert.Status,
			Severity:    alert.Labels["severity"],
			Summary:     alert.Annotations["summary"],
			Description: alert.Annotations["description"],
			Instance:    alert.Labels["instance"],
			Timestamp:   alert.StartsAt,
		}

		if incident.AlertName == "" {
			incident.AlertName = "UnknownAlert"
		}
		if incident.Severity == "" {
			incident.Severity = "warning"
		}
		if incident.Status == "" {
			incident.Status = payload.Status
		}

		incidents = append(incidents, incident)

		// Dispatch external alerts asynchronously
		go DispatchSlackAlert(incident)
		go DispatchPagerDutyAlert(incident)
	}

	return incidents
}

// DispatchSlackAlert formats a rich Slack block kit message and sends it to the SLACK_WEBHOOK_URL env var
func DispatchSlackAlert(inc Incident) {
	slackWebhook := os.Getenv("SLACK_WEBHOOK_URL")
	if slackWebhook == "" {
		return
	}

	color := "#E01E5A" // red for firing
	emoji := "🔴 [FIRING]"
	if inc.Status == "resolved" {
		color = "#2EB67D" // green for resolved
		emoji = "🟢 [RESOLVED]"
	}

	// Structured rich message payload
	message := map[string]interface{}{
		"attachments": []map[string]interface{}{
			{
				"color":      color,
				"title":      fmt.Sprintf("%s %s on %s", emoji, inc.AlertName, inc.Instance),
				"fallback":   fmt.Sprintf("%s Alert: %s - %s", emoji, inc.AlertName, inc.Summary),
				"fields": []map[string]interface{}{
					{
						"title": "Severity",
						"value": stringsToUpper(inc.Severity),
						"short": true,
					},
					{
						"title": "Summary",
						"value": inc.Summary,
						"short": false,
					},
					{
						"title": "Description",
						"value": inc.Description,
						"short": false,
					},
				},
				"ts": inc.Timestamp.Unix(),
			},
		},
	}

	bodyBytes, err := json.Marshal(message)
	if err != nil {
		return
	}

	resp, err := http.Post(slackWebhook, "application/json", bytes.NewBuffer(bodyBytes))
	if err == nil {
		resp.Body.Close()
	}
}

// DispatchPagerDutyAlert constructs PagerDuty Events V2 payload and triggers/resolves active alerts
func DispatchPagerDutyAlert(inc Incident) {
	pdKey := os.Getenv("PAGERDUTY_INTEGRATION_KEY")
	if pdKey == "" {
		return
	}

	action := "trigger"
	if inc.Status == "resolved" {
		action = "resolve"
	}

	message := map[string]interface{}{
		"routing_key":  pdKey,
		"event_action": action,
		"dedup_key":    inc.ID,
		"payload": map[string]interface{}{
			"summary":       fmt.Sprintf("%s on %s: %s", inc.AlertName, inc.Instance, inc.Summary),
			"source":        inc.Instance,
			"severity":      mapPDSeverity(inc.Severity),
			"timestamp":     inc.Timestamp.Format(time.RFC3339),
			"component":     "Kubernetes",
			"group":         "Observability",
			"class":         "Infrastructure",
			"custom_details": map[string]string{
				"description": inc.Description,
			},
		},
	}

	bodyBytes, err := json.Marshal(message)
	if err != nil {
		return
	}

	resp, err := http.Post("https://events.pagerduty.com/v2/enqueue", "application/json", bytes.NewBuffer(bodyBytes))
	if err == nil {
		resp.Body.Close()
	}
}

func stringsToUpper(s string) string {
	var b bytes.Buffer
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' {
			c -= 'a' - 'A'
		}
		b.WriteByte(c)
	}
	return b.String()
}

func mapPDSeverity(s string) string {
	switch s {
	case "critical":
		return "critical"
	case "warning":
		return "warning"
	case "info":
		return "info"
	default:
		return "error"
	}
}
