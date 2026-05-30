package doctor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/shivamshashank/StackPulse/internal/utils"
)

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

	kubectl := func(arg ...string) (string, string, error) {
		if kubeconfig := preferredKubeconfig(); kubeconfig != "" {
			return utils.ExecCommandEnv("", map[string]string{"KUBECONFIG": kubeconfig}, "kubectl", arg...)
		}
		return utils.ExecCommand("", "kubectl", arg...)
	}

	// 1. Check current context
	context, _, err := kubectl("config", "current-context")
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
	_, _, err = kubectl("cluster-info")
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
	nodesOut, _, err := kubectl("get", "nodes", "--no-headers")
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
	scOut, _, err := kubectl("get", "storageclass", "--no-headers")
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
