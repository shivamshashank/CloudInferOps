# StackPulse V2 Roadmap

## Version 2 Objective

StackPulse V2 should evolve from a one-command Kubernetes observability installer
into a service-aware reliability platform for small engineering teams that do not
have a dedicated SRE team.

The V2 product direction is:

> StackPulse installs production-grade observability, onboards applications,
> creates SLOs, detects reliability risks, and helps teams diagnose incidents
> using Kubernetes events, metrics, logs, traces, deployment history, and alert
> context.

The goal is not to add every observability tool. The goal is to turn StackPulse
into a product that can be used repeatedly by a team, sold as a SaaS/control
plane, and valued by an acquirer.

## Product Positioning

### Current V1 Position

StackPulse V1 is a Go CLI that:

- Detects host and Kubernetes readiness.
- Installs or connects to Kubernetes.
- Deploys Prometheus, Grafana, Loki, Tempo, OpenTelemetry, ArgoCD, alerts, and
  related dashboards.
- Provides status, GitOps bootstrap, alert configuration, webhook handling, and
  uninstall workflows.

This is a strong foundation, but V1 is mostly an installation and bootstrap
tool.

### V2 Position

StackPulse V2 should become:

- A Kubernetes reliability assistant.
- An application onboarding layer for observability.
- An incident diagnosis and timeline product.
- A future SaaS control plane for teams and multiple clusters.

### Target Users

- Startup engineering teams running Kubernetes without dedicated SREs.
- Platform engineers who need fast observability bootstrap across clusters.
- DevOps consultants setting up repeatable production baselines.
- Founders/CTOs who need reliability visibility without buying a heavy
  enterprise platform immediately.

### Core Promise

Within 15 minutes, a user should be able to:

1. Connect a Kubernetes cluster.
2. Deploy the StackPulse observability baseline.
3. Onboard one application.
4. Create basic SLOs and alert rules.
5. See a reliability score.
6. Diagnose a simulated or real incident with evidence.

## Strategic Principles

1. Build workflows, not just integrations.
2. Prefer service-level value over cluster-level dashboards.
3. Make every feature useful from the CLI first.
4. Keep the control plane optional until the local product is excellent.
5. Add AI only where there is reliable evidence and deterministic context.
6. Treat security, privacy, and data boundaries as product features.
7. Avoid becoming a generic Datadog clone.

## V2 Success Metrics

### Product Metrics

- Time from install to first useful dashboard: less than 10 minutes.
- Time from app onboarding to first SLO: less than 5 minutes.
- Time to generate an incident diagnosis: less than 60 seconds.
- New user can understand cluster health from one command.
- At least 3 real-world demo scenarios work end to end.

### Engineering Metrics

- `go test ./...` passes reliably.
- Kind-based integration tests run in CI.
- CLI supports JSON output for automation.
- V2 features are covered by unit tests and at least one integration test each.
- No secrets are printed by default unless explicitly requested.

### Startup Metrics

- Clear demo for YC/interviews.
- Clear paid surface area.
- Clear upgrade path from CLI to hosted control plane.
- At least one workflow that teams would use weekly, not just during setup.

## Roadmap Overview

V2 should be delivered in four phases:

1. V2 Foundation: polish V1 into a dependable product base.
2. Service-Aware Observability: onboard applications and define SLOs.
3. Incident Intelligence: persist incidents and diagnose failures.
4. Control Plane Readiness: prepare for SaaS, teams, and multi-cluster usage.

## Phase 1: V2 Foundation

### Goal

Make the existing CLI reliable, consistent, scriptable, and ready for serious
users.

### Milestone 1.1: Documentation And Command Consistency

Current README references commands that should be verified against the actual
CLI surface.

Tasks:

- Audit every README command against Cobra registration.
- Add missing commands or remove unsupported claims.
- Document supported operating systems clearly.
- Clarify Linux-first install behavior.
- Add a V1 vs V2 capability table.
- Add architecture diagrams for local mode and future cloud mode.

