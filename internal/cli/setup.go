package cli

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/shivamshashank/StackPulse/internal/config"
	"github.com/shivamshashank/StackPulse/internal/doctor"
	"github.com/shivamshashank/StackPulse/internal/installer"
	"github.com/shivamshashank/StackPulse/internal/utils"
	"github.com/spf13/cobra"
)

var (
	setupType string
	setupYes  bool
)

// setupCmd represents the parent setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup environment prerequisites like Kubernetes",
	Long:  `Parent command for configuring system infrastructure and bootstrapping local Kubernetes clusters.`,
}

// k8sCmd represents the setup k8s subcommand
// k8sCmd represents the setup k8s subcommand
var k8sCmd = &cobra.Command{
	Use:   "k8s",
	Short: "Install and configure a local Kubernetes cluster (k3s or minikube)",
	Long:  `Automatically installs lightweight single-node Kubernetes (k3s on Linux) or bootstraps Minikube (cross-platform on macOS/Linux) and configures cluster permissions.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. OS platform safeguard: exclusively support Linux
		if runtime.GOOS != "linux" {
			return fmt.Errorf("setup is only supported on Linux")
		}

		// 2. Validate installation type
		setupType = strings.ToLower(strings.TrimSpace(setupType))
		if setupType != "k3s" && setupType != "minikube" {
			return fmt.Errorf("unsupported Kubernetes type '%s' (only 'k3s' and 'minikube' are supported)", setupType)
		}

		// 3. Initialize configuration (create default if missing)
		if err := config.InitConfig(true); err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		// 4. Check Minikube pre-requisites
		if setupType == "minikube" {
			if _, err := exec.LookPath("minikube"); err != nil {
				fmt.Printf("%s'minikube' binary not found in your $PATH.\n", utils.PrefixError)
				fmt.Printf("%sPlease install Minikube using the Linux installer:\n", utils.PrefixInfo)
				fmt.Printf("      Linux:  curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64 && sudo install minikube-linux-amd64 /usr/local/bin/minikube\n")
				return fmt.Errorf("minikube dependency missing")
			}
		}

		// 5. Pre-check if K8s is already present & reachable
		_, hasK8s := doctor.CheckK8sCluster()
		if hasK8s {
			fmt.Printf("%sKubernetes cluster already detected. Skipping installation.\n", utils.PrefixOK)
			return nil
		}

		// 6. Interactive confirmation (unless bypassed via --yes)
		if !setupYes {
			fmt.Printf("%sKubernetes was not detected.\n", utils.PrefixWarn)
			fmt.Printf("%sStackPulse will install %s on this machine.\n", utils.PrefixInfo, setupType)
			confirm, err := promptYesNo("Continue?")
			if err != nil {
				return err
			}
			if !confirm {
				fmt.Printf("%sSetup cancelled.\n", utils.PrefixWarn)
				return nil
			}
		}

		// 7. Execute K8s installation engine
		if setupType == "minikube" {
			fmt.Printf("%sStarting Minikube cluster installation...\n", utils.PrefixInfo)
			if err := installer.InstallMinikube(); err != nil {
				return fmt.Errorf("Minikube installation failed: %w", err)
			}
		} else {
			fmt.Printf("%sStarting K3s cluster installation...\n", utils.PrefixInfo)
			// Expand the path configured inside Viper
			expandedKubeconfig := config.ExpandPath(config.GlobalConfig.Kubernetes.Kubeconfig)
			if err := installer.InstallK3s(expandedKubeconfig); err != nil {
				return fmt.Errorf("K3s installation failed: %w", err)
			}
		}

		fmt.Printf("%sKubernetes cluster is fully set up and configured.\n", utils.PrefixOK)
		fmt.Printf("%sRun '%s' to deploy the observability stack!\n", utils.PrefixReady, utils.ColorBold+"stackpulse deploy observability"+utils.ColorReset)
		return nil
	},
}

func promptYesNo(question string) (bool, error) {
	fmt.Printf("%s [y/N]: ", question)
	var response string
	// Scanln reads space-separated values; an empty newline returns an error
	_, err := fmt.Scanln(&response)
	if err != nil {
		return false, nil
	}
	response = strings.ToLower(strings.TrimSpace(response))
	if response == "y" || response == "yes" {
		return true, nil
	}
	return false, nil
}

func init() {
	k8sCmd.Flags().StringVar(&setupType, "type", "k3s", "Type of Kubernetes cluster to setup ('k3s' or 'minikube')")
	k8sCmd.Flags().BoolVarP(&setupYes, "yes", "y", false, "Force automatic yes to prompts (non-interactive mode)")
	setupCmd.AddCommand(k8sCmd)
	RootCmd.AddCommand(setupCmd)
}
