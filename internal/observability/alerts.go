package observability

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/shivamshashank/StackPulse/internal/utils"
)

// ProvisionAlertRules applies a standard PrometheusRule custom resource containing the SRE Alert Pack.
func ProvisionAlertRules(ns string, dryRun bool) error {
	manifest := fmt.Sprintf(`apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: stackpulse-sre-alerts
  namespace: %s
  labels:
    release: stackpulse-prometheus
spec:
  groups:
  - name: kubernetes-alerts
    rules:
    - alert: PodCrashLooping
      expr: rate(kube_pod_container_status_restarts_total[5m]) * 60 > 2
      for: 2m
      labels:
        severity: critical
        tier: platform
      annotations:
        summary: "Pod {{ $labels.pod }} is crash looping"
        description: "Pod {{ $labels.pod }} in namespace {{ $labels.namespace }} has restarted more than 2 times in the last 5 minutes."
    - alert: PodOOMKilled
      expr: kube_pod_container_status_terminated_reason{reason="OOMKilled"} == 1
      for: 0m
      labels:
        severity: critical
        tier: platform
      annotations:
        summary: "Pod {{ $labels.pod }} OOM killed"
        description: "Pod {{ $labels.pod }} in namespace {{ $labels.namespace }} was terminated due to Out Of Memory."
    - alert: NodeNotReady
      expr: kube_node_status_condition{condition="Ready", status="true"} == 0
      for: 5m
      labels:
        severity: critical
        tier: platform
      annotations:
        summary: "Node {{ $labels.node }} is not ready"
        description: "Node {{ $labels.node }} has been in NotReady state for more than 5 minutes."
    - alert: HighCPUUsage
      expr: sum(rate(container_cpu_usage_seconds_total{container!=""}[5m])) by (pod, namespace) / sum(kube_pod_container_resource_limits{resource="cpu"}) by (pod, namespace) * 100 > 85
      for: 5m
      labels:
        severity: warning
        tier: platform
      annotations:
        summary: "High CPU usage on Pod {{ $labels.pod }}"
        description: "CPU usage of Pod {{ $labels.pod }} in namespace {{ $labels.namespace }} is over 85%%."
    - alert: HighMemoryUsage
      expr: sum(container_memory_working_set_bytes{container!=""}) by (pod, namespace) / sum(kube_pod_container_resource_limits{resource="memory"}) by (pod, namespace) * 100 > 90
      for: 5m
      labels:
        severity: warning
        tier: platform
      annotations:
        summary: "High Memory usage on Pod {{ $labels.pod }}"
        description: "Memory usage of Pod {{ $labels.pod }} in namespace {{ $labels.namespace }} is over 90%%."
  - name: sre-alerts
    rules:
    - alert: HighLatency
      expr: histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le, service)) > 1.5
      for: 2m
      labels:
        severity: warning
        tier: SRE
      annotations:
        summary: "High 95th percentile latency on service {{ $labels.service }}"
        description: "95%% of requests to {{ $labels.service }} are taking more than 1.5s."
    - alert: ErrorRateSpike
      expr: sum(rate(http_requests_total{status=~"5.."}[5m])) by (service) / sum(rate(http_requests_total[5m])) by (service) * 100 > 5
      for: 2m
      labels:
        severity: critical
        tier: SRE
      annotations:
        summary: "Error rate spike on service {{ $labels.service }}"
        description: "HTTP 5xx error rate for service {{ $labels.service }} is over 5%%."
    - alert: ServiceDown
      expr: up == 0
      for: 1m
      labels:
        severity: critical
        tier: SRE
      annotations:
        summary: "Service {{ $labels.job }} is down"
        description: "Scrape job {{ $labels.job }} is failing to respond."
`, ns)

	fmt.Printf("%sProvisioning SRE Alert Rules Pack (PrometheusRule)...\n", utils.PrefixInfo)

	if dryRun {
		fmt.Printf("%s[DRY-RUN] Would apply SRE PrometheusRule alerts\n", utils.PrefixInfo)
		return nil
	}

	tmpPath := filepath.Join(os.TempDir(), "stackpulse-alerts.yaml")
	if err := os.WriteFile(tmpPath, []byte(manifest), 0600); err != nil {
		return fmt.Errorf("failed to write temporary alerts manifest: %w", err)
	}
	defer func() { _ = os.Remove(tmpPath) }()

	_, stderr, err := utils.ExecCommand("", "kubectl", "apply", "-f", tmpPath)
	if err != nil {
		return fmt.Errorf("failed to apply alert rule packs: %w (stderr: %s)", err, stderr)
	}

	fmt.Printf("%sSRE Alert Rules Pack successfully provisioned.\n", utils.PrefixOK)
	return nil
}

