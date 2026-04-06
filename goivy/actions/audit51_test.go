package actions

// Tests for all 15 fixes from AUDIT18MARCH.md §5.1 (MISSING items).

import (
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy/ast"
	il "github.com/glycerine/ivy/goivy/ivylogic"
	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/module"
)

// ---------------------------------------------------------------------------
// Item 1: DebugAction in IntUpdate dispatch
// ---------------------------------------------------------------------------

func TestDebugAction_ActionUpdate_ReturnsNullUpdate(t *testing.T) {
	dbg := NewDebugAction(lg.NewConst("x", lg.TopS))
	ctx := testCtx()
	u := dbg.IntUpdate(ctx)

	if len(u.Modified) != 0 {
		t.Errorf("DebugAction should modify nothing, got %v", u.Modified)
	}
	if !u.TR.IsTrue() {
		t.Error("DebugAction TR should be true (null update)")
	}
}

func TestDebugAction_IntUpdate_Dispatch(t *testing.T) {
	// Verify DebugAction dispatches correctly through IntUpdate (not default)
	dbg := NewDebugAction(lg.NewConst("x", lg.TopS))
	ctx := testCtx()
	u := IntUpdate(dbg, ctx)

	if len(u.Modified) != 0 {
		t.Errorf("IntUpdate(DebugAction) should modify nothing, got %v", u.Modified)
	}
}

// ---------------------------------------------------------------------------
// Item 2: Field action ActionUpdate methods
// ---------------------------------------------------------------------------

// mkBinaryRelSort creates a sort for f: A x B -> Boolean.
func mkBinaryRelSort(dom1, dom2 lg.Sort) lg.Sort {
	return il.RelationSort([]lg.Sort{dom1, dom2})
}

func TestAssignFieldAction_ActionUpdate(t *testing.T) {
	// field(obj, X) := (X = val)
	sortT := mkSort("T")
	sortS := mkSort("S")
	fldSort := mkBinaryRelSort(sortT, sortS)
	fldSym := lg.NewConst("fld", fldSort)
	obj := lg.NewConst("obj", sortT)
	val := lg.NewConst("val", sortS)

	a := NewAssignFieldAction(fldSym, obj, val)
	ctx := testCtx()
	u := a.ActionUpdate(ctx)

	if len(u.Modified) != 1 || u.Modified[0].Name != "fld" {
		t.Errorf("AssignFieldAction should modify [fld], got %v", u.Modified)
	}
	trStr := u.TR.String()
	if !strings.Contains(trStr, "new_fld") {
		t.Errorf("TR should reference new_fld, got: %s", trStr)
	}
}

func TestNullFieldAction_ActionUpdate(t *testing.T) {
	sortT := mkSort("T")
	sortS := mkSort("S")
	fldSort := mkBinaryRelSort(sortT, sortS)
	fldSym := lg.NewConst("fld", fldSort)
	obj := lg.NewConst("obj", sortT)

	a := NewNullFieldAction(fldSym, obj)
	ctx := testCtx()
	u := a.ActionUpdate(ctx)

	if len(u.Modified) != 1 || u.Modified[0].Name != "fld" {
		t.Errorf("NullFieldAction should modify [fld], got %v", u.Modified)
	}
}

func TestCopyFieldAction_ActionUpdate(t *testing.T) {
	sortT := mkSort("T")
	sortS := mkSort("S")
	fldSort := mkBinaryRelSort(sortT, sortS)
	fldSym := lg.NewConst("fld", fldSort)
	dst := lg.NewConst("dst", sortT)
	src := lg.NewConst("src", sortT)

	a := NewCopyFieldAction(dst, fldSym, src, fldSym)
	ctx := testCtx()
	u := a.ActionUpdate(ctx)

	if len(u.Modified) != 1 || u.Modified[0].Name != "fld" {
		t.Errorf("CopyFieldAction should modify [fld], got %v", u.Modified)
	}
}

