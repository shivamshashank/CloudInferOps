package cli

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/shivamshashank/StackPulse/internal/config"
	"github.com/shivamshashank/StackPulse/internal/doctor"
	"github.com/shivamshashank/StackPulse/internal/installer"
	"github.com/shivamshashank/StackPulse/internal/observability"
	"github.com/shivamshashank/StackPulse/internal/utils"
	"github.com/shivamshashank/StackPulse/internal/webhook"
	"github.com/spf13/cobra"
)

var (
	deployDryRun bool
	deployHA     bool
)

// deployCmd represents the parent deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy StackPulse components onto Kubernetes",
	Long:  `Parent command for provisioning observability pipelines and gateway services in your active cluster.`,
}

// observabilityCmd represents the deploy observability subcommand
var observabilityCmd = &cobra.Command{
	Use:   "observability",
	Short: "Deploy the complete observability stack",
	Long:  `Deploys Prometheus, Grafana, Loki, Tempo, OpenTelemetry, alert rules, and standard Grafana dashboards onto Kubernetes.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. Pre-flight check: Verify cluster reachability early
		_, hasK8s := doctor.CheckK8sCluster()
		if !hasK8s {
			if deployDryRun {
				return fmt.Errorf("kubernetes cluster unreachable (dry-run bypassed setup)")
			}

			// Prompt the user to install a local cluster or exit so they can bring their own.
			choice, err := promptClusterOption(os.Stdin)
			if err != nil {
				return err
			}

			if choice == "no" {
				fmt.Printf("%sDeployment cancelled. Install or start Kubernetes (k3s, minikube, kind, Docker Desktop, or another cluster), then rerun deploy.\n", utils.PrefixWarn)
				return fmt.Errorf("kubernetes cluster unreachable")
			}

			switch choice {
			case "kind":
				if _, err := exec.LookPath("docker"); err != nil {
					fmt.Printf("%sDocker is required for kind but was not found. Installing Docker now...\n", utils.PrefixInfo)
					if dockerErr := installer.InstallDocker(); dockerErr != nil {
						return fmt.Errorf("failed to install Docker (required by kind): %w", dockerErr)
					}
				}
				if _, err := exec.LookPath("kind"); err != nil {
					fmt.Printf("%s'kind' not found. Installing it now...\n", utils.PrefixInfo)
					if installErr := installer.DownloadKindBinary(); installErr != nil {
						return fmt.Errorf("failed to install kind: %w", installErr)
					}
				}
				if _, err := exec.LookPath("kubectl"); err != nil {
					fmt.Printf("%s'kubectl' not found. Installing it now...\n", utils.PrefixInfo)
					if installErr := installer.DownloadKubectlBinary(); installErr != nil {
						return fmt.Errorf("failed to install kubectl: %w", installErr)
					}
				}
				fmt.Printf("%sStarting kind cluster installation...\n", utils.PrefixInfo)
				if err := installer.InstallKind(); err != nil {
					return fmt.Errorf("kind installation failed: %w", err)
				}
			case "minikube":
				if _, err := exec.LookPath("docker"); err != nil {
					fmt.Printf("%sDocker is required for minikube but was not found. Installing Docker now...\n", utils.PrefixInfo)
					if dockerErr := installer.InstallDocker(); dockerErr != nil {
						return fmt.Errorf("failed to install Docker (required by minikube): %w", dockerErr)
					}
				}
				if _, err := exec.LookPath("minikube"); err != nil {
					fmt.Printf("%s'minikube' not found. Installing it now...\n", utils.PrefixInfo)
					if installErr := installer.DownloadMinikubeBinary(); installErr != nil {
						return fmt.Errorf("failed to install minikube: %w", installErr)
					}
				}
				if _, err := exec.LookPath("kubectl"); err != nil {
					fmt.Printf("%s'kubectl' not found. Installing it now...\n", utils.PrefixInfo)
					if installErr := installer.DownloadKubectlBinary(); installErr != nil {
						return fmt.Errorf("failed to install kubectl: %w", installErr)
					}
				}
				fmt.Printf("%sStarting Minikube cluster installation...\n", utils.PrefixInfo)
				if err := installer.InstallMinikube(); err != nil {
					return fmt.Errorf("minikube installation failed: %w", err)
				}
			case "k3s":
				if runtime.GOOS != "linux" {
					return fmt.Errorf("k3s installation is only supported on Linux")
				}
				fmt.Printf("%sStarting K3s cluster installation...\n", utils.PrefixInfo)
				if err := config.InitConfig(true); err != nil {
					return fmt.Errorf("failed to load configuration: %w", err)
				}
				expandedKubeconfig := config.ExpandPath(config.GlobalConfig.Kubernetes.Kubeconfig)
				if err := installer.InstallK3s(expandedKubeconfig); err != nil {
					return fmt.Errorf("K3s installation failed: %w", err)
				}
			}

			if _, hasK8s = doctor.CheckK8sCluster(); !hasK8s {
				return fmt.Errorf("kubernetes cluster unreachable after setup")
			}

			fmt.Printf("%sKubernetes cluster setup completed successfully.\n\n", utils.PrefixOK)
		}

		// 1.5 Pre-flight check: Verify Helm is installed and install if missing
		if _, err := exec.LookPath("helm"); err != nil {
			if deployDryRun {
				fmt.Printf("%s[DRY-RUN] Would install Helm\n", utils.PrefixInfo)
			} else {
				fmt.Printf("%sHelm is required but was not found. Installing Helm now...\n", utils.PrefixInfo)
				if _, stderr, installErr := utils.ExecCommandInteractive("", "bash", "-c", "curl -fsSL https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash"); installErr != nil {
					return fmt.Errorf("failed to install Helm: %w (stderr: %s)", installErr, stderr)
				}
				fmt.Printf("%sHelm installed successfully.\n", utils.PrefixOK)
			}
		}

		// 2. Load configuration (fallback on defaults if not initialized)
		if err := config.InitConfig(false); err != nil {
			fmt.Printf("%sConfiguration file not found. Deploying with default settings...\n", utils.PrefixInfo)
			fmt.Printf("%sRun '%s' to configure custom namespaces and metrics.\n\n", utils.PrefixInfo, utils.ColorBold+"sudo stackpulse init"+utils.ColorReset)
			config.GlobalConfig = config.DefaultConfig()
		}

		if deployHA {
			config.GlobalConfig.Observability.Thanos = true
		}

		// 3. Trigger observability stack orchestrator
		if err := observability.DeployObservability(deployDryRun); err != nil {
			return err
		}

		return nil
	},
}

// webhookHandlerCmd represents the deploy webhook-handler subcommand
var webhookHandlerCmd = &cobra.Command{
	Use:   "webhook-handler",
	Short: "Deploy the custom Go webhook incident gateway",
	Long:  `Deploys the custom SRE alert processor which intercepts Alertmanager webhooks and routes incidents to Slack and PagerDuty.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. Pre-flight check: Verify cluster reachability early
		_, hasK8s := doctor.CheckK8sCluster()
		if !hasK8s {
			fmt.Printf("%sKubernetes cluster not detected.\n", utils.PrefixError)
			fmt.Printf("%sPlease ensure a local cluster is running (Docker Desktop, Kind, or Minikube) and rerun this command.\n", utils.PrefixInfo)
			return fmt.Errorf("kubernetes cluster unreachable")
		}

		// 2. Load configuration (fallback on defaults if not initialized)
		if err := config.InitConfig(false); err != nil {
			config.GlobalConfig = config.DefaultConfig()
		}

		// 3. Trigger webhook deployment engine
		if err := webhook.DeployWebhookHandler(deployDryRun); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	observabilityCmd.Flags().BoolVar(&deployDryRun, "dry-run", false, "Print Helm commands and manifest applications without executing them")
	observabilityCmd.Flags().BoolVar(&deployHA, "ha", false, "Enable high-availability Thanos cluster deployment")
	webhookHandlerCmd.Flags().BoolVar(&deployDryRun, "dry-run", false, "Print manifests without executing them")
	deployCmd.AddCommand(observabilityCmd)
	deployCmd.AddCommand(webhookHandlerCmd)
	RootCmd.AddCommand(deployCmd)
}

func promptClusterOption(reader io.Reader) (string, error) {
	fmt.Println("Kubernetes cluster not detected.")
	fmt.Println("Would you like StackPulse to install a local Kubernetes cluster for you?")
	fmt.Println("  1. kind (Docker-based local Kubernetes)")
	fmt.Println("  2. k3s (Lightweight Linux-only Kubernetes)")
	fmt.Println("  3. minikube (Cross-platform local Kubernetes)")
	fmt.Println("  4. No, I will install or manage Kubernetes myself")
	fmt.Print("Choose an option [1-4] (default: 4): ")

	var response string
	_, _ = fmt.Fscanln(reader, &response)
	response = strings.TrimSpace(response)

	switch response {
	case "1":
		return "kind", nil
	case "2":
		return "k3s", nil
	case "3":
		return "minikube", nil
	case "4", "":
		return "no", nil
	default:
		fmt.Printf("%sInvalid choice '%s'. Defaulting to exit.\n", utils.PrefixWarn, response)
		return "no", nil
	}
}
