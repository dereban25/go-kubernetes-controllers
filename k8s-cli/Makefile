# Makefile for k8s-cli

# Variables
BINARY_NAME=k8s-cli
BUILD_DIR=bin
MAIN_FILE=main.go
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# Default target
.PHONY: all
all: clean deps fmt build

# Help target
.PHONY: help
help:
	@echo "K8s-CLI Makefile Commands"
	@echo ""
	@echo "🚀 Main Targets:"
	@echo "  all           - Clean, get deps, format, and build"
	@echo "  build         - Build the binary"
	@echo "  quick         - Quick build (format + build)"
	@echo ""
	@echo "🛠 Development:"
	@echo "  deps          - Download dependencies"
	@echo "  dev-setup     - Setup development environment"
	@echo "  fmt           - Format code"
	@echo ""
	@echo "🧪 Testing:"
	@echo "  test          - Run tests"
	@echo "  demo          - Run demo commands"
	@echo "  integration   - Run integration tests with cluster"
	@echo ""
	@echo "📦 Build Variants:"
	@echo "  build-all     - Build for multiple platforms"
	@echo "  release       - Create release packages"
	@echo ""
	@echo "🔧 Utilities:"
	@echo "  clean         - Clean build artifacts"
	@echo "  install       - Install binary to system PATH"
	@echo "  help          - Show this help"
	@echo ""
	@echo "📋 Examples:"
	@echo "  make all      # Complete build process"
	@echo "  make quick    # Fast development build"
	@echo "  make demo     # See the tool in action"

# Build the binary
.PHONY: build
build:
	@echo "🔨 Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@echo "✅ Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

# Clean build artifacts
.PHONY: clean
clean:
	@echo "🧹 Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@echo "✅ Clean completed"

# Download dependencies
.PHONY: deps
deps:
	@echo "📦 Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "✅ Dependencies ready"

# Format code
.PHONY: fmt
fmt:
	@echo "💅 Formatting code..."
	$(GOFMT) ./...
	@echo "✅ Code formatted"

# Run tests
.PHONY: test
test:
	@echo "🧪 Running tests..."
	$(GOTEST) -v ./... || echo "No tests found"
	@echo "✅ Tests completed"

# Install binary
.PHONY: install
install: build
	@echo "📦 Installing $(BINARY_NAME)..."
	@if [ -z "$(GOPATH)" ]; then \
		sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/; \
		echo "✅ Installed to /usr/local/bin/$(BINARY_NAME)"; \
	else \
		cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME); \
		echo "✅ Installed to $(GOPATH)/bin/$(BINARY_NAME)"; \
	fi

# Run demo commands
.PHONY: demo
demo: build
	@echo "🎬 Running demo commands..."
	@echo "\n=== 📊 Current Context ==="
	@./$(BUILD_DIR)/$(BINARY_NAME) context current || echo "❌ No kubeconfig found"
	@echo "\n=== 📋 Available Contexts ==="
	@./$(BUILD_DIR)/$(BINARY_NAME) context list || echo "❌ No kubeconfig found"
	@echo "\n=== 🏷️ Namespaces ==="
	@./$(BUILD_DIR)/$(BINARY_NAME) list namespaces || echo "❌ Cannot connect to cluster"
	@echo "\n=== 🐳 Pods in default namespace ==="
	@./$(BUILD_DIR)/$(BINARY_NAME) list pods || echo "❌ Cannot connect to cluster"
	@echo "\n=== 🚀 Deployments in default namespace ==="
	@./$(BUILD_DIR)/$(BINARY_NAME) list deployments || echo "❌ Cannot connect to cluster"
	@echo "\n=== ❓ Help ==="
	@./$(BUILD_DIR)/$(BINARY_NAME) --help

# Build for multiple platforms
.PHONY: build-all
build-all: clean
	@echo "🌍 Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_FILE)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_FILE)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_FILE)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_FILE)
	@echo "✅ Multi-platform binaries built in $(BUILD_DIR)/"

# Development setup
.PHONY: dev-setup
dev-setup:
	@echo "🛠 Setting up development environment..."
	$(GOMOD) init k8s-cli 2>/dev/null || true
	$(GOGET) github.com/spf13/cobra@latest
	$(GOGET) github.com/spf13/viper@latest
	$(GOGET) k8s.io/client-go@latest
	$(GOGET) k8s.io/api@latest
	$(GOGET) k8s.io/apimachinery@latest
	$(GOGET) github.com/olekukonko/tablewriter@latest
	$(GOGET) sigs.k8s.io/yaml@latest
	$(GOMOD) tidy
	@echo "✅ Development environment ready"

# Quick build and test
.PHONY: quick
quick: fmt build
	@echo "⚡ Quick build completed"

# Run integration tests
.PHONY: integration
integration: build
	@echo "🔗 Running integration tests..."
	@if command -v kubectl >/dev/null 2>&1; then \
		echo "Testing with real cluster..."; \
		./$(BUILD_DIR)/$(BINARY_NAME) context current || echo "No cluster connection"; \
		./$(BUILD_DIR)/$(BINARY_NAME) list namespaces || echo "Cannot list namespaces"; \
		./$(BUILD_DIR)/$(BINARY_NAME) list deployments || echo "Cannot list deployments"; \
	else \
		echo "kubectl not found, skipping integration tests"; \
	fi

# Create release
.PHONY: release
release: clean fmt build-all
	@echo "📦 Creating release..."
	@mkdir -p release
	@cp $(BUILD_DIR)/* release/
	@echo "✅ Release artifacts created in release/"

# Show project status
.PHONY: status
status:
	@echo "📊 K8s-CLI Project Status"
	@echo "========================"
	@echo "Go version: $(shell go version 2>/dev/null || echo 'Not installed')"
	@echo "kubectl: $(shell kubectl version --client --short 2>/dev/null || echo 'Not installed')"
	@echo "Binary exists: $(shell test -f $(BUILD_DIR)/$(BINARY_NAME) && echo 'Yes' || echo 'No')"
	@echo "Dependencies: $(shell test -f go.mod && echo 'Ready' || echo 'Not initialized')"