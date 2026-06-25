# CloudInferOps Migration Plan

## Project Identity

### Final Name

**CloudInferOps**

### GitHub Repository Name

```text
cloud-infer-ops
```

### GitHub Description

```text
Cloud-native multi-cloud AI inference operations CLI for deploying, monitoring, benchmarking, and scaling LLM workloads on Kubernetes.
```

### Short README Tagline

```text
CloudInferOps - Multi-Cloud AI Inference Operations CLI
```

### One-Line Project Pitch

CloudInferOps is a Go and Python platform-engineering project that deploys and
operates production-style LLM inference workloads across local and cloud
Kubernetes environments with observability, GitOps, benchmarking, autoscaling,
and incident workflows.

### Resume Pitch

Built CloudInferOps, a multi-cloud AI infrastructure CLI that combines a Go
Kubernetes control plane with Python FastAPI inference services to deploy,
monitor, benchmark, and operate LLM workloads using Kubernetes, Helm, ArgoCD,
Prometheus, Grafana, OpenTelemetry, Ollama, and optional vLLM.

## Why Extend CloudInferOps Instead Of Starting Over

CloudInferOps already has the hard platform foundation:

- Go CLI using Cobra.
- Kubernetes readiness and system diagnostics.
- Local cluster bootstrap support for kind, minikube, and k3s.
- Helm-based deployment engine.
- Prometheus, Grafana, Loki, Tempo, OpenTelemetry, Alertmanager, ArgoCD, and
  GitOps flows.
- Alert and incident webhook handling.
- Status commands and test coverage.

CloudInferOps should keep that foundation and add the AI inference layer. This
creates a stronger engineering story than a Python-only rewrite because it shows
that the project uses Go for infrastructure operations and Python for AI-serving
workflows.

## Final Outcome

At the end, CloudInferOps should allow a user to run:

```bash
cloudinferops doctor
cloudinferops deploy platform
cloudinferops deploy inference --provider ollama --model llama3
cloudinferops status
cloudinferops benchmark --model llama3 --requests 100
cloudinferops gitops status
```

The project should produce:

- A renamed and polished GitHub repository: `cloud-infer-ops`.
- A Go CLI binary named `cloudinferops`.
- A Python FastAPI inference gateway.
- Ollama-based local inference support.
- Optional vLLM deployment path for GPU-backed inference.
- Kubernetes manifests and Helm values for inference services.
- Prometheus metrics for AI workloads.
- Grafana dashboards for LLM latency, tokens/sec, TTFT, error rate, and model
  usage.
- OpenTelemetry traces across gateway, router, and inference backend.
- Benchmark reports for latency, throughput, and estimated cost.
- GitOps deployment support through ArgoCD.
- Alerting and incident routing for inference failures and SLO breaches.
- README, architecture diagrams, screenshots, and demo commands.

## End-To-End User Flow

### 1. Preflight Check

Command:

```bash
cloudinferops doctor
```

Checks:

- Operating system and CPU architecture.
- CPU, memory, disk, and ports.
- Docker availability.
- `kubectl` availability.
- Helm availability.
- Kubernetes cluster reachability.
- StorageClass availability.
- Ingress readiness.
- Optional GPU/NVIDIA runtime availability.
- Optional Ollama availability for local inference.

Expected result:

```text
CloudInferOps Doctor

[OK] OS: linux/amd64
[OK] Docker found
[OK] kubectl found
[OK] Helm found
[OK] Kubernetes cluster detected
[OK] StorageClass found
[INFO] GPU not detected; Ollama/local CPU path available

[READY] Run: cloudinferops deploy platform
```

### 2. Platform Deployment

Command:

```bash
cloudinferops deploy platform
```

Deploys:

- NGINX ingress controller.
- Prometheus.
- Grafana.
- Loki.
- Tempo.
- OpenTelemetry Collector.
- Alertmanager.
- ArgoCD.
- Base dashboards and alert rules.

Expected result:

```text
CloudInferOps platform deployed.
Grafana: http://cloudinferops.local/grafana
Prometheus: http://cloudinferops.local/prometheus
ArgoCD: http://cloudinferops.local/argocd
```

### 3. Inference Deployment

Command:

```bash
cloudinferops deploy inference --provider ollama --model llama3
```

Deploys:

