package k8s

import "fmt"

func Deploy(clusterName string) {
	fmt.Printf("[K8s] Deploying Helm charts and ArgoCD to cluster '%s'...\n", clusterName)
}
