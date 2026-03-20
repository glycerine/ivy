package compiler

import (
	"strings"
	"testing"

	"github.com/glycerine/goivy/actions"
	"github.com/glycerine/goivy/ast"
	lg "github.com/glycerine/goivy/logic"
)

// =============================================================================
// Batch B — Compilation Expression Fixes (RED phase)
//
// All tests expose missing/buggy behavior in the Go port relative to Python.
// They should ALL FAIL until the corresponding fixes are implemented.
// =============================================================================

// TestExpr6_CompileActionDef_PrmPrefixSubstitution tests that CompileAction
// renames formal parameters with a "prm:" prefix to avoid name collisions,
// matching Python's compile_action_def which does v.to_const('prm:') +
// substitute_ast(a, subst).
//
// §6.3 #15: Go compiles formals directly without any prm: prefix renaming.
func TestExpr6_CompileActionDef_PrmPrefixSubstitution(t *testing.T) {
	c := newTestCompiler()

	// Set up a sort and a symbol so compilation succeeds.
	natSort := &lg.UninterpretedSort{Name: "nat"}
	c.Sig.Sorts["nat"] = natSort

	// Build an ActionDef: action foo(x:nat) = { x := x }
	paramX := ast.NewAtom("x")
	paramX.ASort = ast.NewSymbol("nat", nil)

	lhs := ast.NewAtom("x")
	rhs := ast.NewAtom("x")
	body := ast.NewAtom(":=", lhs, rhs)

	actionDef := ast.NewActionDef(
		ast.NewSymbol("foo", nil),
		body,
		[]ast.Node{paramX}, // FormalParams
		nil,                // FormalReturns
	)

	result, err := c.CompileAction(actionDef)
	if err != nil {
		t.Fatalf("CompileAction returned error: %v", err)
	}

	// Python renames formals with "prm:" prefix. Check that the compiled
	// action's formal params have names starting with "prm:".
	formals := result.GetFormalParams()
	if len(formals) == 0 {
		t.Fatal("expected at least 1 formal param, got 0")
	}
	if !strings.HasPrefix(formals[0].Name, "prm:") {
		t.Errorf("expected formal param name to start with 'prm:', got %q", formals[0].Name)
	}
}

// TestExpr6_CompileActionDef_FreeVarCheckInCalls tests that after compiling
// an action body, CompileAction checks all CallAction subactions for free
// variables in their arguments, raising "call may not have free variables".
//
// §6.3 #16: Go performs no free-variable check after compilation.
func TestExpr6_CompileActionDef_FreeVarCheckInCalls(t *testing.T) {
	c := newTestCompiler()

	natSort := &lg.UninterpretedSort{Name: "nat"}
	c.Sig.Sorts["nat"] = natSort
	c.Sig.AddSymbol("foo", natSort)

	// Register "foo" as an action in TopContext so the call path is exercised.
	c.TopCtx = &TopContext{
		Actions: map[string]*ActionInfo{
			"foo": {
				Params:  []*lg.Symbol{lg.NewSymbol("a", natSort)},
				Returns: nil,
			},
		},
	}

	// Build an ActionDef whose body is: call foo(X)
	// where X is a logic variable (uppercase = variable in Ivy convention),
	// not a declared constant. This should trigger "call may not have free variables".
	callTarget := ast.NewAtom("foo", ast.NewVariable("X", ast.NewSymbol("nat", nil)))
	callNode := ast.NewAtom("call", callTarget)

	paramA := ast.NewAtom("a")
	paramA.ASort = ast.NewSymbol("nat", nil)

	actionDef := ast.NewActionDef(
		ast.NewSymbol("bar", nil),
		callNode,
		[]ast.Node{paramA},
		nil,
	)

	_, err := c.CompileAction(actionDef)
	if err == nil {
		t.Fatal("expected error about free variables in call, got nil")
	}
	if !strings.Contains(err.Error(), "call may not have free variables") {
		t.Errorf("expected error containing 'call may not have free variables', got: %v", err)
	}
}

