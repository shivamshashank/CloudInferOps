package observability

import (
	"encoding/base64"
	"fmt"
	"net"
	"strings"

	"github.com/shivamshashank/CloudInferOps/internal/config"
	"github.com/shivamshashank/CloudInferOps/internal/utils"
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
	argoStatus := "⚪  Not Deployed"
	vmStatus := "⚪  Not Deployed"
	pyroscopeStatus := "⚪  Not Deployed"
	thanosStatus := "⚪  Not Deployed"
	blackboxStatus := "⚪  Not Deployed"
	alertmanagerStatus := "⚪  Not Deployed"
	uiStatus := "⚪  Not Deployed"

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
			} else if strings.Contains(podName, "argocd") {
				argoStatus = statusStr
			} else if strings.Contains(podName, "victoria-metrics") || strings.Contains(podName, "vmsingle") || strings.Contains(podName, "vmcluster") {
				vmStatus = statusStr
			} else if strings.Contains(podName, "pyroscope") {
				pyroscopeStatus = statusStr
			} else if strings.Contains(podName, "thanos") {
				thanosStatus = statusStr
			} else if strings.Contains(podName, "blackbox-exporter") {
				blackboxStatus = statusStr
			} else if strings.Contains(podName, "alertmanager") {
				alertmanagerStatus = statusStr
			} else if strings.Contains(podName, "cloudinferops-ui") {
				uiStatus = statusStr
			}
		}
	}

	// 2b. Fetch Inference Pod States
	inferencePodsOut, _, err := utils.ExecCommand("", "kubectl", "get", "pods", "-n", "inference", "--no-headers")
	hasInferencePods := err == nil && inferencePodsOut != ""

	inferenceGatewayStatus := "⚪  Not Deployed"
	inferenceBackendStatus := "⚪  Not Deployed"

	if hasInferencePods {
		lines := strings.Split(strings.TrimSpace(inferencePodsOut), "\n")
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

			if strings.Contains(podName, "cloudinferops-gateway") || strings.Contains(podName, "gateway") {
				inferenceGatewayStatus = statusStr
			} else if strings.Contains(podName, "ollama") {
				inferenceBackendStatus = "🟢  Running (Ollama)"
				if podState != "Running" {
					inferenceBackendStatus = "🔴  Failed (Ollama) (" + podState + ")"
				}
			} else if strings.Contains(podName, "vllm") {
				inferenceBackendStatus = "🟢  Running (vLLM)"
				if podState != "Running" {
					inferenceBackendStatus = "🔴  Failed (vLLM) (" + podState + ")"
				}
			}
		}
	}

	// 3. Fetch and Decode Grafana Admin Password
	plainPass := "<unretrievable>"
	pwdSecret, _, err := utils.ExecCommand("", "kubectl", "get", "secret", "cloudinferops-prometheus-grafana", "-n", ns, "-o", "jsonpath={.data.admin-password}")
	if err == nil && pwdSecret != "" {
		decoded, err := DecodeBase64(strings.TrimSpace(pwdSecret))
		if err == nil {
			plainPass = decoded
		}
	}

	// 3b. Fetch and Decode ArgoCD Admin Password
	argoPass := "<unretrievable>"
	argoSecName := "argocd-initial-admin-secret" // Default fallback
	if out, _, err := utils.ExecCommand("", "kubectl", "get", "secret", "-n", ns, "-o", "name"); err == nil {
		for _, line := range strings.Split(out, "\n") {
			if strings.Contains(line, "initial-admin-secret") {
				argoSecName = strings.TrimPrefix(strings.TrimSpace(line), "secret/")
				break
			}
		}
	}
	argoSecret, _, err := utils.ExecCommand("", "kubectl", "get", "secret", argoSecName, "-n", ns, "-o", "jsonpath={.data.password}")
	if err == nil && argoSecret != "" {
		decoded, err := DecodeBase64(strings.TrimSpace(argoSecret))
		if err == nil {
			argoPass = decoded
		}
	}

	// 4. Output Unified Dashboard
	fmt.Println()
	fmt.Printf("%s%s🩺  CloudInferOps Status Dashboard%s\n", utils.PrefixInfo, utils.ColorBold, utils.ColorReset)
	fmt.Println("-----------------------------------------------------------------")
	fmt.Printf("🌐  Kubernetes Context:   %s\n", utils.ColorBold+context+utils.ColorReset)
	fmt.Printf("📦  Namespace:            %s\n", ns)
	fmt.Println()

	fmt.Println("📋  System Components Checklist:")
	fmt.Printf("    %-25s %s\n", "Prometheus Server:", prometheusStatus)
	if config.GlobalConfig.Observability.VictoriaMetrics || vmStatus != "⚪  Not Deployed" {
		fmt.Printf("    %-25s %s\n", "VictoriaMetrics:", vmStatus)
	}
	fmt.Printf("    %-25s %s\n", "Grafana Dashboard:", grafanaStatus)
	fmt.Printf("    %-25s %s\n", "Loki Logging:", lokiStatus)
	fmt.Printf("    %-25s %s\n", "Tempo Tracing:", tempoStatus)
	fmt.Printf("    %-25s %s\n", "OTel Collector:", otelStatus)
	if config.GlobalConfig.Observability.Alertmanager || alertmanagerStatus != "⚪  Not Deployed" {
		fmt.Printf("    %-25s %s\n", "Alertmanager:", alertmanagerStatus)
	}
	if config.GlobalConfig.Observability.BlackboxExporter || blackboxStatus != "⚪  Not Deployed" {
		fmt.Printf("    %-25s %s\n", "Blackbox Exporter:", blackboxStatus)
	}
	if config.GlobalConfig.Observability.Pyroscope || pyroscopeStatus != "⚪  Not Deployed" {
		fmt.Printf("    %-25s %s\n", "Pyroscope Profiling:", pyroscopeStatus)
	}
	if config.GlobalConfig.Observability.Thanos || thanosStatus != "⚪  Not Deployed" {
		fmt.Printf("    %-25s %s\n", "Thanos Storage:", thanosStatus)
	}
	fmt.Printf("    %-25s %s\n", "ArgoCD Delivery:", argoStatus)
	fmt.Printf("    %-25s %s\n", "UI Portal:", uiStatus)
	fmt.Println()

	fmt.Println("🤖  Inference Services:")
	fmt.Printf("    %-25s %s\n", "Inference Gateway:", inferenceGatewayStatus)
	fmt.Printf("    %-25s %s\n", "Model Backend:", inferenceBackendStatus)
	fmt.Println()

	// 5. GitOps Status
	gitOpsMode := "Local Helm"
	appCount := 0
	syncedCount := 0
	healthyCount := 0

	hasGitOpsServer := false
	if gitServerOut, _, err := utils.ExecCommand("", "kubectl", "get", "deployment", "cloudinferops-git-server", "-n", ns, "--no-headers"); err == nil && gitServerOut != "" {
		hasGitOpsServer = true
	}

	if config.GlobalConfig.Observability.ArgoCD {
		if hasGitOpsServer {
			gitOpsMode = "Local GitOps"
		} else {
			gitOpsMode = "ArgoCD Managed"
		}

		appsOut, _, err := utils.ExecCommand("", "kubectl", "get", "applications", "-n", ns, "--no-headers")
		if err == nil && appsOut != "" {
			lines := strings.Split(strings.TrimSpace(appsOut), "\n")
			appCount = len(lines)
			for _, line := range lines {
				if strings.Contains(line, "Synced") {
					syncedCount++
				}
				if strings.Contains(line, "Healthy") {
					healthyCount++
				}
			}
		}
	}

	fmt.Println("📦  GitOps Overview:")
	fmt.Printf("    %-25s %s\n", "Mode:", utils.ColorCyan+gitOpsMode+utils.ColorReset)
	if config.GlobalConfig.Observability.ArgoCD {
		fmt.Printf("    %-25s %d\n", "Applications:", appCount)
		fmt.Printf("    %-25s %d/%d\n", "Synced:", syncedCount, appCount)
		fmt.Printf("    %-25s %d/%d\n", "Healthy:", healthyCount, appCount)
	}
	fmt.Println()

	hasIngressURLs := config.GlobalConfig.Observability.Prometheus || uiStatus != "⚪  Not Deployed" || inferenceGatewayStatus != "⚪  Not Deployed"

	if hasIngressURLs {
		// Fetch instance IP for active display
		instanceIP := "127.0.0.1"
		ingressIP, err := FetchIngressIP(ns, false)
		if err == nil && ingressIP != "" {
			instanceIP = ingressIP
		}

		// If we are on a cloud VM or running locally, the ingress IP might be a private subnet IP or loopback.
		// Attempt to resolve the public IP for correct external browser access.
		var detectedPublicIP string
		if parsedIP := net.ParseIP(instanceIP); parsedIP != nil && (parsedIP.IsPrivate() || parsedIP.IsLoopback()) {
			detectedPublicIP = utils.GetPublicIP()
			if detectedPublicIP != "" {
				instanceIP = detectedPublicIP
			}
		}

		fmt.Println("📊  Access Dashboards & APIs via Ingress:")
		if config.GlobalConfig.Observability.Prometheus {
			fmt.Printf("    🔗  Grafana Dashboard:   %s\n", utils.ColorBold+fmt.Sprintf("http://%s/grafana/", instanceIP)+utils.ColorReset)
			fmt.Printf("    🔗  Prometheus Server:   %s\n", utils.ColorBold+fmt.Sprintf("http://%s/prometheus/", instanceIP)+utils.ColorReset)
			fmt.Printf("    🔗  Alertmanager Panel:  %s\n", utils.ColorBold+fmt.Sprintf("http://%s/alertmanager/", instanceIP)+utils.ColorReset)
		}
		if config.GlobalConfig.Observability.ArgoCD {
			fmt.Printf("    🔗  ArgoCD Dashboard:    %s\n", utils.ColorBold+fmt.Sprintf("http://%s/argocd", instanceIP)+utils.ColorReset)
		}
		if uiStatus != "⚪  Not Deployed" {
			fmt.Printf("    🔗  UI Portal:           %s\n", utils.ColorBold+fmt.Sprintf("http://%s/cloudinferops/", instanceIP)+utils.ColorReset)
		}
		if inferenceGatewayStatus != "⚪  Not Deployed" {
			fmt.Printf("    🔗  Inference Gateway:   %s\n", utils.ColorBold+fmt.Sprintf("http://%s/v1", instanceIP)+utils.ColorReset)
		}

		fmt.Println()
		if config.GlobalConfig.Observability.Prometheus {
			fmt.Printf("    🔑  Username:            admin\n")
			fmt.Printf("    🔑  Grafana Password:    %s\n", utils.ColorGreen+plainPass+utils.ColorReset)
		}
		if config.GlobalConfig.Observability.ArgoCD {
			fmt.Printf("    🔑  ArgoCD Password:     %s\n", utils.ColorGreen+argoPass+utils.ColorReset)
		}
		fmt.Println()
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
