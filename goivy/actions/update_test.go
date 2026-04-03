package actions

import (
	"strings"
	"testing"

	co "github.com/glycerine/ivy/goivy/clauseops"
	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/module"
	"github.com/glycerine/ivy/goivy/transrel"
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
	fmla := lg.NewConst("p", lg.Boolean)
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
	fmla := lg.NewConst("p", lg.Boolean)
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
	if c, ok := not.Body.(*lg.Const); !ok || c.Name != "p" {
		t.Errorf("AssertAction Pre should contain Not(p), got %s", u.Pre.Fmlas[0])
	}
}

// --- AssignAction ---

func TestAssignActionSimple(t *testing.T) {
	// x := y where both are constants with TopS sort
	x := lg.NewConst("x", lg.TopS)
	y := lg.NewConst("y", lg.TopS)
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
	f := lg.NewConst("f", fSort)
	aConst := lg.NewConst("a", lg.TopS)
	bConst := lg.NewConst("b", lg.TopS)
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
	x := lg.NewConst("x", lg.TopS)
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
	code := lg.NewConst("code", lg.TopS)
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
	p := lg.NewConst("p", lg.Boolean)
	q := lg.NewConst("q", lg.Boolean)
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
	x := lg.NewConst("x", lg.TopS)
	y := lg.NewConst("y", lg.TopS)
	p := lg.NewConst("p", lg.Boolean)
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
	p := lg.NewConst("p", lg.Boolean)
	q := lg.NewConst("q", lg.Boolean)
	ch := NewChoiceAction(NewAssumeAction(p), NewAssumeAction(q))
	ctx := testCtx()
	u := ch.IntUpdate(ctx)
	if len(u.Modified) != 0 {
		t.Errorf("Choice of assumes should modify nothing, got %v", u.Modified)
	}
}

// --- IfAction ---

func TestIfActionIntUpdate(t *testing.T) {
	cond := lg.NewConst("c", lg.Boolean)
	p := lg.NewConst("p", lg.Boolean)
	q := lg.NewConst("q", lg.Boolean)
	ifAct := NewIfAction(cond, NewAssumeAction(p), NewAssumeAction(q))
	ctx := testCtx()
	u := ifAct.IntUpdate(ctx)
	if len(u.Modified) != 0 {
		t.Errorf("If of assumes should modify nothing, got %v", u.Modified)
	}
}

// --- LocalAction ---

func TestLocalActionIntUpdate(t *testing.T) {
	x := lg.NewConst("x", lg.TopS)
	y := lg.NewConst("y", lg.TopS)
	// local x { x := y }
	asgn := NewAssignAction(x, y)
	local := NewLocalActionOn(NewActionsConfig(), "test", x, asgn)
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
	p := lg.NewConst("p", lg.Boolean)
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
	x := lg.NewConst("fml:x", lg.TopS)
	y := lg.NewConst("y", lg.TopS)
	asgn := NewAssignAction(x, y)
	asgn.SetFormalParams([]*lg.Const{x})
	ctx := testCtx()
	u := GetUpdate(asgn, ctx)
	// x should be hidden from the modified list
	for _, m := range u.Modified {
		if m.Name == "fml:x" {
			t.Error("GetUpdate should hide formal params from modified")
		}
	}
}

// --- hideFormals __prefix tests ---

// formulaContainsName checks if a formula references a constant with the given name.
func formulaContainsName(node lg.Expr, name string) bool {
	if node == nil {
		return false
	}
	if c, ok := node.(*lg.Const); ok {
		return c.Name == name
	}
	if app, ok := node.(*lg.Apply); ok {
		if formulaContainsName(app.Func, name) {
			return true
		}
	}
	for _, ch := range node.Children() {
		if formulaContainsName(ch, name) {
			return true
		}
	}
	return false
}

