package cli

import (
	"github.com/shivamshashank/CloudInferOps/internal/config"
	"github.com/shivamshashank/CloudInferOps/internal/webhook"
	"github.com/spf13/cobra"
)

var deployWebhookHandlerDryRun bool

var deployWebhookHandlerCmd = &cobra.Command{
	Use:   "webhook-handler",
	Short: "Deploy the custom Go webhook incident gateway",
	Long:  `Deploys the custom Go webhook handler service for processing and routing incident/alert payloads to notification channels like Slack/PagerDuty.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		if err := config.InitConfig(false); err != nil {
			config.GlobalConfig = config.DefaultConfig()
		}
		return webhook.DeployWebhookHandler(deployWebhookHandlerDryRun)
	},
}

func init() {
	deployWebhookHandlerCmd.Flags().BoolVar(&deployWebhookHandlerDryRun, "dry-run", false, "Show what would be deployed without actually deploying")
	deployCmd.AddCommand(deployWebhookHandlerCmd)
}