// ---------------------------------------------------------------------------
// Item 3: makeFieldUpdateFunc helper
// ---------------------------------------------------------------------------

func TestMakeFieldUpdateFunc_NonRelationalSort(t *testing.T) {
	// If the field doesn't have a binary relation sort, should return NullUpdate
	fldSym := lg.NewConst("fld", lg.TopS)
	obj := lg.NewConst("obj", lg.TopS)
	ctx := testCtx()

	u := makeFieldUpdateFunc(fldSym, obj, func(v *lg.Variable) lg.Expr {
		return v
	}, ctx)

	if len(u.Modified) != 0 {
		t.Errorf("Non-relational field should return null update, got modified=%v", u.Modified)
	}
}

func TestMakeFieldUpdateFunc_NilField(t *testing.T) {
	ctx := testCtx()
	u := makeFieldUpdateFunc(nil, nil, func(v *lg.Variable) lg.Expr {
		return v
	}, ctx)

	if len(u.Modified) != 0 {
		t.Errorf("nil field should return null update")
	}
}

// ---------------------------------------------------------------------------
// Item 4: IfAction.Subactions(NewActionsConfig()) and SomeCondition
// ---------------------------------------------------------------------------

func TestIfAction_Subactions_BoolCondition(t *testing.T) {
	cond := lg.NewConst("c", lg.Boolean)
	thenBody := NewAssumeAction(lg.NewConst("p", lg.Boolean))
	elseBody := NewAssumeAction(lg.NewConst("q", lg.Boolean))
	ifAct := NewIfAction(cond, thenBody, elseBody)

	ifPart, elsePart := ifAct.Subactions(NewActionsConfig())

	// ifPart should be Sequence(assume(c), thenBody)
	ifSeq, ok := ifPart.(*Sequence)
	if !ok {
		t.Fatalf("ifPart should be Sequence, got %T", ifPart)
	}
	if len(ifSeq.Elems) != 2 {
		t.Errorf("ifPart should have 2 children (assume+body), got %d", len(ifSeq.Elems))
	}

	// elsePart should be Sequence(assume(dual(c)), elseBody)
	elseSeq, ok := elsePart.(*Sequence)
	if !ok {
		t.Fatalf("elsePart should be Sequence, got %T", elsePart)
	}
	if len(elseSeq.Elems) != 2 {
		t.Errorf("elsePart should have 2 children, got %d", len(elseSeq.Elems))
	}
}

func TestIfAction_Subactions_NoElse(t *testing.T) {
	cond := lg.NewConst("c", lg.Boolean)
	thenBody := NewSequence()
	ifAct := NewIfAction(cond, thenBody)

	_, elsePart := ifAct.Subactions(NewActionsConfig())

	// elsePart should be a Sequence with assume(dual(c)) and an empty sequence
	elseSeq, ok := elsePart.(*Sequence)
	if !ok {
		t.Fatalf("elsePart should be Sequence, got %T", elsePart)
	}
	if len(elseSeq.Elems) != 2 {
		t.Errorf("elsePart should have 2 children (assume dual + empty), got %d", len(elseSeq.Elems))
	}
}

func TestSomeCondition_NodeSort(t *testing.T) {
	p := lg.NewConst("x", lg.TopS)
	fmla := lg.NewConst("f", lg.Boolean)
	sc := &SomeCondition{
		Params: []*lg.Const{p},
		Fmla:   fmla,
		Kind:   "some",
	}
	if sc.NodeSort() != lg.Boolean {
		t.Errorf("SomeCondition.NodeSort() should be Boolean, got %v", sc.NodeSort())
	}
	if !strings.Contains(sc.String(), "some") {
		t.Errorf("SomeCondition.String() should mention 'some', got %q", sc.String())
	}
}

