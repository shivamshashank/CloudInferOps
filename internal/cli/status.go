package cli

import (
	"fmt"

	"github.com/shivamshashank/StackPulse/internal/config"
	"github.com/shivamshashank/StackPulse/internal/doctor"
	"github.com/shivamshashank/StackPulse/internal/observability"
	"github.com/shivamshashank/StackPulse/internal/utils"
	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check the status and retrieve credentials of the deployed stack",
	Long:  `Queries active Kubernetes pods, collects components liveness checkpoints, and decrypts ingress/credentials dashboards.`,
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

		// 3. Trigger observability status dashboard fetcher
		if err := observability.PrintStatus(); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	RootCmd.AddCommand(statusCmd)
}
