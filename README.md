<div align="center">

# 🚀 StackPulse

### One-command observability platform for Kubernetes, Linux VMs, and cloud instances.

**StackPulse** is a Go-based DevOps/SRE CLI that detects your environment,
validates Kubernetes readiness, installs lightweight Kubernetes when needed, and
deploys a production-style observability stack with metrics, logs, traces,
dashboards, alerts, and incident webhooks.

<br />

[![CI](https://img.shields.io/github/actions/workflow/status/shivamshashank/StackPulse/ci.yml?branch=main&label=CI&logo=githubactions&style=flat-square)](https://github.com/shivamshashank/StackPulse/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/actions/workflow/status/shivamshashank/StackPulse/release.yml?branch=main&label=Release&logo=githubactions&style=flat-square)](https://github.com/shivamshashank/StackPulse/actions/workflows/release.yml)
[![Codecov](https://img.shields.io/codecov/c/github/shivamshashank/StackPulse?logo=codecov&style=flat-square)](https://codecov.io/gh/shivamshashank/StackPulse)
[![Go Report Card](https://goreportcard.com/badge/github.com/shivamshashank/StackPulse)](https://goreportcard.com/report/github.com/shivamshashank/StackPulse)
[![GitHub release](https://img.shields.io/github/v/release/shivamshashank/StackPulse?style=flat-square)](https://github.com/shivamshashank/StackPulse/releases)
[![GitHub stars](https://img.shields.io/github/stars/shivamshashank/StackPulse?style=flat-square)](https://github.com/shivamshashank/StackPulse/stargazers)
[![GitHub forks](https://img.shields.io/github/forks/shivamshashank/StackPulse?style=flat-square)](https://github.com/shivamshashank/StackPulse/network/members)
[![License](https://img.shields.io/github/license/shivamshashank/StackPulse?style=flat-square)](LICENSE)

<br />

![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![Kubernetes](https://img.shields.io/badge/Kubernetes-326CE5?style=for-the-badge&logo=kubernetes&logoColor=white)
![Helm](https://img.shields.io/badge/Helm-0F1689?style=for-the-badge&logo=helm&logoColor=white)
![Prometheus](https://img.shields.io/badge/Prometheus-E6522C?style=for-the-badge&logo=prometheus&logoColor=white)
![Grafana](https://img.shields.io/badge/Grafana-F46800?style=for-the-badge&logo=grafana&logoColor=white)
![Loki](https://img.shields.io/badge/Loki-F46800?style=for-the-badge&logo=grafana&logoColor=white)
![Tempo](https://img.shields.io/badge/Tempo-F46800?style=for-the-badge&logo=grafana&logoColor=white)
![OpenTelemetry](https://img.shields.io/badge/OpenTelemetry-000000?style=for-the-badge&logo=opentelemetry&logoColor=white)
![Alertmanager](https://img.shields.io/badge/Alertmanager-E6522C?style=for-the-badge&logo=prometheus&logoColor=white)
![Slack](https://img.shields.io/badge/Slack-4A154B?style=for-the-badge&logo=slack&logoColor=white)
![PagerDuty](https://img.shields.io/badge/PagerDuty-06AC38?style=for-the-badge&logo=pagerduty&logoColor=white)
![AWS](https://img.shields.io/badge/AWS_EC2-FF9900?style=for-the-badge&logo=amazonaws&logoColor=white)
![GCP](https://img.shields.io/badge/GCP_VM-4285F4?style=for-the-badge&logo=googlecloud&logoColor=white)
![Azure](https://img.shields.io/badge/Azure_VM-0078D4?style=for-the-badge&logo=microsoftazure&logoColor=white)
![GitHub Actions](https://img.shields.io/badge/GitHub_Actions-2088FF?style=for-the-badge&logo=githubactions&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white)

<br />

[Quick Start](#-quick-start) • [Features](#-features) •
[Architecture](#-architecture) • [Commands](#-cli-commands) •
[Testing](#-testing) • [CI/CD](#-cicd--gitops) • [Author](#-author)

</div>

---

## 📌 What is StackPulse?

StackPulse turns any Kubernetes-compatible environment into a complete
observability platform.

It works on:

- 💻 Local Linux machines
- ☁️ AWS EC2 instances
- ☁️ GCP Compute Engine VMs
- ☁️ Azure VMs
- ☸️ Existing Kubernetes clusters
- 🧪 Local clusters such as k3s, kind, minikube, and Docker Desktop Kubernetes

StackPulse follows a simple workflow:

```bash
curl -sSL https://raw.githubusercontent.com/shivamshashank/StackPulse/main/scripts/install.sh | bash
sudo stackpulse doctor
sudo stackpulse deploy observability
sudo stackpulse status
```

---

## ✨ Features

### 🧠 Smart Environment Detection

- Detects OS and CPU architecture
- Detects `kubectl`
- Detects Kubernetes cluster availability
- Detects Helm
- Checks memory, CPU, ports, storage class, and namespace permissions
- Warns about existing observability stack conflicts

### ☸️ Kubernetes First

- Uses existing Kubernetes cluster if available
- Installs k3s when Kubernetes is missing
- Supports local and cloud VM environments
- Works consistently across local Linux, AWS EC2, GCP VM, and Azure VM

### 📊 Full Observability Stack

| Layer              | Component                |
| ------------------ | ------------------------ |
| Metrics            | Prometheus               |
| Dashboards         | Grafana                  |
| Logs               | Loki                     |
| Traces             | Tempo                    |
| Telemetry Pipeline | OpenTelemetry Collector  |
| Alerts             | Alertmanager             |
| Node Metrics       | Node Exporter            |
| Kubernetes Metrics | kube-state-metrics       |
| Log Collection     | Grafana Alloy / Promtail |
| Incident Routing   | Slack, PagerDuty         |
| Custom Handler     | Go webhook handler       |

### 🚨 Incident & Alerting

- Slack alert integration
- PagerDuty alert integration
- Alertmanager webhook support
- Custom Go alert webhook handler
- Test alert command
- Prebuilt SRE alert rules

### 📈 Dashboards Included

- Kubernetes cluster overview
- Node CPU, memory, disk
- Pod health
- Container metrics
- Namespace usage
- Loki logs dashboard
- Tempo traces dashboard
- Alertmanager overview
- Resource usage dashboard

---

## ⚡ Quick Start

### 1. Install StackPulse

```bash
curl -sSL https://raw.githubusercontent.com/shivamshashank/StackPulse/main/scripts/install.sh | bash
```

Verify installation:

```bash
stackpulse version
```

---

### 2. Check Your System

> [!IMPORTANT]
> **Sudo Privileges Required** StackPulse requires administrative/root
> privileges (`sudo`) to manage system observability resources, check system
> metrics, and set up networking or local clusters.

```bash
sudo stackpulse doctor
```

Example output:

```text
StackPulse Doctor

[OK] OS: linux/amd64
[OK] Internet connection
[OK] kubectl found
[WARN] Kubernetes cluster not detected
[OK] Helm found
[OK] Minimum memory: 4GB+
[OK] Minimum CPU: 2 cores+

[INFO] Run: sudo stackpulse deploy observability
```

If Kubernetes already exists:

```text
[OK] Kubernetes cluster detected
[OK] Current context: kind-dev
[OK] Nodes ready: 1
[OK] Helm found
[OK] StorageClass found

[READY] Run: sudo stackpulse deploy observability
```

---

### 3. Deploy Observability Stack & Auto-bootstrap Kubernetes

To deploy the observability stack, simply run:

```bash
sudo stackpulse deploy observability
```

> [!TIP]
> **No Kubernetes? No problem!** If StackPulse does not detect an existing
> Kubernetes cluster, it will automatically ask to install and bootstrap a
> lightweight local Kubernetes cluster (supporting `kind`, `minikube`, or `k3s`)
> on-the-fly, then automatically deploy the observability stack onto it. If you
> already have a cluster running, it will deploy directly onto your active
> context.

StackPulse deploys:

- Prometheus
- Grafana
- Loki
- Tempo
- Alertmanager
- OpenTelemetry Collector
- Node Exporter
- kube-state-metrics
- Grafana Alloy / Promtail
- Dashboards
- Alert rules

---

### 5. Check Status

```bash
sudo stackpulse status
```

Example:

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

Grafana:
kubectl port-forward svc/stackpulse-grafana 3000:80 -n observability
```

---

## 🏗️ Architecture

```text
                           ┌──────────────────────────┐
                           │      StackPulse CLI      │
                           │        Go Binary         │
                           └─────────────┬────────────┘
                                         │
                 ┌───────────────────────┼───────────────────────┐
                 │                       │                       │
          ┌──────▼──────┐        ┌───────▼───────┐        ┌──────▼──────┐
          │   Doctor    │        │ Kubernetes    │        │    Helm     │
          │   Checks    │        │  Detection    │        │ Deployment  │
          └──────┬──────┘        └───────┬───────┘        └──────┬──────┘
                 │                       │                       │
                 │              ┌────────▼────────┐              │
                 │              │ Existing K8s or │              │
                 │              │ k3s Installer   │              │
                 │              └────────┬────────┘              │
                 │                       │                       │
                 └───────────────────────▼───────────────────────┘
                                         │
                              ┌──────────▼──────────┐
                              │   observability ns  │
                              └──────────┬──────────┘
                                         │
        ┌────────────────────────────────┼────────────────────────────────┐
        │                                │                                │
┌───────▼────────┐              ┌────────▼────────┐              ┌────────▼───────┐
│   Prometheus   │              │     Grafana     │              │ Alertmanager   │
│    Metrics     │              │   Dashboards    │              │    Alerts      │
└───────┬────────┘              └────────┬────────┘              └────────┬───────┘
        │                                │                                │
┌───────▼────────┐              ┌────────▼────────┐              ┌────────▼───────┐
│ Node Exporter  │              │      Loki       │              │ Slack/PagerDuty│
│ kube-state     │              │      Logs       │              │ Integrations   │
└────────────────┘              └────────┬────────┘              └────────┬───────┘
                                         │                                │
                              ┌──────────▼──────────┐          ┌──────────▼──────────┐
                              │       Tempo         │          │ Go Webhook Handler  │
                              │       Traces        │          │ Incident Processing │
                              └──────────┬──────────┘          └─────────────────────┘
                                         │
                              ┌──────────▼──────────┐
                              │ OpenTelemetry       │
                              │ Collector           │
                              └─────────────────────┘
```

---

## 🧰 CLI Commands

### General

```bash
stackpulse version
sudo stackpulse init
sudo stackpulse doctor
sudo stackpulse status
```

### Observability

```bash
sudo stackpulse deploy observability
sudo stackpulse deploy observability --dry-run
sudo stackpulse dashboards import
sudo stackpulse logs
sudo stackpulse logs --component grafana
sudo stackpulse logs --component prometheus
sudo stackpulse logs --component loki
```

### Alerts

```bash
sudo stackpulse alerts configure --slack
sudo stackpulse alerts configure --pagerduty
sudo stackpulse alerts test
```

### Webhook Handler

```bash
sudo stackpulse deploy webhook-handler
```

### Cleanup

```bash
sudo stackpulse uninstall observability
sudo stackpulse uninstall all
```

---

## ⚙️ Configuration

StackPulse stores local configuration at:

```text
~/.stackpulse/config.yaml
```

Example:

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
  logCollector: alloy

alerts:
  slack:
    enabled: false
    webhookUrlSecret: stackpulse-slack-webhook
  pagerduty:
    enabled: false
    integrationKeySecret: stackpulse-pagerduty-key
```

---

## 🚨 Alert Rules

StackPulse includes SRE-focused alert rules:

| Alert                      | Description                         |
| -------------------------- | ----------------------------------- |
| NodeDown                   | Kubernetes node is not ready        |
| HighCPUUsage               | Node or pod CPU usage is high       |
| HighMemoryUsage            | Node or pod memory usage is high    |
| DiskPressure               | Node disk pressure detected         |
| PodCrashLooping            | Pod is repeatedly crashing          |
| PodRestartSpike            | Pod restart count increased         |
| DeploymentUnavailable      | Deployment has unavailable replicas |
| HighAPILatency             | API latency is above threshold      |
| HighErrorRate              | Application error rate increased    |
| PersistentVolumeAlmostFull | PVC usage is close to capacity      |

---

## 🔔 Slack & PagerDuty

### Configure Slack

```bash
sudo stackpulse alerts configure --slack
```

### Configure PagerDuty

```bash
sudo stackpulse alerts configure --pagerduty
```

### Send Test Alert

```bash
sudo stackpulse alerts test
```

Expected output:

```text
Sending test alert...
[OK] Alert sent to Slack
[OK] Alert sent to PagerDuty
```

---

## 🧩 Go Webhook Handler

StackPulse includes a custom Go service for incident processing.

### Endpoints

```text
GET  /health
POST /webhook/alertmanager
GET  /incidents
```

### Capabilities

- Receives Alertmanager webhooks
- Parses alert payloads
- Formats incident messages
- Sends notifications to Slack
- Sends incidents to PagerDuty
- Stores recent incidents
- Exposes health and incident APIs

Deploy it with:

```bash
sudo stackpulse deploy webhook-handler
```

---

## 🧪 Testing

### Run All Tests

```bash
go test ./...
```

### Run Tests with Coverage

```bash
go test ./... -cover
```

### Generate Coverage Report

```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Run Vet

```bash
go vet ./...
```

### Run Formatting Check

```bash
gofmt -w .
```

### Run Linter

```bash
golangci-lint run
```

### Integration Test with kind

```bash
kind create cluster --name stackpulse-test
sudo stackpulse doctor
sudo stackpulse deploy observability --dry-run
sudo stackpulse uninstall observability --dry-run
kind delete cluster --name stackpulse-test
```

---

## ✅ Test Coverage

| Area          | Tests                                        |
| ------------- | -------------------------------------------- |
| CLI commands  | `version`, `init`, `doctor`, `status`        |
| Config        | Load, validate, default values               |
| Doctor checks | OS, arch, kubectl, Helm, Kubernetes          |
| Kubernetes    | Cluster detection, namespace creation        |
| Helm          | Repo add, release detection, dry-run         |
| Alerts        | Slack payload, PagerDuty payload, test alert |
| Webhook       | Alertmanager payload parsing                 |
| Status        | Component health formatting                  |
| Uninstall     | Dry-run and confirmation flow                |

---

## 🔁 CI/CD & GitOps

StackPulse uses GitHub Actions for automated testing, builds, Docker images, and
releases.

### CI Workflow

Runs on every push and pull request:

```yaml
go test ./...
go vet ./...
gofmt -w .
golangci-lint run
```

### Release Workflow

Create a new release by pushing a tag:

```bash
git tag v0.1.0
git push origin v0.1.0
```

The release workflow builds binaries for:

- Linux amd64
- Linux arm64
- macOS amd64
- macOS arm64

Artifacts are uploaded to GitHub Releases.

### Docker Workflow

Builds and publishes the webhook handler image:

```text
ghcr.io/shivamshashank/stackpulse-webhook-handler:latest
```

---

## 📦 Installation Options

### Curl Install

```bash
curl -sSL https://raw.githubusercontent.com/shivamshashank/StackPulse/main/scripts/install.sh | bash
```

### Linux Manual Installation (via Curl)

You can download and install the latest compiled binary manually for Linux:

```bash
# 1. Download the latest binary for your architecture (e.g., AMD64 / x86_64)
curl -LO https://github.com/shivamshashank/StackPulse/releases/latest/download/stackpulse-linux-amd64

# 2. Make the binary executable
chmod +x stackpulse-linux-amd64

# 3. Move it to your local system's bin directory to make it globally available
sudo mv stackpulse-linux-amd64 /usr/local/bin/stackpulse

# 4. Verify the installation
stackpulse version
```

_(For ARM64 processors, replace `stackpulse-linux-amd64` with
`stackpulse-linux-arm64`)_

### Go Install

```bash
go install github.com/shivamshashank/StackPulse/cmd/stackpulse@latest
```

### GitHub Releases

Alternatively, you can manually download precompiled binaries for all supported
platforms (Linux & macOS) directly from
[GitHub Releases](https://github.com/shivamshashank/StackPulse/releases).

---

## 👑 Running with Sudo / Root Privileges

To install system prerequisites (such as Kubernetes clusters and core networking
configurations) and bind services locally, you can run StackPulse fully under
elevated privileges (`sudo` mode):

```bash
sudo stackpulse deploy observability
sudo stackpulse status
```

> [!NOTE]
> StackPulse is built with **smart environment-aware root fallback**. When run
> as `sudo`, the CLI automatically detects the original invoking user
> (`$SUDO_USER`) and correctly references their standard home directory paths
> (such as `~/.kube/config` and `~/.stackpulse/config.yaml`), preventing
> configuration directory pollution inside the `/root` path.

---

## 🧪 Local Testing via Multipass (Recommended for macOS Users)

Since native Linux VMs are required for k3s, macOS developers can test
StackPulse locally using a lightweight [Multipass](https://multipass.run/)
Ubuntu VM.

Follow this step-by-step pipeline to run globally inside a local VM:

### 1. Launch a Multipass VM

Provision an Ubuntu instance meeting minimum system requirements (2 CPUs, 4GB
RAM):

```bash
multipass launch --name stackpulse-vm --cpus 2 --memory 4G --disk 20G
```

### 2. Move Binary Globally inside the VM

Compile the Linux AMD64 binary locally on your host machine, transfer it to the
VM, and move it to `/usr/local/bin` to make it globally available:

```bash
# Compile for Linux (from host machine)
env GOOS=linux GOARCH=amd64 go build -o stackpulse cmd/stackpulse/main.go

# Transfer to Multipass VM
multipass transfer stackpulse stackpulse-vm:/home/ubuntu/stackpulse

# Shell into the VM
multipass shell stackpulse-vm

# Inside the VM shell: Make it executable and move to global bin path
chmod +x /home/ubuntu/stackpulse
sudo mv /home/ubuntu/stackpulse /usr/local/bin/stackpulse
```

### 3. Verify Global Run

Now you can execute the `stackpulse` CLI globally from anywhere in the VM shell
(just like standard system commands):

```bash
sudo stackpulse doctor
```

### 4. Deploy Observability Stack

```bash
sudo stackpulse deploy observability
```

When prompt options appear, select `2` to automatically install `k3s`
lightweight Kubernetes or `1` for `kind`.

### 5. Access Dashboards from Host Browser

Once fully deployed, retrieve the service status:

```bash
sudo stackpulse status
```

StackPulse will automatically resolve the active VM interface IP. Simply open
the generated links (e.g. `http://<VM_IP>/grafana`) directly in your host
machine's web browser!

---

## 🧹 Uninstall

Remove observability stack:

```bash
sudo stackpulse uninstall observability
```

Remove everything managed by StackPulse:

```bash
sudo stackpulse uninstall all
```

---

## 📸 Screenshots

> Add screenshots after deployment.

### Grafana Kubernetes Overview

```text
docs/images/grafana-kubernetes-overview.png
```

### Loki Logs

```text
docs/images/loki-logs.png
```

### Tempo Traces

```text
docs/images/tempo-traces.png
```

### Slack Alert

```text
docs/images/slack-alert.png
```

---

## 🗂️ Repository Structure

```text
StackPulse/
├── cmd/
│   └── stackpulse/
├── internal/
│   ├── alerts/
│   ├── cli/
│   ├── config/
│   ├── doctor/
│   ├── gitops/
│   ├── helm/
│   ├── installer/
│   ├── kubernetes/
│   ├── observability/
│   ├── utils/
│   └── webhook/
├── charts/
│   └── webhook-handler/
├── configs/
├── dashboards/
├── docs/
├── scripts/
│   └── install.sh
├── .github/
│   └── workflows/
├── Dockerfile.webhook
├── go.mod
├── go.sum
├── README.md
├── ROADMAP.md
└── LICENSE
```

---

## 🛠️ Built With

- [Go](https://go.dev/)
- [Cobra](https://github.com/spf13/cobra)
- [Viper](https://github.com/spf13/viper)
- [Kubernetes](https://kubernetes.io/)
- [Helm](https://helm.sh/)
- [Prometheus](https://prometheus.io/)
- [Grafana](https://grafana.com/)
- [Loki](https://grafana.com/oss/loki/)
- [Tempo](https://grafana.com/oss/tempo/)
- [OpenTelemetry](https://opentelemetry.io/)
- [Alertmanager](https://prometheus.io/docs/alerting/latest/alertmanager/)
- [GitHub Actions](https://github.com/features/actions)

---

## 🤝 Contributing

Contributions are welcome.

```bash
git clone https://github.com/shivamshashank/StackPulse.git
cd StackPulse
go mod tidy
go test ./...
```

Create a branch:

```bash
git checkout -b feature/my-feature
```

Commit and push:

```bash
git commit -m "feat: add my feature"
git push origin feature/my-feature
```

Open a pull request.

---

## 📄 License

This project is licensed under the MIT License.

---

## 👤 Author

**Shivam Shashank**

- 🌐 Portfolio: [shivam-shashank.me](https://www.shivam-shashank.me/)
- 💼 LinkedIn:
  [shivam-shashank-2b5766217](https://www.linkedin.com/in/shivam-shashank-2b5766217/)
- 📧 Email: [shivamkumar872000@gmail.com](mailto:shivamkumar872000@gmail.com)
- 🐙 GitHub: [shivamshashank](https://github.com/shivamshashank)

---

<div align="center">

### ⭐ If StackPulse helps you, please star the repository.

</div>
