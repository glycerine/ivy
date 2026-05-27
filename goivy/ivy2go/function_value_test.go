package ivy2go

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy"
)

func mustFunctionSort(t *testing.T, sorts ...goivy.Sort) *goivy.LogicFunctionSort {
	t.Helper()
	fs, err := goivy.NewFunctionSort(sorts...)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	return fs
}

func mustApplyExpr(t *testing.T, fn goivy.Expr, args ...goivy.Expr) *goivy.Apply {
	t.Helper()
	app, err := goivy.NewApply(fn, args...)
	if err != nil {
		t.Fatalf("NewApply: %v", err)
	}
	return app
}

func manualModWithSort(name string, sort goivy.Sort) *goivy.Module {
	mod := goivy.New()
	mod.Sig.Sorts.Set(name, sort)
	mod.SortOrder = append(mod.SortOrder, name)
	return mod
}

func TestEmitApply_FunctionValuedParamUsesParamStorage(t *testing.T) {
	g := newExprGen(t, `
type slot = {zero, one}
`)
	slot, _ := g.Mod.Sig.Sorts.Get2("slot")
	fnSort := mustFunctionSort(t, slot, goivy.Boolean)
	pred := goivy.NewConst("pred", fnSort)
	x := goivy.NewConst("x", slot)
	app := mustApplyExpr(t, pred, x)

	got, err := g.emitExpr(app)
	if err != nil {
		t.Fatalf("emitExpr: %v", err)
	}
	if got != "pred[x]" {
		t.Fatalf("function-valued param apply = %q, want pred[x]", got)
	}
	if strings.Contains(got, "s.") {
		t.Fatalf("function-valued param apply should not read State: %q", got)
	}
}

func TestEmitApply_FunctionValuedReservedParamUsesMangledParam(t *testing.T) {
	g := newExprGen(t, `
type slot = {zero, one}
`)
	slot, _ := g.Mod.Sig.Sorts.Get2("slot")
	fnSort := mustFunctionSort(t, slot, goivy.Boolean)
	pred := goivy.NewConst("type", fnSort)
	x := goivy.NewConst("x", slot)
	app := mustApplyExpr(t, pred, x)

	got, err := g.emitExpr(app)
	if err != nil {
		t.Fatalf("emitExpr: %v", err)
	}
	if got != "type_[x]" {
		t.Fatalf("function-valued reserved param apply = %q, want type_[x]", got)
	}
}

func TestEmitApply_StateFunctionStillUsesStateStorage(t *testing.T) {
	g := newExprGen(t, `
type slot = {zero, one}
relation pred(X: slot)
`)
	slot, _ := g.Mod.Sig.Sorts.Get2("slot")
	entry, ok := g.Mod.Sig.Symbols.Get2("pred")
	if !ok {
		t.Fatal("pred symbol missing")
	}
	pred := goivy.NewConst("pred", entry.Sort)
	x := goivy.NewConst("x", slot)
	app := mustApplyExpr(t, pred, x)

	got, err := g.emitExpr(app)
	if err != nil {
		t.Fatalf("emitExpr: %v", err)
	}
	if got != "s.Pred[x]" {
		t.Fatalf("state function apply = %q, want s.Pred[x]", got)
	}
}

func TestEmitApply_BoolStateFunctionUsesOrdinalIndex(t *testing.T) {
	g := newExprGen(t, `
relation pred(X: bool)
`)
	entry, ok := g.Mod.Sig.Symbols.Get2("pred")
	if !ok {
		t.Fatal("pred symbol missing")
	}
	pred := goivy.NewConst("pred", entry.Sort)
	x := goivy.NewConst("x", goivy.Boolean)
	app := mustApplyExpr(t, pred, x)

	got, err := g.emitExpr(app)
	if err != nil {
		t.Fatalf("emitExpr: %v", err)
	}
	if got != "s.Pred[ivyBoolIndex(x)]" {
		t.Fatalf("bool-indexed state function apply = %q, want s.Pred[ivyBoolIndex(x)]", got)
	}
	if !g.Ctx.OnceGlobals["__need_boolindex"] {
		t.Fatalf("bool-indexed array access should request ivyBoolIndex helper")
	}
}

