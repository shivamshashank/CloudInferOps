package ui

type Overview struct {
	Cluster       string `json:"cluster"`
	Namespace     string `json:"namespace"`
	Gateway       string `json:"gateway"`
	Observability string `json:"observability"`
	Models        int    `json:"models"`
	Alerts        int    `json:"alerts"`
	Pods          int    `json:"pods"`
	LastUpdated   string `json:"last_updated"`
	Version       string `json:"version"`
	Actions       bool   `json:"actions_enabled"`
}

type DeploymentStatus struct {
	Name        string `json:"name"`
	Namespace   string `json:"namespace"`
	Status      string `json:"status"`
	Replicas    string `json:"replicas"`
	LastUpdated string `json:"last_updated"`
}

type ModelStatus struct {
	Name     string `json:"name"`
	Provider string `json:"provider"`
	Status   string `json:"status"`
	Location string `json:"location"`
}

type AlertItem struct {
	Title     string `json:"title"`
	Severity  string `json:"severity"`
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

type ComponentStatus struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	Namespace string `json:"namespace"`
	URL       string `json:"url,omitempty"`
}

type PodStatus struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Status    string `json:"status"`
}

type LogResponse struct {
	Pod       string   `json:"pod"`
	Namespace string   `json:"namespace"`
	Lines     []string `json:"lines"`
}

type PortalConfig struct {
	Namespace string `json:"namespace"`
	Provider  string `json:"provider"`
	Model     string `json:"model"`
}

type BenchmarkRequest struct {
	Model    string `json:"model"`
	Requests int    `json:"requests"`
}

type BenchmarkResult struct {
	Model       string  `json:"model"`
	Requests    int     `json:"requests"`
	Succeeded   int     `json:"succeeded"`
	Failed      int     `json:"failed"`
	AverageMS   float64 `json:"average_latency_ms"`
	CompletedAt string  `json:"completed_at"`
}
