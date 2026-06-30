package cli

import (
	"github.com/spf13/cobra"
)

// deployCmd represents the parent deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy CloudInferOps components onto Kubernetes",
	Long:  `Parent command for provisioning observability pipelines and gateway services in your active cluster.`,
}

func init() {
	RootCmd.AddCommand(deployCmd)
}
