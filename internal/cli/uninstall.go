package cli

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/shivamshashank/StackPulse/internal/utils"
	"github.com/spf13/cobra"
)

var uninstallDryRun bool
var forceUninstall bool

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall StackPulse components",
	Long:  `Parent command for removing observability pipelines, gateway services, and local clusters.`,
}

var uninstallObservabilityCmd = &cobra.Command{
	Use:   "observability",
	Short: "Remove the observability stack",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !forceUninstall && !promptConfirm("This will remove the StackPulse observability stack. Continue? [y/N]: ") {
			fmt.Printf("%sUninstall cancelled.\n", utils.PrefixWarn)
			return nil
		}
		return performUninstallObservability(uninstallDryRun)
	},
}

var uninstallAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Remove observability stack and tear down local clusters",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !forceUninstall && !promptConfirm("This will remove the observability stack AND delete any local StackPulse clusters (k3s, minikube, kind). Continue? [y/N]: ") {
			fmt.Printf("%sUninstall cancelled.\n", utils.PrefixWarn)
			return nil
		}
		// 1. Try to uninstall observability cleanly first
		_ = performUninstallObservability(uninstallDryRun)

		// 2. Tear down clusters
		_ = performUninstallClusters(uninstallDryRun)

		// 3. Remove downloaded binaries
		_ = performUninstallBinaries(uninstallDryRun)

		// 4. Remove configuration
		return performUninstallConfig(uninstallDryRun)
	},
}

func init() {
	uninstallCmd.PersistentFlags().BoolVar(&uninstallDryRun, "dry-run", false, "Print commands without executing them")
	uninstallCmd.PersistentFlags().BoolVarP(&forceUninstall, "force", "f", false, "Skip confirmation prompts")

	uninstallCmd.AddCommand(uninstallObservabilityCmd)
	uninstallCmd.AddCommand(uninstallAllCmd)
	RootCmd.AddCommand(uninstallCmd)
}

