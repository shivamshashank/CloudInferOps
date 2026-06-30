package score

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/shivamshashank/StackPulse/internal/utils"
)

const defaultNamespace = "default"

type Runner func(args ...string) (stdout string, stderr string, err error)

type Engine struct {
	Namespace string
	Run       Runner
}

type Report struct {
	Score     int       `json:"score"`
	MaxScore  int       `json:"maxScore"`
	Namespace string    `json:"namespace"`
	Issues    []Finding `json:"issues"`
	Summary   Summary   `json:"summary"`
}

type Finding struct {
	Category  string `json:"category"`
	Severity  string `json:"severity"`
	Message   string `json:"message"`
	Deduction int    `json:"deduction"`
}

type Summary struct {
	PodsChecked        int `json:"podsChecked"`
	DeploymentsChecked int `json:"deploymentsChecked"`
	ServicesChecked    int `json:"servicesChecked"`
	SLOCoverageObjects int `json:"sloCoverageObjects"`
}

func NewEngine(namespace string) Engine {
	return Engine{
		Namespace: namespaceOrDefault(namespace),
		Run: func(args ...string) (string, string, error) {
			return utils.ExecCommand("", "kubectl", args...)
		},
	}
}

func (e Engine) Calculate() (Report, error) {
	namespace := e.namespace()

	var pods podList
	if err := e.getJSON(&pods, "get", "pods", "-n", namespace, "-o", "json"); err != nil {
		return Report{}, err
	}

	var deployments deploymentList
	if err := e.getJSON(&deployments, "get", "deployments", "-n", namespace, "-o", "json"); err != nil {
		return Report{}, err
	}

	services := e.services(namespace)
	sloCount := e.sloCoverageCount(namespace)

	findings := []Finding{}
	findings = append(findings, podHealthFindings(pods.Items)...)
	findings = append(findings, restartFindings(pods.Items)...)
	findings = append(findings, deploymentHealthFindings(deployments.Items)...)
	findings = append(findings, probeFindings(deployments.Items)...)
	findings = append(findings, resourceLimitFindings(deployments.Items)...)
	findings = append(findings, sloFindings(sloCount)...)
	findings = append(findings, securityFindings(pods.Items, deployments.Items, services.Items)...)

	findings = dedupeFindings(findings)
	score := 100
	for _, finding := range findings {
		score -= finding.Deduction
	}
	if score < 0 {
		score = 0
	}

	return Report{
		Score:     score,
		MaxScore:  100,
		Namespace: namespace,
		Issues:    findings,
		Summary: Summary{
			PodsChecked:        len(pods.Items),
			DeploymentsChecked: len(deployments.Items),
			ServicesChecked:    len(services.Items),
			SLOCoverageObjects: sloCount,
		},
	}, nil
}

func (r Report) Print(w io.Writer) {
	_, _ = fmt.Fprintf(w, "Reliability Score: %d/%d\n\n", r.Score, r.MaxScore)
	_, _ = fmt.Fprintf(w, "Namespace: %s\n", r.Namespace)
	_, _ = fmt.Fprintf(w, "Checked: %d pods, %d deployments, %d services\n\n", r.Summary.PodsChecked, r.Summary.DeploymentsChecked, r.Summary.ServicesChecked)

	_, _ = fmt.Fprintln(w, "Issues:")
	if len(r.Issues) == 0 {
		_, _ = fmt.Fprintln(w, "- No reliability issues detected")
		return
	}

	for _, issue := range r.Issues {
		_, _ = fmt.Fprintf(w, "- [%s] %s (-%d)\n", issue.Severity, issue.Message, issue.Deduction)
	}
}

func (r Report) JSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

func (e Engine) getJSON(target any, args ...string) error {
	out, stderr, err := e.runner()(args...)
	if err != nil {
		if stderr != "" {
			return fmt.Errorf("kubectl %s failed: %s", strings.Join(args, " "), stderr)
		}
		return fmt.Errorf("kubectl %s failed: %w", strings.Join(args, " "), err)
	}
	if err := json.Unmarshal([]byte(out), target); err != nil {
		return fmt.Errorf("failed to parse kubectl %s output: %w", strings.Join(args, " "), err)
	}
	return nil
}

func (e Engine) services(namespace string) serviceList {
	var services serviceList
	_ = e.getJSON(&services, "get", "services", "-n", namespace, "-o", "json")
	return services
}