// formulaContainsPrefix checks if any constant name in the formula starts with prefix.
func formulaContainsPrefix(node lg.Expr, prefix string) bool {
	if node == nil {
		return false
	}
	if c, ok := node.(*lg.Const); ok {
		return strings.HasPrefix(c.Name, prefix)
	}
	if app, ok := node.(*lg.Apply); ok {
		if formulaContainsPrefix(app.Func, prefix) {
			return true
		}
	}
	for _, ch := range node.Children() {
		if formulaContainsPrefix(ch, prefix) {
			return true
		}
	}
	return false
}

// collectConstNames returns all constant names in a formula.
func collectConstNames(node lg.Expr, out map[string]bool) {
	if node == nil {
		return
	}
	if c, ok := node.(*lg.Const); ok {
		out[c.Name] = true
		return
	}
	if app, ok := node.(*lg.Apply); ok {
		collectConstNames(app.Func, out)
	}
	for _, ch := range node.Children() {
		collectConstNames(ch, out)
	}
}

// TestHideFormalsRenamesWithDoubleUnderscore verifies that hideFormals
// (via transrel.Hide → ExistQuantClauses) renames formal params and their
// new_ versions with __ prefix. This is the core mechanism that should
// produce __new_loc:wr from new_loc:wr in the fragment checker.
func TestHideFormalsRenamesWithDoubleUnderscore(t *testing.T) {
	// Create symbols with a non-TopS sort (UninterpretedSort).
	// This is critical — the original bug was that TopS-keyed lookups
	// failed to match constants with real sorts.
	mySort := &lg.UninterpretedSort{Name: "mytype"}
	loc := lg.NewConst("loc", mySort)
	val := lg.NewConst("val", mySort)

	// Create "loc := val" assignment — this produces new_loc in the TR
	asgn := NewAssignAction(loc, val)
	asgn.SetFormalParams([]*lg.Const{loc})

	ctx := testCtx()
	// Step 1: IntUpdate — should produce new_loc in TR
	u := IntUpdate(asgn, ctx)
	trFormula := u.TRNode()
	t.Logf("After IntUpdate, TR formula: %v", trFormula)

	names := make(map[string]bool)
	collectConstNames(trFormula, names)
	t.Logf("After IntUpdate, constant names: %v", names)

	if !formulaContainsName(trFormula, "new_loc") {
		t.Logf("IntUpdate TR does not contain new_loc (may use different structure)")
	}

	// Step 2: BindOldsAction
	u = transrel.BindOldsAction(u)

	// Step 3: hideFormals — should rename loc and new_loc with __ prefix
	u = hideFormals(asgn, u)
	trAfterHide := u.TRNode()

	names2 := make(map[string]bool)
	collectConstNames(trAfterHide, names2)
	t.Logf("After hideFormals, constant names: %v", names2)

	// The bare "loc" should NOT appear (it was hidden)
	if formulaContainsName(trAfterHide, "loc") {
		t.Error("After hideFormals, bare 'loc' should be renamed with __ prefix")
	}

	// The bare "new_loc" should NOT appear (it was hidden)
	if formulaContainsName(trAfterHide, "new_loc") {
		t.Error("After hideFormals, bare 'new_loc' should be renamed with __ prefix")
	}

	// There should be some __-prefixed name
	if !formulaContainsPrefix(trAfterHide, "__") {
		t.Error("After hideFormals, expected at least one __-prefixed constant")
	}
}

// TestHideFormalsWithTopSSort verifies that Hide works even when the
// formal param has TopS sort (the original code's assumption).
func TestHideFormalsWithTopSSort(t *testing.T) {
	loc := lg.NewConst("loc", lg.TopS)
	x := lg.NewConst("x", lg.TopS)
	asgn := NewAssignAction(loc, x)
	asgn.SetFormalParams([]*lg.Const{loc})

	ctx := testCtx()
	u := IntUpdate(asgn, ctx)
	u = transrel.BindOldsAction(u)
	u = hideFormals(asgn, u)
	trAfterHide := u.TRNode()

	names := make(map[string]bool)
	collectConstNames(trAfterHide, names)
	t.Logf("After hideFormals (TopS), constant names: %v", names)

	if formulaContainsName(trAfterHide, "loc") {
		t.Error("After hideFormals, bare 'loc' should be renamed")
	}
	if formulaContainsName(trAfterHide, "new_loc") {
		t.Error("After hideFormals, bare 'new_loc' should be renamed")
	}
}

