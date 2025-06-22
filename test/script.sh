#!/bin/bash

# Step 5: Complete DevOps Setup Script
# Creates Makefile, distroless Dockerfile, GitHub Actions workflow, and comprehensive tests

set -e

# Color codes
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_header() {
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}================================${NC}"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# Create advanced Makefile
create_makefile() {
    print_header "Creating Advanced Makefile"

    cat > Makefile << 'EOF'
# FastHTTP Server Makefile

# Variables
BINARY_NAME=fasthttp-server
DOCKER_IMAGE=fasthttp-server
DOCKER_TAG=latest
GO_VERSION=1.21
MAIN_PATH=./main.go
BUILD_DIR=build
COVERAGE_DIR=coverage
LOGS_DIR=logs

# Go build flags
LDFLAGS=-ldflags "-w -s -X main.version=$(shell git describe --tags --always --dirty 2>/dev/null || echo 'dev') -X main.buildTime=$(shell date -u +%Y-%m-%dT%H:%M:%SZ) -X main.gitCommit=$(shell git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"

# Default target
.DEFAULT_GOAL := help

# Build targets
.PHONY: build
build: clean ## Build the application
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build completed: $(BUILD_DIR)/$(BINARY_NAME)"

.PHONY: build-local
build-local: clean ## Build for local development
	@echo "Building $(BINARY_NAME) for local development..."
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Local build completed: $(BUILD_DIR)/$(BINARY_NAME)"

.PHONY: build-all
build-all: clean ## Build for all platforms
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	@GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "Multi-platform build completed"
	@ls -la $(BUILD_DIR)/

# Development targets
.PHONY: dev
dev: build-local ## Run in development mode
	@echo "Starting development server..."
	@./$(BUILD_DIR)/$(BINARY_NAME) server -p 8080 -l debug

.PHONY: run
run: build-local ## Build and run the server
	@echo "Starting server..."
	@./$(BUILD_DIR)/$(BINARY_NAME) server -p 8080 -l info

# Testing targets
.PHONY: test
test: ## Run unit tests
	@echo "Running unit tests..."
	@go test -v -race -timeout=60s ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@mkdir -p $(COVERAGE_DIR)
	@go test -v -race -coverprofile=$(COVERAGE_DIR)/coverage.out -covermode=atomic ./...
	@go tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@go tool cover -func=$(COVERAGE_DIR)/coverage.out
	@echo "Coverage report generated: $(COVERAGE_DIR)/coverage.html"

.PHONY: test-integration
test-integration: build-local ## Run integration tests
	@echo "Running integration tests..."
	@./tests/integration_test.sh

.PHONY: test-all
test-all: test test-coverage test-integration ## Run all tests

.PHONY: benchmark
benchmark: ## Run benchmarks
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem -timeout=10m ./...

# Code quality targets
.PHONY: lint
lint: ## Run linters
	@echo "Running linters..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		go vet ./...; \
	fi

.PHONY: fmt
fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@goimports -w .

.PHONY: mod-tidy
mod-tidy: ## Tidy go modules
	@echo "Tidying go modules..."
	@go mod tidy
	@go mod verify

# Security targets
.PHONY: security-scan
security-scan: ## Run security scan
	@echo "Running security scan..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not installed. Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

# Docker targets
.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)"

.PHONY: docker-build-distroless
docker-build-distroless: ## Build distroless Docker image
	@echo "Building distroless Docker image..."
	@docker build -f Dockerfile.distroless -t $(DOCKER_IMAGE):distroless .
	@echo "Distroless Docker image built: $(DOCKER_IMAGE):distroless"

.PHONY: docker-run
docker-run: docker-build ## Run Docker container
	@echo "Running Docker container..."
	@docker run -d --name $(DOCKER_IMAGE) -p 8080:8080 \
		-v $(PWD)/$(LOGS_DIR):/app/$(LOGS_DIR) \
		$(DOCKER_IMAGE):$(DOCKER_TAG)
	@echo "Container started: $(DOCKER_IMAGE)"

