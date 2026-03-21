package actions

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/glycerine/goivy/ast"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
)

// -----------------------------------------------------------------------
// Unit tests for instantiateMacro
// -----------------------------------------------------------------------

// makeMacroDef creates a Definition like: macro name(fparams...) = body
func makeMacroDef(name string, fparams []string, body ast.Node) *ast.Definition {
	terms := make([]ast.Node, len(fparams))
	for i, fp := range fparams {
		terms[i] = ast.NewSymbol(fp, nil)
	}
	lhs := ast.NewAtom(name, terms...)
	return ast.NewDefinition(lhs, body)
}

func TestInstantiateMacroSimpleSubstitution(t *testing.T) {
	// macro incr(x) = x + 1
	// instantiate incr(y)
	// Expected: y + 1
	// Note: In the parser, body references to formals are zero-arity Atoms,
	// not Symbols. RewriteAtom only substitutes zero-arity Atoms.
	body := ast.NewAtom("+", ast.NewAtom("x"), ast.NewAtom("1"))
	defn := makeMacroDef("incr", []string{"x"}, body)

	macros := map[string]interface{}{
		"incr": defn,
	}

	inst := ast.NewAtom("incr", ast.NewAtom("y"))
	result := instantiateMacro(inst, macros)
	if result == nil {
		t.Fatal("instantiateMacro returned nil for valid macro")
	}
	atom, ok := result.(*ast.Atom)
	if !ok {
		t.Fatalf("expected *ast.Atom, got %T: %s", result, result)
	}
	if atom.Rep != "+" {
		t.Errorf("expected '+', got %q", atom.Rep)
	}
	// First arg should be "y" (substituted from "x")
	if len(atom.Terms) != 2 {
		t.Fatalf("expected 2 terms, got %d", len(atom.Terms))
	}
	if a, ok := atom.Terms[0].(*ast.Atom); !ok || a.Rep != "y" {
		t.Errorf("expected first arg 'y', got %s", atom.Terms[0])
	}
	// Second arg should remain "1" (not substituted)
	if a, ok := atom.Terms[1].(*ast.Atom); !ok || a.Rep != "1" {
		t.Errorf("expected second arg '1', got %s", atom.Terms[1])
	}
}

func TestInstantiateMacroMultipleParams(t *testing.T) {
	// macro swap(a, b) = pair(b, a)
	// instantiate swap(x, y)
	// Expected: pair(y, x)
	body := ast.NewAtom("pair", ast.NewAtom("b"), ast.NewAtom("a"))
	defn := makeMacroDef("swap", []string{"a", "b"}, body)

	macros := map[string]interface{}{
		"swap": defn,
	}

	inst := ast.NewAtom("swap", ast.NewAtom("x"), ast.NewAtom("y"))
	result := instantiateMacro(inst, macros)
	if result == nil {
		t.Fatal("instantiateMacro returned nil")
	}
	atom := result.(*ast.Atom)
	if atom.Rep != "pair" {
		t.Errorf("expected 'pair', got %q", atom.Rep)
	}
	// First arg should be "y" (from "b" -> "y")
	if a, ok := atom.Terms[0].(*ast.Atom); !ok || a.Rep != "y" {
		t.Errorf("expected 'y', got %s", atom.Terms[0])
	}
	// Second arg should be "x" (from "a" -> "x")
	if a, ok := atom.Terms[1].(*ast.Atom); !ok || a.Rep != "x" {
		t.Errorf("expected 'x', got %s", atom.Terms[1])
	}
}

func TestInstantiateMacroZeroParams(t *testing.T) {
	// macro truthy() = true_val
	// instantiate truthy()
	body := ast.NewAtom("true_val")
	lhs := ast.NewAtom("truthy")
	defn := ast.NewDefinition(lhs, body)

	macros := map[string]interface{}{
		"truthy": defn,
	}

	inst := ast.NewAtom("truthy")
	result := instantiateMacro(inst, macros)
	if result == nil {
		t.Fatal("instantiateMacro returned nil for zero-param macro")
	}
	if a, ok := result.(*ast.Atom); !ok || a.Rep != "true_val" {
		t.Errorf("expected 'true_val', got %s", result)
	}
}

