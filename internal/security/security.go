package security

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/shivamshashank/StackPulse/internal/utils"
)

const defaultNamespace = "default"

type Runner func(name string, args ...string) (stdout string, stderr string, err error)

type Engine struct {
	Namespace string
	Run       Runner
}

type Report struct {
	Namespace    string       `json:"namespace"`
	Findings     []Finding    `json:"findings"`
	Integrations Integrations `json:"integrations"`
	Summary      Summary      `json:"summary"`
}

type Finding struct {
	Check          string `json:"check"`
	Severity       string `json:"severity"`
	Resource       string `json:"resource"`
	Message        string `json:"message"`
	Recommendation string `json:"recommendation"`
}

type Integrations struct {
	Trivy     IntegrationStatus `json:"trivy"`
	KubeScore IntegrationStatus `json:"kubeScore"`
}

type IntegrationStatus struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type Summary struct {
	PodsChecked        int `json:"podsChecked"`
	DeploymentsChecked int `json:"deploymentsChecked"`
	ServicesChecked    int `json:"servicesChecked"`
	CriticalFindings   int `json:"criticalFindings"`
	HighFindings       int `json:"highFindings"`
	MediumFindings     int `json:"mediumFindings"`
	LowFindings        int `json:"lowFindings"`
}

func NewEngine(namespace string) Engine {
	return Engine{
		Namespace: namespaceOrDefault(namespace),
		Run: func(name string, args ...string) (string, string, error) {
			return utils.ExecCommand("", name, args...)
		},
	}
}

func (e Engine) Scan() (Report, error) {
	namespace := e.namespace()

	var pods podList
	if err := e.getJSON(&pods, "kubectl", "get", "pods", "-n", namespace, "-o", "json"); err != nil {
		return Report{}, err
	}

	var deployments deploymentList
	if err := e.getJSON(&deployments, "kubectl", "get", "deployments", "-n", namespace, "-o", "json"); err != nil {
		return Report{}, err
	}

	var services serviceList
	if err := e.getJSON(&services, "kubectl", "get", "services", "-n", namespace, "-o", "json"); err != nil {
		return Report{}, err
	}

	findings := []Finding{}
	findings = append(findings, privilegedContainerFindings(pods.Items, deployments.Items)...)
	findings = append(findings, missingResourceLimitFindings(deployments.Items)...)
	findings = append(findings, latestImageFindings(pods.Items, deployments.Items)...)
	findings = append(findings, publicExposureFindings(services.Items)...)

	integrations := Integrations{}
	trivyFindings, trivyStatus := e.trivyFindings(namespace)
	findings = append(findings, trivyFindings...)
	integrations.Trivy = trivyStatus

	kubeScoreFindings, kubeScoreStatus := e.kubeScoreFindings(namespace)
	findings = append(findings, kubeScoreFindings...)
	integrations.KubeScore = kubeScoreStatus

	findings = dedupeFindings(findings)
	sortFindings(findings)

	return Report{
		Namespace:    namespace,
		Findings:     findings,
		Integrations: integrations,
		Summary: Summary{
			PodsChecked:        len(pods.Items),
			DeploymentsChecked: len(deployments.Items),
			ServicesChecked:    len(services.Items),
			CriticalFindings:   countSeverity(findings, "critical"),
			HighFindings:       countSeverity(findings, "high"),
			MediumFindings:     countSeverity(findings, "medium"),
			LowFindings:        countSeverity(findings, "low"),
		},
	}, nil
}

func (r Report) Print(w io.Writer) {
	_, _ = fmt.Fprintln(w, "Security Scan Report")
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintf(w, "Namespace: %s\n", r.Namespace)
	_, _ = fmt.Fprintf(w, "Checked: %d pods, %d deployments, %d services\n", r.Summary.PodsChecked, r.Summary.DeploymentsChecked, r.Summary.ServicesChecked)
	_, _ = fmt.Fprintf(w, "Findings: %d critical, %d high, %d medium, %d low\n\n", r.Summary.CriticalFindings, r.Summary.HighFindings, r.Summary.MediumFindings, r.Summary.LowFindings)

	_, _ = fmt.Fprintln(w, "Integrations:")
	_, _ = fmt.Fprintf(w, "- Trivy: %s (%s)\n", r.Integrations.Trivy.Status, r.Integrations.Trivy.Message)
	_, _ = fmt.Fprintf(w, "- kube-score: %s (%s)\n\n", r.Integrations.KubeScore.Status, r.Integrations.KubeScore.Message)

	_, _ = fmt.Fprintln(w, "Findings:")
	if len(r.Findings) == 0 {
		_, _ = fmt.Fprintln(w, "- No security findings detected")
		return
	}

	for _, finding := range r.Findings {
		_, _ = fmt.Fprintf(w, "- [%s] %s: %s\n", finding.Severity, finding.Resource, finding.Message)
		_, _ = fmt.Fprintf(w, "  Recommendation: %s\n", finding.Recommendation)
	}
}