.PHONY: docker-stop
docker-stop: ## Stop and remove Docker container
	@echo "Stopping Docker container..."
	@docker stop $(DOCKER_IMAGE) || true
	@docker rm $(DOCKER_IMAGE) || true

.PHONY: docker-logs
docker-logs: ## Show Docker container logs
	@docker logs -f $(DOCKER_IMAGE)

.PHONY: docker-scan
docker-scan: docker-build ## Scan Docker image for vulnerabilities
	@echo "Scanning Docker image for vulnerabilities..."
	@if command -v trivy >/dev/null 2>&1; then \
		trivy image $(DOCKER_IMAGE):$(DOCKER_TAG); \
	else \
		echo "trivy not installed. Install from: https://aquasecurity.github.io/trivy/"; \
	fi

# Deployment targets
.PHONY: release
release: clean test-all build-all docker-build-distroless ## Create release build
	@echo "Creating release..."
	@mkdir -p release
	@cp $(BUILD_DIR)/* release/
	@tar -czf release/$(BINARY_NAME)-$(shell git describe --tags --always --dirty).tar.gz -C $(BUILD_DIR) .
	@echo "Release created in release/ directory"

# Utility targets
.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -rf $(COVERAGE_DIR)
	@rm -rf release
	@docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) 2>/dev/null || true
	@docker rmi $(DOCKER_IMAGE):distroless 2>/dev/null || true

.PHONY: logs-clean
logs-clean: ## Clean log files
	@echo "Cleaning log files..."
	@rm -f $(LOGS_DIR)/*.log

.PHONY: deps
deps: ## Install development dependencies
	@echo "Installing development dependencies..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

.PHONY: health-check
health-check: ## Check if server is running
	@echo "Checking server health..."
	@curl -f http://localhost:8080/health || echo "Server is not running"

.PHONY: load-test
load-test: ## Run simple load test
	@echo "Running load test..."
	@if command -v ab >/dev/null 2>&1; then \
		ab -n 1000 -c 10 http://localhost:8080/health; \
	else \
		echo "Apache Bench (ab) not installed"; \
		for i in {1..100}; do curl -s http://localhost:8080/health > /dev/null & done; wait; \
	fi

# Help target
.PHONY: help
help: ## Show this help message
	@echo "FastHTTP Server Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "Examples:"
	@echo "  make build              # Build the application"
	@echo "  make dev                # Run in development mode"
	@echo "  make test-all           # Run all tests"
	@echo "  make docker-build       # Build Docker image"
	@echo "  make release            # Create release build"

# Version info
.PHONY: version
version: ## Show version information
	@echo "FastHTTP Server Build Information"
	@echo "================================="
	@echo "Version: $(shell git describe --tags --always --dirty 2>/dev/null || echo 'dev')"
	@echo "Commit:  $(shell git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
	@echo "Date:    $(shell date -u +%Y-%m-%dT%H:%M:%SZ)"
	@echo "Go:      $(shell go version)"
EOF

    print_info "Created advanced Makefile with comprehensive targets"
}

# Create distroless Dockerfile
create_distroless_dockerfile() {
    print_header "Creating Distroless Dockerfile"

    cat > Dockerfile.distroless << 'EOF'
# Multi-stage build for distroless FastHTTP server
# Using distroless for minimal attack surface and smaller image size

# Build stage
FROM golang:1.21-alpine AS builder

# Install ca-certificates and git for dependencies
RUN apk add --no-cache ca-certificates git tzdata

# Set working directory
WORKDIR /app

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo 'dev') -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -a -installsuffix cgo \
    -o fasthttp-server \
    ./main.go

# Verify the binary
RUN file fasthttp-server && ls -la fasthttp-server

# Final stage - Distroless
FROM gcr.io/distroless/static:nonroot

# Copy timezone data and CA certificates from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary from builder stage
COPY --from=builder /app/fasthttp-server /fasthttp-server

# Create logs directory and set ownership
USER 65532:65532

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/fasthttp-server", "health"] || exit 1

# Set environment variables
ENV GIN_MODE=release
ENV LOG_LEVEL=info

# Run the application
ENTRYPOINT ["/fasthttp-server"]
CMD ["server", "-p", "8080", "-l", "info"]
EOF

    # Also create regular Dockerfile for comparison
    cat > Dockerfile << 'EOF'
# Regular Alpine-based Dockerfile for comparison

# Build stage
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache ca-certificates git tzdata

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o fasthttp-server \
    ./main.go

# Final stage - Alpine
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /app/fasthttp-server .

RUN adduser -D -s /bin/sh appuser
USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ./fasthttp-server health || exit 1

CMD ["./fasthttp-server", "server", "-p", "8080", "-l", "info"]
EOF

    print_info "Created Dockerfile.distroless (minimal image) and Dockerfile (Alpine-based)"
}

# Create GitHub Actions workflows
create_github_workflows() {
    print_header "Creating GitHub Actions Workflows"

    mkdir -p .github/workflows

    # Main CI/CD workflow
    cat > .github/workflows/ci-cd.yml << 'EOF'
name: CI/CD Pipeline

on:
  push:
    branches: [ main, develop ]
    tags: [ 'v*' ]
  pull_request:
    branches: [ main ]

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  # Linting and code quality
  lint:
    name: Lint and Code Quality
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'

    - name: Cache Go modules
      uses: actions/cache@v3
      with:
        path: ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
        restore-keys: |
          ${{ runner.os }}-go-

    - name: Install dependencies
      run: |
        go mod download
        go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
        go install golang.org/x/tools/cmd/goimports@latest

    - name: Run goimports
      run: |
        goimports -w .
        git diff --exit-code

    - name: Run golangci-lint
      run: golangci-lint run --timeout=5m

    - name: Run go vet
      run: go vet ./...

  # Unit tests
  test:
    name: Unit Tests
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: ['1.20', '1.21']
    steps:
    - uses: actions/checkout@v4

    - name: Set up Go ${{ matrix.go-version }}
      uses: actions/setup-go@v4
      with:
        go-version: ${{ matrix.go-version }}

    - name: Cache Go modules
      uses: actions/cache@v3
      with:
        path: ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
        restore-keys: |
          ${{ runner.os }}-go-

    - name: Run tests
      run: go test -v -race -timeout=60s ./...

    - name: Run tests with coverage
      run: |
        mkdir -p coverage
        go test -v -race -coverprofile=coverage/coverage.out -covermode=atomic ./...
        go tool cover -func=coverage/coverage.out

    - name: Upload coverage to Codecov
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage/coverage.out
        flags: unittests
        name: codecov-umbrella

  # Integration tests
  integration-test:
    name: Integration Tests
    runs-on: ubuntu-latest
    needs: [lint, test]
    steps:
    - uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'

    - name: Build application
      run: make build-local

    - name: Run integration tests
      run: make test-integration

    - name: Upload integration test results
      uses: actions/upload-artifact@v3
      if: always()
      with:
        name: integration-test-results
        path: tests/results/

  # Security scanning
  security:
    name: Security Scan
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'

    - name: Run Gosec Security Scanner
      uses: securecodewarrior/github-action-gosec@master
      with:
        args: '-fmt sarif -out gosec-results.sarif ./...'

    - name: Upload SARIF file
      uses: github/codeql-action/upload-sarif@v2
      if: always()
      with:
        sarif_file: gosec-results.sarif

  # Build application
  build:
    name: Build Application
    runs-on: ubuntu-latest
    needs: [lint, test]
    strategy:
      matrix:
        goos: [linux, windows, darwin]
        goarch: [amd64, arm64]
        exclude:
          - goos: windows
            goarch: arm64
    steps:
    - uses: actions/checkout@v4
      with:
        fetch-depth: 0 # Needed for git describe

    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'

    - name: Cache Go modules
      uses: actions/cache@v3
      with:
        path: ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
        restore-keys: |
          ${{ runner.os }}-go-

    - name: Build binary
      run: |
        mkdir -p build
        BINARY_NAME=fasthttp-server-${{ matrix.goos }}-${{ matrix.goarch }}
        if [ "${{ matrix.goos }}" = "windows" ]; then
          BINARY_NAME=${BINARY_NAME}.exe
        fi

        CGO_ENABLED=0 GOOS=${{ matrix.goos }} GOARCH=${{ matrix.goarch }} go build \
          -ldflags="-w -s -X main.version=$(git describe --tags --always --dirty) -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ) -X main.gitCommit=$(git rev-parse --short HEAD)" \
          -o build/${BINARY_NAME} \
          ./main.go

    - name: Upload build artifacts
      uses: actions/upload-artifact@v3
      with:
        name: binaries-${{ matrix.goos }}-${{ matrix.goarch }}
        path: build/

  # Docker build and push
  docker:
    name: Docker Build and Push
    runs-on: ubuntu-latest
    needs: [integration-test, security]
    permissions:
      contents: read
      packages: write
    steps:
    - uses: actions/checkout@v4

    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v3

    - name: Log in to Container Registry
      uses: docker/login-action@v3
      with:
        registry: ${{ env.REGISTRY }}
        username: ${{ github.actor }}
        password: ${{ secrets.GITHUB_TOKEN }}

    - name: Extract metadata
      id: meta
      uses: docker/metadata-action@v5
      with:
        images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
        tags: |
          type=ref,event=branch
          type=ref,event=pr
          type=semver,pattern={{version}}
          type=semver,pattern={{major}}.{{minor}}
          type=sha

    - name: Build and push Docker image (Alpine)
      uses: docker/build-push-action@v5
      with:
        context: .
        file: ./Dockerfile
        push: true
        tags: ${{ steps.meta.outputs.tags }}
        labels: ${{ steps.meta.outputs.labels }}
        cache-from: type=gha
        cache-to: type=gha,mode=max

    - name: Build and push Docker image (Distroless)
      uses: docker/build-push-action@v5
      with:
        context: .
        file: ./Dockerfile.distroless
        push: true
        tags: |
          ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:distroless
          ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:distroless-${{ github.sha }}
        labels: ${{ steps.meta.outputs.labels }}
        cache-from: type=gha
        cache-to: type=gha,mode=max

  # Container security scanning
  container-security:
    name: Container Security Scan
    runs-on: ubuntu-latest
    needs: [docker]
    steps:
    - name: Run Trivy vulnerability scanner
      uses: aquasecurity/trivy-action@master
      with:
        image-ref: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:distroless
        format: 'sarif'
        output: 'trivy-results.sarif'

    - name: Upload Trivy scan results
      uses: github/codeql-action/upload-sarif@v2
      if: always()
      with:
        sarif_file: 'trivy-results.sarif'

  # Release
  release:
    name: Create Release
    runs-on: ubuntu-latest
    if: startsWith(github.ref, 'refs/tags/v')
    needs: [build, docker, container-security]
    permissions:
      contents: write
    steps:
    - uses: actions/checkout@v4
      with:
        fetch-depth: 0

    - name: Download all artifacts
      uses: actions/download-artifact@v3

    - name: Create release archives
      run: |
        mkdir -p release
        for dir in binaries-*; do
          if [ -d "$dir" ]; then
            os_arch=$(echo $dir | sed 's/binaries-//')
            tar -czf release/fasthttp-server-${os_arch}.tar.gz -C "$dir" .
          fi
        done

    - name: Generate changelog
      id: changelog
      run: |
        echo "## Changes" > CHANGELOG.md
        git log --oneline --no-merges $(git describe --tags --abbrev=0 HEAD^)..HEAD >> CHANGELOG.md

    - name: Create Release
      uses: softprops/action-gh-release@v1
      with:
        files: release/*
        body_path: CHANGELOG.md
        draft: false
        prerelease: false
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
EOF

    # Dependency update workflow
    cat > .github/workflows/dependencies.yml << 'EOF'
name: Update Dependencies

on:
  schedule:
    - cron: '0 2 * * 1' # Every Monday at 2 AM
  workflow_dispatch:

jobs:
  update-dependencies:
    name: Update Go Dependencies
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
      with:
        token: ${{ secrets.GITHUB_TOKEN }}
 *********************************************************************************************
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'

    - name: Update dependencies
      run: |
        go get -u all
        go mod tidy

    - name: Create Pull Request
      uses: peter-evans/create-pull-request@v5
      with:
        token: ${{ secrets.GITHUB_TOKEN }}
        commit-message: 'chore: update Go dependencies'
        title: 'chore: update Go dependencies'
        body: |
          Automated dependency update

          - Updated all Go dependencies to latest versions
          - Ran go mod tidy
        branch: update-dependencies
        delete-branch: true
EOF

    print_info "Created GitHub Actions workflows for CI/CD and dependency updates"
}

# Create comprehensive test suite
create_test_suite() {
    print_header "Creating Comprehensive Test Suite"

    # Create tests directory structure
    mkdir -p tests/{unit,integration,benchmarks,fixtures}

    # Unit tests for main package
    cat > main_test.go << 'EOF'
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
)

func TestMain(m *testing.M) {
	// Setup test environment
	os.Setenv("LOG_LEVEL", "error") // Reduce log noise during tests

	// Run tests
	code := m.Run()

	// Cleanup
	os.Exit(code)
}

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		expected string
	}{
		{"Debug level", "debug", "debug"},
		{"Info level", "info", "info"},
		{"Warn level", "warn", "warn"},
		{"Error level", "error", "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger(tt.level)
			if logger.level != tt.expected {
				t.Errorf("Expected level %s, got %s", tt.expected, logger.level)
			}
		})
	}
}

func TestGetRequestID(t *testing.T) {
	// Test with valid request ID
	ctx := &fasthttp.RequestCtx{}
	expectedID := uuid.New().String()
	ctx.SetUserValue(requestIDKey, expectedID)

	actualID := getRequestID(ctx)
	if actualID != expectedID {
		t.Errorf("Expected request ID %s, got %s", expectedID, actualID)
	}

	// Test with missing request ID
	ctx2 := &fasthttp.RequestCtx{}
	actualID2 := getRequestID(ctx2)
	if actualID2 != "unknown" {
		t.Errorf("Expected 'unknown', got %s", actualID2)
	}
}

func TestMainHandler(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		method         string
		expectedStatus int
		checkContent   bool
	}{
		{"Root endpoint", "/", "GET", fasthttp.StatusOK, true},
		{"Health endpoint", "/health", "GET", fasthttp.StatusOK, true},
		{"Status endpoint", "/api/v1/status", "GET", fasthttp.StatusOK, true},
		{"Not found", "/nonexistent", "GET", fasthttp.StatusNotFound, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &fasthttp.RequestCtx{}
			ctx.Request.SetRequestURI(tt.path)
			ctx.Request.Header.SetMethod(tt.method)

			// Set request ID for proper logging
			ctx.SetUserValue(requestIDKey, uuid.New().String())

			// Set log level for testing
			logLevel = "error"

			mainHandler(ctx)

			if ctx.Response.StatusCode() != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, ctx.Response.StatusCode())
			}

			if tt.checkContent {
				body := ctx.Response.Body()
				if len(body) == 0 {
					t.Error("Expected response body, got empty")
				}

				// Check if response is valid JSON
				var jsonResp map[string]interface{}
				if err := json.Unmarshal(body, &jsonResp); err != nil {
					t.Errorf("Invalid JSON response: %v", err)
				}

				// Check if request_id is present
				if _, exists := jsonResp["request_id"]; !exists {
					t.Error("Response missing request_id field")
				}
			}
		})
	}
}

func TestLoggingMiddleware(t *testing.T) {
	logger := NewLogger("error")
	called := false

	handler := loggingMiddleware(func(ctx *fasthttp.RequestCtx) {
		called = true
		ctx.SetStatusCode(fasthttp.StatusOK)
	}, logger)

	ctx := &fast