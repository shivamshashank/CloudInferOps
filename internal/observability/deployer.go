package observability

import (
	"fmt"
	"strings"

	"github.com/shivamshashank/StackPulse/internal/config"
	"github.com/shivamshashank/StackPulse/internal/helm"
	"github.com/shivamshashank/StackPulse/internal/utils"
)

// DeployObservability orchestrates the step-by-step installation of the enabled observability charts
func DeployObservability(dryRun bool) error {
	ns := config.GlobalConfig.Namespace
	if ns == "" {
		ns = "observability"
	}

	fmt.Printf("%sStarting StackPulse Observability Stack Deployment...\n", utils.PrefixInfo)
	if dryRun {
		fmt.Printf("%sRunning in [DRY-RUN] mode. No changes will be made to your cluster.\n\n", utils.PrefixInfo)
	}

	// 1. Create target Kubernetes Namespace
	if dryRun {
		fmt.Printf("%s[DRY-RUN] kubectl create namespace %s\n", utils.PrefixInfo, ns)
	} else {
		fmt.Printf("%sCreating Kubernetes namespace '%s' if not exists...\n", utils.PrefixInfo, ns)
		_, _, err := utils.ExecCommand("", "kubectl", "create", "namespace", ns)
		if err != nil {
			// Swallow error if namespace already exists
			if !strings.Contains(err.Error(), "already exists") {
				fmt.Printf("%sNamespace '%s' already exists or was pre-configured.\n", utils.PrefixInfo, ns)
			}
		} else {
			fmt.Printf("%sNamespace '%s' created successfully.\n", utils.PrefixOK, ns)
		}
	}

	// 2. Add required Helm chart registries
	reposAdded := false

	if config.GlobalConfig.Observability.Prometheus {
		if err := helm.AddRepo("prometheus-community", "https://prometheus-community.github.io/helm-charts", dryRun); err != nil {
			return err
		}
		reposAdded = true
	}

	if config.GlobalConfig.Observability.Grafana || config.GlobalConfig.Observability.Loki || config.GlobalConfig.Observability.Tempo {
		if err := helm.AddRepo("grafana", "https://grafana.github.io/helm-charts", dryRun); err != nil {
			return err
		}
		reposAdded = true
	}

	if config.GlobalConfig.Observability.OpenTelemetry {
		if err := helm.AddRepo("open-telemetry", "https://open-telemetry.github.io/opentelemetry-helm-charts", dryRun); err != nil {
			return err
		}
		reposAdded = true
	}

	// Only update repositories if registries were registered
	if reposAdded {
		if err := helm.UpdateRepos(dryRun); err != nil {
			return err
		}
	}

	// 3. Deploy Stack Components
	
	// A. Prometheus & Grafana & Alertmanager Stack
	if config.GlobalConfig.Observability.Prometheus {
		// Production-grade defaults: configure native NGINX Ingress with path-based routing under any host IP
		flags := []string{
			// Grafana Ingress & sub-path config
			"--set", "grafana.ingress.enabled=true",
			"--set", "grafana.ingress.hosts[0]=",
			"--set", "grafana.ingress.path=/grafana",
			"--set", "grafana.ingress.ingressClassName=nginx",
			"--set", "grafana.grafana.ini.server.root_url=%(protocol)s://%(domain)s/grafana/",
			"--set", "grafana.grafana.ini.server.serve_from_sub_path=true",
			
			// Prometheus Ingress & sub-path config
			"--set", "prometheus.ingress.enabled=true",
			"--set", "prometheus.ingress.hosts[0]=",
			"--set", "prometheus.ingress.paths[0]=/prometheus",
			"--set", "prometheus.ingress.ingressClassName=nginx",
			"--set", "prometheus.prometheusSpec.routePrefix=/prometheus",
			"--set", "prometheus.prometheusSpec.externalUrl=/prometheus",

			// Alertmanager Ingress & sub-path config
			"--set", "alertmanager.ingress.enabled=true",
			"--set", "alertmanager.ingress.hosts[0]=",
			"--set", "alertmanager.ingress.paths[0]=/alertmanager",
			"--set", "alertmanager.ingress.ingressClassName=nginx",
			"--set", "alertmanager.alertmanagerSpec.routePrefix=/alertmanager",
			"--set", "alertmanager.alertmanagerSpec.externalUrl=/alertmanager",
		}
		if err := helm.InstallRelease("stackpulse-prometheus", "prometheus-community/kube-prometheus-stack", ns, flags, dryRun); err != nil {
			return err
		}
	}

	// B. Loki Log Collector Stack
	if config.GlobalConfig.Observability.Loki {
		if err := helm.InstallRelease("stackpulse-loki", "grafana/loki-stack", ns, nil, dryRun); err != nil {
			return err
		}
	}

	// C. Tempo Distributed Tracing
	if config.GlobalConfig.Observability.Tempo {
		if err := helm.InstallRelease("stackpulse-tempo", "grafana/tempo", ns, nil, dryRun); err != nil {
			return err
		}
	}

	// D. OpenTelemetry Collector Pipeline
	if config.GlobalConfig.Observability.OpenTelemetry {
		flags := []string{
			"--set", "mode=deployment",
			"--set", "image.repository=ghcr.io/open-telemetry/opentelemetry-collector-releases/opentelemetry-collector-k8s",
		}
		if err := helm.InstallRelease("stackpulse-otel", "open-telemetry/opentelemetry-collector", ns, flags, dryRun); err != nil {
			return err
		}
	}

	// 4. Resolve Dynamic Ingress IP
	instanceIP := "127.0.0.1"
	if config.GlobalConfig.Observability.Prometheus {
		ingressIP, err := FetchIngressIP(ns, dryRun)
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
	}

	// 5. Output complete instructions and credential hooks
	fmt.Println()
	fmt.Printf("%s%sStackPulse Observability Stack deployed successfully!%s\n", utils.PrefixReady, utils.ColorBold, utils.ColorReset)
	fmt.Println("-----------------------------------------------------------------")
	fmt.Printf("🌐  Namespace: %s\n", ns)
	
	if config.GlobalConfig.Observability.Prometheus {
		fmt.Println("\n📊  Access Telemetry Dashboards via Ingress:")
		fmt.Printf("    🔗  Grafana Dashboard:     %s\n", utils.ColorBold+fmt.Sprintf("http://%s/grafana", instanceIP)+utils.ColorReset)
		fmt.Printf("    🔗  Prometheus Server:     %s\n", utils.ColorBold+fmt.Sprintf("http://%s/prometheus", instanceIP)+utils.ColorReset)
		fmt.Printf("    🔗  Alertmanager Panel:    %s\n", utils.ColorBold+fmt.Sprintf("http://%s/alertmanager", instanceIP)+utils.ColorReset)
		fmt.Println("\n    👤  Default credentials:   Username: admin")
		fmt.Printf("                               Password command: %s\n", utils.ColorCyan+fmt.Sprintf("kubectl get secret --namespace %s stackpulse-prometheus-grafana -o jsonpath=\"{.data.admin-password}\" | base64 --decode ; echo", ns)+utils.ColorReset)
	}

	fmt.Println("-----------------------------------------------------------------")
	return nil
}