- FastAPI inference gateway.
- Model router.
- Ollama model backend.
- Kubernetes Deployment.
- Kubernetes Service.
- Ingress route.
- ConfigMap for model routing.
- Prometheus scrape annotations.
- Liveness and readiness probes.

Optional GPU command:

```bash
cloudinferops deploy inference --provider vllm --model mistral --gpu
```

### 4. Request Flow

Request:

```bash
curl -X POST http://cloudinferops.local/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"llama3","messages":[{"role":"user","content":"Explain Kubernetes in simple terms."}]}'
```

Flow:

```text
User
  -> FastAPI Gateway
  -> Model Router
  -> Ollama or vLLM backend
  -> Streaming response
  -> Metrics emitted to Prometheus
  -> Traces emitted through OpenTelemetry
  -> Logs collected by Loki
  -> Dashboards updated in Grafana
  -> Alerts triggered if SLOs fail
```

### 5. Status Check

Command:

```bash
cloudinferops status
```

Shows:

- Platform health.
- Inference gateway health.
- Model backend health.
- GitOps sync status.
- Grafana and Prometheus URLs.
- Active models.
- Average latency.
- Token throughput.
- Error rate.
- Recent incidents.

### 6. Benchmarking

Command:

```bash
cloudinferops benchmark --model llama3 --requests 100 --concurrency 10
```

Reports:

- P50 latency.
- P95 latency.
- P99 latency.
- Time to first token.
- Tokens/sec.
- Requests/sec.
- Error rate.
- Estimated cost/request.
- Recommended scaling changes.

### 7. Reliability Demo

Commands:

```bash
kubectl delete pod -n inference -l app=cloudinferops-gateway
cloudinferops status
```

Expected result:

- Kubernetes recreates the pod.
- Status command detects recovery.
- Grafana shows temporary error spike.
- Alertmanager records alert if threshold is breached.
- Incident webhook stores the event.

## Target Directory Structure

Final structure:

```text
cloud-infer-ops/
├── cmd/
│   └── cloudinferops/
│       └── main.go
├── internal/
│   ├── alerts/
│   ├── benchmark/
│   ├── cli/
│   ├── config/
│   ├── doctor/
│   ├── gitops/
│   ├── helm/
│   ├── inference/
│   ├── installer/
│   ├── observability/
│   ├── platform/
│   ├── security/
│   ├── utils/
│   └── webhook/
├── api/
│   ├── app/
│   │   ├── main.py
│   │   ├── config.py
│   │   ├── metrics.py
│   │   ├── tracing.py
│   │   ├── schemas.py
│   │   ├── router.py
│   │   └── providers/
│   │       ├── base.py
│   │       ├── ollama.py
│   │       └── vllm.py
│   ├── tests/
│   ├── Dockerfile
│   ├── pyproject.toml
│   └── README.md
├── deployments/
│   ├── helm/
│   │   └── cloudinferops/
│   │       ├── Chart.yaml
│   │       ├── values.yaml
│   │       └── templates/
│   ├── kubernetes/
│   │   ├── namespace.yaml
│   │   ├── gateway-deployment.yaml
│   │   ├── gateway-service.yaml
│   │   ├── ollama-deployment.yaml
│   │   ├── vllm-deployment.yaml
│   │   └── ingress.yaml
│   └── terraform/
│       ├── aws/
│       ├── gcp/
│       └── azure/
├── observability/
│   ├── dashboards/
│   │   ├── llm-overview.json
│   │   ├── inference-latency.json
│   │   ├── token-throughput.json
│   │   └── cost-efficiency.json
│   ├── alerts/
│   │   └── inference-alerts.yaml
│   └── otel/
│       └── collector.yaml
├── benchmarking/
│   ├── scenarios/
│   │   ├── chat.yaml
│   │   ├── coding.yaml
│   │   └── reasoning.yaml
│   ├── reports/
│   └── README.md
├── docs/
│   ├── architecture.md
│   ├── request-flow.md
│   ├── multi-cloud.md
│   ├── demo-script.md
│   ├── resume.md
│   └── images/
├── scripts/
│   ├── install.sh
│   ├── dev-up.sh
│   ├── kind-up.sh
│   └── release.sh
├── .github/
│   └── workflows/
│       ├── ci.yml
│       ├── docker.yml
│       └── release.yml
├── README.md
├── CLOUDINFEROPS_MIGRATION_PLAN.md
├── ROADMAP.md
├── SECURITY.md
├── CONTRIBUTING.md
├── go.mod
├── go.sum
└── LICENSE
```

