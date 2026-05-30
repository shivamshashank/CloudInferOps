package webhook

import (
	"fmt"

	"github.com/shivamshashank/StackPulse/internal/config"
	"github.com/shivamshashank/StackPulse/internal/utils"
)

// DeployWebhookHandler renders and applies Kubernetes manifests for the incident processing gateway service
func DeployWebhookHandler(dryRun bool) error {
	ns := config.GlobalConfig.Namespace
	if ns == "" {
		ns = "observability"
	}

	fmt.Printf("%sStarting Custom Go Webhook Handler Deployment...\n", utils.PrefixInfo)
	if dryRun {
		fmt.Printf("%sRunning in [DRY-RUN] mode. No changes will be made to your cluster.\n\n", utils.PrefixInfo)
	}

	// Service manifest
	serviceYaml := fmt.Sprintf(`apiVersion: v1
kind: Service
metadata:
  name: stackpulse-webhook-handler
  namespace: %s
spec:
  ports:
  - port: 80
    targetPort: 8080
    protocol: TCP
  selector:
    app: stackpulse-webhook-handler
  type: ClusterIP`, ns)

	// Deployment manifest
	deploymentYaml := fmt.Sprintf(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: stackpulse-webhook-handler
  namespace: %s
spec:
  replicas: 1
  selector:
    matchLabels:
      app: stackpulse-webhook-handler
  template:
    metadata:
      labels:
        app: stackpulse-webhook-handler
    spec:
      containers:
      - name: webhook-handler
        image: ghcr.io/shivamshashank/stackpulse-webhook-handler:latest
        imagePullPolicy: Always
        ports:
        - containerPort: 8080
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
        env:
        - name: SLACK_WEBHOOK_URL
          valueFrom:
            secretKeyRef:
              name: stackpulse-slack-webhook
              key: webhook-url
              optional: true
        - name: PAGERDUTY_INTEGRATION_KEY
          valueFrom:
            secretKeyRef:
              name: stackpulse-pagerduty-key
              key: integration-key
              optional: true`, ns)

	manifests := fmt.Sprintf("%s\n---\n%s", serviceYaml, deploymentYaml)

	if dryRun {
		fmt.Printf("%s[DRY-RUN] kubectl apply -f -\n", utils.PrefixInfo)
		fmt.Println("---")
		fmt.Println(manifests)
		fmt.Println("---")
		return nil
	}

	// Execute kubectl apply pipeline
	fmt.Printf("%sApplying Custom Webhook manifests into Kubernetes namespace '%s'...\n", utils.PrefixInfo, ns)

	// Create pipe command
	_, stderr, err := utils.ExecCommand("", "sh", "-c", fmt.Sprintf("echo '%s' | kubectl apply -f -", manifests))
	if err != nil {
		return fmt.Errorf("failed to apply Custom Webhook manifests: %w (stderr: %s)", err, stderr)
	}

	fmt.Printf("%sSuccessfully deployed Custom Webhook Handler Gateway.\n", utils.PrefixOK)
	fmt.Println("-----------------------------------------------------------------")
	fmt.Printf("🌐  Service Name:  stackpulse-webhook-handler\n")
	fmt.Printf("📡  Endpoints:     GET  /health\n")
	fmt.Printf("                   POST /webhook/alertmanager\n")
	fmt.Printf("                   GET  /incidents\n")
	fmt.Println("-----------------------------------------------------------------")

	return nil
}
