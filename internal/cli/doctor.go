package cli

import (
	"fmt"
	"os"

	"github.com/shivamshashank/StackPulse/internal/config"
	"github.com/shivamshashank/StackPulse/internal/doctor"
	"github.com/shivamshashank/StackPulse/internal/utils"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check system prerequisites and Kubernetes readiness",
	Long:  `Runs diagnostic checks on your OS, hardware resources, network connectivity, and external tools (kubectl, helm, docker, and active Kubernetes context).`,
	Run: func(cmd *cobra.Command, args []string) {
		// Attempt to load config if it exists
		err := config.InitConfig(false)
		if err != nil {
			// Configuration doesn't exist, we will warn the user but still proceed with defaults
			fmt.Printf("%sConfiguration file not found. Running doctor with default settings...\n", utils.PrefixInfo)
			fmt.Printf("%sRun 'stackpulse init' to create a custom configuration file.\n\n", utils.PrefixInfo)
			
			// Setup default Kubeconfig path env
			defaultKubeconfig := config.ExpandPath("~/.kube/config")
			os.Setenv("KUBECONFIG", defaultKubeconfig)
		} else {
			// Config found, load Kubeconfig into env for kubectl execution
			if config.GlobalConfig.Kubernetes.Kubeconfig != "" {
				os.Setenv("KUBECONFIG", config.GlobalConfig.Kubernetes.Kubeconfig)
			}
		}

		// Run diagnostics
		report := doctor.RunDoctor()
		report.Print()
	},
}

func init() {
	RootCmd.AddCommand(doctorCmd)
}
