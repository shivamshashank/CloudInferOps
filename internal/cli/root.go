package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "stackpulse",
	Short: "🚀 StackPulse is a one-command observability platform for Kubernetes and Linux VMs",
	Long: `🚀 StackPulse is a Go-based DevOps/SRE CLI that detects your environment,
validates Kubernetes readiness, installs lightweight Kubernetes when needed, and
deploys a production-style observability stack with metrics, logs, traces,
dashboards, alerts, and incident webhooks.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
