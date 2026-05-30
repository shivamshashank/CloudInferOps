package observability

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/shivamshashank/StackPulse/internal/config"
	"github.com/shivamshashank/StackPulse/internal/utils"
)

// FetchIngressIP polls the cluster to retrieve the provisioned Ingress LoadBalancer IP/Hostname
func FetchIngressIP(ns string, dryRun bool) (string, error) {
	if dryRun {
		return "127.0.0.1", nil
	}

	fmt.Printf("%sResolving Ingress Controller IP...\n", utils.PrefixInfo)

	// Strategy 1: Get IP from the NGINX Ingress Controller Service (most reliable)
	controllerSvc := "stackpulse-ingress-nginx-controller"
	for i := 0; i < 6; i++ {
		// Try LoadBalancer IP
		ip, _, err := utils.ExecCommand("", "kubectl", "get", "svc", controllerSvc, "-n", ns, "-o", "jsonpath={.status.loadBalancer.ingress[0].ip}")
		if err == nil && ip != "" {
			return strings.TrimSpace(ip), nil
		}
		// Try LoadBalancer Hostname (AWS EKS / cloud)
		host, _, err := utils.ExecCommand("", "kubectl", "get", "svc", controllerSvc, "-n", ns, "-o", "jsonpath={.status.loadBalancer.ingress[0].hostname}")
		if err == nil && host != "" {
			return strings.TrimSpace(host), nil
		}
		// Try ExternalIP
		extIP, _, err := utils.ExecCommand("", "kubectl", "get", "svc", controllerSvc, "-n", ns, "-o", "jsonpath={.spec.externalIPs[0]}")
		if err == nil && extIP != "" {
			return strings.TrimSpace(extIP), nil
		}
		fmt.Printf("%sIngress IP not assigned yet, retrying in 5 seconds...\n", utils.PrefixInfo)
		time.Sleep(5 * time.Second)
	}

	// Strategy 2: Get IP from the Ingress resource status
	ingressName := "stackpulse-prometheus-grafana"
	ip, _, err := utils.ExecCommand("", "kubectl", "get", "ingress", ingressName, "-n", ns, "-o", "jsonpath={.status.loadBalancer.ingress[0].ip}")
	if err == nil && ip != "" {
		return strings.TrimSpace(ip), nil
	}

	return "", fmt.Errorf("ingress IP provisioning timed out")
}

// UpdateHostsFile idempotently maps the Ingress IP to 'grafana.local' in local /etc/hosts
func UpdateHostsFile(ip, host string) error {
	hostsPath := "/etc/hosts"

	// 1. Read existing world-readable hosts file
	data, err := os.ReadFile(hostsPath)
	if err != nil {
		return fmt.Errorf("failed to read /etc/hosts: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	var newLines []string

	// 2. Filter out any existing matching hosts lines
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			newLines = append(newLines, line)
			continue
		}
		// Skip lines mapping our target host
		fields := strings.Fields(trimmed)
		if len(fields) > 1 && fields[1] == host {
			continue
		}
		newLines = append(newLines, line)
	}

	// 3. Append the new clean mapping
	newLines = append(newLines, fmt.Sprintf("%-16s %s", ip, host))
	newContent := strings.Join(newLines, "\n")

	// 4. Write to local temporary file in StackPulse directory
	configDir, err := config.GetConfigDir()
	if err != nil {
		return err
	}
	tempPath := filepath.Join(configDir, "hosts.tmp")
	if writeErr := os.WriteFile(tempPath, []byte(newContent), 0644); writeErr != nil {
		return fmt.Errorf("failed to write temp hosts file: %w", writeErr)
	}

	// 5. Use sudo cp to copy back to /etc/hosts
	fmt.Printf("%sStackPulse will now update your local /etc/hosts file.\n", utils.PrefixInfo)
	fmt.Printf("%sAdministrative privileges required. Requesting sudo password...\n", utils.PrefixInfo)

	var stderr string
	if os.Getuid() == 0 {
		_, stderr, err = utils.ExecCommand("", "cp", tempPath, hostsPath)
	} else {
		_, stderr, err = utils.ExecCommand("", "sudo", "cp", tempPath, hostsPath)
	}

	// Clean up temporary file
	_ = os.Remove(tempPath)

	if err != nil {
		return fmt.Errorf("failed to copy back to /etc/hosts: %w (stderr: %s)", err, stderr)
	}

	return nil
}
