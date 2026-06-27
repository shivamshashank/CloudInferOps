package cli

import (
	"fmt"

	"github.com/shivamshashank/CloudInferOps/internal/config"
	"github.com/shivamshashank/CloudInferOps/internal/doctor"
	"github.com/shivamshashank/CloudInferOps/internal/observability"
	"github.com/shivamshashank/CloudInferOps/internal/utils"
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
			fmt.Printf("%sPlease ensure a local cluster is running (Docker Desktop, Kind, or Minikube) and rerun this command.\n", utils.PrefixInfo)
			return fmt.Errorf("kubernetes cluster unreachable")
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