func TestInstantiateMacroNotFound(t *testing.T) {
	macros := map[string]interface{}{}
	inst := ast.NewAtom("nonexistent", ast.NewSymbol("x", nil))
	result := instantiateMacro(inst, macros)
	if result != nil {
		t.Errorf("expected nil for missing macro, got %s", result)
	}
}

func TestInstantiateMacroNilMacroValue(t *testing.T) {
	macros := map[string]interface{}{
		"m": nil,
	}
	inst := ast.NewAtom("m")
	result := instantiateMacro(inst, macros)
	if result != nil {
		t.Errorf("expected nil for nil macro def, got %s", result)
	}
}

func TestInstantiateMacroWrongType(t *testing.T) {
	// Macro value is not a *ast.Definition
	macros := map[string]interface{}{
		"m": "not a definition",
	}
	inst := ast.NewAtom("m")
	result := instantiateMacro(inst, macros)
	if result != nil {
		t.Errorf("expected nil for wrong type macro def, got %s", result)
	}
}

func TestInstantiateMacroWrongParamCount(t *testing.T) {
	// macro f(x) = x
	// instantiate f(a, b) -- wrong param count
	body := ast.NewAtom("x")
	defn := makeMacroDef("f", []string{"x"}, body)

	macros := map[string]interface{}{
		"f": defn,
	}

	inst := ast.NewAtom("f", ast.NewAtom("a"), ast.NewAtom("b"))
	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic for wrong param count")
		}
		msg := fmt.Sprint(r)
		if !strings.Contains(msg, "wrong number of parameters") {
			t.Errorf("unexpected panic message: %s", msg)
		}
	}()
	instantiateMacro(inst, macros)
}

func TestInstantiateMacroSymbolInst(t *testing.T) {
	// instantiate with a bare Symbol (zero-arity)
	// macro m() = body
	body := ast.NewAtom("result")
	lhs := ast.NewAtom("m")
	defn := ast.NewDefinition(lhs, body)

	macros := map[string]interface{}{
		"m": defn,
	}

	inst := ast.NewSymbol("m", nil)
	result := instantiateMacro(inst, macros)
	if result == nil {
		t.Fatal("expected non-nil for Symbol inst")
	}
}

func TestInstantiateMacroNonNodeInst(t *testing.T) {
	// Pass something that is neither *ast.Atom nor *ast.Symbol
	macros := map[string]interface{}{}
	result := instantiateMacro(&ast.And{}, macros)
	if result != nil {
		t.Errorf("expected nil for non-Atom/Symbol inst, got %s", result)
	}
}

func TestInstantiateMacroPsubstApplied(t *testing.T) {
	// macro m(x) = foo.x.bar
	// The body uses "x" as a subscript in a composite name like "foo.x.bar"
	// psubst should rename "x" → actual param name within dotted names
	//
	// This tests psubst by checking that zero-arity params get string substitution.
	// The psubst is used by AstRewriteSubstConstantsParams.RewriteName which
	// calls SubstSubscripts. We test that the rewriter is constructed correctly.
	body := ast.NewAtom("x") // simple: just the param itself (zero-arity Atom)
	defn := makeMacroDef("m", []string{"x"}, body)

	macros := map[string]interface{}{
		"m": defn,
	}

	// Pass a zero-arity Atom as actual param (triggers psubst)
	inst := ast.NewAtom("m", ast.NewAtom("myval"))
	result := instantiateMacro(inst, macros)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	// The zero-arity atom "myval" should be substituted for "x" via subst
	if a, ok := result.(*ast.Atom); ok {
		if a.Rep != "myval" {
			t.Errorf("expected 'myval', got %q", a.Rep)
		}
	}
}

