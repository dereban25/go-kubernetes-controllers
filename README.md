# k8s-cli

[![Overall CI/CD](https://github.com/dereban25/go-kubernetes-controllers/actions/workflows/ci.yaml/badge.svg?branch=main)](https://github.com/dereban25/go-kubernetes-controllers/actions/workflows/ci.yaml)
[![Go Version](https://img.shields.io/badge/Go-1.21-blue)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

Tool CLI for K8S.

## 🚀 CI/CD Pipeline Status

[![Code Quality](https://img.shields.io/github/actions/workflow/status/dereban25/go-kubernetes-controllers/ci.yml?branch=main&label=Code%20Quality&logo=github)](https://github.com/dereban25/go-kubernetes-controllers/actions/workflows/ci.yaml)
[![Tests](https://img.shields.io/github/actions/workflow/status/dereban25/go-kubernetes-controllers/ci.yml?branch=main&label=Tests&logo=github)](https://github.com/dereban25/go-kubernetes-controllers/actions/workflows/ci.yaml)
[![Multi-Platform Build](https://img.shields.io/github/actions/workflow/status/dereban25/go-kubernetes-controllers/ci.yml?branch=main&label=Multi-Platform%20Build&logo=github)](https://github.com/dereban25/go-kubernetes-controllers/actions/workflows/ci.yaml)

### Pipeline Jobs

| Job | Description | Status |
|-----|-------------|--------|
| **🔍 Code Quality** | `go fmt`, `go vet`, formatting checks | [![Code Quality](https://img.shields.io/badge/Status-✅%20Passing-brightgreen)](https://github.com/dereban25/go-kubernetes-controllers/actions/workflows/ci.yaml) |
| **🧪 Tests** | Unit tests, build verification, syntax validation | [![Tests](https://img.shields.io/badge/Status-✅%20Passing-brightgreen)](https://github.com/dereban25/go-kubernetes-controllers/actions/workflows/ci.yaml) |
| **🔨 Multi-Platform Build** | Linux, macOS, Windows binaries | [![Build](https://img.shields.io/badge/Status-✅%20Success-brightgreen)](https://github.com/dereban25/go-kubernetes-controllers/actions/workflows/ci.yaml) |
| **🚀 Release** | Automatic GitHub releases on tags | [![Release](https://img.shields.io/badge/Status-⏳%20On%20Tags-yellow)](https://github.com/dereban25/go-kubernetes-controllers/actions/workflows/ci.yaml) |
| **📊 Status Report** | Build summary and artifact upload | [![Status](https://img.shields.io/badge/Status-✅%20Complete-brightgreen)](https://github.com/dereban25/go-kubernetes-controllers/actions/workflows/ci.yaml) |

## 📊 Latest Build Info

| Metric | Value |
|--------|-------|
| **Last Build** | [![Last Build](https://img.shields.io/badge/Status-✅%20Success-brightgreen)](https://github.com/dereban25/go-kubernetes-controllers/actions/workflows/ci.yaml) |
| **Build Time** | ~5 minutes |
| **Artifacts** | 3 platform binaries |
| **Success Rate** | ![Success Rate](https://img.shields.io/badge/Success%20Rate-100%25-brightgreen) |

## ⚡ Quick Links

- [📋 **View Latest Run**](https://github.com/dereban25/go-kubernetes-controllers/actions/workflows/ci.yaml) - See current pipeline status
- [📦 **Download Binaries**](https://github.com/dereban25/go-kubernetes-controllers/actions/workflows/ci.yaml) - Get latest artifacts  
- [🏷️ **Releases**](https://github.com/dereban25/go-kubernetes-controllers/releases) - Stable versions

## 🛠️ Installation

### Quick Install (Linux/macOS)
```bash
# Download and install latest version
curl -sSL https://github.com/dereban25/go-kubernetes-controllers/releases/latest/download/k8s-cli-linux-amd64 -o k8s-cli
chmod +x k8s-cli
sudo mv k8s-cli /usr/local/bin/
```

### Manual Download
- [Linux (amd64)](https://github.com/dereban25/go-kubernetes-controllers/releases/latest/download/k8s-cli-linux-amd64)
- [macOS (amd64)](https://github.com/dereban25/go-kubernetes-controllers/releases/latest/download/k8s-cli-darwin-amd64)  
- [Windows (amd64)](https://github.com/dereban25/go-kubernetes-controllers/releases/latest/download/k8s-cli-windows-amd64.exe)

## 🚀 Usage

```bash
# Basic commands
k8s-cli --help
k8s-cli --version

# Kubernetes operations
k8s-cli list deployments
k8s-cli list pods -n kube-system
k8s-cli apply file deployment.yaml
```

## 🛠️ Development

```bash
# Clone and setup
git clone https://github.com/dereban25/go-kubernetes-controllers.git
cd go-kubernetes-controllers/k8s-cli

# Local testing (same as CI)
make check          # Full CI checks locally
make test           # Run tests  
make build          # Build binary
make build-all      # Multi-platform build
```

## 📈 Build History

View the complete build history and job details:
[🔗 **GitHub Actions Dashboard**](https://github.com/dereban25/go-kubernetes-controllers/actions/workflows/ci.yaml)
