#!/bin/bash
set -e

echo "🔧 Setting up code generation..."

# Ensure we're in the right directory
cd "$(dirname "$0")/.."

# Create hack directory if it doesn't exist
mkdir -p hack

# Create boilerplate file if it doesn't exist
if [ ! -f hack/boilerplate.go.txt ]; then
    cat > hack/boilerplate.go.txt << 'EOF'
/*
Copyright 2024 The k8s-cli Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
EOF
fi

# Check if controller-gen is available
if ! command -v controller-gen &> /dev/null; then
    echo "⚠️  controller-gen not found in PATH, checking GOPATH/bin..."
    export PATH=$PATH:$(go env GOPATH)/bin
fi

# Generate DeepCopy methods
echo "🔧 Generating DeepCopy methods..."
echo "📁 Current directory: $(pwd)"
echo "📁 API directory contents:"
ls -la api/v1/ || echo "No api/v1 directory"

controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./api/..." output:dir=api/v1

echo "📁 API directory after generation:"
ls -la api/v1/

# Check if generation was successful
if [ -f "api/v1/zz_generated.deepcopy.go" ]; then
    echo "✅ Code generation completed successfully"
    echo "📄 Generated file contents (first 50 lines):"
    head -n 50 api/v1/zz_generated.deepcopy.go
else
    echo "❌ Code generation failed - no generated files found"
    exit 1
fi

# Generate CRD manifests if requested
if [ "$1" == "crd" ] || [ "$1" == "all" ]; then
    echo "📄 Generating CRD manifests..."
    mkdir -p config/crd/bases
    controller-gen crd paths="./api/..." output:crd:artifacts:config=config/crd/bases
    echo "✅ CRD manifests generated"
fi