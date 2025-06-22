// tests/cli_integration_test.go
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

func skipIfNoCluster(t *testing.T) {
	if err := exec.Command("kubectl", "--request-timeout=5s", "cluster-info").Run(); err != nil {
		t.Skip("cluster is absent, skipping integration tests")
	}
}

func waitForName(t *testing.T, listArgs []string, name string, wantPresent bool) {
	for i := 0; i < 30; i++ { // 30 × 2 сек = ≤ 60 сек
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

/*──────────────────────── basic commands ─────────────────────*/

func TestInstallationCommands(t *testing.T) {
	ensureBinary(t)

	if _, err := exec.Command(cliBin(), "--help").CombinedOutput(); err != nil {
		t.Fatalf("--help failed: %v", err)
	}

	if out, err := exec.Command(cliBin(), "context", "current").CombinedOutput(); err != nil {
		t.Fatalf("context current: %v\n%s", err, out)
	} else if !bytes.Contains(out, []byte("kind-")) {
		t.Fatalf("unexpected context: %s", out)
	}
}

/*──────────────────── verify list requirements ──────────────*/

func TestListVariants(t *testing.T) {
	skipIfNoCluster(t)
	ensureBinary(t)

	if _, err := exec.Command(cliBin(), "list", "deployments").CombinedOutput(); err != nil {
		t.Fatalf("list deployments: %v", err)
	}

	kc := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	if _, err := exec.Command(cliBin(), "--kubeconfig="+kc, "list", "deployments").CombinedOutput(); err != nil {
		t.Fatalf("--kubeconfig list deployments: %v", err)
	}

	// 3) namespace + json-вывод
	if _, err := exec.Command(cliBin(), "list", "deployments", "-n", "default", "-o", "json").CombinedOutput(); err != nil {
		t.Fatalf("list -o json: %v", err)
	}
}

/*─────────────────── apply / delete YAML из examples ────────*/

func TestApplyExampleYamls(t *testing.T) {
	skipIfNoCluster(t)
	ensureBinary(t)

	kindRE := regexp.MustCompile(`(?m)^kind:\s*(\S+)`)
	for _, file := range []string{"deployment.yaml", "pod.yaml", "service.yaml"} {
		fp := filepath.FromSlash("../examples/" + file)
		raw, _ := os.ReadFile(fp)
		m := kindRE.FindSubmatch(raw)
		if m == nil {
			t.Fatalf("cannot detect kind in %s", fp)
		}
		kind := strings.ToLower(string(m[1])) // deployment / pod / service
		name := strings.TrimSuffix(file, ".yaml")
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

/*─────────────────── imperative create / delete ─────────────*/

func TestImperativeCreateDelete(t *testing.T) {
	skipIfNoCluster(t)
	ensureBinary(t)

	ns := "default"
	name := "test"
	img := "nginx:1.20"
	list := []string{"list", "deployments", "-n", ns}

	// run helper
	run := func(args ...string) {
		cmd := exec.Command(cliBin(), args...)
		out, err := cmd.CombinedOutput()
		t.Logf("↪️  k8s-cli %s\n%s", strings.Join(args, " "), out)
		if err != nil {
			t.Fatalf("command failed: %v", err)
		}
	}

	// create deployment
	run("create", "deployment", name, "--image="+img, "--replicas=2", "-n", ns)
	waitForName(t, list, name, true)

	run("delete", "deployment", name, "-n", ns, "--force")

	waitForName(t, list, name, false)
}
