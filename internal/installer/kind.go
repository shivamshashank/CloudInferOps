package installer

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/shivamshashank/StackPulse/internal/utils"
)

// DownloadKindBinary downloads the kind binary from the official release and installs it to /usr/local/bin/kind.
func DownloadKindBinary() error {
	arch := runtime.GOARCH
	if arch == "amd64" {
		arch = "amd64"
	} else if arch == "arm64" {
		arch = "arm64"
	}

	url := fmt.Sprintf("https://kind.sigs.k8s.io/dl/latest/kind-linux-%s", arch)
	tmpPath := "/tmp/kind"

	fmt.Printf("%sDownloading kind binary for linux/%s...\n", utils.PrefixInfo, arch)

	_, stderr, err := utils.ExecCommand("", "sh", "-c", fmt.Sprintf("curl -Lo %s %s && chmod +x %s", tmpPath, url, tmpPath))
	if err != nil {
		return fmt.Errorf("failed to download kind: %w (stderr: %s)", err, stderr)
	}

	// Move to /usr/local/bin (may need sudo)
	if os.Getuid() == 0 {
		_, stderr, err = utils.ExecCommand("", "mv", tmpPath, "/usr/local/bin/kind")
	} else {
		_, stderr, err = utils.ExecCommandInteractive("", "sudo", "mv", tmpPath, "/usr/local/bin/kind")
	}
	if err != nil {
		return fmt.Errorf("failed to install kind to /usr/local/bin: %w (stderr: %s)", err, stderr)
	}

	fmt.Printf("%sKind binary installed successfully.\n", utils.PrefixOK)
	return nil
}

// InstallKind creates a local kind cluster and waits for ready nodes.
func InstallKind() error {
	const clusterName = "stackpulse"

	fmt.Printf("%sInitializing kind cluster setup...\n", utils.PrefixInfo)

	clusters, _, existingErr := utils.ExecCommand("", "kind", "get", "clusters")
	if existingErr != nil {
		return fmt.Errorf("failed to list kind clusters: %w", existingErr)
	}

	if strings.Contains("\n"+clusters+"\n", "\n"+clusterName+"\n") {
		fmt.Printf("%sKind cluster '%s' already exists. Reusing it.\n", utils.PrefixOK, clusterName)
		_, stderr, err := utils.ExecCommand("", "kubectl", "config", "use-context", "kind-"+clusterName)
		if err != nil {
			return fmt.Errorf("failed to switch to existing kind cluster '%s': %w (stderr: %s)", clusterName, err, stderr)
		}
	} else {
		_, stderr, err := utils.ExecCommand("", "kind", "create", "cluster", "--name", clusterName)
		if err != nil {
			return fmt.Errorf("failed to execute 'kind create cluster --name %s': %w (stderr: %s)", clusterName, err, stderr)
		}
	}

	fmt.Printf("%sKind cluster started successfully.\n", utils.PrefixOK)
	fmt.Printf("%sWaiting for Kubernetes nodes to become ready...\n", utils.PrefixInfo)

	success := false
	for i := 0; i < 24; i++ {
		_, stderr, err := utils.ExecCommand("", "kubectl", "wait", "--for=condition=Ready", "node", "--all", "--timeout=10s")
		if err == nil {
			success = true
			break
		}
		fmt.Printf("%sNodes are initializing, retrying in 5 seconds... (%s)\n", utils.PrefixInfo, stderr)
		time.Sleep(5 * time.Second)
	}

	if !success {
		return fmt.Errorf("kind nodes failed to become ready in time. Please check 'kind get clusters' and 'kubectl get nodes'")
	}

	fmt.Printf("%sAll Kubernetes nodes are in ready state.\n", utils.PrefixOK)
	return nil
}