Acceptance criteria:

- Every documented command runs.
- Quick start works on a fresh Linux VM.
- README has no aspirational command examples that are not implemented.

### Milestone 1.2: Machine-Readable CLI Output

Add JSON output for automation and future control-plane integration.

Commands to support:

- `stackpulse doctor --json`
- `stackpulse status --json`
- `stackpulse gitops status --json`
- `stackpulse version --json`

Implementation notes:

- Keep existing human-readable output as default.
- Create typed report structs for status and doctor output.
- Avoid parsing terminal text internally.

Acceptance criteria:

- JSON output is valid and stable.
- Unit tests verify JSON schema fields.
- Exit codes remain meaningful for automation.

### Milestone 1.3: Real Logs Command

Add:

```bash
stackpulse logs
stackpulse logs --component grafana
stackpulse logs --component prometheus
stackpulse logs --component loki
stackpulse logs --component tempo
stackpulse logs --component argocd
stackpulse logs --tail 200
stackpulse logs --follow
```

Implementation notes:

- Map known components to labels/selectors.
- Support namespace override.
- Provide helpful errors when no pods match.

Acceptance criteria:

- Logs work for all core deployed components.
- Tests cover selector construction.
- Command gracefully handles missing deployments.

### Milestone 1.4: Real Alert Test

Upgrade `stackpulse alerts test` from simulated output to real delivery
verification.

Tasks:

- Send a synthetic alert to Alertmanager or the StackPulse webhook handler.
- Verify the webhook receives it.
- Verify Slack/PagerDuty configuration exists when enabled.
- Return a delivery chain report.

Example:

```text
Synthetic alert created
Alertmanager route matched
StackPulse webhook received alert
Slack delivery succeeded
PagerDuty delivery skipped: not configured
```

Acceptance criteria:

- Test fails when Alertmanager is unreachable.
- Test fails when enabled credentials are missing.
- Test does not expose secret values.

### Milestone 1.5: Kind Integration Tests In CI

Add integration tests that deploy the minimum stack into a kind cluster.

Tasks:

- Add CI job for kind.
- Test namespace creation.
- Test Helm repo setup.
- Test dry-run output.
- Test webhook deployment manifest validity.
- Test status command against a controlled fixture or lightweight cluster.

Acceptance criteria:

- CI can distinguish unit tests from integration tests.
- Integration tests are optional locally.
- Main branch blocks on core integration tests.

## Phase 2: Service-Aware Observability

### Goal

Move StackPulse from "cluster observability" to "my app is observable."

### Milestone 2.1: Application Onboarding

Add:

```bash
stackpulse onboard app --namespace payments --service api
stackpulse onboard app --namespace payments --selector app=api
stackpulse onboard app --namespace payments --port 8080
stackpulse onboard app --dry-run
```

Generated resources:

- ServiceMonitor or PodMonitor.
- Grafana service dashboard.
- Default service alert rules.
- Optional OpenTelemetry hints.
- StackPulse app registry entry.

Local config should track onboarded apps:

```yaml
apps:
  - name: api
    namespace: payments
    selector: app=api
    port: 8080
    slo:
      availability: 99.9
      latencyP95Ms: 300
```

Acceptance criteria:

- User can onboard a workload without manually writing YAML.
- Dry run prints generated Kubernetes manifests.
- Onboarded app appears in status output.

### Milestone 2.2: SLO Management

Add:

```bash
stackpulse slo create --service api --availability 99.9 --latency-p95 300ms
stackpulse slo list
stackpulse slo status
stackpulse slo delete --service api
```

Tasks:

- Generate Prometheus recording rules.
- Generate burn-rate alert rules.
- Add dashboard panels for SLOs.
- Track error budget status.

Acceptance criteria:

- SLOs are visible from CLI status.
- SLO rules are generated deterministically.
- Tests cover valid and invalid SLO definitions.