func (r Report) JSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

func (e Engine) getJSON(target any, name string, args ...string) error {
	out, stderr, err := e.runner()(name, args...)
	if err != nil {
		command := strings.TrimSpace(name + " " + strings.Join(args, " "))
		if stderr != "" {
			return fmt.Errorf("%s failed: %s", command, stderr)
		}
		return fmt.Errorf("%s failed: %w", command, err)
	}
	if err := json.Unmarshal([]byte(out), target); err != nil {
		return fmt.Errorf("failed to parse %s output: %w", name, err)
	}
	return nil
}

func (e Engine) trivyFindings(namespace string) ([]Finding, IntegrationStatus) {
	out, stderr, err := e.runner()("trivy", "k8s", "--namespace", namespace, "--severity", "HIGH,CRITICAL", "--format", "json", "all")
	if err != nil {
		return nil, IntegrationStatus{Status: "unavailable", Message: firstNonEmpty(stderr, "trivy not available or scan failed")}
	}

	var trivy trivyReport
	if err := json.Unmarshal([]byte(out), &trivy); err != nil {
		return nil, IntegrationStatus{Status: "error", Message: "failed to parse Trivy JSON output"}
	}

	findings := []Finding{}
	for _, result := range trivy.Results {
		for _, vuln := range result.Vulnerabilities {
			severity := strings.ToLower(vuln.Severity)
			if severity != "critical" && severity != "high" {
				continue
			}
			message := vuln.VulnerabilityID
			if vuln.PkgName != "" {
				message += " in " + vuln.PkgName
			}
			if vuln.Title != "" {
				message += ": " + vuln.Title
			}
			findings = append(findings, Finding{
				Check:          "high_severity_vulnerability",
				Severity:       severity,
				Resource:       emptyAsUnknown(result.Target),
				Message:        message,
				Recommendation: "Upgrade the affected image or package and rerun the vulnerability scan.",
			})
		}
	}
	return findings, IntegrationStatus{Status: "ok", Message: fmt.Sprintf("parsed %d high/critical vulnerabilities", len(findings))}
}

func (e Engine) kubeScoreFindings(namespace string) ([]Finding, IntegrationStatus) {
	out, stderr, err := e.runner()("kube-score", "score", "--namespace", namespace)
	if err != nil {
		return nil, IntegrationStatus{Status: "unavailable", Message: firstNonEmpty(stderr, "kube-score not available or scan failed")}
	}

	findings := []Finding{}
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lower := strings.ToLower(line)
		if strings.Contains(lower, "critical") || strings.Contains(lower, "warning") {
			severity := "medium"
			if strings.Contains(lower, "critical") {
				severity = "high"
			}
			findings = append(findings, Finding{
				Check:          "kube_score",
				Severity:       severity,
				Resource:       "namespace/" + namespace,
				Message:        line,
				Recommendation: "Review the kube-score finding and apply the recommended Kubernetes hardening change.",
			})
		}
	}
	return findings, IntegrationStatus{Status: "ok", Message: fmt.Sprintf("parsed %d kube-score findings", len(findings))}
}

func (e Engine) namespace() string {
	return namespaceOrDefault(e.Namespace)
}

func (e Engine) runner() Runner {
	if e.Run != nil {
		return e.Run
	}
	return NewEngine(e.Namespace).Run
}

func privilegedContainerFindings(pods []pod, deployments []deployment) []Finding {
	findings := []Finding{}
	for _, pod := range pods {
		for _, container := range pod.Spec.Containers {
			if container.SecurityContext.Privileged != nil && *container.SecurityContext.Privileged {
				findings = append(findings, privilegedFinding("pod/"+namespacedName(pod.Metadata), container.Name))
			}
		}
	}
	for _, deployment := range deployments {
		for _, container := range deployment.Spec.Template.Spec.Containers {
			if container.SecurityContext.Privileged != nil && *container.SecurityContext.Privileged {
				findings = append(findings, privilegedFinding("deployment/"+namespacedName(deployment.Metadata), container.Name))
			}
		}
	}
	return findings
}

