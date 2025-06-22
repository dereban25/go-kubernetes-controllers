# k8s-cli

[![Overall CI/CD](https://github.com/dereban25/go-kubernetes-controllers/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/dereban25/go-kubernetes-controllers/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/Go-1.21-blue)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

Tool CLI for K8S.

## 🚀 CI/CD Pipeline Status

[![Code Quality](https://img.shields.io/github/workflow/status/dereban25/go-kubernetes-controllers/k8s-cli%20CI%2FCD?label=Code%20Quality&logo=github&style=flat-square)](https://github.com/dereban25/go-kubernetes-controllers/actions/workflows/ci.yml)
[![Tests](https://img.shields.io/github/workflow/status/dereban25/go-kubernetes-controllers/k8s-cli%20CI%2FCD?label=Tests&logo=github&style=flat-square)](https://github.com/dereban25/go-kubernetes-controllers/actions/workflows/ci.yml)
[![Build](https://img.shields.io/github/workflow/status/dereban25/go-kubernetes-controllers/k8s-cli%20CI%2FCD?label=Multi-Platform%20Build&logo=github&style=flat-square)](https://github.com/dereban25/go-kubernetes-controllers/actions/workflows/ci.yml)

### Pipeline Jobs

- **Code Quality**: `go fmt`, `go vet`, formatting checks
- **Tests**: Unit tests, build verification, syntax validation
- **Multi-Platform Build**: Linux, macOS, Windows binaries
- **Release**: Automatic GitHub releases on tags
- **Status Report**: Build summary and artifact upload

## 📊 Latest Build Info

| Metric | Value |
|--------|-------|
| **Last Build** | [![Last Commit](https://img.shields.io/github/last-commit/dereban25/go-kubernetes-controllers?style=flat-square)](https://github.com/dereban25/go-kubernetes-controllers/commits/main) |
| **Build Time** | ~5 minutes |
| **Artifacts** | 3 platform binaries |
| **Success Rate** | [![Build Success](https://img.shields.io/badge/Success%20Rate-100%25-brightgreen?style=flat-square)](#) |

## ⚡ Quick Links

- [📋 **View Latest Run**](https://github.com/dereban25/go-kubernetes-controllers/actions/workflows/ci.yml) - See current pipeline status
- [📦 **Download Binaries**](https://github.com/dereban25/go-kubernetes-controllers/actions/workflows/ci.yml) - Get latest artifacts
- [🏷️ **Releases**](https://github.com/dereban25/go-kubernetes-controllers/releases) - Stable versions

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
[🔗 **GitHub Actions Dashboard**](https://github.com/dereban25/go-kubernetes-controllers/actions/workflows/ci.yml)
