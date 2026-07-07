package ui

import (
	"fmt"
	"strings"

	"github.com/shivamshashank/CloudInferOps/internal/utils"
)

func DeploymentManifest(image string, enableActions bool) string {
	if strings.TrimSpace(image) == "" {
		image = "ghcr.io/shivamshashank/cloudinferops-ui:latest"
	}
	actions := "0"
	if enableActions {
		actions = "1"
	}
	return fmt.Sprintf(`apiVersion: v1
kind: ServiceAccount
metadata:
  name: cloudinferops-ui
  namespace: observability
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cloudinferops-ui
rules:
  - apiGroups: [""]
    resources: ["pods", "pods/log", "services", "configmaps", "namespaces"]
    verbs: ["get", "list", "watch", "create", "update", "patch"]
  - apiGroups: ["apps"]
    resources: ["deployments", "deployments/scale", "daemonsets", "statefulsets"]
    verbs: ["get", "list", "watch", "create", "update", "patch"]
  - apiGroups: ["networking.k8s.io"]
    resources: ["ingresses"]
    verbs: ["get", "list", "watch", "create", "update", "patch"]
  - apiGroups: ["monitoring.coreos.com"]
    resources: ["prometheusrules", "servicemonitors"]
    verbs: ["get", "list", "watch", "create", "update", "patch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: cloudinferops-ui
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cloudinferops-ui
subjects:
  - kind: ServiceAccount
    name: cloudinferops-ui
    namespace: observability
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cloudinferops-ui
  namespace: observability
spec:
  replicas: 1
  selector:
    matchLabels: {app: cloudinferops-ui}
  template:
    metadata:
      labels: {app: cloudinferops-ui}
    spec:
      serviceAccountName: cloudinferops-ui
      securityContext: {runAsNonRoot: true, runAsUser: 65532, seccompProfile: {type: RuntimeDefault}}
      containers:
        - name: portal
          image: %s
          imagePullPolicy: IfNotPresent
          ports: [{name: http, containerPort: 8080}]
          env:
            - {name: PORT, value: "8080"}
            - {name: CLOUDINFEROPS_NAMESPACE, value: observability}
            - {name: CLOUDINFEROPS_UI_ENABLE_ACTIONS, value: %q}
          readinessProbe: {httpGet: {path: /api/health, port: http}, initialDelaySeconds: 3, periodSeconds: 10}
          livenessProbe: {httpGet: {path: /api/health, port: http}, initialDelaySeconds: 10, periodSeconds: 20}
          resources:
            requests: {cpu: 50m, memory: 64Mi}
            limits: {cpu: 500m, memory: 256Mi}
          securityContext: {allowPrivilegeEscalation: false, readOnlyRootFilesystem: true, capabilities: {drop: ["ALL"]}}
---
apiVersion: v1
kind: Service
metadata:
  name: cloudinferops-ui
  namespace: observability
spec:
  selector: {app: cloudinferops-ui}
  ports: [{name: http, port: 80, targetPort: http}]
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: cloudinferops-ui
  namespace: observability
spec:
  ingressClassName: nginx
  rules:
    - http:
        paths:
          - path: /cloudinferops
            pathType: Prefix
            backend:
              service: {name: cloudinferops-ui, port: {number: 80}}
`, image, actions)
}

func DeployPortal(image string, enableActions, dryRun bool) error {
	manifest := DeploymentManifest(image, enableActions)
	if dryRun {
		fmt.Printf("%s[DRY-RUN] CloudInferOps portal resources:\n%s\n", utils.PrefixInfo, manifest)
		return nil
	}
	if _, stderr, err := utils.ExecCommand("", "kubectl", "create", "namespace", "observability"); err != nil && !strings.Contains(stderr, "AlreadyExists") {
		return fmt.Errorf("create observability namespace: %w (%s)", err, stderr)
	}
	if _, stderr, err := utils.ExecCommandWithStdin(manifest, "", "kubectl", "apply", "-f", "-"); err != nil {
		return fmt.Errorf("deploy portal: %w (%s)", err, stderr)
	}
	if _, stderr, err := utils.ExecCommand("", "kubectl", "rollout", "status", "deployment/cloudinferops-ui", "-n", "observability", "--timeout=180s"); err != nil {
		return fmt.Errorf("wait for portal: %w (%s)", err, stderr)
	}
	fmt.Printf("%sPortal ready at /cloudinferops/\n", utils.PrefixOK)
	return nil
}