## Migration Principles

- Keep Go for CLI, Kubernetes, Helm, GitOps, status, diagnostics, and deployment
  automation.
- Add Python for FastAPI, model routing, LLM provider adapters, streaming, and
  AI-specific metrics.
- Avoid renaming everything in one risky commit; migrate in small phases.
- Keep old command behavior working until equivalent CloudInferOps commands are
  tested.
- Prefer compatibility aliases during transition, then remove legacy names near
  the end.
- Make every major feature demoable from the CLI.
- Keep all implementation tied to tests and README updates.

## Phase 0 - Repository Preparation

Goal: Create a safe branch and capture current behavior.

Tasks:

- Create branch:

```bash
git checkout -b codex/cloudinferops-migration
```

- Run current tests:

```bash
env GOCACHE=/private/tmp/cloudinferops-go-cache go test ./...
```

- Record current command list:

```bash
go run ./cmd/cloudinferops version
go run ./cmd/cloudinferops --help
```

- Add this migration plan to the repo.
- Decide whether `CloudInferOps` should remain as a historical note in README.

Acceptance criteria:

- Current tests pass or known sandbox-only failures are documented.
- Migration branch exists.
- No functional rename has happened yet.

## Phase 1 - Product Rename

Goal: Rename public project identity from CloudInferOps to CloudInferOps.

Files and directories to update:

```text
cmd/cloudinferops/                  -> cmd/cloudinferops/
go.mod module path               -> github.com/shivamshashank/cloud-infer-ops
README.md                        -> CloudInferOps identity
scripts/install.sh               -> cloudinferops binary
.github/workflows/*.yml          -> new binary/repo names
internal/cli/root.go             -> root command Use: cloudinferops
internal/config/config.go        -> ~/.cloudinferops config directory
internal/observability/*.go      -> public messages and release names where appropriate
internal/gitops/*.go             -> generated app names where appropriate
internal/webhook/*.go            -> Kubernetes service/image names
```

Recommended compatibility approach:

- Rename binary to `cloudinferops`.
- Keep a temporary `cloudinferops` alias script or compatibility note for one
  release only.
- Keep Kubernetes resource renames scoped and deliberate because renaming live
  resources creates migration complexity.

Suggested command changes:

```text
cloudinferops doctor                -> cloudinferops doctor
cloudinferops init                  -> cloudinferops init
cloudinferops deploy observability  -> cloudinferops deploy platform
cloudinferops status                -> cloudinferops status
cloudinferops gitops bootstrap      -> cloudinferops gitops bootstrap
cloudinferops alerts                -> cloudinferops alerts
```

Implementation tasks:

- Rename `cmd/cloudinferops` to `cmd/cloudinferops`.
- Change Cobra root `Use` to `cloudinferops`.
- Change root descriptions from Kubernetes observability platform to AI
  inference operations CLI.
- Rename config directory from `~/.cloudinferops` to `~/.cloudinferops`.
- Add migration fallback that reads old `~/.cloudinferops/config.yaml` if the new
  config does not exist.
- Update install script to install `/usr/local/bin/cloudinferops`.
- Update docs and examples.

Tests:

```bash
env GOCACHE=/private/tmp/cloudinferops-go-cache go test ./...
go run ./cmd/cloudinferops --help
go run ./cmd/cloudinferops version
```

Acceptance criteria:

- `cloudinferops --help` works.
- `cloudinferops version` works.
- Existing tests pass.
- README no longer presents the project as CloudInferOps.

## Phase 2 - Command Model Update

Goal: Convert the CLI from observability-only wording to AI infrastructure
operations.

Target command tree:

```text
cloudinferops
├── init
├── doctor
├── deploy
│   ├── platform
│   ├── inference
│   └── webhook-handler
├── status
├── connect
├── gitops
│   ├── bootstrap
│   └── status
├── alerts
│   ├── configure
│   └── test
├── benchmark
│   ├── run
│   └── report
├── models
│   ├── list
│   └── pull
├── security
│   ├── scan
│   └── report
└── version
```

Implementation tasks:

