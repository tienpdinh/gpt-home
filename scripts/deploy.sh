#!/bin/bash

set -e

echo "Deploying GPT-Home to K3s cluster..."

# Build Docker image
echo "Building Docker image..."
docker build -t gpt-home:latest .

# Load image into K3s (for local development)
echo "Loading image into K3s..."
k3s ctr images import <(docker save gpt-home:latest)

# Apply Kubernetes manifests
echo "Applying Kubernetes manifests..."

# Create namespace if it doesn't exist
kubectl create namespace gpt-home --dry-run=client -o yaml | kubectl apply -f -

# Apply configurations
kubectl apply -f deployments/k3s/configmap.yaml -n gpt-home
kubectl apply -f deployments/k3s/pvc.yaml -n gpt-home
kubectl apply -f deployments/k3s/deployment.yaml -n gpt-home

# Wait for deployment to be ready
echo "Waiting for deployment to be ready..."
kubectl wait --for=condition=available --timeout=300s deployment/gpt-home -n gpt-home

# Get service information
echo "Deployment complete!"
echo "Service endpoints:"
kubectl get svc -n gpt-home
echo ""
echo "To access GPT-Home:"
echo "1. Add 'gpt-home.local' to your /etc/hosts pointing to your K3s node IP"
echo "2. Visit http://gpt-home.local in your browser"
echo ""
echo "To check logs:"
echo "kubectl logs -f deployment/gpt-home -n gpt-home"