# GPT-Home Makefile

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOLINT=golangci-lint

# Binary name
BINARY_NAME=gpt-home
BINARY_UNIX=$(BINARY_NAME)_unix

# Build flags
BUILD_FLAGS=-a -installsuffix cgo
LDFLAGS=-ldflags '-extldflags "-static"'

# Test flags
TEST_FLAGS=-v -race -coverprofile=coverage.out -covermode=atomic

.PHONY: all build clean test coverage deps fmt lint help

all: test build

## Build the binary
build:
	CGO_ENABLED=0 GOOS=linux $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BINARY_NAME) ./cmd/main.go

## Build for multiple platforms
build-all:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BINARY_NAME)-linux-amd64 ./cmd/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BINARY_NAME)-linux-arm64 ./cmd/main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BINARY_NAME)-darwin-amd64 ./cmd/main.go

## Run tests
test:
	$(GOTEST) $(TEST_FLAGS) ./...

## Run tests with verbose output
test-verbose:
	$(GOTEST) -v ./...

## Generate test coverage report
coverage: test
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) verify

## Update dependencies
deps-update:
	$(GOMOD) tidy
	$(GOGET) -u ./...

## Format code
fmt:
	$(GOFMT) -s -w .

## Check code formatting
fmt-check:
	@if [ $$($(GOFMT) -s -l . | wc -l) -gt 0 ]; then \
		echo "The following files are not formatted:"; \
		$(GOFMT) -s -l .; \
		exit 1; \
	fi

## Run linter
lint:
	$(GOLINT) run

## Run linter with fixes
lint-fix:
	$(GOLINT) run --fix

## Run security scan
security:
	gosec ./...

## Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-*
	rm -f coverage.out
	rm -f coverage.html

## Run the application locally
run:
	$(GOCMD) run ./cmd/main.go

## Build Docker image
docker-build:
	docker build -t $(BINARY_NAME):latest .

## Run Docker container
docker-run:
	docker run --rm -p 8080:8080 $(BINARY_NAME):latest

## Deploy to Kubernetes
k8s-deploy:
	./scripts/deploy.sh

## Apply Kubernetes manifests
k8s-apply:
	kubectl apply -f deployments/k3s/

## Delete Kubernetes resources
k8s-delete:
	kubectl delete -f deployments/k3s/

## Show Kubernetes status
k8s-status:
	kubectl get pods,svc,ingress -n gpt-home

## View logs
k8s-logs:
	kubectl logs -f deployment/gpt-home -n gpt-home

## Install development tools
dev-tools:
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GOGET) github.com/securecodewarrior/github-action-gosec@latest

## Pre-commit checks (format, lint, test)
pre-commit: fmt-check lint test

## CI pipeline (what runs in GitHub Actions)
ci: deps fmt-check lint test build

## Show help
help:
	@echo 'Available commands:'
	@echo ''
	@echo 'Build:'
	@echo '  build        Build the binary for Linux'
	@echo '  build-all    Build for multiple platforms'
	@echo '  clean        Clean build artifacts'
	@echo ''
	@echo 'Testing:'
	@echo '  test         Run tests with coverage'
	@echo '  test-verbose Run tests with verbose output'
	@echo '  coverage     Generate HTML coverage report'
	@echo ''
	@echo 'Development:'
	@echo '  deps         Download dependencies'
	@echo '  deps-update  Update dependencies'
	@echo '  fmt          Format code'
	@echo '  fmt-check    Check code formatting'
	@echo '  lint         Run linter'
	@echo '  lint-fix     Run linter with fixes'
	@echo '  security     Run security scan'
	@echo '  run          Run application locally'
	@echo ''
	@echo 'Docker:'
	@echo '  docker-build Build Docker image'
	@echo '  docker-run   Run Docker container'
	@echo ''
	@echo 'Kubernetes:'
	@echo '  k8s-deploy   Deploy to Kubernetes'
	@echo '  k8s-apply    Apply Kubernetes manifests'
	@echo '  k8s-delete   Delete Kubernetes resources'
	@echo '  k8s-status   Show Kubernetes status'
	@echo '  k8s-logs     View application logs'
	@echo ''
	@echo 'Tools:'
	@echo '  dev-tools    Install development tools'
	@echo '  pre-commit   Run pre-commit checks'
	@echo '  ci           Run CI pipeline'
	@echo ''

# Default target
.DEFAULT_GOAL := help