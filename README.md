# GPT-Home

[![CI](https://github.com/tienpdinh/gpt-home/workflows/CI/badge.svg)](https://github.com/tienpdinh/gpt-home/actions)
[![codecov](https://codecov.io/gh/tienpdinh/gpt-home/graph/badge.svg?token=KxZZmJs1OP)](https://codecov.io/gh/tienpdinh/gpt-home)
[![Go Report Card](https://goreportcard.com/badge/github.com/tienpdinh/gpt-home)](https://goreportcard.com/report/github.com/tienpdinh/gpt-home)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://golang.org/)

A privacy-first conversational AI system for smart home control, designed to run entirely on edge hardware without cloud dependencies.

## 🏠 Overview

GPT-Home is a specialized ChatGPT-like system that provides natural language control for smart home devices through local language model inference. It runs on a lightweight K3s cluster and connects to Ollama on your home PC for AI processing, ensuring privacy while providing powerful conversational capabilities.

## ✨ Features

- **Natural Language Processing**: Accept varied phrasings for device control
- **HomeAssistant Integration**: Direct REST API integration with your existing HA setup
- **Conversational Context**: Remember previous commands and maintain dialogue context
- **Privacy-First**: All processing occurs locally, no data leaves your network
- **Lightweight k3s Deployment**: Minimal resource requirements (256MB RAM)
- **Kubernetes Ready**: Containerized deployment with K3s orchestration

## 🏗️ Architecture

The system consists of four primary components:

1. **Natural Language Processing Engine**: Connects to Ollama for LLM inference (supports any Ollama model)
2. **Device Integration Layer**: HomeAssistant REST API connectivity
3. **Context Management System**: Maintains conversation state and device references
4. **User Interface**: Web-based chat interface

## 🚀 Quick Start

### Prerequisites

- K3s cluster (can run on minimal hardware)
- HomeAssistant instance with REST API access
- PC/Server running Ollama with desired LLM models
- Go 1.21+ (for development)
- Docker (for containerization)

### Installation

1. **Clone the repository**
   ```bash
   git clone git@github.com:tienpdinh/gpt-home.git
   cd gpt-home
   ```

2. **Install and configure Ollama on your PC**
   ```bash
   # Install Ollama
   curl -fsSL https://ollama.ai/install.sh | sh
   
   # Configure for network access
   export OLLAMA_HOST=0.0.0.0:11434
   ollama serve
   
   # Pull a model (in another terminal)
   ollama pull llama3.2
   ```

3. **Configure and deploy to K3s**
   ```bash
   # Update deployments/k3s/configmap.yaml with your details:
   # - Replace YOUR_PC_IP with your PC's IP address
   # - Set ollama-model to your preferred model
   # - Configure HA URL and token
   
   kubectl apply -f deployments/k3s/
   ```

4. **Access the interface**
   - Add `gpt-home.local` to your `/etc/hosts` pointing to your K3s node IP
   - Visit `http://gpt-home.local`

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_PORT` | HTTP server port | `8080` |
| `HA_URL` | HomeAssistant URL | `http://homeassistant.local:8123` |
| `HA_TOKEN` | HomeAssistant long-lived access token | Required |
| `OLLAMA_URL` | Ollama server URL | `http://localhost:11434` |
| `OLLAMA_MODEL` | Ollama model name | `llama3.2` |
| `LLM_MAX_TOKENS` | Maximum response tokens | `512` |
| `LLM_TEMPERATURE` | Model creativity (0.1-1.0) | `0.7` |
| `LLM_TIMEOUT` | Request timeout (seconds) | `30` |
| `LOG_LEVEL` | Logging level | `info` |

### HomeAssistant Setup

1. Create a long-lived access token in HA
2. Ensure REST API is enabled
3. Update the configuration with your HA URL and token

## 📡 API Endpoints

### Chat
- `POST /api/v1/chat` - Send messages to the AI
- `GET /api/v1/conversations/:id` - Get conversation history

### Device Control
- `GET /api/v1/devices` - List all devices
- `GET /api/v1/devices/:id` - Get device details
- `POST /api/v1/devices/:id/action` - Control specific device

### System
- `GET /api/v1/health` - System health check

## 🤖 Supported Commands

**Lighting**
- "Turn on the living room lights"
- "Dim the bedroom lights"
- "Turn off all lights"

**Climate**
- "Set temperature to 22 degrees"
- "What's the current temperature?"

**General**
- "What devices are available?"
- "Show me the status of all devices"

## 🛠️ Development

### Local Development

```bash
# Install dependencies
go mod download

# Run locally
go run cmd/main.go

# Build
go build -o gpt-home cmd/main.go
```

### Testing

```bash
# Run tests
go test ./...

# Test HomeAssistant connection
curl -X GET "http://localhost:8080/api/v1/health"
```

## 🐳 Deployment

### Docker Build

```bash
docker build -t gpt-home:latest .
```

### Kubernetes Deployment

The deployment includes:
- **Deployment**: Main application pods with resource limits
- **Service**: ClusterIP service for internal communication
- **Ingress**: Traefik ingress for external access
- **ConfigMap**: Non-sensitive configuration
- **Secret**: HomeAssistant token
- **PVC**: Persistent storage for data (models run via Ollama on PC)

## 📊 Resource Requirements

### K3s Cluster Requirements (Minimal)
- **RAM**: 512MB+ (GPT-Home uses 256MB)
- **CPU**: 1 core minimum
- **Storage**: 8GB+ for OS and data

### Ollama PC/Server Requirements
- **RAM**: 2-8GB+ (model dependent)
- **CPU/GPU**: Any modern CPU, GPU optional for faster inference
- **Storage**: 2-20GB+ for models

## 🔒 Security

- All processing occurs locally on your network
- No data transmission to external services
- Conversation history stored locally
- HomeAssistant token secured in Kubernetes secrets

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## 📝 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🙏 Acknowledgments

- HomeAssistant project for the excellent smart home platform
- Ollama team for making LLM deployment simple and accessible
- K3s team for lightweight Kubernetes
- Meta, Alibaba, and other teams creating open source language models

## 🧪 Testing

### Running Tests
```bash
# Run all tests with coverage
make test

# Run tests with verbose output
make test-verbose

# Generate coverage report
make coverage
```

### Test Coverage
- **Unit Tests**: Core business logic and utilities
- **Integration Tests**: HomeAssistant client and device management
- **Mocks**: Comprehensive HomeAssistant API simulation
- **CI/CD**: Automated testing on multiple Go versions


## 🔄 Development Workflow

### Prerequisites
```bash
# Install development tools
make dev-tools

# Download dependencies
make deps
```

### Code Quality
```bash
# Format code
make fmt

# Run linter
make lint

# Security scan
make security

# Pre-commit checks
make pre-commit
```

### Continuous Integration
- **GitHub Actions**: Automated testing, linting, and building
- **Multiple Go Versions**: 1.21.x and 1.22.x
- **Cross-Platform Builds**: Linux and macOS (AMD64/ARM64)
- **Security Scanning**: Gosec and dependency vulnerability checks
- **Code Coverage**: Codecov integration with 80% target

## 📞 Support

For issues and questions:
1. Check the existing issues
2. Create a new issue with detailed information
3. Include logs and configuration (without sensitive data)

---

**Note**: This is an academic research project demonstrating edge AI deployment for smart home automation. Performance and capabilities are optimized for educational purposes and resource-constrained environments.