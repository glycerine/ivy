package actions

// Tests for fixes #10–#16 from §5.3 (BEHAVIORAL_DIFFERENCE items).

import (
	"testing"

	il "github.com/glycerine/ivy/goivy/ivylogic"
	lg "github.com/glycerine/ivy/goivy/logic"
)

// ---------------------------------------------------------------------------
// Fix #10: AssertAction.ActionUpdate — clausification + annotation wrapping
// ---------------------------------------------------------------------------

func TestAssertAction_Clausifies(t *testing.T) {
	// Python: cl = formula_to_clauses(dual_formula(fmla))
	//         cl = Clauses(cl.fmlas, cl.defs, EmptyAnnotation())
	//         return ([], true_clauses(annot=EmptyAnnotation()), cl)
	fmla := lg.NewConst("p", lg.Boolean)
	a := NewAssertAction(fmla)
	ctx := testCtx()
	u := a.ActionUpdate(ctx)

	// TR should be true_clauses (not formula_to_clauses of True)
	if u.TR == nil {
		t.Fatal("AssertAction TR should not be nil")
	}
	if !u.TR.IsTrue() {
		t.Errorf("AssertAction TR should be true clauses, got %v", u.TR)
	}

	// Pre should be clausified dual formula with EmptyAnnotation
	if u.Pre == nil {
		t.Fatal("AssertAction Pre should not be nil")
	}
	if u.Pre.Annot == nil {
		t.Error("AssertAction Pre should have EmptyAnnotation, got nil")
	}
	if _, ok := u.Pre.Annot.(EmptyAnnotation); !ok {
		t.Errorf("AssertAction Pre annotation should be EmptyAnnotation, got %T", u.Pre.Annot)
	}

	// Pre should have formulas (the negated formula)
	if len(u.Pre.Fmlas) == 0 {
		t.Error("AssertAction Pre should have clausified formulas")
	}
}

func TestAssertAction_CheckedAssertSkipReturnsFormula(t *testing.T) {
	// When checked_assert doesn't match, provable assertions return formula as-is
	fmla := lg.NewConst("p", lg.Boolean)
	a := NewAssertAction(fmla)
	ctx := testCtx()
	ctx.CheckedAssert = "other:42"

	u := a.ActionUpdate(ctx)

	// Should return formula as TR (not dual), false Pre
	if u.TR == nil || u.TR.IsTrue() {
		t.Error("Non-matching checked_assert should put formula in TR")
	}
}

// ---------------------------------------------------------------------------
// Fix #11: AssumeAction.ActionUpdate — skolemize + clausify + unfold
// ---------------------------------------------------------------------------

func TestAssumeAction_SkolemizesAndClausifies(t *testing.T) {
	// Python: clauses = formula_to_clauses_tseitin(skolemize_formula(fmla))
	//         clauses = unfold_definitions_clauses(clauses)
	//         clauses = Clauses(clauses.fmlas, clauses.defs, EmptyAnnotation())

	// Use a simple boolean formula (skolemization is tested elsewhere)
	fmla := lg.NewConst("p", lg.Boolean)

	a := NewAssumeAction(fmla)
	ctx := testCtx()
	u := a.ActionUpdate(ctx)

	// TR should have the clausified + skolemized formula
	if u.TR == nil {
		t.Fatal("AssumeAction TR should not be nil")
	}
	// Should have EmptyAnnotation
	if u.TR.Annot == nil {
		t.Error("AssumeAction TR should have EmptyAnnotation")
	}
	if _, ok := u.TR.Annot.(EmptyAnnotation); !ok {
		t.Errorf("AssumeAction TR annotation should be EmptyAnnotation, got %T", u.TR.Annot)
	}

	// Pre should be false clauses
	if u.Pre == nil || !u.Pre.IsFalse() {
		t.Error("AssumeAction Pre should be false clauses")
	}
}

func TestAssumeAction_Unprovable(t *testing.T) {
	fmla := lg.NewConst("p", lg.Boolean)
	a := NewAssumeAction(fmla)
	a.Unprovable = true
	ctx := testCtx()
	u := a.ActionUpdate(ctx)

	if !u.TR.IsTrue() {
		t.Error("Unprovable assume should return true TR")
	}
}

// ---------------------------------------------------------------------------
// Fix #12: VarAction is AST node, not Action
// ---------------------------------------------------------------------------

func TestVarAction_IsNotAction(t *testing.T) {
	// VarAction should NOT implement Action interface
	va := &VarAction{}
	var _ interface{} = va // just make sure it compiles

	// Type-assert to Action should fail
	var iface interface{} = va
	if _, ok := iface.(Action); ok {
		t.Error("VarAction should NOT implement Action interface")
	}
}

// ---------------------------------------------------------------------------
// Fix #13: SubgoalAction extends AssertAction
// ---------------------------------------------------------------------------

