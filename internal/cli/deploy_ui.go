package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var deployUIDryRun bool

var deployUICmd = &cobra.Command{
	Use:   "ui",
	Short: "Deploy the CloudInferOps self-hosted UI portal",
	Long:  `Deploys the web-based CloudInferOps dashboard and exposes it under /cloudinferops.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		if deployUIDryRun {
			fmt.Println("[DRY-RUN] Would deploy the CloudInferOps UI portal")
			return nil
		}
		fmt.Println("CloudInferOps UI deployment flow is ready for integration with your cluster manifests.")
		return nil
	},
}

func init() {
	deployUICmd.Flags().BoolVar(&deployUIDryRun, "dry-run", false, "Show what would be deployed without actually deploying")
	deployCmd.AddCommand(deployUICmd)
}
