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
		// Executing as root — disable Traefik so our NGINX Ingress Controller owns port 80/443
		_, stderr, err = utils.ExecCommand("", "sh", "-c", "curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC='--disable traefik' sh -")
	} else {
		// Executing as standard user, run with sudo
		fmt.Printf("%sAdministrative privileges required. Requesting sudo password...\n", utils.PrefixInfo)
		_, stderr, err = utils.ExecCommandInteractive("", "sudo", "sh", "-c", "curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC='--disable traefik' sh -")
	}

	if err != nil {
		return fmt.Errorf("failed to run get.k3s.io installer: %w (stderr: %s)", err, stderr)
	}
	fmt.Printf("%sK3s installation script completed successfully.\n", utils.PrefixOK)

	// 1b. Ensure the k3s systemd service is running (handles reinstalls where the service may not auto-restart)
	fmt.Printf("%sEnsuring k3s service is active...\n", utils.PrefixInfo)
	if os.Getuid() == 0 {
		_, _, _ = utils.ExecCommand("", "systemctl", "restart", "k3s")
	} else {
		_, _, _ = utils.ExecCommand("", "sudo", "systemctl", "restart", "k3s")
	}
	// Give the API server a moment to bind to port 6443
	fmt.Printf("%sWaiting for K3s API server to initialize...\n", utils.PrefixInfo)
	time.Sleep(10 * time.Second)

	// 2. Read /etc/rancher/k3s/k3s.yaml and write to targetKubeconfig
	fmt.Printf("%sConfiguring Kubeconfig path at: %s...\n", utils.PrefixInfo, targetKubeconfig)
	var configContent string
	if os.Getuid() == 0 {
		data, readErr := os.ReadFile("/etc/rancher/k3s/k3s.yaml")
		if readErr != nil {
			return fmt.Errorf("failed to read /etc/rancher/k3s/k3s.yaml directly: %w", readErr)
		}
		configContent = string(data)
	} else {
		stdout, catStderr, catErr := utils.ExecCommandInteractive("", "sudo", "cat", "/etc/rancher/k3s/k3s.yaml")
		if catErr != nil {
			return fmt.Errorf("failed to read /etc/rancher/k3s/k3s.yaml: %w (stderr: %s)", catErr, catStderr)
		}
		configContent = stdout
	}

	// Ensure destination directory exists
	destDir := filepath.Dir(targetKubeconfig)
	if mkErr := os.MkdirAll(destDir, 0755); mkErr != nil {
		return fmt.Errorf("failed to create kubeconfig directory %s: %w", destDir, mkErr)
	}

	// Write file (automatically owned by current executing user)
	if writeErr := os.WriteFile(targetKubeconfig, []byte(configContent), 0600); writeErr != nil {
		return fmt.Errorf("failed to write kubeconfig file to %s: %w", targetKubeconfig, writeErr)
	}
	if setenvErr := os.Setenv("KUBECONFIG", targetKubeconfig); setenvErr != nil {
		return fmt.Errorf("failed to set KUBECONFIG to %s: %w", targetKubeconfig, setenvErr)
	}
	fmt.Printf("%sKubeconfig written successfully.\n", utils.PrefixOK)

	// 3. Wait for Kubernetes node scheduler readiness
	stopSpinner := utils.StartSpinner("Waiting for Kubernetes nodes to become ready...")

	kubeEnv := map[string]string{"KUBECONFIG": targetKubeconfig}

	// Try waiting for nodes to initialize
	success := false
	for i := 0; i < 60; i++ {
		// Run kubectl wait for node readiness
		_, _, waitErr := utils.ExecCommandEnv("", kubeEnv, "kubectl", "wait", "--for=condition=Ready", "node", "--all", "--timeout=10s")
		if waitErr == nil {
			success = true
			break
		}
		// If kubectl failed or nodes aren't ready, sleep and try again
		time.Sleep(5 * time.Second)
	}
	stopSpinner()

	if !success {
		return fmt.Errorf("kubernetes nodes failed to become ready in time. Please check system logs")
	}

	fmt.Printf("%sAll Kubernetes nodes are in ready state.\n", utils.PrefixOK)
	return nil
}
