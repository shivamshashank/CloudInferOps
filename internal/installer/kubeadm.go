package installer

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/shivamshashank/CloudInferOps/internal/config"
	"github.com/shivamshashank/CloudInferOps/internal/utils"
)

// InstallKubeadm sets up a single-node Kubernetes cluster using kubeadm.
// This is a multi-step process intended for Linux hosts.
func InstallKubeadm() error {
	fmt.Printf("%sStarting kubeadm installation...\n", utils.PrefixInfo)

	steps := []struct {
		name string
		fn   func() error
	}{
		{"Disabling swap", disableSwap},
		{"Configuring kernel modules and sysctl", configureKernel},
		{"Installing containerd", installContainerd},
		{"Installing Kubernetes packages (kubelet, kubeadm, kubectl)", installKubePackages},
		{"Initializing cluster with kubeadm", initKubeadmCluster},
		{"Configuring kubectl for the user", configureKubeconfig},
		{"Patching kube-proxy configuration", patchKubeProxy},
		{"Untainting control-plane node", untaintControlPlaneNode},
		{"Installing Calico CNI", installCalicoCNI},
		{"Installing Metrics Server", installMetricsServer},
		{"Installing Local Path Provisioner for storage", installLocalPathProvisioner},
		{"Waiting for cluster components to be ready", waitForClusterReady},
	}

	for _, step := range steps {
		fmt.Printf("%s%s...\n", utils.PrefixInfo, step.name)
		if err := step.fn(); err != nil {
			return fmt.Errorf("step '%s' failed: %w", step.name, err)
		}
		fmt.Printf("%s%s... Done\n", utils.PrefixOK, step.name)
	}

	fmt.Printf("%sKubeadm cluster installation completed successfully.\n", utils.PrefixOK)
	return nil
}

func disableSwap() error {
	// Disable swap for the current session
	if _, stderr, err := utils.ExecCommand("", "sudo", "swapoff", "-a"); err != nil {
		return fmt.Errorf("failed to run 'swapoff -a': %w (stderr: %s)", err, stderr)
	}

	// Disable swap permanently in fstab
	fstabPath := "/etc/fstab"
	content, err := os.ReadFile(fstabPath)
	if err != nil {
		return fmt.Errorf("failed to read /etc/fstab: %w", err)
	}

	var newLines []string
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.Contains(line, "swap") && !strings.HasPrefix(strings.TrimSpace(line), "#") {
			newLines = append(newLines, "# "+line)
		} else {
			newLines = append(newLines, line)
		}
	}

	newContent := strings.Join(newLines, "\n")
	// #nosec G306 - fstab needs these permissions
	return os.WriteFile(fstabPath, []byte(newContent), 0644)
}

func configureKernel() error {
	// Load required kernel modules
	modules := []string{"overlay", "br_netfilter", "nf_conntrack", "xt_conntrack", "ip_tables", "ip6_tables"}
	for _, mod := range modules {
		if _, stderr, err := utils.ExecCommand("", "sudo", "modprobe", mod); err != nil {
			return fmt.Errorf("failed to modprobe %s: %w (stderr: %s)", mod, err, stderr)
		}
	}

	// Configure required sysctl params
	sysctlConfig := `
net.bridge.bridge-nf-call-iptables  = 1
net.bridge.bridge-nf-call-ip6tables = 1
net.ipv4.ip_forward                 = 1
net.netfilter.nf_conntrack_max      = 131072
net.netfilter.nf_conntrack_tcp_timeout_established = 180
`
	// #nosec G306 - sysctl config needs these permissions
	if err := os.WriteFile("/etc/sysctl.d/99-kubernetes-cri.conf", []byte(sysctlConfig), 0644); err != nil {
		return fmt.Errorf("failed to write sysctl config: %w", err)
	}

	// Apply sysctl params without reboot
	if _, stderr, err := utils.ExecCommand("", "sudo", "sysctl", "--system"); err != nil {
		return fmt.Errorf("failed to apply sysctl settings: %w (stderr: %s)", err, stderr)
	}

	return nil
}

func installContainerd() error {
	installCommands := []string{
		"sudo DEBIAN_FRONTEND=noninteractive apt-get update",
		"sudo DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends ca-certificates curl",
		"sudo install -m 0755 -d /etc/apt/keyrings",
		"sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc",
		"sudo chmod a+r /etc/apt/keyrings/docker.asc",
		`echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null`,
		"sudo DEBIAN_FRONTEND=noninteractive apt-get update",
		"sudo DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends containerd.io",
		"sudo mkdir -p /etc/containerd",
		"sudo containerd config default | sudo tee /etc/containerd/config.toml",
		`sudo sed -i 's/SystemdCgroup = false/SystemdCgroup = true/' /etc/containerd/config.toml`,
		"sudo systemctl restart containerd",
	}

	for _, cmd := range installCommands {
		if err := runInstallCommand(cmd); err != nil {
			return err
		}
	}
	return nil
}

