package k8s

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	spaws "github.com/yourusername/stackpulse/internal/aws"
	"github.com/yourusername/stackpulse/internal/ingress"
	"github.com/yourusername/stackpulse/internal/prometheus"
)

func DeployEC2(ssmClient *ssm.SSM, instanceID string, ip string) {
	fmt.Println("\n[K8s] Checking Kubernetes status via SSM...")

	checkCommands := []*string{
		aws.String("export PATH=$PATH:/usr/local/bin"),
		// Dynamically auto-discover the Kubernetes config on the EC2 instance
		aws.String("for conf in /etc/rancher/k3s/k3s.yaml /etc/kubernetes/admin.conf /root/.kube/config /home/ubuntu/.kube/config /home/ec2-user/.kube/config; do if [ -r \"$conf\" ]; then export KUBECONFIG=\"$conf\"; break; fi; done"),
		aws.String("which kubectl || echo 'K8S_NOT_FOUND'"),
		aws.String("kubectl get nodes || echo 'K8S_NOT_RUNNING'"),
	}

	cmdID, err := sendSSMCommand(ssmClient, instanceID, checkCommands)
	if err != nil {
		if strings.Contains(err.Error(), "InvalidInstanceId") {
			fmt.Printf("❌ Failed to send command: %v\n", err)
			fmt.Println("\n💡 HINT: Your EC2 instance is not ready for AWS Systems Manager (SSM).")
			fmt.Println("⚙️  Attempting to automatically create and attach the necessary IAM Role...")

			region := *ssmClient.Config.Region
			if attachErr := spaws.EnsureSSMRoleAttached(region, instanceID); attachErr != nil {
				fmt.Printf("❌ Failed to attach IAM role automatically: %v\n", attachErr)
				fmt.Println("👉 Please attach it manually in the AWS Console and ensure the SSM Agent is running.")
				return
			}

			fmt.Println("--> 🔄 Retrying SSM Kubernetes check...")
			cmdID, err = sendSSMCommand(ssmClient, instanceID, checkCommands)
			if err != nil {
				fmt.Printf("❌ Failed to send command after retry: %v\n", err)
				return
			}
		} else {
			fmt.Printf("❌ Failed to send command: %v\n", err)
			return
		}
	}

	output, err := getCommandOutput(ssmClient, cmdID, instanceID)
	if err != nil {
		fmt.Printf("❌ Failed to fetch output: %v\n", err)
		return
	}

	// 🧠 Smart parsing
	if contains(output, "K8S_NOT_FOUND") || contains(output, "K8S_NOT_RUNNING") {
		fmt.Println("\n❌ Kubernetes is not installed or not running on this EC2 instance.")
		fmt.Print("💡 Please install Kubernetes (e.g., k3s) before running StackPulse.\n\n")
		return
	}

	fmt.Println("\n✅ Kubernetes is running. Proceeding with deployment...")

	type EC2InstallStep struct {
		Name     string
		Commands []string
	}

	steps := []EC2InstallStep{
		{
			Name: "Helm & Dependencies Setup",
			Commands: []string{
				"which helm || curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash",
				"helm repo add prometheus-community https://prometheus-community.github.io/helm-charts",
				"helm repo update",
				"kubectl delete secret -n observability -l status=pending-install 2>/dev/null || true",
				"kubectl delete secret -n observability -l status=pending-upgrade 2>/dev/null || true",
				"kubectl delete validatingwebhookconfiguration ingress-nginx-admission 2>/dev/null || true",
				"helm uninstall ingress-nginx -n ingress-nginx 2>/dev/null || true",
			},
		},
		{
			Name:     "NGINX Ingress Controller",
			Commands: ingress.GetEC2InstallCommands(),
		},
		{
			Name:     "Prometheus Observability Stack",
			Commands: prometheus.GetEC2InstallCommands(ip),
		},
	}

	for _, step := range steps {
		fmt.Printf("\n[K8s] 🚀 Starting Phase: %s...\n", step.Name)

		deployCommands := []*string{
			aws.String("export PATH=$PATH:/usr/local/bin"),
			// Dynamically auto-discover the Kubernetes config on the EC2 instance
			aws.String("for conf in /etc/rancher/k3s/k3s.yaml /etc/kubernetes/admin.conf /root/.kube/config /home/ubuntu/.kube/config /home/ec2-user/.kube/config; do if [ -r \"$conf\" ]; then export KUBECONFIG=\"$conf\"; break; fi; done"),
		}

		for _, cmd := range step.Commands {
			deployCommands = append(deployCommands, aws.String(cmd))
		}

		deployCmdID, err := sendSSMCommand(ssmClient, instanceID, deployCommands)
		if err != nil {
			fmt.Printf("\n❌ Failed to send deployment command for %s: %v\n", step.Name, err)
			return
		}

		output, err = getCommandOutput(ssmClient, deployCmdID, instanceID)
		if err != nil {
			fmt.Printf("\n❌ Failed to fetch deployment output for %s: %v\n", step.Name, err)
			return
		}

		fmt.Printf("\n--- Output: %s ---\n%s\n-------------------------\n", step.Name, output)
	}

	fmt.Println("\n✅ Observability stack deployed successfully!")
	fmt.Println("\n🔗 Ingress Access Links:")
	fmt.Printf("   ▶ Prometheus Server : http://prometheus.%s.nip.io\n", ip)
	fmt.Printf("   ▶ Pushgateway       : http://pushgateway.%s.nip.io\n", ip)
	fmt.Printf("   ▶ Alertmanager      : http://alertmanager.%s.nip.io\n", ip)
	fmt.Print("\n💡 Note: Using nip.io for automatic DNS resolution. You don't need to configure your hosts file!\n\n")
	fmt.Print("⚠️  IMPORTANT: Ensure Port 80 (HTTP) is OPEN in your EC2 instance's AWS Security Group, otherwise these links will time out!\n\n")
}

