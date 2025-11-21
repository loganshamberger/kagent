# Deployment Guide

This guide provides comprehensive instructions for building and deploying the Kagent system to a Kubernetes cluster. The recommended environment for development and testing is a local Kind (Kubernetes in Docker) cluster.

## Prerequisites

Before you begin, ensure you have the following tools installed:

- **Kind** (v0.27.0+)
- **kubectl** (v1.33.4+)
- **Helm** (v3.x)
- **Go** (v1.24.6+)
- **Docker** (Daemon running)
- **Docker Buildx** (v0.23.0+)
- **Make**

Verify your installation:
```bash
kind version
kubectl version
helm version
go version
docker version
docker buildx version
make --version
```

## Quick Start

The easiest way to get a full system running is to use the provided Makefile targets.

### 1. Create a Kind Cluster

This command sets up a local Kubernetes cluster using Kind and configures MetalLB for load balancing.

```bash
make create-kind-cluster
```

### 2. Configure Model Provider

Set the environment variable for your preferred AI model provider. Supported providers include `openAI`, `anthropic`, `azureOpenAI`, `gemini`, and `ollama`.

```bash
export KAGENT_DEFAULT_MODEL_PROVIDER=openAI
# Set the corresponding API key
export OPENAI_API_KEY=your-openai-api-key
```

For other providers:
```bash
# Anthropic
export KAGENT_DEFAULT_MODEL_PROVIDER=anthropic
export ANTHROPIC_API_KEY=your-key

# Gemini
export KAGENT_DEFAULT_MODEL_PROVIDER=gemini
export GOOGLE_API_KEY=your-key
```

### 3. Build and Deploy

This command builds all Docker images (Controller, UI, App), loads them into the Kind cluster, and deploys the system using Helm.

```bash
make helm-install
```

This process may take a few minutes as it involves:
1. Building Go and Python binaries.
2. creating Docker images.
3. Loading images into Kind.
4. Installing Helm charts.

## Accessing the System

### Kagent UI

To access the web interface, you need to port-forward the UI service:

```bash
kubectl port-forward svc/kagent-ui 8001:80 -n kagent
```

Then open your browser and navigate to [http://localhost:8001](http://localhost:8001).

### CLI Dashboard

You can also use the CLI tool to interact with the system. Build and run it locally:

```bash
make kagent-cli-install
```

## Troubleshooting

### Buildx Connection Issues

If the build fails with an error like `failed to solve: DeadlineExceeded: failed to push localhost:5001/...`, it usually means the build container cannot access the local registry.

Fix this by recreating the buildx builder with host networking:

```bash
make buildx-create
```

Or manually:
```bash
docker buildx rm kagent-builder-v0.23.0
docker buildx create --name kagent-builder-v0.23.0 --platform linux/amd64,linux/arm64 --driver docker-container --use --driver-opt network=host
```

### Clean Up

To delete the cluster and clean up resources:

```bash
make delete-kind-cluster
```

To remove dangling images and build artifacts:

```bash
make clean
```
