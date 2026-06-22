package security

type metadata struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Labels    map[string]string `json:"labels"`
}

type resourceList map[string]string

type container struct {
	Name      string `json:"name"`
	Image     string `json:"image"`
	Resources struct {
		Requests resourceList `json:"requests"`
		Limits   resourceList `json:"limits"`
	} `json:"resources"`
	SecurityContext struct {
		Privileged *bool `json:"privileged"`
	} `json:"securityContext"`
}

type pod struct {
	Metadata metadata `json:"metadata"`
	Spec     struct {
		Containers []container `json:"containers"`
	} `json:"spec"`
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

type trivyReport struct {
	Results []struct {
		Target          string `json:"Target"`
		Vulnerabilities []struct {
			VulnerabilityID string `json:"VulnerabilityID"`
			PkgName         string `json:"PkgName"`
			Severity        string `json:"Severity"`
			Title           string `json:"Title"`
		} `json:"Vulnerabilities"`
	} `json:"Results"`
}
