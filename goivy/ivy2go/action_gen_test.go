package ivy2go

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy"
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

// --- OPEN 055: real Solver integration ------------------------------

func TestEmit_TestTarget_PushStateRunsRealIsSat(t *testing.T) {
	mod := compileIvySource(t, `relation flag`)
	out, err := Generate(mod, Config{Target: "test", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	runtime := out.Files["runtime.go"]
	if !strings.Contains(runtime, "sol.IsSat(") {
		t.Errorf("pushStateIntoSolver should call sol.IsSat, got:\n%s", runtime)
	}
	if !strings.Contains(runtime, "goivy.NewConst") {
		t.Errorf("pushStateIntoSolver should construct goivy expressions:\n%s", runtime)
	}
}

// --- OPEN 055.3: array-storage state facts + reified Pre tests ------

func TestEmit_TestTarget_ArrayStorageStateFacts(t *testing.T) {
	// Relation over a small finite domain → array storage; cells
	// should produce per-cell mkBoolFact assertions.
	mod := compileIvySource(t, `
type idx = {0..3}
relation slot(I: idx)
action set_slot = {
	slot(0) := true
}
`)
	out, err := Generate(mod, Config{Target: "test", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	runtime := out.Files["runtime.go"]
	if !strings.Contains(runtime, "Per-cell facts for array-storage symbol \"slot\"") {
		t.Errorf("array-storage slot should produce per-cell facts, got:\n%s", runtime)
	}
	if !strings.Contains(runtime, "for __i0 := 0; __i0 < 4; __i0++") {
		t.Errorf("array-storage loop bounds wrong, got:\n%s", runtime)
	}
	if !strings.Contains(runtime, "state.Slot[__i0]") {
		t.Errorf("per-cell access should index state.Slot, got:\n%s", runtime)
	}
}

func TestEmit_TestTarget_BuildPreconditionHelperPerAction(t *testing.T) {
	// Each action should get its own buildPrecondition_<Name>.
	mod := compileIvySource(t, `
relation flag
action set_flag(b: bool) = {
	require b;
	flag := b
}
action unset = {
	flag := false
}
`)
	out, err := Generate(mod, Config{Target: "test", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	actions := out.Files["actions.go"]
	if !strings.Contains(actions, "func buildPrecondition_SetFlag(state *State, __in0 *goivy.Const) *goivy.Clauses") {
		t.Errorf("buildPrecondition_SetFlag signature wrong, got:\n%s", actions)
	}
	if !strings.Contains(actions, "func buildPrecondition_Unset(state *State) *goivy.Clauses") {
		t.Errorf("buildPrecondition_Unset signature wrong, got:\n%s", actions)
	}
	if !strings.Contains(actions, "buildPrecondition_SetFlag(state, __in0)") {
		t.Errorf("Generate should call buildPrecondition_SetFlag, got:\n%s", actions)
	}
}

func TestEmit_TestTarget_ReifiedPreReferencesInputSym(t *testing.T) {
	// `require b` produces a Pre that mentions __fml:b; the reifier
	// must rewrite that reference to __in0.
	mod := compileIvySource(t, `
relation flag
action set_flag(b: bool) = {
	require b;
	flag := b
}
`)
	out, err := Generate(mod, Config{Target: "test", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	actions := out.Files["actions.go"]
	if !strings.Contains(actions, "extra = append(extra,") {
		t.Errorf("Pre fmlas should land in extra, got:\n%s", actions)
	}
	if !strings.Contains(actions, "&goivy.LogicNot{Body: __in0}") {
		t.Errorf("require b should reify as LogicNot{__in0}, got:\n%s", actions)
	}
}

func TestReifyExprAsGoCode_BasicShapes(t *testing.T) {
	g := newExprGen(t, "")
	// Plain Const → goivy.NewConst("name", goivy.Boolean).
	c := &goivy.Const{Name: "p", CSort: goivy.Boolean}
	got, ok := g.reifyExprAsGoCode(c, nil)
	if !ok || got != `goivy.NewConst("p", goivy.Boolean)` {
		t.Errorf("Const reify = (%q, %v)", got, ok)
	}
	// LogicNot wraps the body.
	got, ok = g.reifyExprAsGoCode(&goivy.LogicNot{Body: c}, nil)
	if !ok || !strings.Contains(got, "goivy.LogicNot{Body:") {
		t.Errorf("LogicNot reify = (%q, %v)", got, ok)
	}
	// LogicAnd of two consts.
	d := &goivy.Const{Name: "q", CSort: goivy.Boolean}
	got, ok = g.reifyExprAsGoCode(&goivy.LogicAnd{Terms: []goivy.Expr{c, d}}, nil)
	if !ok || !strings.Contains(got, "goivy.LogicAnd{Terms:") {
		t.Errorf("LogicAnd reify = (%q, %v)", got, ok)
	}
}

func TestReifyExprAsGoCode_ParamRewriteToInput(t *testing.T) {
	g := newExprGen(t, "")
	p := &goivy.Const{Name: "b", CSort: goivy.Boolean}
	// Reference to b should rewrite to __in0.
	ref := &goivy.Const{Name: "b", CSort: goivy.Boolean}
	got, ok := g.reifyExprAsGoCode(ref, []*goivy.Const{p})
	if !ok || got != "__in0" {
		t.Errorf("param rewrite = (%q, %v), want __in0", got, ok)
	}
	// __fml:b prefix variant also rewrites.
	ref2 := &goivy.Const{Name: "__fml:b", CSort: goivy.Boolean}
	got, ok = g.reifyExprAsGoCode(ref2, []*goivy.Const{p})
	if !ok || got != "__in0" {
		t.Errorf("__fml: param rewrite = (%q, %v), want __in0", got, ok)
	}
}

// --- OPEN 055.2: state-fact precondition tests ----------------------

func TestEmit_TestTarget_StateFactsClausesEmitted(t *testing.T) {
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
	runtime := out.Files["runtime.go"]
	if !strings.Contains(runtime, "func stateFactsAsClauses(state *State) *goivy.Clauses") {
		t.Errorf("stateFactsAsClauses helper missing, got:\n%s", runtime)
	}
	if !strings.Contains(runtime, `mkBoolFact("flag", state.Flag)`) {
		t.Errorf("scalar bool symbol should be in state facts, got:\n%s", runtime)
	}
	if !strings.Contains(runtime, "func mkBoolFact(name string, val bool) goivy.Expr") {
		t.Errorf("mkBoolFact helper missing, got:\n%s", runtime)
	}
}

func TestEmit_TestTarget_GenerateBuildsPreconditionFromState(t *testing.T) {
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
	if !strings.Contains(actions, "stateFactsAsClauses(state)") {
		t.Errorf("Generate should seed precondition with state facts, got:\n%s", actions)
	}
}

func TestEmit_TestTarget_GenerateUsesGetModelClauses(t *testing.T) {
	// OPEN 055.1: Generate now performs a real solver round-trip
	// via GetModelClauses + per-input model extraction.
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
	if !strings.Contains(actions, "g.sol.GetModelClauses(") {
		t.Errorf("Generate should call GetModelClauses, got:\n%s", actions)
	}
	if !strings.Contains(actions, "goivy.NewConst(") {
		t.Errorf("Generate should construct input goivy.Const symbols, got:\n%s", actions)
	}
	runtime := out.Files["runtime.go"]
	if !strings.Contains(runtime, "func pickBoolOrChoose") {
		t.Errorf("runtime should emit pickBoolOrChoose helper, got:\n%s", runtime)
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
	// Post-OPEN-055.1: input picking goes through pickBoolOrChoose
	// (which itself falls back to ivyChoose when the model can't
	// supply a value).
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
	if !strings.Contains(actions, "pickBoolOrChoose(modelResult,") {
		t.Errorf("bool input should be picked via pickBoolOrChoose, got:\n%s", actions)
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