func (e Engine) sloCoverageCount(namespace string) int {
	for _, resource := range []string{"prometheusrules.monitoring.coreos.com", "prometheusrules"} {
		var rules genericList
		if err := e.getJSON(&rules, "get", resource, "-n", namespace, "-o", "json"); err == nil {
			return len(rules.Items)
		}
	}
	return 0
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

func podHealthFindings(pods []pod) []Finding {
	findings := []Finding{}
	for _, pod := range pods {
		if pod.Status.Phase != "Running" && pod.Status.Phase != "Succeeded" {
			findings = append(findings, Finding{
				Category:  "pod_health",
				Severity:  "high",
				Message:   fmt.Sprintf("Pod %s is %s", namespacedName(pod.Metadata), emptyAsUnknown(pod.Status.Phase)),
				Deduction: 12,
			})
		}
		for _, status := range pod.Status.ContainerStatuses {
			if status.State.Waiting != nil && status.State.Waiting.Reason != "" {
				findings = append(findings, Finding{
					Category:  "pod_health",
					Severity:  severityForWaitingReason(status.State.Waiting.Reason),
					Message:   fmt.Sprintf("Container %s in pod %s is waiting: %s", status.Name, namespacedName(pod.Metadata), status.State.Waiting.Reason),
					Deduction: 10,
				})
			}
		}
	}
	return findings
}

func restartFindings(pods []pod) []Finding {
	findings := []Finding{}
	for _, pod := range pods {
		for _, status := range pod.Status.ContainerStatuses {
			if status.RestartCount >= 10 {
				findings = append(findings, Finding{
					Category:  "restart_frequency",
					Severity:  "high",
					Message:   fmt.Sprintf("High restart rate: pod %s container %s restarted %d times", namespacedName(pod.Metadata), status.Name, status.RestartCount),
					Deduction: 15,
				})
			} else if status.RestartCount >= 3 {
				findings = append(findings, Finding{
					Category:  "restart_frequency",
					Severity:  "medium",
					Message:   fmt.Sprintf("Elevated restart rate: pod %s container %s restarted %d times", namespacedName(pod.Metadata), status.Name, status.RestartCount),
					Deduction: 8,
				})
			}
		}
	}
	return findings
}

func deploymentHealthFindings(deployments []deployment) []Finding {
	findings := []Finding{}
	for _, deployment := range deployments {
		if deployment.Status.UnavailableReplicas > 0 || deployment.Status.ReadyReplicas < deployment.Status.Replicas {
			findings = append(findings, Finding{
				Category:  "deployment_health",
				Severity:  "high",
				Message:   fmt.Sprintf("Deployment %s has %d/%d ready replicas", namespacedName(deployment.Metadata), deployment.Status.ReadyReplicas, deployment.Status.Replicas),
				Deduction: 15,
			})
		}
	}
	return findings
}

func probeFindings(deployments []deployment) []Finding {
	findings := []Finding{}
	for _, deployment := range deployments {
		for _, container := range deployment.Spec.Template.Spec.Containers {
			missing := []string{}
			if container.ReadinessProbe == nil {
				missing = append(missing, "readiness")
			}
			if container.LivenessProbe == nil {
				missing = append(missing, "liveness")
			}
			if len(missing) > 0 {
				findings = append(findings, Finding{
					Category:  "probe_coverage",
					Severity:  "medium",
					Message:   fmt.Sprintf("Deployment %s container %s missing %s probe", namespacedName(deployment.Metadata), container.Name, strings.Join(missing, " and ")),
					Deduction: 7,
				})
			}
		}
	}
	return findings
}

func resourceLimitFindings(deployments []deployment) []Finding {
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
			if len(missing) > 0 {
				findings = append(findings, Finding{
					Category:  "resource_limits",
					Severity:  "medium",
					Message:   fmt.Sprintf("Deployment %s container %s missing resource %s", namespacedName(deployment.Metadata), container.Name, strings.Join(missing, " and ")),
					Deduction: 8,
				})
			}
		}
	}
	return findings
}

func sloFindings(sloCount int) []Finding {
	if sloCount > 0 {
		return nil
	}
	return []Finding{{
		Category:  "slo_coverage",
		Severity:  "low",
		Message:   "No SLO coverage detected in PrometheusRule objects",
		Deduction: 5,
	}}
}

func securityFindings(pods []pod, deployments []deployment, services []service) []Finding {
	findings := []Finding{}
	for _, pod := range pods {
		for _, container := range pod.Spec.Containers {
			findings = append(findings, containerSecurityFindings("Pod "+namespacedName(pod.Metadata), container)...)
		}
	}
	for _, deployment := range deployments {
		for _, container := range deployment.Spec.Template.Spec.Containers {
			findings = append(findings, containerSecurityFindings("Deployment "+namespacedName(deployment.Metadata), container)...)
		}
	}
	for _, service := range services {
		if service.Spec.Type == "LoadBalancer" || service.Spec.Type == "NodePort" {
			findings = append(findings, Finding{
				Category:  "security_findings",
				Severity:  "medium",
				Message:   fmt.Sprintf("Service %s is publicly exposed as %s", namespacedName(service.Metadata), service.Spec.Type),
				Deduction: 7,
			})
		}
	}
	return findings
}

func containerSecurityFindings(owner string, container container) []Finding {
	findings := []Finding{}
	if container.SecurityContext.Privileged != nil && *container.SecurityContext.Privileged {
		findings = append(findings, Finding{
			Category:  "security_findings",
			Severity:  "high",
			Message:   fmt.Sprintf("%s container %s is privileged", owner, container.Name),
			Deduction: 12,
		})
	}
	if usesLatestTag(container.Image) {
		findings = append(findings, Finding{
			Category:  "security_findings",
			Severity:  "medium",
			Message:   fmt.Sprintf("%s container %s uses a mutable latest image tag", owner, container.Name),
			Deduction: 6,
		})
	}
	return findings
}

func dedupeFindings(findings []Finding) []Finding {
	seen := map[string]bool{}
	out := []Finding{}
	for _, finding := range findings {
		key := finding.Category + "|" + finding.Message
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, finding)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Severity != out[j].Severity {
			return severityRank(out[i].Severity) > severityRank(out[j].Severity)
		}
		return out[i].Message < out[j].Message
	})
	return out
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

func severityForWaitingReason(reason string) string {
	switch reason {
	case "CrashLoopBackOff", "ImagePullBackOff", "ErrImagePull":
		return "high"
	default:
		return "medium"
	}
}

func severityRank(severity string) int {
	switch severity {
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