// TestExpr6_CompileLocal_AssignmentSortInference tests that when a local
// declaration has a single assignment child (local x := y), the LHS sort
// is inferred from the RHS via sort_infer(Equals(lhs, rhs)).
//
// §6.3 #17: Go's CompileLocal always compiles locals as bare declarations
// without special single-assignment sort inference.
func TestExpr6_CompileLocal_AssignmentSortInference(t *testing.T) {
	c := newTestCompiler()

	natSort := &lg.UninterpretedSort{Name: "nat"}
	c.Sig.Sorts["nat"] = natSort
	c.Sig.AddSymbol("y", natSort)

	// Build AST: local x := y
	// "x" has no sort annotation; "y" is a known symbol of sort nat.
	// Python infers x's sort as nat from y.
	//
	// Use CompileLocal which is the dedicated method for local compilation.
	xDecl := ast.NewAtom("x") // no ASort annotation — sort should be inferred
	assignBody := ast.NewAtom(":=", ast.NewAtom("x"), ast.NewAtom("y"))

	// Use CompileLocal which takes var decls and body separately.
	result, err := c.CompileLocal([]ast.Node{xDecl}, assignBody)
	if err != nil {
		t.Fatalf("CompileLocal returned error: %v", err)
	}

	// The result should be a LocalAction.
	localAct, ok := result.(*actions.LocalAction)
	if !ok {
		t.Fatalf("expected *actions.LocalAction, got %T", result)
	}

	// The local variable's sort should be "nat" (inferred from y), not TopS.
	if len(localAct.Locals) == 0 {
		t.Fatal("expected at least 1 local, got 0")
	}
	localSym, ok := localAct.Locals[0].(*lg.Symbol)
	if !ok {
		t.Fatalf("expected local to be *lg.Symbol, got %T", localAct.Locals[0])
	}
	sortName := localSym.CSort.String()
	if sortName != "nat" {
		t.Errorf("expected local x to have sort 'nat' (inferred from y), got %q", sortName)
	}
}

// TestExpr6_CompileCall_FieldReferenceFallback tests that CompileCall checks
// whether the callee is in TopContext.Actions, and if not, tries
// compile_field_reference as a fallback. If the field reference succeeds,
// it should raise "call to non-action".
//
// §6.3 #18 (field reference): Go's CompileCall just compiles the callee
// directly without checking TopContext.Actions or trying field reference.
func TestExpr6_CompileCall_FieldReferenceFallback(t *testing.T) {
	c := newTestCompiler()

	natSort := &lg.UninterpretedSort{Name: "nat"}
	c.Sig.Sorts["nat"] = natSort

	// Set up TopCtx with an empty Actions map — "foo" is NOT an action.
	c.TopCtx = &TopContext{
		Actions: map[string]*ActionInfo{},
	}

	// Add "foo" as a regular symbol (a destructor/function, not an action).
	// This means compile_field_reference could resolve it.
	c.Sig.AddSymbol("foo", natSort)
	c.Sig.AddSymbol("x", natSort)

	// Build a call AST: call foo(x)
	calleeNode := ast.NewAtom("foo", ast.NewAtom("x"))

	_, err := c.CompileCall(calleeNode, nil)
	if err == nil {
		t.Fatal("expected error 'call to non-action', got nil")
	}
	if !strings.Contains(err.Error(), "call to non-action") {
		t.Errorf("expected error containing 'call to non-action', got: %v", err)
	}
}

