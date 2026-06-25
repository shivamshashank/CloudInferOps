package gitops

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/shivamshashank/CloudInferOps/internal/config"
	"github.com/shivamshashank/CloudInferOps/internal/helm"
	"github.com/shivamshashank/CloudInferOps/internal/observability"
	"github.com/shivamshashank/CloudInferOps/internal/utils"
)

// BootstrapGitOps orchestrates the full Model B GitOps deployment:
// 1. Installs Ingress NGINX & ArgoCD via Helm
// 2. Provisions the in-cluster Git server
// 3. Generates the GitOps repository layout under ~/.cloudinfer/gitops-repo
// 4. Initializes, commits, and pushes the local repo to the Git server
// 5. Registers ArgoCD Applications pointing to the Git server
// 6. Configures the ingress controller routing
func BootstrapGitOps(dryRun bool) error {
	ns := config.GlobalConfig.Namespace
	if ns == "" {
		ns = "observability"
	}

	fmt.Printf("%sStarting CloudInferOps GitOps Bootstrap (Model B)...\n", utils.PrefixInfo)

	if !CheckGitInstalled() {
		return fmt.Errorf("git CLI is required on the host system but was not found. Please install git and try again")
	}

	// 1. Create namespace
	if !dryRun {
		fmt.Printf("%sCreating Kubernetes namespace '%s' if not exists...\n", utils.PrefixInfo, ns)
		_, _, _ = utils.ExecCommand("", "kubectl", "create", "namespace", ns)
	}

	// 2. Add Helm repos
	if err := helm.AddRepo("ingress-nginx", "https://kubernetes.github.io/ingress-nginx", dryRun); err != nil {
		return err
	}
	if err := helm.AddRepo("argo", "https://argoproj.github.io/argo-helm", dryRun); err != nil {
		return err
	}
	if err := helm.UpdateRepos(dryRun); err != nil {
		return err
	}

	// 3. Deploy NGINX Ingress Controller
	ingressFlags := []string{
		"--set", "controller.watchIngressWithoutClass=true",
	}
	if err := helm.InstallRelease("cloudinfer-ingress-nginx", "ingress-nginx/ingress-nginx", ns, ingressFlags, dryRun); err != nil {
		return err
	}

	// Wait for ingress controller
	if !dryRun {
		stopSpinner := utils.StartSpinner("Waiting for NGINX Ingress Controller to become ready...")
		ready := false
		for i := 0; i < 30; i++ {
			_, _, waitErr := utils.ExecCommand("", "kubectl", "wait", "--namespace", ns,
				"--for=condition=Ready", "pod",
				"-l", "app.kubernetes.io/component=controller,app.kubernetes.io/instance=cloudinfer-ingress-nginx",
				"--timeout=10s")
			if waitErr == nil {
				ready = true
				break
			}
			time.Sleep(1 * time.Second)
		}
		stopSpinner()
		if ready {
			fmt.Printf("%sNGINX Ingress Controller is ready.\n", utils.PrefixOK)
		}
	}

	// 4. Deploy ArgoCD
	argoFlags := []string{
		"--set", "server.extraArgs={--insecure,--rootpath=/argocd}",
		"--set", "server.ingress.enabled=false",
	}
	if err := helm.InstallRelease("cloudinfer-argocd", "argo/argo-cd", ns, argoFlags, dryRun); err != nil {
		return err
	}

	// Wait for ArgoCD CRDs to initialize
	if !dryRun {
		stopSpinner := utils.StartSpinner("Waiting for ArgoCD CRDs to initialize...")
		for i := 0; i < 30; i++ {
			if _, _, err := utils.ExecCommand("", "kubectl", "get", "crd", "applications.argoproj.io"); err == nil {
				time.Sleep(3 * time.Second) // Let controller settle
				break
			}
			time.Sleep(2 * time.Second)
		}
		stopSpinner()

		stopSpinner2 := utils.StartSpinner("Waiting for ArgoCD Repo Server to become ready...")
		repoReady := false
		for i := 0; i < 30; i++ {
			_, _, waitErr := utils.ExecCommand("", "kubectl", "wait", "--namespace", ns,
				"--for=condition=Ready", "pod",
				"-l", "app.kubernetes.io/name=argocd-repo-server",
				"--timeout=10s")
			if waitErr == nil {
				repoReady = true
				break
			}
		}
		stopSpinner2()
		if repoReady {
			fmt.Printf("%sArgoCD Repo Server is ready.\n", utils.PrefixOK)
		}
	}

	// 5. Deploy Git Server
	if err := DeployGitServer(ns, dryRun); err != nil {
		return err
	}

	// 6. Generate local GitOps repository structure
	home, err := utils.GetRealHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}
	repoDir := filepath.Join(home, ".cloudinfer", "gitops-repo")

	if !dryRun {
		if err := generateGitOpsRepo(repoDir); err != nil {
			return fmt.Errorf("failed to generate local GitOps structure: %w", err)
		}
	}

	// 7. Push repository to in-cluster Git server
	if !dryRun {
		fmt.Printf("%sSetting up port-forward to Git Server...\n", utils.PrefixInfo)
		cancelPF, err := StartPortForward(ns, 9418, 9418)
		if err != nil {
			return fmt.Errorf("failed to port-forward to Git Server: %w", err)
		}
		defer cancelPF()

		fmt.Printf("%sPushing GitOps files to Git Server...\n", utils.PrefixInfo)
		if err := InitLocalRepo(repoDir); err != nil {
			return err
		}

		if err := PushToGitServer(repoDir, "git://127.0.0.1:9418/gitops.git"); err != nil {
			return err
		}
		fmt.Printf("%sSuccessfully pushed GitOps repository to Git Server.\n", utils.PrefixOK)
	}

	// 8. Deploy ArgoCD Applications
	if err := deployArgoCDApplications(ns, dryRun); err != nil {
		return err
	}

	// 9. Configure Ingress routes
	if err := applyGitOpsIngress(ns, dryRun); err != nil {
		return err
	}

	// 10. Wait for everything to become healthy
	waitForArgoCDApps(ns, dryRun)

	// Output success instructions
	instanceIP := "127.0.0.1"
	if !dryRun {
		ingressIP, err := fetchIngressIP(ns)
		if err == nil && ingressIP != "" {
			instanceIP = ingressIP
		} else {
			instanceIP = utils.GetLocalIP()
		}

		if parsedIP := net.ParseIP(instanceIP); parsedIP != nil && parsedIP.IsPrivate() {
			detectedPublicIP := utils.GetPublicIP()
			if detectedPublicIP != "" {
				instanceIP = detectedPublicIP
			}
		}
	}

	fmt.Println()
	fmt.Printf("%s%sCloudInferOps GitOps Bootstrap (Model B) Completed!%s\n", utils.PrefixReady, utils.ColorBold, utils.ColorReset)
	fmt.Println("-----------------------------------------------------------------")
	fmt.Printf("🌐  Namespace:           %s\n", ns)
	fmt.Printf("📦  GitOps Mode:         %s\n", utils.ColorCyan+"Local GitOps (ArgoCD managed)"+utils.ColorReset)
	fmt.Printf("📊  ArgoCD Dashboard:    %s\n", utils.ColorBold+fmt.Sprintf("http://%s/argocd", instanceIP)+utils.ColorReset)
	fmt.Printf("📊  Grafana Dashboard:   %s\n", utils.ColorBold+fmt.Sprintf("http://%s/grafana/", instanceIP)+utils.ColorReset)
	fmt.Printf("📊  Prometheus Server:   %s\n", utils.ColorBold+fmt.Sprintf("http://%s/prometheus/", instanceIP)+utils.ColorReset)
	fmt.Printf("📊  Alertmanager Panel:  %s\n", utils.ColorBold+fmt.Sprintf("http://%s/alertmanager/", instanceIP)+utils.ColorReset)
	fmt.Println("\n    👤  Default credentials:   Username: admin")
	fmt.Printf("                               Password: Use 'sudo cloudinfer status' to decrypt\n")
	fmt.Println("\n    ⏳  Note: It may take 5-6 minutes for all services and pods to fully start.")
	fmt.Printf("           Run %s to monitor the live progress.\n", utils.ColorCyan+"sudo cloudinfer status"+utils.ColorReset)
	fmt.Println("-----------------------------------------------------------------")

	return nil
}

