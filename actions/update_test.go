package actions

import (
	"strings"
	"testing"

	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
	"github.com/glycerine/goivy/transrel"
)

func testCtx() *UpdateContext {
	return &UpdateContext{
		Domain: module.New(),
		PVars:  nil,
		ActCfg: NewActionsConfig(),
	}
}

// --- AssumeAction ---

func TestAssumeActionUpdate(t *testing.T) {
	fmla := lg.NewSymbol("p", lg.Boolean)
	a := NewAssumeAction(fmla)
	ctx := testCtx()
	u := a.ActionUpdate(ctx)
	if len(u.Modified) != 0 {
		t.Errorf("AssumeAction should modify nothing, got %v", u.Modified)
	}
	if u.TR == nil || u.TR.IsFalse() {
		t.Error("AssumeAction TR should be the formula, not false")
	}
	if !u.Pre.IsFalse() {
		t.Error("AssumeAction Pre should be false (never fails)")
	}
}

func TestAssumeActionUpdateTrue(t *testing.T) {
	a := NewAssumeAction(lg.True)
	ctx := testCtx()
	u := a.ActionUpdate(ctx)
	if !u.TR.IsTrue() {
		t.Errorf("AssumeAction(true) TR should be true, got %s", u.TR)
	}
}

// --- AssertAction ---

func TestAssertActionUpdate(t *testing.T) {
	fmla := lg.NewSymbol("p", lg.Boolean)
	a := NewAssertAction(fmla)
	ctx := testCtx()
	u := a.ActionUpdate(ctx)
	if len(u.Modified) != 0 {
		t.Errorf("AssertAction should modify nothing, got %v", u.Modified)
	}
	if !u.TR.IsTrue() {
		t.Error("AssertAction TR should be true")
	}
	// Pre should be the dual (negation) of p
	if u.Pre == nil || u.Pre.IsFalse() {
		t.Error("AssertAction Pre should be the negated formula")
	}
	// Pre is a Clauses with one formula: Not(p). Access it directly.
	if len(u.Pre.Fmlas) != 1 {
		t.Fatalf("AssertAction Pre should have 1 formula, got %d", len(u.Pre.Fmlas))
	}
	not, ok := u.Pre.Fmlas[0].(*lg.Not)
	if !ok {
		t.Fatalf("AssertAction Pre should contain Not(p), got %T: %s", u.Pre.Fmlas[0], u.Pre.Fmlas[0])
	}
	if c, ok := not.Body.(*lg.Symbol); !ok || c.Name != "p" {
		t.Errorf("AssertAction Pre should contain Not(p), got %s", u.Pre.Fmlas[0])
	}
}

// --- AssignAction ---

func TestAssignActionSimple(t *testing.T) {
	// x := y where both are constants with TopS sort
	x := lg.NewSymbol("x", lg.TopS)
	y := lg.NewSymbol("y", lg.TopS)
	a := NewAssignAction(x, y)
	ctx := testCtx()
	u := a.ActionUpdate(ctx)
	if len(u.Modified) != 1 || u.Modified[0].Name != "x" {
		t.Errorf("AssignAction should modify [x], got %v", u.Modified)
	}
	// TR should contain an equivalence new_x = y
	trStr := u.TR.String()
	if !strings.Contains(trStr, "new_x") {
		t.Errorf("AssignAction TR should reference new_x, got %s", trStr)
	}
	if !u.Pre.IsFalse() {
		t.Error("AssignAction Pre should be false")
	}
}

func TestAssignActionWithArgs(t *testing.T) {
	// f(a) := b where f is a function constant
	fSort, _ := lg.NewFunctionSort(lg.TopS, lg.TopS)
	f := lg.NewSymbol("f", fSort)
	aConst := lg.NewSymbol("a", lg.TopS)
	bConst := lg.NewSymbol("b", lg.TopS)
	lhs := &lg.Apply{Func: f, Terms: []lg.Expr{aConst}}
	a := NewAssignAction(lhs, bConst)
	ctx := testCtx()
	u := a.ActionUpdate(ctx)
	if len(u.Modified) != 1 || u.Modified[0].Name != "f" {
		t.Errorf("should modify [f], got %v", u.Modified)
	}
	// TR should have an ITE: at index a, use b; elsewhere keep f
	trStr := u.TR.String()
	if !strings.Contains(trStr, "new_f") {
		t.Errorf("TR should reference new_f, got %s", trStr)
	}
}

