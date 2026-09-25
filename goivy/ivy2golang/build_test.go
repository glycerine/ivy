package ivy2golang

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
)

var (
	generatedBuildTestEnvOnce sync.Once
	generatedBuildTestEnv     []string
	generatedBuildTestRoot    string
	generatedBuildTestEnvErr  error
	generatedBinaryRootOnce   sync.Once
	generatedBinaryRoot       string
	generatedBinaryRootErr    error
	generatedBinaryBuilds     sync.Map
	slowTestParallelized      sync.Map
)

func TestMain(m *testing.M) {
	configureSlowTestParallelism()
	code := m.Run()
	if generatedBuildTestRoot != "" {
		_ = os.RemoveAll(generatedBuildTestRoot)
	}
	if generatedBinaryRoot != "" && generatedBinaryRoot != generatedBuildTestRoot {
		_ = os.RemoveAll(generatedBinaryRoot)
	}
	os.Exit(code)
}

func configureSlowTestParallelism() {
	if os.Getenv("SLOWTEST") != "1" {
		return
	}
	explicit := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "test.parallel" {
			explicit = true
		}
	})
	if explicit {
		return
	}
	limit := os.Getenv("IVY2GOLANG_SLOWTEST_PARALLEL")
	if limit == "" {
		n := runtime.GOMAXPROCS(0)
		if n < 4 {
			n = 4
		}
		if n > 9 {
			n = 9
		}
		limit = strconv.Itoa(n)
	}
	_ = flag.Set("test.parallel", limit)
}

func requireSlowTest(t *testing.T) {
	t.Helper()
	if os.Getenv("SLOWTEST") != "1" {
		t.Skip("set SLOWTEST=1 to run generated binary build/run integration checks")
	}
	if strings.Contains(t.Name(), "/") {
		return
	}
	if _, loaded := slowTestParallelized.LoadOrStore(t.Name(), true); !loaded {
		t.Parallel()
	}
}

func compileGeneratedGo(t *testing.T, out *Output) string {
	t.Helper()
	requireSlowTest(t)
	if out == nil || out.Source == "" {
		t.Fatalf("nil or empty generated output")
	}
	dir := t.TempDir()
	if err := WriteOutput(out, dir); err != nil {
		t.Fatalf("write generated Go: %v", err)
	}
	srcPath := filepath.Join(dir, goSourceFileName(out.BaseName))
	if err := formatGoOutputFile(srcPath); err != nil {
		t.Fatalf("format generated Go: %v\nsource: %s", err, srcPath)
	}
	validateGeneratedGoForTest(t, srcPath)
	return srcPath
}

func validateGeneratedGoSourceForTest(t *testing.T, out *Output) {
	t.Helper()
	if out == nil || out.Source == "" {
		t.Fatalf("nil or empty generated output")
	}
	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, goSourceFileName(out.BaseName), out.Source, 0); err != nil {
		t.Fatalf("parse generated Go: %v\nsource:\n%s", err, out.Source)
	}
	if line := firstBareUndeclaredSyntheticTempAssignment(out.Source); line != "" {
		t.Fatalf("generated Go assigns synthetic temp before declaration: %s\nsource:\n%s", line, out.Source)
	}
	if unused := unreadDeclaredSyntheticTemps(out.Source); len(unused) != 0 {
		t.Fatalf("generated Go declares unread synthetic temps: %v\nsource:\n%s", unused, out.Source)
	}
}

func validateGeneratedGoForTest(t *testing.T, srcPath string) {
	t.Helper()
	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, srcPath, nil, 0); err != nil {
		t.Fatalf("parse generated Go: %v\nsource: %s", err, srcPath)
	}
	data, err := os.ReadFile(srcPath)
	if err != nil {
		t.Fatalf("read generated Go: %v\nsource: %s", err, srcPath)
	}
	src := string(data)
	if line := firstBareUndeclaredSyntheticTempAssignment(src); line != "" {
		t.Fatalf("generated Go assigns synthetic temp before declaration: %s\nsource: %s", line, srcPath)
	}
	if unused := unreadDeclaredSyntheticTemps(src); len(unused) != 0 {
		t.Fatalf("generated Go declares unread synthetic temps: %v\nsource: %s", unused, srcPath)
	}
}

func buildOutputForTest(t *testing.T, out *Output, dir string) (string, error) {
	t.Helper()
	return buildOutputWithEnv(out, dir, generatedBuildEnvForTest(t))
}

func generatedBuildEnvForTest(t *testing.T) []string {
	t.Helper()
	generatedBuildTestEnvOnce.Do(func() {
		cacheDir := os.Getenv("GOCACHE")
		tmpDir := os.Getenv("GOTMPDIR")
		if cacheDir == "" || tmpDir == "" {
			parent := os.Getenv("GOTMPDIR")
			if parent == "" {
				parent = os.TempDir()
			}
			generatedBuildTestRoot, generatedBuildTestEnvErr = os.MkdirTemp(parent, "ivy2golang-generated-build-*")
			if generatedBuildTestEnvErr != nil {
				return
			}
			if cacheDir == "" {
				cacheDir = filepath.Join(generatedBuildTestRoot, "gocache")
			}
			if tmpDir == "" {
				tmpDir = filepath.Join(generatedBuildTestRoot, "gotmp")
			}
		}
		for _, dir := range []string{cacheDir, tmpDir} {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				generatedBuildTestEnvErr = err
				return
			}
		}
		generatedBuildTestEnv = []string{"GOCACHE=" + cacheDir, "GOTMPDIR=" + tmpDir}
	})
	if generatedBuildTestEnvErr != nil {
		t.Fatalf("prepare shared generated-build cache: %v", generatedBuildTestEnvErr)
	}
	return generatedBuildTestEnv
}

