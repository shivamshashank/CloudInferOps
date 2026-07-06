package cli

import (
	"fmt"

	"github.com/shivamshashank/CloudInferOps/internal/config"
	"github.com/shivamshashank/CloudInferOps/internal/observability"
	"github.com/shivamshashank/CloudInferOps/internal/utils"
	"github.com/spf13/cobra"
)

var dashboardsDryRun bool

var dashboardsCmd = &cobra.Command{
	Use:   "dashboards",
	Short: "Manage SRE Grafana dashboards",
	Long:  `Parent command for managing, importing, and provisioning SRE Grafana dashboards in your active cluster.`,
}

var dashboardsImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import and provision Grafana dashboards",
	Long:  `Generates and applies all SRE Grafana dashboards as auto-discovered ConfigMaps.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		if err := config.InitConfig(false); err != nil {
			fmt.Printf("%sConfiguration file not found. Provisioning with default settings...\n", utils.PrefixInfo)
			config.GlobalConfig = config.DefaultConfig()
		}

		ns := config.GlobalConfig.Namespace
		if ns == "" {
			ns = "observability"
		}

		return observability.ProvisionDashboards(ns, dashboardsDryRun)
	},
}

func init() {
	dashboardsImportCmd.Flags().BoolVar(&dashboardsDryRun, "dry-run", false, "Show what would be imported without actually importing")
	dashboardsCmd.AddCommand(dashboardsImportCmd)
	RootCmd.AddCommand(dashboardsCmd)
}