func generateGitOpsRepo(repoDir string) error {
	_ = os.RemoveAll(repoDir)
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		return err
	}

	// Create aligned directory layout: apps/, infra/, monitoring/
	dirs := []string{
		"infra",
		"monitoring",
		"apps",
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(repoDir, dir), 0755); err != nil {
			return err
		}
	}

	// 1. Infra Component Chart
	infraChart := `apiVersion: v2
name: cloudinfer-infra
version: 1.0.0
dependencies:
  - name: ingress-nginx
    version: "4.10.0"
    repository: https://kubernetes.github.io/ingress-nginx
`
	if config.GlobalConfig.Observability.Thanos {
		infraChart += `  - name: thanos
    version: "12.5.1"
    repository: https://charts.bitnami.com/bitnami
`
	}

	infraValues := `ingress-nginx:
  controller:
    watchIngressWithoutClass: true
`
	if config.GlobalConfig.Observability.Thanos {
		infraValues += `thanos:
  existingObjstoreSecret: cloudinfer-thanos-objstore
`
	}

	if err := os.WriteFile(filepath.Join(repoDir, "infra", "Chart.yaml"), []byte(infraChart), 0644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(repoDir, "infra", "values.yaml"), []byte(infraValues), 0644); err != nil {
		return err
	}

	// 2. Monitoring Component Chart (umbrella SRE stack)
	ns := config.GlobalConfig.Namespace
	if ns == "" {
		ns = "observability"
	}

	monitoringChart := `apiVersion: v2
name: cloudinfer-monitoring
version: 1.0.0
dependencies:
`
	if config.GlobalConfig.Observability.Prometheus {
		monitoringChart += `  - name: kube-prometheus-stack
    version: "61.3.0"
    repository: https://prometheus-community.github.io/helm-charts
`
	}
	if config.GlobalConfig.Observability.Loki {
		monitoringChart += `  - name: loki-stack
    version: "2.10.2"
    repository: https://grafana.github.io/helm-charts
`
	}
	if config.GlobalConfig.Observability.Tempo {
		monitoringChart += `  - name: tempo
    version: "1.10.1"
    repository: https://grafana.github.io/helm-charts
`
	}
	if config.GlobalConfig.Observability.OpenTelemetry {
		monitoringChart += `  - name: opentelemetry-collector
    version: "0.93.0"
    repository: https://open-telemetry.github.io/opentelemetry-helm-charts
`
	}
	if config.GlobalConfig.Observability.Pyroscope {
		monitoringChart += `  - name: pyroscope
    version: "1.5.0"
    repository: https://grafana.github.io/helm-charts
`
	}
	if config.GlobalConfig.Observability.VictoriaMetrics {
		monitoringChart += `  - name: victoria-metrics-k8s-stack
    version: "0.24.0"
    repository: https://victoriametrics.github.io/helm-charts
`
	}

	// Build values YAML for monitoring dependencies
	var monValues strings.Builder
	if config.GlobalConfig.Observability.Prometheus {
		monValues.WriteString(`kube-prometheus-stack:
  kubeControllerManager:
    enabled: false
  kubeEtcd:
    enabled: false
  kubeScheduler:
    enabled: false
  kubeProxy:
    enabled: false
  grafana:
    ingress:
      enabled: false
    grafana.ini:
      server:
        root_url: "%(protocol)s://%(domain)s/grafana/"
        serve_from_sub_path: true
    sidecar:
      dashboards:
        enabled: true
        label: grafana_dashboard
        searchNamespace: ALL
      datasources:
        enabled: true
        defaultDatasourceEnabled: true
        isDefaultDatasource: true
        name: Prometheus
        uid: prometheus
        url: http://cloudinfer-prometheus-kube-prometheus.observability:9090/prometheus
        alertmanager:
          enabled: true
          name: Alertmanager
          uid: alertmanager
          url: http://cloudinfer-prometheus-kube-alertmanager.observability:9093/alertmanager
  prometheus:
    ingress:
      enabled: false
    prometheusSpec:
      externalLabels:
        cluster: default
      routePrefix: /prometheus
      externalUrl: http://localhost/prometheus
  alertmanager:
    ingress:
      enabled: false
    alertmanagerSpec:
      routePrefix: /alertmanager
      externalUrl: http://localhost/alertmanager
  kube-state-metrics:
`)
		if config.GlobalConfig.Observability.KubeStateMetrics {
			monValues.WriteString("    enabled: true\n")
		} else {
			monValues.WriteString("    enabled: false\n")
		}
	}

	if config.GlobalConfig.Observability.Loki {
		monValues.WriteString(`loki-stack:
  loki:
    isDefault: false
`)
	}
	if config.GlobalConfig.Observability.Tempo {
		monValues.WriteString(`tempo: {}
`)
	}
	if config.GlobalConfig.Observability.OpenTelemetry {
		monValues.WriteString(`opentelemetry-collector:
  mode: deployment
  image:
    repository: ghcr.io/open-telemetry/opentelemetry-collector-releases/opentelemetry-collector-k8s
`)
	}
	if config.GlobalConfig.Observability.Pyroscope {
		monValues.WriteString(`pyroscope: {}
`)
	}
	if config.GlobalConfig.Observability.VictoriaMetrics {
		monValues.WriteString(`victoria-metrics-k8s-stack:
  vmsingle:
    enabled: true
  vmcluster:
    enabled: false
`)
	}

	if err := os.WriteFile(filepath.Join(repoDir, "monitoring", "Chart.yaml"), []byte(monitoringChart), 0644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(repoDir, "monitoring", "values.yaml"), []byte(monValues.String()), 0644); err != nil {
		return err
	}

	// 2.5 Provision alert rules pack directly inside GitOps monitoring templates!
	if config.GlobalConfig.Observability.Prometheus {
		templatesDir := filepath.Join(repoDir, "monitoring", "templates")
		if err := os.MkdirAll(templatesDir, 0755); err != nil {
			return err
		}

		if err := os.WriteFile(filepath.Join(templatesDir, "cloudinfer-alerts.yaml"), []byte(observability.GetAlertRulesManifest(ns)), 0644); err != nil {
			return err
		}
	}

	// 3. User Applications Deployment Component (apps/)
	sampleAppYAML := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: cloudinfer-sample-app
  namespace: default
  labels:
    app: sample-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: sample-app
  template:
    metadata:
      labels:
        app: sample-app
    spec:
      containers:
      - name: web
        image: nginxdemos/hello:plain-text
        ports:
        - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: cloudinfer-sample-app
  namespace: default
