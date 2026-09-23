package ivy2golang

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func requireSlowTest(t *testing.T) {
	t.Helper()
	if os.Getenv("SLOWTEST") != "1" {
		t.Skip("set SLOWTEST=1 to run generated binary build/run integration checks")
	}
}

func compileGeneratedGo(t *testing.T, out *Output) string {
	t.Helper()
	requireSlowTest(t)
	if out == nil || out.Source == "" {
		t.Fatalf("nil or empty generated output")
	}
	dir := t.TempDir()
	bin, err := BuildOutput(out, dir)
	if err != nil {
		srcPath := filepath.Join(dir, goSourceFileName(out.BaseName))
		t.Fatalf("compile generated Go: %v\nsource: %s", err, srcPath)
	}
	if bin == "" {
		t.Fatalf("empty binary path")
	}
	return bin
}

func runBinary(t *testing.T, bin string, args ...string) (string, string, error) {
	t.Helper()
	requireSlowTest(t)
	cmd := exec.Command(bin, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func runBinaryWithInput(t *testing.T, bin, input string, args ...string) (string, string, error) {
	t.Helper()
	requireSlowTest(t)
	cmd := exec.Command(bin, args...)
	cmd.Stdin = strings.NewReader(input)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func TestFormatGoOutputFileFast(t *testing.T) {
	path := filepath.Join(t.TempDir(), "x.go")
	const unformatted = "package main\nfunc main(){println(\"x\")}\n"
	if err := os.WriteFile(path, []byte(unformatted), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	if err := formatGoOutputFile(path); err != nil {
		t.Fatalf("formatGoOutputFile: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read source: %v", err)
	}
	if string(data) == unformatted || !strings.Contains(string(data), "func main() { println(\"x\") }\n") {
		t.Fatalf("generated Go was not gofmt-formatted:\n%s", data)
	}
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

func TestBuildPlanForTestBasenameAvoidsGoTestFile(t *testing.T) {
	out := &Output{BaseName: "table_test", Source: "package main\nfunc main(){}\n"}
	dir := t.TempDir()
	plan, err := BuildPlanFor(out, dir, Config{})
	if err != nil {
		t.Fatalf("BuildPlanFor: %v", err)
	}
	if plan.GoFile != filepath.Join(dir, "table_test_main.go") {
		t.Fatalf("GoFile=%q", plan.GoFile)
	}
	if plan.OutputPath != filepath.Join(dir, "table_test") {
		t.Fatalf("OutputPath=%q", plan.OutputPath)
	}
	if err := WriteOutput(out, dir); err != nil {
		t.Fatalf("WriteOutput: %v", err)
	}
	if _, err := os.Stat(plan.GoFile); err != nil {
		t.Fatalf("expected generated source at %s: %v", plan.GoFile, err)
	}
}

func TestBuildPlanForRelativeDirUsesAbsoluteGoCache(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	out := &Output{BaseName: "x", Source: "package main\nfunc main(){}\n"}
	plan, err := BuildPlanFor(out, "", Config{})
	if err != nil {
		t.Fatalf("BuildPlanFor: %v", err)
	}
	if !filepath.IsAbs(plan.GoFile) || !filepath.IsAbs(plan.OutputPath) {
		t.Fatalf("build plan paths should be absolute: %#v", plan)
	}
	for _, env := range plan.Env {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 || !filepath.IsAbs(parts[1]) {
			t.Fatalf("build env path should be absolute, got %q in %#v", env, plan.Env)
		}
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
	if strings.Contains(out.Source, `"bufio"`) {
		t.Fatalf("class target should not import repl-only scanner package:\n%s", out.Source)
	}
	dir := t.TempDir()
	plan, err := BuildPlanFor(out, dir, out.Config)
	if err != nil {
		t.Fatalf("BuildPlanFor: %v", err)
	}
	if filepath.Ext(plan.OutputPath) != ".a" {
		t.Fatalf("class target should build an archive, got output %q", plan.OutputPath)
	}
	requireSlowTest(t)
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

func TestBuildPlanForImplMirrorsIvy2CppExecutableLinkAttempt(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "impl", ClassName: "ImplPlan"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if out.Target != "impl" || !out.EmitMain || strings.Contains(out.Source, "func main()") {
		t.Fatalf("impl metadata/source should mirror ivy2cpp's no-driver implementation output: target=%q emitMain=%v\n%s", out.Target, out.EmitMain, out.Source)
	}
	plan, err := BuildPlanFor(out, t.TempDir(), out.Config)
	if err != nil {
		t.Fatalf("BuildPlanFor: %v", err)
	}
	if filepath.Ext(plan.OutputPath) == ".a" {
		t.Fatalf("target=impl build should not silently become a Go archive; ivy2cpp build=true impl attempts an executable link: %#v", plan)
	}
	for _, arg := range plan.Args {
		if arg == "-buildmode=archive" {
			t.Fatalf("target=impl build args should not force archive mode: %#v", plan.Args)
		}
	}
}
