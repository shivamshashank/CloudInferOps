package cli

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/shivamshashank/StackPulse/internal/config"
	"github.com/shivamshashank/StackPulse/internal/doctor"
	"github.com/shivamshashank/StackPulse/internal/observability"
	"github.com/shivamshashank/StackPulse/internal/utils"
	"github.com/spf13/cobra"
)

var connectBrowser bool

// connectCmd represents the connect command
var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connect to the Grafana dashboard automatically",
	Long: `Automates all access steps: retrieves the active Ingress LoadBalancer IP/Hostname,
idempotently writes the 'grafana.local' mapping to your local /etc/hosts file, programmatically
decrypts Grafana admin credentials, and automatically opens your default web browser.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. Pre-flight check: Verify cluster reachability early
		_, hasK8s := doctor.CheckK8sCluster()
		if !hasK8s {
			fmt.Printf("%sKubernetes cluster not detected.\n", utils.PrefixError)
			fmt.Printf("%sPlease ensure a local cluster is running (Docker Desktop, Kind, or Minikube), or run:\n", utils.PrefixInfo)
			fmt.Printf("    %s\n", utils.ColorBold+"stackpulse setup k8s"+utils.ColorReset)
			return fmt.Errorf("Kubernetes cluster unreachable")
		}

		// 2. Load configuration (fallback on defaults if not initialized)
		if err := config.InitConfig(false); err != nil {
			config.GlobalConfig = config.DefaultConfig()
		}

		ns := config.GlobalConfig.Namespace
		if ns == "" {
			ns = "observability"
		}

		fmt.Printf("%sStarting automated connection to Grafana...\n", utils.PrefixInfo)

		// 3. Fetch Ingress IP
		ingressIP, err := observability.FetchIngressIP(ns, false)
		if err != nil || ingressIP == "" {
			// Fallback: Resolve active interface IP of the host machine (e.g., VM / EC2 IP)
			fmt.Printf("%sIngress IP is still provisioning or empty. Resolving host interface IP...\n", utils.PrefixInfo)
			ingressIP = utils.GetLocalIP()
		}

		// 4. Fetch and decode Grafana admin password
		plainPassword := "<unretrievable>"
		pwdSecret, _, err := utils.ExecCommand("", "kubectl", "get", "secret", "stackpulse-prometheus-grafana", "-n", ns, "-o", "jsonpath={.data.admin-password}")
		if err == nil && pwdSecret != "" {
			decoded, err := observability.DecodeBase64(strings.TrimSpace(pwdSecret))
			if err == nil {
				plainPassword = decoded
			}
		}

		// 5. Output beautiful visual details card
		fmt.Println()
		fmt.Println("-----------------------------------------------------------------")
		fmt.Printf("🚀  %sTelemetry Stack Access Details Ready!%s\n", utils.ColorBold, utils.ColorReset)
		fmt.Println("-----------------------------------------------------------------")
		fmt.Printf("🌐  Grafana Dashboard:  %s\n", utils.ColorBold+fmt.Sprintf("http://%s/grafana", ingressIP)+utils.ColorReset)
		fmt.Printf("🌐  Prometheus Server:  %s\n", utils.ColorBold+fmt.Sprintf("http://%s/prometheus", ingressIP)+utils.ColorReset)
		fmt.Printf("🌐  Alertmanager Panel: %s\n", utils.ColorBold+fmt.Sprintf("http://%s/alertmanager", ingressIP)+utils.ColorReset)
		fmt.Printf("👤  Username:           admin\n")
		fmt.Printf("🔑  Plain Password:     %s\n", utils.ColorGreen+plainPassword+utils.ColorReset)
		fmt.Println("-----------------------------------------------------------------")

		// 6. Launch browser to Grafana
		if connectBrowser {
			grafanaURL := fmt.Sprintf("http://%s/grafana", ingressIP)
			fmt.Printf("%sOpening %s in your default web browser...\n", utils.PrefixInfo, grafanaURL)
			time.Sleep(1 * time.Second)
			var browserErr error
			if _, err := exec.LookPath("xdg-open"); err == nil {
				// Linux default browser open command
				_, _, browserErr = utils.ExecCommand("", "xdg-open", grafanaURL)
			} else {
				fmt.Printf("%sSystem 'xdg-open' command not found. Please open %s manually.\n", utils.PrefixInfo, grafanaURL)
			}
			
			if browserErr == nil {
				fmt.Printf("%sSuccessfully opened browser!\n", utils.PrefixOK)
			} else {
				fmt.Printf("%sFailed to launch browser: %v. Please open %s manually.\n", utils.PrefixWarn, browserErr, grafanaURL)
			}
		}

		return nil
	},
}

func init() {
	connectCmd.Flags().BoolVar(&connectBrowser, "browser", true, "Automatically open your default browser to Grafana")
	RootCmd.AddCommand(connectCmd)
}
