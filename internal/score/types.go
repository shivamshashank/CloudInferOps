package score

type metadata struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

type resourceList map[string]string

type container struct {
	Name      string `json:"name"`
	Image     string `json:"image"`
	Resources struct {
		Requests resourceList `json:"requests"`
		Limits   resourceList `json:"limits"`
	} `json:"resources"`
	ReadinessProbe  *probe `json:"readinessProbe"`
	LivenessProbe   *probe `json:"livenessProbe"`
	SecurityContext struct {
		Privileged *bool `json:"privileged"`
	} `json:"securityContext"`
}

type probe struct {
	HTTPGet   any `json:"httpGet"`
	TCPSocket any `json:"tcpSocket"`
	Exec      any `json:"exec"`
}

type containerStatus struct {
	Name         string `json:"name"`
	RestartCount int32  `json:"restartCount"`
	State        struct {
		Waiting *struct {
			Reason string `json:"reason"`
		} `json:"waiting"`
	} `json:"state"`
}

type pod struct {
	Metadata metadata `json:"metadata"`
	Spec     struct {
		Containers []container `json:"containers"`
	} `json:"spec"`
	Status struct {
		Phase             string            `json:"phase"`
		ContainerStatuses []containerStatus `json:"containerStatuses"`
	} `json:"status"`
}

type podList struct {
	Items []pod `json:"items"`
}

type deployment struct {
	Metadata metadata `json:"metadata"`
	Spec     struct {
		Template struct {
			Spec struct {
				Containers []container `json:"containers"`
			} `json:"spec"`
		} `json:"template"`
	} `json:"spec"`
	Status struct {
		Replicas            int32 `json:"replicas"`
		ReadyReplicas       int32 `json:"readyReplicas"`
		UnavailableReplicas int32 `json:"unavailableReplicas"`
	} `json:"status"`
}

type deploymentList struct {
	Items []deployment `json:"items"`
}

type service struct {
	Metadata metadata `json:"metadata"`
	Spec     struct {
		Type string `json:"type"`
	} `json:"spec"`
}

type serviceList struct {
	Items []service `json:"items"`
}

type genericList struct {
	Items []struct {
		Metadata metadata `json:"metadata"`
	} `json:"items"`
}
