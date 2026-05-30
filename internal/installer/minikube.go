package installer

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/shivamshashank/StackPulse/internal/utils"
)

// DownloadMinikubeBinary downloads the minikube binary from the official release and installs it to /usr/local/bin/minikube.
func DownloadMinikubeBinary() error {
	arch := runtime.GOARCH
	url := fmt.Sprintf("https://storage.googleapis.com/minikube/releases/latest/minikube-linux-%s", arch)
	tmpPath := "/tmp/minikube"

	fmt.Printf("%sDownloading minikube binary for linux/%s...\n", utils.PrefixInfo, arch)

	_, stderr, err := utils.ExecCommand("", "sh", "-c", fmt.Sprintf("curl -Lo %s %s && chmod +x %s", tmpPath, url, tmpPath))
	if err != nil {
		return fmt.Errorf("failed to download minikube: %w (stderr: %s)", err, stderr)
	}

	// Move to /usr/local/bin (may need sudo)
	if os.Getuid() == 0 {
		_, stderr, err = utils.ExecCommand("", "mv", tmpPath, "/usr/local/bin/minikube")
	} else {
		_, stderr, err = utils.ExecCommandInteractive("", "sudo", "mv", tmpPath, "/usr/local/bin/minikube")
	}
	if err != nil {
		return fmt.Errorf("failed to install minikube to /usr/local/bin: %w (stderr: %s)", err, stderr)
	}

	fmt.Printf("%sMinikube binary installed successfully.\n", utils.PrefixOK)
	return nil
}

// InstallMinikube runs the local minikube start bootstrap process and waits for ready nodes
func InstallMinikube() error {
	fmt.Printf("%sInitializing Minikube cluster setup...\n", utils.PrefixInfo)

	// Build minikube start command — use Docker driver and --force for root compatibility
	minikubeArgs := []string{"start", "--driver=docker"}
	if os.Getuid() == 0 {
		minikubeArgs = append(minikubeArgs, "--force")
	}

	// Run minikube start (blocks until cluster control plane initialized)
	_, stderr, err := utils.ExecCommandStream("", "minikube", minikubeArgs...)
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
		return fmt.Errorf("minikube nodes failed to become ready in time. Please check your minikube status")
	}

	fmt.Printf("%sAll Kubernetes nodes are in ready state.\n", utils.PrefixOK)
	return nil
}