### Milestone 2.3: Service Dashboards

Add generated dashboards per onboarded service.

Panels:

- Request rate.
- Error rate.
- Latency p50/p95/p99.
- Pod restarts.
- CPU and memory.
- Saturation.
- Recent deploy markers if available.

Acceptance criteria:

- Dashboard JSON is valid.
- Datasources are wired correctly.
- Dashboard is labeled for Grafana sidecar discovery.

### Milestone 2.4: OpenTelemetry App Instrumentation Guides

Add:

```bash
stackpulse instrument --language go
stackpulse instrument --language node
stackpulse instrument --language python
stackpulse instrument --language java
```

Output:

- Minimal package install instructions.
- Environment variables.
- Kubernetes deployment patch examples.
- OTLP endpoint for the deployed collector.

Acceptance criteria:

- No unsupported auto-patching by default.
- Generated guidance is language-specific.
- Users can copy the output into their app deployment.

## Phase 3: Incident Intelligence

### Goal

Make StackPulse useful during and after production incidents.

### Milestone 3.1: Persistent Incident Store

The current webhook keeps recent incidents in memory. Replace or extend this
with persistence.

Local mode:

- SQLite under `~/.stackpulse/incidents.db`.

Future cloud mode:

- Postgres in the control plane.

Add:

```bash
stackpulse incidents list
stackpulse incidents show INC-123
stackpulse incidents ack INC-123
stackpulse incidents resolve INC-123
stackpulse incidents export INC-123 --format markdown
```

Incident fields:

- ID.
- Status.
- Severity.
- Service.
- Namespace.
- Alert name.
- First seen.
- Last seen.
- Fingerprint.
- Labels.
- Annotations.
- Related pods.
- Related deployments.
- Timeline events.

Acceptance criteria:

- Incidents survive webhook restart.
- Duplicate alerts are grouped.
- Incident list remains fast with at least 10,000 incidents.

### Milestone 3.2: Diagnosis Bundle Collector

Add:

```bash
stackpulse diagnose pod POD_NAME -n NAMESPACE
stackpulse diagnose service SERVICE_NAME -n NAMESPACE
stackpulse diagnose incident INC-123
stackpulse diagnose cluster
```

Collector should gather:

- Kubernetes events.
- Pod status and restart history.
- Container logs.
- Previous container logs when available.
- Deployment/ReplicaSet state.
- Node conditions.
- Relevant Prometheus metrics.
- ArgoCD sync state.
- Recent alert history.

Privacy controls:

- Redact known secret patterns.
- Do not collect Kubernetes Secrets by default.
- Provide `--include-secrets` only as an explicit dangerous option, if ever.

Acceptance criteria:

- Bundle can be exported as JSON.
- Bundle has deterministic schema.
- Redaction tests cover common secret formats.

### Milestone 3.3: Evidence-Based Incident Explanation

Add:

```bash
stackpulse explain incident INC-123
stackpulse explain pod POD_NAME -n NAMESPACE
```

Output format:

- Summary.
- Likely root cause.
- Evidence.
- Commands to verify.
- Suggested remediation.
- Confidence level.

Example:

```text
Likely cause: pod is crash looping because the container exits with code 1
after failing to connect to POSTGRES_URL.

Evidence:
- 12 restarts in 10 minutes
- Last state: terminated, exit code 1
- Recent logs contain "connection refused"
- No successful readiness probe in the last 8 minutes

Verify:
kubectl logs deployment/api -n payments --previous
kubectl describe pod api-... -n payments
```

Acceptance criteria:

- Explanation always cites evidence from the diagnosis bundle.
- No unsupported claims are made without evidence.
- AI integration is optional and can be disabled.

### Milestone 3.4: Incident Timeline

Build a timeline for each incident:

- Alert fired.
- Deployment changed.
- Pod restarted.
- Node pressure detected.
- ArgoCD sync occurred.
- Alert acknowledged.
- Alert resolved.

CLI:

