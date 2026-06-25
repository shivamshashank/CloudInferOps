package installer

import (
	"fmt"
	"os"
	"runtime"

	"github.com/shivamshashank/CloudInferOps/internal/utils"
)

// DownloadKubectlBinary downloads the kubectl binary from the official release and installs it to /usr/local/bin/kubectl.
func DownloadKubectlBinary() error {
	arch := runtime.GOARCH
	url := fmt.Sprintf("https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/%s/kubectl", arch)
	tmpPath := "/tmp/kubectl"

	fmt.Printf("%sDownloading kubectl binary for linux/%s...\n", utils.PrefixInfo, arch)

	_, stderr, err := utils.ExecCommand("", "sh", "-c", fmt.Sprintf("curl -Lo %s %s && chmod +x %s", tmpPath, url, tmpPath))
	if err != nil {
		return fmt.Errorf("failed to download kubectl: %w (stderr: %s)", err, stderr)
	}

	// Move to /usr/local/bin (may need sudo)
	if os.Getuid() == 0 {
		_, stderr, err = utils.ExecCommand("", "mv", tmpPath, "/usr/local/bin/kubectl")
	} else {
		_, stderr, err = utils.ExecCommandInteractive("", "sudo", "mv", tmpPath, "/usr/local/bin/kubectl")
	}
	if err != nil {
		return fmt.Errorf("failed to install kubectl to /usr/local/bin: %w (stderr: %s)", err, stderr)
	}

	fmt.Printf("%sKubectl binary installed successfully.\n", utils.PrefixOK)
	return nil
}
