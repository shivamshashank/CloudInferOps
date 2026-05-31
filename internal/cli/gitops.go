package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/shivamshashank/StackPulse/internal/config"
	"github.com/shivamshashank/StackPulse/internal/doctor"
	"github.com/shivamshashank/StackPulse/internal/utils"
	"github.com/spf13/cobra"
)

var gitopsCmd = &cobra.Command{
	Use:   "gitops",
	Short: "Manage and view GitOps applications via ArgoCD",
}

var gitopsStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show sync and health status of all ArgoCD applications",
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, hasK8s := doctor.CheckK8sCluster(); !hasK8s {
			return fmt.Errorf("kubernetes cluster unreachable")
		}

		if err := config.InitConfig(false); err != nil {
			config.GlobalConfig = config.DefaultConfig()
		}

		ns := config.GlobalConfig.Namespace
		if ns == "" {
			ns = "observability"
		}

		fmt.Println()
		fmt.Printf("%s%sGitOps Applications Status%s\n", utils.PrefixInfo, utils.ColorBold, utils.ColorReset)
		fmt.Println("-----------------------------------------------------------------")

		appsOut, _, err := utils.ExecCommand("", "kubectl", "get", "applications", "-n", ns, "-o", "json")
		if err != nil || appsOut == "" || strings.HasPrefix(appsOut, "No resources found") {
			fmt.Printf("%sNo GitOps applications found. Ensure ArgoCD is deployed and managing resources.\n", utils.PrefixWarn)
			fmt.Println("-----------------------------------------------------------------")
			return nil
		}

		var appList struct {
			Items []struct {
				Metadata struct {
					Name string `json:"name"`
				} `json:"metadata"`
				Status struct {
					Sync struct {
						Status string `json:"status"`
					} `json:"sync"`
					Health struct {
						Status string `json:"status"`
					} `json:"health"`
				} `json:"status"`
			} `json:"items"`
		}

		if err := json.Unmarshal([]byte(appsOut), &appList); err != nil {
			fmt.Printf("%sFailed to parse GitOps applications status. Raw output:\n\n", utils.PrefixWarn)
			fmt.Println(appsOut)
			fmt.Println("-----------------------------------------------------------------")
			return nil
		}

		appMap := make(map[string]struct{ Sync, Health string })
		for _, item := range appList.Items {
			appMap[item.Metadata.Name] = struct{ Sync, Health string }{
				Sync:   item.Status.Sync.Status,
				Health: item.Status.Health.Status,
			}
		}

		getComponentStatus := func(appName string) (string, string) {
			if vals, exists := appMap[appName]; exists {
				sync := vals.Sync
				if sync == "" {
					sync = "Unknown"
				}
				health := vals.Health
				if health == "" {
					health = "Unknown"
				}
				return sync, health
			}
			return "Missing", "Missing"
		}

		fmt.Printf("%-16s %-10s %-10s\n", "APPLICATION", "STATUS", "HEALTH")

		promSync, promHealth := getComponentStatus("stackpulse-prometheus")
		lokiSync, lokiHealth := getComponentStatus("stackpulse-loki")
		tempoSync, tempoHealth := getComponentStatus("stackpulse-tempo")
		otelSync, otelHealth := getComponentStatus("stackpulse-otel")

		fmt.Printf("%-16s %-10s %-10s\n", "prometheus", promSync, promHealth)
		fmt.Printf("%-16s %-10s %-10s\n", "grafana", promSync, promHealth)
		fmt.Printf("%-16s %-10s %-10s\n", "loki", lokiSync, lokiHealth)
		fmt.Printf("%-16s %-10s %-10s\n", "tempo", tempoSync, tempoHealth)
		fmt.Printf("%-16s %-10s %-10s\n", "otel", otelSync, otelHealth)
		fmt.Printf("%-16s %-10s %-10s\n", "alertmanager", promSync, promHealth)

		fmt.Println("-----------------------------------------------------------------")
		return nil
	},
}

func init() {
	gitopsCmd.AddCommand(gitopsStatusCmd)
	RootCmd.AddCommand(gitopsCmd)
}
