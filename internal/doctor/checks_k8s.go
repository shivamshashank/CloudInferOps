package doctor

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/shivamshashank/StackPulse/internal/utils"
)

// runKubectl runs a kubectl command using the preferred kubeconfig file path.
func runKubectl(arg ...string) (string, string, error) {
	if kubeconfig := preferredKubeconfig(); kubeconfig != "" {
		return utils.ExecCommandEnv("", map[string]string{"KUBECONFIG": kubeconfig}, "kubectl", arg...)
	}
	return utils.ExecCommand("", "kubectl", arg...)
}

// CheckTool looks up a command-line utility in $PATH
func CheckTool(name string, critical bool) CheckResult {
	_, err := exec.LookPath(name)
	if err != nil {
		status := StatusWarn
		if critical {
			status = StatusError
		}
		return CheckResult{
			Name:    name,
			Status:  status,
			Message: fmt.Sprintf("%s not found", name),
		}
	}
	return CheckResult{
		Name:    name,
		Status:  StatusOK,
		Message: fmt.Sprintf("%s found", name),
	}
}

// CheckK8sCluster inspects the current Kubernetes context, connectivity, and resources.
// It returns check results and a boolean indicating if a cluster was successfully detected and reachable.
func CheckK8sCluster() ([]CheckResult, bool) {
	var results []CheckResult

	// First verify kubectl exists in PATH
	if _, err := exec.LookPath("kubectl"); err != nil {
		return []CheckResult{
			{
				Name:    "Kubernetes Cluster",
				Status:  StatusWarn,
				Message: "Kubernetes cluster not detected (kubectl missing)",
			},
		}, false
	}

	// 1. Check current context
	context, _, err := runKubectl("config", "current-context")
	if err != nil || context == "" {
		return []CheckResult{
			{
				Name:    "Kubernetes Cluster",
				Status:  StatusWarn,
				Message: "Kubernetes cluster not detected (failed to get current context)",
			},
		}, false
	}

	// 2. Check cluster reachability (kubectl cluster-info)
	_, _, err = runKubectl("cluster-info")
	if err != nil {
		return []CheckResult{
			{
				Name:    "Kubernetes Reachability",
				Status:  StatusWarn,
				Message: fmt.Sprintf("Kubernetes cluster unreachable (context: %s)", context),
			},
		}, false
	}

	results = append(results, CheckResult{
		Name:    "Kubernetes Context",
		Status:  StatusOK,
		Message: fmt.Sprintf("Kubernetes cluster detected (context: %s)", context),
	})

	// 3. Count ready nodes
	nodesOut, _, err := runKubectl("get", "nodes", "--no-headers")
	if err == nil && nodesOut != "" {
		lines := strings.Split(strings.TrimSpace(nodesOut), "\n")
		readyCount := 0
		for _, line := range lines {
			if strings.Contains(line, " Ready") {
				readyCount++
			}
		}
		results = append(results, CheckResult{
			Name:    "Kubernetes Nodes",
			Status:  StatusOK,
			Message: fmt.Sprintf("Nodes ready: %d", readyCount),
		})
	}

	// 4. Check for StorageClass
	scOut, _, err := runKubectl("get", "storageclass", "--no-headers")
	if err != nil || strings.TrimSpace(scOut) == "" {
		results = append(results, CheckResult{
			Name:    "StorageClass Availability",
			Status:  StatusWarn,
			Message: "StorageClass not found (Persistent Volumes might fail to bind)",
		})
	} else {
		scs := strings.Split(strings.TrimSpace(scOut), "\n")
		scNames := []string{}
		for _, scLine := range scs {
			fields := strings.Fields(scLine)
			if len(fields) > 0 {
				scNames = append(scNames, fields[0])
			}
		}
		results = append(results, CheckResult{
			Name:    "StorageClass Availability",
			Status:  StatusOK,
			Message: fmt.Sprintf("StorageClass found: %s", strings.Join(scNames, ", ")),
		})
	}

	return results, true
}