func TestIfAction_Subactions_SomeCondition(t *testing.T) {
	p := lg.NewConst("x", lg.TopS)
	fmla := lg.NewConst("f", lg.Boolean)
	some := &SomeCondition{
		Params: []*lg.Const{p},
		Fmla:   fmla,
		Kind:   "some",
	}
	thenBody := NewSequence()
	elseBody := NewSequence()
	ifAct := NewIfAction(some, thenBody, elseBody)

	ifPart, elsePart := ifAct.Subactions(NewActionsConfig())

	// Both parts should be non-nil actions
	if ifPart == nil {
		t.Fatal("ifPart should not be nil for SomeCondition")
	}
	if elsePart == nil {
		t.Fatal("elsePart should not be nil for SomeCondition")
	}
}

// ---------------------------------------------------------------------------
// Item 5: IfAction.GetCond()
// ---------------------------------------------------------------------------

func TestIfAction_GetCond_BoolCondition(t *testing.T) {
	cond := lg.NewConst("c", lg.Boolean)
	ifAct := NewIfAction(cond, NewSequence())

	got := ifAct.GetCond()
	if got != cond {
		t.Errorf("GetCond() for bool condition should return cond itself, got %v", got)
	}
}

func TestIfAction_GetCond_SomeCondition(t *testing.T) {
	p := lg.NewConst("x", lg.TopS)
	fmla := lg.NewConst("f", lg.Boolean)
	some := &SomeCondition{
		Params: []*lg.Const{p},
		Fmla:   fmla,
		Kind:   "some",
	}
	ifAct := NewIfAction(some, NewSequence())

	got := ifAct.GetCond()

	// Should return an Exists formula
	if _, ok := got.(*lg.Exists); !ok {
		t.Errorf("GetCond() with SomeCondition should return Exists, got %T: %s", got, got)
	}
}

// ---------------------------------------------------------------------------
// Item 6: CallAction.SplitReturns(NewActionsConfig())
// ---------------------------------------------------------------------------

func TestCallAction_SplitReturns_NoReturns(t *testing.T) {
	callee := lg.NewConst("act", lg.TopS)
	call := NewCallActionOn(NewActionsConfig(), callee)

	result := call.SplitReturns(NewActionsConfig())

	// With no returns, should return self unchanged
	if result != call {
		t.Error("SplitReturns with no returns should return self")
	}
}

func TestCallAction_SplitReturns_WithReturns(t *testing.T) {
	callee := lg.NewConst("act", lg.TopS)
	ret1 := lg.NewConst("r1", lg.TopS)
	call := NewCallActionOn(NewActionsConfig(), callee, ret1)

	result := call.SplitReturns(NewActionsConfig())

	// Result should be a LocalAction wrapping a Sequence
	local, ok := result.(*LocalAction)
	if !ok {
		t.Fatalf("SplitReturns should return LocalAction, got %T", result)
	}
	// The LocalAction should contain a sequence with the call + assignments
	if local.Body == nil {
		t.Fatal("LocalAction body should not be nil")
	}
}

// ---------------------------------------------------------------------------
// Item 7: PrefixCallsFunc callable renamer
// ---------------------------------------------------------------------------

func TestPrefixCalls_StringPrefix(t *testing.T) {
	callee := lg.NewConst("myaction", lg.TopS)
	call := NewCallActionOn(NewActionsConfig(), callee)

	result := PrefixCalls(call, "mod.")
	callResult, ok := result.(*CallAction)
	if !ok {
		t.Fatalf("PrefixCalls should return CallAction, got %T", result)
	}
	if sym, ok := callResult.Callee.(*lg.Const); ok {
		if sym.Name != "mod.myaction" {
			t.Errorf("Expected 'mod.myaction', got %q", sym.Name)
		}
	} else {
		t.Fatal("Callee should be a Symbol")
	}
}