func TestSubgoalAction_InheritsAssertActionUpdate(t *testing.T) {
	// SubgoalAction should use AssertAction.ActionUpdate
	fmla := lg.NewConst("p", lg.Boolean)
	sa := NewSubgoalAction(fmla)
	ctx := testCtx()

	u := sa.ActionUpdate(ctx)

	// Should behave same as AssertAction
	if u.TR == nil || !u.TR.IsTrue() {
		t.Error("SubgoalAction TR should be true (from AssertAction)")
	}
	if u.Pre == nil {
		t.Fatal("SubgoalAction Pre should not be nil")
	}
	if _, ok := u.Pre.Annot.(EmptyAnnotation); !ok {
		t.Errorf("SubgoalAction Pre should have EmptyAnnotation, got %T", u.Pre.Annot)
	}
}

func TestSubgoalAction_HasKind(t *testing.T) {
	fmla := lg.NewConst("p", lg.Boolean)
	sa := NewSubgoalAction(fmla)
	sa.SubgoalKind = "safety"

	if sa.Name() != "assert" { // Python: SubgoalAction inherits name() → "assert"
		t.Errorf("SubgoalAction.Name() = %q, want %q", sa.Name(), "assert")
	}

	// Clone should preserve kind
	cloned := sa.ActionClone([]lg.Expr{fmla})
	if sc, ok := cloned.(*SubgoalAction); ok {
		if sc.SubgoalKind != "safety" {
			t.Errorf("Clone should preserve SubgoalKind, got %q", sc.SubgoalKind)
		}
	} else {
		t.Errorf("Clone should return *SubgoalAction, got %T", cloned)
	}
}

// ---------------------------------------------------------------------------
// Fix #14: CopyFieldAction has 4 args (dst, field, src, srcField)
// ---------------------------------------------------------------------------

func TestCopyFieldAction_FourArgs(t *testing.T) {
	sortT := mkSort("T")
	sortS := mkSort("S")
	fldSort := mkBinaryRelSort(sortT, sortS)
	dstField := lg.NewConst("fld", fldSort)
	srcField := lg.NewConst("sfld", fldSort)
	dst := lg.NewConst("dst", sortT)
	src := lg.NewConst("src", sortT)

	a := NewCopyFieldAction(dst, dstField, src, srcField)

	args := a.ActionArgs()
	if len(args) != 4 {
		t.Fatalf("CopyFieldAction should have 4 args, got %d", len(args))
	}
	if args[0] != dst {
		t.Error("arg 0 should be dst")
	}
	if args[1] != dstField {
		t.Error("arg 1 should be dstField")
	}
	if args[2] != src {
		t.Error("arg 2 should be src")
	}
	if args[3] != srcField {
		t.Error("arg 3 should be srcField")
	}
}

func TestCopyFieldAction_DifferentSourceField(t *testing.T) {
	sortT := mkSort("T")
	sortS := mkSort("S")
	fldSort := mkBinaryRelSort(sortT, sortS)
	dstField := lg.NewConst("fld_a", fldSort)
	srcField := lg.NewConst("fld_b", fldSort)
	dst := lg.NewConst("dst", sortT)
	src := lg.NewConst("src", sortT)

	a := NewCopyFieldAction(dst, dstField, src, srcField)
	ctx := testCtx()
	u := a.ActionUpdate(ctx)

	// Should modify the dest field
	if len(u.Modified) != 1 || u.Modified[0].Name != "fld_a" {
		t.Errorf("CopyFieldAction should modify [fld_a], got %v", u.Modified)
	}
}

// ---------------------------------------------------------------------------
// Fix #15: WhileAction.Unroll
// ---------------------------------------------------------------------------

func TestWhileAction_Unroll(t *testing.T) {
	// Create a simple while loop: while x < bound
	sortT := mkSort("T")
	ltSym := lg.NewConst("<", il.RelationSort([]lg.Sort{sortT, sortT}))
	xSym := lg.NewConst("x", sortT)
	boundSym := lg.NewConst("bound", sortT)
	cond, _ := lg.NewApply(ltSym, xSym, boundSym)

	body := NewAssignAction(xSym, xSym)
	wa := NewWhileAction(cond, body)

	// Card function: returns 3 for sortT
	card := func(s lg.Sort) int {
		if s != nil {
			return 3
		}
		return 0
	}

	unrolled, err := wa.Unroll(card, nil)
	if err != nil {
		t.Fatalf("Unroll failed: %v", err)
	}
	if unrolled == nil {
		t.Fatal("Unroll returned nil")
	}

	// Should be an IfAction
	if _, ok := unrolled.(*IfAction); !ok {
		t.Errorf("Unrolled should be *IfAction, got %T", unrolled)
	}
}

