package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "stackpulse",
	Short: "One-Tap Observability Stack for AWS",
	Long:  "StackPulse is an open-source CLI tool that deploys a complete production-grade observability stack on AWS in a single command.",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
