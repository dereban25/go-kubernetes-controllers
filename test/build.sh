#!/bin/bash

# k8s-cli Build Script
# This script sets up the project structure and builds the k8s-cli tool

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Project configuration
PROJECT_NAME="k8s-cli"
GO_VERSION="1.21"
BINARY_NAME="k8s-cli"
BUILD_DIR="bin"

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo -e "${BLUE}===================================================${NC}"
    echo -e "${BLUE} $1${NC}"
    echo -e "${BLUE}===================================================${NC}"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check Go version
check_go_version() {
    if ! command_exists go; then
        print_error "Go is not installed. Please install Go $GO_VERSION or later."
        exit 1
    fi

    local go_version=$(go version | awk '{print $3}' | sed 's/go//')
    local required_version="1.21"

    if ! printf '%s\n%s\n' "$required_version" "$go_version" | sort -V -C; then
        print_warning "Go version $go_version detected. Recommended: $GO_VERSION or later."
    else
        print_status "Go version $go_version is compatible."
    fi
}

# Function to check kubectl
check_kubectl() {
    if command_exists kubectl; then
        print_status "kubectl found: $(kubectl version --client --short 2>/dev/null || echo 'kubectl installed')"
    else
        print_warning "kubectl not found. Install kubectl to test the CLI with a real cluster."
    fi
}

# Function to setup project structure
setup_project_structure() {
    print_status "Setting up project structure..."

    # Create directories
    mkdir -p cmd
    mkdir -p internal/k8s
    mkdir -p internal/utils
    mkdir -p examples
    mkdir -p bin
    mkdir -p tests

    print_status "Project directories created."
}

# Function to initialize Go module
init_go_module() {
    print_status "Initializing Go module..."

    if [ ! -f "go.mod" ]; then
        go mod init $PROJECT_NAME
        print_status "Go module initialized."
    else
        print_status "Go module already exists."
    fi
}

# Function to download dependencies
download_dependencies() {
    print_status "Downloading dependencies..."

    # Core dependencies
    go get github.com/spf13/cobra@latest
    go get github.com/spf13/viper@latest
    go get k8s.io/client-go@latest
    go get k8s.io/api@latest
    go get k8s.io/apimachinery@latest
    go get github.com/olekukonko/tablewriter@latest
    go get sigs.k8s.io/yaml@latest

    # Tidy up
    go mod tidy

    print_status "Dependencies downloaded and organized."
}

# Function to create example YAML files
create_examples() {
    print_status "Creating example YAML files..."

    # Pod example
    cat > examples/pod.yaml << 'EOF'
apiVersion: v1
kind: Pod
metadata:
  name: nginx-pod
  labels:
    app: nginx
spec:
  containers:
  - name: nginx
    image: nginx:1.20
    ports:
    - containerPort: 80
    resources:
      requests:
        memory: "64Mi"
        cpu: "250m"
      limits:
        memory: "128Mi"
        cpu: "500m"
EOF

    # Deployment example
    cat > examples/deployment.yaml << 'EOF'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  labels:
    app: nginx
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.20
        ports:
        - containerPort: 80
        resources:
          requests:
            memory: "64Mi"
            cpu: "250m"
          limits:
            memory: "128Mi"
            cpu: "500m"
EOF

    # Service example
    cat > examples/service.yaml << 'EOF'
apiVersion: v1
kind: Service
metadata:
  name: nginx-service
  labels:
    app: nginx
spec:
  type: ClusterIP
  ports:
  - port: 80
    targetPort: 80
    protocol: TCP
  selector:
    app: nginx
EOF

    print_status "Example YAML files created in examples/ directory."
}

# Function to build the application
build_application() {
    print_status "Building $PROJECT_NAME..."

    # Format code
    go fmt ./...

    # Run tests if any exist
    if ls *_test.go 1> /dev/null 2>&1 || find . -name "*_test.go" -type f | grep -q .; then
        print_status "Running tests..."
        go test -v ./...
    fi

    # Build binary
    mkdir -p $BUILD_DIR
    go build -ldflags "-X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo 'dev')" -o $BUILD_DIR/$BINARY_NAME main.go

    if [ -f "$BUILD_DIR/$BINARY_NAME" ]; then
        print_status "Binary built successfully: $BUILD_DIR/$BINARY_NAME"

        # Make executable
        chmod +x $BUILD_DIR/$BINARY_NAME

        # Show binary info
        ls -la $BUILD_DIR/$BINARY_NAME
    else
        print_error "Failed to build binary."
        exit 1
    fi
}

# Function to run integration tests
run_integration_tests() {
    print_status "Running integration tests with cluster..."

    if command_exists kubectl; then
        # Test cluster connectivity
        if kubectl cluster-info >/dev/null 2>&1; then
            print_status "✓ Cluster connectivity test passed"

            # Test context commands
            if ./$BUILD_DIR/$BINARY_NAME context current >/dev/null 2>&1; then
                print_status "✓ Context command works"
            else
                print_warning "△ Context command failed (kubeconfig issue?)"
            fi

            # Test list commands
            if ./$BUILD_DIR/$BINARY_NAME list namespaces >/dev/null 2>&1; then
                print_status "✓ List namespaces works"
            else
                print_warning "△ List namespaces failed (permissions?)"
            fi
        else
            print_warning "△ No cluster connection available"
        fi
    else
        print_warning "△ kubectl not available for integration tests"
    fi
}