```bash
stackpulse incidents timeline INC-123
```

Acceptance criteria:

- Timeline is ordered and exportable.
- Related deployment events are included when available.
- Timeline is useful without a hosted UI.

## Phase 4: Control Plane Readiness

### Goal

Prepare StackPulse for SaaS and team usage without weakening the local CLI.

### Milestone 4.1: StackPulse Agent

Add an optional in-cluster agent.

Responsibilities:

- Report cluster metadata.
- Report component health.
- Send incident metadata.
- Send reliability score inputs.
- Receive safe configuration updates.

Non-goals:

- Do not stream all logs by default.
- Do not exfiltrate secrets.
- Do not require cloud mode for local CLI features.

Commands:

```bash
stackpulse cloud login
stackpulse cloud connect
stackpulse deploy agent
stackpulse agent status
```

Acceptance criteria:

- Agent can run in local development mode.
- Agent identity is unique per cluster.
- Communication is authenticated.
- User can see exactly what data is sent.

### Milestone 4.2: Hosted Dashboard MVP

Dashboard views:

- Clusters.
- Services.
- Incidents.
- Reliability score.
- SLO status.
- Alert integrations.

Team features:

- Login.
- Organization.
- Members.
- API keys.
- Audit log.

Acceptance criteria:

- One user can connect one cluster.
- Dashboard reflects cluster health.
- Incident timeline is visible.
- User can delete cluster connection and data.

### Milestone 4.3: Multi-Cluster Support

CLI:

```bash
stackpulse clusters list
stackpulse clusters switch CLUSTER_ID
stackpulse status --cluster CLUSTER_ID
```

Control plane:

- Register multiple clusters.
- Compare reliability scores.
- Show per-cluster incidents.
- Aggregate service health.

Acceptance criteria:

- Local config supports multiple cluster entries.
- Commands can target a specific cluster.
- Hosted dashboard handles more than one cluster per account.

### Milestone 4.4: Billing And Packaging

Potential pricing:

- Free: local CLI, one cluster, community support.
- Pro: hosted dashboard, incident history, AI diagnosis, SLOs.
- Team: multiple users, audit log, multiple clusters.
- Enterprise: SSO, private deployment, compliance exports.

Acceptance criteria:

- Paid features are clearly separated.
- Open-source CLI remains useful.
- Hosted features create recurring value.

## Security And Compliance Workstream

### Milestone S1: Security Scan Command

Add:

```bash
stackpulse security scan
stackpulse security scan --namespace payments
stackpulse security report --format markdown
```

Integrations to evaluate:

- Trivy.
- kube-bench.
- kube-score.
- Polaris.
- Kyverno.

Acceptance criteria:

- Output is grouped by severity.
- Reports include remediation guidance.
- Scans do not require cloud mode.

### Milestone S2: Policy Baselines

Add:

```bash
stackpulse policies apply baseline
stackpulse policies check
stackpulse policies diff
```

Policy areas:

- Privileged containers.
- Missing resource limits.
- HostPath usage.
- Latest image tags.
- Missing probes.
- Public service exposure.

Acceptance criteria:

- Dry-run mode shows impact before applying.
- Users can opt into enforcement.
- Defaults are advisory, not disruptive.

### Milestone S3: Secret Handling

Tasks:

- Stop printing credentials in normal status output by default.
- Add `--show-secrets` when explicitly needed.
- Redact secrets in diagnosis bundles.
- Add tests for redaction.

Acceptance criteria:

- No command accidentally logs secret values.
- Secret display requires explicit user intent.
- JSON output redacts sensitive fields by default.

## Reliability Score Workstream

### Goal

Give teams one understandable score that summarizes readiness and risk.

Add:

```bash
stackpulse score
stackpulse score --json
stackpulse score --namespace payments
```

Score inputs:

- Cluster health.
- Node readiness.
- Pod restart rate.
- Missing probes.
- Missing resource requests/limits.
- SLO coverage.
- Alert coverage.
- Security findings.
- Observability component health.
- Recent incidents.

