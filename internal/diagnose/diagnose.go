package diagnose

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/shivamshashank/StackPulse/internal/utils"
)

const defaultNamespace = "default"

// Runner executes kubectl-style commands.
type Runner func(args ...string) (stdout string, stderr string, err error)

// Engine collects Kubernetes evidence and turns it into an actionable report.
type Engine struct {
	Namespace string
	Run       Runner
}

// Report is the user-facing incident diagnosis.
type Report struct {
	Target         string
	Issue          string
	Evidence       []string
	Recommendation string
}

func NewEngine(namespace string) Engine {
	return Engine{
		Namespace: namespaceOrDefault(namespace),
		Run: func(args ...string) (string, string, error) {
			return utils.ExecCommand("", "kubectl", args...)
		},
	}
}

func (e Engine) DiagnosePod(name string) (Report, error) {
	if strings.TrimSpace(name) == "" {
		return Report{}, fmt.Errorf("pod name cannot be empty")
	}

	var pod pod
	if err := e.getJSON(&pod, "get", "pod", name, "-n", e.namespace(), "-o", "json"); err != nil {
		return Report{}, err
	}

	events := e.eventsFor(name, "Pod")
	previousLogs := e.previousLogs(name)

	report := Report{Target: "Pod " + name}
	report.Evidence = append(report.Evidence, "Phase: "+emptyAsUnknown(pod.Status.Phase))

	for _, condition := range pod.Status.Conditions {
		if condition.Status != "True" {
			report.Evidence = append(report.Evidence, fmt.Sprintf("%s: %s", condition.Type, emptyAsUnknown(condition.Reason)))
		}
	}

	issue := "No obvious issue detected"
	recommendation := "Continue monitoring the pod and inspect application-specific logs if symptoms persist."
	maxRestarts := int32(0)

	for _, status := range pod.Status.ContainerStatuses {
		if status.RestartCount > maxRestarts {
			maxRestarts = status.RestartCount
		}
		report.Evidence = append(report.Evidence, fmt.Sprintf("Container %s restarts: %d", status.Name, status.RestartCount))

		if status.State.Waiting != nil && status.State.Waiting.Reason != "" {
			issue = status.State.Waiting.Reason
			report.Evidence = append(report.Evidence, fmt.Sprintf("Container %s waiting: %s", status.Name, status.State.Waiting.Message))
			recommendation = recommendationFor(issue)
		}

		if status.LastState.Terminated != nil {
			terminated := status.LastState.Terminated
			report.Evidence = append(report.Evidence, fmt.Sprintf("Last exit code: %d", terminated.ExitCode))
			if terminated.Reason != "" {
				report.Evidence = append(report.Evidence, "Last termination reason: "+terminated.Reason)
			}
			if terminated.Reason == "OOMKilled" {
				issue = "OOMKilled"
				recommendation = recommendationFor(issue)
			}
		}
	}

	if issue == "No obvious issue detected" && maxRestarts >= 3 {
		issue = "HighRestartCount"
		recommendation = recommendationFor(issue)
	}

	if issue == "No obvious issue detected" && strings.EqualFold(pod.Status.Phase, "Pending") {
		issue = "Pending"
		recommendation = recommendationFor(issue)
	}

	report.Evidence = append(report.Evidence, summarizeEvents(events)...)
	report.Evidence = append(report.Evidence, e.recentAlertSignals()...)
	if previousLogs != "" {
		report.Evidence = append(report.Evidence, "Previous logs available: yes")
		report.Evidence = append(report.Evidence, firstLogSignal(previousLogs))
	}

	report.Issue = issue
	report.Recommendation = recommendation
	return report, nil
}