- Rename `deploy observability` to `deploy platform`.
- Keep `deploy observability` as a deprecated alias initially.
- Add `deploy inference` command with dry-run support.
- Add `models list` command to call the deployed gateway or local config.
- Add `benchmark run` and `benchmark report` command shells.
- Update `status` output to include inference services even before they are
  built.

Tests:

- Command wiring tests for each new command.
- Deprecated alias test for `deploy observability`.
- Help text tests where practical.

Acceptance criteria:

- CLI shape reflects AI infrastructure operations.
- Existing platform deployment still works.
- New commands can be called with `--help`.

## Phase 3 - Python FastAPI Inference Gateway

Goal: Add the first real AI inference service.

Create:

```text
api/
├── app/
│   ├── main.py
│   ├── schemas.py
│   ├── router.py
│   ├── metrics.py
│   ├── tracing.py
│   └── providers/
│       ├── base.py
│       ├── ollama.py
│       └── vllm.py
├── tests/
│   ├── test_health.py
│   ├── test_models.py
│   ├── test_chat.py
│   └── test_router.py
├── Dockerfile
├── pyproject.toml
└── README.md
```

Required endpoints:

```http
GET  /health
GET  /models
POST /v1/chat/completions
GET  /metrics
```

Minimum behavior:

- `/health` returns service status.
- `/models` returns configured models.
- `/v1/chat/completions` supports non-streaming responses first.
- `/metrics` exposes Prometheus-compatible metrics.
- Provider interface supports Ollama first and vLLM later.

Core metrics:

```text
cloudinferops_inference_requests_total
cloudinferops_inference_errors_total
cloudinferops_inference_latency_seconds
cloudinferops_inference_tokens_total
cloudinferops_inference_tokens_per_second
cloudinferops_inference_ttft_seconds
cloudinferops_inference_model_requests_total
```

Python dependencies:

```text
fastapi
uvicorn
httpx
pydantic
prometheus-client
opentelemetry-api
opentelemetry-sdk
opentelemetry-instrumentation-fastapi
pytest
pytest-asyncio
```

Tests:

```bash
cd api
pytest
```

Acceptance criteria:

- FastAPI app starts locally.
- Health endpoint passes.
- Mock provider tests pass.
- Metrics endpoint exposes inference counters.

## Phase 4 - Local Ollama Inference Path

Goal: Make the project demoable without GPU access.

Implementation tasks:

- Add Ollama provider in Python.
- Add CLI checks for Ollama availability.
- Add `cloudinferops models pull llama3`.
- Add `cloudinferops deploy inference --provider ollama --model llama3`.
- Add Docker/Kubernetes manifests for the gateway.
- Decide whether Ollama runs inside Kubernetes or externally for local mode.

Recommended first version:

- Gateway runs in Kubernetes.
- Ollama can run locally on the host or as an optional Kubernetes deployment.
- Config value controls backend URL.

Example config:

```yaml
inference:
  provider: ollama
  model: llama3
  gateway:
    replicas: 1
  ollama:
    url: http://ollama.inference.svc.cluster.local:11434
```

Acceptance criteria:

- User can deploy local inference demo.
- User can call `/v1/chat/completions`.
- Metrics update after requests.
- Grafana can scrape/display request metrics.

## Phase 5 - Kubernetes Inference Deployment

Goal: Deploy the AI gateway as part of the platform.

Create manifests:

```text
deployments/kubernetes/namespace.yaml
deployments/kubernetes/gateway-deployment.yaml
deployments/kubernetes/gateway-service.yaml
deployments/kubernetes/gateway-ingress.yaml
deployments/kubernetes/ollama-deployment.yaml
deployments/kubernetes/ollama-service.yaml
deployments/kubernetes/model-config.yaml
```

Deployment requirements:

- Namespace: `inference`.
- Gateway Deployment with readiness and liveness probes.
- Gateway Service.
- Ingress route `/v1`.
- Prometheus scrape annotations.
- ConfigMap for model routing rules.
- Resource requests and limits.
- Optional persistent volume for Ollama models.

CLI implementation:

- Add Go package `internal/inference`.
- Render/apply manifests through Go.
- Add dry-run mode.
- Add tests using injectable command runners.

Acceptance criteria:

- `cloudinferops deploy inference --dry-run` prints planned resources.
- `cloudinferops deploy inference` applies manifests.
- `cloudinferops status` displays inference health.

