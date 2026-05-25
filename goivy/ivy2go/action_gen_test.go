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

// --- OPEN 055.7: solver-driven struct-input synthesis ---------------

func TestEmit_TestTarget_StructParamSynthesisShape(t *testing.T) {
	mod := compileIvySource(t, `
type point = struct { x : bool, y : bool }
relation flag
action set_to_x(p: point) = {
	require x(p);
	flag := x(p)
}
`)
	out, err := Generate(mod, Config{Target: "test", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	actions := out.Files["actions.go"]

	// Per-field input symbols declared in Generate.
	for _, want := range []string{
		`__in0 := goivy.NewConst("__in0_p"`,
		`__in0_x := goivy.NewConst("__in0_p_x"`,
		`__in0_y := goivy.NewConst("__in0_p_y"`,
	} {
		if !strings.Contains(actions, want) {
			t.Errorf("missing per-field input decl %q:\n%s", want, actions)
		}
	}
	// buildPrecondition_<Name> signature receives all three.
	if !strings.Contains(actions, "buildPrecondition_SetToX(state *State, __in0 *goivy.Const, __in0_x *goivy.Const, __in0_y *goivy.Const)") {
		t.Errorf("buildPrecondition signature missing per-field args:\n%s", actions)
	}
	// Destructor-equality conjuncts in the precondition body.
	if !strings.Contains(actions, `&goivy.Eq{T1: mustApply(goivy.NewConst("x"`) {
		t.Errorf("precondition should add x(p)=p_x equality:\n%s", actions)
	}
	if !strings.Contains(actions, `&goivy.Eq{T1: mustApply(goivy.NewConst("y"`) {
		t.Errorf("precondition should add y(p)=p_y equality:\n%s", actions)
	}
	// Per-field pick + struct assembly in Generate.
	if !strings.Contains(actions, "v0_x := pickBoolOrChoose(g.sol, modelResult, __in0_x)") {
		t.Errorf("Generate should pick __in0_x via pickBoolOrChoose:\n%s", actions)
	}
	if !strings.Contains(actions, "v0_y := pickBoolOrChoose(g.sol, modelResult, __in0_y)") {
		t.Errorf("Generate should pick __in0_y:\n%s", actions)
	}
	if !strings.Contains(actions, "v0 := Point{X: v0_x, Y: v0_y}") {
		t.Errorf("Generate should assemble Point{X: …, Y: …}:\n%s", actions)
	}
	if !strings.Contains(actions, "state.SetToX(v0)") {
		t.Errorf("Generate should call SetToX(v0):\n%s", actions)
	}
}

func TestEmit_TestTarget_IndexedDestructorFieldsSkipped(t *testing.T) {
	// Indexed destructor (data(I: idx) : bool) should NOT contribute
	// per-field synthesis — those fields can't be solved as a single
	// value. The receiver still gets an __in<i>; the indexed field
	// stays the zero value.
	mod := compileIvySource(t, `
type idx = {0..3}
type box = struct {
	cell(I: idx) : bool
}
relation flag
action use_box(b: box) = {
	flag := true
}
`)
	out, err := Generate(mod, Config{Target: "test", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	actions := out.Files["actions.go"]
	// No per-field input symbol for `cell` since it's indexed.
	if strings.Contains(actions, "__in0_cell") {
		t.Errorf("indexed destructor field cell should not get a per-field input:\n%s", actions)
	}
}

// --- OPEN 055.8: variant-typed param synthesis ----------------------

func TestEmit_TestTarget_VariantParamPlainLeavesUseConstructorSwitch(t *testing.T) {
	mod := compileIvySource(t, `
type super
type leaf_a
type leaf_b
variant leaf_a of super
variant leaf_b of super
relation flag
action probe(s: super) = {
	flag := true
}
`)
	out, err := Generate(mod, Config{Target: "test", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	actions := out.Files["actions.go"]

	// Receiver Const declared.
	if !strings.Contains(actions, `__in0 := goivy.NewConst("__in0_s"`) {
		t.Errorf("missing variant receiver Const:\n%s", actions)
	}
	// Tag pick via ivyChoose(N).
	if !strings.Contains(actions, "v0_tag := ivyChoose(2)") {
		t.Errorf("variant should pick tag via ivyChoose(2):\n%s", actions)
	}
	// Switch with per-leaf constructor calls.
	if !strings.Contains(actions, "switch v0_tag {") {
		t.Errorf("variant should switch on tag:\n%s", actions)
	}
	if !strings.Contains(actions, "v0 = NewSuperLeafA()") {
		t.Errorf("case 0 should construct LeafA via NewSuperLeafA:\n%s", actions)
	}
	if !strings.Contains(actions, "v0 = NewSuperLeafB()") {
		t.Errorf("case 1 should construct LeafB via NewSuperLeafB:\n%s", actions)
	}
	// Action invoked with synthesised value.
	if !strings.Contains(actions, "state.Probe(v0)") {
		t.Errorf("Generate should call Probe(v0):\n%s", actions)
	}
}

func TestEmit_TestTarget_VariantNoLeaves_ZeroValue(t *testing.T) {
	// Variant supertype with zero leaves should fall back to the
	// zero-value path (no switch).
	mod := compileIvySource(t, `
type super
relation flag
action probe(s: super) = {
	flag := true
}
`)
	out, err := Generate(mod, Config{Target: "test", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	actions := out.Files["actions.go"]
	// `super` here is a plain UninterpretedSort (no variants registered),
	// so the synthesis path should NOT emit a variant switch.
	if strings.Contains(actions, "v0_tag := ivyChoose") {
		t.Errorf("non-variant uninterp should not get variant tag pick:\n%s", actions)
	}
}

func TestSmoke_BuildEmittedTest_VariantParam(t *testing.T) {
	if !SlowGoTest {
		t.Skip("SLOW_GO_TEST not set")
	}
	mod := compileIvySource(t, `
type super
type leaf_a
type leaf_b
variant leaf_a of super
variant leaf_b of super
relation flag
action probe(s: super) = {
	flag := true
}
`)
	buildEmittedAgainstGoivy(t, mod, "ivygo_variant_param")
}

// --- OPEN 055.5 / 055.6: end-to-end build smokes for harder Pre ----

func TestSmoke_BuildEmittedTest_EnumPre(t *testing.T) {
	if !SlowGoTest {
		t.Skip("SLOW_GO_TEST not set")
	}
	mod := compileIvySource(t, `
type color = {red, green, blue}
function pick : color
action set_pick(c: color) = {
	require c = red;
	pick := c
}
`)
	buildEmittedAgainstGoivy(t, mod, "ivygo_enum_pre")
}

func TestSmoke_BuildEmittedTest_RangePre(t *testing.T) {
	if !SlowGoTest {
		t.Skip("SLOW_GO_TEST not set")
	}
	mod := compileIvySource(t, `
type idx = {0..7}
relation slot(I: idx)
action set_slot(i: idx) = {
	require i < 4;
	slot(i) := true
}
`)
	buildEmittedAgainstGoivy(t, mod, "ivygo_range_pre")
}

func TestSmoke_BuildEmittedTest_ForAllPre(t *testing.T) {
	if !SlowGoTest {
		t.Skip("SLOW_GO_TEST not set")
	}
	mod := compileIvySource(t, `
type node = {0..3}
relation link(N1: node, N2: node)
action sym = {
	require forall X. link(X, X)
}
`)
	buildEmittedAgainstGoivy(t, mod, "ivygo_forall_pre")
}

func TestSmoke_BuildEmittedTest_StructDestrParam(t *testing.T) {
	if !SlowGoTest {
		t.Skip("SLOW_GO_TEST not set")
	}
	mod := compileIvySource(t, `
type point = struct { x : bool, y : bool }
relation flag
action set_to_x(p: point) = {
	require x(p);
	flag := x(p)
}
`)
	buildEmittedAgainstGoivy(t, mod, "ivygo_struct_param")
}

// buildEmittedAgainstGoivy is a shared helper: write the test-target
// output to a unique playpen subdir (under ~/ivy/goivy/playpen/
// _smoke-…) and run `go build .`.
//
// Because the playpen lives inside the goivy module, the emitted
// package's `import "github.com/glycerine/ivy/goivy"` resolves via
// the enclosing go.mod — no replace, no separate go.mod, no proxy
// fetch. The `_smoke-…` prefix means `go build ./...` from the
// goivy module root skips these directories automatically.
//
// modName is informational only; it appears in the temp dir name
// for grep-ability if a test leaves output behind on a failure.
func buildEmittedAgainstGoivy(t *testing.T, mod *goivy.Module, modName string) {
	t.Helper()
	_ = modName
	out, err := Generate(mod, Config{Target: "test", PackageName: "main"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	dir := playpenDir(t)
	if err := WriteOutput(out, dir); err != nil {
		t.Fatalf("WriteOutput: %v", err)
	}
	pkgDir := outputDirectory(dir, out.BaseName)
	build := exec.Command("go", "build", ".")
	build.Dir = pkgDir
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build:\n%s", string(output))
	}
}

// --- Test-loop main: build AND run the emitted binary -------------

func TestSmoke_BuildAndRunEmittedTestBinary(t *testing.T) {
	if !SlowGoTest {
		t.Skip("SLOW_GO_TEST not set")
	}
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
	out, err := Generate(mod, Config{Target: "test", PackageName: "main"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	dir := playpenDir(t)
	if err := WriteOutput(out, dir); err != nil {
		t.Fatalf("WriteOutput: %v", err)
	}
	pkgDir := outputDirectory(dir, out.BaseName)
	build := exec.Command("go", "build", "-o", "test_bin", ".")
	build.Dir = pkgDir
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build:\n%s", string(output))
	}
	// Run the binary with a small iter count so the test is fast.
	run := exec.Command(filepath.Join(pkgDir, "test_bin"), "--iters=5")
	output, err := run.CombinedOutput()
	if err != nil {
		t.Fatalf("run failed: %v\n%s", err, string(output))
	}
	if !strings.Contains(string(output), "test_completed") {
		t.Errorf("test binary should print test_completed; got:\n%s", string(output))
	}
}

// --- OPEN 055.4: extended reifier coverage --------------------------

func TestReifyExprAsGoCode_ApplyWrapsInMustApply(t *testing.T) {
	g := newExprGen(t, "")
	// f : bool * bool -> bool, applied to two consts.
	fnSort, _ := goivy.NewFunctionSort(goivy.Boolean, goivy.Boolean, goivy.Boolean)
	fn := goivy.NewConst("f", fnSort)
	a := &goivy.Const{Name: "x", CSort: goivy.Boolean}
	b := &goivy.Const{Name: "y", CSort: goivy.Boolean}
	apply, err := goivy.NewApply(fn, a, b)
	if err != nil {
		t.Fatalf("NewApply: %v", err)
	}
	got, ok := g.reifyExprAsGoCode(apply, nil)
	if !ok {
		t.Fatal("Apply reify not ok")
	}
	if !strings.HasPrefix(got, "mustApply(") {
		t.Errorf("Apply reify should start with mustApply(, got: %q", got)
	}
	if !strings.Contains(got, "mustNewFunctionSort(") {
		t.Errorf("Apply reify should reify function sort via mustNewFunctionSort, got: %q", got)
	}
	// Reifier should have flagged must helpers.
	if !g.Ctx.OnceGlobals["__need_musthelpers"] {
		t.Error("Apply reify should require must helpers")
	}
}

func TestReifyExprAsGoCode_ForAllEmitsVariableSlice(t *testing.T) {
	g := newExprGen(t, "")
	v, _ := goivy.NewVariable("X", goivy.Boolean)
	body := &goivy.Const{Name: "true", CSort: goivy.Boolean}
	got, ok := g.reifyExprAsGoCode(&goivy.ForAll{Variables: []*goivy.LogicVariable{v}, Body: body}, nil)
	if !ok {
		t.Fatal("ForAll reify not ok")
	}
	if !strings.Contains(got, "goivy.ForAll{Variables: []*goivy.LogicVariable{") {
		t.Errorf("ForAll reify should include variable slice, got: %q", got)
	}
	if !strings.Contains(got, `mustNewVariable("X", goivy.Boolean)`) {
		t.Errorf("ForAll should reify variable via mustNewVariable, got: %q", got)
	}
}

func TestReifyExprAsGoCode_ExistsAndIte(t *testing.T) {
	g := newExprGen(t, "")
	v, _ := goivy.NewVariable("Y", goivy.Boolean)
	body := &goivy.Const{Name: "true", CSort: goivy.Boolean}
	got, ok := g.reifyExprAsGoCode(&goivy.LogicExists{Variables: []*goivy.LogicVariable{v}, Body: body}, nil)
	if !ok || !strings.Contains(got, "goivy.LogicExists{Variables:") {
		t.Errorf("LogicExists reify = (%q, %v)", got, ok)
	}
	// Ite uses mustNewIte.
	c := &goivy.Const{Name: "true", CSort: goivy.Boolean}
	thn := &goivy.Const{Name: "true", CSort: goivy.Boolean}
	els := &goivy.Const{Name: "false", CSort: goivy.Boolean}
	ite, _ := goivy.NewIte(c, thn, els)
	got, ok = g.reifyExprAsGoCode(ite, nil)
	if !ok || !strings.HasPrefix(got, "mustNewIte(") {
		t.Errorf("Ite reify = (%q, %v), want prefix mustNewIte(", got, ok)
	}
}

func TestReifyExprAsGoCode_LogicVariableUsesMustNewVariable(t *testing.T) {
	g := newExprGen(t, "")
	v, _ := goivy.NewVariable("X", goivy.Boolean)
	got, ok := g.reifyExprAsGoCode(v, nil)
	if !ok || got != `mustNewVariable("X", goivy.Boolean)` {
		t.Errorf("LogicVariable reify = (%q, %v)", got, ok)
	}
}

func TestReifySortAsGoCode_EnumeratedSort(t *testing.T) {
	g := newExprGen(t, "")
	s := &goivy.LogicEnumeratedSort{Name: "color", Extension: []string{"red", "green"}}
	got, ok := g.reifySortAsGoCode(s)
	if !ok {
		t.Fatal("enum reify not ok")
	}
	if !strings.Contains(got, `&goivy.LogicEnumeratedSort{Name: "color"`) {
		t.Errorf("enum reify missing Name field, got: %q", got)
	}
	if !strings.Contains(got, `Extension: []string{"red", "green"}`) {
		t.Errorf("enum reify should include Extension list, got: %q", got)
	}
}

func TestReifySortAsGoCode_RangeSortWithNumeralBounds(t *testing.T) {
	g := newExprGen(t, "")
	s := &goivy.RangeSort{
		Name: "idx",
		Lb:   goivy.NumeralBound{Value: "0", Sort: nil},
		Ub:   goivy.NumeralBound{Value: "7", Sort: nil},
	}
	got, ok := g.reifySortAsGoCode(s)
	if !ok {
		t.Fatal("range reify not ok")
	}
	if !strings.Contains(got, `&goivy.RangeSort{Name: "idx"`) {
		t.Errorf("range reify missing Name, got: %q", got)
	}
	if !strings.Contains(got, `Lb: goivy.NumeralBound{Value: "0"`) {
		t.Errorf("range reify missing Lb, got: %q", got)
	}
	if !strings.Contains(got, `Ub: goivy.NumeralBound{Value: "7"`) {
		t.Errorf("range reify missing Ub, got: %q", got)
	}
}

func TestReifySortAsGoCode_FunctionSortUsesMustFactory(t *testing.T) {
	g := newExprGen(t, "")
	fs, _ := goivy.NewFunctionSort(goivy.Boolean, goivy.Boolean, goivy.Boolean)
	got, ok := g.reifySortAsGoCode(fs)
	if !ok {
		t.Fatal("function sort reify not ok")
	}
	if !strings.HasPrefix(got, "mustNewFunctionSort(") {
		t.Errorf("function sort reify should use mustNewFunctionSort, got: %q", got)
	}
}

func TestEmitRuntime_MustHelpersEmittedOnDemand(t *testing.T) {
	// A module whose action.Pre needs a function sort triggers the
	// must helpers via the reifier.
	mod := compileIvySource(t, `
type idx = {0..3}
relation slot(I: idx)
action set_slot(i: idx) = {
	require slot(i);
	slot(i) := true
}
`)
	out, err := Generate(mod, Config{Target: "test", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	runtime := out.Files["runtime.go"]
	if !strings.Contains(runtime, "func mustApply(fn goivy.Expr") {
		t.Errorf("mustApply helper missing, got:\n%s", runtime)
	}
	if !strings.Contains(runtime, "func mustNewVariable(") {
		t.Errorf("mustNewVariable helper missing, got:\n%s", runtime)
	}
	if !strings.Contains(runtime, "func mustNewFunctionSort(") {
		t.Errorf("mustNewFunctionSort helper missing, got:\n%s", runtime)
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
	if !strings.Contains(actions, "pickBoolOrChoose(g.sol, modelResult,") {
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
	dir := playpenDir(t)
	if err := WriteOutput(out, dir); err != nil {
		t.Fatalf("WriteOutput: %v", err)
	}
	pkgDir := outputDirectory(dir, out.BaseName)
	cmd := exec.Command("go", "build", "-o", "test_bin", ".")
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
