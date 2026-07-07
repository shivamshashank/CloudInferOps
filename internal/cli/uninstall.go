package cli

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/shivamshashank/CloudInferOps/internal/config"
	"github.com/shivamshashank/CloudInferOps/internal/utils"
	"github.com/spf13/cobra"
)

var uninstallDryRun bool
var forceUninstall bool

var uninstallKubernetesCluster = func() error {
	fmt.Printf("%sRemoving Kubernetes cluster components and kubeadm state...\n", utils.PrefixInfo)

	cleanupCommands := []string{
		"sudo kubeadm reset --force || true",
		"sudo systemctl stop kubelet containerd || true",
		"sudo systemctl disable kubelet containerd || true",
		"sudo apt-get purge -y kubeadm kubelet kubectl kubernetes-cni cri-tools containerd.io || true",
		"sudo apt-get autoremove -y || true",
	}

	for _, cmd := range cleanupCommands {
		if _, stderr, err := utils.ExecCommandInteractive("", "bash", "-c", cmd); err != nil {
			fmt.Printf("%sWarning: cleanup step failed (%s): %s\n", utils.PrefixWarn, cmd, stderr)
		}
	}

	pathsToRemove := []string{
		"/etc/kubernetes",
		"/var/lib/etcd",
		"/var/lib/kubelet",
		"/var/lib/cni",
		"/var/lib/calico",
		"/var/run/calico",
		"/etc/cni/net.d",
		"/opt/cni/bin",
		"/var/log/pods",
	}
	for _, path := range pathsToRemove {
		if err := os.RemoveAll(path); err != nil && !os.IsNotExist(err) {
			fmt.Printf("%sWarning: could not remove %s: %v\n", utils.PrefixWarn, path, err)
		}
	}

	realHome, err := utils.GetRealHomeDir()
	if err == nil {
		_ = os.RemoveAll(filepath.Join(realHome, ".kube"))
	}

	fmt.Printf("%sKubernetes teardown completed.\n", utils.PrefixOK)
	return nil
}

var performUninstallBinaries = func(dryRun bool) error {
	if dryRun {
		fmt.Printf("%s[DRY-RUN] Removing Kubernetes-related binaries and config\n", utils.PrefixInfo)
		return nil
	}
	return uninstallKubernetesCluster()
}

var performUninstallInference = func(dryRun bool) error {
	fmt.Printf("%sUninstalling inference stack...\n", utils.PrefixInfo)

	if _, err := exec.LookPath("kubectl"); err != nil {
		fmt.Printf("%s'kubectl' not found, assuming inference stack is already removed.\n", utils.PrefixWarn)
		return nil
	}

	if dryRun {
		fmt.Printf("%s[DRY-RUN] kubectl delete namespace inference --ignore-not-found=true\n", utils.PrefixInfo)
		return nil
	}

	stopSpinner := utils.StartSpinner("Deleting inference namespace...")
	realHome, _ := utils.GetRealHomeDir()
	kubeEnv := map[string]string{
		"KUBECONFIG": filepath.Join(realHome, ".kube", "config"),
	}

	_, stderr, err := utils.ExecCommandEnv("", kubeEnv, "kubectl", "delete", "namespace", "inference", "--ignore-not-found=true")
	stopSpinner()

	if err != nil {
		return fmt.Errorf("failed to delete inference namespace: %w (stderr: %s)", err, stderr)
	}

	fmt.Printf("%sInference stack uninstalled successfully.\n", utils.PrefixOK)
	return nil
}