## Phase 6 - AI Observability Dashboards

Goal: Turn the existing observability foundation into AI-specific observability.

Dashboards to add:

```text
observability/dashboards/llm-overview.json
observability/dashboards/inference-latency.json
observability/dashboards/token-throughput.json
observability/dashboards/model-usage.json
observability/dashboards/cost-efficiency.json
```

Dashboard panels:

- Total inference requests.
- Error rate.
- P50/P95/P99 latency.
- Time to first token.
- Tokens/sec.
- Requests by model.
- Requests by provider.
- Inference backend health.
- Estimated cost/request.
- Pod CPU and memory.
- Optional GPU utilization.

Alert rules:

```text
HighInferenceErrorRate
HighInferenceLatencyP95
LowTokenThroughput
InferenceGatewayDown
ModelBackendDown
HighCostPerRequest
```

Acceptance criteria:

- Dashboards are deployed with platform.
- Inference requests appear in Grafana.
- Alert rules are visible in Prometheus/Alertmanager.

## Phase 7 - Benchmarking Engine

Goal: Add a clear performance and cost story.

Go CLI commands:

```bash
cloudinferops benchmark run --model llama3 --requests 100 --concurrency 10
cloudinferops benchmark report --latest
```

Benchmark dimensions:

- Model.
- Prompt class.
- Request count.
- Concurrency.
- Streaming or non-streaming.
- Provider.
- Input tokens.
- Output tokens.

Metrics to calculate:

- P50 latency.
- P95 latency.
- P99 latency.
- Time to first token.
- Tokens/sec.
- Requests/sec.
- Error rate.
- Cost/request estimate.

Report outputs:

```text
benchmarking/reports/latest.json
benchmarking/reports/latest.md
benchmarking/reports/latest.html
```

Acceptance criteria:

- Benchmark can run against local gateway.
- JSON and Markdown reports are generated.
- README includes a sample benchmark result.

## Phase 8 - Optional vLLM GPU Path

Goal: Add production-grade inference credibility.

Implementation tasks:

- Add vLLM provider adapter in Python.
- Add Kubernetes deployment manifest for vLLM.
- Add GPU scheduling options.
- Add CLI flags:

```bash
cloudinferops deploy inference --provider vllm --model mistral --gpu
```

Config:

```yaml
inference:
  provider: vllm
  model: mistral
  vllm:
    image: vllm/vllm-openai:latest
    gpu: true
    tensorParallelSize: 1
```

Acceptance criteria:

- Dry-run shows vLLM deployment.
- Documentation explains GPU requirement.
- Gateway provider interface supports vLLM-compatible OpenAI API.

## Phase 9 - Multi-Cloud Positioning

Goal: Make the "Cloud" part real without overbuilding.

Supported environments:

- Local Docker Desktop Kubernetes.
- kind.
- minikube.
- k3s on Linux VM.
- AWS EC2 with k3s.
- AWS EKS path through Terraform as an advanced option.
- Future: GKE and AKS docs.

Terraform structure:

```text
deployments/terraform/
├── aws/
│   ├── eks/
│   └── ec2-k3s/
├── gcp/
│   └── gke/
└── azure/
    └── aks/
```

Minimum credible version:

- Document local, EC2+k3s, and EKS.
- Add Terraform skeleton for AWS EKS.
- Do not claim full GCP/Azure support until tested.

Acceptance criteria:

- README has a tested local path.
- Docs have an AWS path.
- Multi-cloud claims are honest and specific.

## Phase 10 - GitOps Integration

Goal: Use existing ArgoCD strength for inference workloads.

Implementation tasks:

- Extend GitOps repo generation to include inference manifests.
- Add `cloudinferops gitops bootstrap --with-inference`.
- Add ArgoCD Application for inference namespace.
- Add model routing ConfigMap to GitOps-managed templates.

Generated GitOps apps:

```text
cloudinferops-platform
cloudinferops-observability
cloudinferops-inference
cloudinferops-apps
```

Acceptance criteria:

- ArgoCD shows inference app synced and healthy.
- `cloudinferops gitops status` includes inference workloads.

## Phase 11 - Security And Reliability

Goal: Make the project look production aware.

Security features:

- Kubernetes manifest checks.
- Image tag warnings for `latest`.
- Resource limit checks.
- Privileged container detection.
- Public ingress warnings.
- Secret presence checks.
- Optional Trivy integration.
- Optional kube-score integration.