func (e Engine) DiagnoseDeployment(name string) (Report, error) {
	if strings.TrimSpace(name) == "" {
		return Report{}, fmt.Errorf("deployment name cannot be empty")
	}

	var deployment deployment
	if err := e.getJSON(&deployment, "get", "deployment", name, "-n", e.namespace(), "-o", "json"); err != nil {
		return Report{}, err
	}

	report := Report{Target: "Deployment " + name}
	report.Evidence = append(report.Evidence,
		fmt.Sprintf("Ready replicas: %d/%d", deployment.Status.ReadyReplicas, deployment.Status.Replicas),
		fmt.Sprintf("Unavailable replicas: %d", deployment.Status.UnavailableReplicas),
	)

	issue := "No obvious issue detected"
	recommendation := "Deployment appears healthy. Continue monitoring rollout and workload metrics."

	for _, condition := range deployment.Status.Conditions {
		report.Evidence = append(report.Evidence, fmt.Sprintf("%s: %s (%s)", condition.Type, condition.Status, emptyAsUnknown(condition.Reason)))
		if condition.Type == "Progressing" && condition.Reason == "ProgressDeadlineExceeded" {
			issue = "ProgressDeadlineExceeded"
			recommendation = recommendationFor(issue)
		}
	}

	if issue == "No obvious issue detected" && deployment.Status.UnavailableReplicas > 0 {
		issue = "DeploymentUnavailable"
		recommendation = recommendationFor(issue)
	}

	selector := labelSelector(deployment.Spec.Selector.MatchLabels)
	if selector != "" {
		var pods podList
		if err := e.getJSON(&pods, "get", "pods", "-n", e.namespace(), "-l", selector, "-o", "json"); err == nil {
			unhealthy := unhealthyPods(pods.Items)
			report.Evidence = append(report.Evidence, fmt.Sprintf("Pods selected: %d", len(pods.Items)))
			if len(unhealthy) > 0 {
				report.Evidence = append(report.Evidence, "Unhealthy pods: "+strings.Join(unhealthy, ", "))
				if issue == "No obvious issue detected" {
					issue = "PodsUnhealthy"
					recommendation = recommendationFor(issue)
				}
			}
		}
	}

	report.Evidence = append(report.Evidence, summarizeEvents(e.eventsFor(name, "Deployment"))...)
	report.Evidence = append(report.Evidence, e.recentAlertSignals()...)
	report.Issue = issue
	report.Recommendation = recommendation
	return report, nil
}

func (e Engine) DiagnoseService(name string) (Report, error) {
	if strings.TrimSpace(name) == "" {
		return Report{}, fmt.Errorf("service name cannot be empty")
	}

	var svc service
	if err := e.getJSON(&svc, "get", "service", name, "-n", e.namespace(), "-o", "json"); err != nil {
		return Report{}, err
	}

	report := Report{Target: "Service " + name}
	report.Evidence = append(report.Evidence, "Type: "+emptyAsUnknown(svc.Spec.Type))

	issue := "No obvious issue detected"
	recommendation := "Service has endpoints. Continue checking client errors, DNS, ingress, and backend logs."

	if len(svc.Spec.Selector) == 0 {
		issue = "ServiceMissingSelector"
		recommendation = recommendationFor(issue)
		report.Evidence = append(report.Evidence, "Selector: none")
	} else {
		report.Evidence = append(report.Evidence, "Selector: "+labelSelector(svc.Spec.Selector))
	}

	var endpoints endpoints
	if err := e.getJSON(&endpoints, "get", "endpoints", name, "-n", e.namespace(), "-o", "json"); err == nil {
		ready := endpointCount(endpoints)
		report.Evidence = append(report.Evidence, fmt.Sprintf("Ready endpoints: %d", ready))
		if ready == 0 && issue == "No obvious issue detected" {
			issue = "ServiceNoEndpoints"
			recommendation = recommendationFor(issue)
		}
	}

	if svc.Spec.Type == "LoadBalancer" && len(svc.Status.LoadBalancer.Ingress) == 0 {
		report.Evidence = append(report.Evidence, "LoadBalancer ingress: pending")
		if issue == "No obvious issue detected" {
			issue = "LoadBalancerPending"
			recommendation = recommendationFor(issue)
		}
	}

	report.Evidence = append(report.Evidence, summarizeEvents(e.eventsFor(name, "Service"))...)
	report.Evidence = append(report.Evidence, e.recentAlertSignals()...)
	report.Issue = issue
	report.Recommendation = recommendation
	return report, nil
}

