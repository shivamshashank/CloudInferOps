package cli

import (
	"github.com/shivamshashank/CloudInferOps/internal/ui"
	"github.com/spf13/cobra"
)

var deployUIDryRun bool
var deployUIImage string
var deployUIEnableActions bool

var deployUICmd = &cobra.Command{
	Use:   "ui",
	Short: "Deploy the CloudInferOps self-hosted UI portal",
	Long:  `Deploys the web-based CloudInferOps dashboard and exposes it under /cloudinferops.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		return ui.DeployPortal(deployUIImage, deployUIEnableActions, deployUIDryRun)
	},
}

func init() {
	deployUICmd.Flags().BoolVar(&deployUIDryRun, "dry-run", false, "Show what would be deployed without actually deploying")
	deployUICmd.Flags().StringVar(&deployUIImage, "image", "ghcr.io/shivamshashank/cloudinferops-ui:latest", "Portal container image")
	deployUICmd.Flags().BoolVar(&deployUIEnableActions, "enable-actions", false, "Enable guarded cluster write actions in the portal")
	deployCmd.AddCommand(deployUICmd)
}