func TestInstantiateMacroNestedBody(t *testing.T) {
	// macro m(x) = f(g(x), x)
	// instantiate m(y)
	// Expected: f(g(y), y)
	inner := ast.NewAtom("g", ast.NewAtom("x"))
	body := ast.NewAtom("f", inner, ast.NewAtom("x"))
	defn := makeMacroDef("m", []string{"x"}, body)

	macros := map[string]interface{}{
		"m": defn,
	}

	inst := ast.NewAtom("m", ast.NewAtom("y"))
	result := instantiateMacro(inst, macros)
	if result == nil {
		t.Fatal("instantiateMacro returned nil for nested body")
	}
	outer := result.(*ast.Atom)
	if outer.Rep != "f" {
		t.Errorf("outer should be 'f', got %q", outer.Rep)
	}
	if len(outer.Terms) != 2 {
		t.Fatalf("expected 2 terms, got %d", len(outer.Terms))
	}
	// Check g(y)
	innerResult, ok := outer.Terms[0].(*ast.Atom)
	if !ok || innerResult.Rep != "g" {
		t.Errorf("inner should be 'g', got %s", outer.Terms[0])
	}
	if a, ok := innerResult.Terms[0].(*ast.Atom); !ok || a.Rep != "y" {
		t.Errorf("inner arg should be 'y', got %s", innerResult.Terms[0])
	}
	// Check outer second arg is y
	if a, ok := outer.Terms[1].(*ast.Atom); !ok || a.Rep != "y" {
		t.Errorf("second arg should be 'y', got %s", outer.Terms[1])
	}
}

// -----------------------------------------------------------------------
// Unit tests for InstantiateAction
// -----------------------------------------------------------------------

func TestInstantiateActionBasic(t *testing.T) {
	a := NewInstantiateAction(nil)
	if a.Name() != "instantiate" {
		t.Errorf("Name() = %q", a.Name())
	}
	if a.AstInst != nil {
		t.Error("AstInst should be nil by default")
	}
}

func TestInstantiateActionWithAstInst(t *testing.T) {
	a := NewInstantiateAction(nil)
	a.AstInst = ast.NewAtom("mymacro", ast.NewSymbol("arg", nil))
	if a.AstInst == nil {
		t.Error("AstInst should be set")
	}
}

func TestInstantiateActionClonePreservesAstInst(t *testing.T) {
	a := NewInstantiateAction(nil)
	astNode := ast.NewAtom("mymacro", ast.NewSymbol("arg", nil))
	a.AstInst = astNode

	cloned := a.ActionClone(a.ActionArgs())
	ia, ok := cloned.(*InstantiateAction)
	if !ok {
		t.Fatal("clone should return *InstantiateAction")
	}
	if ia.AstInst != astNode {
		t.Error("clone should preserve AstInst")
	}
}

func TestInstantiateActionIntUpdateNilDomain(t *testing.T) {
	a := NewInstantiateAction(nil)
	ctx := &UpdateContext{Domain: nil}
	u := a.IntUpdate(ctx)
	if u == nil {
		t.Fatal("IntUpdate should return non-nil")
	}
	if len(u.Modified) != 0 {
		t.Error("should modify nothing with nil domain")
	}
}

