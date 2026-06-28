package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/shivamshashank/CloudInferOps/internal/config"
	"github.com/shivamshashank/CloudInferOps/internal/doctor"
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
	Short: "Alias for 'cloudinferops bootstrap'. Installs Kubernetes and deploys the GitOps stack.",
	Long: `This command is an alias for 'cloudinferops bootstrap'.

It installs Kubernetes if not found, then deploys the full observability stack using GitOps with ArgoCD.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// This is an alias. We just call the main bootstrap command's logic.
		bootstrapCmd.SetArgs(args)
		bootstrapDryRun = gitopsBootstrapDryRun // Sync the dry-run flag
		return bootstrapCmd.RunE(cmd, args)

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