spec:
  ports:
  - port: 80
    targetPort: 80
  selector:
    app: sample-app
`
	if err := os.WriteFile(filepath.Join(repoDir, "apps", "sample-app.yaml"), []byte(sampleAppYAML), 0644); err != nil {
		return err
	}

	return nil
}

func deployArgoCDApplications(ns string, dryRun bool) error {
	apps := []struct {
		name string
		path string
	}{
		{"cloudinfer-infra", "infra"},
		{"cloudinfer-monitoring", "monitoring"},
		{"cloudinfer-apps", "apps"},
	}

	for _, app := range apps {
		manifest := fmt.Sprintf(`apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: %s
  namespace: %s
spec:
  project: default
  source:
    repoURL: 'git://cloudinfer-git-server.observability.svc.cluster.local/gitops.git'
    path: %s
    targetRevision: HEAD
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
  ignoreDifferences:
    - group: apps
      kind: Deployment
      jsonPointers:
        - /spec/template/metadata/annotations/checksum~1secret
        - /spec/template/metadata/annotations/checksum~1config
    - group: apps
      kind: StatefulSet
      jsonPointers:
        - /spec/template/metadata/annotations/checksum~1secret
        - /spec/template/metadata/annotations/checksum~1config
    - group: admissionregistration.k8s.io
      kind: ValidatingWebhookConfiguration
      jsonPointers:
        - /webhooks/0/clientConfig/caBundle
    - group: admissionregistration.k8s.io
      kind: MutatingWebhookConfiguration
      jsonPointers:
        - /webhooks/0/clientConfig/caBundle
