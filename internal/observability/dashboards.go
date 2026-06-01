package observability

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shivamshashank/StackPulse/internal/utils"
)

// ProvisionDashboards generates and applies all SRE Grafana dashboards as auto-discovered ConfigMaps.
func ProvisionDashboards(ns string, dryRun bool) error {
	dashboards := map[string]string{
		"stackpulse-cluster-overview":   getClusterOverviewJSON(),
		"stackpulse-node-dashboard":     getNodeDashboardJSON(),
		"stackpulse-pod-dashboard":      getPodDashboardJSON(),
		"stackpulse-app-dashboard":      getAppDashboardJSON(),
		"stackpulse-otel-dashboard":     getOtelDashboardJSON(),
		"stackpulse-loki-dashboard":     getLokiDashboardJSON(),
		"stackpulse-blackbox-dashboard": getBlackboxDashboardJSON(),
	}

	fmt.Printf("%sProvisioning auto-discovered SRE Grafana dashboards...\n", utils.PrefixInfo)

	for name, jsonContent := range dashboards {
		manifest := fmt.Sprintf(`apiVersion: v1
kind: ConfigMap
metadata:
  name: %s
  namespace: %s
  labels:
    grafana_dashboard: "1"
data:
  %s.json: |
%s
`, name, ns, name, indentString(jsonContent, 4))

		if dryRun {
			fmt.Printf("%s[DRY-RUN] Would apply Grafana Dashboard ConfigMap '%s'\n", utils.PrefixInfo, name)
			continue
		}

		tmpPath := filepath.Join(os.TempDir(), name+"-dashboard.yaml")
		if err := os.WriteFile(tmpPath, []byte(manifest), 0600); err != nil {
			return fmt.Errorf("failed to write temporary dashboard manifest for %s: %w", name, err)
		}
		defer func(path string) { _ = os.Remove(path) }(tmpPath)

		_, stderr, err := utils.ExecCommand("", "kubectl", "apply", "-f", tmpPath)
		if err != nil {
			return fmt.Errorf("failed to apply Grafana dashboard %s: %w (stderr: %s)", name, err, stderr)
		}
		fmt.Printf("%sDashboard '%s' successfully provisioned.\n", utils.PrefixOK, name)
	}

	return nil
}

func indentString(str string, spaces int) string {
	indent := ""
	for i := 0; i < spaces; i++ {
		indent += " "
	}
	lines := []string{}
	for _, line := range strings.Split(str, "\n") {
		if line == "" {
			lines = append(lines, "")
		} else {
			lines = append(lines, indent+line)
		}
	}
	return strings.Join(lines, "\n")
}

func getClusterOverviewJSON() string {
	return `{
  "annotations": {
    "list": []
  },
  "editable": true,
  "fiscalYearStartMonth": 0,
  "graphTooltip": 0,
  "id": null,
  "links": [],
  "liveNow": false,
  "panels": [
    {
      "collapsed": false,
      "gridPos": {
        "h": 8,
        "w": 12,
        "x": 0,
        "y": 0
      },
      "id": 1,
      "title": "Cluster CPU Usage Percentage",
      "type": "timeseries",
      "targets": [
        {
          "datasource": {
            "type": "prometheus",
            "uid": "prometheus"
          },
          "editorMode": "code",
          "expr": "sum(rate(container_cpu_usage_seconds_total{container!=\"\"}[5m])) / sum(kube_node_status_capacity{resource=\"cpu\"}) * 100",
          "legendFormat": "CPU Usage (%)",
          "range": true
        }
      ]
    },
    {
      "collapsed": false,
      "gridPos": {
        "h": 8,
        "w": 12,
        "x": 12,
        "y": 0
      },
      "id": 2,
      "title": "Cluster Memory Usage Percentage",
      "type": "timeseries",
      "targets": [
        {
          "datasource": {
            "type": "prometheus",
            "uid": "prometheus"
          },
          "editorMode": "code",
          "expr": "sum(container_memory_working_set_bytes{container!=\"\"}) / sum(kube_node_status_capacity{resource=\"memory\"}) * 100",
          "legendFormat": "Memory Usage (%)",
          "range": true
        }
      ]
    },
    {
      "collapsed": false,
      "gridPos": {
        "h": 8,
        "w": 12,
        "x": 0,
        "y": 8
      },
      "id": 3,
      "title": "Cluster Network Transmit Rate",
      "type": "timeseries",
      "targets": [
        {
          "datasource": {
            "type": "prometheus",
            "uid": "prometheus"
          },
          "editorMode": "code",
          "expr": "sum(rate(container_network_transmit_bytes_total[5m]))",
          "legendFormat": "Transmit (bps)",
          "range": true
        }
      ]
    },
    {
      "collapsed": false,
      "gridPos": {
        "h": 8,
        "w": 12,
        "x": 12,
        "y": 8
      },
      "id": 4,
      "title": "Cluster Disk Read/Write Operations",
      "type": "timeseries",
      "targets": [
        {
          "datasource": {
            "type": "prometheus",
            "uid": "prometheus"
          },
          "editorMode": "code",
          "expr": "sum(rate(container_fs_reads_total[5m]))",
          "legendFormat": "Reads",
          "range": true
        },
        {
          "datasource": {
            "type": "prometheus",
            "uid": "prometheus"
          },
          "editorMode": "code",
          "expr": "sum(rate(container_fs_writes_total[5m]))",
          "legendFormat": "Writes",
          "range": true
        }
      ]
    }
  ],
  "refresh": "5s",
  "schemaVersion": 36,
  "style": "dark",
  "tags": ["stackpulse", "kubernetes"],
  "templating": {
    "list": []
  },
  "time": {
    "from": "now-1h",
    "to": "now"
  },
  "timepicker": {},
  "timezone": "browser",
  "title": "StackPulse Cluster Overview",
  "uid": "stackpulse-cluster-overview",
  "version": 1
}`
}

