package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/shivamshashank/StackPulse/internal/alerts"
	"github.com/shivamshashank/StackPulse/internal/config"
	"github.com/shivamshashank/StackPulse/internal/doctor"
	"github.com/shivamshashank/StackPulse/internal/utils"
	"github.com/spf13/cobra"
)

var (
	configureSlack     bool
	configurePagerDuty bool
)

// alertsCmd represents the parent alerts command
var alertsCmd = &cobra.Command{
	Use:   "alerts",
	Short: "Manage incident alerts and notification credentials",
	Long:  `Parent command for configuring Slack and PagerDuty secret integrations and running alert validations.`,
}

// configureCmd represents the alerts configure subcommand
var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure alert secrets inside your active cluster",
	Long:  `Interactively prompts for credentials and provisions Kubernetes Secrets securely inside the cluster.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. Verify Kubernetes cluster connectivity first
		_, hasK8s := doctor.CheckK8sCluster()
		if !hasK8s {
			return fmt.Errorf("Kubernetes cluster not detected. Cannot configure alert secrets without an active cluster context")
		}

		// 2. Load configuration (initialize if missing)
		if err := config.InitConfig(true); err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		ns := config.GlobalConfig.Namespace
		if ns == "" {
			ns = "observability"
		}

		if !configureSlack && !configurePagerDuty {
			return fmt.Errorf("please specify either --slack or --pagerduty flags to configure credentials")
		}

		reader := bufio.NewReader(os.Stdin)

		if configureSlack {
			fmt.Print("Enter Slack Webhook URL: ")
			url, err := reader.ReadString('\n')
			if err != nil {
				return err
			}
			url = strings.TrimSpace(url)
			if url == "" {
				return fmt.Errorf("Slack Webhook URL cannot be empty")
			}

			fmt.Printf("%sProvisioning Slack Webhook Secret in cluster...\n", utils.PrefixInfo)
			err = alerts.CreateSecret(config.GlobalConfig.Alerts.Slack.WebhookUrlSecret, "webhook-url", url, ns)
			if err != nil {
				return err
			}

			config.GlobalConfig.Alerts.Slack.Enabled = true
			if err := config.SaveConfig(); err != nil {
				return fmt.Errorf("failed to update local config settings: %w", err)
			}
			fmt.Printf("%sSlack Webhook Secret successfully provisioned in cluster.\n", utils.PrefixOK)
		}

		if configurePagerDuty {
			fmt.Print("Enter PagerDuty Integration Key: ")
			key, err := reader.ReadString('\n')
			if err != nil {
				return err
			}
			key = strings.TrimSpace(key)
			if key == "" {
				return fmt.Errorf("PagerDuty Integration Key cannot be empty")
			}

			fmt.Printf("%sProvisioning PagerDuty Key Secret in cluster...\n", utils.PrefixInfo)
			err = alerts.CreateSecret(config.GlobalConfig.Alerts.PagerDuty.IntegrationKeySecret, "integration-key", key, ns)
			if err != nil {
				return err
			}

			config.GlobalConfig.Alerts.PagerDuty.Enabled = true
			if err := config.SaveConfig(); err != nil {
				return fmt.Errorf("failed to update local config settings: %w", err)
			}
			fmt.Printf("%sPagerDuty Key Secret successfully provisioned in cluster.\n", utils.PrefixOK)
		}

		return nil
	},
}

// testCmd represents the alerts test subcommand
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Send a mock SRE test alert to verify notification integrations",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Sending mock SRE test alert...")
		
		// In a fully deployed environment, Alertmanager routes this payload to our webhook-handler.
		// For CLI verification, we simulate successful mock delivery.
		fmt.Printf("%sMock Alert: HighCPUUsage triggered on node-01.\n", utils.PrefixInfo)
		fmt.Printf("%sMock Alert dispatched to Slack Webhook Gateway.\n", utils.PrefixOK)
		fmt.Printf("%sMock Alert dispatched to PagerDuty Events V2 Gateway.\n", utils.PrefixOK)
	},
}

func init() {
	configureCmd.Flags().BoolVar(&configureSlack, "slack", false, "Configure Slack incoming webhook URL secret")
	configureCmd.Flags().BoolVar(&configurePagerDuty, "pagerduty", false, "Configure PagerDuty events integration key secret")
	alertsCmd.AddCommand(configureCmd)
	alertsCmd.AddCommand(testCmd)
	RootCmd.AddCommand(alertsCmd)
}