func missingResourceLimitFindings(deployments []deployment) []Finding {
	findings := []Finding{}
	for _, deployment := range deployments {
		for _, container := range deployment.Spec.Template.Spec.Containers {
			missing := []string{}
			if container.Resources.Requests == nil || container.Resources.Requests["cpu"] == "" || container.Resources.Requests["memory"] == "" {
				missing = append(missing, "requests")
			}
			if container.Resources.Limits == nil || container.Resources.Limits["cpu"] == "" || container.Resources.Limits["memory"] == "" {
				missing = append(missing, "limits")
			}
			if len(missing) == 0 {
				continue
			}
			findings = append(findings, Finding{
				Check:          "missing_resource_limits",
				Severity:       "medium",
				Resource:       "deployment/" + namespacedName(deployment.Metadata),
				Message:        fmt.Sprintf("Container %s is missing resource %s", container.Name, strings.Join(missing, " and ")),
				Recommendation: "Set CPU and memory requests and limits for every workload container.",
			})
		}
	}
	return findings
}

func latestImageFindings(pods []pod, deployments []deployment) []Finding {
	findings := []Finding{}
	for _, pod := range pods {
		for _, container := range pod.Spec.Containers {
			if usesLatestTag(container.Image) {
				findings = append(findings, latestTagFinding("pod/"+namespacedName(pod.Metadata), container.Name, container.Image))
			}
		}
	}
	for _, deployment := range deployments {
		for _, container := range deployment.Spec.Template.Spec.Containers {
			if usesLatestTag(container.Image) {
				findings = append(findings, latestTagFinding("deployment/"+namespacedName(deployment.Metadata), container.Name, container.Image))
			}
		}
	}
	return findings
}

func publicExposureFindings(services []service) []Finding {
	findings := []Finding{}
	for _, service := range services {
		if service.Spec.Type == "LoadBalancer" || service.Spec.Type == "NodePort" {
			findings = append(findings, Finding{
				Check:          "public_exposure",
				Severity:       "medium",
				Resource:       "service/" + namespacedName(service.Metadata),
				Message:        fmt.Sprintf("Service is exposed as %s", service.Spec.Type),
				Recommendation: "Verify the exposure is intentional and protected by ingress rules, network policy, authentication, and TLS.",
			})
		}
	}
	return findings
}

func privilegedFinding(resource, containerName string) Finding {
	return Finding{
		Check:          "privileged_container",
		Severity:       "high",
		Resource:       resource,
		Message:        fmt.Sprintf("Container %s is running privileged", containerName),
		Recommendation: "Disable privileged mode and grant only the specific Linux capabilities the workload needs.",
	}
}

func latestTagFinding(resource, containerName, image string) Finding {
	return Finding{
		Check:          "latest_image_tag",
		Severity:       "medium",
		Resource:       resource,
		Message:        fmt.Sprintf("Container %s uses mutable image tag %q", containerName, image),
		Recommendation: "Pin images to immutable version tags or digests.",
	}
}

func dedupeFindings(findings []Finding) []Finding {
	seen := map[string]bool{}
	out := []Finding{}
	for _, finding := range findings {
		key := finding.Check + "|" + finding.Resource + "|" + finding.Message
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, finding)
	}
	return out
}

func sortFindings(findings []Finding) {
	sort.SliceStable(findings, func(i, j int) bool {
		if findings[i].Severity != findings[j].Severity {
			return severityRank(findings[i].Severity) > severityRank(findings[j].Severity)
		}
		if findings[i].Resource != findings[j].Resource {
			return findings[i].Resource < findings[j].Resource
		}
		return findings[i].Message < findings[j].Message
	})
}

func countSeverity(findings []Finding, severity string) int {
	count := 0
	for _, finding := range findings {
		if finding.Severity == severity {
			count++
		}
	}
	return count
}

func usesLatestTag(image string) bool {
	image = strings.TrimSpace(image)
	if image == "" {
		return false
	}
	lastSlash := strings.LastIndex(image, "/")
	lastColon := strings.LastIndex(image, ":")
	if lastColon <= lastSlash {
		return true
	}
	return image[lastColon+1:] == "latest"
}

func namespacedName(meta metadata) string {
	if meta.Namespace == "" {
		return meta.Name
	}
	return meta.Namespace + "/" + meta.Name
}

func namespaceOrDefault(namespace string) string {
	if strings.TrimSpace(namespace) == "" {
		return defaultNamespace
	}
	return strings.TrimSpace(namespace)
}

func emptyAsUnknown(value string) string {
	if strings.TrimSpace(value) == "" {
		return "unknown"
	}
	return strings.TrimSpace(value)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func severityRank(severity string) int {
	switch severity {
	case "critical":
		return 4
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
	default:
		return 0
	}
}
