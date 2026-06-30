package cli

import (
	"fmt"
	"os"

	"github.com/shivamshashank/StackPulse/internal/config"
	"github.com/shivamshashank/StackPulse/internal/security"
	"github.com/spf13/cobra"
)

var (
	securityNamespace string
	securityJSON      bool
)

var securityCmd = &cobra.Command{
	Use:   "security",
	Short: "Scan Kubernetes workloads for production-readiness security risks",
	Long:  `Runs Kubernetes security checks and optional Trivy/kube-score integrations for privileged containers, missing resource limits, mutable image tags, public exposure, and high-severity vulnerabilities.`,
}

var securityScanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Run a Kubernetes security scan",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSecurityReport()
	},
}

var securityReportCmd = &cobra.Command{
	Use:   "report",
	Short: "Print a Kubernetes security report",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSecurityReport()
	},
}

func init() {
	securityCmd.PersistentFlags().StringVarP(&securityNamespace, "namespace", "n", "", "Kubernetes namespace to scan")
	securityCmd.PersistentFlags().BoolVar(&securityJSON, "json", false, "Print machine-readable JSON output")
	securityCmd.AddCommand(securityScanCmd)
	securityCmd.AddCommand(securityReportCmd)
	RootCmd.AddCommand(securityCmd)
}

func runSecurityReport() error {
	engine := security.NewEngine(resolveSecurityNamespace())
	report, err := engine.Scan()
	if err != nil {
		return err
	}

	if securityJSON {
		out, err := report.JSON()
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintln(os.Stdout, string(out))
		return nil
	}

	report.Print(os.Stdout)
	return nil
}

func resolveSecurityNamespace() string {
	if securityNamespace != "" {
		return securityNamespace
	}

	if err := config.InitConfig(false); err != nil {
		config.GlobalConfig = config.DefaultConfig()
	}
	if config.GlobalConfig.Namespace != "" {
		return config.GlobalConfig.Namespace
	}
	return "default"
}

func validateSecurityCommands() error {
	for _, child := range []*cobra.Command{securityScanCmd, securityReportCmd} {
		if child.Parent() != securityCmd {
			return fmt.Errorf("%s command is not attached to security", child.Use)
		}
	}
	return nil
}