func TestPrefixCallsFunc_Callable(t *testing.T) {
	callee := lg.NewConst("myaction", lg.TopS)
	call := NewCallActionOn(NewActionsConfig(), callee)

	result := PrefixCallsFunc(call, func(name string) string {
		return strings.ToUpper(name)
	})
	callResult, ok := result.(*CallAction)
	if !ok {
		t.Fatalf("PrefixCallsFunc should return CallAction, got %T", result)
	}
	if sym, ok := callResult.Callee.(*lg.Const); ok {
		if sym.Name != "MYACTION" {
			t.Errorf("Expected 'MYACTION', got %q", sym.Name)
		}
	}
}

func TestPrefixCallsFunc_Nested(t *testing.T) {
	// PrefixCallsFunc should recurse into sub-actions
	callee := lg.NewConst("inner", lg.TopS)
	call := NewCallActionOn(NewActionsConfig(), callee)
	seq := NewSequence(call)

	result := PrefixCallsFunc(seq, func(name string) string {
		return "prefix_" + name
	})

	// Walk into the result sequence to check the renamed call
	seqResult, ok := result.(*Sequence)
	if !ok {
		t.Fatalf("Should return Sequence, got %T", result)
	}
	if len(seqResult.Elems) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(seqResult.Elems))
	}
	inner, _ := seqResult.Elems[0].(Action)
	if inner == nil {
		t.Fatal("inner action should not be nil")
	}
	innerCall, ok := inner.(*CallAction)
	if !ok {
		t.Fatalf("inner should be CallAction, got %T", inner)
	}
	if sym, ok := innerCall.Callee.(*lg.Const); ok {
		if sym.Name != "prefix_inner" {
			t.Errorf("Expected 'prefix_inner', got %q", sym.Name)
		}
	}
}

func TestPrefixCallsFunc_PreservesReturns(t *testing.T) {
	callee := lg.NewConst("act", lg.TopS)
	ret := lg.NewConst("r", lg.TopS)
	call := NewCallActionOn(NewActionsConfig(), callee, ret)

	result := PrefixCallsFunc(call, func(name string) string {
		return "ns." + name
	})
	callResult := result.(*CallAction)
	if len(callResult.ActualReturns) != 1 {
		t.Errorf("PrefixCallsFunc should preserve ActualReturns, got %d", len(callResult.ActualReturns))
	}
}

func TestPrefixCalls_Nil(t *testing.T) {
	if PrefixCalls(nil, "x.") != nil {
		t.Error("PrefixCalls(nil) should return nil")
	}
	call := NewCallActionOn(NewActionsConfig(), lg.NewConst("a", lg.TopS))
	if PrefixCalls(call, "") != call {
		t.Error("PrefixCalls with empty prefix should return action unchanged")
	}
}

// ---------------------------------------------------------------------------
// Item 8: IterInternalDefines and InternalDefine
// ---------------------------------------------------------------------------

func TestIterInternalDefines_NoThunks(t *testing.T) {
	seq := NewSequence(NewAssumeAction(lg.NewConst("p", lg.Boolean)))
	defs := IterInternalDefines(seq)
	if len(defs) != 0 {
		t.Errorf("No thunks means no internal defines, got %d", len(defs))
	}
}

func TestIterInternalDefines_ThunkAction(t *testing.T) {
	thunk := NewThunkAction(lg.NewConst("handler", lg.TopS), lg.NewConst("body", lg.TopS))
	thunk.SetLineno(ast.Location{Filename: "test.ivy", Line: 42})

	defs := IterInternalDefines(thunk)

	// ThunkAction should yield 2 defines: "handler" and "handler.run"
	if len(defs) != 2 {
		t.Fatalf("ThunkAction should yield 2 internal defines, got %d", len(defs))
	}
	if defs[0].Name != "handler" {
		t.Errorf("First define should be 'handler', got %q", defs[0].Name)
	}
	if !strings.HasSuffix(defs[1].Name, "run") {
		t.Errorf("Second define should end with 'run', got %q", defs[1].Name)
	}
	if defs[0].Lineno.Line != 42 {
		t.Errorf("Lineno should be 42, got %d", defs[0].Lineno.Line)
	}
}

