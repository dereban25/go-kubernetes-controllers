package tests

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestIntegrationWithCluster(t *testing.T) {
	// Skip if not in CI environment or no cluster available
	if os.Getenv("CI") == "" {
		t.Skip("Skipping integration test outside CI environment")
	}

	// Build the CLI
	cmd := exec.Command("go", "build", "-o", "../bin/k8s-cli", "../main.go")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build CLI: %v", err)
	}

	// Test context commands
	t.Run("ContextCommands", func(t *testing.T) {
		cmd := exec.Command("../bin/k8s-cli", "context", "current")
		if err := cmd.Run(); err != nil {
			t.Errorf("Failed to get current context: %v", err)
		}

		cmd = exec.Command("../bin/k8s-cli", "context", "list")
		if err := cmd.Run(); err != nil {
			t.Errorf("Failed to list contexts: %v", err)
		}
	})

	// Test listing commands
	t.Run("ListCommands", func(t *testing.T) {
		commands := [][]string{
			{"list", "namespaces"},
			{"list", "pods", "-n", "kube-system"},
			{"list", "deployments"},
			{"list", "services"},
		}

		for _, args := range commands {
			cmd := exec.Command("../bin/k8s-cli", args...)
			if err := cmd.Run(); err != nil {
				t.Errorf("Failed to run command %v: %v", args, err)
			}
		}
	})

	// Test Step 6 requirements
	t.Run("Step6Requirements", func(t *testing.T) {
		// Test deployment listing with kubeconfig auth
		cmd := exec.Command("../bin/k8s-cli", "list", "deployments")
		if err := cmd.Run(); err != nil {
			t.Errorf("Failed to list deployments: %v", err)
		}

		// Test with output formats
		cmd = exec.Command("../bin/k8s-cli", "list", "deployments", "-o", "json")
		if err := cmd.Run(); err != nil {
			t.Errorf("Failed to list deployments in JSON format: %v", err)
		}

		cmd = exec.Command("../bin/k8s-cli", "list", "deployments", "-o", "yaml")
		if err := cmd.Run(); err != nil {
			t.Errorf("Failed to list deployments in YAML format: %v", err)
		}
	})

	// Test imperative creation and deletion
	t.Run("ImperativeOperations", func(t *testing.T) {
		// Create deployment
		cmd := exec.Command("../bin/k8s-cli", "create", "deployment", "test-app", "--image=nginx:1.20")
		if err := cmd.Run(); err != nil {
			t.Errorf("Failed to create deployment: %v", err)
		}

		// Wait a bit for deployment to be created
		time.Sleep(10 * time.Second)

		// List deployments to verify
		cmd = exec.Command("../bin/k8s-cli", "list", "deployments")
		if err := cmd.Run(); err != nil {
			t.Errorf("Failed to list deployments after creation: %v", err)
		}

		// Delete deployment
		cmd = exec.Command("../bin/k8s-cli", "delete", "deployment", "test-app", "--force")
		if err := cmd.Run(); err != nil {
			t.Errorf("Failed to delete deployment: %v", err)
		}
	})
}

func TestYAMLOperations(t *testing.T) {
	if os.Getenv("CI") == "" {
		t.Skip("Skipping integration test outside CI environment")
	}

	// Test apply and delete from YAML
	t.Run("YAMLApplyDelete", func(t *testing.T) {
		// Apply pod
		cmd := exec.Command("../bin/k8s-cli", "apply", "file", "../examples/pod.yaml")
		if err := cmd.Run(); err != nil {
			t.Errorf("Failed to apply pod YAML: %v", err)
		}

		// Wait for pod to be ready
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		for {
			select {
			case <-ctx.Done():
				t.Error("Timeout waiting for pod to be ready")
				return
			default:
				cmd := exec.Command("kubectl", "get", "pod", "nginx-pod", "-o", "jsonpath={.status.phase}")
				output, err := cmd.Output()
				if err == nil && string(output) == "Running" {
					goto podReady
				}
				time.Sleep(5 * time.Second)
			}
		}

	podReady:
		// List pods to verify
		cmd = exec.Command("../bin/k8s-cli", "list", "pods")
		if err := cmd.Run(); err != nil {
			t.Errorf("Failed to list pods after apply: %v", err)
		}

		// Delete pod
		cmd = exec.Command("../bin/k8s-cli", "delete", "file", "../examples/pod.yaml", "--force")
		if err := cmd.Run(); err != nil {
			t.Errorf("Failed to delete pod from YAML: %v", err)
		}
	})
}
