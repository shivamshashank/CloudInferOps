package observability

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/shivamshashank/StackPulse/internal/config"
	"github.com/shivamshashank/StackPulse/internal/utils"
)

// PrintStatus queries the active cluster to collect pod health, decrypt secrets, and print a unified dashboard.
func PrintStatus() error {
	ns := config.GlobalConfig.Namespace
	if ns == "" {
		ns = "observability"
	}

	// 1. Get Kubernetes Active Context
	context, _, err := utils.ExecCommand("", "kubectl", "config", "current-context")
	if err != nil || context == "" {
		context = "unknown"
	}

	// 2. Fetch Pod States
	podsOut, _, err := utils.ExecCommand("", "kubectl", "get", "pods", "-n", ns, "--no-headers")
	hasPods := err == nil && podsOut != ""

	prometheusStatus := "⚪  Not Deployed"
	grafanaStatus := "⚪  Not Deployed"
	lokiStatus := "⚪  Not Deployed"
	tempoStatus := "⚪  Not Deployed"
	otelStatus := "⚪  Not Deployed"
	webhookStatus := "⚪  Not Deployed"

	if hasPods {
		lines := strings.Split(strings.TrimSpace(podsOut), "\n")
		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) < 3 {
				continue
			}
			podName := fields[0]
			podState := fields[2]

			statusStr := "🔴  Failed (" + podState + ")"
			switch podState {
			case "Running":
				statusStr = "🟢  Running"
			case "Pending", "ContainerCreating", "PodInitializing":
				statusStr = "🟡  Initializing"
			}

			if strings.Contains(podName, "prometheus-server") || strings.Contains(podName, "prometheus-prometheus") {
				prometheusStatus = statusStr
			} else if strings.Contains(podName, "grafana") {
				grafanaStatus = statusStr
			} else if strings.Contains(podName, "loki") {
				lokiStatus = statusStr
			} else if strings.Contains(podName, "tempo") {
				tempoStatus = statusStr
			} else if strings.Contains(podName, "opentelemetry") || strings.Contains(podName, "otel") {
				otelStatus = statusStr
			} else if strings.Contains(podName, "webhook-handler") {
				webhookStatus = statusStr
			}
		}
	}

	// 3. Fetch and Decode Grafana Admin Password
	plainPassword := "<unretrievable>"
	pwdSecret, _, err := utils.ExecCommand("", "kubectl", "get", "secret", "stackpulse-prometheus-grafana", "-n", ns, "-o", "jsonpath={.data.admin-password}")
	if err == nil && pwdSecret != "" {
		decoded, err := DecodeBase64(strings.TrimSpace(pwdSecret))
		if err == nil {
			plainPassword = decoded
		}
	}

	// 4. Output Unified Dashboard
	fmt.Println()
	fmt.Printf("%s%s🩺  StackPulse Status Dashboard%s\n", utils.PrefixInfo, utils.ColorBold, utils.ColorReset)
	fmt.Println("-----------------------------------------------------------------")
	fmt.Printf("🌐  Kubernetes Context:   %s\n", utils.ColorBold+context+utils.ColorReset)
	fmt.Printf("📦  Namespace:            %s\n", ns)
	fmt.Println()

	fmt.Println("📋  System Components Checklist:")
	fmt.Printf("    %-25s %s\n", "Prometheus Server:", prometheusStatus)
	fmt.Printf("    %-25s %s\n", "Grafana Dashboard:", grafanaStatus)
	fmt.Printf("    %-25s %s\n", "Loki Logging:", lokiStatus)
	fmt.Printf("    %-25s %s\n", "Tempo Tracing:", tempoStatus)
	fmt.Printf("    %-25s %s\n", "OTel Collector:", otelStatus)
	fmt.Printf("    %-25s %s\n", "Custom Webhook Handler:", webhookStatus)
	fmt.Println()

	if config.GlobalConfig.Observability.Prometheus {
		// Fetch instance IP for active display
		instanceIP := "127.0.0.1"
		ingressIP, err := FetchIngressIP(ns, false)
		if err == nil && ingressIP != "" {
			instanceIP = ingressIP
		} else {
			context, _, _ := utils.ExecCommand("", "kubectl", "config", "current-context")
			context = strings.TrimSpace(context)
			if strings.Contains(context, "minikube") {
				minikubeIP, _, err := utils.ExecCommand("", "minikube", "ip")
				if err == nil && minikubeIP != "" {
					instanceIP = strings.TrimSpace(minikubeIP)
				}
			}
		}

		fmt.Println("📊  Access Telemetry Dashboards via Ingress:")
		fmt.Printf("    🔗  Grafana Dashboard:   %s\n", utils.ColorBold+fmt.Sprintf("http://%s/grafana/", instanceIP)+utils.ColorReset)
		fmt.Printf("    🔗  Prometheus Server:   %s\n", utils.ColorBold+fmt.Sprintf("http://%s/prometheus/", instanceIP)+utils.ColorReset)
		fmt.Printf("    🔗  Alertmanager Panel:  %s\n", utils.ColorBold+fmt.Sprintf("http://%s/alertmanager/", instanceIP)+utils.ColorReset)
		fmt.Printf("    🔑  Username:            admin\n")
		fmt.Printf("    🔑  Password:            %s\n", utils.ColorGreen+plainPassword+utils.ColorReset)
		fmt.Println()
	}

	if config.GlobalConfig.Alerts.Slack.Enabled || config.GlobalConfig.Alerts.PagerDuty.Enabled {
		fmt.Println("📡  Custom Webhook Incident REST APIs:")
		fmt.Printf("    🔗  Liveness Probe:      GET  http://stackpulse-webhook-handler.%s.svc.cluster.local/health\n", ns)
		fmt.Printf("    🔗  Alertmanager Router: POST http://stackpulse-webhook-handler.%s.svc.cluster.local/webhook/alertmanager\n", ns)
		fmt.Printf("    🔗  Incident Logs API:   GET  http://stackpulse-webhook-handler.%s.svc.cluster.local/incidents\n", ns)
	}

	fmt.Println("-----------------------------------------------------------------")
	return nil
}

// DecodeBase64 safely decodes a base64 string
func DecodeBase64(input string) (string, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return "", err
	}
	return string(decodedBytes), nil
}
