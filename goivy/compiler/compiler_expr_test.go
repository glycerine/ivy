package compiler

import (
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy/actions"
	"github.com/glycerine/ivy/goivy/ast"
	lg "github.com/glycerine/ivy/goivy/logic"
)

// =============================================================================
// Batch B — Compilation Expression Fixes (RED phase)
//
// All tests expose missing/buggy behavior in the Go port relative to Python.
// They should ALL FAIL until the corresponding fixes are implemented.
// =============================================================================

// TestExpr6_CompileActionDef_PrmPrefixSubstitution tests that CompileAction
// handles formal parameters correctly, matching Python's compile_action_def.
//
// Python's compile_action_def gets object params from a.args[0].args (the Name
// atom's children) and prm:-prefixes them. Formal params from a.formal_params
// keep their fml: prefix. When the LALR parser creates ActionDefs, params go
// into FormalParams (fml:-prefixed) and the Name atom has no children, so no
// prm: prefixing occurs — formals keep their fml: prefix.
func TestExpr6_CompileActionDef_PrmPrefixSubstitution(t *testing.T) {
	cfg := ast.NewAstConfig()
	c := newTestCompiler()

	// Set up a sort and a symbol so compilation succeeds.
	natSort := &lg.UninterpretedSort{Name: "nat"}
	c.Sig.Sorts.Set("nat", natSort)

	// Build an ActionDef: action foo(x:nat) = { x := x }
	// Matches LALR parser: Name atom has no children, params in FormalParams.
	paramX := cfg.NewAtom("x")
	paramX.ASort = cfg.NewSymbol("nat", nil)

	lhs := cfg.NewAtom("x")
	rhs := cfg.NewAtom("x")
	body := cfg.NewAssignAction(lhs, rhs)

	actionDef := cfg.NewActionDef(
		cfg.NewSymbol("foo", nil),
		body,
		[]ast.Node{paramX}, // FormalParams — gets fml: prefix in NewActionDef
		nil,                // FormalReturns
	)

	result, err := c.CompileAction(actionDef)
	if err != nil {
		t.Fatalf("CompileAction returned error: %v", err)
	}

	// Python: when Name atom has no children (as in LALR parser output),
	// pformals=[], and formals = a.formal_params (fml:-prefixed).
	// So compiled formals should have "fml:" prefix.
	formals := result.GetFormalParams()
	if len(formals) == 0 {
		t.Fatal("expected at least 1 formal param, got 0")
	}
	if !strings.HasPrefix(formals[0].Name, "fml:") {
		t.Errorf("expected formal param name to start with 'fml:', got %q", formals[0].Name)
	}
}

