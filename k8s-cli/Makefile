# Makefile for k8s-cli
.PHONY: build test clean run help fmt vet

BINARY_NAME=k8s-cli
BIN_DIR=bin
GO_FILES=$(shell find . -name "*.go" -type f)

# Default target - help
help:
	@echo "Available commands:"
	@echo "  build     - Build the application"
	@echo "  test      - Run tests"
	@echo "  run       - Run the application"
	@echo "  clean     - Clean built files"
	@echo "  fmt       - Format code"
	@echo "  vet       - Check code"
	@echo "  install   - Install to system"

# Build
build:
	@echo "🔨 Building $(BINARY_NAME)..."
	@mkdir -p $(BIN_DIR)
	@cd k8s-cli && go build -o ../$(BIN_DIR)/$(BINARY_NAME) main.go
	@echo "✅ Build completed: $(BIN_DIR)/$(BINARY_NAME)"

# Tests
test:
	@echo "🧪 Running tests..."
	@cd k8s-cli && go test ./tests/... -v
	@echo "✅ Tests completed"

# Format
fmt:
	@echo "📝 Formatting code..."
	@cd k8s-cli && go fmt ./...
	@echo "✅ Code formatted"

# Check
vet:
	@echo "🔍 Checking code..."
	@cd k8s-cli && go vet ./...
	@echo "✅ Check completed"

# Run
run: build
	@echo "🚀 Running $(BINARY_NAME)..."
	@./$(BIN_DIR)/$(BINARY_NAME) --help

# Clean
clean:
	@echo "🧹 Cleaning..."
	@rm -rf $(BIN_DIR)
	@echo "✅ Cleaned"

# Install to system
install: build
	@echo "📦 Installing to /usr/local/bin..."
	@sudo cp $(BIN_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "✅ Installed: /usr/local/bin/$(BINARY_NAME)"

# Check everything
check: fmt vet test build
	@echo "✅ All checks passed successfully!"

# Build for all platforms
build-all:
	@echo "🔨 Building for all platforms..."
	@mkdir -p $(BIN_DIR)
	@cd k8s-cli && \
	GOOS=linux GOARCH=amd64 go build -o ../$(BIN_DIR)/$(BINARY_NAME)-linux-amd64 main.go && \
	GOOS=linux GOARCH=arm64 go build -o ../$(BIN_DIR)/$(BINARY_NAME)-linux-arm64 main.go && \
	GOOS=darwin GOARCH=amd64 go build -o ../$(BIN_DIR)/$(BINARY_NAME)-darwin-amd64 main.go && \
	GOOS=darwin GOARCH=arm64 go build -o ../$(BIN_DIR)/$(BINARY_NAME)-darwin-arm64 main.go && \
	GOOS=windows GOARCH=amd64 go build -o ../$(BIN_DIR)/$(BINARY_NAME)-windows-amd64.exe main.go
	@echo "✅ Multi-platform build completed"
	@ls -la $(BIN_DIR)/