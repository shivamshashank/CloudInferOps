package cli

import (
	"fmt"

	"github.com/shivamshashank/StackPulse/internal/config"
	"github.com/shivamshashank/StackPulse/internal/doctor"
	"github.com/shivamshashank/StackPulse/internal/observability"
	"github.com/shivamshashank/StackPulse/internal/utils"
	"github.com/shivamshashank/StackPulse/internal/webhook"
	"github.com/spf13/cobra"
)

var deployDryRun bool

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
			fmt.Printf("%sKubernetes cluster not detected.\n", utils.PrefixError)
			fmt.Printf("%sPlease ensure a local cluster is running (Docker Desktop, Kind, or Minikube), or run:\n", utils.PrefixInfo)
			fmt.Printf("    %s\n", utils.ColorBold+"stackpulse setup k8s"+utils.ColorReset)
			return fmt.Errorf("Kubernetes cluster unreachable")
		}

		// 2. Load configuration (fallback on defaults if not initialized)
		if err := config.InitConfig(false); err != nil {
			fmt.Printf("%sConfiguration file not found. Deploying with default settings...\n", utils.PrefixInfo)
			fmt.Printf("%sRun '%s' to configure custom namespaces and metrics.\n\n", utils.PrefixInfo, utils.ColorBold+"stackpulse init"+utils.ColorReset)
			config.GlobalConfig = config.DefaultConfig()
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
			fmt.Printf("%sPlease ensure a local cluster is running (Docker Desktop, Kind, or Minikube), or run:\n", utils.PrefixInfo)
			fmt.Printf("    %s\n", utils.ColorBold+"stackpulse setup k8s"+utils.ColorReset)
			return fmt.Errorf("Kubernetes cluster unreachable")
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
	webhookHandlerCmd.Flags().BoolVar(&deployDryRun, "dry-run", false, "Print manifests without executing them")
	deployCmd.AddCommand(observabilityCmd)
	deployCmd.AddCommand(webhookHandlerCmd)
	RootCmd.AddCommand(deployCmd)
}