// GetAlertRulesManifest returns the raw PrometheusRule manifest string (used for GitOps generation)
func GetAlertRulesManifest(ns string) string {
	return fmt.Sprintf(`apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: stackpulse-sre-alerts
  namespace: %s
  labels:
    release: stackpulse-prometheus
spec:
  groups:
  - name: kubernetes-alerts
    rules:
    - alert: PodCrashLooping
      expr: rate(kube_pod_container_status_restarts_total[5m]) * 60 > 2
      for: 2m
      labels:
        severity: critical
        tier: platform
      annotations:
        summary: "Pod {{ $labels.pod }} is crash looping"
        description: "Pod {{ $labels.pod }} in namespace {{ $labels.namespace }} has restarted more than 2 times in the last 5 minutes."
    - alert: PodOOMKilled
      expr: kube_pod_container_status_terminated_reason{reason="OOMKilled"} == 1
      for: 0m
      labels:
        severity: critical
        tier: platform
      annotations:
        summary: "Pod {{ $labels.pod }} OOM killed"
        description: "Pod {{ $labels.pod }} in namespace {{ $labels.namespace }} was terminated due to Out Of Memory."
    - alert: NodeNotReady
      expr: kube_node_status_condition{condition="Ready", status="true"} == 0
      for: 5m
      labels:
        severity: critical
        tier: platform
      annotations:
        summary: "Node {{ $labels.node }} is not ready"
        description: "Node {{ $labels.node }} has been in NotReady state for more than 5 minutes."
    - alert: HighCPUUsage
      expr: sum(rate(container_cpu_usage_seconds_total{container!=""}[5m])) by (pod, namespace) / sum(kube_pod_container_resource_limits{resource="cpu"}) by (pod, namespace) * 100 > 85
      for: 5m
      labels:
        severity: warning
        tier: platform
      annotations:
        summary: "High CPU usage on Pod {{ $labels.pod }}"
        description: "CPU usage of Pod {{ $labels.pod }} in namespace {{ $labels.namespace }} is over 85%%."
    - alert: HighMemoryUsage
      expr: sum(container_memory_working_set_bytes{container!=""}) by (pod, namespace) / sum(kube_pod_container_resource_limits{resource="memory"}) by (pod, namespace) * 100 > 90
      for: 5m
      labels:
        severity: warning
        tier: platform
      annotations:
        summary: "High Memory usage on Pod {{ $labels.pod }}"
        description: "Memory usage of Pod {{ $labels.pod }} in namespace {{ $labels.namespace }} is over 90%%."
  - name: sre-alerts
    rules:
    - alert: HighLatency
      expr: histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le, service)) > 1.5
      for: 2m
      labels:
        severity: warning
        tier: SRE
      annotations:
        summary: "High 95th percentile latency on service {{ $labels.service }}"
        description: "95%% of requests to {{ $labels.service }} are taking more than 1.5s."
    - alert: ErrorRateSpike
      expr: sum(rate(http_requests_total{status=~"5.."}[5m])) by (service) / sum(rate(http_requests_total[5m])) by (service) * 100 > 5
      for: 2m
      labels:
        severity: critical
        tier: SRE
      annotations:
        summary: "Error rate spike on service {{ $labels.service }}"
        description: "HTTP 5xx error rate for service {{ $labels.service }} is over 5%%."
    - alert: ServiceDown
      expr: up == 0
      for: 1m
      labels:
        severity: critical
        tier: SRE
      annotations:
        summary: "Service {{ $labels.job }} is down"
        description: "Scrape job {{ $labels.job }} is failing to respond."
`, ns)
}