var performUninstallUI = func(dryRun bool) error {
	fmt.Printf("%sUninstalling UI portal...\n", utils.PrefixInfo)

	if _, err := exec.LookPath("kubectl"); err != nil {
		fmt.Printf("%s'kubectl' not found, assuming UI portal is already removed.\n", utils.PrefixWarn)
		return nil
	}

	ns := config.GlobalConfig.Namespace
	if ns == "" {
		ns = "observability"
	}

	resources := []struct {
		kind string
		name string
	}{
		{"ingress", "cloudinferops-ui"},
		{"service", "cloudinferops-ui"},
		{"deployment", "cloudinferops-ui"},
		{"clusterrolebinding", "cloudinferops-ui"},
		{"clusterrole", "cloudinferops-ui"},
		{"serviceaccount", "cloudinferops-ui"},
	}

	if dryRun {
		for _, r := range resources {
			if r.kind == "clusterrole" || r.kind == "clusterrolebinding" {
				fmt.Printf("%s[DRY-RUN] kubectl delete %s %s\n", utils.PrefixInfo, r.kind, r.name)
			} else {
				fmt.Printf("%s[DRY-RUN] kubectl delete %s %s -n %s\n", utils.PrefixInfo, r.kind, r.name, ns)
			}
		}
		return nil
	}

	stopSpinner := utils.StartSpinner("Deleting UI portal resources...")
	realHome, _ := utils.GetRealHomeDir()
	kubeEnv := map[string]string{
		"KUBECONFIG": filepath.Join(realHome, ".kube", "config"),
	}

	for _, r := range resources {
		var args []string
		if r.kind == "clusterrole" || r.kind == "clusterrolebinding" {
			args = []string{"delete", r.kind, r.name, "--ignore-not-found=true"}
		} else {
			args = []string{"delete", r.kind, r.name, "-n", ns, "--ignore-not-found=true"}
		}
		_, _, _ = utils.ExecCommandEnv("", kubeEnv, "kubectl", args...)
	}
	stopSpinner()

	fmt.Printf("%sUI portal uninstalled successfully.\n", utils.PrefixOK)
	return nil
}

var uninstallK8sCmd = &cobra.Command{
	Use:   "k8s",
	Short: "Uninstall Kubernetes-related binaries and resources",
	RunE: func(_ *cobra.Command, _ []string) error {
		return performUninstallBinaries(uninstallDryRun)
	},
}

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall CloudInferOps components",
	Long:  `Parent command for removing observability pipelines, gateway services, and local clusters.`,
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
		if !hasRootPrivileges() {
			return fmt.Errorf("the 'uninstall' command requires root privileges. Please run with sudo")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("\n%s🧹 CloudInferOps Uninstall Menu%s\n", utils.ColorBold, utils.ColorReset)
		fmt.Println("-----------------------------------------------------------------")
		fmt.Println("Select a component to uninstall:")
		fmt.Println("  1. Observability Stack (Prometheus, Grafana, Loki, Tempo, ArgoCD, etc.)")
		fmt.Println("  2. Inference Stack (Ollama, Gateway, etc.)")
		fmt.Println("  3. UI Portal")
		fmt.Println("  4. All Components (including Kubernetes cluster and config)")
		fmt.Print("Choose an option [1-4]: ")

		var response string
		_, _ = fmt.Fscanln(os.Stdin, &response)
		response = strings.TrimSpace(response)

		if err := config.InitConfig(false); err != nil {
			config.GlobalConfig = config.DefaultConfig()
		}

		switch response {
		case "1":
			if !forceUninstall && !promptConfirm("This will remove the CloudInferOps observability stack. Continue? [y/N]: ") {
				fmt.Printf("%sUninstall cancelled.\n", utils.PrefixWarn)
				return nil
			}
			return performUninstallObservability(uninstallDryRun)
		case "2":
			if !forceUninstall && !promptConfirm("This will remove the CloudInferOps inference stack. Continue? [y/N]: ") {
				fmt.Printf("%sUninstall cancelled.\n", utils.PrefixWarn)
				return nil
			}
			return performUninstallInference(uninstallDryRun)
		case "3":
			if !forceUninstall && !promptConfirm("This will remove the CloudInferOps UI portal. Continue? [y/N]: ") {
				fmt.Printf("%sUninstall cancelled.\n", utils.PrefixWarn)
				return nil
			}
			return performUninstallUI(uninstallDryRun)
		case "4":
			return runUninstallAll(uninstallDryRun, forceUninstall)
		default:
			fmt.Printf("%sInvalid choice '%s'. Uninstall cancelled.\n", utils.PrefixWarn, response)
			return nil
		}
	},
}

var uninstallObservabilityCmd = &cobra.Command{
	Use:   "observability",
	Short: "Remove the observability stack",
	RunE: func(_ *cobra.Command, _ []string) error {
		if err := config.InitConfig(false); err != nil {
			config.GlobalConfig = config.DefaultConfig()
		}
		if !forceUninstall && !promptConfirm("This will remove the CloudInferOps observability stack. Continue? [y/N]: ") {
			fmt.Printf("%sUninstall cancelled.\n", utils.PrefixWarn)
			return nil
		}
		return performUninstallObservability(uninstallDryRun)
	},
}

