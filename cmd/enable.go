package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var enableCmd = &cobra.Command{
	Use:   "enable [plugin]",
	Short: "Enable optional integrations (e.g., dynatrace, datadog, newrelic)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		plugin := args[0]
		fmt.Printf("[Plugins] Enabling integration for: %s\n", plugin)
	},
}

func init() {
	rootCmd.AddCommand(enableCmd)
}
