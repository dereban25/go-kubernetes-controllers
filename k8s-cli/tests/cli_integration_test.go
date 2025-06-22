package tests

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

/* -------------------------------------------------------------------------- */
/*                               helper-functions                             */
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

// Быстрый smoke-тест, что кластер жив
func skipIfNoCluster(t *testing.T) {
	if err := exec.Command("kubectl", "--request-timeout=5s", "cluster-info").Run(); err != nil {
		t.Skip("cluster is absent, skipping integration tests")
	}
}

// map "Deployment" → "deployments" и т. д.
func plural(kind string) (string, bool) {
	switch strings.ToLower(kind) {
	case "deployment":
		return "deployments", true
	case "pod":
		return "pods", true
	case "service":
		return "services", true
	default:
		return "", false
	}
}

var kindRe = regexp.MustCompile(`(?m)^kind:\s*(\S+)`)

func yamlKind(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	m := kindRe.FindSubmatch(data)
	if m == nil {
		return "", nil
	}
	return string(m[1]), nil
}

func waitUntil(t *testing.T, listArgs []string, wantChange bool) {
	before, _ := exec.Command(cliBin(), listArgs...).CombinedOutput()

	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Second)
		after, _ := exec.Command(cliBin(), listArgs...).CombinedOutput()
		changed := !bytes.Equal(bytes.TrimSpace(before), bytes.TrimSpace(after))
		if changed == wantChange {
			return
		}
	}
	state := "change"
	if !wantChange {
		state = "restore"
	}
	t.Fatalf("list %v did not %s", listArgs, state)
}

/* -------------------------------------------------------------------------- */
/*                               integration-tests                            */
/* -------------------------------------------------------------------------- */

func TestApplyExampleYamls(t *testing.T) {
	skipIfNoCluster(t)
	ensureBinary(t)

	examples := []string{
		filepath.FromSlash("../examples/deployment.yaml"),
		filepath.FromSlash("../examples/pod.yaml"),
		filepath.FromSlash("../examples/service.yaml"),
	}

	for _, fp := range examples {
		kind, err := yamlKind(fp)
		if err != nil || kind == "" {
			t.Fatalf("cannot detect kind in %s: %v", fp, err)
		}
		plur, ok := plural(kind)
		if !ok {
			t.Skipf("kind %s not handled – skipping", kind)
		}

		listArgs := []string{"list", plur}

		// apply file
		if out, err := exec.Command(cliBin(), "apply", "file", fp).CombinedOutput(); err != nil {
			t.Fatalf("apply %s: %v\n%s", fp, err, out)
		}

		waitUntil(t, listArgs, true)

		// delete file
		if out, err := exec.Command(cliBin(), "delete", "file", fp, "--force").CombinedOutput(); err != nil {
			t.Fatalf("delete %s: %v\n%s", fp, err, out)
		}

		waitUntil(t, listArgs, false)
	}
}

func TestCreateAndDeleteDeploymentCLI(t *testing.T) {
	skipIfNoCluster(t)
	ensureBinary(t)

	name := "int-deploy-cli"
	image := "nginx:1.25.5"
	listArgs := []string{"list", "deployments"}

	// create deployment
	if out, err := exec.Command(cliBin(), "create", "deployment", name, "--image="+image, "--replicas=1").CombinedOutput(); err != nil {
		t.Fatalf("create deployment: %v\n%s", err, out)
	}
	waitUntil(t, listArgs, true)

	// delete deployment
	if out, err := exec.Command(cliBin(), "delete", "deployment", name, "--force").CombinedOutput(); err != nil {
		t.Fatalf("delete deployment: %v\n%s", err, out)
	}
	waitUntil(t, listArgs, false)
}
