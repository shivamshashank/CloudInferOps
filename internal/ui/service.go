package ui

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Service struct {
	version string
}

func NewService() *Service {
	return &Service{version: "dev"}
}

func (s *Service) Overview() Overview {
	return Overview{
		Cluster:       s.clusterState(),
		Gateway:       s.gatewayState(),
		Observability: s.observabilityState(),
		Models:        len(s.Models()),
		Alerts:        len(s.Alerts()),
		Pods:          s.podCount(),
		LastUpdated:   time.Now().UTC().Format(time.RFC3339),
		Version:       s.version,
	}
}

func (s *Service) Deployments() []DeploymentStatus {
	return []DeploymentStatus{
		{Name: "cloudinferops-gateway", Namespace: "inference", Status: s.resourceState("gateway"), Replicas: "1/1"},
		{Name: "grafana", Namespace: "observability", Status: s.resourceState("grafana"), Replicas: "1/1"},
		{Name: "prometheus", Namespace: "observability", Status: s.resourceState("prometheus"), Replicas: "1/1"},
	}
}

func (s *Service) Models() []ModelStatus {
	if os.Getenv("CLOUDINFEROPS_UI_USE_MOCK") == "0" {
		return []ModelStatus{
			{Name: "llama3", Provider: "ollama", Status: "Active", Location: "Local"},
			{Name: "mistral", Provider: "ollama", Status: "Active", Location: "Local"},
		}
	}
	return []ModelStatus{
		{Name: "llama3", Provider: "ollama", Status: "Active", Location: "Local"},
		{Name: "mistral", Provider: "ollama", Status: "Active", Location: "Local"},
		{Name: "phi3", Provider: "vllm", Status: "Available", Location: "Remote"},
	}
}

func (s *Service) Alerts() []AlertItem {
	return []AlertItem{
		{Title: "Gateway latency spike", Severity: "warning", Status: "active", Timestamp: time.Now().Add(-10 * time.Minute).UTC().Format(time.RFC3339)},
		{Title: "Model warmup completed", Severity: "info", Status: "resolved", Timestamp: time.Now().Add(-30 * time.Minute).UTC().Format(time.RFC3339)},
	}
}

func (s *Service) Health() map[string]string {
	return map[string]string{"status": "ok", "service": "cloudinferops-ui"}
}

func (s *Service) LogSummary(service string) []string {
	if strings.TrimSpace(service) == "" {
		service = "cloudinferops-ui"
	}
	return []string{
		fmt.Sprintf("[%s] %s ready", time.Now().Add(-2*time.Minute).UTC().Format(time.RFC3339), service),
		fmt.Sprintf("[%s] %s health check passed", time.Now().Add(-5*time.Minute).UTC().Format(time.RFC3339), service),
	}
}

func (s *Service) DeployAction(name string) map[string]string {
	if strings.TrimSpace(name) == "" {
		return map[string]string{"status": "error", "message": "deployment target is required"}
	}
	return map[string]string{"status": "ok", "message": fmt.Sprintf("deployment requested for %s", name)}
}

func (s *Service) clusterState() string {
	if os.Getenv("KUBECONFIG") != "" || os.Getenv("CLOUDINFEROPS_UI_FORCE_LIVE") == "1" {
		return "Connected"
	}
	return "Demo"
}

func (s *Service) gatewayState() string {
	if os.Getenv("CLOUDINFEROPS_UI_FORCE_LIVE") == "1" {
		return "Healthy"
	}
	return "Healthy"
}

func (s *Service) observabilityState() string {
	if os.Getenv("CLOUDINFEROPS_UI_FORCE_LIVE") == "1" {
		return "Healthy"
	}
	return "Healthy"
}

func (s *Service) podCount() int {
	return 8
}

func (s *Service) resourceState(resource string) string {
	if resource == "gateway" && os.Getenv("CLOUDINFEROPS_UI_FORCE_LIVE") == "1" {
		return "Running"
	}
	if resource == "grafana" && os.Getenv("CLOUDINFEROPS_UI_FORCE_LIVE") == "1" {
		return "Running"
	}
	if resource == "prometheus" && os.Getenv("CLOUDINFEROPS_UI_FORCE_LIVE") == "1" {
		return "Running"
	}
	return "Pending"
}
