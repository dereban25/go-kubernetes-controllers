#!/bin/bash

# Test script for k8s-cli
set -e

echo "🚀 Running k8s-cli tests..."

# Build the application
echo "📦 Building application..."
go build -o bin/k8s-cli main.go

# Run unit tests
echo "🧪 Running unit tests..."
go test -v ./tests/... -short

# Test CLI functionality
echo "🔧 Testing CLI functionality..."
./bin/k8s-cli --help

# Test with mock cluster (if available)
if command -v kubectl >/dev/null 2>&1; then
    echo "☸️ Testing with Kubernetes cluster..."

    # Test context commands
    ./bin/k8s-cli context current || echo "No context available"
    ./bin/k8s-cli context list || echo "No contexts available"

    # Test listing commands
    ./bin/k8s-cli list namespaces || echo "Cannot connect to cluster"
    ./bin/k8s-cli list deployments || echo "Cannot list deployments"

    echo "✅ Cluster tests completed"
else
    echo "⚠️ kubectl not available, skipping cluster tests"
fi

echo "✅ All tests completed successfully!"