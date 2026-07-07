# CloudInferOps CLI Command Reference

This document provides a comprehensive reference of all available commands in the CloudInferOps CLI.

---

## General Commands

### `version`
Prints the version information of the installed CloudInferOps CLI.
```bash
cloudinferops version
```

### `init`
Initializes the default configuration file for CloudInferOps.
```bash
sudo cloudinferops init
```

### `doctor`
Performs environment pre-flight diagnostics (OS/arch checks, internet connectivity, kubectl/Helm availability, and hardware resource checking).
```bash
sudo cloudinferops doctor
```

### `status`
Displays a unified status dashboard indicating the health and state of all system components, running inference models, GitOps overview, and access URLs for web interfaces.
```bash
sudo cloudinferops status
```

---

## Deployment Commands

### `deploy all`
Deploys all CloudInferOps components sequentially in a single step (observability stack, inference stack, self-hosted UI, and webhook handler).
```bash
sudo cloudinferops deploy all [flags]
```
**Flags:**
- `--dry-run`: Show planned deployment without committing.
- `--provider string`: Model provider to deploy (default: `"ollama"`).
- `--model string`: Model name to deploy (default: `"llama3"`).
- `--image string`: UI portal container image (default: `"ghcr.io/shivamshashank/cloudinferops-ui:latest"`).
- `--enable-actions`: Enable guarded cluster write actions in the portal (default: `false`).

### `deploy observability`
Deploys the core observability platform (Prometheus, Grafana, Loki, Tempo, ArgoCD, Alertmanager, OpenTelemetry Collector, and Node Exporter).
```bash
sudo cloudinferops deploy observability [flags]
```
**Flags:**
- `--dry-run`: Show what would be deployed without committing.
- `--ha`: Enable high-availability storage architecture via Thanos.

### `deploy inference`
Deploys an AI/ML model inference service and the gateway interface.
```bash
sudo cloudinferops deploy inference [flags]
```
**Flags:**
- `--provider string`: Inference provider, e.g. `ollama` or `vllm` (default: `"ollama"`).
- `--model string`: Model name to download and deploy (default: `"llama3"`).
- `--dry-run`: Output planned manifests without executing.

### `deploy ui`
Deploys the self-hosted CloudInferOps web portal dashboard.
```bash
sudo cloudinferops deploy ui [flags]
```
**Flags:**
- `--image string`: Custom portal container image.
- `--enable-actions`: Enable allowlisted writes and configuration updates.
- `--dry-run`: Output deployment manifests without deploying.

### `deploy webhook-handler`
Deploys the custom Go webhook gateway for incident routing.
```bash
sudo cloudinferops deploy webhook-handler [flags]
```
**Flags:**
- `--dry-run`: Show planned deployment.

---

## Diagnostics & Operations

### `dashboards import`
Imports preconfigured SRE Grafana dashboards into the running Grafana instance.
```bash
sudo cloudinferops dashboards import
```

### `logs`
Retrieves logs for deployed system components.
```bash
sudo cloudinferops logs --component <component-name>
```
**Example:**
```bash
sudo cloudinferops logs --component grafana
sudo cloudinferops logs --component prometheus
```

### `models list`
Lists all downloaded models inside the inference service container.
```bash
cloudinferops models list
```

### `benchmark run`
Runs throughput and latency benchmarking tests on a deployed inference model.
```bash
cloudinferops benchmark run --model <model-name>
```

---

## GitOps & Delivery

### `gitops bootstrap`
Bootstraps GitOps workflows by provisioning an internal Git server and linking ArgoCD applications to track git repositories.
```bash
sudo cloudinferops gitops bootstrap [flags]
```

### `gitops status`
Displays the synchronization status of GitOps-managed applications.
```bash
sudo cloudinferops gitops status
```

---

## Alerts & Integrations

### `alerts configure`
Configures notification integrations for alerting.
```bash
sudo cloudinferops alerts configure --slack
sudo cloudinferops alerts configure --pagerduty
```

### `alerts test`
Fires a dummy test alert through Alertmanager to verify notifications.
```bash
sudo cloudinferops alerts test
```

---

## Uninstallation & Cleanup

### `uninstall`
If run without subcommands, it displays an interactive menu with options to uninstall specific parts of the platform or everything.
```bash
sudo cloudinferops uninstall [flags]
```

### `uninstall all`
Runs the complete interactive or automated teardown of all observability, inference, UI, webhook, Kubernetes cluster, and configuration components.
```bash
sudo cloudinferops uninstall all [flags]
```

### `uninstall observability`
Removes the core observability stack namespace and its resources.
```bash
sudo cloudinferops uninstall observability [flags]
```

### `uninstall inference`
Cleans up and deletes the inference stack namespace.
```bash
sudo cloudinferops uninstall inference [flags]
```

### `uninstall ui`
Removes the UI portal resources from the observability namespace.
```bash
sudo cloudinferops uninstall ui [flags]
```

### `uninstall k8s`
Uninstalls the local Kubernetes cluster and purges system packages (kubeadm, containerd, etc.).
```bash
sudo cloudinferops uninstall k8s [flags]
```

**Common Uninstall Flags:**
- `-f, --force`: Bypass confirmation prompts.
- `--dry-run`: Print commands without deleting resources.
