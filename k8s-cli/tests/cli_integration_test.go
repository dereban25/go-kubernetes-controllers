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

/* -------------------------------------------------------------------------- */
/*                               helper-functions                              */
/* -------------------------------------------------------------------------- */

func cliBin() string { return filepath.FromSlash("../bin/k8s-cli") }

func ensureBinary(t *testing.T) {
	if _, err := os.Stat(cliBin()); err == nil {
		return
	}
	t.Log("building k8s-cli …")
	if out, err := exec.Command("go", "build", "-o", cliBin(), "../main.go").CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
}

func skipIfNoCluster(t *testing.T) {
	if err := exec.Command("kubectl", "--request-timeout=5s", "cluster-info").Run(); err != nil {
		t.Skip("cluster is absent, skipping integration tests")
	}
}

func waitForCLI(t *testing.T, args []string, name string, wantPresent bool) {
	for i := 0; i < 10; i++ {
		out, _ := exec.Command(cliBin(), args...).CombinedOutput()
		has := bytes.Contains(out, []byte(name))
		if has == wantPresent {
			return
		}
		time.Sleep(time.Second)
	}
	state := "appear"
	if !wantPresent {
		state = "disappear"
	}
	t.Fatalf("resource %q did not %s via %v", name, state, args)
}

/* -------------------------------------------------------------------------- */
/*                                integration                                 */
/* -------------------------------------------------------------------------- */

func TestCreateListDeleteDeployment(t *testing.T) {
	skipIfNoCluster(t)
	ensureBinary(t)

	name := "ci-int-deploy"
	image := "nginx:1.25.5"

	// create
	if out, err := exec.Command(cliBin(), "create", "deployment", name, "--image="+image, "--replicas=2").CombinedOutput(); err != nil {
		t.Fatalf("create deployment: %v\n%s", err, out)
	}

	// verify presence
	waitForCLI(t, []string{"list", "deployments"}, name, true)

	// delete
	if out, err := exec.Command(cliBin(), "delete", "deployment", name, "--force").CombinedOutput(); err != nil {
		t.Fatalf("delete deployment: %v\n%s", err, out)
	}

	// verify absence
	waitForCLI(t, []string{"list", "deployments"}, name, false)
}

func TestCreateDeletePod(t *testing.T) {
	skipIfNoCluster(t)
	ensureBinary(t)

	name := "ci-int-pod"
	image := "nginx:1.25.5"

	// create
	if out, err := exec.Command(cliBin(), "create", "pod", name, "--image="+image).CombinedOutput(); err != nil {
		t.Fatalf("create pod: %v\n%s", err, out)
	}

	waitForCLI(t, []string{"list", "pods"}, name, true)

	// delete
	if out, err := exec.Command(cliBin(), "delete", "pod", name, "--force").CombinedOutput(); err != nil {
		t.Fatalf("delete pod: %v\n%s", err, out)
	}

	waitForCLI(t, []string{"list", "pods"}, name, false)
}

func TestCreateDeleteService(t *testing.T) {
	skipIfNoCluster(t)
	ensureBinary(t)

	// предварительно деплой, к которому привяжем service
	_ = exec.Command(cliBin(), "create", "deployment", "svc-demo", "--image=nginx:1.25.5").Run()
	waitForCLI(t, []string{"list", "deployments"}, "svc-demo", true)

	svc := "ci-int-svc"
	if out, err := exec.Command(cliBin(), "create", "service", svc, "--port=80").CombinedOutput(); err != nil {
		t.Fatalf("create service: %v\n%s", err, out)
	}

	waitForCLI(t, []string{"list", "services"}, svc, true)

	if out, err := exec.Command(cliBin(), "delete", "service", svc, "--force").CombinedOutput(); err != nil {
		t.Fatalf("delete service: %v\n%s", err, out)
	}

	waitForCLI(t, []string{"list", "services"}, svc, false)
	_ = exec.Command(cliBin(), "delete", "deployment", "svc-demo", "--force").Run()
}

func TestApplyAndDeleteFile(t *testing.T) {
	skipIfNoCluster(t)
	ensureBinary(t)

	// tmp-yaml c ConfigMap
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

	waitForCLI(t, []string{"list", "configmaps"}, "ci-int-cm", true)

	if out, err := exec.Command(cliBin(), "delete", "file", tmp, "--force").CombinedOutput(); err != nil {
		t.Fatalf("delete file: %v\n%s", err, out)
	}

	waitForCLI(t, []string{"list", "configmaps"}, "ci-int-cm", false)
}
