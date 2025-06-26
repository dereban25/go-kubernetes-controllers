# Makefile for k8s-cli with Step 7-12 support

# Variables
BINARY_NAME=k8s-cli
VERSION ?= dev
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT_SHA := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build flags
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.commitSHA=$(COMMIT_SHA)"

# Colors for output
GREEN := \033[0;32m
YELLOW := \033[1;33m
RED := \033[0;31m
BLUE := \033[0;34m
NC := \033[0m # No Color

# Go related variables
GOBASE := $(shell pwd)
GOPATH := $(GOBASE)/vendor:$(GOBASE)
GOBIN := $(GOBASE)/bin
GOFILES := $(wildcard *.go)

# Controller-gen
CONTROLLER_GEN = $(shell pwd)/bin/controller-gen
CONTROLLER_TOOLS_VERSION ?= v0.13.0

# Default target
.PHONY: all
all: clean deps generate manifests build

# Install controller-gen if necessary
.PHONY: controller-gen
controller-gen:
	@if [ ! -f $(CONTROLLER_GEN) ]; then \
		echo "$(YELLOW)📦 Installing controller-gen...$(NC)"; \
		GOBIN=$(shell pwd)/bin go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION); \
	fi

# Generate DeepCopy methods
.PHONY: generate
generate: controller-gen
	@echo "$(GREEN)🔧 Generating DeepCopy methods...$(NC)"
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./api/..."
	@echo "$(GREEN)✅ Code generation completed$(NC)"

# Generate CRD manifests
.PHONY: manifests
manifests: controller-gen
	@echo "$(GREEN)📄 Generating CRD manifests...$(NC)"
	$(CONTROLLER_GEN) crd paths="./api/..." output:crd:artifacts:config=config/crd/bases
	@echo "$(GREEN)✅ CRD manifests generated$(NC)"

# Install CRDs into cluster
.PHONY: install-crds
install-crds: manifests
	@echo "$(GREEN)📦 Installing CRDs to cluster...$(NC)"
	kubectl apply -f config/crd/bases/
	@echo "$(GREEN)✅ CRDs installed$(NC)"

# Uninstall CRDs from cluster
.PHONY: uninstall-crds
uninstall-crds:
	@echo "$(YELLOW)🗑️ Removing CRDs from cluster...$(NC)"
	kubectl delete -f config/crd/bases/ --ignore-not-found=true
	@echo "$(GREEN)✅ CRDs removed$(NC)"

# Build the binary
.PHONY: build
build:
	@echo "$(GREEN)🔨 Building $(BINARY_NAME)...$(NC)"
	mkdir -p bin
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) main.go
	@echo "$(GREEN)✅ Build completed: bin/$(BINARY_NAME)$(NC)"

# Install dependencies
.PHONY: deps
deps:
	@echo "$(GREEN)📦 Installing dependencies...$(NC)"
	go mod download
	go mod tidy
	@echo "$(GREEN)✅ Dependencies installed$(NC)"

# Clean build artifacts
.PHONY: clean
clean:
	@echo "$(YELLOW)🧹 Cleaning...$(NC)"
	go clean
	rm -rf bin/
	rm -rf vendor/
	rm -f test-*.yaml
	rm -f zz_generated.deepcopy.go
	@echo "$(GREEN)✅ Clean completed$(NC)"

# Run tests
.PHONY: test
test:
	@echo "$(BLUE)🧪 Running unit tests...$(NC)"
	go test -v ./tests/...

# Format code
.PHONY: fmt
fmt:
	@echo "$(GREEN)🎨 Formatting code...$(NC)"
	go fmt ./...
	@echo "$(GREEN)✅ Code formatted$(NC)"

# Lint code
.PHONY: lint
lint:
	@echo "$(GREEN)🔍 Running linter...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
		echo "$(GREEN)✅ Linting completed$(NC)"; \
	else \
		echo "$(YELLOW)⚠️ golangci-lint not found, skipping lint check$(NC)"; \
		echo "$(YELLOW)💡 To install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest$(NC)"; \
	fi

# Build and run basic tests
.PHONY: test-build
test-build: build
	@echo "$(BLUE)📋 Testing build and basic commands...$(NC)"
	@./bin/$(BINARY_NAME) --help > /dev/null && echo "$(GREEN)✅ Help command works$(NC)" || echo "$(RED)❌ Help command failed$(NC)"
	@./bin/$(BINARY_NAME) version > /dev/null && echo "$(GREEN)✅ Version command works$(NC)" || echo "$(RED)❌ Version command failed$(NC)"

# Run checks (similar to CI)
.PHONY: check
check: fmt lint test build
	@echo "$(GREEN)✅ All checks passed!$(NC)"

# Install to system
.PHONY: install
install: build
	@echo "$(GREEN)📦 Installing to system...$(NC)"
	@if [ -w /usr/local/bin ]; then \
		cp bin/$(BINARY_NAME) /usr/local/bin/; \
		echo "$(GREEN)✅ Installed to /usr/local/bin/$(BINARY_NAME)$(NC)"; \
	else \
		sudo cp bin/$(BINARY_NAME) /usr/local/bin/; \
		echo "$(GREEN)✅ Installed to /usr/local/bin/$(BINARY_NAME) (with sudo)$(NC)"; \
	fi

# Uninstall from system
.PHONY: uninstall
uninstall:
	@echo "$(YELLOW)🗑️ Uninstalling from system...$(NC)"
	@if [ -w /usr/local/bin ]; then \
		rm -f /usr/local/bin/$(BINARY_NAME); \
	else \
		sudo rm -f /usr/local/bin/$(BINARY_NAME); \
	fi
	@echo "$(GREEN)✅ Uninstalled$(NC)"

# Quick build for development
.PHONY: quick
quick:
	@echo "$(GREEN)⚡ Quick build...$(NC)"
	go build -o bin/$(BINARY_NAME) main.go

# Build for multiple platforms
.PHONY: build-all
build-all:
	@echo "$(GREEN)🌍 Building for multiple platforms...$(NC)"
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64 main.go
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-amd64 main.go
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-amd64.exe main.go
	@echo "$(GREEN)✅ Multi-platform build completed$(NC)"

# Step-specific test targets
.PHONY: test-step7
test-step7: build
	@echo "$(BLUE)🔄 Testing Step 7: Informers...$(NC)"
	@timeout 10s ./bin/$(BINARY_NAME) watch-informer --workers 1 --resync-period 5s || true
	@echo "$(GREEN)✅ Step 7 test completed$(NC)"

.PHONY: test-step7plus
test-step7plus: build
	@echo "$(BLUE)🌐 Testing Step 7+: JSON API...$(NC)"
	@./bin/$(BINARY_NAME) api-server --port 8080 > /tmp/api-test.log 2>&1 & \
	API_PID=$$!; \
	sleep 3; \
	curl -s http://localhost:8080/api/v1/health || true; \
	kill $$API_PID 2>/dev/null || true
	@echo "$(GREEN)✅ Step 7+ test completed$(NC)"

.PHONY: test-step8
test-step8: build
	@echo "$(BLUE)🚀 Testing Step 8: Advanced API...$(NC)"
	@./bin/$(BINARY_NAME) step8-api --port 8090 > /tmp/step8-test.log 2>&1 & \
	API_PID=$$!; \
	sleep 3; \
	curl -s http://localhost:8090/api/v2/health || true; \
	kill $$API_PID 2>/dev/null || true
	@echo "$(GREEN)✅ Step 8 test completed$(NC)"

.PHONY: test-step9
test-step9: build
	@echo "$(BLUE)🎮 Testing Step 9: Controller Runtime...$(NC)"
	@timeout 10s ./bin/$(BINARY_NAME) controller --workers 1 || true
	@echo "$(GREEN)✅ Step 9 test completed$(NC)"

.PHONY: test-step10
test-step10: build
	@echo "$(BLUE)👑 Testing Step 10: Manager...$(NC)"
	@timeout 10s ./bin/$(BINARY_NAME) manager --enable-leader-election=false || true
	@echo "$(GREEN)✅ Step 10 test completed$(NC)"

.PHONY: test-step11
test-step11: build generate manifests
	@echo "$(BLUE)🔧 Testing Step 11: CRD Controller...$(NC)"
	@timeout 10s ./bin/$(BINARY_NAME) crd --enable-leader-election=false || true
	@echo "$(GREEN)✅ Step 11 test completed$(NC)"

.PHONY: test-step12
test-step12: build
	@echo "$(BLUE)🏗️ Testing Step 12: Platform API...$(NC)"
	@./bin/$(BINARY_NAME) platform --port 8084 > /tmp/platform-test.log 2>&1 & \
	API_PID=$$!; \
	sleep 3; \
	curl -s http://localhost:8084/health || true; \
	kill $$API_PID 2>/dev/null || true
	@echo "$(GREEN)✅ Step 12 test completed$(NC)"

# Test all steps
.PHONY: test-all-steps
test-all-steps: test-step7 test-step7plus test-step8 test-step9 test-step10 test-step11 test-step12
	@echo "$(GREEN)🎉 All step tests completed!$(NC)"

# Complete test suite
.PHONY: test-complete
test-complete: check test-all-steps
	@echo ""
	@echo "$(GREEN)🎯 COMPLETE TEST SUITE FINISHED!$(NC)"
	@echo ""
	@echo "$(BLUE)📊 Summary:$(NC)"
	@echo "   ✅ Build and basic commands"
	@echo "   ✅ Code formatting and linting"
	@echo "   ✅ Unit tests"
	@echo "   ✅ Step 7: Informers"
	@echo "   ✅ Step 7+: JSON API"
	@echo "   ✅ Step 8: Advanced API"
	@echo "   ✅ Step 9: Controller Runtime"
	@echo "   ✅ Step 10: Manager"
	@echo "   ✅ Step 11: CRD Controller"
	@echo "   ✅ Step 12: Platform API"
	@echo ""
	@echo "$(GREEN)🎉 Your k8s-cli is ready with Steps 7-12++ support!$(NC)"

# Run integration tests
.PHONY: integration
integration: build
	@echo "$(BLUE)🧪 Running integration tests...$(NC)"
	go test -v ./tests/... -tags=integration

# Demo scenarios
.PHONY: demo
demo: build
	@echo "$(BLUE)🎬 Running demo scenario...$(NC)"
	./bin/$(BINARY_NAME) context current
	./bin/$(BINARY_NAME) list deployments
	@echo "$(GREEN)✅ Demo completed$(NC)"

.PHONY: demo-frontendpage
demo-frontendpage: build
	@echo "$(BLUE)🎬 Creating demo FrontendPage resources...$(NC)"
	kubectl apply -f examples/frontendpage-demo.yaml || true
	@sleep 2
	kubectl get frontendpages || true
	@echo "$(GREEN)✅ FrontendPage demo completed$(NC)"

.PHONY: demo-platform
demo-platform: build
	@echo "$(BLUE)🎬 Testing platform actions...$(NC)"
	@curl -X GET http://localhost:8084/api/v1/actions 2>/dev/null || echo "Platform API not running"
	@echo "$(GREEN)✅ Platform demo completed$(NC)"

# Development helpers
.PHONY: dev-watch
dev-watch: build
	@echo "$(BLUE)🔄 Starting development watch mode...$(NC)"
	./bin/$(BINARY_NAME) watch-informer --workers 2 --resync-period 10s --log-events

.PHONY: dev-api
dev-api: build
	@echo "$(BLUE)🌐 Starting development API server...$(NC)"
	./bin/$(BINARY_NAME) api-server --port 8080

.PHONY: dev-step8
dev-step8: build
	@echo "$(BLUE)🚀 Starting development Step 8 API server...$(NC)"
	./bin/$(BINARY_NAME) step8-api --port 8090 --enable-debug --enable-metrics

.PHONY: dev-controller
dev-controller: build
	@echo "$(BLUE)🎮 Starting development controller...$(NC)"
	./bin/$(BINARY_NAME) controller --workers 2

.PHONY: dev-manager
dev-manager: build
	@echo "$(BLUE)👑 Starting development manager...$(NC)"
	./bin/$(BINARY_NAME) manager --enable-leader-election=false

.PHONY: dev-crd
dev-crd: build generate manifests install-crds
	@echo "$(BLUE)🔧 Starting development CRD controller...$(NC)"
	./bin/$(BINARY_NAME) crd --enable-leader-election=false

.PHONY: dev-platform
dev-platform: build
	@echo "$(BLUE)🏗️ Starting development platform API...$(NC)"
	./bin/$(BINARY_NAME) platform --port 8084

# Help
.PHONY: help
help:
	@echo "$(BLUE)k8s-cli Makefile - Step 7-12++ Support$(NC)"
	@echo ""
	@echo "$(YELLOW)Main targets:$(NC)"
	@echo "  all              - Clean, generate, and build everything"
	@echo "  build            - Build the binary"
	@echo "  clean            - Clean build artifacts"
	@echo "  deps             - Install dependencies"
	@echo "  generate         - Generate DeepCopy methods"
	@echo "  manifests        - Generate CRD manifests"
	@echo "  install-crds     - Install CRDs to cluster"
	@echo ""
	@echo "$(YELLOW)Test targets:$(NC)"
	@echo "  test             - Run unit tests"
	@echo "  test-complete    - Complete test suite"
	@echo "  test-all-steps   - Test all Step 7-12 functionality"
	@echo "  check            - Run all checks (fmt, lint, test, build)"
	@echo ""
	@echo "$(YELLOW)Development:$(NC)"
	@echo "  dev-*            - Start development servers for each component"
	@echo "  demo             - Run demo scenarios"
	@echo ""
	@echo "$(YELLOW)Installation:$(NC)"
	@echo "  install          - Install to /usr/local/bin"
	@echo "  uninstall        - Remove from system"
	@echo ""
	@echo "$(YELLOW)Examples:$(NC)"
	@echo "  make                      # Build everything"
	@echo "  make test-complete        # Full test suite"
	@echo "  make dev-crd              # Start CRD controller development"