var uninstallInferenceCmd = &cobra.Command{
	Use:   "inference",
	Short: "Remove the inference stack",
	RunE: func(_ *cobra.Command, _ []string) error {
		if err := config.InitConfig(false); err != nil {
			config.GlobalConfig = config.DefaultConfig()
		}
		if !forceUninstall && !promptConfirm("This will remove the CloudInferOps inference stack. Continue? [y/N]: ") {
			fmt.Printf("%sUninstall cancelled.\n", utils.PrefixWarn)
			return nil
		}
		return performUninstallInference(uninstallDryRun)
	},
}

var uninstallUICmd = &cobra.Command{
	Use:   "ui",
	Short: "Remove the UI portal",
	RunE: func(_ *cobra.Command, _ []string) error {
		if err := config.InitConfig(false); err != nil {
			config.GlobalConfig = config.DefaultConfig()
		}
		if !forceUninstall && !promptConfirm("This will remove the CloudInferOps UI portal. Continue? [y/N]: ") {
			fmt.Printf("%sUninstall cancelled.\n", utils.PrefixWarn)
			return nil
		}
		return performUninstallUI(uninstallDryRun)
	},
}

var uninstallWebhookHandlerCmd = &cobra.Command{
	Use:   "webhook-handler",
	Short: "Remove the custom Go webhook incident gateway",
	RunE: func(_ *cobra.Command, _ []string) error {
		if err := config.InitConfig(false); err != nil {
			config.GlobalConfig = config.DefaultConfig()
		}
		if !forceUninstall && !promptConfirm("This will remove the CloudInferOps webhook handler. Continue? [y/N]: ") {
			fmt.Printf("%sUninstall cancelled.\n", utils.PrefixWarn)
			return nil
		}
		return performUninstallWebhookHandler(uninstallDryRun)
	},
}

var runUninstallAll = func(dryRun, force bool) error {
	if force {
		fmt.Printf("\n%s--force flag detected. Starting complete uninstallation...\n", utils.PrefixInfo)
		_ = performUninstallObservability(dryRun)
		_ = performUninstallInference(dryRun)
		_ = performUninstallUI(dryRun)
		_ = performUninstallWebhookHandler(dryRun)
		_ = performUninstallBinaries(dryRun)
		_ = performUninstallConfig(dryRun)
		fmt.Printf("\n%sComplete uninstall process finished.\n", utils.PrefixOK)
		return nil
	}

	fmt.Printf("\n%s🧹 CloudInferOps Interactive Uninstall%s\n", utils.ColorBold, utils.ColorReset)
	fmt.Println("-----------------------------------------------------------------")

	if promptConfirm("1. Remove the Observability Stack (Prometheus, Grafana, etc.)? [y/N]: ") {
		_ = performUninstallObservability(dryRun)
	} else {
		fmt.Printf("%sSkipping observability stack removal.\n", utils.PrefixInfo)
	}

	if promptConfirm("2. Remove the Inference Stack (Ollama, Gateway, etc.)? [y/N]: ") {
		_ = performUninstallInference(dryRun)
	} else {
		fmt.Printf("%sSkipping inference stack removal.\n", utils.PrefixInfo)
	}

	if promptConfirm("3. Remove the UI Portal? [y/N]: ") {
		_ = performUninstallUI(dryRun)
	} else {
		fmt.Printf("%sSkipping UI portal removal.\n", utils.PrefixInfo)
	}

	if promptConfirm("4. Remove the Webhook Handler? [y/N]: ") {
		_ = performUninstallWebhookHandler(dryRun)
	} else {
		fmt.Printf("%sSkipping webhook handler removal.\n", utils.PrefixInfo)
	}

	if promptConfirm("5. Remove Kubernetes cluster and kubeadm artifacts? [y/N]: ") {
		_ = performUninstallBinaries(dryRun)
	} else {
		fmt.Printf("%sSkipping Kubernetes cleanup.\n", utils.PrefixInfo)
	}

	if promptConfirm("6. Delete CloudInferOps configuration (~/.cloudinferops)? [y/N]: ") {
		_ = performUninstallConfig(dryRun)
	} else {
		fmt.Printf("%sSkipping configuration removal.\n", utils.PrefixInfo)
	}

	fmt.Printf("\n%sUninstall process finished.\n", utils.PrefixOK)
	return nil
}

var uninstallAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Interactively remove all CloudInferOps components",
	RunE: func(_ *cobra.Command, _ []string) error {
		if err := config.InitConfig(false); err != nil {
			config.GlobalConfig = config.DefaultConfig()
		}
		return runUninstallAll(uninstallDryRun, forceUninstall)
	},
}

func init() {
	uninstallCmd.PersistentFlags().BoolVar(&uninstallDryRun, "dry-run", false, "Print commands without executing them")
	uninstallCmd.PersistentFlags().BoolVarP(&forceUninstall, "force", "f", false, "Skip confirmation prompts")

	uninstallCmd.AddCommand(uninstallObservabilityCmd)
	uninstallCmd.AddCommand(uninstallInferenceCmd)
	uninstallCmd.AddCommand(uninstallUICmd)
	uninstallCmd.AddCommand(uninstallWebhookHandlerCmd)
	uninstallCmd.AddCommand(uninstallK8sCmd)
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

	ns := config.GlobalConfig.Namespace
	if ns == "" {
		ns = "observability"
	}

	if dryRun {
		fmt.Printf("%s[DRY-RUN] kubectl delete namespace %s --ignore-not-found=true\n", utils.PrefixInfo, ns)
		return nil
	}

	// Delete the namespace to clear all Helm releases, ConfigMaps, Secrets, and Deployments inside it
	stopSpinner := utils.StartSpinner(fmt.Sprintf("Deleting %s namespace (this may take a few minutes)...", ns))

	// Provide kubeEnv fallback in case it's run via sudo
	realHome, _ := utils.GetRealHomeDir()
	kubeEnv := map[string]string{
		"KUBECONFIG": filepath.Join(realHome, ".kube", "config"),
	}

	_, stderr, err := utils.ExecCommandEnv("", kubeEnv, "kubectl", "delete", "namespace", ns, "--ignore-not-found=true")
	stopSpinner()

	if err != nil {
		return fmt.Errorf("failed to delete %s namespace: %w (stderr: %s)", ns, err, stderr)
	}

	fmt.Printf("%sObservability stack uninstalled successfully.\n", utils.PrefixOK)
	return nil
}

func performUninstallWebhookHandler(dryRun bool) error {
	fmt.Printf("%sUninstalling webhook handler...\n", utils.PrefixInfo)

	if _, err := exec.LookPath("kubectl"); err != nil {
		fmt.Printf("%s'kubectl' not found, assuming webhook handler is already removed.\n", utils.PrefixWarn)
		return nil
	}

	ns := config.GlobalConfig.Namespace
	if ns == "" {
		ns = "observability"
	}

	if dryRun {
		fmt.Printf("%s[DRY-RUN] helm uninstall cloudinferops-webhook-handler -n %s\n", utils.PrefixInfo, ns)
		fmt.Printf("%s[DRY-RUN] kubectl delete deployment,svc cloudinferops-webhook-handler -n %s --ignore-not-found=true\n", utils.PrefixInfo, ns)
		return nil
	}

	stopSpinner := utils.StartSpinner("Deleting webhook handler...")
	realHome, _ := utils.GetRealHomeDir()
	kubeEnv := map[string]string{
		"KUBECONFIG": filepath.Join(realHome, ".kube", "config"),
	}

	_, _, _ = utils.ExecCommandEnv("", kubeEnv, "helm", "uninstall", "cloudinferops-webhook-handler", "-n", ns)
	_, _, _ = utils.ExecCommandEnv("", kubeEnv, "kubectl", "delete", "deployment,svc", "cloudinferops-webhook-handler", "-n", ns, "--ignore-not-found=true")
	stopSpinner()

	fmt.Printf("%sWebhook handler uninstalled successfully.\n", utils.PrefixOK)
	return nil
}

func performUninstallConfig(dryRun bool) error {
	fmt.Printf("%sChecking for CloudInferOps configuration to remove...\n", utils.PrefixInfo)

	realHome, err := utils.GetRealHomeDir()
	if err != nil {
		return nil
	}

	configDir := filepath.Join(realHome, ".cloudinferops")
	if _, err := os.Stat(configDir); err == nil {
		if dryRun {
			fmt.Printf("%s[DRY-RUN] rm -rf %s\n", utils.PrefixInfo, configDir)
		} else {
			fmt.Printf("%sRemoving CloudInferOps configuration directory at %s...\n", utils.PrefixInfo, configDir)
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