// TestExpr6_CompileCall_ParamCountValidation tests that CompileCall validates
// the number of input and output parameters against the action's declaration.
//
// §6.3 #18 (param count): Go's CompileCall performs no parameter count
// validation at all.
func TestExpr6_CompileCall_ParamCountValidation(t *testing.T) {
	c := newTestCompiler()

	natSort := &lg.UninterpretedSort{Name: "nat"}
	c.Sig.Sorts["nat"] = natSort

	// Add symbols to sig so they compile.
	c.Sig.AddSymbol("x", natSort)
	c.Sig.AddSymbol("y", natSort)
	c.Sig.AddSymbol("r1", natSort)
	c.Sig.AddSymbol("r2", natSort)
	// Add "foo" as a 2-arg function in the sig so CompileNode can resolve it.
	// Python looks up "foo" in top_context.actions and validates param counts
	// BEFORE compiling; Go skips this validation entirely.
	fooSort, _ := lg.NewFunctionSort(natSort, natSort, natSort)
	c.Sig.AddSymbol("foo", fooSort)

	// Register action "foo" that takes 2 input params and 1 return.
	c.TopCtx = &TopContext{
		Actions: map[string]*ActionInfo{
			"foo": {
				Params:  []*lg.Symbol{lg.NewSymbol("a", natSort), lg.NewSymbol("b", natSort)},
				Returns: []*lg.Symbol{lg.NewSymbol("r", natSort)},
			},
		},
	}

	// Test wrong number of input params: call foo(x) with only 1 arg (needs 2).
	// Supply the correct number of return targets (1) so the output count passes
	// and the input count error is triggered.
	// Python checks output params first, then input params (ivy_compiler.py:594-597).
	calleeNode := ast.NewAtom("foo", ast.NewAtom("x"))
	retTarget := []ast.Node{ast.NewAtom("r1")}
	_, err := c.CompileCall(calleeNode, retTarget)
	if err == nil {
		t.Error("expected error about wrong number of input parameters, got nil")
	} else if !strings.Contains(err.Error(), "wrong number of input parameters") {
		t.Errorf("expected error containing 'wrong number of input parameters', got: %v", err)
	}

	// Test wrong number of output params: call with 2 return targets when action expects 1.
	// Go's CompileCall accepts any number of return targets without checking.
	calleeNode2 := ast.NewAtom("foo", ast.NewAtom("x"), ast.NewAtom("y"))
	retTargets := []ast.Node{ast.NewAtom("r1"), ast.NewAtom("r2")}
	_, err = c.CompileCall(calleeNode2, retTargets)
	if err == nil {
		t.Fatal("expected error about wrong number of output parameters, got nil")
	}
	if !strings.Contains(err.Error(), "wrong number of output parameters") {
		t.Errorf("expected error containing 'wrong number of output parameters', got: %v", err)
	}
}

// TestExpr6_ExprContextExtract_LocalActionWrapping tests that
// ExprContext.Extract() wraps multiple code items with local symbols into
// a LocalAction, matching Python's ExprContext.extract().
//
// §6.3 #32: Go's ExprContext has no Extract() method at all.
// CompileInlineCode() exists but doesn't wrap in LocalAction with local_syms.
func TestExpr6_ExprContextExtract_LocalActionWrapping(t *testing.T) {
	// Create an ExprContext with 2 code items and 1 local symbol.
	natSort := &lg.UninterpretedSort{Name: "nat"}
	localSym := lg.NewSymbol("tmp", natSort)

	code1 := actions.WrapAction(actions.NewAssumeAction(lg.True))
	code2 := actions.WrapAction(actions.NewAssumeAction(lg.True))

	ec := &ExprContext{
		Code:      []lg.Expr{code1, code2},
		LocalSyms: []*lg.Symbol{localSym},
	}

	// Call Extract() — this method doesn't exist yet, so this tests that
	// it exists and produces the correct result.
	result := ec.Extract()
	if result == nil {
		t.Fatal("Extract() returned nil")
	}

	// The result should be a LocalAction wrapping the local symbol and a Sequence.
	act := actions.UnwrapAction(result)
	if act == nil {
		t.Fatal("Extract() result is not a wrapped action")
	}

	localAct, ok := act.(*actions.LocalAction)
	if !ok {
		t.Fatalf("expected *actions.LocalAction from Extract(), got %T", act)
	}

	// Check that the LocalAction contains the local symbol.
	if len(localAct.Locals) != 1 {
		t.Fatalf("expected 1 local, got %d", len(localAct.Locals))
	}
	sym, ok := localAct.Locals[0].(*lg.Symbol)
	if !ok {
		t.Fatalf("expected local to be *lg.Symbol, got %T", localAct.Locals[0])
	}
	if sym.Name != "tmp" {
		t.Errorf("expected local symbol name 'tmp', got %q", sym.Name)
	}

	// Check that the body is a Sequence with 2 children.
	bodyAct := actions.UnwrapAction(localAct.Body)
	if bodyAct == nil {
		t.Fatal("LocalAction body is not a wrapped action")
	}
	seq, ok := bodyAct.(*actions.Sequence)
	if !ok {
		t.Fatalf("expected body to be *actions.Sequence, got %T", bodyAct)
	}
	if len(seq.Children) != 2 {
		t.Errorf("expected Sequence with 2 children, got %d", len(seq.Children))
	}
}
