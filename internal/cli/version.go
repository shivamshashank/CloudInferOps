package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Dynamic build variables set during compilation via -ldflags
var (
	Version   = "v1.0.0"
	Commit    = "unknown"
	BuildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of CloudInferOps",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("🚀 CloudInferOps version: %s\n", Version)
		fmt.Printf("🔑 Commit: %s\n", Commit)
		fmt.Printf("📅 Build date: %s\n", BuildDate)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
