package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/shivamshashank/StackPulse/internal/config"
	"github.com/shivamshashank/StackPulse/internal/doctor"
	"github.com/shivamshashank/StackPulse/internal/gitops"
	"github.com/shivamshashank/StackPulse/internal/installer"
	"github.com/shivamshashank/StackPulse/internal/utils"
	"github.com/spf13/cobra"
)

var gitopsCmd = &cobra.Command{
	Use:   "gitops",
	Short: "Manage and view GitOps applications via ArgoCD",
}

var gitopsBootstrapDryRun bool

var gitopsBootstrapCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "Deploy the GitOps-managed observability stack via ArgoCD",
	Long:  `Installs ArgoCD, a local in-cluster Git server, and registers GitOps Applications to continuously sync and manage your observability platform.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. Pre-flight check: Verify cluster reachability early
		_, hasK8s := doctor.CheckK8sCluster()
		if !hasK8s {
			if gitopsBootstrapDryRun {
				return fmt.Errorf("kubernetes cluster unreachable (dry-run bypassed setup)")
			}

			// Prompt the user to install a local cluster or exit so they can bring their own.
			choice, err := promptClusterOption(os.Stdin)
			if err != nil {
				return err
			}

			if choice == "no" {
				fmt.Printf("%sBootstrap cancelled. Install or start Kubernetes, then rerun bootstrap.\n", utils.PrefixWarn)
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
			if gitopsBootstrapDryRun {
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
			config.GlobalConfig = config.DefaultConfig()
		}

		// Always force ArgoCD to true for GitOps bootstrap
		config.GlobalConfig.Observability.ArgoCD = true

		// 3. Trigger GitOps bootstrap orchestrator
		if err := gitops.BootstrapGitOps(gitopsBootstrapDryRun); err != nil {
			return err
		}

		return nil
	},
}

var gitopsStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show sync and health status of all ArgoCD applications",
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, hasK8s := doctor.CheckK8sCluster(); !hasK8s {
			return fmt.Errorf("kubernetes cluster unreachable")
		}

		if err := config.InitConfig(false); err != nil {
			config.GlobalConfig = config.DefaultConfig()
		}

		ns := config.GlobalConfig.Namespace
		if ns == "" {
			ns = "observability"
		}

		fmt.Println()
		fmt.Printf("%s%sGitOps Applications Status%s\n", utils.PrefixInfo, utils.ColorBold, utils.ColorReset)
		fmt.Println("-----------------------------------------------------------------")

		appsOut, _, err := utils.ExecCommand("", "kubectl", "get", "applications", "-n", ns, "-o", "json")
		if err != nil || appsOut == "" || strings.HasPrefix(appsOut, "No resources found") {
			fmt.Printf("%sNo GitOps applications found. Ensure ArgoCD is deployed and managing resources.\n", utils.PrefixWarn)
			fmt.Println("-----------------------------------------------------------------")
			return nil
		}

		var appList struct {
			Items []struct {
				Metadata struct {
					Name string `json:"name"`
				} `json:"metadata"`
				Status struct {
					Sync struct {
						Status string `json:"status"`
					} `json:"sync"`
					Health struct {
						Status string `json:"status"`
					} `json:"health"`
				} `json:"status"`
			} `json:"items"`
		}

		if err := json.Unmarshal([]byte(appsOut), &appList); err != nil {
			fmt.Printf("%sFailed to parse GitOps applications status. Raw output:\n\n", utils.PrefixWarn)
			fmt.Println(appsOut)
			fmt.Println("-----------------------------------------------------------------")
			return nil
		}

		fmt.Printf("%-25s %-10s %-10s\n", "APPLICATION", "STATUS", "HEALTH")

		for _, item := range appList.Items {
			name := item.Metadata.Name
			sync := item.Status.Sync.Status
			if sync == "" {
				sync = "Unknown"
			}
			health := item.Status.Health.Status
			if health == "" {
				health = "Unknown"
			}
			fmt.Printf("%-25s %-10s %-10s\n", name, sync, health)
		}

		fmt.Println("-----------------------------------------------------------------")
		return nil
	},
}

func init() {
	gitopsBootstrapCmd.Flags().BoolVar(&gitopsBootstrapDryRun, "dry-run", false, "Initialize files and output manifests without committing changes")
	gitopsCmd.AddCommand(gitopsBootstrapCmd)
	gitopsCmd.AddCommand(gitopsStatusCmd)
	RootCmd.AddCommand(gitopsCmd)
}