Example:

```text
StackPulse Reliability Score: 78/100

Top issues:
- 4 workloads missing CPU limits
- 2 services have no SLO
- 1 deployment is crash looping
- Slack alert integration is configured but not verified
```

Acceptance criteria:

- Score is explainable.
- Each deduction maps to a concrete finding.
- JSON output supports hosted control-plane ingestion.

## Managed Kubernetes Workstream

### Goal

Make StackPulse credible for production clusters on EKS, GKE, and AKS.

Add:

```bash
stackpulse doctor --provider eks
stackpulse doctor --provider gke
stackpulse doctor --provider aks
```

Checks:

- Kubernetes version.
- Node pool health.
- Load balancer provisioning.
- Storage class.
- Ingress controller.
- DNS and TLS readiness.
- Metrics server.
- IAM/workload identity basics.
- Network policy availability.
- Public endpoint exposure.
- Autoscaling readiness.

Acceptance criteria:

- Provider-specific checks are modular.
- Missing cloud CLI tools are reported clearly.
- Checks do not mutate the cluster.

## Developer Experience Workstream

### Tasks

- Add command reference docs generated from Cobra.
- Add examples directory.
- Add demo scripts.
- Add issue templates.
- Add architecture decision records under `docs/adr`.
- Add contributing guide for new components.
- Add Makefile for common tasks.

Suggested commands:

```bash
make test
make lint
make build
make integration
make demo-kind
```

Acceptance criteria:

- New contributor can run tests in less than 10 minutes.
- Demo can be repeated from a clean machine.
- Release process is documented.

## Demo Scenarios For V2

### Demo 1: Startup Observability In 15 Minutes

1. Start a fresh Kubernetes cluster.
2. Install StackPulse.
3. Run doctor.
4. Deploy observability.
5. Onboard a sample app.
6. Create SLO.
7. Open service dashboard.

### Demo 2: Diagnose A CrashLoop

1. Deploy a broken app.
2. Trigger alert.
3. View incident.
4. Run diagnosis.
5. Show evidence and remediation.

### Demo 3: Reliability Score

1. Scan cluster.
2. Show score.
3. Fix missing limits/probes.
4. Rerun score.
5. Show improvement.

### Demo 4: Cloud Control Plane Preview

1. Log in.
2. Deploy agent.
3. Connect cluster.
4. Show hosted dashboard.
5. Show incident timeline.

## What Not To Build In V2

Avoid these until the core workflow is strong:

- A full Datadog replacement.
- A custom metrics database.
- A custom log storage engine.
- A generic chat interface without diagnosis context.
- Dozens of integrations before Slack/PagerDuty are reliable.
- Complex enterprise SSO before single-team SaaS works.
- Automatic code changes to user applications.
- Secret collection by default.

## Suggested Repository Structure Changes

Potential new packages:

```text
internal/apps
internal/slo
internal/incidents
internal/diagnostics
internal/score
internal/security
internal/agent
internal/cloud
```

Potential docs:

```text
docs/v2/
docs/v2/architecture.md
docs/v2/app-onboarding.md
docs/v2/incidents.md
docs/v2/security.md
docs/adr/
examples/
```

## Priority Order

If only three V2 features can be built first, build these:

1. Application onboarding.
2. Incident persistence and diagnosis bundle.
3. Reliability score.

These three features turn StackPulse from an installer into a product.

## V2 Release Definition

StackPulse V2 is ready when:

- A new user can deploy the stack on a clean cluster.
- A real application can be onboarded.
- The app has generated dashboards and SLO alerts.
- A real or simulated incident is persisted.
- `stackpulse diagnose` explains the issue with evidence.
- `stackpulse score` gives a useful reliability score.
- The README accurately reflects implemented features.
- Unit and integration tests pass in CI.
- The demo story is clear enough for investors, users, and potential acquirers.

