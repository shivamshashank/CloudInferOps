package grafana

import (
	"fmt"
	"os"
	"os/exec"
)

// InstallOnEKS handles the deployment of Grafana on EKS via Helm
func InstallOnEKS() {
	fmt.Println("\n--- [Phase 2] Installing Grafana ---")
	fmt.Println("\n⚙️  Adding Grafana Helm repository...")

	repoAddCmd := exec.Command("helm", "repo", "add", "grafana", "https://grafana.github.io/helm-charts")
	repoAddCmd.Stdout = os.Stdout
	repoAddCmd.Stderr = os.Stderr
	if err := repoAddCmd.Run(); err != nil {
		fmt.Printf("\n❌ Failed to add Grafana Helm repository: %v\n", err)
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

	fmt.Println("\n⚙️  Deploying Grafana to Kubernetes with Ingress enabled...")
	installCmd := exec.Command("helm", "upgrade", "--install", "grafana", "grafana/grafana",
		"--namespace", "observability",
		"--create-namespace",
		"--set", "ingress.enabled=true",
		"--set", "ingress.ingressClassName=nginx",
		"--set", "ingress.hosts[0]=grafana.stackpulse.dev",
		"--set", "ingress.path=/",
		"--set", "ingress.pathType=Prefix",
		"--set", "adminPassword=admin",
		"--set", "sidecar.dashboards.enabled=true",
		"--set", "sidecar.dashboards.label=grafana_dashboard",
		"--set", "dashboardProviders.dashboardproviders.yaml.apiVersion=1",
		"--set", "dashboardProviders.dashboardproviders.yaml.providers[0].name=default",
		"--set", "dashboardProviders.dashboardproviders.yaml.providers[0].orgId=1",
		"--set", "dashboardProviders.dashboardproviders.yaml.providers[0].folder=General",
		"--set", "dashboardProviders.dashboardproviders.yaml.providers[0].type=file",
		"--set", "dashboardProviders.dashboardproviders.yaml.providers[0].disableDeletion=false",
		"--set", "dashboardProviders.dashboardproviders.yaml.providers[0].editable=true",
		"--set", "dashboardProviders.dashboardproviders.yaml.providers[0].options.path=/var/lib/grafana/dashboards/default",
		"--atomic",
		"--wait")
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	if err := installCmd.Run(); err != nil {
		fmt.Printf("\n❌ Failed to install Grafana: %v\n", err)
		return
	}

	fmt.Println("\n✅ Grafana installed successfully!")
}

// GetEC2InstallCommands returns the SSM shell commands for installing Grafana on EC2
func GetEC2InstallCommands(ip string) []string {
	grafanaHost := fmt.Sprintf("grafana.%s.nip.io", ip)

	return []string{
		"helm repo add grafana https://grafana.github.io/helm-charts",
		"helm repo update",
		fmt.Sprintf("helm upgrade --install grafana grafana/grafana --namespace observability --create-namespace "+
			"--set ingress.enabled=true --set ingress.hosts[0]=%s "+
			"--set ingress.path=/ --set ingress.pathType=Prefix "+
			"--set adminPassword=admin "+
			"--atomic --timeout 5m --wait", grafanaHost),
	}
}