// --- HavocAction ---

func TestHavocActionUpdate(t *testing.T) {
	x := lg.NewSymbol("x", lg.TopS)
	a := NewHavocAction(x)
	ctx := testCtx()
	u := a.ActionUpdate(ctx)
	if len(u.Modified) != 1 || u.Modified[0].Name != "x" {
		t.Errorf("HavocAction should modify [x], got %v", u.Modified)
	}
	if !u.Pre.IsFalse() {
		t.Error("HavocAction Pre should be false")
	}
}

// --- NativeAction ---

func TestNativeActionUpdate(t *testing.T) {
	code := lg.NewSymbol("code", lg.TopS)
	a := NewNativeAction(code)
	ctx := testCtx()
	u := a.IntUpdate(ctx)
	if len(u.Modified) != 0 {
		t.Errorf("NativeAction should modify nothing, got %v", u.Modified)
	}
	if !u.TR.IsTrue() {
		t.Error("NativeAction TR should be true")
	}
}

// --- Sequence ---

func TestSequenceIntUpdate(t *testing.T) {
	// Sequence of two assumes: assume p; assume q
	p := lg.NewSymbol("p", lg.Boolean)
	q := lg.NewSymbol("q", lg.Boolean)
	seq := NewSequence(NewAssumeAction(p), NewAssumeAction(q))
	ctx := testCtx()
	u := seq.IntUpdate(ctx)
	// Should modify nothing
	if len(u.Modified) != 0 {
		t.Errorf("Sequence of assumes should modify nothing, got %v", u.Modified)
	}
	// TR should contain both p and q conjoined
	if u.TR.IsTrue() {
		t.Error("Sequence TR should not be just true")
	}
}

func TestSequenceAssignAssume(t *testing.T) {
	// x := y; assume p — should modify [x]
	x := lg.NewSymbol("x", lg.TopS)
	y := lg.NewSymbol("y", lg.TopS)
	p := lg.NewSymbol("p", lg.Boolean)
	seq := NewSequence(NewAssignAction(x, y), NewAssumeAction(p))
	ctx := testCtx()
	u := seq.IntUpdate(ctx)
	hasX := false
	for _, m := range u.Modified {
		if m.Name == "x" {
			hasX = true
		}
	}
	if !hasX {
		t.Errorf("Sequence should modify x, got %v", u.Modified)
	}
}

// --- ChoiceAction ---

func TestChoiceActionIntUpdate(t *testing.T) {
	// choice { assume p } or { assume q }
	p := lg.NewSymbol("p", lg.Boolean)
	q := lg.NewSymbol("q", lg.Boolean)
	ch := NewChoiceAction(NewAssumeAction(p), NewAssumeAction(q))
	ctx := testCtx()
	u := ch.IntUpdate(ctx)
	if len(u.Modified) != 0 {
		t.Errorf("Choice of assumes should modify nothing, got %v", u.Modified)
	}
}

// --- IfAction ---

func TestIfActionIntUpdate(t *testing.T) {
	cond := lg.NewSymbol("c", lg.Boolean)
	p := lg.NewSymbol("p", lg.Boolean)
	q := lg.NewSymbol("q", lg.Boolean)
	ifAct := NewIfAction(cond, NewAssumeAction(p), NewAssumeAction(q))
	ctx := testCtx()
	u := ifAct.IntUpdate(ctx)
	if len(u.Modified) != 0 {
		t.Errorf("If of assumes should modify nothing, got %v", u.Modified)
	}
}

// --- LocalAction ---

func TestLocalActionIntUpdate(t *testing.T) {
	x := lg.NewSymbol("x", lg.TopS)
	y := lg.NewSymbol("y", lg.TopS)
	// local x { x := y }
	asgn := NewAssignAction(x, y)
	local := NewActionsConfig().NewLocalAction(x, asgn)
	ctx := testCtx()
	u := local.IntUpdate(ctx)
	// x should be hidden — not in modified list
	for _, m := range u.Modified {
		if m.Name == "x" {
			t.Error("Local should hide x from modified list")
		}
	}
}

