#!/bin/bash
set -e

VM_NAME="stackpulse-ec2"
BINARY="stackpulse"
REMOTE_PATH="/home/ubuntu/${BINARY}"
NS="observability"

echo "=== [1/5] Cross-compiling for Linux arm64 ==="
GOOS=linux GOARCH=arm64 go build -o ${BINARY} cmd/stackpulse/main.go

echo "=== [2/5] Transferring binary to VM ==="
multipass exec ${VM_NAME} -- sudo rm -f ${REMOTE_PATH}
multipass transfer ${BINARY} ${VM_NAME}:${REMOTE_PATH}
multipass exec ${VM_NAME} -- chmod +x ${REMOTE_PATH}

echo "=== [3/5] Cleaning Kubernetes cluster inside VM ==="
# Uninstall all StackPulse helm releases (ignore errors if they don't exist)
multipass exec ${VM_NAME} -- sudo bash -c "
  export KUBECONFIG=/etc/rancher/k3s/k3s.yaml 2>/dev/null || true

  echo '  Removing Helm releases in namespace ${NS}...'
  helm uninstall stackpulse-ingress-nginx -n ${NS} 2>/dev/null && echo '    ✓ stackpulse-ingress-nginx removed' || echo '    - stackpulse-ingress-nginx not found (skip)'
  helm uninstall stackpulse-prometheus    -n ${NS} 2>/dev/null && echo '    ✓ stackpulse-prometheus removed'    || echo '    - stackpulse-prometheus not found (skip)'
  helm uninstall stackpulse-loki          -n ${NS} 2>/dev/null && echo '    ✓ stackpulse-loki removed'          || echo '    - stackpulse-loki not found (skip)'
  helm uninstall stackpulse-tempo         -n ${NS} 2>/dev/null && echo '    ✓ stackpulse-tempo removed'         || echo '    - stackpulse-tempo not found (skip)'
  helm uninstall stackpulse-otel          -n ${NS} 2>/dev/null && echo '    ✓ stackpulse-otel removed'          || echo '    - stackpulse-otel not found (skip)'

  echo '  Deleting namespace ${NS}...'
  kubectl delete namespace ${NS} --timeout=60s 2>/dev/null && echo '    ✓ namespace deleted' || echo '    - namespace not found (skip)'

  echo '  Deleting leftover CRDs from kube-prometheus-stack...'
  kubectl delete crd alertmanagerconfigs.monitoring.coreos.com 2>/dev/null || true
  kubectl delete crd alertmanagers.monitoring.coreos.com 2>/dev/null || true
  kubectl delete crd podmonitors.monitoring.coreos.com 2>/dev/null || true
  kubectl delete crd probes.monitoring.coreos.com 2>/dev/null || true
  kubectl delete crd prometheusagents.monitoring.coreos.com 2>/dev/null || true
  kubectl delete crd prometheuses.monitoring.coreos.com 2>/dev/null || true
  kubectl delete crd prometheusrules.monitoring.coreos.com 2>/dev/null || true
  kubectl delete crd scrapeconfigs.monitoring.coreos.com 2>/dev/null || true
  kubectl delete crd servicemonitors.monitoring.coreos.com 2>/dev/null || true
  kubectl delete crd thanosrulers.monitoring.coreos.com 2>/dev/null || true
  echo '    ✓ CRDs cleaned'

  echo '  Cluster reset complete.'
" || true

echo "=== [4/5] Stopping k3s service (to test cluster onboarding prompt) ==="
multipass exec ${VM_NAME} -- sudo systemctl stop k3s 2>/dev/null || true
echo "  ✓ k3s stopped"

echo "=== [5/5] Dropping into VM shell ==="
echo ""
echo "  Run inside the VM:"
echo "    sudo ./stackpulse deploy observability"
echo ""
multipass shell ${VM_NAME}