func TestInstantiateActionIntUpdateMacroExpansion(t *testing.T) {
	// Set up a macro: macro incr(x) = assume x
	// The body after rewriting should be "assume y"
	body := ast.NewAtom("assume", ast.NewAtom("x"))
	defn := makeMacroDef("incr", []string{"x"}, body)

	mod := module.New()
	mod.Macros = map[string]interface{}{
		"incr": defn,
	}

	// Create the action with AstInst
	a := NewInstantiateAction(nil)
	a.AstInst = ast.NewAtom("incr", ast.NewAtom("y"))

	// Set up a CompileActionBody callback
	compileCalled := false
	ctx := &UpdateContext{
		Domain: mod,
		PVars:  nil,
		CompileActionBody: func(node ast.Node) (Action, error) {
			compileCalled = true
			// The rewritten node should be "assume y"
			atom, ok := node.(*ast.Atom)
			if !ok {
				t.Errorf("compiled node should be *ast.Atom, got %T", node)
				return NewSequence(), nil
			}
			if atom.Rep != "assume" {
				t.Errorf("expected 'assume', got %q", atom.Rep)
			}
			// Return a simple assume action
			return NewAssumeAction(lg.True), nil
		},
	}

	u := a.IntUpdate(ctx)
	if !compileCalled {
		t.Error("CompileActionBody should have been called")
	}
	if u == nil {
		t.Fatal("IntUpdate should return non-nil")
	}
	// The assume(true) should produce a true TR
	if !u.TR.IsTrue() {
		t.Errorf("expected true TR from assume(true), got %s", u.TR)
	}
}

func TestInstantiateActionIntUpdateModuleCallback(t *testing.T) {
	// Test that the module's CompileActionBodyFn is used as fallback
	body := ast.NewAtom("x")
	defn := makeMacroDef("m", []string{"x"}, body)

	mod := module.New()
	mod.Macros = map[string]interface{}{
		"m": defn,
	}

	moduleFnCalled := false
	mod.CompileActionBodyFn = func(node ast.Node) (module.Action, error) {
		moduleFnCalled = true
		return NewAssumeAction(lg.True), nil
	}

	a := NewInstantiateAction(nil)
	a.AstInst = ast.NewAtom("m", ast.NewAtom("y"))

	ctx := &UpdateContext{
		Domain: mod,
		PVars:  nil,
		// No CompileActionBody set — should fall back to module's callback
	}

	u := a.IntUpdate(ctx)
	if !moduleFnCalled {
		t.Error("module CompileActionBodyFn should have been called")
	}
	if u == nil {
		t.Fatal("IntUpdate returned nil")
	}
}

func TestInstantiateActionIntUpdateSchemaFallback(t *testing.T) {
	// When macro not found, should fall back to schemata
	mod := module.New()
	mod.Macros = map[string]interface{}{} // no macros

	fmla := lg.NewSymbol("p", lg.Boolean)
	mod.Schemata["myschema"] = ast.NewLabeledFormula(ast.NewSymbol("myschema", nil), fmla)

	a := NewInstantiateAction(nil)
	a.AstInst = ast.NewAtom("myschema")

	ctx := &UpdateContext{
		Domain: mod,
		PVars:  nil,
	}

	u := a.IntUpdate(ctx)
	if u == nil {
		t.Fatal("IntUpdate returned nil")
	}
	// Should have non-null TR from schema
	if u.TR == nil {
		t.Error("TR should be non-nil from schema path")
	}
}

func TestInstantiateActionIntUpdateNoMacroNoSchema(t *testing.T) {
	mod := module.New()
	mod.Macros = map[string]interface{}{}

	a := NewInstantiateAction(nil)
	a.AstInst = ast.NewAtom("unknown")

	ctx := &UpdateContext{
		Domain: mod,
		PVars:  nil,
	}

	u := a.IntUpdate(ctx)
	// Should be null update
	if len(u.Modified) != 0 {
		t.Error("unknown instantiate should produce null update")
	}
}

func TestInstantiateActionIntUpdateCompiledExprFallback(t *testing.T) {
	// When AstInst is nil, use compiled Inst for schema lookup
	mod := module.New()

	fmla := lg.NewSymbol("q", lg.Boolean)
	mod.Schemata["myschema"] = ast.NewLabeledFormula(ast.NewSymbol("myschema", nil), fmla)

	sym := lg.NewSymbol("myschema", lg.TopS)
	a := NewInstantiateAction(sym)
	// No AstInst set

	ctx := &UpdateContext{
		Domain: mod,
		PVars:  nil,
	}

	u := a.IntUpdate(ctx)
	if u == nil {
		t.Fatal("IntUpdate returned nil")
	}
	if u.TR == nil {
		t.Error("TR should be non-nil from schema using compiled Inst")
	}
}

