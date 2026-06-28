package cli

import (
	"fmt"
	"os"

	"github.com/shivamshashank/CloudInferOps/internal/config"
	"github.com/shivamshashank/CloudInferOps/internal/utils"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize CloudInferOps configuration",
	Long:  `Creates the default ~/.cloudinferops/config.yaml configuration file if it does not exist.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := config.GetConfigPath()
		if err != nil {
			return err
		}

		if _, err := os.Stat(path); err == nil {
			fmt.Printf("%sStackPulse is already initialized. Config file found at: %s\n", utils.PrefixOK, path)
			return nil
		}

		fmt.Println("Initializing CloudInferOps configuration...")
		if err := config.InitConfig(true); err != nil {
			return fmt.Errorf("failed to initialize configuration: %w", err)
		}

		fmt.Printf("%sSuccessfully initialized default configuration at: %s\n", utils.PrefixOK, path)
		fmt.Printf("%sReview or modify this file to configure your namespaces and credentials.\n", utils.PrefixInfo)
		return nil
	},
}

func init() {
	RootCmd.AddCommand(initCmd)
}
