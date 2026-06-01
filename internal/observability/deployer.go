package observability

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

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

	if config.GlobalConfig.Observability.Grafana || config.GlobalConfig.Observability.Loki || config.GlobalConfig.Observability.Tempo || config.GlobalConfig.Observability.Pyroscope {
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

	if config.GlobalConfig.Observability.ArgoCD {
		if err := helm.AddRepo("argo", "https://argoproj.github.io/argo-helm", dryRun); err != nil {
			return err
		}
		reposAdded = true
	}

	if config.GlobalConfig.Observability.Thanos {
		if err := helm.AddRepo("bitnami", "https://charts.bitnami.com/bitnami", dryRun); err != nil {
			return err
		}
		reposAdded = true
	}

	if config.GlobalConfig.Observability.VictoriaMetrics {
		if err := helm.AddRepo("vm", "https://victoriametrics.github.io/helm-charts", dryRun); err != nil {
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

	// E. ArgoCD Continuous Delivery (Bootstrapped first for GitOps management)
	if config.GlobalConfig.Observability.ArgoCD {
		argoFlags := []string{
			"--set", "server.extraArgs={--insecure,--rootpath=/argocd}",
			"--set", "server.ingress.enabled=false",
		}
		if err := helm.InstallRelease("stackpulse-argocd", "argo/argo-cd", ns, argoFlags, dryRun); err != nil {
			return err
		}
		// Wait for ArgoCD Application CRDs to initialize before we apply applications
		if !dryRun {
			fmt.Printf("%sWaiting for ArgoCD CRDs to initialize...\n", utils.PrefixInfo)
			for i := 0; i < 30; i++ {
				if _, _, err := utils.ExecCommand("", "kubectl", "get", "crd", "applications.argoproj.io"); err == nil {
					time.Sleep(3 * time.Second) // Let controller settle
					break
				}
				time.Sleep(2 * time.Second)
			}
		}
	}

	// A. Prometheus & Grafana & Alertmanager Stack
	if config.GlobalConfig.Observability.Prometheus {
		// Production-grade dynamic values construction to bypass cli string escaping constraints
		var valuesYAML strings.Builder
		valuesYAML.WriteString("grafana:\n")
		valuesYAML.WriteString("  ingress:\n")
		valuesYAML.WriteString("    enabled: false\n")
		valuesYAML.WriteString("  grafana.ini:\n")
		valuesYAML.WriteString("    server:\n")
		valuesYAML.WriteString("      root_url: \"%(protocol)s://%(domain)s/grafana/\"\n")
		valuesYAML.WriteString("      serve_from_sub_path: true\n")
		valuesYAML.WriteString("  sidecar:\n")
		valuesYAML.WriteString("    dashboards:\n")
		valuesYAML.WriteString("      enabled: true\n")
		valuesYAML.WriteString("      label: grafana_dashboard\n")
		valuesYAML.WriteString("      searchNamespace: ALL\n")
		valuesYAML.WriteString("    datasources:\n")
		valuesYAML.WriteString("      enabled: true\n")
		valuesYAML.WriteString("      defaultDatasourceEnabled: true\n")
		valuesYAML.WriteString("      isDefaultDatasource: true\n")
		valuesYAML.WriteString("      name: Prometheus\n")
		valuesYAML.WriteString("      uid: prometheus\n")
		fmt.Fprintf(&valuesYAML, "      url: http://stackpulse-prometheus-kube-prometheus.%s:9090/prometheus\n", ns)
		valuesYAML.WriteString("      alertmanager:\n")
		valuesYAML.WriteString("        enabled: true\n")
		valuesYAML.WriteString("        name: Alertmanager\n")
		valuesYAML.WriteString("        uid: alertmanager\n")
		fmt.Fprintf(&valuesYAML, "        url: http://stackpulse-prometheus-kube-alertmanager.%s:9093/alertmanager\n", ns)

		// Build additional Grafana data sources dynamically
		valuesYAML.WriteString("  additionalDataSources:\n")
		if config.GlobalConfig.Observability.Loki {
			valuesYAML.WriteString("    - name: Loki\n")
			valuesYAML.WriteString("      type: loki\n")
			valuesYAML.WriteString("      access: proxy\n")
			fmt.Fprintf(&valuesYAML, "      url: http://stackpulse-loki.%s:3100\n", ns)
			valuesYAML.WriteString("      uid: loki\n")
		}
		if config.GlobalConfig.Observability.Tempo {
			valuesYAML.WriteString("    - name: Tempo\n")
			valuesYAML.WriteString("      type: tempo\n")
			valuesYAML.WriteString("      access: proxy\n")
			fmt.Fprintf(&valuesYAML, "      url: http://stackpulse-tempo.%s:3100\n", ns)
			valuesYAML.WriteString("      uid: tempo\n")
		}
		if config.GlobalConfig.Observability.Pyroscope {
			valuesYAML.WriteString("    - name: Pyroscope\n")
			valuesYAML.WriteString("      type: pyroscope\n")
			valuesYAML.WriteString("      access: proxy\n")
			fmt.Fprintf(&valuesYAML, "      url: http://stackpulse-pyroscope.%s:4040\n", ns)
			valuesYAML.WriteString("      uid: pyroscope\n")
		}

		valuesYAML.WriteString("kubeControllerManager:\n")
		valuesYAML.WriteString("  enabled: false\n")
		valuesYAML.WriteString("kubeEtcd:\n")
		valuesYAML.WriteString("  enabled: false\n")
		valuesYAML.WriteString("kubeScheduler:\n")
		valuesYAML.WriteString("  enabled: false\n")
		valuesYAML.WriteString("kubeProxy:\n")
		valuesYAML.WriteString("  enabled: false\n")

		valuesYAML.WriteString("prometheus:\n")
		valuesYAML.WriteString("  ingress:\n")
		valuesYAML.WriteString("    enabled: false\n")
		valuesYAML.WriteString("  prometheusSpec:\n")
		valuesYAML.WriteString("    externalLabels:\n")
		valuesYAML.WriteString("      cluster: default\n")
		valuesYAML.WriteString("    routePrefix: /prometheus\n")
		valuesYAML.WriteString("    externalUrl: http://localhost/prometheus\n")

		if config.GlobalConfig.Observability.Thanos {
			valuesYAML.WriteString("    thanos:\n")
			valuesYAML.WriteString("      version: v0.31.0\n")
		}

		// Configure Blackbox exporter scraping targets under Prometheus
		if config.GlobalConfig.Observability.BlackboxExporter && len(config.GlobalConfig.Observability.BlackboxTargets) > 0 {
			valuesYAML.WriteString("    additionalScrapeConfigs:\n")
			valuesYAML.WriteString("      - job_name: 'blackbox'\n")
			valuesYAML.WriteString("        metrics_path: /probe\n")
			valuesYAML.WriteString("        params:\n")
			valuesYAML.WriteString("          module: [http_2xx]\n")
			valuesYAML.WriteString("        static_configs:\n")
			valuesYAML.WriteString("          - targets:\n")
			for _, target := range config.GlobalConfig.Observability.BlackboxTargets {
				fmt.Fprintf(&valuesYAML, "            - '%s'\n", target)
			}
			valuesYAML.WriteString("        relabel_configs:\n")
			valuesYAML.WriteString("          - source_labels: [__address__]\n")
			valuesYAML.WriteString("            target_label: __param_target\n")
			valuesYAML.WriteString("          - source_labels: [__param_target]\n")
			valuesYAML.WriteString("            target_label: instance\n")
			valuesYAML.WriteString("          - target_label: __address__\n")
			fmt.Fprintf(&valuesYAML, "            replacement: stackpulse-blackbox-prometheus-blackbox-exporter.%s:9115\n", ns)
		}

		valuesYAML.WriteString("kube-state-metrics:\n")
		if config.GlobalConfig.Observability.KubeStateMetrics {
			valuesYAML.WriteString("  enabled: true\n")
		} else {
			valuesYAML.WriteString("  enabled: false\n")
		}

		valuesYAML.WriteString("alertmanager:\n")
		valuesYAML.WriteString("  ingress:\n")
		valuesYAML.WriteString("    enabled: false\n")
		valuesYAML.WriteString("  alertmanagerSpec:\n")
		valuesYAML.WriteString("    routePrefix: /alertmanager\n")
		valuesYAML.WriteString("    externalUrl: http://localhost/alertmanager\n")

		if config.GlobalConfig.Observability.ArgoCD {
			if err := deployViaArgoCD("stackpulse-prometheus", "https://prometheus-community.github.io/helm-charts", "kube-prometheus-stack", ns, nil, valuesYAML.String(), dryRun); err != nil {
				return err
			}
		} else {
			tmpValuesPath := filepath.Join(os.TempDir(), "stackpulse-prometheus-values.yaml")
			if err := os.WriteFile(tmpValuesPath, []byte(valuesYAML.String()), 0600); err != nil {
				return fmt.Errorf("failed to write temporary prometheus values file: %w", err)
			}
			defer func() { _ = os.Remove(tmpValuesPath) }()

			flags := []string{"-f", tmpValuesPath}
			if err := helm.InstallRelease("stackpulse-prometheus", "prometheus-community/kube-prometheus-stack", ns, flags, dryRun); err != nil {
				return err
			}
		}
		if err := applyObservabilityIngress(ns, dryRun); err != nil {
			return err
		}

		// Trigger Auto-Provisioning of Dashboards & SRE Alert Packs
		if err := ProvisionDashboards(ns, dryRun); err != nil {
			fmt.Printf("%sWarning: failed to provision dashboards: %v\n", utils.PrefixWarn, err)
		}
		if err := ProvisionAlertRules(ns, dryRun); err != nil {
			fmt.Printf("%sWarning: failed to provision alert rules: %v\n", utils.PrefixWarn, err)
		}
	}

	// B. Loki Log Collector Stack
	if config.GlobalConfig.Observability.Loki {
		flags := []string{
			"--set", "loki.isDefault=false",
		}
		if config.GlobalConfig.Observability.ArgoCD {
			if err := deployViaArgoCD("stackpulse-loki", "https://grafana.github.io/helm-charts", "loki-stack", ns, flags, "", dryRun); err != nil {
				return err
			}
		} else {
			if err := helm.InstallRelease("stackpulse-loki", "grafana/loki-stack", ns, flags, dryRun); err != nil {
				return err
			}
		}
	}

	// C. Tempo Distributed Tracing
	if config.GlobalConfig.Observability.Tempo {
		if config.GlobalConfig.Observability.ArgoCD {
			if err := deployViaArgoCD("stackpulse-tempo", "https://grafana.github.io/helm-charts", "tempo", ns, nil, "", dryRun); err != nil {
				return err
			}
		} else {
			if err := helm.InstallRelease("stackpulse-tempo", "grafana/tempo", ns, nil, dryRun); err != nil {
				return err
			}
		}
	}

	// D. OpenTelemetry Collector Pipeline
	if config.GlobalConfig.Observability.OpenTelemetry {
		flags := []string{
			"--set", "mode=deployment",
			"--set", "image.repository=ghcr.io/open-telemetry/opentelemetry-collector-releases/opentelemetry-collector-k8s",
		}
		if config.GlobalConfig.Observability.ArgoCD {
			if err := deployViaArgoCD("stackpulse-otel", "https://open-telemetry.github.io/opentelemetry-helm-charts", "opentelemetry-collector", ns, flags, "", dryRun); err != nil {
				return err
			}
		} else {
			if err := helm.InstallRelease("stackpulse-otel", "open-telemetry/opentelemetry-collector", ns, flags, dryRun); err != nil {
				return err
			}
		}
	}

	// A1. Blackbox Exporter Deployment
	if config.GlobalConfig.Observability.BlackboxExporter {
		if config.GlobalConfig.Observability.ArgoCD {
			if err := deployViaArgoCD("stackpulse-blackbox", "https://prometheus-community.github.io/helm-charts", "prometheus-blackbox-exporter", ns, nil, "", dryRun); err != nil {
				return err
			}
		} else {
			if err := helm.InstallRelease("stackpulse-blackbox", "prometheus-community/prometheus-blackbox-exporter", ns, nil, dryRun); err != nil {
				return err
			}
		}
	}

	// A2. Pyroscope Profiler Deployment
	if config.GlobalConfig.Observability.Pyroscope {
		if config.GlobalConfig.Observability.ArgoCD {
			if err := deployViaArgoCD("stackpulse-pyroscope", "https://grafana.github.io/helm-charts", "pyroscope", ns, nil, "", dryRun); err != nil {
				return err
			}
		} else {
			if err := helm.InstallRelease("stackpulse-pyroscope", "grafana/pyroscope", ns, nil, dryRun); err != nil {
				return err
			}
		}
	}

	// A3. Thanos HA optional Deployment
	if config.GlobalConfig.Observability.Thanos {
		if err := createThanosSecret(ns, dryRun); err != nil {
			return fmt.Errorf("failed to create Thanos objstore secret: %w", err)
		}
		thanosFlags := []string{
			"--set", "existingObjstoreSecret=stackpulse-thanos-objstore",
		}
		if config.GlobalConfig.Observability.ArgoCD {
			if err := deployViaArgoCD("stackpulse-thanos", "https://charts.bitnami.com/bitnami", "thanos", ns, thanosFlags, "", dryRun); err != nil {
				return err
			}
		} else {
			if err := helm.InstallRelease("stackpulse-thanos", "bitnami/thanos", ns, thanosFlags, dryRun); err != nil {
				return err
			}
		}
	}

	// A4. VictoriaMetrics optional Deployment
	if config.GlobalConfig.Observability.VictoriaMetrics {
		vmFlags := []string{
			"--set", "vmsingle.enabled=true",
			"--set", "vmcluster.enabled=false",
		}
		if config.GlobalConfig.Observability.ArgoCD {
			if err := deployViaArgoCD("stackpulse-vm", "https://victoriametrics.github.io/helm-charts", "victoria-metrics-k8s-stack", ns, vmFlags, "", dryRun); err != nil {
				return err
			}
		} else {
			if err := helm.InstallRelease("stackpulse-vm", "vm/victoria-metrics-k8s-stack", ns, vmFlags, dryRun); err != nil {
				return err
			}
		}
	}

	// 4. Resolve Dynamic Ingress IP
	instanceIP := "127.0.0.1"
	var detectedPublicIP string
	if config.GlobalConfig.Observability.Prometheus {
		ingressIP, err := FetchIngressIP(ns, dryRun)
		if err == nil && ingressIP != "" {
			instanceIP = ingressIP
		} else {
			// Fallback: Resolve active interface IP of the host machine (e.g., VM / EC2 IP)
			instanceIP = utils.GetLocalIP()
		}

		// If we are on a cloud VM, the ingress IP might be the private subnet IP.
		// Attempt to resolve the public IP for correct external browser access.
		if parsedIP := net.ParseIP(instanceIP); parsedIP != nil && parsedIP.IsPrivate() {
			if utils.IsCloudVM() {
				detectedPublicIP = utils.GetPublicIP()
				if detectedPublicIP != "" {
					instanceIP = detectedPublicIP
				}
			}
		}
	}

	argoSecretName := "argocd-initial-admin-secret"
	if out, _, err := utils.ExecCommand("", "kubectl", "get", "secret", "-n", ns, "-o", "name"); err == nil {
		for _, line := range strings.Split(out, "\n") {
			if strings.Contains(line, "initial-admin-secret") {
				argoSecretName = strings.TrimPrefix(strings.TrimSpace(line), "secret/")
				break
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
		fmt.Printf("    🔗  Grafana Dashboard:     %s\n", utils.ColorBold+fmt.Sprintf("http://%s/grafana/", instanceIP)+utils.ColorReset)
		fmt.Printf("    🔗  Prometheus Server:     %s\n", utils.ColorBold+fmt.Sprintf("http://%s/prometheus/", instanceIP)+utils.ColorReset)
		fmt.Printf("    🔗  Alertmanager Panel:    %s\n", utils.ColorBold+fmt.Sprintf("http://%s/alertmanager/", instanceIP)+utils.ColorReset)
		if config.GlobalConfig.Observability.ArgoCD {
			fmt.Printf("    🔗  ArgoCD Dashboard:      %s\n", utils.ColorBold+fmt.Sprintf("http://%s/argocd", instanceIP)+utils.ColorReset)
		}

		fmt.Println("\n    👤  Default credentials:   Username: admin")
		fmt.Printf("                               Password command: %s\n", utils.ColorCyan+fmt.Sprintf("kubectl get secret --namespace %s stackpulse-prometheus-grafana -o jsonpath=\"{.data.admin-password}\" | base64 --decode ; echo", ns)+utils.ColorReset)
		if config.GlobalConfig.Observability.ArgoCD {
			fmt.Printf("                               ArgoCD Password:  %s\n", utils.ColorCyan+fmt.Sprintf("kubectl get secret --namespace %s %s -o jsonpath=\"{.data.password}\" | base64 --decode ; echo", ns, argoSecretName)+utils.ColorReset)
		}
	}

	fmt.Println("-----------------------------------------------------------------")
	return nil
}

func createThanosSecret(ns string, dryRun bool) error {
	secretYAML := fmt.Sprintf(`apiVersion: v1
kind: Secret
metadata:
  name: stackpulse-thanos-objstore
  namespace: %s
type: Opaque
stringData:
  thanos.yaml: |
    type: FILESYSTEM
    config:
      directory: /var/thanos
`, ns)

	if dryRun {
		fmt.Printf("%s[DRY-RUN] Would create Thanos storage config secret\n", utils.PrefixInfo)
		return nil
	}

	tmpPath := filepath.Join(os.TempDir(), "stackpulse-thanos-objstore.yaml")
	if err := os.WriteFile(tmpPath, []byte(secretYAML), 0600); err != nil {
		return err
	}
	defer func() { _ = os.Remove(tmpPath) }()

	_, _, err := utils.ExecCommand("", "kubectl", "apply", "-f", tmpPath)
	return err
}

func applyObservabilityIngress(ns string, dryRun bool) error {
	grafanaSvc := "stackpulse-prometheus-grafana"
	prometheusSvc := "stackpulse-prometheus-kube-prometheus"
	alertmanagerSvc := "stackpulse-prometheus-kube-alertmanager"
	argoSvc := "stackpulse-argocd-server"

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
		if config.GlobalConfig.Observability.ArgoCD {
			argoSvc, err = findServiceByPort(ns, "argocd-server", 80, argoSvc)
			if err != nil {
				return err
			}
		}
	}

	argoRoute := ""
	if config.GlobalConfig.Observability.ArgoCD {
		argoRoute = fmt.Sprintf(`
          - path: /argocd(/|$)(.*)
            pathType: ImplementationSpecific
            backend:
              service:
                name: %s
                port:
                  number: 80`, argoSvc)
	}

	manifest := fmt.Sprintf(`apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: stackpulse-observability
  namespace: %s
  annotations:
    kubernetes.io/ingress.class: nginx
    nginx.ingress.kubernetes.io/ssl-redirect: "false"
    nginx.ingress.kubernetes.io/use-regex: "true"
spec:
  ingressClassName: nginx
  rules:
    - http:
        paths:
          - path: /grafana(/|$)(.*)
            pathType: ImplementationSpecific
            backend:
              service:
                name: %s
                port:
                  number: 80
          - path: /prometheus(/|$)(.*)
            pathType: ImplementationSpecific
            backend:
              service:
                name: %s
                port:
                  number: 9090
          - path: /alertmanager(/|$)(.*)
            pathType: ImplementationSpecific
            backend:
              service:
                name: %s
                port:
                  number: 9093%s
`, ns, grafanaSvc, prometheusSvc, alertmanagerSvc, argoRoute)

	if dryRun {
		fmt.Printf("%s[DRY-RUN] kubectl apply -f stackpulse-observability-ingress.yaml\n", utils.PrefixInfo)
		return nil
	}

	fmt.Printf("%sApplying StackPulse observability Ingress routes...\n", utils.PrefixInfo)
	tmpPath := filepath.Join(os.TempDir(), "stackpulse-observability-ingress.yaml")
	if err := os.WriteFile(tmpPath, []byte(manifest), 0600); err != nil {
		return fmt.Errorf("failed to write temporary ingress manifest: %w", err)
	}
	defer func() { _ = os.Remove(tmpPath) }()

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

func deployViaArgoCD(name, repoURL, chart, ns string, flags []string, valuesStr string, dryRun bool) error {
	helmParameters := ""
	for i := 0; i < len(flags); i++ {
		if flags[i] == "--set" && i+1 < len(flags) {
			kv := strings.SplitN(flags[i+1], "=", 2)
			if len(kv) == 2 {
				name := strings.ReplaceAll(kv[0], "'", "''")
				value := strings.ReplaceAll(kv[1], "'", "''")
				helmParameters += fmt.Sprintf("\n        - name: '%s'\n          value: '%s'", name, value)
			}
			i++
		} else if flags[i] == "--set-string" && i+1 < len(flags) {
			kv := strings.SplitN(flags[i+1], "=", 2)
			if len(kv) == 2 {
				name := strings.ReplaceAll(kv[0], "'", "''")
				value := strings.ReplaceAll(kv[1], "'", "''")
				helmParameters += fmt.Sprintf("\n        - name: '%s'\n          value: '%s'\n          forceString: true", name, value)
			}
			i++
		}
	}

	helmBlock := ""
	if helmParameters != "" {
		helmBlock = fmt.Sprintf("\n    helm:\n      parameters:%s", helmParameters)
	} else if valuesStr != "" {
		helmBlock = fmt.Sprintf("\n    helm:\n      values: |\n%s", indentString(valuesStr, 8))
	}

	manifest := fmt.Sprintf(`apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: %s
  namespace: %s
spec:
  project: default
  source:
    repoURL: '%s'
    chart: '%s'
    targetRevision: '*'%s
  destination:
    server: https://kubernetes.default.svc
    namespace: %s
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - ServerSideApply=true
    retry:
      limit: 5
      backoff:
        duration: 5s
        factor: 2
        maxDuration: 3m
`, name, ns, repoURL, chart, helmBlock, ns)

	if dryRun {
		fmt.Printf("%s[DRY-RUN] Deploying %s via ArgoCD Application\n", utils.PrefixInfo, name)
		return nil
	}

	tmpPath := filepath.Join(os.TempDir(), name+"-argocd-app.yaml")
	if err := os.WriteFile(tmpPath, []byte(manifest), 0600); err != nil {
		return err
	}
	defer func() { _ = os.Remove(tmpPath) }()

	_, stderr, err := utils.ExecCommand("", "kubectl", "apply", "-f", tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create ArgoCD Application for %s: %w (stderr: %s)", name, err, stderr)
	}
	fmt.Printf("%sGitOps Application %s created successfully.\n", utils.PrefixOK, name)
	return nil
}