func TestIterInternalDefines_NestedThunk(t *testing.T) {
	thunk := NewThunkAction(lg.NewConst("h", lg.TopS))
	seq := NewSequence(thunk)

	defs := IterInternalDefines(seq)

	// Should find the thunk's defines by recursing
	if len(defs) != 2 {
		t.Errorf("Should find nested thunk defines, got %d", len(defs))
	}
}

func TestIterInternalDefines_Nil(t *testing.T) {
	defs := IterInternalDefines(nil)
	if defs != nil {
		t.Errorf("nil action should return nil, got %v", defs)
	}
}

// ---------------------------------------------------------------------------
// Item 9: GetTypeNames
// ---------------------------------------------------------------------------

func TestGetTypeNames_NoLocals(t *testing.T) {
	seq := NewSequence(NewAssumeAction(lg.NewConst("p", lg.Boolean)))
	names := make(map[string]bool)
	GetTypeNames(seq, names)
	if len(names) != 0 {
		t.Errorf("No locals means no type names, got %v", names)
	}
}

func TestGetTypeNames_WithLocal(t *testing.T) {
	sortT := mkSort("T")
	localDecl := lg.NewConst("v", sortT)
	body := NewSequence()
	local := NewLocalActionOn(NewActionsConfig(), "test", localDecl, body)
	seq := NewSequence(local)

	names := make(map[string]bool)
	GetTypeNames(seq, names)

	// Should collect sort name "T" from the local declaration
	if !names["T"] {
		t.Errorf("Should collect type name 'T' from local declaration, got %v", names)
	}
}

func TestGetTypeNames_Nil(t *testing.T) {
	names := make(map[string]bool)
	GetTypeNames(nil, names)
	if len(names) != 0 {
		t.Errorf("nil action should add no names, got %v", names)
	}
}

// ---------------------------------------------------------------------------
// Item 10: Schema.GetInstance() and Instantiate
// ---------------------------------------------------------------------------

func TestSchema_GetInstance(t *testing.T) {
	// Schema with definition: f(a) = body where body references a
	aSym := lg.NewConst("a", lg.TopS)
	fSym := lg.NewConst("f", lg.TopS)
	lhs := &lg.Apply{Func: fSym, Terms: []lg.Expr{aSym}}
	rhs := aSym // body is just "a"
	defn := lg.NewDefinition(lhs, rhs)
	s := NewSchema(defn)

	// Instantiate with param "val"
	valSym := lg.NewConst("val", lg.TopS)
	result, err := s.GetInstance([]lg.Expr{valSym}, false)
	if err != nil {
		t.Fatalf("GetInstance failed: %v", err)
	}

	// Result should be "val" (the body with "a" substituted for "val")
	if sym, ok := result.(*lg.Const); ok {
		if sym.Name != "val" {
			t.Errorf("Expected substituted body to be 'val', got %q", sym.Name)
		}
	} else {
		t.Errorf("Expected Symbol, got %T: %s", result, result)
	}
}

func TestSchema_GetInstance_ParamCountMismatch(t *testing.T) {
	aSym := lg.NewConst("a", lg.TopS)
	fSym := lg.NewConst("f", lg.TopS)
	lhs := &lg.Apply{Func: fSym, Terms: []lg.Expr{aSym}}
	defn := lg.NewDefinition(lhs, aSym)
	s := NewSchema(defn)

	// Wrong number of params
	_, err := s.GetInstance([]lg.Expr{}, false)
	if err == nil {
		t.Error("GetInstance should fail with wrong param count")
	}
}

func TestSchema_Instantiate(t *testing.T) {
	aSym := lg.NewConst("a", lg.TopS)
	fSym := lg.NewConst("f", lg.TopS)
	lhs := &lg.Apply{Func: fSym, Terms: []lg.Expr{aSym}}
	defn := lg.NewDefinition(lhs, aSym)
	s := NewSchema(defn)

	// Instantiate should add to Instances
	valSym := lg.NewConst("val", lg.TopS)
	s.Instantiate([]lg.Expr{valSym})

	if len(s.Instances) != 1 {
		t.Fatalf("Instantiate should add 1 instance, got %d", len(s.Instances))
	}
}