func (e Engine) DiagnoseCluster() (Report, error) {
	report := Report{Target: "Cluster"}
	issue := "No obvious issue detected"
	recommendation := "Cluster baseline looks healthy. Continue monitoring node pressure, workload restarts, and alert trends."

	var nodes nodeList
	if err := e.getJSON(&nodes, "get", "nodes", "-o", "json"); err != nil {
		return Report{}, err
	}

	readyNodes := 0
	notReadyNodes := []string{}
	for _, node := range nodes.Items {
		if nodeReady(node) {
			readyNodes++
		} else {
			notReadyNodes = append(notReadyNodes, node.Metadata.Name)
		}
		for _, condition := range node.Status.Conditions {
			if condition.Status == "True" && condition.Type != "Ready" {
				report.Evidence = append(report.Evidence, fmt.Sprintf("Node %s condition: %s", node.Metadata.Name, condition.Type))
			}
		}
	}
	report.Evidence = append(report.Evidence, fmt.Sprintf("Ready nodes: %d/%d", readyNodes, len(nodes.Items)))
	if len(notReadyNodes) > 0 {
		issue = "NodeNotReady"
		recommendation = recommendationFor(issue)
		report.Evidence = append(report.Evidence, "Not ready nodes: "+strings.Join(notReadyNodes, ", "))
	}

	var pods podList
	if err := e.getJSON(&pods, "get", "pods", "-A", "-o", "json"); err == nil {
		unhealthy := unhealthyPods(pods.Items)
		report.Evidence = append(report.Evidence, fmt.Sprintf("Pods checked: %d", len(pods.Items)))
		if len(unhealthy) > 0 {
			report.Evidence = append(report.Evidence, "Unhealthy pods: "+strings.Join(limitStrings(unhealthy, 5), ", "))
			if issue == "No obvious issue detected" {
				issue = "PodsUnhealthy"
				recommendation = recommendationFor(issue)
			}
		}
	}

	report.Evidence = append(report.Evidence, summarizeEvents(e.warningEvents())...)
	report.Evidence = append(report.Evidence, e.recentAlertSignals()...)
	report.Issue = issue
	report.Recommendation = recommendation
	return report, nil
}

