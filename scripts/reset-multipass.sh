#!/usr/bin/env bash
set -euo pipefail

echo "🧹 Deleting and purging all Multipass instances..."
multipass delete --all
multipass purge

echo "🧹 Clearing Go cache..."
go clean -cache -modcache -testcache

echo "🔨 Building StackPulse binary for Linux..."
# Cross-compile for Linux using your host's architecture (handles Apple Silicon / Intel seamlessly)
HOST_ARCH=$(go env GOARCH)
GOOS=linux GOARCH=${HOST_ARCH} go build -o stackpulse cmd/stackpulse/main.go

VM_NAME="stackpulse-vm"

echo "🚀 Launching new Multipass VM ($VM_NAME)..."
# Allocating standard resources suitable for running K3s and the Observability stack
multipass launch --name "$VM_NAME" --cpus 2 --memory 4G --disk 20G

echo "📦 Transferring binary to the VM..."
multipass transfer stackpulse "$VM_NAME":/home/ubuntu/stackpulse

echo "🔧 Setting execute permissions..."
multipass exec "$VM_NAME" -- chmod +x /home/ubuntu/stackpulse

echo "✅ Setup complete! Your fresh VM is ready."

multipass shell $VM_NAME

echo "👉 Then run: ./stackpulse doctor"
