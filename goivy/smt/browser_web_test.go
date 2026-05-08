//go:build web && !tinygo && !wasip1 && (darwin || linux || windows)

package smt

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestSolverBoolRoundTrip(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not resolve smt test path")
	}
	goivyDir := filepath.Dir(filepath.Dir(file))
	webvueDir := filepath.Join(goivyDir, "webvue")
	staticDir := filepath.Join(webvueDir, "static")
	roundTripWasm := filepath.Join(staticDir, "smt-wasip1-roundtrip.wasm")

	requireFile(t, filepath.Join(staticDir, "z3-471-api.js"), "run make z3-wasm-api")
	requireFile(t, filepath.Join(staticDir, "z3-471-api.wasm"), "run make z3-wasm-api")
	requireFile(t, filepath.Join(webvueDir, "node_modules", ".bin", playwrightBin()), "run make webvue-setup")

	runCommand(t, goivyDir, []string{
		"GOCACHE=/private/tmp/go-build",
		"GOOS=wasip1",
		"GOARCH=wasm",
	}, "go", "build", "-o", roundTripWasm, "./smt/testdata/wasip1_roundtrip")

	runCommand(t, webvueDir, nil,
		filepath.Join(".", "node_modules", ".bin", playwrightBin()),
		"test", "--config", "playwright.config.mjs", "tests/smtWasip1Worker.browser.spec.js",
	)
}

func playwrightBin() string {
	if runtime.GOOS == "windows" {
		return "playwright.cmd"
	}
	return "playwright"
}

func requireFile(t *testing.T, path, hint string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("missing required file %s: %v; %s", path, err, hint)
	}
}

func runCommand(t *testing.T, dir string, env []string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		t.Fatalf("%s %v failed in %s: %v\n%s", name, args, dir, err, out.String())
	}
	if out.Len() > 0 {
		t.Logf("%s %v output:\n%s", name, args, out.String())
	}
}