func TestWhileAction_Unroll_RefusesLargeCard(t *testing.T) {
	sortT := mkSort("T")
	ltSym := lg.NewConst("<", il.RelationSort([]lg.Sort{sortT, sortT}))
	xSym := lg.NewConst("x", sortT)
	boundSym := lg.NewConst("bound", sortT)
	cond, _ := lg.NewApply(ltSym, xSym, boundSym)

	body := NewSequence()
	wa := NewWhileAction(cond, body)

	card := func(s lg.Sort) int { return 200 }

	_, err := wa.Unroll(card, nil)
	if err == nil {
		t.Error("Unroll should refuse cardinality > 100")
	}
}

func TestWhileAction_Unroll_NotEqCondition(t *testing.T) {
	// while !(x = bound)
	sortT := mkSort("T")
	xSym := lg.NewConst("x", sortT)
	boundSym := lg.NewConst("bound", sortT)
	eq := &lg.Eq{T1: xSym, T2: boundSym}
	cond := &lg.Not{Body: eq}

	body := NewSequence()
	wa := NewWhileAction(cond, body)

	card := func(s lg.Sort) int {
		if s != nil {
			return 2
		}
		return 0
	}

	unrolled, err := wa.Unroll(card, nil)
	if err != nil {
		t.Fatalf("Unroll with Not(Eq) condition failed: %v", err)
	}
	if _, ok := unrolled.(*IfAction); !ok {
		t.Errorf("Unrolled should be *IfAction, got %T", unrolled)
	}
}

func TestWhileAction_IntUpdate_UsesUnrollContext(t *testing.T) {
	// Test that IntUpdate checks for UnrollContext
	sortT := mkSort("T")
	ltSym := lg.NewConst("<", il.RelationSort([]lg.Sort{sortT, sortT}))
	xSym := lg.NewConst("x", sortT)
	boundSym := lg.NewConst("bound", sortT)
	cond, _ := lg.NewApply(ltSym, xSym, boundSym)

	body := NewSequence()
	wa := NewWhileAction(cond, body)

	// Set up UnrollContext with ActionsConfig
	cfg := NewActionsConfig()
	uc := NewUnrollContext(func(s lg.Sort) int { return 2 }, nil, cfg)
	uc.Enter()
	defer uc.Exit()

	ctx := testCtx()
	ctx.ActCfg = cfg
	u := wa.IntUpdate(ctx)
	if u == nil {
		t.Fatal("IntUpdate with UnrollContext should not return nil")
	}
}

// ---------------------------------------------------------------------------
// Fix #16: AssignAction partial application + variable check
// ---------------------------------------------------------------------------

func TestAssignAction_PartialApplication(t *testing.T) {
	// f : S -> S -> Bool, assign f(a) := g(a)
	// This is a partial application; xtra = 1 (needs 2 args, has 1)
	sortS := mkSort("S")
	fSort := il.RelationSort([]lg.Sort{sortS, sortS})
	fSym := lg.NewConst("f", fSort)
	gSym := lg.NewConst("g", fSort)
	aSym := lg.NewConst("a", sortS)

	// Construct Apply directly to bypass arity check (partial application)
	lhs := &lg.Apply{Func: fSym, Terms: []lg.Expr{aSym}}
	rhs := &lg.Apply{Func: gSym, Terms: []lg.Expr{aSym}}
	a := NewAssignAction(lhs, rhs)
	ctx := testCtx()
	u := a.ActionUpdate(ctx)

	// Should still produce a valid update (not null)
	if len(u.Modified) != 1 || u.Modified[0].Name != "f" {
		t.Errorf("Partial application assign should modify [f], got %v", u.Modified)
	}
}

func TestAssignAction_VariableCheck(t *testing.T) {
	// f(X) := g(Y) where Y is not in LHS — should return null update
	sortS := mkSort("S")
	fSort := il.RelationSort([]lg.Sort{sortS})
	fSym := lg.NewConst("f", fSort)
	gSym := lg.NewConst("g", fSort)
	xVar, err := lg.NewVariable("X", sortS)
	if err != nil {
		t.Fatalf("NewVariable X: %v", err)
	}
	yVar, err := lg.NewVariable("Y", sortS)
	if err != nil {
		t.Fatalf("NewVariable Y: %v", err)
	}

	lhs, _ := lg.NewApply(fSym, xVar)
	rhs, _ := lg.NewApply(gSym, yVar)
	a := NewAssignAction(lhs, rhs)
	ctx := testCtx()
	u := a.ActionUpdate(ctx)

	// Should return null update (multiply assigned)
	if len(u.Modified) != 0 {
		t.Errorf("Variable check should reject: RHS var not in LHS, got modified=%v", u.Modified)
	}
}

// ---------------------------------------------------------------------------
// Helper: testCtx (reuse from existing tests or define here)
// ---------------------------------------------------------------------------

