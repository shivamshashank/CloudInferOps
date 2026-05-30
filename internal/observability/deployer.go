package observability

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

	// NGINX Ingress Controller is always required for path-based routing
	if err := helm.AddRepo("ingress-nginx", "https://kubernetes.github.io/ingress-nginx", dryRun); err != nil {
		return err
	}
	reposAdded = true

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

	// A0. NGINX Ingress Controller (required for path-based routing to Grafana/Prometheus/Alertmanager)
	ingressFlags := []string{
		"--set", "controller.watchIngressWithoutClass=true",
	}
	if err := helm.InstallRelease("stackpulse-ingress-nginx", "ingress-nginx/ingress-nginx", ns, ingressFlags, dryRun); err != nil {
		return err
	}

	// Wait for ingress controller pod to become ready before deploying apps that create Ingress resources
	if !dryRun {
		fmt.Printf("%sWaiting for NGINX Ingress Controller to become ready...\n", utils.PrefixInfo)
		for i := 0; i < 30; i++ {
			_, _, waitErr := utils.ExecCommand("", "kubectl", "wait", "--namespace", ns,
				"--for=condition=Ready", "pod",
				"-l", "app.kubernetes.io/component=controller,app.kubernetes.io/instance=stackpulse-ingress-nginx",
				"--timeout=10s")
			if waitErr == nil {
				fmt.Printf("%sNGINX Ingress Controller is ready.\n", utils.PrefixOK)
				break
			}
			if i == 29 {
				fmt.Printf("%sIngress Controller not ready yet, continuing deployment anyway...\n", utils.PrefixWarn)
			}
		}
	}

	// A. Prometheus & Grafana & Alertmanager Stack
	if config.GlobalConfig.Observability.Prometheus {
		// Production-grade defaults: configure native NGINX Ingress with path-based routing under any host IP
		flags := []string{
			// StackPulse creates one explicit Ingress below; keep chart-generated Ingresses disabled.
			"--set", "grafana.ingress.enabled=false",
			"--set", "prometheus.ingress.enabled=false",
			"--set", "alertmanager.ingress.enabled=false",

			// Sub-path config for the apps behind /grafana, /prometheus, and /alertmanager.
			"--set", "grafana.grafana.ini.server.root_url=%(protocol)s://%(domain)s/grafana/",
			"--set", "grafana.grafana.ini.server.serve_from_sub_path=true",
			"--set", "prometheus.prometheusSpec.routePrefix=/prometheus",
			"--set", "prometheus.prometheusSpec.externalUrl=/prometheus",
			"--set", "alertmanager.alertmanagerSpec.routePrefix=/alertmanager",
			"--set", "alertmanager.alertmanagerSpec.externalUrl=/alertmanager",
		}
		if err := helm.InstallRelease("stackpulse-prometheus", "prometheus-community/kube-prometheus-stack", ns, flags, dryRun); err != nil {
			return err
		}
		if err := applyObservabilityIngress(ns, dryRun); err != nil {
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
			// Fallback: Resolve active interface IP of the host machine (e.g., VM / EC2 IP)
			instanceIP = utils.GetLocalIP()
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

func applyObservabilityIngress(ns string, dryRun bool) error {
	grafanaSvc := "stackpulse-prometheus-grafana"
	prometheusSvc := "stackpulse-prometheus-kube-prom-prometheus"
	alertmanagerSvc := "stackpulse-prometheus-kube-prom-alertmanager"

	if !dryRun {
		var err error
		grafanaSvc, err = findServiceByPort(ns, "grafana", 80, grafanaSvc)
		if err != nil {
			return err
		}
		prometheusSvc, err = findServiceByPort(ns, "prometheus", 9090, prometheusSvc)
		if err != nil {
			return err
		}
		alertmanagerSvc, err = findServiceByPort(ns, "alertmanager", 9093, alertmanagerSvc)
		if err != nil {
			return err
		}
	}

	manifest := fmt.Sprintf(`apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: stackpulse-observability
  namespace: %s
spec:
  ingressClassName: nginx
  rules:
    - http:
        paths:
          - path: /grafana
            pathType: Prefix
            backend:
              service:
                name: %s
                port:
                  number: 80
          - path: /prometheus
            pathType: Prefix
            backend:
              service:
                name: %s
                port:
                  number: 9090
          - path: /alertmanager
            pathType: Prefix
            backend:
              service:
                name: %s
                port:
                  number: 9093
`, ns, grafanaSvc, prometheusSvc, alertmanagerSvc)

	if dryRun {
		fmt.Printf("%s[DRY-RUN] kubectl apply -f stackpulse-observability-ingress.yaml\n", utils.PrefixInfo)
		return nil
	}

	fmt.Printf("%sApplying StackPulse observability Ingress routes...\n", utils.PrefixInfo)
	tmpPath := filepath.Join(os.TempDir(), "stackpulse-observability-ingress.yaml")
	if err := os.WriteFile(tmpPath, []byte(manifest), 0600); err != nil {
		return fmt.Errorf("failed to write temporary ingress manifest: %w", err)
	}
	defer os.Remove(tmpPath)

	_, stderr, err := utils.ExecCommand("", "kubectl", "apply", "-f", tmpPath)
	if err != nil {
		return fmt.Errorf("failed to apply observability ingress routes: %w (stderr: %s)", err, stderr)
	}
	fmt.Printf("%sStackPulse observability Ingress routes applied.\n", utils.PrefixOK)
	return nil
}

func findServiceByPort(ns, nameHint string, port int, fallback string) (string, error) {
	output, stderr, err := utils.ExecCommand("", "kubectl", "get", "svc", "-n", ns, "-o", "json")
	if err != nil {
		return "", fmt.Errorf("failed to list services in namespace %s: %w (stderr: %s)", ns, err, stderr)
	}

	var services struct {
		Items []struct {
			Metadata struct {
				Name string `json:"name"`
			} `json:"metadata"`
			Spec struct {
				Ports []struct {
					Port int `json:"port"`
				} `json:"ports"`
			} `json:"spec"`
		} `json:"items"`
	}
	if err := json.Unmarshal([]byte(output), &services); err != nil {
		return "", fmt.Errorf("failed to parse Kubernetes services: %w", err)
	}

	for _, svc := range services.Items {
		name := strings.ToLower(svc.Metadata.Name)
		if !strings.Contains(name, "stackpulse") || !strings.Contains(name, nameHint) {
			continue
		}
		for _, svcPort := range svc.Spec.Ports {
			if svcPort.Port == port {
				return svc.Metadata.Name, nil
			}
		}
	}

	for _, svc := range services.Items {
		name := strings.ToLower(svc.Metadata.Name)
		if !strings.Contains(name, nameHint) {
			continue
		}
		for _, svcPort := range svc.Spec.Ports {
			if svcPort.Port == port {
				return svc.Metadata.Name, nil
			}
		}
	}

	fmt.Printf("%sCould not discover %s service on port %d. Falling back to %s.\n", utils.PrefixWarn, nameHint, port, fallback)
	return fallback, nil
}