// ---------------------------------------------------------------------------
// Item 11: TypeCheckContext — already correct, verify basic behavior
// ---------------------------------------------------------------------------

func TestTypeCheckContext_Get_ReturnsEmptyWithFormals(t *testing.T) {
	mod := mkTestModule()
	// Add an action to the module
	act := NewSequence()
	act.SetFormalParams([]*lg.Const{lg.NewConst("x", lg.TopS)})
	act.SetFormalReturns([]*lg.Const{lg.NewConst("r", lg.TopS)})
	mod.Actions.Set("myact", act)

	tc := NewTypeCheckContext(mod)
	result := tc.Get("myact")

	if result == nil {
		t.Fatal("TypeCheckContext.Get should return an action")
	}
	// Should be a Sequence (empty) but with same formals
	seq, ok := result.(*Sequence)
	if !ok {
		t.Fatalf("Should return Sequence, got %T", result)
	}
	if len(seq.Elems) != 0 {
		t.Error("TypeCheckContext should return empty action")
	}
	if len(result.GetFormalParams()) != 1 || result.GetFormalParams()[0].Name != "x" {
		t.Error("Should preserve formal params")
	}
	if len(result.GetFormalReturns()) != 1 || result.GetFormalReturns()[0].Name != "r" {
		t.Error("Should preserve formal returns")
	}
}

func TestTypeCheckContext_Get_NotFound(t *testing.T) {
	mod := mkTestModule()
	tc := NewTypeCheckContext(mod)

	result := tc.Get("nonexistent")
	if result != nil {
		t.Error("Should return nil for unknown action")
	}
}

// ---------------------------------------------------------------------------
// Item 12: checked_assert / check_unprovable in AssertAction.ActionUpdate
// ---------------------------------------------------------------------------

func TestAssertAction_ActionUpdate_CheckUnprovableMismatch(t *testing.T) {
	fmla := lg.NewConst("p", lg.Boolean)
	a := NewAssertAction(fmla)
	a.Unprovable = false

	ctx := testCtx()
	ctx.CheckUnprovable = true // mismatch: checking unprovable but assertion is provable

	u := a.ActionUpdate(ctx)

	// Should skip: TR=true, Pre=false
	if !u.TR.IsTrue() {
		t.Error("Mismatch should skip: TR should be true")
	}
	if !u.Pre.IsFalse() {
		t.Error("Mismatch should skip: Pre should be false")
	}
}

func TestAssertAction_ActionUpdate_CheckedAssert_Selected(t *testing.T) {
	fmla := lg.NewConst("p", lg.Boolean)
	a := NewAssertAction(fmla)
	a.SetLineno(ast.Location{Filename: "test.ivy", Line: 42})

	ctx := testCtx()
	ctx.CheckedAssert = a.GetLineno().String() // matches this assertion

	u := a.ActionUpdate(ctx)

	// Selected assertion should get dual formula treatment
	if u.Pre == nil || u.Pre.IsFalse() {
		t.Error("Selected assertion should have non-false Pre (dual formula)")
	}
}

func TestAssertAction_ActionUpdate_CheckedAssert_NotSelected_Provable(t *testing.T) {
	fmla := lg.NewConst("p", lg.Boolean)
	a := NewAssertAction(fmla)
	a.SetLineno(ast.Location{Filename: "test.ivy", Line: 42})
	a.Unprovable = false

	ctx := testCtx()
	ctx.CheckedAssert = "other.ivy:99" // different assertion

	u := a.ActionUpdate(ctx)

	// Not selected, provable: return formula as-is (not dual), Pre=false
	if u.TR.IsFalse() {
		t.Error("Not-selected provable assertion should have non-false TR (formula)")
	}
	if !u.Pre.IsFalse() {
		t.Error("Not-selected provable assertion should have false Pre")
	}
}

