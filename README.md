# k8s-cli

[![k8s-cli CI/CD](https://github.com/dereban25/go-kubernetes-controllers/actions/workflows/ci.yaml/badge.svg)](https://github.com/dereban25/go-kubernetes-controllers/actions/workflows/ci.yaml)
[![Go Version](https://img.shields.io/badge/Go-1.21-blue)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

Tool CLI for K8S.

## 🚀 CI/CD Pipeline Status

[![CI/CD](https://github.com/dereban25/go-kubernetes-controllers/actions/workflows/ci.yaml/badge.svg?branch=main)](https://github.com/dereban25/go-kubernetes-controllers/actions/workflows/ci.yaml)

### Pipeline Jobs

| Job | Description | Status |
|-----|-------------|--------|
| **🔍 Code Quality** | `go fmt`, `go vet`, formatting checks | Check workflow → |
| **🧪 Tests** | Unit tests, build verification, syntax validation | Check workflow → |
| **🔨 Multi-Platform Build** | Linux, macOS, Windows binaries | Check workflow → |
| **🚀 Release** | Automatic GitHub releases on tags | On tags only |
| **📊 Status Report** | Build summary and artifact upload | Check workflow → |

## 📊 Latest Build Info

| Metric | Value |
|--------|-------|
| **Last Build** | [![CI/CD](https://github.com/dereban25/go-kubernetes-controllers/actions/workflows/ci.yaml/badge.svg)](https://github.com/dereban25/go-kubernetes-controllers/actions/workflows/ci.yaml) |
| **Build Time** | ~5 minutes |
| **Artifacts** | 3 platform binaries |

⚡ Quick Links

📋 [View Latest Run](https://github.com/dereban25/go-kubernetes-controllers/actions/workflows/ci.yaml) - See current pipeline status
📦 [Download Binaries](https://github.com/dereban25/go-kubernetes-controllers/actions) - Get latest artifacts
🏷️ [Releases](https://github.com/dereban25/go-kubernetes-controllers/releases) - Stable versions

🎯 Features Overview
Core CLI Functionality (Steps 1-6)

✅ Context Management: Switch between Kubernetes contexts seamlessly
✅ Resource Viewing: List pods, deployments, services, and namespaces
✅ Declarative Management: Apply and delete resources from YAML files
✅ Imperative Management: Create resources directly with command-line flags
✅ Step 6 Compliance: Full support for listing Kubernetes deployment resources

Advanced Informer Features (Steps 7-8)

✅ Step 7: k8s.io/client-go Informers: Watch deployment events using SharedInformerFactory
✅ Step 7+: JSON API Server: Cache access via HTTP API with real-time deployment data
✅ Step 7++: Configuration Management: YAML-based configuration with validation
✅ Step 8: Advanced API: Enhanced cache handlers with filtering, search, and analytics

Controller Runtime Features (Steps 9-10)

✅ Step 9: sigs.k8s.io/controller-runtime: Event-driven controller with reconciliation logic
✅ Step 9+: Multi-cluster Informers: Dynamically created informers for multiple clusters
✅ Step 10: Controller Manager: Centralized management with leader election using lease resources

Custom Resource Features (Steps 11-12++)

✅ Step 11: Custom CRD: FrontendPage custom resource with full lifecycle management
✅ Step 11++: Multi-cluster Management: Cross-cluster resource synchronization
✅ Step 12: Platform Engineering: Port.io integration for self-service experiences
✅ Step 12+: Update Actions: IDP controller integration with enhanced CRUD operations
✅ Step 12++: Discord Integration: Rich notification system with embed messages

🛠️ Installation
Quick Install (Linux/macOS)
```bash
# Download and install latest version
curl -sSL https://github.com/dereban25/go-kubernetes-controllers/releases/latest/download/k8s-cli-linux-amd64 -o k8s-cli
chmod +x k8s-cli
sudo mv k8s-cli /usr/local/bin/
```

Manual Download

[Linux (amd64)](https://github.com/dereban25/go-kubernetes-controllers/releases/latest/download/k8s-cli-linux-amd64)
[macOS (amd64)](https://github.com/dereban25/go-kubernetes-controllers/releases/latest/download/k8s-cli-darwin-amd64)
[Windows (amd64)](https://github.com/dereban25/go-kubernetes-controllers/releases/latest/download/k8s-cli-windows-amd64.exe)

Development Setup
```bash
# Clone and setup
git clone https://github.com/dereban25/go-kubernetes-controllers.git
cd go-kubernetes-controllers/k8s-cli

# Install dependencies and setup CRDs
make deps
make generate
make manifests
make install-crds

# Build and test
make build
make test-complete
```

🚀 Usage
Basic Commands
```bash
# Basic commands
k8s-cli --help
k8s-cli --version

# Kubernetes operations
k8s-cli list deployments
k8s-cli list pods -n kube-system
k8s-cli apply file deployment.yaml
```

Advanced Features
Step 7-8: Informers and APIs
```bash
# Start informer with event logging
k8s-cli watch-informer --workers 2 --resync-period 30s --log-events

# Start JSON API server for cache access
k8s-cli api-server --port 8080

# Start advanced API with filtering and search
k8s-cli step8-api --port 8090 --enable-debug --enable-metrics

# Test API endpoints
curl http://localhost:8080/api/v1/deployments
curl 'http://localhost:8090/api/v2/deployments?sortBy=name&pageSize=5'
curl 'http://localhost:8090/api/v2/cache/search?q=nginx'
```

Step 9-10: Controller Runtime and Manager
```bash
# Start controller with reconciliation logic
k8s-cli controller --workers 2

# Start manager with leader election
k8s-cli manager --enable-leader-election --metrics-port 8080 --health-port 8081

# Check leader election and health
kubectl get leases -n kube-system | grep k8s-cli
curl http://localhost:8081/healthz
curl http://localhost:8080/metrics
```

Step 11: Custom CRD (FrontendPage)
```bash
# Start CRD controller
k8s-cli crd --metrics-port 8082 --health-port 8083

# Create FrontendPage resource
kubectl apply -f - <<EOF
apiVersion: k8scli.dev/v1
kind: FrontendPage
metadata:
  name: my-frontend
spec:
  title: "My Frontend App"
  description: "A sample frontend application"
  path: "/app"
  replicas: 2
  image: "nginx:1.20"
  config:
    ENVIRONMENT: "production"
EOF

# Check created resources
kubectl get frontendpages
kubectl describe frontendpage my-frontend
kubectl get deployments,services | grep my-frontend
```

Step 12++: Platform Engineering with Discord
```bash
# Start platform API with Port.io and Discord integration
k8s-cli platform --port 8084 \
  --port-token $PORT_API_TOKEN \
  --discord-webhook $DISCORD_WEBHOOK_URL

# Create via CRUD API
curl -X POST http://localhost:8084/api/v1/frontendpages \
  -H 'Content-Type: application/json' \
  -d '{
    "metadata": {"name": "api-frontend"},
    "spec": {
      "title": "API Created Frontend",
      "path": "/api",
      "replicas": 2
    }
  }'

# Trigger Port.io action (will send Discord notification)
curl -X POST http://localhost:8084/webhook/port \
  -H 'Content-Type: application/json' \
  -d '{
    "action": "create_frontend",
    "resourceId": "frontend-123",
    "inputs": {
      "name": "port-frontend",
      "title": "Port Created Frontend",
      "path": "/port",
      "replicas": 3
    }
  }'

# Update action (Step 12+)
curl -X POST http://localhost:8084/api/v1/frontendpages/update \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "api-frontend",
    "updates": {
      "title": "Updated Frontend",
      "replicas": 5
    }
  }'
```

🧪 Testing and Verification
Quick Test All Steps
```bash
# Run complete test suite
make test-complete

# Expected output:
# ✅ Step 7: Informers - TESTED
# ✅ Step 7+: JSON API - TESTED
# ✅ Step 8: Advanced API - TESTED
# ✅ Step 9: Controller-Runtime - TESTED
# ✅ Step 10: Manager with Leader Election - TESTED
# ✅ Step 11: Custom CRD - TESTED
# ✅ Step 12: Platform Engineering API - TESTED
# 🎉 Your k8s-cli is ready with Steps 7-12++ support!
```

Test Individual Steps
```bash
# Test specific steps
make test-step7        # Informers
make test-step7plus    # JSON API
make test-step8        # Advanced API
make test-step9        # Controller Runtime
make test-step10       # Manager with Leader Election
make test-step11       # CRD Controller
make test-step12       # Platform Engineering API

# Run all step tests
make test-all-steps

# Demo scenarios
make demo-frontendpage  # Create sample FrontendPage
make demo-platform      # Test platform actions
```

Manual Testing Workflows
End-to-End Workflow Test
```bash
# 1. Start all services in separate terminals
k8s-cli watch-informer --config ~/.k8s-cli/config.yaml &
k8s-cli api-server --port 8080 &
k8s-cli step8-api --port 8090 --enable-debug &
k8s-cli controller --workers 2 &
k8s-cli manager --enable-leader-election=false --metrics-port 8081 &
k8s-cli crd --metrics-port 8082 &
k8s-cli platform --port 8084 &

# 2. Create resources via different methods
kubectl create deployment manual-deploy --image=nginx:1.20
k8s-cli create deployment cli-deploy --image=nginx:1.20 --replicas=2

# 3. Create FrontendPage via platform API
curl -X POST http://localhost:8084/webhook/port \
  -H 'Content-Type: application/json' \
  -d '{
    "action": "create_frontend",
    "inputs": {
      "name": "test-complete",
      "title": "Complete Test",
      "path": "/test",
      "replicas": 3
    }
  }'

# 4. Monitor all services
curl http://localhost:8080/api/v1/deployments | jq .count
curl http://localhost:8090/api/v2/cache/metrics | jq .total_deployments
curl http://localhost:8081/metrics | grep k8s_cli
curl http://localhost:8082/healthz
curl http://localhost:8084/health

# 5. Verify Discord notifications (if configured)
# Check Discord channel for action notifications

# 6. Cleanup
kubectl delete deployments --all
kubectl delete frontendpages --all
```

Step-by-Step Verification
Step 7-8: Informers and APIs
```bash
# Terminal 1: Start informer
k8s-cli watch-informer --workers 2 --log-events

# Terminal 2: Create test deployment
kubectl create deployment test-step7 --image=nginx:1.20
kubectl scale deployment test-step7 --replicas=3
kubectl delete deployment test-step7

# Observe logs in Terminal 1 showing ADD/UPDATE/DELETE events

# Terminal 3: Test API access
curl http://localhost:8080/api/v1/deployments | jq .
curl 'http://localhost:8090/api/v2/deployments?sortBy=name&pageSize=3' | jq .
```

Step 9-10: Controllers and Managers
```bash
# Terminal 1: Start controller
k8s-cli controller --workers 2

# Terminal 2: Create test resources
kubectl create deployment test-step9 --image=nginx:1.20
kubectl scale deployment test-step9 --replicas=2

# Observe reconciliation logs in Terminal 1 with "Step 9" prefix

# Terminal 3: Test manager with leader election
k8s-cli manager --enable-leader-election --metrics-port 8080 --health-port 8081

# Check leader election and health
kubectl get leases -n kube-system | grep k8s-cli
curl http://localhost:8081/healthz
```

Step 11: Custom CRDs
```bash
# Install CRDs and start controller
make install-crds
k8s-cli crd --metrics-port 8082 --health-port 8083

# Create FrontendPage and verify reconciliation
kubectl apply -f examples/frontendpage-demo.yaml

# Check all created resources
kubectl get frontendpages
kubectl describe frontendpage demo-frontend
kubectl get deployments,services | grep demo-frontend

# Update the FrontendPage and observe controller response
kubectl patch frontendpage demo-frontend --type='merge' -p='{"spec":{"replicas":5}}'
```

Step 12++: Platform Engineering
```bash
# Start platform API with Discord integration
k8s-cli platform --port 8084 --discord-webhook $DISCORD_WEBHOOK_URL

# Test all CRUD operations
curl -X POST http://localhost:8084/api/v1/frontendpages \
  -H 'Content-Type: application/json' \
  -d '{"metadata":{"name":"crud-test"},"spec":{"title":"CRUD Test","path":"/crud","replicas":1}}'

curl -X GET http://localhost:8084/api/v1/frontendpages | jq .

curl -X PUT http://localhost:8084/api/v1/frontendpages/crud-test \
  -H 'Content-Type: application/json' \
  -d '{"spec":{"title":"Updated CRUD Test","replicas":3}}'

curl -X DELETE http://localhost:8084/api/v1/frontendpages/crud-test

# Test Port.io webhook actions
curl -X POST http://localhost:8084/webhook/port \
  -H 'Content-Type: application/json' \
  -d '{
    "action": "create_frontend",
    "resourceId": "frontend-test",
    "trigger": "manual",
    "inputs": {
      "name": "webhook-test",
      "title": "Webhook Test",
      "path": "/webhook",
      "replicas": 2
    }
  }'

# Check Discord for rich notification with action details
```

🔧 Configuration
Environment Variables
```bash
# Step 7++ Configuration
export K8S_CLI_KUBECONFIG=/path/to/config
export K8S_CLI_NAMESPACE=my-namespace

# Step 12 Platform Integration
export PORT_API_TOKEN=your_port_api_token
export DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/your/webhook

# Step 10 Leader Election
export K8S_CLI_LEADER_ELECTION=true
export K8S_CLI_LEADER_ELECTION_ID=k8s-cli-manager
```

Configuration Files
Step 7++ Config (~/.k8s-cli/config.yaml):
```yaml
resync_period: "30s"
workers: 2
namespaces: ["default", "kube-system"]
log_events: true

api_server:
  enabled: true
  port: 8080

custom_logic:
  enable_update_handling: true
  enable_delete_handling: true
```

Step 11++ Multi-cluster Config (~/.k8s-cli/clusters.yaml):
```yaml
clusters:
  production:
    kubeconfig: ~/.kube/config-prod
    context: prod-cluster
    namespace: frontend-prod
    enabled: true
  staging:
    kubeconfig: ~/.kube/config-staging
    context: staging-cluster
    namespace: frontend-staging
    enabled: true
```

🛠️ Development
Build Commands
```bash
# Complete development setup
make deps              # Install dependencies including controller-gen
make generate          # Generate DeepCopy methods
make manifests         # Generate CRD manifests
make install-crds      # Install CRDs to cluster
make build             # Build binary
make test-complete     # Run complete test suite

# Development servers (run in separate terminals)
make dev-watch         # Step 7: Informer
make dev-api           # Step 7+: JSON API
make dev-step8         # Step 8: Advanced API
make dev-controller    # Step 9: Controller
make dev-manager       # Step 10: Manager
make dev-crd           # Step 11: CRD Controller
make dev-platform      # Step 12: Platform API
```

Local testing (same as CI)
```bash
make check          # Full CI checks locally
make test           # Run tests  
make build          # Build binary
make build-all      # Multi-platform build
```

Project Structure
```
k8s-cli/
├── cmd/                           # CLI commands
│   ├── root.go                   # Root command and global flags
│   ├── context.go, list.go, etc. # Core commands (Steps 1-6)
│   ├── config.go, informer.go    # Step 7: Informers
│   ├── api.go, cache.go          # Step 7+/8: APIs
│   ├── controller.go             # Step 9: Controller Runtime
│   ├── manager.go                # Step 10: Manager
│   ├── crd.go                    # Step 11: CRD Controller
│   └── platform.go               # Step 12: Platform API
├── api/v1/                       # Step 11: Custom Resources
│   ├── frontendpage_types.go     # FrontendPage CRD
│   └── groupversion_info.go      # API group info
├── controllers/                  # Step 11: Controllers
│   └── frontendpage_controller.go # FrontendPage controller
├── config/crd/bases/             # Generated CRD manifests
├── internal/                     # Internal packages
├── examples/                     # Example YAML files
├── tests/                        # Test files
├── Makefile                      # Build automation with all steps
└── README.md                     # This file
```

🎉 Success Criteria
Complete Feature Verification

Steps 1-6: Core CLI functionality works
Step 7: Informers report events in logs using k8s.io/client-go
Step 7+: JSON API provides cache access via HTTP
Step 7++: Configuration management with YAML validation
Step 8: Advanced API with filtering, search, and analytics
Step 9: Controller-runtime with reconciliation logic reporting events
Step 9+: Multi-cluster informers with dynamic creation
Step 10: Controller manager with leader election using lease resources
Step 11: Custom FrontendPage CRD with full reconciliation
Step 11++: Multi-cluster management configuration
Step 12: Platform engineering API with Port.io integration
Step 12+: Update action support for IDP controller
Step 12++: Discord notifications with rich embeds

Architecture Flow Verification
```
kubectl operations → Informers (Step 7) → JSON APIs (Step 7+/8) →
Controller Runtime (Step 9) → Manager (Step 10) →
Custom CRDs (Step 11) → Platform API (Step 12) → Discord (Step 12++)
```

📈 Build History
View the complete build history and job details:
🔗 [GitHub Actions Dashboard](https://github.com/dereban25/go-kubernetes-controllers/actions)

🤝 Contributing

1. Fork the repository
2. Create a feature branch for your step (`git checkout -b feature/step13-new-feature`)
3. Implement following the established patterns
4. Add tests and update Makefile
5. Update this README with your step documentation
6. Commit and create Pull Request

📄 License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---
**k8s-cli: From simple CLI to advanced platform engineering with Steps 7-12++! 🚀**