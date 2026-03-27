package aws

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eks"
)

func UpdateKubeconfig(region, clusterName string) error {
	fmt.Printf("[AWS] Updating local kubeconfig for EKS cluster '%s'...\n", clusterName)
	cmd := exec.Command("aws", "eks", "update-kubeconfig", "--region", region, "--name", clusterName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to update kubeconfig: %w", err)
	}

	return nil
}

// CheckCredentials verifies that AWS credentials can be resolved from the environment.
func CheckCredentials() error {
	fmt.Println("Checking for AWS credentials...")
	sess, err := session.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create AWS session: %w", err)
	}

	if _, err := sess.Config.Credentials.Get(); err != nil {
		return fmt.Errorf("could not resolve AWS credentials. Please run 'aws configure' or set up credentials via environment variables or an IAM role")
	}

	fmt.Println("✓ AWS credentials found.")
	return nil
}

// CheckClusterExists checks if an EKS cluster with the given name already exists in the specified region.
func CheckClusterExists(region, clusterName string) (bool, error) {
	fmt.Printf("--> Checking if EKS cluster '%s' already exists in region %s...\n", clusterName, region)

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return false, fmt.Errorf("failed to create AWS session for cluster check: %w", err)
	}

	eksSvc := eks.New(sess)

	_, err = eksSvc.DescribeCluster(&eks.DescribeClusterInput{
		Name: aws.String(clusterName),
	})

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == eks.ErrCodeResourceNotFoundException {
			return false, nil // Cluster does not exist, which is the desired state for provisioning.
		}
		return false, fmt.Errorf("error describing EKS cluster: %w", err) // Another error occurred.
	}

	return true, nil // No error means the cluster was found.
}