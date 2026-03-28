package k8s

import (
	"fmt"
	"time"

	"github.com/yourusername/stackpulse/internal/prometheus"
)

func DeployEKS(clusterName string) {
	fmt.Printf("\n[K8s] Deploying Helm charts and ArgoCD to cluster '%s'...\n", clusterName)

	// Phase 1: Prometheus
	prometheus.InstallOnEKS()

	components := []string{
		"Mimir (Long-term Metrics)",
		"Loki (Logs)",
		"Tempo (Traces)",
		"Grafana (Dashboards)",
		"Alertmanager (Alerting)",
		"ArgoCD (GitOps Engine)",
	}

	fmt.Println("\n--- [Phase 2] Deploying Remaining Components ---")
	for _, component := range components {
		fmt.Printf(" 📦 Installing %s...\n", component)
		// Simulating deployment time
		time.Sleep(1 * time.Second)
		fmt.Println("✅ Done!")
	}
	fmt.Println()
}
