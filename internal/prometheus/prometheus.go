package prometheus

import (
	"fmt"
	"os"
	"os/exec"
)

// InstallOnEKS handles the deployment of Prometheus on EKS via Helm
func InstallOnEKS() {
	fmt.Println("\n--- [Phase 1] Installing Prometheus ---")
	fmt.Println("\n⚙️  Adding Prometheus Helm repository...")

	repoAddCmd := exec.Command("helm", "repo", "add", "prometheus-community", "https://prometheus-community.github.io/helm-charts")
	repoAddCmd.Stdout = os.Stdout
	repoAddCmd.Stderr = os.Stderr
	if err := repoAddCmd.Run(); err != nil {
		fmt.Printf("\n❌ Failed to add Helm repository: %v\n", err)
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

	fmt.Println("\n⚙️  Deploying Prometheus to Kubernetes with Ingress enabled...")
	// The --wait flag ensures Helm waits until all Pods are running before exiting.
	installCmd := exec.Command("helm", "upgrade", "--install", "prometheus", "prometheus-community/prometheus",
		"--namespace", "observability",
		"--create-namespace",
		"--set", "server.ingress.enabled=true",
		"--set", "server.ingress.hosts[0]=prometheus.stackpulse.dev",
		"--set", "alertmanager.ingress.enabled=true",
		"--set", "alertmanager.ingress.hosts[0].host=alertmanager.stackpulse.dev",
		"--set", "alertmanager.ingress.hosts[0].paths[0].path=/",
		"--set", "alertmanager.ingress.hosts[0].paths[0].pathType=Prefix",
		"--set", "pushgateway.ingress.enabled=true",
		"--set", "pushgateway.ingress.hosts[0].host=pushgateway.stackpulse.dev",
		"--set", "pushgateway.ingress.hosts[0].paths[0].path=/",
		"--set", "pushgateway.ingress.hosts[0].paths[0].pathType=Prefix",
		"--atomic",
		"--wait")
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	if err := installCmd.Run(); err != nil {
		fmt.Printf("\n❌ Failed to install Prometheus: %v\n", err)
		fmt.Println("💡 HINT: Helm cannot connect to your Kubernetes cluster. Ensure your KUBECONFIG is properly configured.")
		return
	}

	fmt.Println("\n✅ Prometheus installed successfully!")
	fmt.Println("\n🔗 Ingress Access Links:")
	fmt.Println("   ▶ Prometheus Server (Port 9090): http://prometheus.stackpulse.dev")
	fmt.Println("   ▶ Pushgateway       (Port 9091): http://pushgateway.stackpulse.dev")
	fmt.Println("   ▶ Alertmanager      (Port 9093): http://alertmanager.stackpulse.dev")
	fmt.Println("\n💡 Note: Make sure your DNS or local /etc/hosts file maps these domains to your Ingress controller's IP.")
	fmt.Println("\n🎉 1st phase done!")
}

// GetEC2InstallCommands returns the SSM shell commands for installing Prometheus on EC2
func GetEC2InstallCommands(ip string) []string {
	promHost := fmt.Sprintf("prometheus.%s.nip.io", ip)
	alertHost := fmt.Sprintf("alertmanager.%s.nip.io", ip)
	pushHost := fmt.Sprintf("pushgateway.%s.nip.io", ip)

	return []string{
		fmt.Sprintf("helm upgrade --install prometheus prometheus-community/prometheus --namespace observability --create-namespace "+
			"--set server.ingress.enabled=true --set server.ingress.hosts[0]=%s "+
			"--set alertmanager.ingress.enabled=true --set alertmanager.ingress.hosts[0].host=%s "+
			"--set alertmanager.ingress.hosts[0].paths[0].path=/ --set alertmanager.ingress.hosts[0].paths[0].pathType=Prefix "+
			"--set pushgateway.ingress.enabled=true --set pushgateway.ingress.hosts[0].host=%s "+
			"--set pushgateway.ingress.hosts[0].paths[0].path=/ --set pushgateway.ingress.hosts[0].paths[0].pathType=Prefix "+
			"--atomic --timeout 5m --wait", promHost, alertHost, pushHost),
	}
}
