package cmd

import (
	"github.com/spf13/cobra"
	"github.com/yourusername/stackpulse/internal/integrations"
)

var webhookURL string
var routingKey string

var integrateCmd = &cobra.Command{
	Use:   "integrate",
	Short: "Configure alert integrations like Slack and PagerDuty",
}

var integrateSlackCmd = &cobra.Command{
	Use:   "slack",
	Short: "Integrate with Slack",
	Run: func(cmd *cobra.Command, args []string) {
		integrations.ConfigureSlack(webhookURL)
	},
}

var integratePagerDutyCmd = &cobra.Command{
	Use:   "pagerduty",
	Short: "Integrate with PagerDuty",
	Run: func(cmd *cobra.Command, args []string) {
		integrations.ConfigurePagerDuty(routingKey)
	},
}

func init() {
	rootCmd.AddCommand(integrateCmd)
	
	integrateCmd.AddCommand(integrateSlackCmd)
	integrateSlackCmd.Flags().StringVarP(&webhookURL, "webhook-url", "w", "", "Slack Webhook URL")
	integrateSlackCmd.MarkFlagRequired("webhook-url")

	integrateCmd.AddCommand(integratePagerDutyCmd)
	integratePagerDutyCmd.Flags().StringVarP(&routingKey, "routing-key", "r", "", "PagerDuty Routing Key")
	integratePagerDutyCmd.MarkFlagRequired("routing-key")
}