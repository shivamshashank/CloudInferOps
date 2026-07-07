package cli

import (
	"fmt"
	"os/exec"

	"github.com/shivamshashank/CloudInferOps/internal/config"
	"github.com/shivamshashank/CloudInferOps/internal/inference"
	"github.com/shivamshashank/CloudInferOps/internal/observability"
	"github.com/shivamshashank/CloudInferOps/internal/ui"
	"github.com/shivamshashank/CloudInferOps/internal/utils"
	"github.com/shivamshashank/CloudInferOps/internal/webhook"
	"github.com/spf13/cobra"
)

var (
	deployAllDryRun        bool
	deployAllProvider      string
	deployAllModel         string
	deployAllUIImage       string
	deployAllEnableActions bool
)

var deployAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Deploy all CloudInferOps components",
	Long:  `Deploys the complete stack including observability, inference, self-hosted UI, and webhook handler.`,
	PreRunE: func(_ *cobra.Command, _ []string) error {
		if !hasRootPrivileges() {
			return fmt.Errorf("the 'deploy all' command requires root privileges. Please run with sudo")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. Pre-flight check: Verify cluster reachability and install if missing.
		if err := ensureKubernetes(deployAllDryRun, "deploy all"); err != nil {
			return err
		}

		// 1.5 Pre-flight check: Verify Helm is installed and install if missing
		if _, err := exec.LookPath("helm"); err != nil {
			if deployAllDryRun {
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

		// 3. Deploy Observability
		fmt.Printf("\n%sDeploying the observability stack...\n", utils.PrefixInfo)
		if err := observability.DeployObservability(deployAllDryRun); err != nil {
			return fmt.Errorf("failed to deploy observability stack: %w", err)
		}

		// 4. Deploy Inference
		fmt.Printf("\n%sDeploying the inference stack...\n", utils.PrefixInfo)
		if err := inference.DeployInference(deployAllProvider, deployAllModel, deployAllDryRun); err != nil {
			return fmt.Errorf("failed to deploy inference stack: %w", err)
		}

		// 5. Deploy UI Portal
		fmt.Printf("\n%sDeploying the UI portal...\n", utils.PrefixInfo)
		if err := ui.DeployPortal(deployAllUIImage, deployAllEnableActions, deployAllDryRun); err != nil {
			return fmt.Errorf("failed to deploy UI portal: %w", err)
		}

		// 6. Deploy Webhook Handler
		fmt.Printf("\n%sDeploying the webhook handler...\n", utils.PrefixInfo)
		if err := webhook.DeployWebhookHandler(deployAllDryRun); err != nil {
			return fmt.Errorf("failed to deploy webhook handler: %w", err)
		}

		fmt.Printf("\n%sAll components successfully deployed!\n", utils.PrefixOK)
		return nil
	},
}

func init() {
	deployAllCmd.Flags().BoolVar(&deployAllDryRun, "dry-run", false, "Show what would be deployed without actually deploying")
	deployAllCmd.Flags().StringVar(&deployAllProvider, "provider", "ollama", "Inference model provider (e.g. ollama, vllm)")
	deployAllCmd.Flags().StringVar(&deployAllModel, "model", "llama3", "Model name to deploy")
	deployAllCmd.Flags().StringVar(&deployAllUIImage, "image", "ghcr.io/shivamshashank/cloudinferops-ui:latest", "Portal container image")
	deployAllCmd.Flags().BoolVar(&deployAllEnableActions, "enable-actions", false, "Enable guarded cluster write actions in the portal")

	deployCmd.AddCommand(deployAllCmd)
}
