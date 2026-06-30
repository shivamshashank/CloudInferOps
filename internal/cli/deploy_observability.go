package cli

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/shivamshashank/CloudInferOps/internal/config"
	"github.com/shivamshashank/CloudInferOps/internal/doctor"
	"github.com/shivamshashank/CloudInferOps/internal/installer"
	"github.com/shivamshashank/CloudInferOps/internal/observability"
	"github.com/shivamshashank/CloudInferOps/internal/utils"
	"github.com/spf13/cobra"
)

var deployDryRun bool

func addDeployObservabilityFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&deployDryRun, "dry-run", false, "Show what would be deployed without actually deploying")
}

func runDeployObservability(_ *cobra.Command, _ []string) error {
	// 1. Pre-flight check: Verify cluster reachability and install if missing.
	if err := ensureKubernetes(deployDryRun, "deploy platform"); err != nil {
		return err
	}

	// 1.5 Pre-flight check: Verify Helm is installed and install if missing
	if _, err := exec.LookPath("helm"); err != nil {
		if deployDryRun {
			fmt.Printf("%s[DRY-RUN] Would install Helm\n", utils.PrefixInfo)
		} else {
			fmt.Printf("%sHelm is required but was not found. Installing Helm now...\n", utils.PrefixInfo)
			if _, stderr, installErr := utils.ExecCommandInteractive("", "bash", "-c", "curl -fsSL https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash"); installErr != nil { // #nosec G204
				return fmt.Errorf("failed to install Helm: %w (stderr: %s)", installErr, stderr)
			}
			fmt.Printf("%sHelm installed successfully.\n", utils.PrefixOK)
		}
	}

	// 2. Load configuration (fallback on defaults if not initialized)
	if err := config.InitConfig(false); err != nil {
		fmt.Printf("%sConfiguration file not found. Deploying with default settings...\n", utils.PrefixInfo)
		config.GlobalConfig = config.DefaultConfig()
	}

	// 3. Trigger observability stack deployment
	fmt.Printf("\n%sKubernetes is ready. Deploying the observability stack...\n", utils.PrefixInfo)
	if err := observability.DeployObservability(deployDryRun); err != nil {
		return fmt.Errorf("failed to deploy observability stack: %w", err)
	}

	return nil
}

// ensureKubernetes checks for a Kubernetes cluster and prompts for installation if one is not found.
// This function is shared between bootstrap and deploy commands.
func ensureKubernetes(dryRun bool, actionName string) error {
	_, hasK8s := doctor.CheckK8sCluster()
	if hasK8s {
		fmt.Printf("%sExisting Kubernetes cluster detected. Skipping installation.\n", utils.PrefixOK)
		return nil
	}

	if dryRun {
		return fmt.Errorf("kubernetes cluster unreachable (dry-run bypassed setup)")
	}

	// Prompt the user to install a local cluster or exit so they can bring their own.
	choice, err := promptClusterOption(os.Stdin)
	if err != nil {
		return err
	}

	if choice == "no" {
		fmt.Printf("%s%s cancelled. Install or start Kubernetes, then rerun %s.\n", utils.PrefixWarn, actionName, actionName)
		return fmt.Errorf("kubernetes cluster unreachable")
	}

	switch choice {
	case "kubeadm":
		if runtime.GOOS != "linux" {
			return fmt.Errorf("kubeadm installation is only supported on Linux")
		}
		fmt.Printf("%sStarting Kubernetes cluster installation via kubeadm...\n", utils.PrefixInfo)
		if err := installer.InstallKubeadm(); err != nil {
			return fmt.Errorf("kubeadm installation failed: %w", err)
		}
		// Set KUBECONFIG so that subsequent commands (Helm, kubectl, doctor checks)
		// in this process can reach the cluster. When running as sudo, the default
		// kubeconfig lookup (~/.kube/config) points to root's home which doesn't
		// have the config. admin.conf is always available after kubeadm init.
		_ = os.Setenv("KUBECONFIG", "/etc/kubernetes/admin.conf")
	}

	if _, hasK8s = doctor.CheckK8sCluster(); !hasK8s {
		return fmt.Errorf("kubernetes cluster unreachable after setup")
	}

	fmt.Printf("%sKubernetes cluster setup completed successfully.\n\n", utils.PrefixOK)
	return nil
}

func promptClusterOption(reader io.Reader) (string, error) {
	fmt.Println("Kubernetes cluster not detected.")
	fmt.Println("Would you like CloudInferOps to install a Kubernetes cluster via kubeadm for you?")
	fmt.Println("  1. kubeadm (Recommended for production on Linux)")
	fmt.Println("  2. No, I will install or manage Kubernetes myself")
	fmt.Print("Choose an option [1-2] (default: 2): ")

	var response string
	_, _ = fmt.Fscanln(reader, &response)
	response = strings.TrimSpace(response)

	switch response {
	case "1":
		return "kubeadm", nil
	case "2", "":
		return "no", nil
	default:
		fmt.Printf("%sInvalid choice '%s'. Defaulting to exit.\n", utils.PrefixWarn, response)
		return "no", nil
	}
}
