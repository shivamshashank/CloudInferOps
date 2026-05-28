# StackPulse MVP Roadmap

## Project Goal

StackPulse is a Go-based DevOps/SRE CLI that allows a user to download a binary and deploy a complete observability stack on their own setup.

The user should be able to run StackPulse on:

- Local Linux machine
- AWS EC2 instance
- GCP VM
- Azure VM
- Existing Kubernetes cluster
- Local Kubernetes environments such as k3s, kind, minikube, or Docker Desktop Kubernetes

The MVP should focus on this flow:

```bash
curl -sSL https://raw.githubusercontent.com/shivamshashank/stackpulse/main/install.sh | bash
stackpulse doctor
stackpulse setup k8s --type k3s
stackpulse deploy observability
stackpulse status
```

---

## MVP Scope

### MVP Must Have

- Go CLI binary
- One-command install script
- Kubernetes detection
- Helm detection and installation guidance
- k3s installation option
- Observability namespace creation
- Helm-based deployment
- Prometheus
- Grafana
- Loki
- Tempo
- Alertmanager
- OpenTelemetry Collector
- kube-state-metrics
- Node Exporter
- Grafana dashboards
- Slack alert integration
- Basic PagerDuty integration
- Go-based alert webhook handler
- Health/status checks
- Uninstall command
- GitHub Actions CI/CD
- Automated GitHub Releases
- Tests for CLI commands and internal packages
- README with screenshots and demo commands

### MVP Should Not Have Initially

- Full EKS/GKE/AKS provisioning
- Complex Terraform cloud provisioning
- Multi-node Kubernetes cluster automation
- Production-grade Mimir HA setup
- Paid domain dependency

These can be added after MVP.

---

## Recommended Repository Structure

```text
stackpulse/
├── cmd/
│   └── stackpulse/
│       └── main.go
├── internal/
│   ├── cli/
│   ├── config/
│   ├── doctor/
│   ├── installer/
│   ├── kubernetes/
│   ├── helm/
│   ├── observability/
│   ├── alerts/
│   ├── webhook/
│   ├── gitops/
│   └── utils/
├── charts/
│   └── webhook-handler/
├── configs/
│   ├── prometheus/
│   ├── grafana/
│   ├── loki/
│   ├── tempo/
│   ├── alertmanager/
│   └── otel/
├── dashboards/
├── scripts/
│   └── install.sh
├── docs/
│   ├── architecture.md
│   ├── commands.md
│   ├── troubleshooting.md
│   └── demo.md
├── .github/
│   └── workflows/
│       ├── ci.yml
│       ├── release.yml
│       └── docker.yml
├── Dockerfile.webhook
├── go.mod
├── go.sum
├── README.md
├── ROADMAP.md
└── LICENSE
```

---

## CLI Command Design

### Core Commands

```bash
stackpulse version
stackpulse init
stackpulse doctor
stackpulse setup k8s --type k3s
stackpulse deploy observability
stackpulse deploy webhook-handler
stackpulse status
stackpulse alerts configure --slack
stackpulse alerts configure --pagerduty
stackpulse alerts test
stackpulse dashboards import
stackpulse logs
stackpulse uninstall observability
stackpulse uninstall all
```

---

# Phase 1: Go CLI Foundation

## Goal

Create the base Go CLI with clean command structure.

## Recommended Library

Use Cobra for CLI commands.

```bash
go get github.com/spf13/cobra
go get github.com/spf13/viper
```

## Commands to Implement

```bash
stackpulse version
stackpulse init
stackpulse doctor
```

## Expected Behaviour

### `stackpulse version`

Prints:

```text
StackPulse version: v0.1.0
Commit: abc123
Build date: 2026-05-28
```

### `stackpulse init`

Creates local config:

```text
~/.stackpulse/config.yaml
```

Example config:

```yaml
namespace: observability
kubernetes:
  type: auto
  kubeconfig: ~/.kube/config

observability:
  prometheus: true
  grafana: true
  loki: true
  tempo: true
  alertmanager: true
  opentelemetry: true
  nodeExporter: true
  kubeStateMetrics: true

alerts:
  slack:
    enabled: false
    webhookUrl: ""
  pagerduty:
    enabled: false
    integrationKey: ""
```

---

# Phase 2: Installer Script and Binary Distribution

## Goal

User can install StackPulse with curl.

## Install Command

```bash
curl -sSL https://raw.githubusercontent.com/shivamshashank/stackpulse/main/scripts/install.sh | bash
```