func promptConfirm(msg string) bool {
	fmt.Print(msg)
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

func performUninstallObservability(dryRun bool) error {
	fmt.Printf("%sUninstalling observability stack...\n", utils.PrefixInfo)

	if _, err := exec.LookPath("kubectl"); err != nil {
		fmt.Printf("%s'kubectl' not found, assuming observability stack is already removed.\n", utils.PrefixWarn)
		return nil
	}

	if dryRun {
		fmt.Printf("%s[DRY-RUN] kubectl delete namespace observability --ignore-not-found=true\n", utils.PrefixInfo)
		return nil
	}

	// Delete the namespace to clear all Helm releases, ConfigMaps, Secrets, and Deployments inside it
	stopSpinner := utils.StartSpinner("Deleting observability namespace (this may take a few minutes)...")

	// Provide kubeEnv fallback in case it's run via sudo
	realHome, _ := utils.GetRealHomeDir()
	kubeEnv := map[string]string{
		"KUBECONFIG": filepath.Join(realHome, ".kube", "config"),
	}

	_, stderr, err := utils.ExecCommandEnv("", kubeEnv, "kubectl", "delete", "namespace", "observability", "--ignore-not-found=true")
	stopSpinner()

	if err != nil {
		return fmt.Errorf("failed to delete observability namespace: %w (stderr: %s)", err, stderr)
	}

	fmt.Printf("%sObservability stack uninstalled successfully.\n", utils.PrefixOK)
	return nil
}

func performUninstallClusters(dryRun bool) error {
	fmt.Printf("%sChecking for local clusters to tear down...\n", utils.PrefixInfo)

	realHome, _ := utils.GetRealHomeDir()
	kubeEnv := map[string]string{
		"KUBECONFIG":    filepath.Join(realHome, ".kube", "config"),
		"MINIKUBE_HOME": realHome,
	}

	// 1. Kind
	if _, err := exec.LookPath("kind"); err == nil {
		// Verify if cluster exists
		out, _, _ := utils.ExecCommandEnv("", kubeEnv, "kind", "get", "clusters")
		if strings.Contains(out, "stackpulse") {
			if dryRun {
				fmt.Printf("%s[DRY-RUN] kind delete cluster --name stackpulse\n", utils.PrefixInfo)
			} else {
				fmt.Printf("%sRemoving kind cluster 'stackpulse'...\n", utils.PrefixInfo)
				_, _, _ = utils.ExecCommandEnv("", kubeEnv, "kind", "delete", "cluster", "--name", "stackpulse")
			}
		}
	}

	// 2. Minikube
	if _, err := exec.LookPath("minikube"); err == nil {
		// Verify if minikube profile exists
		out, _, _ := utils.ExecCommandEnv("", kubeEnv, "minikube", "profile", "list")
		if strings.Contains(out, "minikube") {
			if dryRun {
				fmt.Printf("%s[DRY-RUN] minikube delete\n", utils.PrefixInfo)
			} else {
				fmt.Printf("%sRemoving minikube cluster...\n", utils.PrefixInfo)
				_, _, _ = utils.ExecCommandEnv("", kubeEnv, "minikube", "delete")
			}
		}
	}

	// 3. K3s
	if _, err := os.Stat("/usr/local/bin/k3s-uninstall.sh"); err == nil {
		if dryRun {
			fmt.Printf("%s[DRY-RUN] /usr/local/bin/k3s-uninstall.sh\n", utils.PrefixInfo)
		} else {
			fmt.Printf("%sRemoving k3s cluster...\n", utils.PrefixInfo)
			if os.Getuid() == 0 {
				_, _, _ = utils.ExecCommand("", "/usr/local/bin/k3s-uninstall.sh")
			} else {
				_, _, _ = utils.ExecCommandInteractive("", "sudo", "/usr/local/bin/k3s-uninstall.sh")
			}
		}
	}

	fmt.Printf("%sLocal cluster cleanup complete.\n", utils.PrefixOK)
	return nil
}

func performUninstallBinaries(dryRun bool) error {
	fmt.Printf("%sChecking for downloaded binaries to remove...\n", utils.PrefixInfo)

	binaries := []string{"kubectl", "kind", "minikube"}
	for _, bin := range binaries {
		path := filepath.Join("/usr/local/bin", bin)
		if _, err := os.Stat(path); err == nil {
			if dryRun {
				fmt.Printf("%s[DRY-RUN] rm -f %s\n", utils.PrefixInfo, path)
			} else {
				fmt.Printf("%sRemoving %s binary from %s...\n", utils.PrefixInfo, bin, path)
				if os.Getuid() == 0 {
					_, _, _ = utils.ExecCommand("", "rm", "-f", path)
				} else {
					_, _, _ = utils.ExecCommandInteractive("", "sudo", "rm", "-f", path)
				}
			}
		}
	}

	fmt.Printf("%sBinary cleanup complete.\n", utils.PrefixOK)
	return nil
}

func performUninstallConfig(dryRun bool) error {
	fmt.Printf("%sChecking for StackPulse configuration to remove...\n", utils.PrefixInfo)

	realHome, err := utils.GetRealHomeDir()
	if err != nil {
		return nil
	}

	configDir := filepath.Join(realHome, ".stackpulse")
	if _, err := os.Stat(configDir); err == nil {
		if dryRun {
			fmt.Printf("%s[DRY-RUN] rm -rf %s\n", utils.PrefixInfo, configDir)
		} else {
			fmt.Printf("%sRemoving StackPulse configuration directory at %s...\n", utils.PrefixInfo, configDir)
			if os.Getuid() == 0 {
				_, _, _ = utils.ExecCommand("", "rm", "-rf", configDir)
			} else {
				_, _, _ = utils.ExecCommandInteractive("", "sudo", "rm", "-rf", configDir)
			}
		}
	}

	fmt.Printf("%sConfiguration cleanup complete.\n", utils.PrefixOK)
	return nil
}
