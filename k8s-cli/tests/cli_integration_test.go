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

/*──────────────────────── helpers ───────────────────────*/

func cliBin() string { return filepath.FromSlash("../bin/k8s-cli") }

func ensureBinary(t *testing.T) {
	if _, err := os.Stat(cliBin()); err == nil {
		return
	}
	out, err := exec.Command("go", "build", "-o", cliBin(), "../main.go").CombinedOutput()
	if err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
}

// быстрый smoke-тест: кластер жив?
func skipIfNoCluster(t *testing.T) {
	if err := exec.Command("kubectl", "--request-timeout=5s", "cluster-info").Run(); err != nil {
		t.Skip("cluster is absent, skipping integration tests")
	}
}

// ждём появления/исчезновения строки `name` в выводе k8s-cli list …
func waitForName(t *testing.T, listArgs []string, name string, wantPresent bool) {
	for i := 0; i < 30; i++ { // ≤ 60 с
		out, _ := exec.Command(cliBin(), listArgs...).CombinedOutput()
		has := bytes.Contains(out, []byte(name))
		if has == wantPresent {
			return
		}
		time.Sleep(2 * time.Second)
	}
	state := "appear"
	if !wantPresent {
		state = "disappear"
	}
	t.Fatalf("resource %q did not %s via %v", name, state, listArgs)
}

/*──────────────────── basic installation ───────────────────*/

func TestInstallationCommands(t *testing.T) {
	ensureBinary(t)

	// k8s-cli --help
	if out, err := exec.Command(cliBin(), "--help").CombinedOutput(); err != nil {
		t.Fatalf("--help failed: %v\n%s", err, out)
	}

	// k8s-cli context current (должен показать текущий контекст Kind)
	out, err := exec.Command(cliBin(), "context", "current").CombinedOutput()
	if err != nil {
		t.Fatalf("context current: %v\n%s", err, out)
	}
	if !bytes.Contains(out, []byte("kind-")) {
		t.Fatalf("unexpected context: %s", out)
	}
}

/*──────────────────── verify step-6 requirements ───────────*/

func TestListVariants(t *testing.T) {
	skipIfNoCluster(t)
	ensureBinary(t)

	// 1. k8s-cli list deployments
	_, err := exec.Command(cliBin(), "list", "deployments").CombinedOutput()
	if err != nil {
		t.Fatalf("list deployments: %v", err)
	}

	// 2. k8s-cli --kubeconfig=<path> list deployments
	kc := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	_, err = exec.Command(cliBin(), "--kubeconfig="+kc, "list", "deployments").CombinedOutput()
	if err != nil {
		t.Fatalf("--kubeconfig list deployments: %v", err)
	}

	// 3. k8s-cli list deployments -n default -o json
	_, err = exec.Command(cliBin(), "list", "deployments", "-n", "default", "-o", "json").CombinedOutput()
	if err != nil {
		t.Fatalf("list deployments json: %v", err)
	}
}

/*──────────────────── full functionality ───────────────────*/

// 1. apply / delete YAML-ы из examples/
func TestApplyExampleYamls(t *testing.T) {
	skipIfNoCluster(t)
	ensureBinary(t)

	files := []string{
		"deployment.yaml",
		"pod.yaml",
		"service.yaml",
	}
	kindRe := regexp.MustCompile(`(?m)^kind:\s*(\S+)`)

	for _, f := range files {
		fp := filepath.FromSlash("../examples/" + f)
		raw, _ := os.ReadFile(fp)
		m := kindRe.FindSubmatch(raw)
		if m == nil {
			t.Fatalf("cannot detect kind in %s", fp)
		}
		kind := strings.ToLower(string(m[1])) // deployment/pod/service
		name := strings.TrimSuffix(f, ".yaml")
		list := []string{"list", kind + "s"}

		// apply
		if out, err := exec.Command(cliBin(), "apply", "file", fp).CombinedOutput(); err != nil {
			t.Fatalf("apply %s: %v\n%s", fp, err, out)
		}
		waitForName(t, list, name, true)

		// delete
		if out, err := exec.Command(cliBin(), "delete", "file", fp, "--force").CombinedOutput(); err != nil {
			t.Fatalf("delete %s: %v\n%s", fp, err, out)
		}
		waitForName(t, list, name, false)
	}
}

// 2. imperative create / delete
func TestImperativeCreateDelete(t *testing.T) {
	skipIfNoCluster(t)
	ensureBinary(t)

	name := "test"
	list := []string{"list", "deployments"}

	// k8s-cli create deployment test --image=nginx:1.20 --replicas=2 -n nginx
	if out, err := exec.Command(cliBin(), "create", "deployment", name,
		"--image=nginx:1.20", "--replicas=2", "-n", "default").CombinedOutput(); err != nil {
		t.Fatalf("create deployment: %v\n%s", err, out)
	}
	waitForName(t, append(list, "-n", "default"), name, true)

	// k8s-cli delete deployment test -n nginx
	if out, err := exec.Command(cliBin(), "delete", "deployment", name, "-n", "default").
		CombinedOutput(); err != nil {
		t.Fatalf("delete deployment: %v\n%s", err, out)
	}
	waitForName(t, append(list, "-n", "default"), name, false)
}
