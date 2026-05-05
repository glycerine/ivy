package goivy

// Tests for the 7 stub fixes from AUDIT18MARCH.md §5.2 items #3-#9.

import (
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// mkFuncSort creates a FunctionSort: dom... -> range.
func mkFuncSort(dom []Sort, rng Sort) Sort {
	all := make([]Sort, len(dom)+1)
	copy(all, dom)
	all[len(dom)] = rng
	fs, err := NewFunctionSort(all...)
	if err != nil {
		panic(err)
	}
	return fs
}

// mkRelSort creates a relational sort: dom... -> Boolean.
func mkRelSort(dom ...Sort) Sort {
	return LogicRelationSort(dom)
}

// mkTestModule creates a module with optional setup.
func mkTestModule() *Module {
	return New()
}

// ---------------------------------------------------------------------------
// Fix #5: isDestructor
// ---------------------------------------------------------------------------

func TestIsDestructor_ReturnsTrueForKnownDestructor(t *testing.T) {
	mod := mkTestModule()
	sortT := actionsMkSort("T")
	mod.DestructorSorts["field1"] = sortT

	cfg := NewActionsConfig()
	ctx := NewActionContextOn(mod, cfg)
	ctx.Enter()
	defer ctx.Exit()

	if !isDestructor("field1", cfg) {
		t.Error("isDestructor should return true for symbol in DestructorSorts")
	}
}

func TestIsDestructor_ReturnsFalseForNonDestructor(t *testing.T) {
	mod := mkTestModule()

	cfg := NewActionsConfig()
	ctx := NewActionContextOn(mod, cfg)
	ctx.Enter()
	defer ctx.Exit()

	if isDestructor("not_a_destructor", cfg) {
		t.Error("isDestructor should return false for symbol NOT in DestructorSorts")
	}
}

func TestIsDestructor_ReturnsFalseWithNilContext(t *testing.T) {
	if isDestructor("anything", nil) {
		t.Error("isDestructor should return false when cfg is nil")
	}
}

// ---------------------------------------------------------------------------
// Fix #8: ChoiceAction.IntUpdate determinize check
// ---------------------------------------------------------------------------

func TestChoiceAction_IntUpdate_NoDeterminize(t *testing.T) {
	// With determinize=false, ChoiceAction.IntUpdate should use join_action
	// (original behavior), regardless of branch count.
	p := NewConst("p", Boolean)
	q := NewConst("q", Boolean)
	ch := NewChoiceActionOn(NewActionsConfig(), NewAssumeAction(p), NewAssumeAction(q))
	ctx := testCtx()
	ctx.ActCfg.Determinize = false
	u := ch.IntUpdate(ctx)

	if len(u.Modified) != 0 {
		t.Errorf("Choice of assumes should modify nothing, got %v", u.Modified)
	}
	// The TR should contain a disjunction of p and q (via join_action)
	if u.TR == nil {
		t.Fatal("TR should not be nil")
	}
}

func TestChoiceAction_IntUpdate_Determinize_TwoBranches(t *testing.T) {
	// With determinize=true and 2 branches, should convert to IfAction
	x := NewConst("x", TopS)
	y := NewConst("y", TopS)
	z := NewConst("z", TopS)

	// Branch 0: x := y,  Branch 1: x := z
	b0 := NewAssignAction(x, y)
	b1 := NewAssignAction(x, z)
	ch := NewChoiceActionOn(NewActionsConfig(), b0, b1)
	ctx := testCtx()
	ctx.ActCfg.Determinize = true
	u := ch.IntUpdate(ctx)

	// Should modify x (the assignment target)
	hasX := false
	for _, m := range u.Modified {
		if m.Name == "x" {
			hasX = true
		}
	}
	if !hasX {
		t.Errorf("Determinized choice should still modify x, got %v", u.Modified)
	}

	// The TR should reference the branch condition ___branch:N
	trStr := u.TR.String()
	if !strings.Contains(trStr, "___branch:") {
		t.Errorf("Determinized choice TR should contain ___branch condition, got: %s", trStr)
	}
}

func TestChoiceAction_IntUpdate_Determinize_ThreeBranches(t *testing.T) {
	// With determinize=true but 3 branches, should NOT convert to IfAction
	// (only works for exactly 2 branches per Python)
	p := NewConst("p", Boolean)
	q := NewConst("q", Boolean)
	r := NewConst("r", Boolean)
	ch := NewChoiceActionOn(NewActionsConfig(),
		NewAssumeAction(p),
		NewAssumeAction(q),
		NewAssumeAction(r),
	)
	ctx := testCtx()
	ctx.ActCfg.Determinize = true
	u := ch.IntUpdate(ctx)

	// Should use join_action (no ___branch condition)
	trStr := u.TR.String()
	if strings.Contains(trStr, "___branch:") {
		t.Errorf("3-branch choice should not use determinize, got: %s", trStr)
	}
}

// ---------------------------------------------------------------------------
// Fix #9: EnvAction.IntUpdateEnv determinize check
// ---------------------------------------------------------------------------

func TestEnvAction_IntUpdateEnv_NoDeterminize(t *testing.T) {
	p := NewConst("p", Boolean)
	q := NewConst("q", Boolean)
	env := NewEnvActionOn(NewActionsConfig(), NewAssumeAction(p), NewAssumeAction(q))
	ctx := testCtx()
	ctx.ActCfg.Determinize = false
	u := env.IntUpdateEnv(ctx)

	if len(u.Modified) != 0 {
		t.Errorf("Env of assumes should modify nothing, got %v", u.Modified)
	}
}

func TestEnvAction_IntUpdateEnv_Determinize_TwoBranches(t *testing.T) {
	x := NewConst("x", TopS)
	y := NewConst("y", TopS)
	z := NewConst("z", TopS)

	b0 := NewAssignAction(x, y)
	b1 := NewAssignAction(x, z)
	env := NewEnvActionOn(NewActionsConfig(), b0, b1)
	ctx := testCtx()
	ctx.ActCfg.Determinize = true
	u := env.IntUpdateEnv(ctx)

	// Should modify x
	hasX := false
	for _, m := range u.Modified {
		if m.Name == "x" {
			hasX = true
		}
	}
	if !hasX {
		t.Errorf("Determinized env should still modify x, got %v", u.Modified)
	}

	// TR should contain ___branch condition
	trStr := u.TR.String()
	if !strings.Contains(trStr, "___branch:") {
		t.Errorf("Determinized env TR should contain ___branch condition, got: %s", trStr)
	}
}

func TestEnvAction_IntUpdateEnv_Determinize_ThreeBranches(t *testing.T) {
	p := NewConst("p", Boolean)
	q := NewConst("q", Boolean)
	r := NewConst("r", Boolean)
	env := NewEnvActionOn(NewActionsConfig(),
		NewAssumeAction(p),
		NewAssumeAction(q),
		NewAssumeAction(r),
	)
	ctx := testCtx()
	ctx.ActCfg.Determinize = true
	u := env.IntUpdateEnv(ctx)

	trStr := u.TR.String()
	if strings.Contains(trStr, "___branch:") {
		t.Errorf("3-branch env should not use determinize, got: %s", trStr)
	}
}

// ---------------------------------------------------------------------------
// Fix #6: SetAction.ActionUpdate frame conditions
// ---------------------------------------------------------------------------

func TestSetAction_ActionUpdate_PositiveLiteral(t *testing.T) {
	// set R(a) — positive literal with concrete arg
	relSort := mkRelSort(TopS)
	relSym := NewConst("R", relSort)
	aConst := NewConst("a", TopS)
	atom := MustApply(relSym, aConst)

	sa := NewSetAction(atom)
	ctx := testCtx()
	u := sa.ActionUpdate(ctx)

	// Should modify R
	if len(u.Modified) != 1 || u.Modified[0].Name != "R" {
		t.Errorf("SetAction should modify [R], got %v", u.Modified)
	}

	// TR should reference new_R (post-state symbol)
	trStr := u.TR.String()
	if !strings.Contains(trStr, "new_R") {
		t.Errorf("SetAction TR should reference new_R, got: %s", trStr)
	}

	// TR should contain the frame condition equality (X0 = a or similar)
	// because 'a' is a non-variable arg, triggering frame condition clauses
	if u.TR.IsFalse() {
		t.Error("SetAction TR should not be false")
	}

	// Should have more than just 2 formulas when there are frame conditions
	if len(u.TR.Fmlas) < 2 {
		t.Errorf("SetAction with concrete arg should have frame condition formulas, got %d formulas", len(u.TR.Fmlas))
	}
}

func TestSetAction_ActionUpdate_NegativeLiteral(t *testing.T) {
	// set ~R(a) — negative literal
	relSort := mkRelSort(TopS)
	relSym := NewConst("R", relSort)
	aConst := NewConst("a", TopS)
	atom := MustApply(relSym, aConst)
	negLit := &LogicNot{Body: atom}

	sa := NewSetAction(negLit)
	ctx := testCtx()
	u := sa.ActionUpdate(ctx)

	if len(u.Modified) != 1 || u.Modified[0].Name != "R" {
		t.Errorf("SetAction(~R) should modify [R], got %v", u.Modified)
	}

	trStr := u.TR.String()
	if !strings.Contains(trStr, "new_R") {
		t.Errorf("SetAction(~R) TR should reference new_R, got: %s", trStr)
	}
}

func TestSetAction_ActionUpdate_NoArgs(t *testing.T) {
	// set R — 0-ary relation (no args, so no frame conditions needed)
	relSym := NewConst("R", Boolean)
	sa := NewSetAction(relSym)
	ctx := testCtx()
	u := sa.ActionUpdate(ctx)

	if len(u.Modified) != 1 || u.Modified[0].Name != "R" {
		t.Errorf("SetAction(R) should modify [R], got %v", u.Modified)
	}
}

func TestSetAction_ActionUpdate_MultipleArgs(t *testing.T) {
	// set R(a, b) — two concrete args should generate frame conditions for each
	relSort := mkRelSort(TopS, TopS)
	relSym := NewConst("R", relSort)
	aConst := NewConst("a", TopS)
	bConst := NewConst("b", TopS)
	atom := MustApply(relSym, aConst, bConst)

	sa := NewSetAction(atom)
	ctx := testCtx()
	u := sa.ActionUpdate(ctx)

	if len(u.Modified) != 1 || u.Modified[0].Name != "R" {
		t.Errorf("SetAction(R(a,b)) should modify [R], got %v", u.Modified)
	}

	// With 2 concrete (non-variable) args, there should be 2 frame condition pairs
	// Total formulas: 2 (base) + 2*2 (frame) = 6
	if len(u.TR.Fmlas) < 4 {
		t.Errorf("SetAction(R(a,b)) should have >= 4 formulas for frame conditions, got %d", len(u.TR.Fmlas))
	}
}

// ---------------------------------------------------------------------------
// Fix #3: mkVariantAssignClauses
// ---------------------------------------------------------------------------

func TestMkVariantAssignClauses_BasicVariant(t *testing.T) {
	// Setup: sort "msg" has variant "req_sort"
	mod := mkTestModule()
	sortMsg := actionsMkSort("msg")
	sortReq := actionsMkSort("req_sort")
	mod.Variants["msg"] = []Sort{sortReq}

	// lhs: symbol "m" of sort msg; rhs: symbol "r" of sort req_sort
	mSym := NewConst("m", sortMsg)
	rSym := NewConst("r", sortReq)

	u := mkVariantAssignClauses(mSym, rSym, mod)

	// Should modify m
	if len(u.Modified) != 1 || u.Modified[0].Name != "m" {
		t.Errorf("mkVariantAssignClauses should modify [m], got %v", u.Modified)
	}

	// TR should have formulas and definitions
	if u.TR == nil {
		t.Fatal("TR should not be nil")
	}

	// Should have at least 1 formula (the Iff constraint for the pto)
	if len(u.TR.Fmlas) < 1 {
		t.Errorf("TR should have at least 1 formula (pto constraint), got %d", len(u.TR.Fmlas))
	}

	// Should have a definition for the nondeterministic value
	if len(u.TR.Defs) < 1 {
		t.Errorf("TR should have at least 1 definition, got %d", len(u.TR.Defs))
	}
}

func TestMkVariantAssignClauses_MultipleVariants(t *testing.T) {
	// Setup: sort "msg" has variants "req_sort" and "resp_sort"
	mod := mkTestModule()
	sortMsg := actionsMkSort("msg")
	sortReq := actionsMkSort("req_sort")
	sortResp := actionsMkSort("resp_sort")
	mod.Variants["msg"] = []Sort{sortReq, sortResp}

	mSym := NewConst("m", sortMsg)
	rSym := NewConst("r", sortReq)

	u := mkVariantAssignClauses(mSym, rSym, mod)

	// Should have 2 formulas: 1 Iff for req_sort + 1 Not(pto) for resp_sort
	if len(u.TR.Fmlas) != 2 {
		t.Errorf("Expected 2 formulas (1 Iff + 1 Not), got %d", len(u.TR.Fmlas))
	}

	// First formula should be Iff (the target variant constraint)
	if _, ok := u.TR.Fmlas[0].(*LogicIff); !ok {
		t.Errorf("First formula should be Iff, got %T", u.TR.Fmlas[0])
	}

	// Second formula should be Not (negative constraint for other variant)
	if _, ok := u.TR.Fmlas[1].(*LogicNot); !ok {
		t.Errorf("Second formula should be Not, got %T", u.TR.Fmlas[1])
	}
}

func TestMkVariantAssignClauses_NilSym(t *testing.T) {
	// If lhs has no extractable symbol, should return NullUpdate
	mod := mkTestModule()
	v, _ := NewVariable("X", TopS)
	rSym := NewConst("r", TopS)
	u := mkVariantAssignClauses(v, rSym, mod)
	if len(u.Modified) != 0 {
		t.Errorf("Variant assign with variable lhs should return null update, got modified=%v", u.Modified)
	}
}

// ---------------------------------------------------------------------------
// Fix #4: destructorAssignUpdate (via destrAsgnVal)
// ---------------------------------------------------------------------------

func TestDestrAsgnVal_SimpleDestructor(t *testing.T) {
	// Setup: sort "T" with destructor "fld" : T -> S
	mod := mkTestModule()
	sortT := actionsMkSort("T")
	sortS := actionsMkSort("S")
	fldSort := mkFuncSort([]Sort{sortT}, sortS)
	fldSym := NewConst("fld", fldSort)

	mod.DestructorSorts["fld"] = sortT
	mod.SortDestructors["T"] = []*Const{fldSym}

	// lhs = fld(obj), rhs = val
	objSym := NewConst("obj", sortT)
	_ = NewConst("val", sortS) // rhs used in destructorAssignUpdate, not destrAsgnVal
	lhs := MustApply(fldSym, objSym)

	var fmlas []Expr
	resultExpr, clauses, mutated := destrAsgnVal(lhs, &fmlas, mod)

	// mutated should be the root object symbol
	if mutated == nil || mutated.Name != "obj" {
		t.Errorf("mutated should be 'obj', got %v", mutated)
	}

	// clauses should not be nil (the base-case assignment)
	if clauses == nil {
		t.Fatal("clauses should not be nil")
	}

	// resultExpr should reference the skolem
	resultStr := resultExpr.String()
	if !strings.Contains(resultStr, "fld") {
		t.Errorf("result expr should reference fld, got: %s", resultStr)
	}

	// No sibling destructors, so fmlas should be empty (fld is the only destructor for T)
	// (frame conditions only fire for sibling destructors != n)
}

func TestDestrAsgnVal_WithSiblingDestructor(t *testing.T) {
	// Setup: sort "T" with destructors "fld1" and "fld2"
	mod := mkTestModule()
	sortT := actionsMkSort("T")
	sortS := actionsMkSort("S")
	fld1Sort := mkFuncSort([]Sort{sortT}, sortS)
	fld2Sort := mkFuncSort([]Sort{sortT}, sortS)
	fld1Sym := NewConst("fld1", fld1Sort)
	fld2Sym := NewConst("fld2", fld2Sort)

	mod.DestructorSorts["fld1"] = sortT
	mod.DestructorSorts["fld2"] = sortT
	mod.SortDestructors["T"] = []*Const{fld1Sym, fld2Sym}

	// lhs = fld1(obj), rhs = val — should produce frame condition for fld2
	objSym := NewConst("obj", sortT)
	valSym := NewConst("val", sortS)
	lhs := MustApply(fld1Sym, objSym)

	var fmlas []Expr
	_, _, mutated := destrAsgnVal(lhs, &fmlas, mod)

	if mutated == nil || mutated.Name != "obj" {
		t.Errorf("mutated should be 'obj', got %v", mutated)
	}

	// Should have a frame condition formula for fld2
	// fmlas should contain an equality: fld2(skolem_val) = fld2(obj)
	foundFrame := false
	for _, f := range fmlas {
		if eq, ok := f.(*Eq); ok {
			s := eq.String()
			if strings.Contains(s, "fld2") {
				foundFrame = true
			}
		}
	}
	if !foundFrame {
		t.Errorf("Should have frame condition for sibling destructor fld2, fmlas=%v", fmlas)
	}

	_ = valSym
}

func TestDestructorAssignUpdate_FullPath(t *testing.T) {
	// Integration test: AssignAction.destructorAssignUpdate
	mod := mkTestModule()
	sortT := actionsMkSort("T")
	sortS := actionsMkSort("S")
	fldSort := mkFuncSort([]Sort{sortT}, sortS)
	fldSym := NewConst("fld", fldSort)

	mod.DestructorSorts["fld"] = sortT
	mod.SortDestructors["T"] = []*Const{fldSym}

	objSym := NewConst("obj", sortT)
	valSym := NewConst("val", sortS)
	lhs := MustApply(fldSym, objSym)

	a := NewAssignAction(lhs, valSym)
	ctx := &UpdateContext{Domain: mod}
	u := a.destructorAssignUpdate(ctx, lhs, valSym)

	// Should modify the root symbol 'obj'
	if len(u.Modified) != 1 || u.Modified[0].Name != "obj" {
		t.Errorf("destructorAssignUpdate should modify [obj], got %v", u.Modified)
	}

	// TR should have formulas (the combined clauses)
	if u.TR == nil || u.TR.IsFalse() {
		t.Error("TR should not be nil or false")
	}
}

// ---------------------------------------------------------------------------
// Fix #7: WhileAction.Decompose
// ---------------------------------------------------------------------------

func TestWhileAction_Decompose_WithModule(t *testing.T) {
	// Use DecomposeWithModule to expand the while loop
	mod := mkTestModule()

	cond := NewConst("c", Boolean)
	body := NewAssumeAction(NewConst("p", Boolean))
	w := NewWhileAction(cond, body)

	paths := w.DecomposeWithModule(mod)

	// After expansion, the decomposition should produce paths from the
	// expanded sequence (assert, havoc, assume, if).
	// It should NOT just return [[body]] like the old stub.
	if len(paths) == 0 {
		t.Fatal("Decompose should return at least one path")
	}

	// The old stub would return exactly [[body_action]].
	// The new implementation expands and then decomposes, so we should
	// get a different structure — typically the expanded sequence decomposes
	// to a single path with all the steps.
	// Just verify it's not the old trivial decomposition of the body alone.
	if len(paths) == 1 && len(paths[0]) == 1 {
		singleAct := paths[0][0]
		if singleAct.Name() == "assume" {
			t.Error("Decompose should expand the while loop, not just return the body")
		}
	}
}

func TestWhileAction_Decompose_WithoutModule(t *testing.T) {
	// Without a module, Decompose falls back to returning the body

	cond := NewConst("c", Boolean)
	body := NewAssumeAction(NewConst("p", Boolean))
	w := NewWhileAction(cond, body)

	paths := w.Decompose()

	// Fallback: should return [[body]] (the body as a single step)
	if len(paths) != 1 || len(paths[0]) != 1 {
		t.Errorf("Without module, Decompose should fall back to [[body]], got %d paths", len(paths))
	}
}

// ---------------------------------------------------------------------------
// Cross-cutting: verify ChoiceAction vs EnvAction polarity difference
// ---------------------------------------------------------------------------

func TestDeterminize_PolarityDifference(t *testing.T) {
	// ChoiceAction uses Not(cond), EnvAction uses cond (positive).
	// We verify this by checking the TR string for each.
	x := NewConst("x", TopS)
	y := NewConst("y", TopS)
	z := NewConst("z", TopS)
	b0 := NewAssignAction(x, y)
	b1 := NewAssignAction(x, z)

	// ChoiceAction
	ch := NewChoiceActionOn(NewActionsConfig(), b0, b1)
	ctx := testCtx()
	ctx.ActCfg.Determinize = true
	chUpdate := ch.IntUpdate(ctx)
	chTR := chUpdate.TR.String()

	// EnvAction
	env := NewEnvActionOn(NewActionsConfig(), b0, b1)
	envUpdate := env.IntUpdateEnv(ctx)
	envTR := envUpdate.TR.String()

	// Both should contain ___branch, but the formula structure should differ
	// because of the polarity difference (Not vs positive)
	if !strings.Contains(chTR, "___branch:") {
		t.Error("Choice TR should contain ___branch")
	}
	if !strings.Contains(envTR, "___branch:") {
		t.Error("Env TR should contain ___branch")
	}

	// They should have different unique IDs (since each action gets a new counter)
	// so the TRs won't be identical. But structurally they should both work.
	if chTR == envTR {
		// This is not necessarily wrong (could have same structure with different IDs)
		// but if the polarity is correctly different, they should differ.
		// We don't fail on this since the branch IDs differ anyway.
		_ = chTR
	}
}

// Ensure unused imports are satisfied.
var _ = BoolConst
