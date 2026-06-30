package cli

import (
	"github.com/spf13/cobra"
)

var deployPlatformCmd = &cobra.Command{
	Use:   "platform",
	Short: "Deploy the core CloudInferOps observability and management platform",
	Long: `Deploys the core platform, which includes:
- Prometheus for metrics
- Grafana for dashboards
- Loki for logs
- Tempo for traces
- ArgoCD for GitOps
- Alertmanager for alerts
- OpenTelemetry Collector for telemetry pipelines`,
	RunE: runDeployObservability,
}

// This is the old `deploy observability` command, now acting as a hidden alias.
var deployObservabilityCmd = &cobra.Command{
	Use:        "observability",
	Short:      "Alias for 'deploy platform'",
	Long:       "This command is an alias for 'deploy platform' and will be removed in a future version.",
	RunE:       runDeployObservability, // It can directly call the same function
	Hidden:     true,                   // Hide it from the main help command
	Deprecated: "use 'deploy platform' instead",
}

func init() {
	addDeployObservabilityFlags(deployPlatformCmd)
	addDeployObservabilityFlags(deployObservabilityCmd) // Also add flags to the alias
	deployCmd.AddCommand(deployPlatformCmd)
	deployCmd.AddCommand(deployObservabilityCmd)
}
