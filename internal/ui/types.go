package ui

type Overview struct {
	Cluster       string `json:"cluster"`
	Gateway       string `json:"gateway"`
	Observability string `json:"observability"`
	Models        int    `json:"models"`
	Alerts        int    `json:"alerts"`
	Pods          int    `json:"pods"`
	LastUpdated   string `json:"last_updated"`
	Version       string `json:"version"`
}

type DeploymentStatus struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Status    string `json:"status"`
	Replicas  string `json:"replicas"`
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
