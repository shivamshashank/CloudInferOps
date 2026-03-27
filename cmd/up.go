package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yourusername/stackpulse/internal/aws"
	"github.com/yourusername/stackpulse/internal/k8s"
)

var region string
var clusterName string

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Deploy the StackPulse observability stack",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := aws.CheckCredentials(); err != nil {
			return err
		}

		// Ensure the target cluster actually exists before trying to deploy to it.
		exists, err := aws.CheckClusterExists(region, clusterName)
		if err != nil {
			return fmt.Errorf("failed during cluster existence check: %w", err)
		}
		if !exists {
			return fmt.Errorf("cluster '%s' does not exist in region %s. Please provide an existing EKS cluster", clusterName, region)
		}

		fmt.Printf("Deploying StackPulse observability stack to cluster %s in region %s...\n", clusterName, region)
		if err := aws.UpdateKubeconfig(region, clusterName); err != nil {
			return fmt.Errorf("failed to connect to cluster: %w", err)
		}

		k8s.Deploy(clusterName)
		fmt.Println("Deployment complete!")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(upCmd)
	upCmd.Flags().StringVarP(&region, "region", "r", "", "AWS region for deployment")
	upCmd.Flags().StringVarP(&clusterName, "cluster-name", "c", "", "Name for the EKS cluster")
	upCmd.MarkFlagRequired("region")
	upCmd.MarkFlagRequired("cluster-name")
}