// --- BindOldsAction ---

func TestBindOldsActionIntUpdate(t *testing.T) {
	p := lg.NewSymbol("p", lg.Boolean)
	inner := NewAssumeAction(p)
	bindOlds := NewBindOldsAction(inner)
	ctx := testCtx()
	u := bindOlds.IntUpdate(ctx)
	if len(u.Modified) != 0 {
		t.Errorf("BindOlds of assume should modify nothing, got %v", u.Modified)
	}
}

// --- GetUpdate ---

func TestGetUpdateHidesFormals(t *testing.T) {
	x := lg.NewSymbol("fml:x", lg.TopS)
	y := lg.NewSymbol("y", lg.TopS)
	asgn := NewAssignAction(x, y)
	asgn.SetFormalParams([]*lg.Symbol{x})
	ctx := testCtx()
	u := GetUpdate(asgn, ctx)
	// x should be hidden from the modified list
	for _, m := range u.Modified {
		if m.Name == "fml:x" {
			t.Error("GetUpdate should hide formal params from modified")
		}
	}
}

// --- NullUpdate ---

func TestNullUpdate(t *testing.T) {
	u := transrel.NullUpdate()
	if u.Modified == nil {
		t.Error("NullUpdate Modified should be non-nil empty slice")
	}
	if len(u.Modified) != 0 {
		t.Errorf("NullUpdate Modified should be empty, got %v", u.Modified)
	}
}

// --- Helper functions ---

func TestEquivASTIndividual(t *testing.T) {
	x := lg.NewSymbol("x", lg.TopS)
	y := lg.NewSymbol("y", lg.TopS)
	eq := equivAST(x, y)
	if _, ok := eq.(*lg.Eq); !ok {
		t.Errorf("equivAST for individuals should return Eq, got %T", eq)
	}
}

func TestEquivASTBoolean(t *testing.T) {
	p := lg.NewSymbol("p", lg.Boolean)
	q := lg.NewSymbol("q", lg.Boolean)
	result := equivAST(p, q)
	if _, ok := result.(*lg.And); !ok {
		t.Errorf("equivAST for booleans should return And, got %T: %s", result, result)
	}
}

func TestDualFormula(t *testing.T) {
	p := lg.NewSymbol("p", lg.Boolean)
	dual := dualFormula(p)
	not, ok := dual.(*lg.Not)
	if !ok {
		t.Fatalf("dualFormula of constant should be Not, got %T: %s", dual, dual)
	}
	if c, ok := not.Body.(*lg.Symbol); !ok || c.Name != "p" {
		t.Errorf("dualFormula body should be p, got %s", not.Body)
	}
}

func TestSkolemizeFormula(t *testing.T) {
	v, _ := lg.NewVariable("X", lg.TopS)
	body := v
	ex, _ := lg.NewExists([]*lg.Variable{v}, body)
	result := skolemizeFormula(ex)
	// Should replace X with __sk__X
	if c, ok := result.(*lg.Symbol); ok {
		if c.Name != "__sk__X" {
			t.Errorf("expected __sk__X, got %s", c.Name)
		}
	} else {
		t.Errorf("expected Const after skolemization, got %T: %s", result, result)
	}
}

func TestConjoin(t *testing.T) {
	if !isTrue(conjoin(lg.True, lg.True)) {
		t.Error("conjoin(true, true) should be true")
	}
	p := lg.NewSymbol("p", lg.Boolean)
	result := conjoin(lg.True, p)
	if result != p {
		t.Errorf("conjoin(true, p) should be p, got %s", result)
	}
	if !isFalse(conjoin(lg.False, p)) {
		t.Error("conjoin(false, p) should be false")
	}
}

func TestDisjoin(t *testing.T) {
	if !isFalse(disjoin(lg.False, lg.False)) {
		t.Error("disjoin(false, false) should be false")
	}
	p := lg.NewSymbol("p", lg.Boolean)
	result := disjoin(lg.False, p)
	if result != p {
		t.Errorf("disjoin(false, p) should be p, got %s", result)
	}
	if !isTrue(disjoin(lg.True, p)) {
		t.Error("disjoin(true, p) should be true")
	}
}
