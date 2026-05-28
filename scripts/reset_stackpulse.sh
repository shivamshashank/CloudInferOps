#!/bin/bash
# Reset script to delete all StackPulse-related Kubernetes resources
# Only deletes resources deployed by this repo (namespace: observability)

set -e

NAMESPACE="observability"

# Delete Helm releases in the observability namespace
echo "Deleting all Helm releases in namespace $NAMESPACE..."
helm list -n $NAMESPACE -q | xargs -r -L1 helm uninstall -n $NAMESPACE

# Delete the namespace (removes all resources)
echo "Deleting namespace $NAMESPACE..."
kubectl delete namespace $NAMESPACE --wait

echo "StackPulse resources deleted."
