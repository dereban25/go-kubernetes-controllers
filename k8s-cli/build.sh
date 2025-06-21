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
    print_status "Go version $go_version is compatible."
}

# Function to check kubectl
check_kubectl() {
    if command_exists kubectl; then
        print_status "kubectl found"
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
    go get sigs.k8s.io/yaml@latest

    # Tidy up
    go mod tidy

    print_status "Dependencies downloaded and organized."
}

# Function to create example YAML files
create_examples() {
    print_status "Creating example YAML files..."

    # Only create if examples directory exists and files don't exist
    if [ -d "examples" ]; then
        if [ ! -f "examples/pod.yaml" ]; then
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
        fi

        if [ ! -f "examples/deployment.yaml" ]; then
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
        fi

        if [ ! -f "examples/service.yaml" ]; then
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
        fi
    fi

    print_status "Example YAML files ready."
}

# Function to build the application
build_application() {
    print_status "Building $PROJECT_NAME..."

    # Format code
    go fmt ./...

    # Build binary
    mkdir -p $BUILD_DIR
    go build -o $BUILD_DIR/$BINARY_NAME main.go

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

# Function to install the binary
install_binary() {
    print_status "Installing $BINARY_NAME to system..."

    if [ ! -f "$BUILD_DIR/$BINARY_NAME" ]; then
        print_error "Binary not found. Run build first."
        return 1
    fi

    # Try to install to /usr/local/bin first
    if sudo cp "$BUILD_DIR/$BINARY_NAME" /usr/local/bin/ 2>/dev/null; then
        print_status "✅ Installed to /usr/local/bin/$BINARY_NAME"
        print_status "✅ You can now run: k8s-cli --help"
        return 0
    fi

    # If sudo failed, try user's bin directory
    if [ ! -z "$HOME" ]; then
        mkdir -p "$HOME/bin"
        if cp "$BUILD_DIR/$BINARY_NAME" "$HOME/bin/"; then
            print_status "✅ Installed to $HOME/bin/$BINARY_NAME"
            print_warning "⚠️  Add $HOME/bin to your PATH if not already added:"
            print_warning "   echo 'export PATH=\$PATH:\$HOME/bin' >> ~/.bashrc"
            print_warning "   source ~/.bashrc"
            return 0
        fi
    fi

    # If both failed, suggest manual installation
    print_warning "⚠️  Could not install automatically. Manual installation:"
    print_warning "   sudo cp $BUILD_DIR/$BINARY_NAME /usr/local/bin/"
    print_warning "   Or add current directory to PATH:"
    print_warning "   export PATH=\$PATH:\$(pwd)/$BUILD_DIR"
    return 1
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

    print_status "Basic tests completed."
}

# Function to show usage examples
show_usage_examples() {
    print_header "Usage Examples"

    # Check if binary is in PATH
    if command_exists k8s-cli; then
        echo "🎉 k8s-cli is installed! You can use:"
        echo ""
        echo "# Show help"
        echo "k8s-cli --help"
        echo ""
        echo "# List contexts (requires kubeconfig)"
        echo "k8s-cli context list"
        echo ""
        echo "# List deployments (Step 6)"
        echo "k8s-cli list deployments"
        echo ""
        echo "# Create deployment imperatively"
        echo "k8s-cli create deployment demo --image=nginx:1.20"
        echo ""
        echo "# Apply YAML file"
        echo "k8s-cli apply file examples/pod.yaml"
        echo ""
        echo "# Delete resources"
        echo "k8s-cli delete pod nginx-pod"
        echo ""
    else
        echo "After installation, you can use the following commands:"
        echo ""
        echo "# Local usage (current directory)"
        echo "./$BUILD_DIR/$BINARY_NAME --help"
        echo "./$BUILD_DIR/$BINARY_NAME context list"
        echo "./$BUILD_DIR/$BINARY_NAME list deployments"
        echo ""
        echo "# Or install and use globally:"
        echo "./$0 install"
        echo "k8s-cli --help"
        echo ""
    fi
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

    # Post-build tests
    run_basic_tests

    # Show completion message
    print_header "Build Completed Successfully!"

    print_status "🎯 Project: $PROJECT_NAME"
    print_status "📦 Binary location: $BUILD_DIR/$BINARY_NAME"
    print_status "📁 Examples: examples/"

    show_usage_examples

    print_status "🔧 Installation Options:"
    print_status "  ./$0 install              # Install to system"
    print_status "  sudo cp $BUILD_DIR/$BINARY_NAME /usr/local/bin/"
    print_status "  export PATH=\$PATH:\$(pwd)/$BUILD_DIR"

    echo ""
    print_status "Build script completed successfully! 🎉"

    # Ask if user wants to install
    echo ""
    echo -n "Would you like to install k8s-cli to system now? (y/N): "
    read -r response
    if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
        install_binary
    fi
}

# Help function
show_help() {
    echo "K8s-CLI Build Script"
    echo ""
    echo "Usage: $0 [option]"
    echo ""
    echo "Build Options:"
    echo "  (no args)     Full build process with all features"
    echo "  quick         Quick build without full setup"
    echo "  build-only    Only build (skip setup)"
    echo "  deps-only     Only download dependencies"
    echo ""
    echo "Installation Options:"
    echo "  install       Install binary to system PATH"
    echo "  uninstall     Remove binary from system"
    echo ""
    echo "Utility Options:"
    echo "  clean         Clean build artifacts"
    echo "  help          Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0            # Full build with all features"
    echo "  $0 install    # Install to system"
    echo "  $0 quick      # Quick build for development"
    echo "  $0 clean      # Clean all build artifacts"
}

# Install function
install_only() {
    print_header "Installing k8s-cli"

    if [ ! -f "$BUILD_DIR/$BINARY_NAME" ]; then
        print_error "Binary not found. Building first..."
        build_application
    fi

    install_binary
}

# Uninstall function
uninstall_binary() {
    print_status "Uninstalling k8s-cli..."

    # Remove from system locations
    if [ -f "/usr/local/bin/$BINARY_NAME" ]; then
        if sudo rm "/usr/local/bin/$BINARY_NAME" 2>/dev/null; then
            print_status "✅ Removed from /usr/local/bin/"
        fi
    fi

    if [ -f "$HOME/bin/$BINARY_NAME" ]; then
        if rm "$HOME/bin/$BINARY_NAME" 2>/dev/null; then
            print_status "✅ Removed from $HOME/bin/"
        fi
    fi

    if command_exists k8s-cli; then
        print_warning "⚠️  k8s-cli still found in PATH. Check other locations."
    else
        print_status "✅ k8s-cli successfully uninstalled"
    fi
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
    "quick")
        check_go_version
        build_application
        run_basic_tests
        print_status "Quick build completed."
        exit 0
        ;;
    "install")
        install_only
        exit 0
        ;;
    "uninstall")
        uninstall_binary
        exit 0
        ;;
    "")
        main
        ;;
    *)
        print_error "Unknown option: $1"
        show_help
        exit 1
        ;;
esac