`, app.name, ns, app.path, ns)

		if dryRun {
			fmt.Printf("%s[DRY-RUN] Create ArgoCD Application '%s' sync from path '%s'\n", utils.PrefixInfo, app.name, app.path)
			continue
		}

		tmpPath := filepath.Join(os.TempDir(), app.name+"-application.yaml")
		if err := os.WriteFile(tmpPath, []byte(manifest), 0600); err != nil {
			return err
		}
		defer func() { _ = os.Remove(tmpPath) }()

		_, stderr, err := utils.ExecCommand("", "kubectl", "apply", "-f", tmpPath)
		if err != nil {
			return fmt.Errorf("failed to deploy ArgoCD Application '%s': %w (stderr: %s)", app.name, err, stderr)
		}
		fmt.Printf("%sGitOps Application '%s' initialized and registered in ArgoCD.\n", utils.PrefixOK, app.name)

		if !dryRun {
			_, _, _ = utils.ExecCommand("", "kubectl", "patch", "application", app.name, "-n", ns, "--type", "merge", "-p", `{"operation":{"sync":{"prune":true,"syncStrategy":{"apply":{}}}}}`)
		}
	}

	return nil
}

func applyGitOpsIngress(ns string, dryRun bool) error {
	manifest := fmt.Sprintf(`apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: cloudinfer-observability
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
                name: cloudinfer-prometheus-grafana
                port:
                  number: 80
          - path: /prometheus(/|$)(.*)
            pathType: ImplementationSpecific
            backend:
              service:
                name: cloudinfer-prometheus-kube-prometheus
                port:
                  number: 9090
          - path: /alertmanager(/|$)(.*)
            pathType: ImplementationSpecific
            backend:
              service:
                name: cloudinfer-prometheus-kube-alertmanager
                port:
                  number: 9093
          - path: /argocd(/|$)(.*)
            pathType: ImplementationSpecific
            backend:
              service:
                name: cloudinfer-argocd-server
                port:
                  number: 80