## Installer Responsibilities

The installer should:

1. Detect OS
2. Detect CPU architecture
3. Download latest GitHub release binary
4. Move binary to `/usr/local/bin/stackpulse`
5. Make it executable
6. Verify installation

## Supported OS for MVP

- Linux amd64
- Linux arm64
- macOS amd64
- macOS arm64

## Example Install Script Behaviour

```text
Detecting OS...
Detected linux/amd64
Downloading StackPulse v0.1.0...
Installing to /usr/local/bin/stackpulse...
Installation complete.
Run: stackpulse doctor
```

---

# Phase 3: Doctor Checks

## Goal

Before deployment, StackPulse should check the user's setup.

## Command

```bash
stackpulse doctor
```

## Checks

StackPulse should check:

- OS supported
- CPU architecture supported
- Internet connection available
- `kubectl` installed
- Kubernetes cluster reachable
- Current Kubernetes context
- Helm installed
- Docker/containerd available
- Sufficient CPU and memory
- Required ports available
- Namespace permissions
- Storage class available
- Existing observability stack conflict

## Example Output

```text
StackPulse Doctor

[OK] OS: linux/amd64
[OK] Internet connection
[OK] kubectl found
[WARN] Kubernetes cluster not detected
[OK] Helm found
[OK] Docker found
[OK] Minimum memory: 4GB+
[OK] Minimum CPU: 2 cores+
[INFO] Run: stackpulse setup k8s --type k3s
```

## If Kubernetes Exists

```text
[OK] Kubernetes cluster detected
[OK] Current context: kind-dev
[OK] Nodes ready: 1
[OK] Helm found
[OK] StorageClass found
[READY] You can run: stackpulse deploy observability
```

## If Kubernetes Does Not Exist

```text
[WARN] Kubernetes cluster not found
[INFO] You can install k3s using:
stackpulse setup k8s --type k3s
```

---

# Phase 4: Kubernetes Setup

## Goal

If Kubernetes does not exist, user can install k3s.

## Command

```bash
stackpulse setup k8s --type k3s
```

## Behaviour

The command should:

1. Check if Kubernetes already exists
2. If not, install k3s
3. Configure kubeconfig
4. Wait for node readiness
5. Verify cluster access

## Safety

Before installing k3s:

```text
Kubernetes was not detected.
StackPulse can install k3s on this machine.
Continue? [y/N]
```

## Example Output

```text
Installing k3s...
Configuring kubeconfig...
Waiting for node to become ready...
Kubernetes is ready.
Run: stackpulse deploy observability
```

---

# Phase 5: Helm Integration

## Goal

Use Helm to deploy observability components.

## Commands

```bash
stackpulse deploy observability
```

## Helm Repositories

Add required Helm repos:

```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo add grafana https://grafana.github.io/helm-charts
helm repo add open-telemetry https://open-telemetry.github.io/opentelemetry-helm-charts
helm repo update
```

## Namespace

Create namespace:

```bash
kubectl create namespace observability
```

If it already exists, continue safely.

---

# Phase 6: Observability Stack Deployment

## Goal

Deploy a complete MVP observability stack.

## Command

```bash
stackpulse deploy observability
```

## Default Stack

The MVP stack should include:

| Component | Purpose |
|---|---|
| Prometheus | Metrics collection |
| Grafana | Dashboards and visualisation |
| Loki | Log aggregation |
| Tempo | Distributed tracing |
| Alertmanager | Alert routing |
| OpenTelemetry Collector | Telemetry pipeline |
| Node Exporter | Linux node metrics |
| kube-state-metrics | Kubernetes object metrics |
| Grafana Alloy or Promtail | Log collection |

## Deployment Order

1. Create namespace
2. Add Helm repos
3. Install kube-prometheus-stack
4. Install Loki
5. Install Tempo
6. Install OpenTelemetry Collector
7. Install Alloy or Promtail
8. Apply alert rules
9. Import dashboards
10. Print access instructions

## Example Output

```text
Deploying StackPulse Observability Stack...

[OK] Namespace: observability
[OK] Prometheus installed
[OK] Grafana installed
[OK] Loki installed
[OK] Tempo installed
[OK] Alertmanager installed
[OK] OpenTelemetry Collector installed
[OK] Node Exporter installed
[OK] kube-state-metrics installed
[OK] Dashboards imported

Grafana:
kubectl port-forward svc/stackpulse-grafana 3000:80 -n observability

Username: admin
Password: run `stackpulse status --secrets`
```

---

