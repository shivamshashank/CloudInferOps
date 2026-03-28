package ingress

// GetEC2InstallCommands returns the SSM shell commands for checking and installing an NGINX Ingress Controller
func GetEC2InstallCommands() []string {
	return []string{
		"if [ -z \"$(kubectl get ingressclass -o name 2>/dev/null)\" ]; then " +
			"echo '📦 No Ingress Controller found. Installing NGINX...'; " +
			"helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx && helm repo update; " +
			"helm upgrade --install ingress-nginx ingress-nginx/ingress-nginx --namespace ingress-nginx --create-namespace " +
			"--set controller.hostNetwork=true --set controller.service.type=ClusterIP " +
			"--set controller.admissionWebhooks.enabled=false --set controller.ingressClassResource.default=true " +
			"--set controller.tolerations[0].operator=Exists " +
			"--timeout 90s --wait || (echo '\n❌ NGINX Ingress failed to start. Pod details:' && kubectl describe pods -n ingress-nginx && exit 1); " +
			"else " +
			"echo '✅ Existing Ingress Controller detected. Skipping NGINX installation...'; " +
			"fi",
	}
}
