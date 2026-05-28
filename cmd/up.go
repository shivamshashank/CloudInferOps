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

		// First check if an EC2 instance with the same Name tag exists
		isEC2, ip, instanceID, err := aws.CheckEC2InstanceExists(region, clusterName)
		if err != nil {
			return fmt.Errorf("failed during EC2 instance check: %w", err)
		}

		if isEC2 {
			fmt.Printf("\nFound EC2 instance '%s' with IP: %s (ID: %s)\n", clusterName, ip, instanceID)

			fmt.Println("\nDeploying StackPulse observability stack to EC2 instance via SSM...")
			ssmClient, err := aws.NewSSMClient(region)
			if err != nil {
				return fmt.Errorf("failed to create SSM client: %w", err)
			}
			k8s.DeployEC2(ssmClient, instanceID, ip)
		} else {
			// If not EC2, check if an EKS cluster with this name exists
			fmt.Printf("\nEC2 instance '%s' not found. Checking for an EKS cluster instead...\n", clusterName)
			isEKS, err := aws.CheckClusterExists(region, clusterName)
			if err != nil {
				return fmt.Errorf("failed during EKS cluster existence check: %w", err)
			}

			if isEKS {
				fmt.Printf("\nDeploying StackPulse observability stack to EKS cluster %s in region %s...\n", clusterName, region)
				if err := aws.UpdateKubeconfig(region, clusterName); err != nil {
					return fmt.Errorf("failed to connect to EKS cluster: %w", err)
				}
				k8s.DeployEKS(clusterName)
			} else {
				return fmt.Errorf("neither EC2 instance nor EKS cluster named '%s' exists in region %s", clusterName, region)
			}
		}

		fmt.Println("\n🚀 Deployment sequence finished!")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(upCmd)
	upCmd.Flags().StringVarP(&region, "region", "r", "", "AWS region for deployment")
	upCmd.Flags().StringVarP(&clusterName, "cluster-name", "c", "", "Name for the EKS cluster or EC2 instance")
	upCmd.MarkFlagRequired("region")
	upCmd.MarkFlagRequired("cluster-name")
}
