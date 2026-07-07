package ui

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

type commandRunner func(context.Context, string, ...string) ([]byte, error)

type Service struct {
	version        string
	run            commandRunner
	httpClient     *http.Client
	actionsEnabled bool
	benchmarkMu    sync.RWMutex
	benchmark      *BenchmarkResult
}

func NewService() *Service {
	return NewServiceWithRunner(func(ctx context.Context, name string, args ...string) ([]byte, error) {
		return exec.CommandContext(ctx, name, args...).CombinedOutput()
	})
}

func NewServiceWithRunner(run commandRunner) *Service {
	return &Service{
		version: os.Getenv("CLOUDINFEROPS_VERSION"), run: run,
		httpClient:     &http.Client{Timeout: 2 * time.Second},
		actionsEnabled: os.Getenv("CLOUDINFEROPS_UI_ENABLE_ACTIONS") == "1",
	}
}

func (s *Service) kubectl(args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return s.run(ctx, "kubectl", args...)
}

func (s *Service) namespace() string {
	if value := strings.TrimSpace(os.Getenv("CLOUDINFEROPS_NAMESPACE")); value != "" {
		return value
	}
	return "observability"
}

func (s *Service) Overview() Overview {
	deployments := s.Deployments()
	models := s.Models()
	alerts := s.Alerts()
	cluster, pods := "Disconnected", 0
	if raw, err := s.kubectl("get", "pods", "-A", "-o", "json"); err == nil {
		var list struct {
			Items []json.RawMessage `json:"items"`
		}
		if json.Unmarshal(raw, &list) == nil {
			cluster, pods = "Connected", len(list.Items)
		}
	}
	gateway, observability := "Not deployed", "Not deployed"
	obsReady, obsTotal := 0, 0
	for _, item := range deployments {
		if item.Name == "gateway-deployment" {
			gateway = item.Status
		}
		if item.Namespace == s.namespace() {
			obsTotal++
			if item.Status == "Running" {
				obsReady++
			}
		}
	}
	if obsTotal > 0 {
		observability = "Partial"
		if obsReady == obsTotal {
			observability = "Healthy"
		}
	}
	version := s.version
	if version == "" {
		version = "dev"
	}
	return Overview{Cluster: cluster, Namespace: s.namespace(), Gateway: gateway, Observability: observability,
		Models: len(models), Alerts: len(alerts), Pods: pods, LastUpdated: time.Now().UTC().Format(time.RFC3339),
		Version: version, Actions: s.actionsEnabled}
}

func (s *Service) Deployments() []DeploymentStatus {
	raw, err := s.kubectl("get", "deployments", "-A", "-o", "json")
	if err != nil {
		return []DeploymentStatus{}
	}
	var list struct {
		Items []struct {
			Metadata struct{ Name, Namespace, CreationTimestamp string } `json:"metadata"`
			Spec     struct {
				Replicas int `json:"replicas"`
			} `json:"spec"`
			Status struct{ ReadyReplicas, AvailableReplicas int } `json:"status"`
		} `json:"items"`
	}
	if json.Unmarshal(raw, &list) != nil {
		return []DeploymentStatus{}
	}
	items := make([]DeploymentStatus, 0, len(list.Items))
	for _, item := range list.Items {
		status := "Pending"
		if item.Spec.Replicas > 0 && item.Status.AvailableReplicas == item.Spec.Replicas {
			status = "Running"
		}
		items = append(items, DeploymentStatus{Name: item.Metadata.Name, Namespace: item.Metadata.Namespace,
			Status: status, Replicas: fmt.Sprintf("%d/%d", item.Status.ReadyReplicas, item.Spec.Replicas), LastUpdated: item.Metadata.CreationTimestamp})
	}
	return items
}

func (s *Service) Models() []ModelStatus {
	base := strings.TrimRight(envOr("CLOUDINFEROPS_GATEWAY_URL", "http://gateway-service.inference.svc.cluster.local:8000"), "/")
	resp, err := s.httpClient.Get(base + "/models")
	if err != nil {
		return []ModelStatus{}
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return []ModelStatus{}
	}
	var source []struct {
		ID string `json:"id"`
	}
	if json.NewDecoder(resp.Body).Decode(&source) != nil {
		return []ModelStatus{}
	}
	provider := envOr("CLOUDINFEROPS_PROVIDER", "ollama")
	result := make([]ModelStatus, 0, len(source))
	for _, model := range source {
		result = append(result, ModelStatus{Name: model.ID, Provider: provider, Status: "Available", Location: "Cluster"})
	}
	return result
}

