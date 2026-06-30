package cli

import (
	"fmt"
	"os"

	"github.com/shivamshashank/StackPulse/internal/config"
	"github.com/shivamshashank/StackPulse/internal/score"
	"github.com/spf13/cobra"
)

var (
	scoreNamespace string
	scoreJSON      bool
)

var scoreCmd = &cobra.Command{
	Use:   "score",
	Short: "Generate a Kubernetes workload reliability score",
	Long:  `Scores workload reliability from pod health, deployment readiness, restarts, probe coverage, resource limits, SLO coverage, and security findings.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		engine := score.NewEngine(resolveScoreNamespace())
		report, err := engine.Calculate()
		if err != nil {
			return err
		}

		if scoreJSON {
			out, err := report.JSON()
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintln(os.Stdout, string(out))
			return nil
		}

		report.Print(os.Stdout)
		return nil
	},
}

func init() {
	scoreCmd.Flags().StringVarP(&scoreNamespace, "namespace", "n", "", "Kubernetes namespace to score")
	scoreCmd.Flags().BoolVar(&scoreJSON, "json", false, "Print machine-readable JSON output")
	RootCmd.AddCommand(scoreCmd)
}

func resolveScoreNamespace() string {
	if scoreNamespace != "" {
		return scoreNamespace
	}

	if err := config.InitConfig(false); err != nil {
		config.GlobalConfig = config.DefaultConfig()
	}
	if config.GlobalConfig.Namespace != "" {
		return config.GlobalConfig.Namespace
	}
	return "default"
}
