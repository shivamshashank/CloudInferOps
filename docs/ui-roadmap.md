# CloudInferOps UI and Self-Hosted Portal Roadmap

## 1. Product Goal

Create a simple, self-hosted web portal for CloudInferOps so that after installation and deployment, users can access a dashboard at:

- http://<public-ip>/cloudinferops

The experience should feel like this:

1. User installs the CLI
2. User runs the deployment command
3. The system provisions the stack
4. The UI becomes available automatically
5. The user sees deployment health, model status, logs, and observability from the browser

The goal is to make CloudInferOps feel like a complete control plane rather than just a CLI tool.

---

## 2. Vision

CloudInferOps should provide:

- a free, self-hosted control plane for AI inference operations
- a browser-based dashboard for Kubernetes-native inference workloads
- a simple path from install to deployment to visibility
- no login requirement for local/self-hosted usage

---

## 3. Recommended Architecture

### Recommended stack

- Frontend: React
- Backend: Go
- Data source: Kubernetes API and existing CloudInferOps logic
- Ingress route: /cloudinferops

### Why this stack

- React gives you a fast and modern dashboard experience
- Go fits naturally with the existing CloudInferOps backend and CLI implementation
- Kubernetes APIs are easy to call from Go
- The UI can be served behind the existing ingress system without much complexity

---

## 4. Product Scope

### Phase 1 — Minimal dashboard
A lightweight portal that shows:

- cluster status
- deployed services
- inference gateway status
- models list
- alerts and incidents
- logs summary
- deployment health

### Phase 2 — Self-hosted portal
Expose the UI at:

- http://<public-ip>/cloudinferops

This portal should include pages for:

- Overview
- Deployments
- Models
- Observability
- Alerts

### Phase 3 — Control plane actions
Allow the user to perform basic operations from the UI:

- deploy inference
- deploy observability
- restart services
- view logs
- trigger benchmark
- view benchmark report

---

## 5. MVP Feature Checklist

### Completed in this milestone

- [x] Login-free local/self-hosted portal foundation
- [x] Overview dashboard
- [x] Cluster health status display
- [x] Gateway status display
- [x] Inference services status display
- [x] List of deployed models
- [x] Deploy action entry point
- [x] Logs endpoint foundation
- [x] Basic health indicators and modern UI styling

### Still pending

- [ ] Real Kubernetes-backed data instead of seeded/default values
- [ ] Full logs viewer with pod selection and stream support
- [ ] Auth or local-only guardrails for non-local environments
- [ ] Real deployment action that triggers actual cluster operations
- [ ] Container image build and publish flow
- [ ] Auto-install and auto-provision command for the UI
- [ ] Dedicated pages for Observability and Alerts
- [ ] Benchmark results page
- [ ] Config editor page
- [ ] One-click redeploy
- [ ] Service restart action
- [ ] Search/filter in logs and models

---

## 6. Suggested Pages

### 6.1 Overview page
Show a summary of the system:

- cluster connected or not
- namespace status
- gateway health
- observability stack health
- number of running pods
- latest incident or alert
- last deployment status

### 6.2 Deployments page
Show:

- inference deployment status
- observability deployment status
- webhook handler status
- deployment history
- buttons to redeploy or refresh

### 6.3 Models page
Show:

- available models
- active models
- pulled models
- provider type such as Ollama or vLLM
- model status and location

### 6.4 Observability page
Show:

- Grafana status
- Prometheus status
- Loki status
- Tempo status
- alert rules status
- dashboard availability

### 6.5 Alerts page
Show:

- active incidents
- alert history
- recent Slack/PagerDuty notifications
- incident severity
- timestamps

---

## 7. Recommended Implementation Plan

## Step 1 — Create the UI structure

Create a new frontend folder in the repository, for example:

- web/
  - src/
  - public/
  - package.json
  - vite.config.ts

Use React with a minimal dashboard layout.

### Suggested initial components

- AppShell
- Sidebar
- OverviewPage
- DeploymentsPage
- ModelsPage
- ObservabilityPage
- AlertsPage
- StatusCard
- TableView
- EmptyState

---

## Step 2 — Build a Go API backend

Create a backend service that exposes API endpoints for the UI.

Suggested folder structure:

- cmd/ui-api/main.go
- internal/ui/
  - handler.go
  - service.go
  - types.go

### Suggested API endpoints

- GET /api/health
- GET /api/overview
- GET /api/deployments
- GET /api/models
- GET /api/observability
- GET /api/alerts
- POST /api/deploy/inference
- POST /api/deploy/observability
- POST /api/restart/:service
- GET /api/logs/:service

These endpoints can call existing internal logic already used by the CLI.

---

## Step 3 — Reuse existing CloudInferOps logic

Instead of rewriting the system logic, connect the UI to the same internal packages that the CLI already uses.

Good candidates:

- internal/doctor
- internal/inference
- internal/observability
- internal/webhook
- internal/config
- internal/utils

The UI backend should act as a thin adapter over these packages.

This keeps the architecture consistent and avoids duplication.

---

## Step 4 — Create a simple ingress route

Expose the UI under the path:

- /cloudinferops

Use an ingress rule that routes traffic to the frontend service.

### Example ingress idea

- path: /cloudinferops
- service: cloudinferops-ui

If the UI is served as a single-page app, make sure frontend routing is handled correctly.

---

## Step 5 — Add a container image

Create a Dockerfile for the UI frontend and backend.

Example target images:

- ghcr.io/<org>/cloudinferops-ui
- ghcr.io/<org>/cloudinferops-ui-api

If you want the simplest setup, you can also serve the React build from a small Go server.

---

## Step 6 — Add deployment manifests