func getNodeDashboardJSON() string {
	return `{
  "editable": true,
  "id": null,
  "panels": [
    {
      "gridPos": {"h": 8, "w": 24, "x": 0, "y": 0},
      "id": 1,
      "title": "CPU Usage per Node",
      "type": "timeseries",
      "targets": [
        {
          "datasource": {"type": "prometheus", "uid": "prometheus"},
          "expr": "sum(rate(node_cpu_seconds_total{mode!=\"idle\"}[5m])) by (instance) / sum(kube_node_status_capacity{resource=\"cpu\"}) by (instance) * 100",
          "legendFormat": "{{instance}}",
          "range": true
        }
      ]
    }
  ],
  "schemaVersion": 36,
  "style": "dark",
  "tags": ["stackpulse", "node"],
  "title": "StackPulse Node Dashboard",
  "uid": "stackpulse-node-dashboard",
  "version": 1
}`
}

func getPodDashboardJSON() string {
	return `{
  "editable": true,
  "id": null,
  "panels": [
    {
      "gridPos": {"h": 8, "w": 12, "x": 0, "y": 0},
      "id": 1,
      "title": "Pod Memory Usage (bytes)",
      "type": "timeseries",
      "targets": [
        {
          "datasource": {"type": "prometheus", "uid": "prometheus"},
          "expr": "sum(container_memory_working_set_bytes{container!=\"\"}) by (pod)",
          "legendFormat": "{{pod}}",
          "range": true
        }
      ]
    },
    {
      "gridPos": {"h": 8, "w": 12, "x": 12, "y": 0},
      "id": 2,
      "title": "Pod Restarts (restarts/5m)",
      "type": "timeseries",
      "targets": [
        {
          "datasource": {"type": "prometheus", "uid": "prometheus"},
          "expr": "sum(rate(kube_pod_container_status_restarts_total[5m])) by (pod)",
          "legendFormat": "{{pod}}",
          "range": true
        }
      ]
    }
  ],
  "schemaVersion": 36,
  "style": "dark",
  "tags": ["stackpulse", "pod"],
  "title": "StackPulse Pod Dashboard",
  "uid": "stackpulse-pod-dashboard",
  "version": 1
}`
}

