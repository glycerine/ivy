package ivy2go

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// --- M9: test/gen target emission tests -----------------------------

func TestEmit_TestTarget_ProducesActionGenStruct(t *testing.T) {
	mod := compileIvySource(t, `
relation flag
action set_flag = {
	flag := true
}
`)
	out, err := Generate(mod, Config{Target: "test", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	actions := out.Files["actions.go"]
	if !strings.Contains(actions, "type actionGen_SetFlag struct") {
		t.Errorf("test target should emit actionGen_SetFlag struct, got:\n%s", actions)
	}
	if !strings.Contains(actions, "sol *goivy.Solver") {
		t.Errorf("actionGen should hold *goivy.Solver:\n%s", actions)
	}
	if !strings.Contains(actions, "func newactionGen_SetFlag(state *State)") {
		t.Errorf("actionGen needs constructor:\n%s", actions)
	}
	if !strings.Contains(actions, "func (g *actionGen_SetFlag) Generate(state *State)") {
		t.Errorf("actionGen needs Generate method:\n%s", actions)
	}
}

func TestEmit_TestTarget_ImportsGoivy(t *testing.T) {
	mod := compileIvySource(t, `
relation flag
action set_flag = {
	flag := true
}
`)
	out, err := Generate(mod, Config{Target: "test", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, where := range []string{"actions.go", "runtime.go"} {
		text := out.Files[where]
		if !strings.Contains(text, `"github.com/glycerine/ivy/goivy"`) {
			t.Errorf("%s should import goivy, got:\n%s", where, text)
		}
	}
}

func TestEmit_TestTarget_NewIvySolverHelper(t *testing.T) {
	mod := compileIvySource(t, `relation flag`)
	out, err := Generate(mod, Config{Target: "test", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	runtime := out.Files["runtime.go"]
	if !strings.Contains(runtime, "func newIvySolver() *goivy.Solver") {
		t.Errorf("runtime missing newIvySolver:\n%s", runtime)
	}
	if !strings.Contains(runtime, "goivy.NewSolver(nil, goivy.DefaultSolverOptions())") {
		t.Errorf("newIvySolver body wrong:\n%s", runtime)
	}
}

func TestEmit_ImplTarget_NoActionGenStruct(t *testing.T) {
	mod := compileIvySource(t, `
relation flag
action set_flag = {
	flag := true
}
`)
	out, err := Generate(mod, Config{Target: "impl", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	actions := out.Files["actions.go"]
	if strings.Contains(actions, "actionGen_") {
		t.Errorf("impl target should NOT emit actionGen_*, got:\n%s", actions)
	}
}

func TestEmit_GenTarget_AlsoEmitsActionGen(t *testing.T) {
	mod := compileIvySource(t, `
relation flag
action set_flag = {
	flag := true
}
`)
	out, err := Generate(mod, Config{Target: "gen", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	actions := out.Files["actions.go"]
	if !strings.Contains(actions, "actionGen_SetFlag") {
		t.Errorf("gen target should emit actionGen_SetFlag:\n%s", actions)
	}
}

func TestEmit_TestTarget_GeneratePicksInputsByCardinality(t *testing.T) {
	mod := compileIvySource(t, `
relation flag
action set_flag(b: bool) = {
	flag := b
}
`)
	out, err := Generate(mod, Config{Target: "test", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	actions := out.Files["actions.go"]
	if !strings.Contains(actions, "ivyChoose(2)") {
		t.Errorf("bool input should be picked via ivyChoose(2):\n%s", actions)
	}
}

func TestEmit_TestTarget_ActionGenHasClose(t *testing.T) {
	mod := compileIvySource(t, `
relation flag
action set_flag = {
	flag := true
}
`)
	out, err := Generate(mod, Config{Target: "test", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	actions := out.Files["actions.go"]
	if !strings.Contains(actions, "func (g *actionGen_SetFlag) Close() error") {
		t.Errorf("actionGen needs Close method:\n%s", actions)
	}
}

// --- M9: Tier 2 smoke — test target go-builds against goivy ---------

func TestSmoke_BuildEmittedTest(t *testing.T) {
	if !SlowGoTest {
		t.Skip("SLOW_GO_TEST not set")
	}
	mod := compileIvySource(t, `
relation flag
action set_flag = {
	flag := true
}
`)
	out, err := Generate(mod, Config{Target: "test", PackageName: "main"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	dir := t.TempDir()
	if err := WriteOutput(out, dir); err != nil {
		t.Fatalf("WriteOutput: %v", err)
	}
	pkgDir := outputDirectory(dir, out.BaseName)
	// Write a go.mod that maps the goivy import to our in-tree
	// goivy via a replace directive — the workspace files are next
	// to the test process's $GOPATH/src/github.com/glycerine/ivy/goivy.
	goivyAbs := repoRoot(t)
	gomod := "module ivygo_test_smoke\n\ngo 1.25\n\n" +
		"require github.com/glycerine/ivy/goivy v0.0.0\n\n" +
		"replace github.com/glycerine/ivy/goivy => " + goivyAbs + "\n"
	if err := os.WriteFile(filepath.Join(pkgDir, "go.mod"), []byte(gomod), 0o644); err != nil {
		t.Fatalf("go.mod: %v", err)
	}
	// Bring in a real go.sum so `go build` doesn't try to download.
	// Easiest: run `go mod tidy` first.
	tidy := exec.Command("go", "mod", "tidy")
	tidy.Dir = pkgDir
	if output, err := tidy.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy failed:\n%s", string(output))
	}
	cmd := exec.Command("go", "build", "-o", filepath.Join(pkgDir, "test_bin"), "./...")
	cmd.Dir = pkgDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build failed:\n%s", string(output))
	}
}

// repoRoot returns the absolute path to the goivy repo root, derived
// from the test binary's working directory (always ivy2go/).
func repoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	// wd is .../goivy/ivy2go; parent is goivy itself.
	return filepath.Dir(wd)
}