Create Kubernetes manifests for:

- deployment for the UI backend
- deployment for the frontend or static assets
- service for the UI
- ingress route for /cloudinferops

You can place them under:

- deployments/kubernetes/

Example names:

- cloudinferops-ui.yaml
- cloudinferops-ui-service.yaml
- cloudinferops-ui-ingress.yaml

---

## Step 7 — Integrate with existing CLI

Add CLI commands or helper flows so the deployment process is smooth.

You can make the CLI automatically:

- deploy the UI service
- expose the ingress route
- print the access URL after deployment

Suggested command:

- cloudinferops deploy ui

This should be the bridge between the terminal workflow and the browser portal.

---

## 8. Recommended UI Data Flow

### Overview data
The backend can collect:

- current namespace
- pods
- services
- deployment state
- ingress state
- observability status

### Models data
The backend can call:

- /models from the gateway
- /api/tags from Ollama

### Alerts data
The backend can read from:

- webhook incident memory or persisted storage
- alertmanager data if available

### Logs summary
The backend can return recent log excerpts from selected pods.

---

## 9. Suggested MVP UI Screens

### 9.1 Dashboard overview
Show cards such as:

- Cluster: Healthy / Unhealthy
- Gateway: Running / Not Running
- Observability: Healthy / Partial
- Models: 5 available
- Alerts: 2 active

### 9.2 Deployments screen
Show a table with:

- service name
- namespace
- ready status
- replicas
- last updated

### 9.3 Models screen
Show a table with:

- model name
- provider
- status
- location

### 9.4 Logs screen
Show:

- pod name
- recent log lines
- refresh button

### 9.5 Alerts screen
Show:

- incident title
- severity
- created time
- status

---

## 10. Backend Implementation Details

### 10.1 Health endpoint
Return a simple response:

```json
{
  "status": "ok",
  "version": "dev"
}
```

### 10.2 Overview endpoint
Return a structured summary:

```json
{
  "cluster": "connected",
  "gateway": "healthy",
  "observability": "healthy",
  "models": 5,
  "alerts": 2
}
```

### 10.3 Deployment endpoint
Trigger a deployment action using the existing internal deployer packages.

### 10.4 Logs endpoint
Return the last 100 lines for a selected pod.

---

## 11. Security Considerations

Because this is a self-hosted portal, keep it simple and safe:

- no mandatory login for local deployment
- optional basic auth later
- avoid exposing secrets in the UI
- only show non-sensitive operational information by default

Later, if you want enterprise readiness, you can add:

- auth
- RBAC-aware access control
- token-based API protection
- role-based views

---

## 12. Suggested Development Phases

### Phase 1 — Prototype UI (Completed)
Goal:

- serve a polished dashboard from a simple web app
- display structured data and UI states
- verify route and deployment structure

### Phase 2 — Real backend integration (In progress)
Goal:

- connect live data from Kubernetes and CloudInferOps logic
- show real status cards and tables

### Phase 3 — Deployment actions (In progress)
Goal:

- add buttons to deploy or refresh services
- make UI useful for operations

### Phase 4 — Polish and packaging (Planned)
Goal:

- make it easy to install and expose under /cloudinferops
- add screenshots and demo instructions
- package the UI for real cluster deployment

---

## 13. Suggested File Structure

```text
web/
  src/
    components/
    pages/
    api/
    styles/
  public/
  package.json
  vite.config.ts

cmd/ui-api/main.go
internal/ui/
  handler.go
  service.go
  types.go

deployments/kubernetes/
  ui-deployment.yaml
  ui-service.yaml
  ui-ingress.yaml
```

---

## 14. Recommended First Milestone

Completed:

- a modern React dashboard
- a Go backend with overview, deployment, model, alert, and log endpoints
- ingress route structure for /cloudinferops
- deployment manifests for the UI portal

Next:

- switch the backend from seeded/default state to live Kubernetes status
- add a real deploy action that triggers actual operations
- add a true log viewer experience

This gives you a strong first milestone and reduces the risk of building too much too soon.

---

## 15. Suggested Launch Experience

Once implemented, the flow should be:

1. User installs CloudInferOps
2. User runs deployment command
3. The UI is deployed automatically
4. The browser opens the portal at:
   - http://<public-ip>/cloudinferops
5. The user sees the operational dashboard instantly

That is the experience you want to market.

---

## 16. Marketing Angle for the UI

This UI becomes a strong story for LinkedIn and community growth:

- “From CLI install to live inference operations portal”
- “Self-hosted observability and deployment control plane for LLM workloads”
- “Free platform to deploy, monitor, and manage inference workloads on Kubernetes”

---

## 17. Recommended Next Actions

### Immediate next steps

- [ ] Create the frontend project structure
- [ ] Create the Go API service
- [ ] Add one overview dashboard page
- [ ] Add one deployments page
- [ ] Add a simple ingress route
- [ ] Connect the backend to live Kubernetes status
- [ ] Add a CLI command to deploy the UI

### Short-term goals

- [ ] Show cluster status
- [ ] Show deployed services
- [ ] Show model list
- [ ] Make the portal reachable at /cloudinferops

### Medium-term goals

- [ ] Add deployment actions
- [ ] Add logs and alert views
- [ ] Add benchmark reports
- [ ] Polish the UI experience

---

## 18. Final Recommendation

The best approach is to build the UI as a lightweight, self-hosted control plane that is simple, fast, and useful from day one.

Do not try to build a full enterprise portal immediately.

Start with:

- one overview dashboard
- one deployments page
- one models page
- one observability page
- one alerts page

Then add actions gradually.

This gives you a strong foundation for both product growth and community adoption.
