# 🚀 StackPulse — One-Tap Observability Stack for AWS

StackPulse is an open-source CLI tool that deploys a **complete production-grade observability stack** on AWS in a single command.

It sets up metrics, logs, traces, dashboards, alerting, and GitOps — fully integrated and ready to use.

---

## 🌟 Features

* ⚡ One-command deployment (`stackpulse up`)
* 📊 Metrics with Prometheus & Mimir
* 📜 Logs with Loki
* 🔍 Traces with OpenTelemetry & Tempo
* 📈 Dashboards with Grafana
* 🚨 Alerting with Alertmanager
* 🔔 Slack & PagerDuty integration
* 🔁 GitOps with ArgoCD
* ☁️ AWS-native deployment (EKS-based)
* 🔌 Extensible plugin system (Dynatrace, Datadog, etc.)

---

## 🏗️ Architecture

```
                ┌───────────────────────┐
                │      Applications     │
                └─────────┬─────────────┘
                          │
                ┌─────────▼─────────┐
                │   OpenTelemetry   │
                └──────┬─────┬──────┘
                       │     │
        ┌──────────────┘     └──────────────┐
        ▼                                   ▼
  ┌────────────┐                     ┌────────────┐
  │ Prometheus │                     │    Loki    │
  └────┬───────┘                     └────┬───────┘
       ▼                                  ▼
  ┌────────────┐                     ┌────────────┐
  │   Mimir    │                     │   Tempo    │
  └────┬───────┘                     └────┬───────┘
       └──────────────┬──────────────────┘
                      ▼
               ┌────────────┐
               │  Grafana   │
               └────┬───────┘
                    ▼
           ┌───────────────────┐
           │ Alertmanager      │
           └────┬──────────────┘
                ▼
      ┌───────────────────────┐
      │ Slack / PagerDuty     │
      └───────────────────────┘

                ┌───────────────────────┐
                │       ArgoCD          │
                │   (GitOps Engine)     │
                └───────────────────────┘
```

---

## 🛠️ Tech Stack

| Layer                   | Technology             |
| ----------------------- | ---------------------- |
| CLI                     | Go                     |
| Infrastructure          | Terraform              |
| Container Orchestration | Kubernetes (EKS)       |
| Deployment              | Helm + ArgoCD (GitOps) |
| Metrics                 | Prometheus + Mimir     |
| Logs                    | Loki                   |
| Tracing                 | OpenTelemetry + Tempo  |
| Visualization           | Grafana                |
| Alerting                | Alertmanager           |

---

## ⚡ Quick Start

### 1. Install CLI

```bash
git clone https://github.com/yourusername/stackpulse.git
cd stackpulse
go build -o stackpulse
```

---

### 2. Configure AWS

```bash
aws configure
```

---

### 3. Deploy Stack

```bash
./stackpulse up \
  --region eu-west-2 \
  --cluster-name stackpulse-cluster
```

---

## 🌐 What You Get After Deployment

StackPulse automatically provisions infrastructure and deploys the full stack.

### 🔗 Access URLs (Auto-Generated)

* 📊 Grafana → https://grafana.stackpulse.dev
* 📦 ArgoCD → https://argocd.stackpulse.dev
* 📈 Prometheus → https://prometheus.stackpulse.dev
* 📜 Loki → https://loki.stackpulse.dev
* 🔍 Tempo → https://tempo.stackpulse.dev

---

### 🔑 Default Credentials

* 👤 Grafana → admin / `<generated-password>`
* 👤 ArgoCD → admin / `<generated-password>`

---

### 📡 Integration Status

* 🟢 Prometheus → Connected
* 🟢 Loki → Connected
* 🟢 Tempo → Connected
* 🟢 Mimir → Connected
* 🟢 OpenTelemetry → Active
* 🟢 Alertmanager → Running
* 🟢 Slack → Enabled
* 🟢 PagerDuty → Enabled
* 🟢 ArgoCD → Synced

---

## 🔁 GitOps with ArgoCD

StackPulse includes ArgoCD for managing deployments declaratively.

* 🔄 Auto-sync from Git repositories
* 📦 Manage observability stack as code
* 🚀 Easy rollback and versioning
* 🔍 Visual deployment monitoring

---

## 🔔 Alert Integrations

### Slack

```bash
./stackpulse integrate slack \
  --webhook-url <your-webhook>
```

---

### PagerDuty

```bash
./stackpulse integrate pagerduty \
  --routing-key <your-key>
```

---

## 🔌 Optional Integrations

* Dynatrace
* Datadog
* New Relic

```bash
./stackpulse enable dynatrace
```

---

## 📂 Project Structure

```
stackpulse/
│
├── cmd/                # CLI entrypoint
├── internal/
│   ├── aws/           # AWS provisioning logic
│   ├── k8s/           # Kubernetes deployment
│   ├── integrations/  # Slack, PagerDuty
│   └── config/
│
├── terraform/         # Infra setup
├── helm/              # Helm charts
└── README.md
```

---

## 🧪 Roadmap

* [ ] Multi-cloud support (GCP, Azure)
* [ ] Web UI dashboard
* [ ] Auto domain + SSL
* [ ] Cost optimization insights
* [ ] AI-based anomaly detection

---

## 📸 Screenshots (Add These!)

* Grafana dashboards
* Loki logs view
* Trace visualization
* ArgoCD UI
* Slack alert example
* CLI deployment output

---

## 🤝 Contributing

Contributions are welcome!

```bash
git checkout -b feature/amazing-feature
```

---

## ⭐ Why StackPulse?

Most observability setups take hours or days.

StackPulse reduces it to:

> **One command. One stack. Full visibility + GitOps.**

---

## 📜 License

MIT License