func installKubePackages() error {
	installCommands := []string{
		"sudo DEBIAN_FRONTEND=noninteractive apt-get update",
		"sudo DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends apt-transport-https ca-certificates curl gpg",
		"sudo install -m 0755 -d /etc/apt/keyrings",
		"curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.30/deb/Release.key | sudo gpg --batch --yes --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg",
		`echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.30/deb/ /' | sudo tee /etc/apt/sources.list.d/kubernetes.list > /dev/null`,
		"sudo DEBIAN_FRONTEND=noninteractive apt-get update",
		"sudo DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends kubelet kubeadm kubectl",
		"sudo apt-mark hold kubelet kubeadm kubectl",
	}

	for _, cmd := range installCommands {
		if err := runInstallCommand(cmd); err != nil {
			return err
		}
	}
	return nil
}

func runInstallCommand(command string) error {
	_, stderr, err := utils.ExecCommandStream("", "sh", "-c", command)
	if err != nil {
		return fmt.Errorf("command failed: %w (stderr: %s)", err, stderr)
	}
	return nil
}

func initKubeadmCluster() error {
	// 10.244.0.0/16 is a common non-conflicting CIDR for pod networking
	_, stderr, err := utils.ExecCommandInteractive("", "sudo", "kubeadm", "init", "--pod-network-cidr=10.244.0.0/16")
	if err != nil {
		return fmt.Errorf("kubeadm init failed: %w (stderr: %s)", err, stderr)
	}
	return nil
}

func configureKubeconfig() error {
	if err := config.InitConfig(true); err != nil {
		return fmt.Errorf("failed to load configuration to find home dir: %w", err)
	}
	homeDir, err := utils.GetRealHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}
	kubeDir := homeDir + "/.kube"
	kubeConfigFile := kubeDir + "/config"

	// Create .kube directory for the user
	if _, stderr, err := utils.ExecCommandInteractive("", "bash", "-c", fmt.Sprintf("mkdir -p %s", kubeDir)); err != nil {
		return fmt.Errorf("failed to create .kube directory: %w (stderr: %s)", err, stderr)
	}

	// Copy admin.conf to user's .kube/config
	copyCmd := fmt.Sprintf("sudo cp -i /etc/kubernetes/admin.conf %s", kubeConfigFile)
	if _, stderr, err := utils.ExecCommandInteractive("", "bash", "-c", copyCmd); err != nil {
		return fmt.Errorf("failed to copy admin.conf: %w (stderr: %s)", err, stderr)
	}

	// Change ownership to the original user
	var chownCmd string
	sudoUser := os.Getenv("SUDO_USER")
	if sudoUser != "" {
		chownCmd = fmt.Sprintf("sudo chown $(id -u %s):$(id -g %s) %s", sudoUser, sudoUser, kubeConfigFile)
	} else {
		chownCmd = fmt.Sprintf("sudo chown $(id -u):$(id -g) %s", kubeConfigFile)
	}
	if _, stderr, err := utils.ExecCommandInteractive("", "bash", "-c", chownCmd); err != nil {
		return fmt.Errorf("failed to chown kubeconfig: %w (stderr: %s)", err, stderr)
	}
	return nil
}

