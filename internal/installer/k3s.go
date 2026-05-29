package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/shivamshashank/StackPulse/internal/utils"
)

// InstallK3s runs the K3s installation script, copies its kubeconfig, fixes permissions, and waits for cluster nodes to be ready.
func InstallK3s(targetKubeconfig string) error {
	// 1. Run official get.k3s.io installer script
	fmt.Printf("%sDownloading and running k3s installer...\n", utils.PrefixInfo)

	var err error
	var stderr string
	if os.Getuid() == 0 {
		// Executing as root
		_, stderr, err = utils.ExecCommand("", "sh", "-c", "curl -sfL https://get.k3s.io | sh -")
	} else {
		// Executing as standard user, run with sudo
		fmt.Printf("%sAdministrative privileges required. Requesting sudo password...\n", utils.PrefixInfo)
		_, stderr, err = utils.ExecCommand("", "sudo", "sh", "-c", "curl -sfL https://get.k3s.io | sh -")
	}

	if err != nil {
		return fmt.Errorf("failed to run get.k3s.io installer: %w (stderr: %s)", err, stderr)
	}
	fmt.Printf("%sK3s installation script completed successfully.\n", utils.PrefixOK)

	// 2. Read /etc/rancher/k3s/k3s.yaml and write to targetKubeconfig
	fmt.Printf("%sConfiguring Kubeconfig path at: %s...\n", utils.PrefixInfo, targetKubeconfig)
	var configContent string
	if os.Getuid() == 0 {
		data, err := os.ReadFile("/etc/rancher/k3s/k3s.yaml")
		if err != nil {
			return fmt.Errorf("failed to read /etc/rancher/k3s/k3s.yaml directly: %w", err)
		}
		configContent = string(data)
	} else {
		stdout, stderr, err := utils.ExecCommand("", "sudo", "cat", "/etc/rancher/k3s/k3s.yaml")
		if err != nil {
			return fmt.Errorf("failed to read /etc/rancher/k3s/k3s.yaml: %w (stderr: %s)", err, stderr)
		}
		configContent = stdout
	}

	// Ensure destination directory exists
	destDir := filepath.Dir(targetKubeconfig)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create kubeconfig directory %s: %w", destDir, err)
	}

	// Write file (automatically owned by current executing user)
	if err := os.WriteFile(targetKubeconfig, []byte(configContent), 0600); err != nil {
		return fmt.Errorf("failed to write kubeconfig file to %s: %w", targetKubeconfig, err)
	}
	fmt.Printf("%sKubeconfig written successfully.\n", utils.PrefixOK)

	// 3. Wait for Kubernetes node scheduler readiness
	fmt.Printf("%sWaiting for Kubernetes nodes to become ready...\n", utils.PrefixInfo)
	
	// Try waiting up to 60 seconds
	success := false
	for i := 0; i < 12; i++ {
		// Run kubectl wait for node readiness
		_, stderr, err = utils.ExecCommand("", "kubectl", "wait", "--for=condition=Ready", "node", "--all", "--timeout=10s")
		if err == nil {
			success = true
			break
		}
		// If kubectl failed or nodes aren't ready, sleep and try again
		fmt.Printf("%sNodes are initializing, retrying in 5 seconds... (%s)\n", utils.PrefixInfo, stderr)
		time.Sleep(5 * time.Second)
	}

	if !success {
		return fmt.Errorf("Kubernetes nodes failed to become ready in time. Please check system logs")
	}

	fmt.Printf("%sAll Kubernetes nodes are in ready state.\n", utils.PrefixOK)
	return nil
}