func (r Report) Print(w io.Writer) {
	_, _ = fmt.Fprintf(w, "Target:\n%s\n\n", r.Target)
	_, _ = fmt.Fprintf(w, "Issue:\n%s\n\n", r.Issue)
	_, _ = fmt.Fprintln(w, "Evidence:")
	if len(r.Evidence) == 0 {
		_, _ = fmt.Fprintln(w, "- No Kubernetes evidence collected")
	} else {
		for _, evidence := range dedupe(r.Evidence) {
			if strings.TrimSpace(evidence) != "" {
				_, _ = fmt.Fprintf(w, "- %s\n", evidence)
			}
		}
	}
	_, _ = fmt.Fprintf(w, "\nRecommendation:\n%s\n", r.Recommendation)
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

func (e Engine) eventsFor(name, kind string) []event {
	var events eventList
	fieldSelector := fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=%s", name, kind)
	if err := e.getJSON(&events, "get", "events", "-n", e.namespace(), "--field-selector", fieldSelector, "-o", "json"); err != nil {
		return nil
	}
	return events.Items
}

func (e Engine) warningEvents() []event {
	var events eventList
	if err := e.getJSON(&events, "get", "events", "-A", "--field-selector", "type=Warning", "-o", "json"); err != nil {
		return nil
	}
	return events.Items
}

func (e Engine) previousLogs(podName string) string {
	out, _, err := e.runner()("logs", podName, "-n", e.namespace(), "--previous", "--tail=50")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

func (e Engine) recentAlertSignals() []string {
	pods := e.alertmanagerPods()
	if len(pods) == 0 {
		return nil
	}

	signals := []string{}
	for _, pod := range limitPods(pods, 2) {
		namespace := pod.Metadata.Namespace
		if namespace == "" {
			namespace = e.namespace()
		}
		out, _, err := e.runner()("logs", pod.Metadata.Name, "-n", namespace, "--tail=30")
		if err != nil {
			continue
		}
		for _, line := range strings.Split(out, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			lower := strings.ToLower(line)
			if strings.Contains(lower, "alert") || strings.Contains(lower, "firing") {
				signals = append(signals, "Recent alert signal: "+line)
				break
			}
		}
	}
	return limitStrings(signals, 3)
}

func (e Engine) alertmanagerPods() []pod {
	for _, selector := range []string{"app.kubernetes.io/name=alertmanager", "app=alertmanager"} {
		var pods podList
		if err := e.getJSON(&pods, "get", "pods", "-A", "-l", selector, "-o", "json"); err == nil && len(pods.Items) > 0 {
			return pods.Items
		}
	}
	return nil
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

func namespaceOrDefault(namespace string) string {
	if strings.TrimSpace(namespace) == "" {
		return defaultNamespace
	}
	return strings.TrimSpace(namespace)
}

func summarizeEvents(events []event) []string {
	if len(events) == 0 {
		return nil
	}
	sort.SliceStable(events, func(i, j int) bool {
		return events[i].LastTimestamp > events[j].LastTimestamp
	})

	evidence := []string{}
	for _, event := range limitEvents(events, 5) {
		reason := emptyAsUnknown(event.Reason)
		message := strings.TrimSpace(event.Message)
		if message == "" {
			evidence = append(evidence, "Event: "+reason)
			continue
		}
		evidence = append(evidence, fmt.Sprintf("Event %s: %s", reason, message))
	}
	return evidence
}

func unhealthyPods(pods []pod) []string {
	unhealthy := []string{}
	for _, pod := range pods {
		if pod.Status.Phase != "Running" && pod.Status.Phase != "Succeeded" {
			unhealthy = append(unhealthy, pod.Metadata.Name+"("+pod.Status.Phase+")")
			continue
		}
		for _, status := range pod.Status.ContainerStatuses {
			if status.RestartCount >= 3 {
				unhealthy = append(unhealthy, fmt.Sprintf("%s(restarts=%d)", pod.Metadata.Name, status.RestartCount))
				break
			}
			if status.State.Waiting != nil && status.State.Waiting.Reason != "" {
				unhealthy = append(unhealthy, pod.Metadata.Name+"("+status.State.Waiting.Reason+")")
				break
			}
		}
	}
	return unhealthy
}

func nodeReady(node node) bool {
	for _, condition := range node.Status.Conditions {
		if condition.Type == "Ready" {
			return condition.Status == "True"
		}
	}
	return false
}

func endpointCount(endpoints endpoints) int {
	count := 0
	for _, subset := range endpoints.Subsets {
		count += len(subset.Addresses)
	}
	return count
}

func labelSelector(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}
	keys := make([]string, 0, len(labels))
	for key := range labels {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+labels[key])
	}
	return strings.Join(parts, ",")
}

func recommendationFor(issue string) string {
	switch issue {
	case "CrashLoopBackOff":
		return "Inspect the previous container logs, recent config changes, probes, and dependencies causing the process to exit."
	case "OOMKilled":
		return "Increase memory limits or reduce application memory usage; validate recent traffic and allocation changes."
	case "ImagePullBackOff", "ErrImagePull":
		return "Verify the image name, tag, registry credentials, and network access from cluster nodes."
	case "Pending":
		return "Check scheduling events, node capacity, taints, tolerations, persistent volume claims, and resource requests."
	case "HighRestartCount":
		return "Review previous logs and readiness/liveness probes; identify whether restarts align with resource pressure or application failures."
	case "ProgressDeadlineExceeded":
		return "Inspect rollout events and selected pods; fix failing pods before retrying or rolling back the deployment."
	case "DeploymentUnavailable":
		return "Check unavailable pods, replica counts, resource pressure, and rollout history for the deployment."
	case "PodsUnhealthy":
		return "Diagnose the listed unhealthy pods and review recent events for scheduling, image, probe, or runtime failures."
	case "ServiceMissingSelector":
		return "Add a selector that matches backend pod labels or intentionally manage EndpointSlices for this service."
	case "ServiceNoEndpoints":
		return "Verify backend pods are running, ready, and match the service selector labels."
	case "LoadBalancerPending":
		return "Check cloud load balancer integration, service annotations, and cluster networking support."
	case "NodeNotReady":
		return "Inspect not-ready node conditions, kubelet status, disk/memory pressure, and recent node events."
	default:
		return "Review the collected evidence and inspect related Kubernetes events, logs, and recent changes."
	}
}

func firstLogSignal(logs string) string {
	lines := strings.Split(strings.TrimSpace(logs), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			return "Previous log signal: " + line
		}
	}
	return "Previous logs were empty"
}

func emptyAsUnknown(value string) string {
	if strings.TrimSpace(value) == "" {
		return "unknown"
	}
	return strings.TrimSpace(value)
}

func dedupe(values []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, value := range values {
		if !seen[value] {
			seen[value] = true
			out = append(out, value)
		}
	}
	return out
}

func limitEvents(events []event, max int) []event {
	if len(events) <= max {
		return events
	}
	return events[:max]
}

func limitStrings(values []string, max int) []string {
	if len(values) <= max {
		return values
	}
	return values[:max]
}

func limitPods(values []pod, max int) []pod {
	if len(values) <= max {
		return values
	}
	return values[:max]
}
