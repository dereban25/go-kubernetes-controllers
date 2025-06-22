// tests/cli_integration_test.go
package tests

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Путь к bin-файлу после сборки
func cliBin() string {
	return filepath.FromSlash("../bin/k8s-cli")
}

// Сборка бинаря, если его нет
func ensureBinary(t *testing.T) {
	if _, err := os.Stat(cliBin()); err == nil {
		return
	}
	t.Log("building k8s-cli …")
	cmd := exec.Command("go", "build", "-o", cliBin(), "../main.go")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
}

func skipIfNoCluster(t *testing.T) {
	if err := exec.Command("kubectl", "--request-timeout=5s", "cluster-info").Run(); err != nil {
		t.Skip("cluster is absent, skipping integration tests")
	}
}

func waitFor(t *testing.T, cmd *exec.Cmd) {
	for i := 0; i < 10; i++ {
		if err := cmd.Run(); err == nil {
			return
		}
		time.Sleep(time.Second)
	}
	out, _ := cmd.CombinedOutput()
	t.Fatalf("resource did not appear in time: %s", out)
}

func TestCreateListDeleteDeployment(t *testing.T) {
	skipIfNoCluster(t)
	ensureBinary(t)

	name := "ci-int-deploy"
	image := "nginx:1.25.5"

	// create
	if out, err := exec.Command(cliBin(), "create", "deployment", name, "--image="+image, "--replicas=2").CombinedOutput(); err != nil {
		t.Fatalf("create deployment: %v\n%s", err, out)
	}

	// verify via kubectl
	waitFor(t, exec.Command("kubectl", "get", "deploy", name))

	// list via CLI
	out, err := exec.Command(cliBin(), "list", "deployments").CombinedOutput()
	if err != nil || !bytes.Contains(out, []byte(name)) {
		t.Fatalf("deployment not in list: %v\n%s", err, out)
	}

	// delete
	if out, err := exec.Command(cliBin(), "delete", "deployment", name, "--force").CombinedOutput(); err != nil {
		t.Fatalf("delete deployment: %v\n%s", err, out)
	}

	// ensure gone
	if err := exec.Command("kubectl", "get", "deploy", name).Run(); err == nil {
		t.Fatalf("deployment %q still exists after deletion", name)
	}
}

func TestCreateDeletePod(t *testing.T) {
	skipIfNoCluster(t)
	ensureBinary(t)

	name := "ci-int-pod"
	image := "nginx:1.25.5"

	if out, err := exec.Command(cliBin(), "create", "pod", name, "--image="+image).CombinedOutput(); err != nil {
		t.Fatalf("create pod: %v\n%s", err, out)
	}
	waitFor(t, exec.Command("kubectl", "get", "pod", name))

	if out, err := exec.Command(cliBin(), "delete", "pod", name, "--force").CombinedOutput(); err != nil {
		t.Fatalf("delete pod: %v\n%s", err, out)
	}
}

func TestCreateDeleteService(t *testing.T) {
	skipIfNoCluster(t)
	ensureBinary(t)

	// Предварительно нужен pod/deployment с label app=svc-demo
	_ = exec.Command(cliBin(), "create", "deployment", "svc-demo", "--image=nginx:1.25.5").Run()
	waitFor(t, exec.Command("kubectl", "rollout", "status", "deployment/svc-demo"))

	svc := "ci-int-svc"
	if out, err := exec.Command(cliBin(), "create", "service", svc, "--port=80").CombinedOutput(); err != nil {
		t.Fatalf("create service: %v\n%s", err, out)
	}
	waitFor(t, exec.Command("kubectl", "get", "svc", svc))

	if out, err := exec.Command(cliBin(), "delete", "service", svc, "--force").CombinedOutput(); err != nil {
		t.Fatalf("delete service: %v\n%s", err, out)
	}
	_ = exec.Command(cliBin(), "delete", "deployment", "svc-demo", "--force").Run()
}

func TestApplyAndDeleteFile(t *testing.T) {
	skipIfNoCluster(t)
	ensureBinary(t)

	// tmp-файл с ConfigMap
	yaml := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: ci-int-cm
data:
  foo: bar
`
	tmp := filepath.Join(t.TempDir(), "cm.yaml")
	if err := os.WriteFile(tmp, []byte(strings.TrimSpace(yaml)), 0644); err != nil {
		t.Fatalf("write tmp yaml: %v", err)
	}

	if out, err := exec.Command(cliBin(), "apply", "file", tmp).CombinedOutput(); err != nil {
		t.Fatalf("apply file: %v\n%s", err, out)
	}
	waitFor(t, exec.Command("kubectl", "get", "cm", "ci-int-cm"))

	if out, err := exec.Command(cliBin(), "delete", "file", tmp, "--force").CombinedOutput(); err != nil {
		t.Fatalf("delete file: %v\n%s", err, out)
	}
}