func (s *Service) Alerts() []AlertItem {
	url := strings.TrimRight(envOr("CLOUDINFEROPS_INCIDENTS_URL", "http://cloudinferops-webhook-handler.observability.svc.cluster.local"), "/") + "/incidents"
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return []AlertItem{}
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return []AlertItem{}
	}
	var raw []struct{ Name, Severity, Status, StartsAt string }
	if json.NewDecoder(resp.Body).Decode(&raw) != nil {
		return []AlertItem{}
	}
	items := make([]AlertItem, 0, len(raw))
	for _, item := range raw {
		items = append(items, AlertItem{Title: item.Name, Severity: item.Severity, Status: item.Status, Timestamp: item.StartsAt})
	}
	return items
}

func (s *Service) Observability() []ComponentStatus {
	components := []string{"grafana", "prometheus", "loki", "tempo", "alertmanager", "otel", "argocd"}
	pods := s.Pods()
	result := make([]ComponentStatus, 0, len(components))
	for _, component := range components {
		status := "Not deployed"
		for _, pod := range pods {
			if pod.Namespace == s.namespace() && strings.Contains(strings.ToLower(pod.Name), component) {
				if pod.Status == "Running" {
					status = "Running"
					break
				} else if pod.Status == "Pending" {
					status = "Pending"
				} else if status == "Not deployed" {
					status = pod.Status
				}
			}
		}
		result = append(result, ComponentStatus{Name: component, Namespace: s.namespace(), Status: status})
	}
	return result
}

func (s *Service) Pods() []PodStatus {
	raw, err := s.kubectl("get", "pods", "-A", "-o", "json")
	if err != nil {
		return []PodStatus{}
	}
	var list struct {
		Items []struct {
			Metadata struct{ Name, Namespace string } `json:"metadata"`
			Status   struct{ Phase string }           `json:"status"`
		} `json:"items"`
	}
	if json.Unmarshal(raw, &list) != nil {
		return []PodStatus{}
	}
	items := make([]PodStatus, 0, len(list.Items))
	for _, pod := range list.Items {
		items = append(items, PodStatus{Name: pod.Metadata.Name, Namespace: pod.Metadata.Namespace, Status: pod.Status.Phase})
	}
	return items
}

func (s *Service) Logs(namespace, pod string, lines int) (LogResponse, error) {
	if !safeName(namespace) || !safeName(pod) {
		return LogResponse{}, errors.New("valid namespace and pod are required")
	}
	if lines < 1 || lines > 500 {
		lines = 100
	}
	raw, err := s.kubectl("logs", "-n", namespace, pod, "--tail", strconv.Itoa(lines))
	if err != nil {
		return LogResponse{}, fmt.Errorf("unable to read pod logs: %w", err)
	}
	text := strings.Split(strings.TrimSpace(string(raw)), "\n")
	if len(text) == 1 && text[0] == "" {
		text = []string{}
	}
	return LogResponse{Pod: pod, Namespace: namespace, Lines: text}, nil
}

func (s *Service) StreamLogs(ctx context.Context, namespace, pod string, output io.Writer) error {
	if !safeName(namespace) || !safeName(pod) {
		return errors.New("valid namespace and pod are required")
	}
	cmd := exec.CommandContext(ctx, "kubectl", "logs", "-n", namespace, pod, "--tail", "100", "--follow=true")
	cmd.Stdout = output
	cmd.Stderr = output
	return cmd.Run()
}

func (s *Service) Config() PortalConfig {
	config := PortalConfig{Namespace: s.namespace(), Provider: envOr("CLOUDINFEROPS_PROVIDER", "ollama"), Model: envOr("CLOUDINFEROPS_MODEL", "llama3")}
	raw, err := s.kubectl("get", "configmap", "model-config", "-n", "inference", "-o", "json")
	if err != nil {
		return config
	}
	var item struct {
		Data map[string]string `json:"data"`
	}
	if json.Unmarshal(raw, &item) == nil {
		if item.Data["provider"] != "" {
			config.Provider = item.Data["provider"]
		}
		if item.Data["model"] != "" {
			config.Model = item.Data["model"]
		}
	}
	return config
}

