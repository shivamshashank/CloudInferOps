# CloudInferOps Phase 2 and Phase 3 Flow

> Phase 0 and Phase 1 are complete. This file documents the end-to-end technical
> flow, deliverables, and checklist for Phase 2 and Phase 3.

## Overview

Phase 2 expands `cloudinferops` from an observability platform CLI into a full
AI infrastructure operations CLI. Phase 3 adds the first Python FastAPI
inference gateway, model provider interface, metrics, tracing, and deployment
integration.

---

## Phase 2 - Command Model Update

### Goal

Convert the CLI wording and command shape from observability-only to AI
infrastructure operations.

### Target CLI Tree

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

### Technical Stack

- Go CLI using Cobra
- Command wiring and help text
- Kubernetes deployment orchestration through Helm values and manifests
- Existing platform deployment preserved
- New inference deployment command shell
- Status output updated for inference infrastructure
- Deprecated alias support for backwards compatibility

### Phase 2 End-to-End Flow

1. Run `cloudinferops doctor`
   - Validate OS, Docker, `kubectl`, Helm, cluster access, storage, ingress, and
     optional GPU/OLLAMA readiness.
2. Run `cloudinferops deploy platform`
   - Deploy platform dependencies: ingress, Prometheus, Grafana, Loki, Tempo,
     OpenTelemetry Collector, Alertmanager, ArgoCD.
3. Run `cloudinferops deploy inference --dry-run`
   - Validate inference deployment plan without creating resources.
4. Run `cloudinferops deploy inference --provider ollama --model llama3`
   - Deploy inference gateway and model backend.
5. Run `cloudinferops status`
   - Confirm inference service health, gateway, and platform endpoints.
6. Run `cloudinferops models list`
   - Validate configured models and gateway discovery.

### Phase 2 Checklist

- [x] Rename `deploy observability` to `deploy platform`
- [x] Keep `deploy observability` as a deprecated alias
- [x] Add `deploy inference` command with dry-run support
- [x] Add `models list` command with gateway/local config discovery
- [x] Add `benchmark run` and `benchmark report` command shells
- [x] Update `status` output to include inference service status
- [x] Update CLI help text and command descriptions
- [x] Add command wiring tests for new CLI tree
- [x] Add deprecated alias regression tests
- [x] Ensure `cloudinferops --help` and `cloudinferops version` work
- [x] Preserve existing platform deployment behavior

---

## Phase 3 - Python FastAPI Inference Gateway

### Goal

Add the first AI inference service with a Python FastAPI gateway, provider
interface, metrics, tracing, and Kubernetes deployment integration.

### Required Components

- Python FastAPI app for inference traffic
- Provider abstraction for Ollama and later vLLM
- Health, models, completions, and metrics endpoints
- Prometheus-compatible metrics exposition
- OpenTelemetry tracing integration
- Docker image and Kubernetes manifests

### Required Endpoints

- `GET /health`
- `GET /models`
- `POST /v1/chat/completions`
- `GET /metrics`

### Minimum Runtime Behavior

- `/health` returns service health status
- `/models` returns available configured models
- `/v1/chat/completions` supports request/response completion semantics
- `/metrics` exposes Prometheus metrics for scraping
- Provider interface supports Ollama initially, with vLLM as a later extension

### Core Metrics

- `cloudinferops_inference_requests_total`
- `cloudinferops_inference_errors_total`
- `cloudinferops_inference_latency_seconds`
- `cloudinferops_inference_tokens_total`
- `cloudinferops_inference_tokens_per_second`
- `cloudinferops_inference_ttft_seconds`
- `cloudinferops_inference_model_requests_total`

### Phase 3 End-to-End Flow

1. Deploy the FastAPI inference service via `cloudinferops deploy inference`.
2. Confirm the gateway pod is running and the service is reachable.
3. Call `GET /health` to validate the API is live.
4. Call `GET /models` to validate configured model metadata.
5. Send a request to `POST /v1/chat/completions` and verify inference output.
6. Scrape `GET /metrics` from Prometheus and confirm metrics are exposed.
7. Observe logs and traces in Loki and Tempo / OpenTelemetry.
8. Use `cloudinferops status` to verify inference service state.

### Python Stack

- `fastapi`
- `uvicorn`
- `httpx`
- `pydantic`
- `prometheus-client`
- `opentelemetry-api`
- `opentelemetry-sdk`
- `opentelemetry-instrumentation-fastapi`
- `pytest`
- `pytest-asyncio`

### Deployment and Integration

- Dockerfile for the FastAPI app
- Kubernetes Deployment and Service for inference gateway
- Ingress route for internal/external access
- Prometheus scrape annotations on the service
- ConfigMap for model routing configuration
- Liveness/readiness probes
- Optional GPU support path for vLLM provider
- Optional Ollama local inference path

### Phase 3 Checklist

- [ ] Create Python FastAPI gateway package structure
- [ ] Implement health endpoint
- [ ] Implement model listing endpoint
- [ ] Implement completions endpoint
- [ ] Implement metrics endpoint
- [ ] Add provider interface with Ollama support
- [ ] Add Docker packaging and Python dependencies
- [ ] Add tests for health, models, chat, and routing
- [ ] Add Prometheus metric instrumentation
- [ ] Add OpenTelemetry tracing for requests
- [ ] Add Kubernetes deployment manifests for inference service
- [ ] Validate the FastAPI app starts locally
- [ ] Validate health endpoint
- [ ] Validate mock provider tests

---

## End-to-End Tech Stack Summary

- Go / Cobra CLI for operations and deployment orchestration
- Helm / Kubernetes for platform and inference deployment
- Python / FastAPI for inference gateway
- Ollama provider for local LLM hosting
- Optional vLLM provider for GPU-backed inference
- Prometheus and Grafana for metrics and dashboards
- OpenTelemetry / Tempo for tracing
- Loki for log aggregation
- ArgoCD for GitOps delivery
- Alertmanager for SLO/alerting
- pytest for Python validation
- Go test coverage for CLI and deployment logic

---

## Validation and Acceptance

- CLI shape and commands must reflect the AI infrastructure operations plan.
- Platform deployment remains functional after the command model update.
- Inference deployment command must be available and support `--help`.
- FastAPI gateway must expose the required endpoints and metrics.
- Metrics and tracing must be wired so platform observability can capture
  inference traffic.
- Existing repository behavior should remain intact while enabling the new AI
  inference path.

---

## Notes

This file is intentionally focused on **Phase 2 and Phase 3 only** and can be
used as a standalone implementation checklist for the next work phase.
