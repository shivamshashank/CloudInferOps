package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config represents the schema of StackPulse configuration.
type Config struct {
	Namespace     string        `mapstructure:"namespace"`
	Kubernetes    K8sConfig     `mapstructure:"kubernetes"`
	Observability ObsConfig     `mapstructure:"observability"`
	Alerts        AlertsConfig  `mapstructure:"alerts"`
}

type K8sConfig struct {
	Type       string `mapstructure:"type"`
	Kubeconfig string `mapstructure:"kubeconfig"`
}

type ObsConfig struct {
	Prometheus       bool   `mapstructure:"prometheus"`
	Grafana          bool   `mapstructure:"grafana"`
	Loki             bool   `mapstructure:"loki"`
	Tempo            bool   `mapstructure:"tempo"`
	Alertmanager     bool   `mapstructure:"alertmanager"`
	OpenTelemetry    bool   `mapstructure:"opentelemetry"`
	NodeExporter     bool   `mapstructure:"nodeExporter"`
	KubeStateMetrics bool   `mapstructure:"kubeStateMetrics"`
	LogCollector     string `mapstructure:"logCollector"`
}

type AlertsConfig struct {
	Slack     SlackConfig     `mapstructure:"slack"`
	PagerDuty PagerDutyConfig `mapstructure:"pagerduty"`
}

type SlackConfig struct {
	Enabled          bool   `mapstructure:"enabled"`
	WebhookUrlSecret string `mapstructure:"webhookUrlSecret"`
}

type PagerDutyConfig struct {
	Enabled              bool   `mapstructure:"enabled"`
	IntegrationKeySecret string `mapstructure:"integrationKeySecret"`
}

// GlobalConfig holds the loaded configuration instance
var GlobalConfig Config

// DefaultConfig returns a pre-populated default Config struct
func DefaultConfig() Config {
	return Config{
		Namespace: "observability",
		Kubernetes: K8sConfig{
			Type:       "auto",
			Kubeconfig: "~/.kube/config",
		},
		Observability: ObsConfig{
			Prometheus:       true,
			Grafana:          true,
			Loki:             true,
			Tempo:            true,
			Alertmanager:     true,
			OpenTelemetry:    true,
			NodeExporter:     true,
			KubeStateMetrics: true,
			LogCollector:     "alloy",
		},
		Alerts: AlertsConfig{
			Slack: SlackConfig{
				Enabled:          false,
				WebhookUrlSecret: "stackpulse-slack-webhook",
			},
			PagerDuty: PagerDutyConfig{
				Enabled:              false,
				IntegrationKeySecret: "stackpulse-pagerduty-key",
			},
		},
	}
}

// GetConfigDir returns the absolute path to StackPulse configuration directory (~/.stackpulse)
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("unable to determine home directory: %w", err)
	}
	return filepath.Join(home, ".stackpulse"), nil
}

// GetConfigPath returns the absolute path to the config file (~/.stackpulse/config.yaml)
func GetConfigPath() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

// ExpandPath replaces leading "~" with user's home directory.
func ExpandPath(path string) string {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Clean(filepath.Join(home, path[1:]))
		}
	}
	return filepath.Clean(path)
}

// InitConfig loads the configuration from disk, or initializes default if it does not exist.
func InitConfig(createIfMissing bool) error {
	dir, err := GetConfigDir()
	if err != nil {
		return err
	}
	
	configPath := filepath.Join(dir, "config.yaml")

	// Ensure the config directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if createIfMissing {
			defaults := DefaultConfig()
			viper.Set("namespace", defaults.Namespace)
			viper.Set("kubernetes.type", defaults.Kubernetes.Type)
			viper.Set("kubernetes.kubeconfig", defaults.Kubernetes.Kubeconfig)
			viper.Set("observability.prometheus", defaults.Observability.Prometheus)
			viper.Set("observability.grafana", defaults.Observability.Grafana)
			viper.Set("observability.loki", defaults.Observability.Loki)
			viper.Set("observability.tempo", defaults.Observability.Tempo)
			viper.Set("observability.alertmanager", defaults.Observability.Alertmanager)
			viper.Set("observability.opentelemetry", defaults.Observability.OpenTelemetry)
			viper.Set("observability.nodeExporter", defaults.Observability.NodeExporter)
			viper.Set("observability.kubeStateMetrics", defaults.Observability.KubeStateMetrics)
			viper.Set("observability.logCollector", defaults.Observability.LogCollector)
			viper.Set("alerts.slack.enabled", defaults.Alerts.Slack.Enabled)
			viper.Set("alerts.slack.webhookUrlSecret", defaults.Alerts.Slack.WebhookUrlSecret)
			viper.Set("alerts.pagerduty.enabled", defaults.Alerts.PagerDuty.Enabled)
			viper.Set("alerts.pagerduty.integrationKeySecret", defaults.Alerts.PagerDuty.IntegrationKeySecret)

			if err := viper.WriteConfigAs(configPath); err != nil {
				return fmt.Errorf("failed to write default config: %w", err)
			}
		} else {
			return fmt.Errorf("configuration file does not exist at %s", configPath)
		}
	}

	// Read config
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	// Unmarshal config
	if err := viper.Unmarshal(&GlobalConfig); err != nil {
		return fmt.Errorf("failed to parse config schema: %w", err)
	}

	// Post-processing: expand paths
	GlobalConfig.Kubernetes.Kubeconfig = ExpandPath(GlobalConfig.Kubernetes.Kubeconfig)

	return nil
}

// SaveConfig saves the current GlobalConfig to ~/.stackpulse/config.yaml
func SaveConfig() error {
	dir, err := GetConfigDir()
	if err != nil {
		return err
	}
	configPath := filepath.Join(dir, "config.yaml")

	viper.Set("namespace", GlobalConfig.Namespace)
	viper.Set("kubernetes.type", GlobalConfig.Kubernetes.Type)
	viper.Set("kubernetes.kubeconfig", GlobalConfig.Kubernetes.Kubeconfig)
	viper.Set("observability.prometheus", GlobalConfig.Observability.Prometheus)
	viper.Set("observability.grafana", GlobalConfig.Observability.Grafana)
	viper.Set("observability.loki", GlobalConfig.Observability.Loki)
	viper.Set("observability.tempo", GlobalConfig.Observability.Tempo)
	viper.Set("observability.alertmanager", GlobalConfig.Observability.Alertmanager)
	viper.Set("observability.opentelemetry", GlobalConfig.Observability.OpenTelemetry)
	viper.Set("observability.nodeExporter", GlobalConfig.Observability.NodeExporter)
	viper.Set("observability.kubeStateMetrics", GlobalConfig.Observability.KubeStateMetrics)
	viper.Set("observability.logCollector", GlobalConfig.Observability.LogCollector)
	viper.Set("alerts.slack.enabled", GlobalConfig.Alerts.Slack.Enabled)
	viper.Set("alerts.slack.webhookUrlSecret", GlobalConfig.Alerts.Slack.WebhookUrlSecret)
	viper.Set("alerts.pagerduty.enabled", GlobalConfig.Alerts.PagerDuty.Enabled)
	viper.Set("alerts.pagerduty.integrationKeySecret", GlobalConfig.Alerts.PagerDuty.IntegrationKeySecret)

	if err := viper.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	return nil
}