// -----------------------------------------------------------------------
// Randomized tests for instantiateMacro
// -----------------------------------------------------------------------

func TestInstantiateMacroRandomized(t *testing.T) {
	// Generate random macro definitions with N params and verify
	// that substitution produces correct results.
	rng := rand.New(rand.NewSource(42))

	for trial := 0; trial < 100; trial++ {
		nParams := rng.Intn(5) // 0-4 params

		// Generate formal param names
		fparams := make([]string, nParams)
		for i := range fparams {
			fparams[i] = fmt.Sprintf("p%d", i)
		}

		// Generate actual param names (zero-arity Atoms, as the parser produces)
		aparams := make([]ast.Node, nParams)
		apNames := make([]string, nParams)
		for i := range aparams {
			apNames[i] = fmt.Sprintf("a%d_%d", trial, i)
			aparams[i] = ast.NewAtom(apNames[i])
		}

		// Body: an atom referencing all formal params (as zero-arity Atoms)
		var bodyTerms []ast.Node
		for _, fp := range fparams {
			bodyTerms = append(bodyTerms, ast.NewAtom(fp))
		}
		body := ast.NewAtom("body", bodyTerms...)

		macroName := fmt.Sprintf("m%d", trial)
		defn := makeMacroDef(macroName, fparams, body)
		macros := map[string]interface{}{
			macroName: defn,
		}

		inst := ast.NewAtom(macroName, aparams...)
		result := instantiateMacro(inst, macros)
		if result == nil {
			t.Fatalf("trial %d: instantiateMacro returned nil", trial)
		}

		resultAtom, ok := result.(*ast.Atom)
		if !ok {
			t.Fatalf("trial %d: expected *ast.Atom, got %T", trial, result)
		}
		if resultAtom.Rep != "body" {
			t.Errorf("trial %d: expected 'body', got %q", trial, resultAtom.Rep)
		}
		if len(resultAtom.Terms) != nParams {
			t.Fatalf("trial %d: expected %d terms, got %d", trial, nParams, len(resultAtom.Terms))
		}

		// Each formal param should be replaced by the corresponding actual param
		for i, term := range resultAtom.Terms {
			a, ok := term.(*ast.Atom)
			if !ok {
				t.Errorf("trial %d, param %d: expected Atom, got %T", trial, i, term)
				continue
			}
			if a.Rep != apNames[i] {
				t.Errorf("trial %d, param %d: expected %q, got %q", trial, i, apNames[i], a.Rep)
			}
		}
	}
}

func TestInstantiateMacroRandomizedNestedBodies(t *testing.T) {
	// Generate macros with randomly nested bodies and verify substitution.
	rng := rand.New(rand.NewSource(123))

	for trial := 0; trial < 50; trial++ {
		nParams := 1 + rng.Intn(3) // 1-3 params
		depth := 1 + rng.Intn(4)   // 1-4 nesting depth

		fparams := make([]string, nParams)
		for i := range fparams {
			fparams[i] = fmt.Sprintf("fp%d", i)
		}

		aparams := make([]ast.Node, nParams)
		apNames := make([]string, nParams)
		for i := range aparams {
			apNames[i] = fmt.Sprintf("ap%d_%d", trial, i)
			aparams[i] = ast.NewAtom(apNames[i])
		}

		// Build a nested body: f0(f1(f2(...(fp0)...)))
		var body ast.Node = ast.NewAtom(fparams[0])
		for d := 0; d < depth; d++ {
			body = ast.NewAtom(fmt.Sprintf("f%d", d), body)
		}

		macroName := fmt.Sprintf("nested%d", trial)
		defn := makeMacroDef(macroName, fparams, body)
		macros := map[string]interface{}{
			macroName: defn,
		}

		inst := ast.NewAtom(macroName, aparams...)
		result := instantiateMacro(inst, macros)
		if result == nil {
			t.Fatalf("trial %d: instantiateMacro returned nil", trial)
		}

		// Walk down the nesting and verify the leaf is the actual param
		node := result
		for d := depth - 1; d >= 0; d-- {
			atom, ok := node.(*ast.Atom)
			if !ok {
				t.Fatalf("trial %d, depth %d: expected *ast.Atom, got %T", trial, d, node)
			}
			expected := fmt.Sprintf("f%d", d)
			if atom.Rep != expected {
				t.Errorf("trial %d, depth %d: expected %q, got %q", trial, d, expected, atom.Rep)
			}
			if len(atom.Terms) != 1 {
				t.Fatalf("trial %d, depth %d: expected 1 term, got %d", trial, d, len(atom.Terms))
			}
			node = atom.Terms[0]
		}
		// Leaf should be the substituted actual param (zero-arity Atom)
		leafAtom, ok := node.(*ast.Atom)
		if !ok {
			t.Fatalf("trial %d: leaf should be Atom, got %T", trial, node)
		}
		if leafAtom.Rep != apNames[0] {
			t.Errorf("trial %d: leaf should be %q, got %q", trial, apNames[0], leafAtom.Rep)
		}
	}
}

