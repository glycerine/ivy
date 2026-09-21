package ivy2golang

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func TestGenerateTargetClassBuildsAsArchive(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "class", ClassName: "OnlyClass"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if out.Target != "class" || out.EffectiveTarget != "repl" || out.EmitMain {
		t.Fatalf("class target metadata = target %q effective %q emitMain %v", out.Target, out.EffectiveTarget, out.EmitMain)
	}
	if !strings.Contains(out.Source, "package ivygenerated") {
		t.Fatalf("class target should emit non-main package:\n%s", out.Source)
	}
	if strings.Contains(out.Source, "func main()") {
		t.Fatalf("class target should not emit main:\n%s", out.Source)
	}
	dir := t.TempDir()
	plan, err := BuildPlanFor(out, dir, out.Config)
	if err != nil {
		t.Fatalf("BuildPlanFor: %v", err)
	}
	if filepath.Ext(plan.OutputPath) != ".a" {
		t.Fatalf("class target should build an archive, got output %q", plan.OutputPath)
	}
	built, err := BuildOutput(out, dir)
	if err != nil {
		t.Fatalf("BuildOutput class target: %v\nsource:\n%s", err, out.Source)
	}
	if built != plan.OutputPath {
		t.Fatalf("BuildOutput path = %q, want %q", built, plan.OutputPath)
	}
	if _, err := os.Stat(built); err != nil {
		t.Fatalf("archive not written: %v", err)
	}
}
