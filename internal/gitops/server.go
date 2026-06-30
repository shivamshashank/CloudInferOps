package gitops

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/shivamshashank/CloudInferOps/internal/utils"
)

// GitServerManifests returns the deployment and service YAML manifests for cloudinferops-git-server.
func GitServerManifests(ns string) string {
	return fmt.Sprintf(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: cloudinferops-git-server
  namespace: %s
  labels:
    app: cloudinferops-git-server
spec:
  replicas: 1
  selector:
    matchLabels:
      app: cloudinferops-git-server
  template:
    metadata:
      labels:
        app: cloudinferops-git-server
    spec:
      containers:
      - name: git-server
        image: alpine/git:latest
        imagePullPolicy: IfNotPresent
        command:
        - sh
        - -c
        - |
          mkdir -p /git/gitops.git
          cd /git/gitops.git
          git init --bare --shared=all
          git config receive.denyCurrentBranch ignore
          git daemon --verbose --export-all --enable=receive-pack --reuseaddr --base-path=/git /git
        ports:
        - containerPort: 9418
          name: git
        volumeMounts:
        - name: git-storage
          mountPath: /git
      volumes:
      - name: git-storage
        emptyDir: {}
---
apiVersion: v1
kind: Service
metadata:
  name: cloudinferops-git-server
  namespace: %s
spec:
  selector:
    app: cloudinferops-git-server
  ports:
  - port: 9418
    targetPort: 9418
    name: git
  type: ClusterIP
`, ns, ns)
}

// DeployGitServer applies the Git server manifest to the Kubernetes cluster.
func DeployGitServer(ns string, dryRun bool) error {
	if dryRun {
		fmt.Printf("%s[DRY-RUN] Deploying Git Server in namespace '%s'\n", utils.PrefixInfo, ns)
		return nil
	}

	manifest := GitServerManifests(ns)
	tmpPath := filepath.Join(os.TempDir(), "cloudinferops-git-server.yaml")
	if err := os.WriteFile(tmpPath, []byte(manifest), 0600); err != nil {
		return fmt.Errorf("failed to write git server manifest: %w", err)
	}
	defer func() { _ = os.Remove(tmpPath) }()

	_, stderr, err := utils.ExecCommand("", "kubectl", "apply", "-f", tmpPath)
	if err != nil {
		return fmt.Errorf("failed to apply git server deployment: %w (stderr: %s)", err, stderr)
	}

	fmt.Printf("%sWaiting for cloudinferops-git-server to be ready...\n", utils.PrefixInfo)
	for i := 0; i < 30; i++ {
		_, _, waitErr := utils.ExecCommand("", "kubectl", "wait", "--namespace", ns,
			"--for=condition=Ready", "pod",
			"-l", "app=cloudinferops-git-server",
			"--timeout=10s")
		if waitErr == nil {
			fmt.Printf("%sGit Server is ready.\n", utils.PrefixOK)
			return nil
		}
		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("git server failed to become ready in time")
}

// StartPortForward sets up a background port-forward for the git-server and returns a cancel function.
func StartPortForward(ns string, localPort, targetPort int) (func(), error) {
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, "kubectl", "port-forward", "-n", ns, "svc/cloudinferops-git-server", fmt.Sprintf("%d:%d", localPort, targetPort)) //nolint:gosec

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start port-forward process: %w", err)
	}

	// Give it a moment to bind and run a quick connectivity check
	time.Sleep(2 * time.Second)

	cleanup := func() {
		cancel()
		_ = cmd.Wait()
	}

	return cleanup, nil
}

// IsGitServerRunning checks if the cloudinferops-git-server deployment exists in the cluster.
func IsGitServerRunning(ns string) bool {
	out, _, err := utils.ExecCommand("", "kubectl", "get", "deployment", "cloudinferops-git-server", "-n", ns, "-o", "jsonpath={.status.readyReplicas}")
	if err != nil {
		return false
	}
	return strings.TrimSpace(out) != "" && strings.TrimSpace(out) != "0"
}
