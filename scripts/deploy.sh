#!/bin/bash

set -e

echo "Deploying GPT-Home to K3s cluster..."

# Build Docker image
echo "Building Docker image..."
docker build -t gpt-home:latest .

# Tag and push to Docker Hub
echo "Tagging and pushing image to Docker Hub..."
docker tag gpt-home:latest tiendinhphuc/gpt-home:latest
docker build --platform linux/arm64 -t tiendinhphuc/gpt-home:arm64 . --push

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
echo "1. Ensure DNS record for gpt-home.tdinternal.com points to your K3s node"
echo "2. Visit http://gpt-home.tdinternal.com in your browser"
echo ""
echo "To check logs:"
echo "kubectl logs -f deployment/gpt-home -n gpt-home"