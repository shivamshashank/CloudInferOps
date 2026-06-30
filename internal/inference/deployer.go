package inference

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/shivamshashank/CloudInferOps/internal/utils"
)

// NamespaceManifest returns the inference namespace YAML
func NamespaceManifest() string {
	return `apiVersion: v1
kind: Namespace
metadata:
  name: inference
  labels:
    name: inference
`
}

// ModelConfigManifest returns the configuration map YAML populated with parameters
func ModelConfigManifest(provider, model string) string {
	ollamaHost := "http://ollama-service:11434"
	return fmt.Sprintf(`apiVersion: v1
kind: ConfigMap
metadata:
  name: model-config
  namespace: inference
data:
  provider: %q
  model: %q
  ollama_host: %q
`, provider, model, ollamaHost)
}

// GatewayDeploymentManifest returns the gateway deployment YAML
func GatewayDeploymentManifest() string {
	return `apiVersion: apps/v1
kind: Deployment
metadata:
  name: gateway-deployment
  namespace: inference
  labels:
    app: cloudinferops-gateway
spec:
  replicas: 1
  selector:
    matchLabels:
      app: cloudinferops-gateway
  template:
    metadata:
      labels:
        app: cloudinferops-gateway
    spec:
      containers:
      - name: gateway
        image: ghcr.io/shivamshashank/cloudinferops-gateway:latest
        imagePullPolicy: Always
        ports:
        - containerPort: 8000
          name: http
        env:
        - name: PROVIDER
          valueFrom:
            configMapKeyRef:
              name: model-config
              key: provider
        - name: OLLAMA_HOST
          valueFrom:
            configMapKeyRef:
              name: model-config
              key: ollama_host
        - name: OTEL_EXPORTER_OTLP_ENDPOINT
          value: "http://cloudinferops-otel.observability.svc.cluster.local:4318/v1/traces"
        livenessProbe:
          httpGet:
            path: /health
            port: 8000
          initialDelaySeconds: 5
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8000
          initialDelaySeconds: 5
          periodSeconds: 10
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 500m
            memory: 512Mi
`
}

// GatewayServiceManifest returns the gateway service YAML
func GatewayServiceManifest() string {
	return `apiVersion: v1
kind: Service
metadata:
  name: gateway-service
  namespace: inference
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/port: "8000"
    prometheus.io/path: "/metrics"
  labels:
    app: cloudinferops-gateway
spec:
  ports:
  - port: 8000
    targetPort: 8000
    protocol: TCP
    name: http
  selector:
    app: cloudinferops-gateway
  type: ClusterIP
`
}

// GatewayIngressManifest returns the gateway ingress YAML
func GatewayIngressManifest() string {
	return `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: gateway-ingress
  namespace: inference
  annotations:
    kubernetes.io/ingress.class: nginx
    nginx.ingress.kubernetes.io/ssl-redirect: "false"
spec:
  ingressClassName: nginx
  rules:
  - http:
      paths:
      - path: /v1
        pathType: Prefix
        backend:
          service:
            name: gateway-service
            port:
              number: 8000
      - path: /models
        pathType: Prefix
        backend:
          service:
            name: gateway-service
            port:
              number: 8000
`
}

// OllamaDeploymentManifest returns the Ollama deployment YAML
func OllamaDeploymentManifest() string {
	return `apiVersion: apps/v1
kind: Deployment
metadata:
  name: ollama-deployment
  namespace: inference
  labels:
    app: ollama
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ollama
  template:
    metadata:
      labels:
        app: ollama
    spec:
      containers:
      - name: ollama
        image: ollama/ollama:latest
        imagePullPolicy: IfNotPresent
        ports:
        - containerPort: 11434
          name: api
        volumeMounts:
        - name: ollama-storage
          mountPath: /root/.ollama
        resources:
          requests:
            cpu: 1000m
            memory: 2Gi
          limits:
            cpu: 4000m
            memory: 4Gi
      volumes:
      - name: ollama-storage
        emptyDir: {}
`
}

// OllamaServiceManifest returns the Ollama service YAML
func OllamaServiceManifest() string {
	return `apiVersion: v1
kind: Service
metadata:
  name: ollama-service
  namespace: inference
  labels:
    app: ollama
spec:
  ports:
  - port: 11434
    targetPort: 11434
    protocol: TCP
    name: api
  selector:
    app: ollama
  type: ClusterIP
`
}

// DeployInference orchestrates deployment of the inference stack
func DeployInference(provider, model string, dryRun bool) error {
	manifests := []struct {
		name string
		yaml string
	}{
		{"namespace", NamespaceManifest()},
		{"model-config", ModelConfigManifest(provider, model)},
		{"ollama-deployment", OllamaDeploymentManifest()},
		{"ollama-service", OllamaServiceManifest()},
		{"gateway-deployment", GatewayDeploymentManifest()},
		{"gateway-service", GatewayServiceManifest()},
		{"gateway-ingress", GatewayIngressManifest()},
	}

	if dryRun {
		fmt.Printf("%s[DRY-RUN] Planned resources for namespace 'inference':\n", utils.PrefixInfo)
		fmt.Printf("%s[DRY-RUN] Provider: %s, Model: %s\n", utils.PrefixInfo, provider, model)
		for _, m := range manifests {
			fmt.Printf("--- \n# Manifest: %s.yaml\n%s\n", m.name, m.yaml)
		}
		return nil
	}

	for _, m := range manifests {
		fmt.Printf("%sApplying manifest: %s.yaml...\n", utils.PrefixInfo, m.name)
		tmpPath := filepath.Join(os.TempDir(), fmt.Sprintf("cloudinferops-inference-%s.yaml", m.name))
		if err := os.WriteFile(tmpPath, []byte(m.yaml), 0600); err != nil {
			return fmt.Errorf("failed to write temporary manifest %s: %w", m.name, err)
		}
		defer func(path string) { _ = os.Remove(path) }(tmpPath)

		_, stderr, err := utils.ExecCommand("", "kubectl", "apply", "-f", tmpPath)
		if err != nil {
			return fmt.Errorf("failed to apply manifest %s: %w (stderr: %s)", m.name, err, stderr)
		}
	}

	fmt.Printf("%sInference stack successfully deployed!\n", utils.PrefixOK)
	return nil
}
