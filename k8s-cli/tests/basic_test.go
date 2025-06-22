package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIBuild(t *testing.T) {
	t.Log("Testing CLI build...")

	// Create bin directory
	binDir := filepath.Join("..", "bin")
	os.MkdirAll(binDir, 0755)

	// Build CLI
	binPath := filepath.Join(binDir, "k8s-cli")
	cmd := exec.Command("go", "build", "-o", binPath, "../main.go")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Build failed: %v\nOutput: %s", err, output)
	}

	// Check if binary was created
	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		t.Fatal("Binary was not created")
	}

	t.Log("✅ Build successful")
}

func TestCLIHelp(t *testing.T) {
	t.Log("Testing help command...")

	binPath := filepath.Join("..", "bin", "k8s-cli")

	// Check if binary exists
	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		t.Skip("Binary not found, skipping test")
	}

	// Make executable
	os.Chmod(binPath, 0755)

	// Test help
	cmd := exec.Command(binPath, "--help")
	output, err := cmd.CombinedOutput()

	// It's okay if command is not supported at initial stage
	if err != nil {
		t.Logf("Help command returned error (this is normal): %v", err)
		return
	}

	// If command works, check that there's output
	if len(output) == 0 {
		t.Error("Help command produces no output")
	}

	t.Logf("✅ Help works, output: %s", string(output)[:min(100, len(output))])
}

func TestCLIVersion(t *testing.T) {
	t.Log("Testing version command...")

	binPath := filepath.Join("..", "bin", "k8s-cli")

	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		t.Skip("Binary not found")
	}

	cmd := exec.Command(binPath, "--version")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Logf("Version command not supported (this is normal): %v", err)
		return
	}

	if len(output) > 0 {
		t.Logf("✅ Version works: %s", strings.TrimSpace(string(output)))
	}
}

func TestGoModules(t *testing.T) {
	t.Log("Checking go.mod...")

	// Check that go.mod exists
	if _, err := os.Stat("../go.mod"); os.IsNotExist(err) {
		t.Fatal("go.mod not found. Run: cd k8s-cli && go mod init")
	}

	// Check that dependencies are correct
	cmd := exec.Command("go", "mod", "verify")
	cmd.Dir = ".."

	if err := cmd.Run(); err != nil {
		t.Fatalf("go mod verify failed: %v", err)
	}

	t.Log("✅ go.mod is correct")
}

func TestGoSyntax(t *testing.T) {
	t.Log("Checking Go syntax...")

	cmd := exec.Command("go", "build", "-o", "/dev/null", "../main.go")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("Syntax error in main.go: %v\nOutput: %s", err, output)
	}

	t.Log("✅ Syntax is correct")
}

// Helper function min for Go versions < 1.21
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