# Phase 7: Grafana Dashboards

## Goal

Make the project visually strong for LinkedIn and CV.

## Command

```bash
stackpulse dashboards import
```

## Dashboards to Include

- Kubernetes cluster overview
- Node CPU/memory/disk
- Pod health
- Container metrics
- Namespace usage
- Loki logs dashboard
- Tempo traces dashboard
- Alertmanager overview
- Resource usage dashboard

## Dashboard Storage

Store JSON dashboards in:

```text
dashboards/
```

---

# Phase 8: Alerts

## Goal

Add real SRE incident workflow.

## Commands

```bash
stackpulse alerts configure --slack
stackpulse alerts configure --pagerduty
stackpulse alerts test
```

## Alert Rules

Add PrometheusRule resources for:

- Node down
- High CPU usage
- High memory usage
- Disk pressure
- Pod crash loop
- Pod restart spike
- Deployment unavailable
- API server latency
- High error rate
- PersistentVolume almost full

## Slack Configuration

```bash
stackpulse alerts configure --slack
```

Prompt:

```text
Enter Slack webhook URL:
```

Store as Kubernetes secret:

```text
stackpulse-slack-webhook
```

## PagerDuty Configuration

```bash
stackpulse alerts configure --pagerduty
```

Prompt:

```text
Enter PagerDuty integration key:
```

Store as Kubernetes secret:

```text
stackpulse-pagerduty-key
```

## Test Alert

```bash
stackpulse alerts test
```

Expected:

```text
Sending test alert...
[OK] Alert sent to Slack
[OK] Alert sent to PagerDuty
```

---

# Phase 9: Custom Go Webhook Handler

## Goal

Make StackPulse unique, not just a Helm wrapper.

## Component

Build a small Go service:

```text
stackpulse-webhook-handler
```

## Features

- Receives Alertmanager webhook
- Parses alert payload
- Formats incident message
- Sends to Slack
- Sends to PagerDuty
- Stores recent incidents in memory or SQLite
- Exposes health endpoint
- Exposes incident list endpoint

## Endpoints

```text
GET /health
POST /webhook/alertmanager
GET /incidents
```

## Deploy Command

```bash
stackpulse deploy webhook-handler
```

## Docker Image

```text
ghcr.io/shivamshashank/stackpulse-webhook-handler:latest
```

---

# Phase 10: Status Command

## Goal

User can quickly see if everything is working.

## Command

```bash
stackpulse status
```

## Checks

Show status for:

- Kubernetes cluster
- Namespace
- Prometheus
- Grafana
- Loki
- Tempo
- Alertmanager
- OpenTelemetry Collector
- Webhook handler
- Alerts
- Dashboards

## Example Output

```text
StackPulse Status

Cluster: ready
Namespace: observability

Prometheus: running
Grafana: running
Loki: running
Tempo: running
Alertmanager: running
OpenTelemetry Collector: running
Webhook Handler: running

Access Grafana:
kubectl port-forward svc/stackpulse-grafana 3000:80 -n observability
```

---

# Phase 11: Logs Command

## Goal

Give basic CLI debugging capability.

## Commands

```bash
stackpulse logs
stackpulse logs --component grafana
stackpulse logs --component prometheus
stackpulse logs --component loki
```

## Behaviour

Fetch logs from observability namespace pods.

---

# Phase 12: Uninstall

## Goal

Allow clean removal.

## Commands

```bash
stackpulse uninstall observability
stackpulse uninstall all
```

## Behaviour

Remove:

- Helm releases
- Observability namespace
- Alert secrets
- Dashboards configmaps
- Webhook handler

## Safety Prompt

```text
This will remove the StackPulse observability stack.
Continue? [y/N]
```

---

# Phase 13: GitOps and CI/CD

## Goal

Every commit should be tested. Every tagged release should publish binaries.

## GitHub Actions Workflows

### 1. CI Workflow

File:

```text
.github/workflows/ci.yml
```

Runs on:

- Pull request
- Push to main

Checks:

```bash
go test ./...
go vet ./...
go fmt ./...
go mod tidy
```

Optional:

```bash
golangci-lint run
```

### 2. Release Workflow

File:

```text
.github/workflows/release.yml
```

Runs when a tag is pushed:

```bash
git tag v0.1.0
git push origin v0.1.0
```

Builds binaries for:

- linux/amd64
- linux/arm64
- darwin/amd64
- darwin/arm64

Uploads binaries to GitHub Releases.

### 3. Docker Workflow

File:

