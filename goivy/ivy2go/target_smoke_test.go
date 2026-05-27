package ivy2go

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestSmoke_TestRuntimeOptionsRunsOutAndModelFile(t *testing.T) {
	if !SlowGoTest {
		t.Skip("SLOW_GO_TEST not set")
	}
	mod := compileIvySource(t, `
action step = {}
export step
`)
	out, err := Generate(mod, Config{Target: "test", PackageName: "main", TestIters: "0", TestRuns: "1"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	dir := playpenDir(t)
	if err := WriteOutput(out, dir); err != nil {
		t.Fatalf("WriteOutput: %v", err)
	}
	pkgDir := outputDirectory(dir, out.BaseName)
	binName := "test_runtime_options_bin"
	build := exec.Command("go", "build", "-o", binName, ".")
	build.Dir = pkgDir
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build failed:\n%s", string(output))
	}
	outPath := filepath.Join(dir, "trace.out")
	modelPath := filepath.Join(dir, "model.log")
	run := exec.Command(filepath.Join(pkgDir, binName),
		"iters=0",
		"runs=2",
		"seed=9",
		"delay=0",
		"wait=0",
		"out="+outPath,
		"modelfile="+modelPath,
	)
	stdout, err := run.CombinedOutput()
	if err != nil {
		t.Fatalf("generated binary failed: %v\n%s", err, string(stdout))
	}
	if strings.TrimSpace(string(stdout)) != "" {
		t.Fatalf("out= should redirect test output away from stdout, got:\n%s", string(stdout))
	}
	trace, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read out file: %v", err)
	}
	if got, want := strings.Count(string(trace), "test_completed"), 2; got != want {
		t.Fatalf("runs=2 should emit two completions to out file, got %d:\n%s", got, string(trace))
	}
	if _, err := os.Stat(modelPath); err != nil {
		t.Fatalf("modelfile= should create model log file: %v", err)
	}
}