func (s *Service) SaveConfig(config PortalConfig) error {
	if !s.actionsEnabled {
		return errors.New("control-plane actions are disabled")
	}
	if !safeName(config.Namespace) || !safeName(config.Provider) || strings.TrimSpace(config.Model) == "" || len(config.Model) > 128 {
		return errors.New("invalid configuration")
	}
	patch, _ := json.Marshal(map[string]any{"data": map[string]string{"provider": config.Provider, "model": config.Model, "ollama_host": "http://ollama-service:11434"}})
	if out, err := s.kubectl("patch", "configmap", "model-config", "-n", "inference", "--type", "merge", "-p", string(patch)); err != nil {
		return fmt.Errorf("config update failed: %s", strings.TrimSpace(string(out)))
	}
	return nil
}

func (s *Service) Restart(namespace, deployment string) error {
	if !s.actionsEnabled {
		return errors.New("control-plane actions are disabled")
	}
	if !safeName(namespace) || !safeName(deployment) {
		return errors.New("invalid deployment")
	}
	if namespace != "inference" && namespace != s.namespace() {
		return errors.New("namespace is not managed by CloudInferOps")
	}
	_, err := s.kubectl("rollout", "restart", "deployment/"+deployment, "-n", namespace)
	return err
}

func (s *Service) Deploy(target string) error {
	if !s.actionsEnabled {
		return errors.New("control-plane actions are disabled")
	}
	if target != "inference" && target != "observability" && target != "ui" {
		return errors.New("unsupported deployment target")
	}
	// The UI performs a safe reconciliation through the installed CLI. No shell is involved.
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()
	_, err := s.run(ctx, "cloudinferops", "deploy", target)
	return err
}

func (s *Service) Undeploy(target string) error {
	if !s.actionsEnabled {
		return errors.New("control-plane actions are disabled")
	}
	if target != "inference" && target != "observability" {
		return errors.New("unsupported undeployment target")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()
	switch target {
	case "observability":
		_, err := s.run(ctx, "cloudinferops", "uninstall", "observability", "-f")
		return err
	case "inference":
		_, err := s.run(ctx, "kubectl", "delete", "namespace", "inference", "--ignore-not-found=true")
		return err
	default:
		return errors.New("unsupported undeployment target")
	}
}

func (s *Service) RunBenchmark(request BenchmarkRequest) (BenchmarkResult, error) {
	if !s.actionsEnabled {
		return BenchmarkResult{}, errors.New("control-plane actions are disabled")
	}
	if request.Requests < 1 || request.Requests > 20 {
		request.Requests = 5
	}
	if strings.TrimSpace(request.Model) == "" {
		request.Model = s.Config().Model
	}
	url := strings.TrimRight(envOr("CLOUDINFEROPS_GATEWAY_URL", "http://gateway-service.inference.svc.cluster.local:8000"), "/") + "/v1/chat/completions"
	result := BenchmarkResult{Model: request.Model, Requests: request.Requests}
	var total time.Duration
	for i := 0; i < request.Requests; i++ {
		payload, _ := json.Marshal(map[string]any{"model": request.Model, "messages": []map[string]string{{"role": "user", "content": "Reply with one word: ready"}}})
		start := time.Now()
		resp, err := s.httpClient.Post(url, "application/json", bytes.NewReader(payload))
		total += time.Since(start)
		if err != nil {
			result.Failed++
			continue
		}
		_ = resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			result.Succeeded++
		} else {
			result.Failed++
		}
	}
	result.AverageMS = float64(total.Milliseconds()) / float64(request.Requests)
	result.CompletedAt = time.Now().UTC().Format(time.RFC3339)
	s.benchmarkMu.Lock()
	s.benchmark = &result
	s.benchmarkMu.Unlock()
	return result, nil
}

func (s *Service) Benchmark() *BenchmarkResult {
	s.benchmarkMu.RLock()
	defer s.benchmarkMu.RUnlock()
	if s.benchmark == nil {
		return nil
	}
	copy := *s.benchmark
	return &copy
}
func (s *Service) Health() map[string]any {
	version := s.version
	if version == "" {
		version = "dev"
	}
	return map[string]any{"status": "ok", "version": version}
}
func envOr(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
func safeName(value string) bool {
	if value == "" || len(value) > 253 {
		return false
	}
	for _, r := range value {
		if r != '.' && r != '-' && (r < 'a' || r > 'z') && (r < '0' || r > '9') {
			return false
		}
	}
	return true
}