func TestEmitApply_MapBackedStateFunctionStillUsesGetter(t *testing.T) {
	g := newExprGen(t, `
type key
relation rel(X: key, Y: key)
`)
	key, _ := g.Mod.Sig.Sorts.Get2("key")
	entry, ok := g.Mod.Sig.Symbols.Get2("rel")
	if !ok {
		t.Fatal("rel symbol missing")
	}
	rel := goivy.NewConst("rel", entry.Sort)
	x := goivy.NewConst("x", key)
	y := goivy.NewConst("y", key)
	app := mustApplyExpr(t, rel, x, y)

	got, err := g.emitExpr(app)
	if err != nil {
		t.Fatalf("emitExpr: %v", err)
	}
	if got != "s.getRel(tup__Key__Key{x, y})" {
		t.Fatalf("map-backed state function apply = %q, want getter read", got)
	}
}

func TestEmitActions_FunctionValuedParamUsesParamNotState(t *testing.T) {
	slot := &goivy.LogicEnumeratedSort{Name: "slot", Extension: []string{"zero", "one"}}
	mod := manualModWithSort("slot", slot)
	fnSort := mustFunctionSort(t, slot, goivy.Boolean)
	pred := goivy.NewConst("pred", fnSort)
	x := goivy.NewConst("x", slot)
	outParam := goivy.NewConst("out", goivy.Boolean)
	body := goivy.NewSequence(goivy.NewAssignAction(outParam, mustApplyExpr(t, pred, x)))
	body.SetFormalParams([]*goivy.Const{pred, x})
	body.SetFormalReturns([]*goivy.Const{outParam})
	mod.Actions.Set("check", body)

	out, err := Generate(mod, Config{Target: "impl", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	actions := out.Files["actions.go"]
	requireHasLineWithAllTerms(t, actions, "func", "(s *State)", "Check", "pred [2]bool", "x Slot", "out bool")
	requireHasLineWithAllTerms(t, actions, "out = pred[x]")
	if strings.Contains(actions, "s.Pred") {
		t.Fatalf("function-valued action parameter should not lower through State:\n%s", actions)
	}
	for name, text := range out.Files {
		assertGoSourceGofmt(t, name, text)
	}
}

func TestEmitActions_FunctionValuedLocalReadAndWriteUsesLocalStorage(t *testing.T) {
	slot := &goivy.LogicEnumeratedSort{Name: "slot", Extension: []string{"zero", "one"}}
	mod := manualModWithSort("slot", slot)
	fnSort := mustFunctionSort(t, slot, goivy.Boolean)
	loc := goivy.NewConst("loc", fnSort)
	x := goivy.NewConst("x", slot)
	outParam := goivy.NewConst("out", goivy.Boolean)
	locAtX := mustApplyExpr(t, loc, x)
	body := &goivy.LogicLocalAction{
		Locals: []goivy.Expr{loc},
		Body: goivy.NewSequence(
			goivy.NewAssignAction(locAtX, goivy.True),
			goivy.NewAssignAction(outParam, locAtX),
		),
	}
	action := goivy.NewSequence(body)
	action.SetFormalParams([]*goivy.Const{x})
	action.SetFormalReturns([]*goivy.Const{outParam})
	mod.Actions.Set("check_local", action)

	out, err := Generate(mod, Config{Target: "impl", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	actions := out.Files["actions.go"]
	requireHasLineWithAllTerms(t, actions, "var loc [2]bool")
	requireHasLineWithAllTerms(t, actions, "loc[x] = true")
	requireHasLineWithAllTerms(t, actions, "out = loc[x]")
	if strings.Contains(actions, "s.Loc") {
		t.Fatalf("function-valued local should not lower through State:\n%s", actions)
	}
	for name, text := range out.Files {
		assertGoSourceGofmt(t, name, text)
	}
}

func TestEmitActions_MapBackedFunctionParamDeclaresTupleAndUsesMapIndex(t *testing.T) {
	key := &goivy.UninterpretedSort{Name: "key"}
	mod := manualModWithSort("key", key)
	fnSort := mustFunctionSort(t, key, key, goivy.Boolean)
	rel := goivy.NewConst("rel", fnSort)
	x := goivy.NewConst("x", key)
	y := goivy.NewConst("y", key)
	outParam := goivy.NewConst("out", goivy.Boolean)
	body := goivy.NewSequence(goivy.NewAssignAction(outParam, mustApplyExpr(t, rel, x, y)))
	body.SetFormalParams([]*goivy.Const{rel, x, y})
	body.SetFormalReturns([]*goivy.Const{outParam})
	mod.Actions.Set("check_rel", body)

	out, err := Generate(mod, Config{Target: "impl", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	types := out.Files["types.go"]
	requireHasLineWithAllTerms(t, types, "type tup__Key__Key struct")
	requireHasLineWithAllTerms(t, types, "Arg0 Key")
	requireHasLineWithAllTerms(t, types, "Arg1 Key")

	actions := out.Files["actions.go"]
	requireHasLineWithAllTerms(t, actions, "rel map[tup__Key__Key]bool")
	requireHasLineWithAllTerms(t, actions, "out = rel[tup__Key__Key{x, y}]")
	if strings.Contains(actions, "s.getRel") || strings.Contains(actions, "s.Rel") {
		t.Fatalf("map-backed function-valued param should not lower through State:\n%s", actions)
	}
	for name, text := range out.Files {
		assertGoSourceGofmt(t, name, text)
	}
}

func TestSmoke_BuildEmittedFunctionValuedParamAndLocal(t *testing.T) {
	if !SlowGoTest {
		t.Skip("SLOW_GO_TEST not set")
	}
	slot := &goivy.LogicEnumeratedSort{Name: "slot", Extension: []string{"zero", "one"}}
	key := &goivy.UninterpretedSort{Name: "key"}
	mod := goivy.New()
	mod.Sig.Sorts.Set("slot", slot)
	mod.Sig.Sorts.Set("key", key)
	mod.SortOrder = append(mod.SortOrder, "slot", "key")

	predSort := mustFunctionSort(t, slot, goivy.Boolean)
	pred := goivy.NewConst("pred", predSort)
	x := goivy.NewConst("x", slot)
	outPred := goivy.NewConst("out", goivy.Boolean)
	checkParam := goivy.NewSequence(goivy.NewAssignAction(outPred, mustApplyExpr(t, pred, x)))
	checkParam.SetFormalParams([]*goivy.Const{pred, x})
	checkParam.SetFormalReturns([]*goivy.Const{outPred})
	mod.Actions.Set("check_param", checkParam)

	loc := goivy.NewConst("loc", predSort)
	outLocal := goivy.NewConst("out", goivy.Boolean)
	locAtX := mustApplyExpr(t, loc, x)
	checkLocal := goivy.NewSequence(&goivy.LogicLocalAction{
		Locals: []goivy.Expr{loc},
		Body: goivy.NewSequence(
			goivy.NewAssignAction(locAtX, goivy.True),
			goivy.NewAssignAction(outLocal, locAtX),
		),
	})
	checkLocal.SetFormalParams([]*goivy.Const{x})
	checkLocal.SetFormalReturns([]*goivy.Const{outLocal})
	mod.Actions.Set("check_local", checkLocal)

	relSort := mustFunctionSort(t, key, key, goivy.Boolean)
	rel := goivy.NewConst("rel", relSort)
	k0 := goivy.NewConst("k0", key)
	k1 := goivy.NewConst("k1", key)
	outRel := goivy.NewConst("out", goivy.Boolean)
	checkRel := goivy.NewSequence(goivy.NewAssignAction(outRel, mustApplyExpr(t, rel, k0, k1)))
	checkRel.SetFormalParams([]*goivy.Const{rel, k0, k1})
	checkRel.SetFormalReturns([]*goivy.Const{outRel})
	mod.Actions.Set("check_rel", checkRel)

	out, err := Generate(mod, Config{Target: "impl", PackageName: "main"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	dir := playpenDir(t)
	if err := WriteOutput(out, dir); err != nil {
		t.Fatalf("WriteOutput: %v", err)
	}
	pkgDir := outputDirectory(dir, out.BaseName)
	cmd := exec.Command("go", "build", ".")
	cmd.Dir = pkgDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build failed:\n%s", string(output))
	}
}