func TestInstantiateMacroRandomizedMissing(t *testing.T) {
	// Verify that random names not in the macros map always return nil
	rng := rand.New(rand.NewSource(7))
	macros := map[string]interface{}{
		"existing": makeMacroDef("existing", nil, ast.NewAtom("body")),
	}

	for trial := 0; trial < 100; trial++ {
		name := fmt.Sprintf("notfound%d", rng.Intn(1000))
		inst := ast.NewAtom(name)
		result := instantiateMacro(inst, macros)
		if result != nil {
			t.Errorf("trial %d: expected nil for missing macro %q, got %s", trial, name, result)
		}
	}
}

// -----------------------------------------------------------------------
// Integration test: UpdateContext.CompileActionBody wiring
// -----------------------------------------------------------------------

func TestUpdateContextCompileActionBodyField(t *testing.T) {
	ctx := &UpdateContext{
		Domain: module.New(),
		PVars:  nil,
	}
	// Should be nil by default
	if ctx.CompileActionBody != nil {
		t.Error("CompileActionBody should be nil by default")
	}

	// Set it and verify it works
	called := false
	ctx.CompileActionBody = func(node ast.Node) (Action, error) {
		called = true
		return NewSequence(), nil
	}

	act, err := ctx.CompileActionBody(ast.NewSymbol("test", nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("callback not called")
	}
	if act == nil {
		t.Error("callback returned nil")
	}
}

// -----------------------------------------------------------------------
// Integration test: full macro expansion through IntUpdate
// -----------------------------------------------------------------------

func TestInstantiateActionMacroExpansionEndToEnd(t *testing.T) {
	// Scenario: macro double_assume(x) = assume x
	// Action: instantiate double_assume(p)
	// Expected: macro expands to "assume p", which compiles to AssumeAction(p)
	body := ast.NewAtom("assume", ast.NewAtom("x"))
	defn := makeMacroDef("double_assume", []string{"x"}, body)

	mod := module.New()
	mod.Macros = map[string]interface{}{
		"double_assume": defn,
	}

	// Create InstantiateAction with AST
	a := NewInstantiateAction(nil)
	a.AstInst = ast.NewAtom("double_assume", ast.NewAtom("p"))

	p := lg.NewSymbol("p", lg.Boolean)

	ctx := &UpdateContext{
		Domain: mod,
		PVars:  nil,
		CompileActionBody: func(node ast.Node) (Action, error) {
			// Verify the rewritten AST is "assume p"
			atom, ok := node.(*ast.Atom)
			if !ok {
				return nil, fmt.Errorf("expected Atom, got %T", node)
			}
			if atom.Rep != "assume" {
				return nil, fmt.Errorf("expected 'assume', got %q", atom.Rep)
			}
			if len(atom.Terms) != 1 {
				return nil, fmt.Errorf("expected 1 term, got %d", len(atom.Terms))
			}
			termAtom, ok := atom.Terms[0].(*ast.Atom)
			if !ok || termAtom.Rep != "p" {
				return nil, fmt.Errorf("expected 'p', got %s", atom.Terms[0])
			}
			return NewAssumeAction(p), nil
		},
	}

	u := a.IntUpdate(ctx)
	if u == nil {
		t.Fatal("IntUpdate returned nil")
	}
	// AssumeAction(p) should produce TR with p, Pre = false
	if u.TR == nil {
		t.Fatal("TR should be non-nil")
	}
	if !u.Pre.IsFalse() {
		t.Error("Pre should be false for assume action")
	}
}

// -----------------------------------------------------------------------
// Fuzz test for instantiateMacro
// -----------------------------------------------------------------------

func FuzzInstantiateMacro(f *testing.F) {
	f.Add("m", "x", "y", "actual")
	f.Add("foo", "a", "body", "bar")
	f.Add("inc", "n", "plus_one", "count")
	f.Add("", "x", "y", "z")

	f.Fuzz(func(t *testing.T, macroName, paramName, bodyName, actualName string) {
		// Skip empty names that would cause issues
		if macroName == "" || paramName == "" || bodyName == "" || actualName == "" {
			return
		}
		// Skip names with special chars that might confuse the AST
		for _, s := range []string{macroName, paramName, bodyName, actualName} {
			if strings.ContainsAny(s, "\x00\n\r\t") {
				return
			}
		}

		body := ast.NewAtom(bodyName, ast.NewAtom(paramName))
		defn := makeMacroDef(macroName, []string{paramName}, body)
		macros := map[string]interface{}{
			macroName: defn,
		}

		inst := ast.NewAtom(macroName, ast.NewAtom(actualName))

		// Should not panic
		result := instantiateMacro(inst, macros)
		if result == nil {
			t.Error("expected non-nil result for valid macro")
			return
		}

		// Result should be an Atom with bodyName as Rep
		atom, ok := result.(*ast.Atom)
		if !ok {
			t.Errorf("expected *ast.Atom, got %T", result)
			return
		}
		if atom.Rep != bodyName {
			t.Errorf("expected body name %q, got %q", bodyName, atom.Rep)
		}
		if len(atom.Terms) != 1 {
			t.Errorf("expected 1 term, got %d", len(atom.Terms))
			return
		}

		// The term should be the actual param (substituted, as Atom)
		a, ok := atom.Terms[0].(*ast.Atom)
		if !ok {
			// If paramName == bodyName, the subst might replace the Atom
			// itself. This is an edge case that depends on the rewriter.
			return
		}
		if a.Rep != actualName {
			t.Errorf("expected actual param %q, got %q", actualName, a.Rep)
		}
	})
}

// -----------------------------------------------------------------------
// Test extractInstInfo
// -----------------------------------------------------------------------

func TestExtractInstInfoSymbol(t *testing.T) {
	sym := lg.NewSymbol("foo", lg.TopS)
	name, args := extractInstInfo(sym)
	if name != "foo" {
		t.Errorf("expected 'foo', got %q", name)
	}
	if args != nil {
		t.Errorf("expected nil args, got %v", args)
	}
}

func TestExtractInstInfoApply(t *testing.T) {
	fSym := lg.NewSymbol("bar", lg.TopS)
	arg := lg.NewSymbol("x", lg.TopS)
	app := &lg.Apply{Func: fSym, Terms: []lg.Expr{arg}}
	name, args := extractInstInfo(app)
	if name != "bar" {
		t.Errorf("expected 'bar', got %q", name)
	}
	if len(args) != 1 {
		t.Errorf("expected 1 arg, got %d", len(args))
	}
}

func TestExtractInstInfoOther(t *testing.T) {
	name, _ := extractInstInfo(lg.True)
	if name != "" {
		t.Errorf("expected empty name for non-Symbol/Apply, got %q", name)
	}
}
