package ivy2golang

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func compileGeneratedGo(t *testing.T, out *Output) string {
	t.Helper()
	if out == nil || out.Source == "" {
		t.Fatalf("nil or empty generated output")
	}
	dir := t.TempDir()
	bin, err := BuildOutput(out, dir)
	if err != nil {
		srcPath := filepath.Join(dir, out.BaseName+".go")
		t.Fatalf("compile generated Go: %v\nsource: %s", err, srcPath)
	}
	if bin == "" {
		t.Fatalf("empty binary path")
	}
	return bin
}

func runBinary(t *testing.T, bin string, args ...string) (string, string, error) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func TestBuildPlanForUsesGoFile(t *testing.T) {
	out := &Output{BaseName: "x", Source: "package main\nfunc main(){}\n"}
	dir := t.TempDir()
	plan, err := BuildPlanFor(out, dir, Config{})
	if err != nil {
		t.Fatalf("BuildPlanFor: %v", err)
	}
	if plan.GoFile != filepath.Join(dir, "x.go") {
		t.Fatalf("GoFile=%q", plan.GoFile)
	}
	if _, err := os.Stat(filepath.Join(dir, ".ivy2golang-gocache")); err != nil {
		t.Fatalf("gocache not created: %v", err)
	}
}
