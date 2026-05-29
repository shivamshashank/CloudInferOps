package installer

import (
	"fmt"
	"time"

	"github.com/shivamshashank/StackPulse/internal/utils"
)

// InstallMinikube runs the local minikube start bootstrap process and waits for ready nodes
func InstallMinikube() error {
	fmt.Printf("%sInitializing Minikube cluster setup...\n", utils.PrefixInfo)

	// Run minikube start (blocks until cluster control plane initialized)
	_, stderr, err := utils.ExecCommand("", "minikube", "start")
	if err != nil {
		return fmt.Errorf("failed to execute 'minikube start': %w (stderr: %s)", err, stderr)
	}

	fmt.Printf("%sMinikube cluster started successfully.\n", utils.PrefixOK)

	// Wait for Kubernetes node scheduler readiness
	fmt.Printf("%sWaiting for Kubernetes nodes to become ready...\n", utils.PrefixInfo)
	
	success := false
	for i := 0; i < 12; i++ {
		_, stderr, err = utils.ExecCommand("", "kubectl", "wait", "--for=condition=Ready", "node", "--all", "--timeout=10s")
		if err == nil {
			success = true
			break
		}
		fmt.Printf("%sNodes are initializing, retrying in 5 seconds... (%s)\n", utils.PrefixInfo, stderr)
		time.Sleep(5 * time.Second)
	}

	if !success {
		return fmt.Errorf("Minikube nodes failed to become ready in time. Please check your minikube status")
	}

	fmt.Printf("%sAll Kubernetes nodes are in ready state.\n", utils.PrefixOK)
	return nil
}