# Function to run basic tests
run_basic_tests() {
    print_status "Running basic functionality tests..."

    # Test help command
    if ./$BUILD_DIR/$BINARY_NAME --help >/dev/null 2>&1; then
        print_status "✓ Help command works"
    else
        print_error "✗ Help command failed"
        return 1
    fi

    # Test version command
    if ./$BUILD_DIR/$BINARY_NAME --version >/dev/null 2>&1; then
        print_status "✓ Version command works"
    else
        print_warning "△ Version command not available (this is okay)"
    fi

    print_status "Basic tests completed."
}

# Function to show usage examples
show_usage_examples() {
    print_header "Usage Examples"

    echo "After building, you can use the following commands:"
    echo ""
    echo "# Show help"
    echo "./$BUILD_DIR/$BINARY_NAME --help"
    echo ""
    echo "# List contexts (requires kubeconfig)"
    echo "./$BUILD_DIR/$BINARY_NAME context list"
    echo ""
    echo "# List pods in default namespace"
    echo "./$BUILD_DIR/$BINARY_NAME list pods"
    echo ""
    echo "# List pods in specific namespace"
    echo "./$BUILD_DIR/$BINARY_NAME list pods -n kube-system"
    echo ""
    echo "# Apply YAML file"
    echo "./$BUILD_DIR/$BINARY_NAME apply file examples/pod.yaml"
    echo ""
    echo "# Delete resources"
    echo "./$BUILD_DIR/$BINARY_NAME delete pod nginx-pod"
    echo "./$BUILD_DIR/$BINARY_NAME delete file examples/pod.yaml"
    echo ""
    echo "# Different output formats"
    echo "./$BUILD_DIR/$BINARY_NAME list pods -o json"
    echo "./$BUILD_DIR/$BINARY_NAME list pods -o yaml"
    echo ""
}

# Function to create a simple test
create_basic_test() {
    print_status "Creating basic test file..."

    cat > tests/cli_test.go << 'EOF'
package tests

import (
    "os/exec"
    "testing"
)

func TestCLIBinary(t *testing.T) {
    // Test if binary exists and runs
    cmd := exec.Command("../bin/k8s-cli", "--help")
    err := cmd.Run()
    if err != nil {
        t.Fatalf("CLI binary failed to run: %v", err)
    }
}

func TestVersionCommand(t *testing.T) {
    cmd := exec.Command("../bin/k8s-cli", "--version")
    err := cmd.Run()
    // Version command might not be implemented, so we just check it doesn't crash
    if err != nil {
        t.Logf("Version command not available: %v", err)
    }
}
EOF

    print_status "Basic test file created."
}

# Main execution
main() {
    print_header "K8s-CLI Build Script"

    print_status "Starting build process for $PROJECT_NAME..."

    # Pre-build checks
    check_go_version
    check_kubectl

    # Setup and build
    setup_project_structure
    init_go_module
    download_dependencies
    create_examples
    build_application
    create_basic_test

    # Post-build tests
    run_basic_tests

    # Show completion message
    print_header "Build Completed Successfully!"

    print_status "Project: $PROJECT_NAME"
    print_status "Binary location: $BUILD_DIR/$BINARY_NAME"
    print_status "Examples: examples/"

    show_usage_examples

    print_status "To install system-wide, run: sudo cp $BUILD_DIR/$BINARY_NAME /usr/local/bin/"
    print_status "Or add the bin directory to your PATH: export PATH=\$PATH:\$(pwd)/$BUILD_DIR"

    echo ""
    print_status "Build script completed successfully! 🎉"
}

# Help function
show_help() {
    echo "K8s-CLI Build Script"
    echo ""
    echo "Usage: $0 [option]"
    echo ""
    echo "Options:"
    echo "  help          Show this help message"
    echo "  clean         Clean build artifacts"
    echo "  deps-only     Only download dependencies"
    echo "  build-only    Only build (skip setup)"
    echo "  quick         Quick build without full setup"
    echo ""
    echo "Default: Run full build process"
}

# Handle command line arguments
case "${1:-}" in
    "help"|"-h"|"--help")
        show_help
        exit 0
        ;;
    "clean")
        print_status "Cleaning build artifacts..."
        rm -rf $BUILD_DIR
        rm -rf vendor
        go clean
        print_status "Clean completed."
        exit 0
        ;;
    "deps-only")
        init_go_module
        download_dependencies
        exit 0
        ;;
    "build-only")
        build_application
        exit 0
        ;;
    "test")
        check_go_version
        build_application
        run_basic_tests
        print_status "Tests completed."
        exit 0
        ;;
    "integration")
        check_go_version
        check_kubectl
        build_application
        run_basic_tests
        run_integration_tests
        print_status "Integration tests completed."
        exit 0
        ;;
    "quick")
        check_go_version
        build_application
        run_basic_tests
        print_status "Quick build completed."
        exit 0
        ;;
    "")
        print_error "Unknown option: $1"
        show_help
        exit 1
        ;;
esac