// CheckK8sVersion checks the server version of the active Kubernetes cluster
func CheckK8sVersion() CheckResult {
	if _, err := exec.LookPath("kubectl"); err != nil {
		return CheckResult{
			Name:    "Kubernetes Version",
			Status:  StatusWarn,
			Message: "Kubernetes Version: Unknown (kubectl missing)",
		}
	}
	out, _, err := runKubectl("version", "--output=json")
	if err != nil {
		outShort, _, errShort := runKubectl("version", "--short")
		if errShort == nil && outShort != "" {
			return CheckResult{
				Name:    "Kubernetes Version",
				Status:  StatusOK,
				Message: fmt.Sprintf("Kubernetes server version: %s", strings.ReplaceAll(strings.TrimSpace(outShort), "\n", " ")),
			}
		}
		return CheckResult{
			Name:    "Kubernetes Version",
			Status:  StatusWarn,
			Message: "Kubernetes Version: Unknown (failed to query cluster version)",
		}
	}

	var vData struct {
		ServerVersion struct {
			GitVersion string `json:"gitVersion"`
		} `json:"serverVersion"`
	}
	if err := json.Unmarshal([]byte(out), &vData); err == nil && vData.ServerVersion.GitVersion != "" {
		return CheckResult{
			Name:    "Kubernetes Version",
			Status:  StatusOK,
			Message: fmt.Sprintf("Kubernetes server version: %s", vData.ServerVersion.GitVersion),
		}
	}

	if strings.Contains(out, `"gitVersion"`) {
		parts := strings.Split(out, `"gitVersion":`)
		if len(parts) > 1 {
			sub := strings.Split(parts[1], ",")
			if len(sub) > 0 {
				ver := strings.Trim(sub[0], ` "`)
				return CheckResult{
					Name:    "Kubernetes Version",
					Status:  StatusOK,
					Message: fmt.Sprintf("Kubernetes server version: %s", ver),
				}
			}
		}
	}

	return CheckResult{
		Name:    "Kubernetes Version",
		Status:  StatusOK,
		Message: "Kubernetes server version detected",
	}
}

// CheckIngressController checks if an Ingress controller is installed
func CheckIngressController() CheckResult {
	if _, err := exec.LookPath("kubectl"); err != nil {
		return CheckResult{
			Name:    "Kubernetes Ingress",
			Status:  StatusWarn,
			Message: "Kubernetes Ingress: Ingress controller not detected (kubectl missing)",
		}
	}
	out, _, err := runKubectl("get", "ingressclass", "--no-headers")
	if err == nil && strings.TrimSpace(out) != "" {
		classes := []string{}
		lines := strings.Split(strings.TrimSpace(out), "\n")
		for _, line := range lines {
			f := strings.Fields(line)
			if len(f) > 0 {
				classes = append(classes, f[0])
			}
		}
		return CheckResult{
			Name:    "Kubernetes Ingress",
			Status:  StatusOK,
			Message: fmt.Sprintf("Ingress classes found: %s", strings.Join(classes, ", ")),
		}
	}

	podsOut, _, err := runKubectl("get", "pods", "-A", "-l", "app.kubernetes.io/name=ingress-nginx", "--no-headers")
	if err == nil && strings.TrimSpace(podsOut) != "" {
		return CheckResult{
			Name:    "Kubernetes Ingress",
			Status:  StatusOK,
			Message: "Ingress controller: NGINX Ingress Controller detected in cluster",
		}
	}
	return CheckResult{
		Name:    "Kubernetes Ingress",
		Status:  StatusWarn,
		Message: "Ingress controller: None detected (ingress-nginx recommended for path routing)",
	}
}

// CheckCloudCredentials scans local and env credentials for AWS, GCP, and Azure
func CheckCloudCredentials() CheckResult {
	detected := []string{}
	home, err := utils.GetRealHomeDir()

	// 1. AWS
	hasAWS := false
	if os.Getenv("AWS_ACCESS_KEY_ID") != "" {
		hasAWS = true
	} else if err == nil {
		if _, statErr := os.Stat(filepath.Join(home, ".aws", "credentials")); statErr == nil {
			hasAWS = true
		}
	}
	if hasAWS {
		detected = append(detected, "AWS")
	}

	// 2. GCP
	hasGCP := false
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") != "" {
		hasGCP = true
	} else if err == nil {
		if _, statErr := os.Stat(filepath.Join(home, ".config", "gcloud", "application_default_credentials.json")); statErr == nil {
			hasGCP = true
		}
	}
	if hasGCP {
		detected = append(detected, "GCP")
	}

	// 3. Azure
	hasAzure := false
	if os.Getenv("AZURE_TENANT_ID") != "" || os.Getenv("AZURE_CLIENT_ID") != "" {
		hasAzure = true
	} else if err == nil {
		if _, statErr := os.Stat(filepath.Join(home, ".azure")); statErr == nil {
			hasAzure = true
		}
	}
	if hasAzure {
		detected = append(detected, "Azure")
	}

	if len(detected) == 0 {
		return CheckResult{
			Name:    "Cloud Credentials",
			Status:  StatusInfo,
			Message: "Cloud Credentials: None detected (running in local/bare-metal mode)",
		}
	}

	return CheckResult{
		Name:    "Cloud Credentials",
		Status:  StatusOK,
		Message: fmt.Sprintf("Cloud Integrations active: %s", strings.Join(detected, ", ")),
	}
}

func preferredKubeconfig() string {
	if kubeconfig := strings.TrimSpace(os.Getenv("KUBECONFIG")); kubeconfig != "" {
		return kubeconfig
	}

	home, err := utils.GetRealHomeDir()
	if err != nil {
		return ""
	}
	kubeconfig := filepath.Join(home, ".kube", "config")
	if _, err := os.Stat(kubeconfig); err == nil {
		return kubeconfig
	}

	return ""
}