func runBinary(t *testing.T, bin string, args ...string) (string, string, error) {
	t.Helper()
	requireSlowTest(t)
	bin = ensureGeneratedBinaryForTest(t, bin)
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
	bin = ensureGeneratedBinaryForTest(t, bin)
	cmd := exec.Command(bin, args...)
	cmd.Stdin = strings.NewReader(input)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func ensureGeneratedBinaryForTest(t *testing.T, path string) string {
	t.Helper()
	if !strings.HasSuffix(path, ".go") {
		return path
	}
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read generated Go before build: %v\nsource: %s", err, path)
	}
	sum := sha256.Sum256(src)
	key := runtime.GOOS + "-" + runtime.GOARCH + "-" + hex.EncodeToString(sum[:])
	build := generatedBinaryBuildForTest(t, key)
	build.once.Do(func() {
		bin := filepath.Join(generatedBinaryRootForTest(t), "generated-"+key)
		if runtime.GOOS == "windows" {
			bin += ".exe"
		}
		build.path = bin
		if st, err := os.Stat(bin); err == nil && st.Mode().IsRegular() {
			return
		}
		cmd := exec.Command("go", "build", "-o", bin, path)
		cmd.Env = append(os.Environ(), generatedBuildEnvForTest(t)...)
		cmd.Dir = moduleRootForGeneratedBuild()
		build.output, build.err = cmd.CombinedOutput()
	})
	if build.err != nil {
		t.Fatalf("build generated Go for run: %v\n%s\nsource: %s", build.err, string(build.output), path)
	}
	if build.path == "" {
		t.Fatalf("build generated Go produced empty binary path\nsource: %s", path)
	}
	return build.path
}

type generatedBinaryBuild struct {
	once   sync.Once
	path   string
	output []byte
	err    error
}

func generatedBinaryBuildForTest(t *testing.T, key string) *generatedBinaryBuild {
	t.Helper()
	value, _ := generatedBinaryBuilds.LoadOrStore(key, &generatedBinaryBuild{})
	return value.(*generatedBinaryBuild)
}

func generatedBinaryRootForTest(t *testing.T) string {
	t.Helper()
	generatedBinaryRootOnce.Do(func() {
		parent := os.Getenv("GOTMPDIR")
		if parent == "" {
			parent = os.TempDir()
		}
		generatedBinaryRoot, generatedBinaryRootErr = os.MkdirTemp(parent, "ivy2golang-generated-bin-*")
	})
	if generatedBinaryRootErr != nil {
		t.Fatalf("prepare shared generated binary cache: %v", generatedBinaryRootErr)
	}
	return generatedBinaryRoot
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
		if strings.HasPrefix(env, "XTRACE_OFF=") {
			continue
		}
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 || !filepath.IsAbs(parts[1]) {
			t.Fatalf("build env path should be absolute, got %q in %#v", env, plan.Env)
		}
	}
}

func TestBuildPlanForUsesModuleRootWorkDirFast(t *testing.T) {
	out := &Output{BaseName: "x", Source: "package main\nfunc main(){}\n"}
	plan, err := BuildPlanFor(out, t.TempDir(), Config{})
	if err != nil {
		t.Fatalf("BuildPlanFor: %v", err)
	}
	root := moduleRootForGeneratedBuild()
	if root == "" {
		t.Skip("module source root not available in this build")
	}
	if plan.WorkDir != root {
		t.Fatalf("WorkDir=%q, want module root %q", plan.WorkDir, root)
	}
	if _, err := os.Stat(filepath.Join(plan.WorkDir, "go.mod")); err != nil {
		t.Fatalf("WorkDir should contain go.mod: %v", err)
	}
}

func TestBuildPlanForGeneratedBuildCacheEnvOverrideFast(t *testing.T) {
	dir := t.TempDir()
	cacheDir := filepath.Join(dir, "shared-cache")
	tmpDir := filepath.Join(dir, "shared-tmp")
	t.Setenv("IVY2GOLANG_GOCACHE", cacheDir)
	t.Setenv("IVY2GOLANG_GOTMPDIR", tmpDir)
	out := &Output{BaseName: "x", Source: "package main\nfunc main(){}\n"}
	plan, err := BuildPlanFor(out, filepath.Join(dir, "out"), Config{})
	if err != nil {
		t.Fatalf("BuildPlanFor: %v", err)
	}
	wantEnv := []string{"GOCACHE=" + cacheDir, "GOTMPDIR=" + tmpDir}
	for _, want := range wantEnv {
		if !containsString(plan.Env, want) {
			t.Fatalf("BuildPlanFor env missing %q in %#v", want, plan.Env)
		}
	}
	for _, want := range []string{cacheDir, tmpDir} {
		if _, err := os.Stat(want); err != nil {
			t.Fatalf("expected generated build dir %s: %v", want, err)
		}
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
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
	built, err := buildOutputForTest(t, out, dir)
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