func TestAssertAction_ActionUpdate_CheckedAssert_NotSelected_Unprovable(t *testing.T) {
	fmla := lg.NewConst("p", lg.Boolean)
	a := NewAssertAction(fmla)
	a.SetLineno(ast.Location{Filename: "test.ivy", Line: 42})
	a.Unprovable = true

	ctx := testCtx()
	ctx.CheckUnprovable = true         // match the unprovable flag
	ctx.CheckedAssert = "other.ivy:99" // different assertion

	u := a.ActionUpdate(ctx)

	// Not selected, unprovable: skip entirely
	if !u.TR.IsTrue() {
		t.Error("Not-selected unprovable assertion should have true TR")
	}
	if !u.Pre.IsFalse() {
		t.Error("Not-selected unprovable assertion should have false Pre")
	}
}

// ---------------------------------------------------------------------------
// Item 13: Unprovable field on AssumeAction/AssertAction
// ---------------------------------------------------------------------------

func TestAssumeAction_Unprovable_Skips(t *testing.T) {
	fmla := lg.NewConst("p", lg.Boolean)
	a := NewAssumeAction(fmla)
	a.Unprovable = true

	ctx := testCtx()
	u := a.ActionUpdate(ctx)

	// Unprovable assume should produce TR=true (skip)
	if !u.TR.IsTrue() {
		t.Error("Unprovable AssumeAction should produce true TR (skip)")
	}
}

func TestAssumeAction_Provable_Normal(t *testing.T) {
	fmla := lg.NewConst("p", lg.Boolean)
	a := NewAssumeAction(fmla)
	a.Unprovable = false

	ctx := testCtx()
	u := a.ActionUpdate(ctx)

	// Normal assume should include the formula in TR
	if u.TR.IsTrue() {
		t.Error("Normal AssumeAction should have non-trivial TR")
	}
}

func TestAssertAction_Unprovable_Field(t *testing.T) {
	fmla := lg.NewConst("p", lg.Boolean)
	a := NewAssertAction(fmla)
	a.Unprovable = true

	// Verify the field exists and is propagated through clone
	cloned := a.ActionClone(a.ActionArgs())
	if assertCloned, ok := cloned.(*AssertAction); ok {
		if !assertCloned.Unprovable {
			t.Error("ActionClone should preserve Unprovable field")
		}
	}
}

// ---------------------------------------------------------------------------
// Item 14: ActionContext with Enter/Exit/Get using ActionsConfig
// ---------------------------------------------------------------------------

func TestActionContext_EnterExit(t *testing.T) {
	cfg := NewActionsConfig()

	mod := mkTestModule()
	ctx := NewActionContextOn(mod, cfg)
	original := cfg.Context
	ctx.Enter()

	if cfg.Context != ctx {
		t.Error("Enter() should set cfg.Context to this context")
	}

	ctx.Exit()

	if cfg.Context != original {
		t.Error("Exit() should restore original context")
	}
}

func TestActionContext_NestedEnterExit(t *testing.T) {
	cfg := NewActionsConfig()
	original := cfg.Context

	ctx1 := NewActionContextOn(&module.Module{Name: "domain1"}, cfg)
	ctx2 := NewActionContextOn(&module.Module{Name: "domain2"}, cfg)

	ctx1.Enter()
	if cfg.Context != ctx1 {
		t.Error("After ctx1.Enter(), cfg.Context should be ctx1")
	}

	ctx2.Enter()
	if cfg.Context != ctx2 {
		t.Error("After ctx2.Enter(), cfg.Context should be ctx2")
	}

	ctx2.Exit()
	if cfg.Context != ctx1 {
		t.Error("After ctx2.Exit(), cfg.Context should be ctx1")
	}

	ctx1.Exit()
	if cfg.Context != original {
		t.Error("After ctx1.Exit(), cfg.Context should be original")
	}
}