func patchKubeProxy() error {
	// Patch the kube-proxy ConfigMap to disable conntrack tuning.
	// Even with privileged=true, newer containerd runtimes block writes to
	// /proc/sys/net/netfilter/nf_conntrack_max. Since configureKernel() already
	// sets nf_conntrack_max on the host via sysctl, kube-proxy doesn't need to.
	// Setting maxPerCore=0 tells kube-proxy to skip the conntrack sysctl write.
	stdout, _, err := utils.ExecCommand("", "kubectl", "--kubeconfig=/etc/kubernetes/admin.conf",
		"-n", "kube-system", "get", "cm", "kube-proxy", "-o", "jsonpath={.data.config\\.conf}")
	if err != nil {
		return fmt.Errorf("failed to get kube-proxy configmap: %w", err)
	}

	// Inject conntrack maxPerCore: 0 into the config
	var modifiedConfig string
	if strings.Contains(stdout, "conntrack:") {
		modifiedConfig = strings.Replace(stdout, "maxPerCore: null", "maxPerCore: 0", 1)
		if modifiedConfig == stdout {
			modifiedConfig = strings.Replace(stdout, "maxPerCore: 32768", "maxPerCore: 0", 1)
		}
		if modifiedConfig == stdout {
			modifiedConfig = strings.Replace(stdout, "conntrack:", "conntrack:\n  maxPerCore: 0", 1)
		}
	} else {
		modifiedConfig = stdout + "\nconntrack:\n  maxPerCore: 0\n"
	}

	// Escape config YAML for JSON embedding
	escapedConfig := strings.ReplaceAll(modifiedConfig, `\`, `\\`)
	escapedConfig = strings.ReplaceAll(escapedConfig, `"`, `\"`)
	escapedConfig = strings.ReplaceAll(escapedConfig, "\n", `\n`)
	escapedConfig = strings.ReplaceAll(escapedConfig, "\t", `\t`)
	cmPatch := fmt.Sprintf(`{"data":{"config.conf":"%s"}}`, escapedConfig)

	_, stderr, err := utils.ExecCommand("", "kubectl", "--kubeconfig=/etc/kubernetes/admin.conf",
		"patch", "configmap", "kube-proxy", "-n", "kube-system",
		"--type=merge", "-p", cmPatch)
	if err != nil {
		return fmt.Errorf("failed to patch kube-proxy configmap: %w (stderr: %s)", err, stderr)
	}
	fmt.Printf("%sPatched kube-proxy ConfigMap to disable conntrack tuning.\n", utils.PrefixOK)

	// Also ensure kube-proxy runs as privileged with hostPID
	dsPatch := `{"spec":{"template":{"spec":{"hostPID":true,"containers":[{"name":"kube-proxy","securityContext":{"privileged":true}}]}}}}`
	_, stderr, err = utils.ExecCommand("", "kubectl", "--kubeconfig=/etc/kubernetes/admin.conf",
		"patch", "daemonset", "kube-proxy", "-n", "kube-system",
		"--type=strategic", "-p", dsPatch)
	if err != nil {
		return fmt.Errorf("failed to patch kube-proxy security context: %w (stderr: %s)", err, stderr)
	}

	// Restart kube-proxy to pick up both changes
	_, stderr, err = utils.ExecCommand("", "kubectl", "--kubeconfig=/etc/kubernetes/admin.conf",
		"-n", "kube-system", "rollout", "restart", "daemonset/kube-proxy")
	if err != nil {
		fmt.Printf("%sFailed to restart kube-proxy: %s\n", utils.PrefixWarn, stderr)
	}

	// Wait for rollout to complete
	_, stderr, err = utils.ExecCommandInteractive("", "kubectl", "--kubeconfig=/etc/kubernetes/admin.conf",
		"rollout", "status", "daemonset/kube-proxy", "-n", "kube-system", "--timeout=2m")
	if err != nil {
		fmt.Printf("%skube-proxy rollout status check: %s\n", utils.PrefixWarn, stderr)
	}
	return nil
}

func untaintControlPlaneNode() error {
	_, stderr, err := utils.ExecCommand("", "kubectl", "--kubeconfig=/etc/kubernetes/admin.conf", "taint", "nodes", "--all", "node-role.kubernetes.io/control-plane-")
	if err != nil {
		// This may fail if the taint is already gone, which is not a critical error for a single-node setup.
		fmt.Printf("%sCould not untaint control-plane node (might already be untainted): %s\n", utils.PrefixWarn, stderr)
	}
	return nil
}

func installCalicoCNI() error {
	// Download the Calico manifest
	resp, err := http.Get("https://raw.githubusercontent.com/projectcalico/calico/v3.28.0/manifests/calico.yaml")
	if err != nil {
		return fmt.Errorf("failed to download Calico manifest: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read Calico manifest: %w", err)
	}

	// Enable and set CALICO_IPV4POOL_CIDR to match kubeadm's --pod-network-cidr.
	// In the default manifest these lines are commented out; we uncomment and set them.
	manifest := string(body)
	manifest = strings.ReplaceAll(manifest,
		"# - name: CALICO_IPV4POOL_CIDR",
		"- name: CALICO_IPV4POOL_CIDR")
	manifest = strings.ReplaceAll(manifest,
		"#   value: \"192.168.0.0/16\"",
		"  value: \"10.244.0.0/16\"")

	// Apply the modified manifest
	_, stderr, err := utils.ExecCommandWithStdin(manifest, "", "kubectl", "--kubeconfig=/etc/kubernetes/admin.conf", "apply", "-f", "-")
	if err != nil {
		return fmt.Errorf("failed to apply Calico manifest: %w (stderr: %s)", err, stderr)
	}
	return nil
}

func installMetricsServer() error {
	// Apply the metrics-server manifest
	_, stderr, err := utils.ExecCommand("", "kubectl", "--kubeconfig=/etc/kubernetes/admin.conf",
		"apply", "-f", "https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml")
	if err != nil {
		return fmt.Errorf("failed to apply metrics-server manifest: %w (stderr: %s)", err, stderr)
	}

	// Patch the deployment to add --kubelet-insecure-tls for kubeadm's self-signed certs.
	// Using kubectl patch is more reliable than string-replacing YAML (indentation-agnostic).
	patch := `[{"op":"add","path":"/spec/template/spec/containers/0/args/-","value":"--kubelet-insecure-tls"}]`
	_, stderr, err = utils.ExecCommand("", "kubectl", "--kubeconfig=/etc/kubernetes/admin.conf",
		"patch", "deployment", "metrics-server", "-n", "kube-system",
		"--type=json", "-p", patch)
	if err != nil {
		return fmt.Errorf("failed to patch metrics-server for insecure TLS: %w (stderr: %s)", err, stderr)
	}
	return nil
}

func installLocalPathProvisioner() error {
	_, stderr, err := utils.ExecCommand("", "kubectl", "--kubeconfig=/etc/kubernetes/admin.conf", "apply", "-f", "https://raw.githubusercontent.com/rancher/local-path-provisioner/v0.0.26/deploy/local-path-storage.yaml")
	if err != nil {
		return fmt.Errorf("failed to apply local-path-provisioner manifest: %w (stderr: %s)", err, stderr)
	}
	return nil
}

func waitForClusterReady() error {
	namespaces := []string{"kube-system"}
	for _, ns := range namespaces {
		fmt.Printf("%sWaiting for pods in namespace %s to become ready...\n", utils.PrefixInfo, ns)
		if err := waitForNamespace(ns); err != nil {
			return err
		}
		if err := waitForPodsReady(ns); err != nil {
			return err
		}
	}
	return nil
}

func waitForNamespace(namespace string) error {
	for i := 0; i < 60; i++ {
		_, stderr, err := utils.ExecCommand("", "kubectl", "--kubeconfig=/etc/kubernetes/admin.conf", "get", "namespace", namespace)
		if err == nil {
			return nil
		}
		if i%6 == 5 {
			fmt.Printf("%sStill waiting for namespace %s to appear...\n", utils.PrefixInfo, namespace)
		}
		if i == 59 {
			return fmt.Errorf("namespace %s did not appear in time: %w (stderr: %s)", namespace, err, stderr)
		}
		time.Sleep(5 * time.Second)
	}
	return nil
}

func waitForPodsReady(namespace string) error {
	// Poll pods until all non-terminal pods are Ready, with progress updates.
	// We use a polling approach rather than "kubectl wait --all" because:
	// 1. It shows which specific pods are still pending
	// 2. Job pods (Succeeded/Failed) are naturally excluded
	// 3. It gives clear timeout behavior with actionable output
	timeout := 3 * time.Minute
	start := time.Now()

	for {
		elapsed := time.Since(start)
		if elapsed > timeout {
			// On timeout, warn but don't fail - non-critical pods (metrics-server, etc.)
			// shouldn't block the entire installation
			stdout, _, _ := utils.ExecCommand("", "kubectl", "--kubeconfig=/etc/kubernetes/admin.conf",
				"get", "pods", "-n", namespace, "--field-selector=status.phase!=Succeeded,status.phase!=Failed",
				"--no-headers")
			fmt.Printf("%sSome pods in namespace %s are not yet ready after %s:\n%s\n", utils.PrefixWarn, namespace, timeout, stdout)
			fmt.Printf("%sProceeding with installation. These pods should become ready shortly.\n", utils.PrefixInfo)
			return nil
		}

		// Get all non-terminal pods and check their status
		stdout, _, err := utils.ExecCommand("", "kubectl", "--kubeconfig=/etc/kubernetes/admin.conf",
			"get", "pods", "-n", namespace,
			"--field-selector=status.phase!=Succeeded,status.phase!=Failed",
			"-o", "jsonpath={range .items[*]}{.metadata.name}:{.status.conditions[?(@.type==\"Ready\")].status} {end}")
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		if stdout == "" {
			// No non-terminal pods (all completed or none exist)
			return nil
		}

		// Parse pod:status pairs to find unready pods
		var notReady []string
		pods := strings.Fields(stdout)
		for _, pod := range pods {
			parts := strings.SplitN(pod, ":", 2)
			if len(parts) == 2 && parts[1] != "True" {
				notReady = append(notReady, parts[0])
			}
		}

		if len(notReady) == 0 {
			return nil
		}

		// Print progress every 15 seconds
		if int(elapsed.Seconds())%15 == 0 && elapsed.Seconds() > 1 {
			fmt.Printf("%sWaiting for %d pod(s): %s\n", utils.PrefixInfo, len(notReady), strings.Join(notReady, ", "))
		}

		time.Sleep(5 * time.Second)
	}
}
