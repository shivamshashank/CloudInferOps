package cli

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/shivamshashank/CloudInferOps/internal/config"
	"github.com/shivamshashank/CloudInferOps/internal/utils"
	"github.com/spf13/cobra"
)

var (
	logsComponent string
	logsFollow    bool
	logsTail      int
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View logs of CloudInferOps components",
	Long:  `Retrieve and stream logs for CloudInferOps infrastructure components deployed in the Kubernetes cluster.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		if err := config.InitConfig(false); err != nil {
			config.GlobalConfig = config.DefaultConfig()
		}

		ns := config.GlobalConfig.Namespace
		if ns == "" {
			ns = "observability"
		}

		// Check if kubectl is available and cluster is reachable
		_, _, err := utils.ExecCommand("", "kubectl", "version", "--client")
		if err != nil {
			return fmt.Errorf("kubectl is required but was not found or is not working: %w", err)
		}

		// Fetch all pods in the namespace
		stdout, stderr, err := utils.ExecCommand("", "kubectl", "get", "pods", "-n", ns, "-o", "jsonpath={range .items[*]}{.metadata.name}{\"\\n\"}{end}")
		if err != nil {
			return fmt.Errorf("failed to fetch pods: %w (stderr: %s)", err, stderr)
		}

		pods := []string{}
		for _, p := range strings.Split(stdout, "\n") {
			p = strings.TrimSpace(p)
			if p != "" {
				pods = append(pods, p)
			}
		}

		if len(pods) == 0 {
			fmt.Printf("%sNo pods found in namespace '%s'.\n", utils.PrefixWarn, ns)
			return nil
		}

		var targetPods []string

		if logsComponent != "" {
			// Filter pods by component name
			compLower := strings.ToLower(logsComponent)
			for _, p := range pods {
				if strings.Contains(strings.ToLower(p), compLower) {
					targetPods = append(targetPods, p)
				}
			}

			if len(targetPods) == 0 {
				return fmt.Errorf("no pods found matching component '%s' in namespace '%s'", logsComponent, ns)
			}
		} else {
			// Interactive select or default
			fmt.Printf("Multiple pods found in namespace '%s'. Please select a pod to view logs:\n", ns)
			for i, p := range pods {
				fmt.Printf("  %d. %s\n", i+1, p)
			}
			fmt.Print("Choose a pod [1-", len(pods), "]: ")

			reader := bufio.NewReader(os.Stdin)
			choiceStr, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read choice: %w", err)
			}
			choiceStr = strings.TrimSpace(choiceStr)
			choice, err := strconv.Atoi(choiceStr)
			if err != nil || choice < 1 || choice > len(pods) {
				return fmt.Errorf("invalid choice: %s", choiceStr)
			}
			targetPods = append(targetPods, pods[choice-1])
		}

		// If multiple pods matched, let the user choose which one
		var podToLog string
		if len(targetPods) > 1 {
			fmt.Printf("Multiple pods matched component '%s':\n", logsComponent)
			for i, p := range targetPods {
				fmt.Printf("  %d. %s\n", i+1, p)
			}
			fmt.Print("Choose a pod [1-", len(targetPods), "]: ")

			reader := bufio.NewReader(os.Stdin)
			choiceStr, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read choice: %w", err)
			}
			choiceStr = strings.TrimSpace(choiceStr)
			choice, err := strconv.Atoi(choiceStr)
			if err != nil || choice < 1 || choice > len(targetPods) {
				return fmt.Errorf("invalid choice: %s", choiceStr)
			}
			podToLog = targetPods[choice-1]
		} else {
			podToLog = targetPods[0]
		}

		// Build kubectl logs command arguments
		args := []string{"logs", podToLog, "-n", ns}
		if logsFollow {
			args = append(args, "-f")
		}
		if logsTail >= 0 {
			args = append(args, fmt.Sprintf("--tail=%d", logsTail))
		}

		fmt.Printf("%sFetching logs for pod '%s' in namespace '%s'...\n", utils.PrefixInfo, podToLog, ns)
		_, _, err = utils.ExecCommandStream("", "kubectl", args...)
		if err != nil {
			// Check if the error is just due to process being interrupted/killed (Ctrl+C)
			if strings.Contains(err.Error(), "interrupt") || strings.Contains(err.Error(), "killed") || strings.Contains(err.Error(), "exit status 130") {
				return nil
			}
			return fmt.Errorf("failed to stream logs: %w", err)
		}

		return nil
	},
}

func init() {
	logsCmd.Flags().StringVarP(&logsComponent, "component", "c", "", "Component name to view logs for (e.g. grafana, prometheus, loki)")
	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "Specify if the logs should be streamed continuously")
	logsCmd.Flags().IntVarP(&logsTail, "tail", "t", -1, "Lines of recent log history to show (default is all)")
	RootCmd.AddCommand(logsCmd)
}