func getAppDashboardJSON() string {
	return `{
  "editable": true,
  "id": null,
  "panels": [
    {
      "gridPos": {"h": 8, "w": 8, "x": 0, "y": 0},
      "id": 1,
      "title": "Request Rate (RPS)",
      "type": "timeseries",
      "targets": [
        {
          "datasource": {"type": "prometheus", "uid": "prometheus"},
          "expr": "sum(rate(http_requests_total[5m])) by (service)",
          "legendFormat": "{{service}} rps",
          "range": true
        }
      ]
    },
    {
      "gridPos": {"h": 8, "w": 8, "x": 8, "y": 0},
      "id": 2,
      "title": "Error Rate (5xx %)",
      "type": "timeseries",
      "targets": [
        {
          "datasource": {"type": "prometheus", "uid": "prometheus"},
          "expr": "sum(rate(http_requests_total{status=~\"5..\"}[5m])) by (service) / sum(rate(http_requests_total[5m])) by (service) * 100",
          "legendFormat": "{{service}} error %",
          "range": true
        }
      ]
    },
    {
      "gridPos": {"h": 8, "w": 8, "x": 16, "y": 0},
      "id": 3,
      "title": "Latency (95th quantile)",
      "type": "timeseries",
      "targets": [
        {
          "datasource": {"type": "prometheus", "uid": "prometheus"},
          "expr": "histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le, service))",
          "legendFormat": "{{service}} p95",
          "range": true
        }
      ]
    }
  ],
  "schemaVersion": 36,
  "style": "dark",
  "tags": ["stackpulse", "sre"],
  "title": "StackPulse RED Application Dashboard",
  "uid": "stackpulse-app-dashboard",
  "version": 1
}`
}

func getOtelDashboardJSON() string {
	return `{
  "editable": true,
  "id": null,
  "panels": [
    {
      "gridPos": {"h": 8, "w": 24, "x": 0, "y": 0},
      "id": 1,
      "title": "OTel Collector - Spans Received/Dropped",
      "type": "timeseries",
      "targets": [
        {
          "datasource": {"type": "prometheus", "uid": "prometheus"},
          "expr": "sum(rate(otelcol_receiver_accepted_spans[5m])) by (receiver)",
          "legendFormat": "Accepted - {{receiver}}",
          "range": true
        },
        {
          "datasource": {"type": "prometheus", "uid": "prometheus"},
          "expr": "sum(rate(otelcol_receiver_refused_spans[5m])) by (receiver)",
          "legendFormat": "Refused - {{receiver}}",
          "range": true
        }
      ]
    }
  ],
  "schemaVersion": 36,
  "style": "dark",
  "tags": ["stackpulse", "otel"],
  "title": "StackPulse OpenTelemetry Dashboard",
  "uid": "stackpulse-otel-dashboard",
  "version": 1
}`
}

func getLokiDashboardJSON() string {
	return `{
  "editable": true,
  "id": null,
  "panels": [
    {
      "gridPos": {"h": 8, "w": 24, "x": 0, "y": 0},
      "id": 1,
      "title": "Loki Logs Rate (lines/sec)",
      "type": "timeseries",
      "targets": [
        {
          "datasource": {"type": "prometheus", "uid": "prometheus"},
          "expr": "sum(rate(loki_distributor_bytes_received_total[5m])) / 1024",
          "legendFormat": "Bytes Received (KB/s)",
          "range": true
        }
      ]
    }
  ],
  "schemaVersion": 36,
  "style": "dark",
  "tags": ["stackpulse", "loki"],
  "title": "StackPulse Loki Dashboard",
  "uid": "stackpulse-loki-dashboard",
  "version": 1
}`
}

func getBlackboxDashboardJSON() string {
	return `{
  "editable": true,
  "id": null,
  "panels": [
    {
      "gridPos": {"h": 8, "w": 12, "x": 0, "y": 0},
      "id": 1,
      "title": "Target Endpoint Reachability",
      "type": "stat",
      "targets": [
        {
          "datasource": {"type": "prometheus", "uid": "prometheus"},
          "expr": "probe_success",
          "legendFormat": "{{instance}}",
          "range": true
        }
      ]
    },
    {
      "gridPos": {"h": 8, "w": 12, "x": 12, "y": 0},
      "id": 2,
      "title": "HTTP Response Latency (seconds)",
      "type": "timeseries",
      "targets": [
        {
          "datasource": {"type": "prometheus", "uid": "prometheus"},
          "expr": "probe_duration_seconds",
          "legendFormat": "{{instance}}",
          "range": true
        }
      ]
    }
  ],
  "schemaVersion": 36,
  "style": "dark",
  "tags": ["stackpulse", "blackbox"],
  "title": "StackPulse Blackbox Endpoint Dashboard",
  "uid": "stackpulse-blackbox-dashboard",
  "version": 1
}`
}
