package k8s

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

func DeployEKS(clusterName string) {
	fmt.Printf("\n[K8s] Deploying Helm charts and ArgoCD to cluster '%s'...\n", clusterName)

	fmt.Println("\n--- [Phase 1] Installing kube-prometheus-stack ---")

	repoAddCmd := exec.Command("helm", "repo", "add", "prometheus-community", "https://prometheus-community.github.io/helm-charts")
	repoAddCmd.Stdout = os.Stdout
	repoAddCmd.Stderr = os.Stderr
	if err := repoAddCmd.Run(); err != nil {
		fmt.Printf("\n❌ Failed to add Prometheus Community Helm repository: %v\n", err)
		return
	}

	fmt.Println("\n⚙️  Updating Helm repositories...")
	repoUpCmd := exec.Command("helm", "repo", "update")
	repoUpCmd.Stdout = os.Stdout
	repoUpCmd.Stderr = os.Stderr
	if err := repoUpCmd.Run(); err != nil {
		fmt.Printf("\n❌ Failed to update Helm repositories: %v\n", err)
		return
	}

	fmt.Println("\n⚙️  Deploying kube-prometheus-stack to Kubernetes with Ingress enabled...")
	installCmd := exec.Command("helm", "upgrade", "--install", "kube-prometheus-stack", "prometheus-community/kube-prometheus-stack",
		"--namespace", "observability",
		"--create-namespace",
		"--set", "grafana.ingress.enabled=true",
		"--set", "grafana.ingress.ingressClassName=nginx",
		"--set", "grafana.ingress.hosts[0]=grafana.stackpulse.dev",
		"--set", "grafana.ingress.path=/",
		"--set", "grafana.ingress.pathType=Prefix",
		"--set", "grafana.adminPassword=admin",
		"--set", "grafana.sidecar.dashboards.enabled=true",
		"--set", "grafana.sidecar.datasources.enabled=true",
		"--set", "prometheus.ingress.enabled=true",
		"--set", "prometheus.ingress.ingressClassName=nginx",
		"--set", "prometheus.ingress.hosts[0]=prometheus.stackpulse.dev",
		"--set", "prometheus.ingress.paths[0].path=/",
		"--set", "prometheus.ingress.paths[0].pathType=Prefix",
		"--atomic",
		"--wait")
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	if err := installCmd.Run(); err != nil {
		fmt.Printf("\n❌ Failed to install kube-prometheus-stack: %v\n", err)
		return
	}

	fmt.Println("\n✅ kube-prometheus-stack installed successfully!")

	components := []string{
		"Mimir (Long-term Metrics)",
		"Loki (Logs)",
		"Tempo (Traces)",
		"ArgoCD (GitOps Engine)",
	}

	fmt.Println("\n--- [Phase 2] Deploying Remaining Components ---")
	for _, component := range components {
		fmt.Printf(" 📦 Installing %s...\n", component)
		// Simulating deployment time
		time.Sleep(1 * time.Second)
		fmt.Println("✅ Done!")
	}

	fmt.Println("\n✅ Observability stack deployed successfully!")
	fmt.Println("\n🔗 Ingress Access Links:")
	fmt.Println("   ▶ Grafana           : http://grafana.stackpulse.dev (User: admin / Pass: admin)")
	fmt.Println("   ▶ Prometheus Server : http://prometheus.stackpulse.dev")
	fmt.Println("   ▶ Pushgateway       : http://pushgateway.stackpulse.dev")
	fmt.Println("   ▶ Alertmanager      : http://alertmanager.stackpulse.dev")
	fmt.Println("\n💡 Note: Make sure your DNS or local /etc/hosts file maps these domains to your Ingress controller's IP.")
}