// TestTransrelHideDirectly tests transrel.Hide directly with a
// FunctionSort symbol to verify ExistQuantClauses works with real sorts.
func TestTransrelHideDirectly(t *testing.T) {
	mySort := &lg.UninterpretedSort{Name: "mytype"}
	loc := lg.NewConst("loc", mySort)
	newLoc := lg.NewConst("new_loc", mySort)

	// Build a simple update with loc in Modified and new_loc in TR
	eq, _ := lg.NewEq(newLoc, loc)
	tr := &lg.And{Terms: []lg.Expr{eq}}

	u := &transrel.Update{
		Modified: []*lg.Const{loc},
		TR:       co.FormulaToClauses(tr, nil),
		Pre:      co.FormulaToClauses(lg.False, nil),
	}

	// Hide loc — should also hide new_loc, both renamed with __
	hidden := transrel.Hide([]*lg.Const{loc}, u)
	trFormula := hidden.TRNode()

	names := make(map[string]bool)
	collectConstNames(trFormula, names)
	t.Logf("After Hide (FunctionSort), constant names: %v", names)

	if formulaContainsName(trFormula, "loc") {
		t.Errorf("After Hide, bare 'loc' should be renamed to __loc, got names: %v", names)
	}
	if formulaContainsName(trFormula, "new_loc") {
		t.Errorf("After Hide, bare 'new_loc' should be renamed to __new_loc, got names: %v", names)
	}
	if !formulaContainsPrefix(trFormula, "__") {
		t.Errorf("After Hide, expected __-prefixed constants, got names: %v", names)
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
	x := lg.NewConst("x", lg.TopS)
	y := lg.NewConst("y", lg.TopS)
	eq := equivAST(x, y)
	if _, ok := eq.(*lg.Eq); !ok {
		t.Errorf("equivAST for individuals should return Eq, got %T", eq)
	}
}

func TestEquivASTBoolean(t *testing.T) {
	p := lg.NewConst("p", lg.Boolean)
	q := lg.NewConst("q", lg.Boolean)
	result := equivAST(p, q)
	if _, ok := result.(*lg.And); !ok {
		t.Errorf("equivAST for booleans should return And, got %T: %s", result, result)
	}
}

func TestDualFormula(t *testing.T) {
	p := lg.NewConst("p", lg.Boolean)
	dual := dualFormula(p)
	not, ok := dual.(*lg.Not)
	if !ok {
		t.Fatalf("dualFormula of constant should be Not, got %T: %s", dual, dual)
	}
	if c, ok := not.Body.(*lg.Const); !ok || c.Name != "p" {
		t.Errorf("dualFormula body should be p, got %s", not.Body)
	}
}

func TestSkolemizeFormula(t *testing.T) {
	v, _ := lg.NewVariable("X", lg.TopS)
	body := v
	ex, _ := lg.NewExists([]*lg.Variable{v}, body)
	result := skolemizeFormula(ex)
	// Should replace X with __sk__X
	if c, ok := result.(*lg.Const); ok {
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
	p := lg.NewConst("p", lg.Boolean)
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
	p := lg.NewConst("p", lg.Boolean)
	result := disjoin(lg.False, p)
	if result != p {
		t.Errorf("disjoin(false, p) should be p, got %s", result)
	}
	if !isTrue(disjoin(lg.True, p)) {
		t.Error("disjoin(true, p) should be true")
	}
}
