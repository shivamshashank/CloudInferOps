package cli

import (
	"github.com/shivamshashank/CloudInferOps/internal/inference"
	"github.com/spf13/cobra"
)

var (
	inferenceProvider string
	inferenceModel    string
	inferenceDryRun   bool
)

var deployInferenceCmd = &cobra.Command{
	Use:   "inference",
	Short: "Deploy an AI/ML model inference service",
	Long: `Deploys an inference service and gateway for serving AI/ML models.

You can specify the model provider and the model to be deployed.
Example: cloudinferops deploy inference --provider ollama --model llama3`,
	RunE: func(_ *cobra.Command, _ []string) error {
		return inference.DeployInference(inferenceProvider, inferenceModel, inferenceDryRun)
	},
}

func init() {
	deployInferenceCmd.Flags().StringVar(&inferenceProvider, "provider", "ollama", "Inference model provider (e.g. ollama, vllm)")
	deployInferenceCmd.Flags().StringVar(&inferenceModel, "model", "llama3", "Model name to deploy")
	deployInferenceCmd.Flags().BoolVar(&inferenceDryRun, "dry-run", false, "Show what would be deployed without actually deploying")
	deployCmd.AddCommand(deployInferenceCmd)
}
