# k8s-cli - Complete Kubernetes CLI Tool

A powerful, feature-rich command-line tool for managing Kubernetes clusters with both imperative and declarative approaches. Built with Go using the Cobra CLI framework and the official Kubernetes client-go library.

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![Kubernetes](https://img.shields.io/badge/Kubernetes-Compatible-326CE5?style=flat&logo=kubernetes)](https://kubernetes.io)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## 🎯 Features

### **Core Functionality**
- ✅ **Context Management**: Switch between Kubernetes contexts seamlessly
- ✅ **Resource Viewing**: List pods, deployments, services, and namespaces with beautiful table output
- ✅ **Declarative Management**: Apply and delete resources from YAML files
- ✅ **Imperative Management**: Create resources directly with command-line flags (like kubectl create)
- ✅ **Resource Deletion**: Delete resources by name or from YAML files with confirmation prompts

### **Advanced Features**
- ✅ **Flexible Output**: Support for table, JSON, and YAML output formats
- ✅ **Kubeconfig Support**: Custom kubeconfig file paths with automatic detection
- ✅ **Namespace Support**: Work with specific namespaces or use defaults
- ✅ **Label Selectors**: Filter resources using Kubernetes label selectors
- ✅ **Force Operations**: Skip confirmation prompts for automated workflows
- ✅ **Cross-Platform**: Builds for Linux, macOS, and Windows

### **Step 6 Compliance**
- ✅ **Deployment Listing**: Full support for listing Kubernetes deployment resources
- ✅ **Kubeconfig Authentication**: Proper authentication via kubeconfig files
- ✅ **Kubeconfig Flags**: Complete support for custom kubeconfig paths
- ✅ **Default Namespace**: Lists deployments in default namespace with all output formats

## 🚀 Quick Start

### Prerequisites

- **Go 1.21+**: [Download Go](https://golang.org/dl/)
- **kubectl**: [Install kubectl](https://kubernetes.io/docs/tasks/tools/)
- **Kubernetes cluster**: Local (Docker Desktop, minikube, kind) or remote cluster access

### Installation Methods

#### Method 1: Automated Build Script (Recommended)

```bash
# Clone the repository
git clone <repository-url> k8s-cli
cd k8s-cli

# Make build script executable and run
chmod +x build.sh
./build.sh

# Install to system PATH (optional)
./build.sh install
```

#### Method 2: Using Make

```bash
# Complete build process
make all

# Install to system
make install

# Show all available commands
make help
```

#### Method 3: Manual Build

```bash
# Initialize and download dependencies
go mod init k8s-cli
go mod tidy

# Build the binary
go build -o bin/k8s-cli main.go

# Install to system (optional)
sudo cp bin/k8s-cli /usr/local/bin/
```

### Verification

```bash
# Test installation
k8s-cli --help
k8s-cli context current
k8s-cli list deployments
```

## 📋 Complete Command Reference

### Global Flags

All commands support these global flags:

```bash
--kubeconfig string    Path to kubeconfig file (default: ~/.kube/config)
-n, --namespace string Namespace for operations (default: "default")  
-o, --output string    Output format: table, json, yaml (default: "table")
```

### Context Management

```bash
# List all available contexts
k8s-cli context list

# Show current active context
k8s-cli context current

# Switch to a different context
k8s-cli context set <context-name>

# Examples
k8s-cli context set docker-desktop
k8s-cli context set minikube
k8s-cli context set production-cluster
```

### Resource Listing

```bash
# List pods in default namespace
k8s-cli list pods

# List pods in specific namespace
k8s-cli list pods -n kube-system

# List with label selector
k8s-cli list pods -l app=nginx

# List deployments (Step 6 requirement)
k8s-cli list deployments
k8s-cli list deployments -n production

# List services
k8s-cli list services

# List all namespaces
k8s-cli list namespaces

# Different output formats
k8s-cli list pods -o json
k8s-cli list deployments -o yaml
k8s-cli list services -o table
```

### Declarative Resource Management (YAML Files)

#### Apply Resources

```bash
# Apply resource from YAML file
k8s-cli apply file examples/pod.yaml
k8s-cli apply file examples/deployment.yaml
k8s-cli apply file examples/service.yaml

# Apply to specific namespace
k8s-cli apply file deployment.yaml -n my-app
```

#### Delete Resources from YAML

```bash
# Delete resource from YAML file
k8s-cli delete file examples/pod.yaml

# Delete with force (skip confirmation)
k8s-cli delete file examples/deployment.yaml --force

# Delete from specific namespace
k8s-cli delete file service.yaml -n production
```

### Imperative Resource Management (kubectl-style)

#### Create Deployments

```bash
# Basic deployment creation (like kubectl create deploy)
k8s-cli create deployment demo2 --image=gcr.io/kuber-351315/week-3:v1.0.0

# With replica count
k8s-cli create deployment nginx --image=nginx:1.20 --replicas=3

# With port exposure
k8s-cli create deployment web-app --image=nginx:1.20 --port=80 --replicas=2

# In specific namespace
k8s-cli create deployment api --image=my-api:v1.0.0 -n production --replicas=5
```

#### Create Pods

```bash
# Basic pod creation
k8s-cli create pod test-pod --image=nginx:1.20

# With port exposure
k8s-cli create pod web-pod --image=gcr.io/kuber-351315/week-3:v1.0.0 --port=8080

# In specific namespace
k8s-cli create pod debug-pod --image=busybox:latest -n development
```

#### Create Services

```bash
# ClusterIP service (default)
k8s-cli create service my-service --port=80 --target-port=8080

# NodePort service
k8s-cli create service web-svc --port=80 --type=NodePort

# LoadBalancer service
k8s-cli create service api-svc --port=443 --type=LoadBalancer

# With custom selector
k8s-cli create service demo-svc --port=80 --selector=app=demo2
```

### Resource Deletion by Name

```bash
# Delete specific resources by name
k8s-cli delete pod test-pod
k8s-cli delete deployment nginx-deployment
k8s-cli delete service my-service

# Delete with force (skip confirmation)
k8s-cli delete pod test-pod --force
k8s-cli delete deployment api --force

# Delete from specific namespace
k8s-cli delete deployment my-app -n production
k8s-cli delete service api-service -n staging
```

## 🛠 Development

### Build Commands

```bash
# Complete build process
make all

# Quick development build  
make quick

# Download dependencies
make deps

# Format code
make fmt

# Run tests
make test

# Clean build artifacts
make clean

# Install to system
make install

# Build for multiple platforms
make build-all

# Run demonstration
make demo

# Run integration tests
make integration
```

### Build Script Options

```bash
# Full build with all features
./build.sh

# Install to system PATH
./build.sh install

# Quick development build
./build.sh quick

# Download dependencies only
./build.sh deps-only

# Build only (skip setup)
./build.sh build-only

# Clean build artifacts
./build.sh clean

# Remove from system
./build.sh uninstall

# Show help
./build.sh help
```

### Project Structure

```
k8s-cli/
├── cmd/                    # CLI commands
│   ├── root.go            # Root command and global flags
│   ├── context.go         # Context management commands
│   ├── list.go            # Resource listing commands
│   ├── apply.go           # Declarative apply commands
│   ├── create.go          # Imperative create commands
│   └── delete.go          # Resource delete commands
├── internal/              # Internal packages
│   ├── k8s/              # Kubernetes client wrapper
│   │   └── client.go     # Client implementation with Step 6 features
│   └── utils/            # Utility functions
│       └── output.go     # Output formatting (table/json/yaml)
├── examples/             # Example YAML files
│   ├── pod.yaml         # Sample pod
│   ├── deployment.yaml  # Sample deployment
│   └── service.yaml     # Sample service
├── bin/                 # Built binaries (gitignored)
├── main.go             # Application entry point
├── go.mod              # Go module definition
├── Makefile            # Build automation
├── build.sh            # Comprehensive build script
└── README.md           # This file
```

## 🔧 Configuration

### Kubeconfig

The tool uses kubectl's standard kubeconfig file locations in this order:

1. `--kubeconfig` flag value
2. `K8S_CLI_KUBECONFIG` environment variable
3. `~/.kube/config` (default)

### Environment Variables

```bash
# Set custom kubeconfig
export K8S_CLI_KUBECONFIG=/path/to/config

# Set default namespace
export K8S_CLI_NAMESPACE=my-namespace

# Set default output format
export K8S_CLI_OUTPUT=json
```

### Configuration File

Create `~/.k8s-cli.yaml` for persistent settings:

```yaml
kubeconfig: /path/to/custom/config
namespace: my-default-namespace
output: table
```

## 📖 Usage Examples

### Basic Workflow

```bash
# Check current context
k8s-cli context current

# List available contexts
k8s-cli context list
# Available contexts:
#   minikube
# * docker-desktop (current)
#   production

# Switch to development context
k8s-cli context set docker-desktop

# View resources (Step 6 compliance)
k8s-cli list namespaces
k8s-cli list pods
k8s-cli list deployments

# Create resources imperatively (kubectl-style)
k8s-cli create deployment demo2 --image=gcr.io/kuber-351315/week-3:v1.0.0 --replicas=3
k8s-cli create service demo2-svc --port=80 --selector=app=demo2

# Or create resources declaratively
k8s-cli apply file examples/deployment.yaml
k8s-cli apply file examples/service.yaml

# Check deployment status
k8s-cli list pods -l app=demo2
k8s-cli list deployments

# Clean up
k8s-cli delete deployment demo2
k8s-cli delete service demo2-svc
```

### Step 6 Requirements Demo

```bash
# List Kubernetes deployment resources in default namespace
# Auth via kubeconfig with flags for custom kubeconfig file
k8s-cli list deployments
k8s-cli --kubeconfig=/custom/path list deployments
k8s-cli list deployments -n default -o json
k8s-cli list deployments --kubeconfig=~/.kube/config -o yaml

# Create deployment imperatively and list
k8s-cli create deployment test-app --image=nginx:1.20 --replicas=2
k8s-cli list deployments
k8s-cli delete deployment test-app
```

### Complete kubectl Replacement Workflow

```bash
# Instead of: kubectl create deploy nginx --image=nginx:1.20
k8s-cli create deployment nginx --image=nginx:1.20

# Instead of: kubectl get deployments
k8s-cli list deployments

# Instead of: kubectl apply -f deployment.yaml
k8s-cli apply file deployment.yaml

# Instead of: kubectl delete deploy nginx
k8s-cli delete deployment nginx

# Instead of: kubectl get pods -o json
k8s-cli list pods -o json

# Instead of: kubectl get pods -n kube-system
k8s-cli list pods -n kube-system
```

### Multi-Environment Management

```bash
# Production environment
k8s-cli context set production
k8s-cli list deployments -n app-production

# Staging environment  
k8s-cli context set staging
k8s-cli create deployment api-v2 --image=myapp:v2.0.0 -n staging

# Development environment
k8s-cli context set docker-desktop
k8s-cli apply file examples/pod.yaml
```

### Advanced Usage

```bash
# Use custom kubeconfig
k8s-cli --kubeconfig=/path/to/config list pods

# Work with different namespace as default
k8s-cli -n production list deployments

# Combine multiple flags
k8s-cli --kubeconfig=./config -n kube-system list pods -o json

# Use environment variables
export K8S_CLI_KUBECONFIG=/path/to/config
export K8S_CLI_NAMESPACE=production
k8s-cli list deployments
```

## 🔍 Troubleshooting

### Common Issues

1. **"Unable to connect to cluster"**
   ```bash
   # Check kubeconfig
   kubectl config current-context
   k8s-cli context current
   
   # Test connectivity
   kubectl get nodes
   k8s-cli list namespaces
   
   # Use custom kubeconfig
   k8s-cli --kubeconfig=/path/to/config list pods
   ```

2. **"Context not found"**
   ```bash
   # List available contexts
   kubectl config get-contexts
   k8s-cli context list
   
   # Switch to valid context
   k8s-cli context set docker-desktop
   ```

3. **"Permission denied" or RBAC errors**
   ```bash
   # Check RBAC permissions
   kubectl auth can-i get pods
   kubectl auth can-i create deployments
   kubectl auth can-i list deployments
   ```

4. **"EOF" errors during resource creation**
   ```bash
   # Check cluster status
   kubectl cluster-info
   
   # Restart cluster if needed (Docker Desktop)
   # Or: minikube start
   
   # Test with simple operations first
   k8s-cli list namespaces
   ```

5. **Build issues**
   ```bash
   # Clean and rebuild
   make clean
   make all
   
   # Or use build script
   ./build.sh clean
   ./build.sh
   ```

### Verification Commands

```bash
# Test CLI functionality
k8s-cli --help
k8s-cli context current
k8s-cli list namespaces

# Test deployment listing (Step 6)
k8s-cli list deployments
k8s-cli --kubeconfig=~/.kube/config list deployments -o json

# Test imperative creation
k8s-cli create pod test-pod --image=nginx:1.20
k8s-cli list pods
k8s-cli delete pod test-pod

# Test declarative operations
k8s-cli apply file examples/pod.yaml
k8s-cli delete file examples/pod.yaml
```

## 🚀 Quick Setup Guide

### 1. Install and Build
```bash
git clone <repository-url> k8s-cli
cd k8s-cli
chmod +x build.sh
./build.sh install
```

### 2. Test Installation
```bash
k8s-cli --help
k8s-cli context current
k8s-cli list deployments
```

### 3. Verify Step 6 Requirements
```bash
# List deployments with kubeconfig auth and flags
k8s-cli list deployments
k8s-cli --kubeconfig=/path/to/config list deployments
k8s-cli list deployments -n default -o json
```

### 4. Test Full Functionality
```bash
# Imperative creation (kubectl-style)
k8s-cli create deployment test --image=nginx:1.20
k8s-cli create service test-svc --port=80 --selector=app=test

# Declarative management
k8s-cli apply file examples/deployment.yaml

# Resource management
k8s-cli list deployments
k8s-cli delete deployment test
```

## 🎉 Success Criteria

✅ **Apply/Delete Commands**: Complete YAML file and resource name support  
✅ **Step 6 Compliance**: Deployment listing with kubeconfig authentication  
✅ **Kubeconfig Flags**: Full support for custom kubeconfig paths  
✅ **List CLI Commands**: All resource types with proper client integration  
✅ **Imperative Creation**: kubectl-style resource creation commands  
✅ **Build Automation**: Bash script and Makefile for easy setup  
✅ **English Documentation**: Comprehensive README with examples  
✅ **Cross-Platform**: Builds for Linux, macOS, and Windows  
✅ **System Installation**: Easy installation to system PATH

## 🔧 Requirements

- Go 1.21 or higher
- kubectl configured for cluster access
- Access to a Kubernetes cluster (local or remote)

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go best practices and formatting
- Add tests for new features
- Update documentation for new commands
- Use conventional commit messages
- Ensure cross-platform compatibility

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Cobra CLI](https://github.com/spf13/cobra) - Powerful CLI framework for Go
- [Kubernetes Client-Go](https://github.com/kubernetes/client-go) - Official Kubernetes API client
- [Viper](https://github.com/spf13/viper) - Complete configuration solution
- [kubectl](https://kubernetes.io/docs/reference/kubectl/) - Inspiration for command patterns

## 📞 Support

- 🐛 **Bug Reports**: Create an issue with detailed reproduction steps
- 💡 **Feature Requests**: Submit enhancement proposals
- 📚 **Documentation**: Check this README and inline help commands
- 🔧 **Troubleshooting**: Follow the troubleshooting section above

---

**Happy Kubernetes management with k8s-cli! 🚀**

*Built with ❤️ for the Kubernetes community*