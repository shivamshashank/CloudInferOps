package cli

import (
	"fmt"
	"os"

	"github.com/shivamshashank/StackPulse/internal/config"
	"github.com/shivamshashank/StackPulse/internal/diagnose"
	"github.com/spf13/cobra"
)

var diagnoseNamespace string

var diagnoseCmd = &cobra.Command{
	Use:   "diagnose",
	Short: "Diagnose Kubernetes incidents from events, status, logs, and node signals",
	Long:  `Collects Kubernetes evidence and prints an actionable incident diagnosis for pods, deployments, services, or the whole cluster.`,
}

var diagnosePodCmd = &cobra.Command{
	Use:   "pod POD_NAME",
	Short: "Diagnose a Kubernetes pod",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		engine := diagnose.NewEngine(resolveDiagnoseNamespace())
		report, err := engine.DiagnosePod(args[0])
		if err != nil {
			return err
		}
		report.Print(os.Stdout)
		return nil
	},
}

var diagnoseDeploymentCmd = &cobra.Command{
	Use:   "deployment DEPLOYMENT",
	Short: "Diagnose a Kubernetes deployment",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		engine := diagnose.NewEngine(resolveDiagnoseNamespace())
		report, err := engine.DiagnoseDeployment(args[0])
		if err != nil {
			return err
		}
		report.Print(os.Stdout)
		return nil
	},
}

var diagnoseServiceCmd = &cobra.Command{
	Use:   "service SERVICE",
	Short: "Diagnose a Kubernetes service",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		engine := diagnose.NewEngine(resolveDiagnoseNamespace())
		report, err := engine.DiagnoseService(args[0])
		if err != nil {
			return err
		}
		report.Print(os.Stdout)
		return nil
	},
}

var diagnoseClusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Diagnose cluster-wide Kubernetes health",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		engine := diagnose.NewEngine(resolveDiagnoseNamespace())
		report, err := engine.DiagnoseCluster()
		if err != nil {
			return err
		}
		report.Print(os.Stdout)
		return nil
	},
}

func init() {
	diagnoseCmd.PersistentFlags().StringVarP(&diagnoseNamespace, "namespace", "n", "", "Kubernetes namespace for namespaced resources")
	diagnoseCmd.AddCommand(diagnosePodCmd)
	diagnoseCmd.AddCommand(diagnoseDeploymentCmd)
	diagnoseCmd.AddCommand(diagnoseServiceCmd)
	diagnoseCmd.AddCommand(diagnoseClusterCmd)
	RootCmd.AddCommand(diagnoseCmd)
}

func resolveDiagnoseNamespace() string {
	if diagnoseNamespace != "" {
		return diagnoseNamespace
	}

	if err := config.InitConfig(false); err != nil {
		config.GlobalConfig = config.DefaultConfig()
	}
	if config.GlobalConfig.Namespace != "" {
		return config.GlobalConfig.Namespace
	}
	return "default"
}

func validateDiagnoseCommands() error {
	for _, child := range []*cobra.Command{diagnosePodCmd, diagnoseDeploymentCmd, diagnoseServiceCmd, diagnoseClusterCmd} {
		if child.Parent() != diagnoseCmd {
			return fmt.Errorf("%s command is not attached to diagnose", child.Use)
		}
	}
	return nil
}