// TestExpr6_CompileActionDef_FreeVarCheckInCalls tests that after compiling
// an action body, CompileAction checks all CallAction subactions for free
// variables in their arguments, raising "call may not have free variables".
//
// §6.3 #16: Go performs no free-variable check after compilation.
func TestExpr6_CompileActionDef_FreeVarCheckInCalls(t *testing.T) {
	cfg := ast.NewAstConfig()
	c := newTestCompiler()

	natSort := &lg.UninterpretedSort{Name: "nat"}
	c.Sig.Sorts.Set("nat", natSort)
	c.Sig.AddSymbol("foo", natSort)

	// Register "foo" as an action in TopContext so the call path is exercised.
	c.TopCtx = &TopContext{
		Actions: map[string]*ActionInfo{
			"foo": {
				Params:  []*lg.Const{lg.NewConst("a", natSort)},
				Returns: nil,
			},
		},
	}

	// Build an ActionDef whose body is: call foo(X)
	// where X is a logic variable (uppercase = variable in Ivy convention),
	// not a declared constant. This should trigger "call may not have free variables".
	callTarget := cfg.NewAtom("foo", cfg.NewVariable("X", "nat"))
	callNode := cfg.NewCallAction(callTarget)

	paramA := cfg.NewAtom("a")
	paramA.ASort = cfg.NewSymbol("nat", nil)

	actionDef := cfg.NewActionDef(
		cfg.NewSymbol("bar", nil),
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
// declaration is a single AssignAction (from LowerVarStatements: var x := y),
// the LHS sort is inferred from the RHS via sort_infer(Equals(lhs, rhs)).
//
// Python: isinstance(ls[0], AssignAction) path in compile_local.
// LowerVarStatements produces: LocalAction(AssignAction(loc:x, y), body).
func TestExpr6_CompileLocal_AssignmentSortInference(t *testing.T) {
	cfg := ast.NewAstConfig()
	c := newTestCompiler()

	natSort := &lg.UninterpretedSort{Name: "nat"}
	c.Sig.Sorts.Set("nat", natSort)
	c.Sig.AddSymbol("y", natSort)

	// Build AST matching LowerVarStatements output:
	// localDecls = [AssignAction(loc:x, y)]
	// body = Sequence() (empty continuation)
	lhsNode := cfg.NewAtom("loc:x") // no ASort — sort should be inferred
	rhsNode := cfg.NewAtom("y")
	localDecls := []ast.Node{cfg.NewAssignAction(lhsNode, rhsNode)}
	body := cfg.NewSequence()

	result, err := c.CompileLocal(localDecls, body)
	if err != nil {
		t.Fatalf("CompileLocal returned error: %v", err)
	}

	localAct, ok := result.(*actions.LocalAction)
	if !ok {
		t.Fatalf("expected *actions.LocalAction, got %T", result)
	}

	// The local variable's sort should be "nat" (inferred from y), not TopS.
	if len(localAct.Locals) == 0 {
		t.Fatal("expected at least 1 local, got 0")
	}
	localSym, ok := localAct.Locals[0].(*lg.Const)
	if !ok {
		t.Fatalf("expected local to be *lg.Const, got %T", localAct.Locals[0])
	}
	sortName := localSym.CSort.String()
	if sortName != "nat" {
		t.Errorf("expected local x to have sort 'nat' (inferred from y), got %q", sortName)
	}

	// Body should contain the assignment prepended to continuation.
	if localAct.Body == nil {
		t.Fatal("expected non-nil body")
	}
	bodySeq, ok := localAct.Body.(*actions.Sequence)
	if !ok {
		t.Fatalf("expected body to be *actions.Sequence, got %T", localAct.Body)
	}
	if len(bodySeq.Elems) < 1 {
		t.Fatal("expected body sequence to contain at least the assignment")
	}
	if _, ok := bodySeq.Elems[0].(*actions.AssignAction); !ok {
		t.Errorf("expected first body element to be *actions.AssignAction, got %T", bodySeq.Elems[0])
	}
}

// TestExpr6_CompileLocal_ExplicitSortAnnotation tests that when the LHS
// has an explicit sort annotation, it is used instead of inferring from RHS.
func TestExpr6_CompileLocal_ExplicitSortAnnotation(t *testing.T) {
	cfg := ast.NewAstConfig()
	c := newTestCompiler()

	natSort := &lg.UninterpretedSort{Name: "nat"}
	c.Sig.Sorts.Set("nat", natSort)
	c.Sig.AddSymbol("y", natSort)

	lhsNode := cfg.NewAtom("loc:x")
	lhsNode.ASort = cfg.NewSymbol("nat", nil) // explicit sort
	rhsNode := cfg.NewAtom("y")
	localDecls := []ast.Node{cfg.NewAssignAction(lhsNode, rhsNode)}
	body := cfg.NewSequence()

	result, err := c.CompileLocal(localDecls, body)
	if err != nil {
		t.Fatalf("CompileLocal returned error: %v", err)
	}

	localAct, ok := result.(*actions.LocalAction)
	if !ok {
		t.Fatalf("expected *actions.LocalAction, got %T", result)
	}
	if len(localAct.Locals) == 0 {
		t.Fatal("expected at least 1 local")
	}
	localSym, ok := localAct.Locals[0].(*lg.Const)
	if !ok {
		t.Fatalf("expected local to be *lg.Const, got %T", localAct.Locals[0])
	}
	if localSym.CSort.String() != "nat" {
		t.Errorf("expected sort 'nat', got %q", localSym.CSort.String())
	}
}

// TestExpr6_CompileLocal_FunctionLikeLHS tests CompileLocal when the LHS
// is an App (function with args), e.g. "local f(P) := true".
// The local variable extracted should be the function symbol, not the Apply.
func TestExpr6_CompileLocal_FunctionLikeLHS(t *testing.T) {
	cfg := ast.NewAstConfig()
	c := newTestCompiler()

	procSort := &lg.UninterpretedSort{Name: "proc"}
	c.Sig.Sorts.Set("proc", procSort)
	c.Sig.AddSymbol("P", procSort)

	lhsNode := cfg.NewApp(cfg.NewSymbol("loc:f", nil), cfg.NewAtom("P"))
	rhsNode := cfg.NewAtom("true")
	localDecls := []ast.Node{cfg.NewAssignAction(lhsNode, rhsNode)}
	body := cfg.NewSequence()

	result, err := c.CompileLocal(localDecls, body)
	if err != nil {
		t.Fatalf("CompileLocal returned error: %v", err)
	}

	localAct, ok := result.(*actions.LocalAction)
	if !ok {
		t.Fatalf("expected *actions.LocalAction, got %T", result)
	}
	if len(localAct.Locals) == 0 {
		t.Fatal("expected at least 1 local")
	}
	// Local should be a Symbol (the function symbol), not an Apply.
	localSym, ok := localAct.Locals[0].(*lg.Const)
	if !ok {
		t.Fatalf("expected local to be *lg.Const (function symbol), got %T", localAct.Locals[0])
	}
	if !strings.Contains(localSym.Name, "loc:f") {
		t.Errorf("expected local symbol name containing 'loc:f', got %q", localSym.Name)
	}
}

// TestExpr6_CompileLocal_BareDeclaration tests the generic case where
// localDecls[0] is NOT an AssignAction — just a bare variable declaration.
func TestExpr6_CompileLocal_BareDeclaration(t *testing.T) {
	cfg := ast.NewAstConfig()
	c := newTestCompiler()

	natSort := &lg.UninterpretedSort{Name: "nat"}
	c.Sig.Sorts.Set("nat", natSort)

	// Bare declaration: local loc:x { true }
	xDecl := cfg.NewAtom("loc:x")
	xDecl.ASort = cfg.NewSymbol("nat", nil)
	body := cfg.NewAtom("true")
	localDecls := []ast.Node{xDecl}

	result, err := c.CompileLocal(localDecls, body)
	if err != nil {
		t.Fatalf("CompileLocal returned error: %v", err)
	}

	localAct, ok := result.(*actions.LocalAction)
	if !ok {
		t.Fatalf("expected *actions.LocalAction, got %T", result)
	}
	if len(localAct.Locals) == 0 {
		t.Fatal("expected at least 1 local")
	}
}

// TestExpr6_CompileLocal_MultipleDeclarations tests the generic path
// with multiple local declarations.
func TestExpr6_CompileLocal_MultipleDeclarations(t *testing.T) {
	cfg := ast.NewAstConfig()
	c := newTestCompiler()

	natSort := &lg.UninterpretedSort{Name: "nat"}
	c.Sig.Sorts.Set("nat", natSort)

	xDecl := cfg.NewAtom("loc:x")
	xDecl.ASort = cfg.NewSymbol("nat", nil)
	yDecl := cfg.NewAtom("loc:y")
	yDecl.ASort = cfg.NewSymbol("nat", nil)
	body := cfg.NewAtom("true")
	localDecls := []ast.Node{xDecl, yDecl}

	result, err := c.CompileLocal(localDecls, body)
	if err != nil {
		t.Fatalf("CompileLocal returned error: %v", err)
	}

	localAct, ok := result.(*actions.LocalAction)
	if !ok {
		t.Fatalf("expected *actions.LocalAction, got %T", result)
	}
	if len(localAct.Locals) != 2 {
		t.Errorf("expected 2 locals, got %d", len(localAct.Locals))
	}
}

// TestExpr6_CompileLocal_BodyWithMultipleStatements tests that the assignment
// is prepended to the body's statements when body is a Sequence.
func TestExpr6_CompileLocal_BodyWithMultipleStatements(t *testing.T) {
	cfg := ast.NewAstConfig()
	c := newTestCompiler()

	natSort := &lg.UninterpretedSort{Name: "nat"}
	c.Sig.Sorts.Set("nat", natSort)
	c.Sig.AddSymbol("y", natSort)
	c.Sig.AddSymbol("z", natSort)

	lhsNode := cfg.NewAtom("loc:x")
	rhsNode := cfg.NewAtom("y")
	localDecls := []ast.Node{cfg.NewAssignAction(lhsNode, rhsNode)}

	// Body with two statements
	stmt1 := cfg.NewAtom("true")
	stmt2 := cfg.NewAtom("true")
	body := cfg.NewSequence(stmt1, stmt2)

	result, err := c.CompileLocal(localDecls, body)
	if err != nil {
		t.Fatalf("CompileLocal returned error: %v", err)
	}

	localAct, ok := result.(*actions.LocalAction)
	if !ok {
		t.Fatalf("expected *actions.LocalAction, got %T", result)
	}

	// Body should be Sequence with assignment + 2 original statements = 3 elements
	bodySeq, ok := localAct.Body.(*actions.Sequence)
	if !ok {
		t.Fatalf("expected body to be *actions.Sequence, got %T", localAct.Body)
	}
	if len(bodySeq.Elems) != 3 {
		t.Errorf("expected 3 body elements (asgn + 2 stmts), got %d", len(bodySeq.Elems))
	}
	if _, ok := bodySeq.Elems[0].(*actions.AssignAction); !ok {
		t.Errorf("expected first element to be assignment, got %T", bodySeq.Elems[0])
	}
}

// TestExpr6_CompileLocal_EnsureSortAnnotation tests that ensureSortAnnotation
// sets ASort to "S" when it is nil, preventing CompileConst from failing.
func TestExpr6_CompileLocal_EnsureSortAnnotation(t *testing.T) {
	// Unit test of ensureSortAnnotation itself
	cfg := ast.NewAstConfig()

	atom := cfg.NewAtom("x")
	if atom.ASort != nil {
		t.Fatal("precondition: ASort should be nil")
	}
	ensureSortAnnotation(atom, cfg)
	if atom.ASort == nil {
		t.Fatal("ensureSortAnnotation should have set ASort")
	}
	sym, ok := atom.ASort.(*ast.Symbol)
	if !ok {
		t.Fatalf("expected ASort to be *ast.Symbol, got %T", atom.ASort)
	}
	if sym.Rep != "S" {
		t.Errorf("expected ASort rep 'S', got %q", sym.Rep)
	}

	// Also test App
	app := cfg.NewApp(cfg.NewSymbol("f", nil), cfg.NewAtom("a"))
	if app.ASort != nil {
		t.Fatal("precondition: App.ASort should be nil")
	}
	ensureSortAnnotation(app, cfg)
	if app.ASort == nil {
		t.Fatal("ensureSortAnnotation should have set App.ASort")
	}
	sym2, ok := app.ASort.(*ast.Symbol)
	if !ok {
		t.Fatalf("expected App.ASort to be *ast.Symbol, got %T", app.ASort)
	}
	if sym2.Rep != "S" {
		t.Errorf("expected App.ASort rep 'S', got %q", sym2.Rep)
	}
}

// TestExpr6_CompileLocal_LowerVarRoundTrip tests the full pipeline:
// VarAction → LowerVarStatements → CompileLocal.
func TestExpr6_CompileLocal_LowerVarRoundTrip(t *testing.T) {
	cfg := ast.NewAstConfig()
	c := newTestCompiler()

	natSort := &lg.UninterpretedSort{Name: "nat"}
	c.Sig.Sorts.Set("nat", natSort)
	c.Sig.AddSymbol("y", natSort)

	// Create: var x := y; true
	xAtom := cfg.NewAtom("x")
	yAtom := cfg.NewAtom("y")
	varAction := cfg.NewVarAction(xAtom, yAtom)
	trueStmt := cfg.NewAtom("true")

	// LowerVarStatements transforms [varAction, trueStmt] into
	// [LocalAction(AssignAction(loc:x, y), Sequence(trueStmt))]
	lowered := ast.LowerVarStatements([]ast.Node{varAction, trueStmt})
	if len(lowered) != 1 {
		t.Fatalf("expected 1 lowered node, got %d", len(lowered))
	}
	la, ok := lowered[0].(*ast.AstLocalAction)
	if !ok {
		t.Fatalf("expected *ast.LocalAction, got %T", lowered[0])
	}
	if len(la.Elems) < 2 {
		t.Fatalf("expected at least 2 elems in LocalAction, got %d", len(la.Elems))
	}

	// Extract localDecls and body (same as CompileActionBody does)
	localDecls := la.Elems[:len(la.Elems)-1]
	body := la.Elems[len(la.Elems)-1]

	// The first local decl should be an AssignAction
	if _, ok := localDecls[0].(*ast.AstAssignAction); !ok {
		t.Fatalf("expected localDecls[0] to be *ast.AssignAction, got %T", localDecls[0])
	}

	result, err := c.CompileLocal(localDecls, body)
	if err != nil {
		t.Fatalf("CompileLocal returned error: %v", err)
	}

	localAct, ok := result.(*actions.LocalAction)
	if !ok {
		t.Fatalf("expected *actions.LocalAction, got %T", result)
	}
	if len(localAct.Locals) == 0 {
		t.Fatal("expected at least 1 local")
	}
	// Verify sort inference worked
	localSym, ok := localAct.Locals[0].(*lg.Const)
	if !ok {
		t.Fatalf("expected local to be *lg.Const, got %T", localAct.Locals[0])
	}
	if localSym.CSort == nil {
		t.Error("expected local symbol to have a sort")
	}
}

// TestExpr6_CompileLocal_SymbolShadowing tests that a local variable
// correctly shadows an outer-scope symbol of the same name.
func TestExpr6_CompileLocal_SymbolShadowing(t *testing.T) {
	cfg := ast.NewAstConfig()
	c := newTestCompiler()

	natSort := &lg.UninterpretedSort{Name: "nat"}
	boolSort := &lg.UninterpretedSort{Name: "bool"}
	c.Sig.Sorts.Set("nat", natSort)
	c.Sig.Sorts.Set("bool", boolSort)
	// Outer scope: x is nat
	c.Sig.AddSymbol("x", natSort)
	c.Sig.AddSymbol("y", boolSort)

	// local loc:x := y (where y is bool, so x becomes bool)
	lhsNode := cfg.NewAtom("loc:x")
	rhsNode := cfg.NewAtom("y")
	localDecls := []ast.Node{cfg.NewAssignAction(lhsNode, rhsNode)}
	body := cfg.NewSequence()

	result, err := c.CompileLocal(localDecls, body)
	if err != nil {
		t.Fatalf("CompileLocal returned error: %v", err)
	}

	localAct, ok := result.(*actions.LocalAction)
	if !ok {
		t.Fatalf("expected *actions.LocalAction, got %T", result)
	}
	if len(localAct.Locals) == 0 {
		t.Fatal("expected at least 1 local")
	}

	// The local should have sort "bool" (from y), NOT "nat" (from outer x)
	localSym, ok := localAct.Locals[0].(*lg.Const)
	if !ok {
		t.Fatalf("expected local to be *lg.Const, got %T", localAct.Locals[0])
	}
	if localSym.CSort.String() != "bool" {
		t.Errorf("expected local to have sort 'bool' (inferred from y), got %q", localSym.CSort.String())
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
	cfg := ast.NewAstConfig()
	c := newTestCompiler()

	natSort := &lg.UninterpretedSort{Name: "nat"}
	c.Sig.Sorts.Set("nat", natSort)

	// Set up TopCtx with an empty Actions map — "foo" is NOT an action.
	c.TopCtx = &TopContext{
		Actions: map[string]*ActionInfo{},
	}

	// Add "foo" as a regular symbol (a destructor/function, not an action).
	// This means compile_field_reference could resolve it.
	c.Sig.AddSymbol("foo", natSort)
	c.Sig.AddSymbol("x", natSort)

	// Build a call AST: call foo(x)
	calleeNode := cfg.NewAtom("foo", cfg.NewAtom("x"))

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
	cfg := ast.NewAstConfig()
	c := newTestCompiler()

	natSort := &lg.UninterpretedSort{Name: "nat"}
	c.Sig.Sorts.Set("nat", natSort)

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
				Params:  []*lg.Const{lg.NewConst("a", natSort), lg.NewConst("b", natSort)},
				Returns: []*lg.Const{lg.NewConst("r", natSort)},
			},
		},
	}

	// Test wrong number of input params: call foo(x) with only 1 arg (needs 2).
	// Supply the correct number of return targets (1) so the output count passes
	// and the input count error is triggered.
	// Python checks output params first, then input params (ivy_compiler.py:594-597).
	calleeNode := cfg.NewAtom("foo", cfg.NewAtom("x"))
	retTarget := []ast.Node{cfg.NewAtom("r1")}
	_, err := c.CompileCall(calleeNode, retTarget)
	if err == nil {
		t.Error("expected error about wrong number of input parameters, got nil")
	} else if !strings.Contains(err.Error(), "wrong number of input parameters") {
		t.Errorf("expected error containing 'wrong number of input parameters', got: %v", err)
	}

	// Test wrong number of output params: call with 2 return targets when action expects 1.
	// Go's CompileCall accepts any number of return targets without checking.
	calleeNode2 := cfg.NewAtom("foo", cfg.NewAtom("x"), cfg.NewAtom("y"))
	retTargets := []ast.Node{cfg.NewAtom("r1"), cfg.NewAtom("r2")}
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
	localSym := lg.NewConst("tmp", natSort)

	code1 := actions.NewAssumeAction(lg.True)
	code2 := actions.NewAssumeAction(lg.True)

	ec := &ExprContext{
		Code:      []lg.Expr{code1, code2},
		LocalSyms: []*lg.Const{localSym},
		ActCfg:    actions.NewActionsConfig(),
	}

	// Call Extract() — this method doesn't exist yet, so this tests that
	// it exists and produces the correct result.
	result := ec.Extract()
	if result == nil {
		t.Fatal("Extract() returned nil")
	}

	// The result should be a LocalAction wrapping the local symbol and a Sequence.
	act, _ := result.(actions.Action)
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
	sym, ok := localAct.Locals[0].(*lg.Const)
	if !ok {
		t.Fatalf("expected local to be *lg.Const, got %T", localAct.Locals[0])
	}
	if sym.Name != "tmp" {
		t.Errorf("expected local symbol name 'tmp', got %q", sym.Name)
	}

	// Check that the body is a Sequence with 2 children.
	bodyAct, _ := localAct.Body.(actions.Action)
	if bodyAct == nil {
		t.Fatal("LocalAction body is not a wrapped action")
	}
	seq, ok := bodyAct.(*actions.Sequence)
	if !ok {
		t.Fatalf("expected body to be *actions.Sequence, got %T", bodyAct)
	}
	if len(seq.Elems) != 2 {
		t.Errorf("expected Sequence with 2 children, got %d", len(seq.Elems))
	}
}
