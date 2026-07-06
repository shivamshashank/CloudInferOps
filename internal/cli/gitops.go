package cli

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/shivamshashank/CloudInferOps/internal/config"
	"github.com/shivamshashank/CloudInferOps/internal/doctor"
	"github.com/shivamshashank/CloudInferOps/internal/gitops"
	"github.com/shivamshashank/CloudInferOps/internal/utils"
	"github.com/spf13/cobra"
)

var gitopsCmd = &cobra.Command{
	Use:   "gitops",
	Short: "Manage and view GitOps applications via ArgoCD",
}

var gitopsBootstrapDryRun bool

var gitopsBootstrapCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "Install Kubernetes and bootstrap the local GitOps (ArgoCD) stack",
	Long: `Performs end-to-end setup of the local GitOps pipeline:
1. Checks for an existing Kubernetes cluster (and prompts to install one via kubeadm if missing).
2. Verifies and installs Helm if missing.
3. Provisions the in-cluster Git server.
4. Generates local GitOps repository structure under ~/.cloudinferops/gitops-repo.
5. Initializes, commits, and pushes configurations to the in-cluster Git server.
6. Deploys Ingress NGINX, ArgoCD, and registers ArgoCD applications for GitOps deployment.`,
	PreRunE: func(_ *cobra.Command, _ []string) error {
		if !hasRootPrivileges() {
			return fmt.Errorf("the 'gitops bootstrap' command requires root privileges. Please run with sudo")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. Pre-flight check: Verify cluster reachability and install if missing.
		if err := ensureKubernetes(gitopsBootstrapDryRun, "gitops bootstrap"); err != nil {
			return err
		}

		// 2. Pre-flight check: Verify Helm is installed and install if missing
		if _, err := exec.LookPath("helm"); err != nil {
			if gitopsBootstrapDryRun {
				fmt.Printf("%s[DRY-RUN] Would install Helm\n", utils.PrefixInfo)
			} else {
				fmt.Printf("%sHelm is required but was not found. Installing Helm now...\n", utils.PrefixInfo)
				if _, stderr, installErr := utils.ExecCommandInteractive("", "bash", "-c", "curl -fsSL https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash"); installErr != nil { // #nosec G204
					return fmt.Errorf("failed to install Helm: %w (stderr: %s)", installErr, stderr)
				}
				fmt.Printf("%sHelm installed successfully.\n", utils.PrefixOK)
			}
		}

		// 3. Load configuration (fallback on defaults if not initialized)
		if err := config.InitConfig(false); err != nil {
			fmt.Printf("%sConfiguration file not found. Bootstrapping GitOps with default settings...\n", utils.PrefixInfo)
			config.GlobalConfig = config.DefaultConfig()
		}

		// 4. Trigger GitOps bootstrap
		return gitops.BootstrapGitOps(gitopsBootstrapDryRun)
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