Reliability features:

- Readiness probes.
- Liveness probes.
- Graceful shutdown in FastAPI service.
- Retry and timeout handling in provider clients.
- Circuit breaker for unhealthy providers.
- HPA examples.
- Failure demo docs.

Acceptance criteria:

- `cloudinferops security scan` includes inference namespace.
- Gateway handles backend timeout gracefully.
- Alerts fire for gateway/backend failure.

## Phase 12 - README And Documentation Polish

Goal: Make the repo instantly understandable to recruiters and engineers.

README sections:

- Project title and badges.
- What is CloudInferOps?
- Why this exists.
- Architecture diagram.
- Features.
- Quick start.
- CLI commands.
- End-to-end demo.
- Screenshots.
- Benchmark example.
- Kubernetes/GitOps architecture.
- Local demo path.
- AWS demo path.
- Tech stack.
- Testing.
- Roadmap.
- Resume highlights.

Recommended README badges:

```md
![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![Python](https://img.shields.io/badge/Python-3776AB?style=for-the-badge&logo=python&logoColor=white)
![FastAPI](https://img.shields.io/badge/FastAPI-009688?style=for-the-badge&logo=fastapi&logoColor=white)
![Kubernetes](https://img.shields.io/badge/Kubernetes-326CE5?style=for-the-badge&logo=kubernetes&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white)
![Helm](https://img.shields.io/badge/Helm-0F1689?style=for-the-badge&logo=helm&logoColor=white)
![ArgoCD](https://img.shields.io/badge/ArgoCD-EF7B4D?style=for-the-badge&logo=argo&logoColor=white)
![Prometheus](https://img.shields.io/badge/Prometheus-E6522C?style=for-the-badge&logo=prometheus&logoColor=white)
![Grafana](https://img.shields.io/badge/Grafana-F46800?style=for-the-badge&logo=grafana&logoColor=white)
![OpenTelemetry](https://img.shields.io/badge/OpenTelemetry-000000?style=for-the-badge&logo=opentelemetry&logoColor=white)
![Ollama](https://img.shields.io/badge/Ollama-Local%20LLM-black?style=for-the-badge)
![vLLM](https://img.shields.io/badge/vLLM-GPU%20Inference-blue?style=for-the-badge)
```

Docs to create:

```text
docs/architecture.md
docs/request-flow.md
docs/multi-cloud.md
docs/demo-script.md
docs/resume.md
docs/screenshots.md
```

Acceptance criteria:

- A recruiter can understand the project in 60 seconds.
- An engineer can run the local demo from README.
- Architecture and request flow are clear.

## Phase 13 - CI/CD And Release

Goal: Make the project look maintained and production-grade.

GitHub Actions:

```text
.github/workflows/ci.yml
.github/workflows/docker.yml
.github/workflows/release.yml
```

CI should run:

- Go tests.
- Go vet.
- Go formatting check.
- Python tests.
- Python linting.
- Docker build for API gateway.
- Helm/template validation if available.

Release should build:

- `cloudinferops-darwin-amd64`
- `cloudinferops-darwin-arm64`
- `cloudinferops-linux-amd64`
- `cloudinferops-linux-arm64`

Acceptance criteria:

- Pull requests run Go and Python tests.
- Releases publish CLI binaries.
- Docker image builds for the gateway.

## Phase 14 - Demo Package

Goal: Make the project easy to show.

Demo script:

```bash
cloudinferops doctor
cloudinferops deploy platform --dry-run
cloudinferops deploy inference --provider ollama --model llama3 --dry-run
cloudinferops deploy platform
cloudinferops deploy inference --provider ollama --model llama3
cloudinferops status
curl http://cloudinferops.local/v1/chat/completions
cloudinferops benchmark run --model llama3 --requests 25 --concurrency 5
cloudinferops benchmark report --latest
cloudinferops gitops status
```

Screenshots to capture:

- CLI doctor output.
- CLI deploy output.
- CLI status output.
- FastAPI docs page.
- Grafana LLM overview dashboard.
- Grafana latency dashboard.
- Prometheus targets.
- ArgoCD applications.
- Benchmark report.

Video demo structure:

