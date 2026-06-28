package cli

import (
	"fmt"
	"os/exec"

	"github.com/shivamshashank/CloudInferOps/internal/config"
	"github.com/shivamshashank/CloudInferOps/internal/doctor"
	"github.com/shivamshashank/CloudInferOps/internal/observability"
	"github.com/shivamshashank/CloudInferOps/internal/utils"
	"github.com/shivamshashank/CloudInferOps/internal/webhook"
	"github.com/spf13/cobra"
)

var (
	deployDryRun bool
	deployHA     bool
)

// deployCmd represents the parent deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy CloudInferOps components onto Kubernetes",
	Long:  `Parent command for provisioning observability pipelines and gateway services in your active cluster.`,
}

// observabilityCmd represents the deploy observability subcommand
var observabilityCmd = &cobra.Command{
	Use:   "observability",
	Short: "Deploy the complete observability stack",
	Long: `Deploys the observability stack onto an EXISTING Kubernetes cluster.

This command is for users who bring their own cluster. If no cluster is detected, the command will fail with instructions.
To have CloudInferOps install Kubernetes for you, use the 'cloudinferops bootstrap' command instead.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if !hasRootPrivileges() {
			return fmt.Errorf("the 'deploy' command requires root privileges. Please run with sudo")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ensureKubernetes(deployDryRun, "deploy observability"); err != nil {
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
			fmt.Printf("%sRun '%s' to configure custom namespaces and metrics.\n\n", utils.PrefixInfo, utils.ColorBold+"sudo cloudinferops init"+utils.ColorReset)
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
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if !hasRootPrivileges() {
			return fmt.Errorf("the 'deploy webhook-handler' command requires root privileges. Please run with sudo")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. Pre-flight check: Verify cluster reachability early
		_, hasK8s := doctor.CheckK8sCluster()
		if !hasK8s {
			fmt.Printf("%sKubernetes cluster not detected.\n", utils.PrefixError)
			fmt.Printf("%sPlease ensure Kubernetes is running and rerun this command.\n", utils.PrefixInfo)
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