```text
.github/workflows/docker.yml
```

Builds and pushes webhook handler image to GitHub Container Registry:

```text
ghcr.io/shivamshashank/stackpulse-webhook-handler
```

---

# Phase 14: Test Cases

## Goal

Show engineering quality.

## Unit Tests

Test:

- Config loading
- OS detection
- Kubernetes detection
- Helm detection
- Command validation
- Alert payload parsing
- Slack message formatting
- PagerDuty payload formatting
- Status output formatting

## Integration Tests

Use kind in GitHub Actions.

Test:

1. Create kind cluster
2. Run `stackpulse doctor`
3. Run `stackpulse deploy observability --dry-run`
4. Validate manifests
5. Run `stackpulse uninstall observability --dry-run`

## CLI Tests

Test commands:

```bash
stackpulse version
stackpulse init
stackpulse doctor
stackpulse status
```

---

# Phase 15: Dry Run Mode

## Goal

Useful for safety and testing.

## Commands

```bash
stackpulse deploy observability --dry-run
stackpulse uninstall observability --dry-run
```

## Behaviour

Print what would happen without changing the system.

---

# Phase 16: Documentation

## README Must Include

- Project name and badges
- What StackPulse does
- Why it exists
- Architecture diagram
- Quick start
- Installation
- Commands
- Observability stack components
- Screenshots
- Demo GIF
- Slack/PagerDuty alert screenshots
- Roadmap
- Contributing
- License

## Badges

Add badges for:

- CI
- Release
- Go version
- License
- Docker image
- GitHub stars

## Demo Section

```bash
curl -sSL https://raw.githubusercontent.com/shivamshashank/stackpulse/main/scripts/install.sh | bash
stackpulse doctor
stackpulse setup k8s --type k3s
stackpulse deploy observability
stackpulse status
```

---

# Phase 17: MVP Release Checklist

## Before v0.1.0

- [ ] CLI command structure complete
- [ ] `version` command works
- [ ] `init` command creates config
- [ ] `doctor` checks system
- [ ] k3s setup works on Linux VM
- [ ] Helm repo setup works
- [ ] Observability namespace created
- [ ] Prometheus installed
- [ ] Grafana installed
- [ ] Loki installed
- [ ] Tempo installed
- [ ] Alertmanager installed
- [ ] OpenTelemetry Collector installed
- [ ] Node Exporter installed
- [ ] kube-state-metrics installed
- [ ] Dashboards imported
- [ ] Slack alert test works
- [ ] PagerDuty alert test works
- [ ] Webhook handler deployed
- [ ] `status` command works
- [ ] `logs` command works
- [ ] `uninstall` command works
- [ ] Unit tests added
- [ ] GitHub Actions CI added
- [ ] GitHub Release workflow added
- [ ] Install script added
- [ ] README completed
- [ ] Screenshots added
- [ ] Demo video/GIF added

---

# Suggested MVP Timeline

## Week 1

- Go CLI setup
- Commands: version, init, doctor
- Install script
- GitHub Actions CI

## Week 2

- Kubernetes detection
- k3s setup
- Helm integration
- Namespace management

## Week 3

- Deploy Prometheus, Grafana, Loki, Tempo
- Deploy Alertmanager
- Deploy OpenTelemetry Collector
- Status command

## Week 4

- Slack/PagerDuty integration
- Webhook handler
- Dashboards
- Logs command
- Uninstall command

## Week 5

- Tests
- Release workflow
- Docker workflow
- README
- Screenshots
- LinkedIn demo post

---

# Final MVP Positioning

Use this wording in README:

> StackPulse is a Go-based DevOps/SRE CLI that turns any Kubernetes-compatible environment into a full observability platform. It can detect an existing Kubernetes cluster, install k3s when needed, and deploy Prometheus, Grafana, Loki, Tempo, Alertmanager, OpenTelemetry Collector, dashboards, and alert integrations with a simple command-line workflow.

Use this wording in CV:

> Built StackPulse, a Go-based DevOps/SRE CLI that detects or installs Kubernetes with k3s and deploys a production-style observability stack including Prometheus, Grafana, Loki, Tempo, OpenTelemetry Collector, Alertmanager, Slack/PagerDuty alerts, dashboards, and a custom Go webhook handler.

Use this wording for LinkedIn:

> I built StackPulse — a Go CLI that turns a Linux VM or Kubernetes cluster into a full observability platform with Prometheus, Grafana, Loki, Tempo, OpenTelemetry, dashboards, and alerting in one command.