func TestRunWithActionContext(t *testing.T) {
	cfg := NewActionsConfig()
	original := cfg.Context
	mod := mkTestModule()
	ctx := NewActionContextOn(mod, cfg)

	var insideCtx IActionContext
	RunWithActionContext(ctx, func() {
		insideCtx = cfg.Context
	})

	if insideCtx != ctx {
		t.Error("Inside RunWithActionContext, cfg.Context should be the provided context")
	}
	if cfg.Context != original {
		t.Error("After RunWithActionContext, cfg.Context should be restored")
	}
}

func TestActionContext_GetDomain(t *testing.T) {
	mod := mkTestModule()
	ctx := NewActionContext(mod)
	if ctx.GetDomain() != mod {
		t.Error("GetDomain should return the domain")
	}
}

func TestActionContext_Get_WithModule(t *testing.T) {
	mod := mkTestModule()
	act := NewSequence()
	mod.Actions.Set("test_action", act)

	ctx := NewActionContext(mod)
	result := ctx.Get("test_action")

	if result == nil {
		t.Error("Get should find the action in the module")
	}
}

func TestActionContext_Get_NotFound(t *testing.T) {
	mod := mkTestModule()
	ctx := NewActionContext(mod)
	result := ctx.Get("nonexistent")
	if result != nil {
		t.Error("Get should return nil for unknown action")
	}
}

// ---------------------------------------------------------------------------
// Item 15: SymExContext with Enter/Exit and SymexParams global
// ---------------------------------------------------------------------------

func TestSymExContext_EnterExit(t *testing.T) {
	cfg := NewActionsConfig()

	params := []lg.Expr{lg.NewConst("a", lg.TopS), lg.NewConst("b", lg.TopS)}
	ctx := NewSymExContext(cfg, params)

	ctx.Enter()
	if len(cfg.SymexParams) != 2 {
		t.Errorf("After Enter(), SymexParams should have 2 elements, got %d", len(cfg.SymexParams))
	}

	ctx.Exit()
	if cfg.SymexParams != nil {
		t.Error("After Exit(), SymexParams should be restored to nil")
	}
}

func TestSymExContext_NestedEnterExit(t *testing.T) {
	cfg := NewActionsConfig()

	params1 := []lg.Expr{lg.NewConst("x", lg.TopS)}
	params2 := []lg.Expr{lg.NewConst("y", lg.TopS), lg.NewConst("z", lg.TopS)}

	ctx1 := NewSymExContext(cfg, params1)
	ctx2 := NewSymExContext(cfg, params2)

	ctx1.Enter()
	if len(cfg.SymexParams) != 1 {
		t.Errorf("After ctx1.Enter(), expected 1 param, got %d", len(cfg.SymexParams))
	}

	ctx2.Enter()
	if len(cfg.SymexParams) != 2 {
		t.Errorf("After ctx2.Enter(), expected 2 params, got %d", len(cfg.SymexParams))
	}

	ctx2.Exit()
	if len(cfg.SymexParams) != 1 {
		t.Errorf("After ctx2.Exit(), expected 1 param, got %d", len(cfg.SymexParams))
	}

	ctx1.Exit()
	if cfg.SymexParams != nil {
		t.Error("After ctx1.Exit(), SymexParams should be nil")
	}
}

func TestRunWithSymExContext(t *testing.T) {
	cfg := NewActionsConfig()

	params := []lg.Expr{lg.NewConst("p", lg.TopS)}
	var insideParams []lg.Expr

	RunWithSymExContext(cfg, params, func() {
		insideParams = cfg.SymexParams
	})

	if len(insideParams) != 1 {
		t.Errorf("Inside RunWithSymExContext, SymexParams should have 1 element, got %d", len(insideParams))
	}
	if cfg.SymexParams != nil {
		t.Error("After RunWithSymExContext, SymexParams should be restored to nil")
	}
}

// Ensure unused imports are satisfied.
var _ = module.BoolConst
var _ = il.NewEqualsNode
