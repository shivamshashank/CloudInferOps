package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var bootstrapDryRun bool

var bootstrapCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "Install Kubernetes and deploy the full observability stack",
	Long: `The recommended command for a fresh setup.

It performs the following actions:
1. Checks for an existing Kubernetes cluster.
2. If no cluster is found, it prompts to install one for you via kubeadm.
3. Deploys the complete observability stack.`,
	PreRunE: func(_ *cobra.Command, _ []string) error {
		if !hasRootPrivileges() {
			return fmt.Errorf("the 'bootstrap' command requires root privileges. Please run with sudo")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// The bootstrap command now just calls the same logic as 'deploy platform'
		deployDryRun = bootstrapDryRun // Ensure the dry-run flag is respected
		return runDeployObservability(cmd, args)
	},
}

func init() {
	bootstrapCmd.Flags().BoolVar(&bootstrapDryRun, "dry-run", false, "Initialize files and output manifests without committing changes")
	RootCmd.AddCommand(bootstrapCmd)
}