`, ns)

	if dryRun {
		fmt.Printf("%s[DRY-RUN] kubectl apply -f cloudinfer-observability-ingress.yaml\n", utils.PrefixInfo)
		return nil
	}

	tmpPath := filepath.Join(os.TempDir(), "cloudinfer-observability-ingress.yaml")
	if err := os.WriteFile(tmpPath, []byte(manifest), 0600); err != nil {
		return err
	}
	defer func() { _ = os.Remove(tmpPath) }()

	_, stderr, err := utils.ExecCommand("", "kubectl", "apply", "-f", tmpPath)
	if err != nil {
		return fmt.Errorf("failed to apply ingress routing: %w (stderr: %s)", err, stderr)
	}
	return nil
}

func fetchIngressIP(ns string) (string, error) {
	for i := 0; i < 15; i++ {
		out, _, err := utils.ExecCommand("", "kubectl", "get", "ingress", "cloudinfer-observability", "-n", ns, "-o", "jsonpath={.status.loadBalancer.ingress[0].ip}")
		if err == nil && out != "" {
			return out, nil
		}
		time.Sleep(2 * time.Second)
	}
	return "", fmt.Errorf("ingress IP not resolved yet")
}

func waitForArgoCDApps(ns string, dryRun bool) {
	if dryRun {
		return
	}
	stopSpinner := utils.StartSpinner("Waiting for ArgoCD applications to become Healthy and Synced (this may take up to 2 minutes)...")
	var lastPendingApps string

	for i := 0; i < 24; i++ { // Wait up to 2 minutes
		out, _, err := utils.ExecCommand("", "kubectl", "get", "applications", "-n", ns, "-o", "jsonpath={range .items[*]}{.metadata.name}={.status.sync.status},{.status.health.status}\n{end}")
		if err == nil && out != "" {
			var pendingApps []string
			validAppsCount := 0
			lines := strings.Split(strings.TrimSpace(out), "\n")

			for _, line := range lines {
				if line == "" {
					continue
				}
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					validAppsCount++
					name, status := parts[0], parts[1]
					if status != "Synced,Healthy" {
						pendingApps = append(pendingApps, name)
					}
				}
			}
			if validAppsCount > 0 && len(pendingApps) == 0 {
				stopSpinner()
				fmt.Printf("%sAll ArgoCD applications are successfully Synced and Healthy!\n", utils.PrefixOK)
				return
			}

			if len(pendingApps) > 0 {
				sort.Strings(pendingApps)
				currentPending := strings.Join(pendingApps, ", ")
				if currentPending != lastPendingApps {
					stopSpinner()
					stopSpinner = utils.StartSpinner(fmt.Sprintf("Waiting for ArgoCD applications (pending: %s)...", currentPending))
					lastPendingApps = currentPending
				}
			}
		}
		time.Sleep(5 * time.Second)
	}
	stopSpinner()
	fmt.Printf("%sTimeout waiting for applications to become healthy. Check progress using 'cloudinfer status'.\n", utils.PrefixWarn)
}
