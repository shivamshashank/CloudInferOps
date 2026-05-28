package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestUpCommand(t *testing.T) {
	// 1. Create a buffer to capture the command's output.
	// We assume a `rootCmd` exists in the package, which is standard for Cobra.
	output := &bytes.Buffer{}
	rootCmd.SetOut(output)
	rootCmd.SetErr(output)

	// 2. Set the arguments for the 'up' command, simulating user input.
	testRegion := "us-east-1"
	testCluster := "test-cluster"
	rootCmd.SetArgs([]string{"up", "--region", testRegion, "--cluster-name", testCluster})

	// 3. Execute the root command.
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	// 4. Get the captured output.
	got := output.String()

	// 5. Assert that the output contains all the expected messages.
	expectedMessages := []string{
		"Deploying StackPulse observability stack to cluster test-cluster in region us-east-1...",
		"[AWS] Updating local kubeconfig for EKS cluster 'test-cluster'...",
		"[K8s] Deploying Helm charts and ArgoCD to cluster 'test-cluster'...",
		"Deployment complete!",
	}

	for _, msg := range expectedMessages {
		if !strings.Contains(got, msg) {
			t.Errorf("expected output to contain %q, but got %q", msg, got)
		}
	}

	// 6. Reset flag values to their defaults to avoid polluting other tests.
	region = "eu-west-2"
	clusterName = "stackpulse-cluster"
}