```text
1. Problem: AI inference is hard to operate.
2. Architecture: Go CLI + Python gateway + Kubernetes + observability.
3. Deploy platform.
4. Deploy inference.
5. Send LLM request.
6. Show metrics and traces.
7. Run benchmark.
8. Show reliability failure/recovery.
9. Explain resume impact.
```

Acceptance criteria:

- Demo can be completed in 10-15 minutes.
- Screenshots are in `docs/images`.
- README references the screenshots.

## Phase 15 - Resume And Portfolio Output

Goal: Convert the engineering work into a clear career signal.

Resume project title:

```text
CloudInferOps - Multi-Cloud AI Inference Operations CLI
```

Resume bullets:

```text
- Built CloudInferOps, a Go and Python AI infrastructure CLI for deploying, monitoring, benchmarking, and operating LLM inference workloads on Kubernetes.
- Implemented a Go-based Kubernetes control plane for cluster diagnostics, Helm deployment, GitOps bootstrap, alerting, and status reporting across local and cloud environments.
- Developed a FastAPI inference gateway with Ollama/vLLM provider adapters, OpenAI-compatible chat endpoints, Prometheus metrics, and OpenTelemetry tracing.
- Added AI-specific observability dashboards for request latency, TTFT, token throughput, model usage, error rate, and backend health using Prometheus and Grafana.
- Created benchmarking workflows that report P50/P95/P99 latency, requests/sec, tokens/sec, error rate, and estimated cost/request for LLM workloads.
```

Portfolio outcome:

- GitHub repository.
- README with screenshots.
- Architecture diagram.
- Demo video.
- Benchmark report.
- Technical blog post.
- Resume bullets.

## Recommended Implementation Order

If time is limited, build in this order:

1. Rename CLI and README identity.
2. Add `deploy platform` alias for existing observability deployment.
3. Add FastAPI gateway with mock provider.
4. Add Ollama provider.
5. Add Kubernetes manifests for gateway.
6. Add Prometheus metrics.
7. Add Grafana LLM dashboard.
8. Add benchmark command and report.
9. Add GitOps inference app.
10. Add vLLM dry-run support.
11. Polish README/screenshots/demo.

## Minimum Viable CloudInferOps

The smallest version that is still resume-worthy:

- `cloudinferops` Go CLI.
- `doctor`, `deploy platform`, `deploy inference`, `status`, and `benchmark`.
- Python FastAPI gateway.
- Ollama provider.
- Kubernetes deployment for gateway.
- Prometheus metrics.
- One Grafana LLM overview dashboard.
- Benchmark Markdown report.
- README with architecture and demo.

Avoid spending too early on:

- Full AWS/GCP/Azure support.
- Complex UI.
- Multi-tenant auth.
- Full vLLM GPU testing if no GPU is available.
- Production billing logic.

## Definition Of Done

CloudInferOps is complete when:

- The repository is named `cloud-infer-ops`.
- The binary is named `cloudinferops`.
- README explains the project as an AI inference operations platform.
- A user can deploy platform observability.
- A user can deploy local LLM inference.
- A user can send an inference request.
- Metrics appear in Prometheus/Grafana.
- A benchmark report can be generated.
- GitOps status can be shown.
- Tests cover Go CLI and Python gateway.
- The demo flow is documented and reproducible.
- Resume bullets are included in `docs/resume.md`.

## Suggested GitHub Topics

```text
ai-infrastructure
llmops
mlops
kubernetes
golang
python
fastapi
vllm
ollama
prometheus
grafana
opentelemetry
gitops
argocd
sre
devops
cloud-native
benchmarking
observability
platform-engineering
```

## Suggested Repo Social Preview Text

```text
CloudInferOps
Multi-Cloud AI Inference Operations CLI

Deploy LLMs. Observe tokens. Benchmark latency. Operate AI workloads on Kubernetes.
```

## Final Architecture Summary

```text
Developer or Platform Engineer
  -> cloudinferops CLI
  -> Kubernetes / Helm / GitOps control plane
  -> FastAPI inference gateway
  -> Model router
  -> Ollama or vLLM backend
  -> Prometheus metrics
  -> Grafana dashboards
  -> Loki logs
  -> Tempo traces
  -> Alertmanager incidents
  -> Benchmark reports
```

CloudInferOps should feel like a real internal platform tool an AI
infrastructure team would use to bootstrap, observe, benchmark, and operate LLM
workloads.
