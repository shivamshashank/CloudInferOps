package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Dynamic build variables set during compilation via -ldflags
var (
	Version   = "v0.1.3"
	Commit    = "unknown"
	BuildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of StackPulse",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("🚀 StackPulse version: %s\n", Version)
		fmt.Printf("🔑 Commit: %s\n", Commit)
		fmt.Printf("📅 Build date: %s\n", BuildDate)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