func sendSSMCommand(client *ssm.SSM, instanceID string, commands []*string) (string, error) {
	input := &ssm.SendCommandInput{
		InstanceIds:  []*string{aws.String(instanceID)},
		DocumentName: aws.String("AWS-RunShellScript"),
		Parameters: map[string][]*string{
			"commands": commands,
		},
	}

	result, err := client.SendCommand(input)
	if err != nil {
		return "", err
	}

	return *result.Command.CommandId, nil
}

func getCommandOutput(client *ssm.SSM, commandID, instanceID string) (string, error) {
	input := &ssm.GetCommandInvocationInput{
		CommandId:  aws.String(commandID),
		InstanceId: aws.String(instanceID),
	}

	fmt.Print("⏳ Waiting for SSM command to finish")

	// ANSI Color Codes
	colorReset := "\033[0m"
	colorRed := "\033[31m"
	colorYellow := "\033[33m"
	colorCyan := "\033[36m"

	for {
		result, err := client.GetCommandInvocation(input)
		if err != nil {
			// SSM commands take a moment to register, so we ignore the NotExist error briefly
			if !strings.Contains(err.Error(), "InvocationDoesNotExist") {
				return "", err
			}
		} else if result.Status != nil {
			status := *result.Status
			if status == ssm.CommandInvocationStatusSuccess ||
				status == ssm.CommandInvocationStatusFailed ||
				status == ssm.CommandInvocationStatusCancelled ||
				status == ssm.CommandInvocationStatusTimedOut {

				fmt.Println(" Done!")
				var buffer bytes.Buffer
				if result.StandardOutputContent != nil && strings.TrimSpace(*result.StandardOutputContent) != "" {
					buffer.WriteString(colorCyan + *result.StandardOutputContent + colorReset)
				}
				if result.StandardErrorContent != nil && strings.TrimSpace(*result.StandardErrorContent) != "" {
					if status == ssm.CommandInvocationStatusFailed {
						buffer.WriteString("\n\n" + colorRed + "--- ERROR LOGS ---" + colorReset + "\n")
						buffer.WriteString(colorRed + strings.TrimSpace(*result.StandardErrorContent) + colorReset)
					} else {
						buffer.WriteString("\n\n" + colorYellow + "--- STANDARD ERROR (Info/Warnings) ---" + colorReset + "\n")
						buffer.WriteString(colorYellow + strings.TrimSpace(*result.StandardErrorContent) + colorReset)
					}
				}
				return strings.TrimSpace(buffer.String()), nil
			}
		}
		fmt.Print(".")
		time.Sleep(3 * time.Second)
	}
}

func contains(output, keyword string) bool {
	return strings.Contains(output, keyword)
}
