package diagnose

type metadata struct {
	Name      string            `json:"name"`
	Labels    map[string]string `json:"labels"`
	Namespace string            `json:"namespace"`
}

type condition struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

type waitingState struct {
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

type terminatedState struct {
	Reason   string `json:"reason"`
	ExitCode int32  `json:"exitCode"`
	Message  string `json:"message"`
}

type containerStatus struct {
	Name         string `json:"name"`
	RestartCount int32  `json:"restartCount"`
	State        struct {
		Waiting *waitingState `json:"waiting"`
	} `json:"state"`
	LastState struct {
		Terminated *terminatedState `json:"terminated"`
	} `json:"lastState"`
}

type pod struct {
	Metadata metadata `json:"metadata"`
	Status   struct {
		Phase             string            `json:"phase"`
		Conditions        []condition       `json:"conditions"`
		ContainerStatuses []containerStatus `json:"containerStatuses"`
	} `json:"status"`
}

type podList struct {
	Items []pod `json:"items"`
}

type deployment struct {
	Status struct {
		Replicas            int32       `json:"replicas"`
		ReadyReplicas       int32       `json:"readyReplicas"`
		UnavailableReplicas int32       `json:"unavailableReplicas"`
		Conditions          []condition `json:"conditions"`
	} `json:"status"`
	Spec struct {
		Selector struct {
			MatchLabels map[string]string `json:"matchLabels"`
		} `json:"selector"`
	} `json:"spec"`
}

type service struct {
	Spec struct {
		Type     string            `json:"type"`
		Selector map[string]string `json:"selector"`
	} `json:"spec"`
	Status struct {
		LoadBalancer struct {
			Ingress []struct {
				IP       string `json:"ip"`
				Hostname string `json:"hostname"`
			} `json:"ingress"`
		} `json:"loadBalancer"`
	} `json:"status"`
}

type endpoints struct {
	Subsets []struct {
		Addresses []struct {
			IP string `json:"ip"`
		} `json:"addresses"`
	} `json:"subsets"`
}

type node struct {
	Metadata metadata `json:"metadata"`
	Status   struct {
		Conditions []condition `json:"conditions"`
	} `json:"status"`
}

type nodeList struct {
	Items []node `json:"items"`
}

type event struct {
	Reason        string `json:"reason"`
	Message       string `json:"message"`
	Type          string `json:"type"`
	LastTimestamp string `json:"lastTimestamp"`
}

type eventList struct {
	Items []event `json:"items"`
}
