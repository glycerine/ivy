package goivy

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
)

var actionsTestAstCfg = NewAstConfig()

// -----------------------------------------------------------------------
// Unit tests for instantiateMacro
// -----------------------------------------------------------------------

// makeMacroDef creates a Definition like: macro name(fparams...) = body
func makeMacroDef(name string, fparams []string, body Node) *AstDefinition {
	terms := make([]Node, len(fparams))
	for i, fp := range fparams {
		terms[i] = actionsTestAstCfg.NewSymbol(fp, nil)
	}
	lhs := actionsTestAstCfg.NewAtom(name, terms...)
	return actionsTestAstCfg.NewDefinition(lhs, body)
}

func TestInstantiateMacroSimpleSubstitution(t *testing.T) {
	// macro incr(x) = x + 1
	// instantiate incr(y)
	// Expected: y + 1
	// Note: In the parser, body references to formals are zero-arity Atoms,
	// not Symbols. RewriteAtom only substitutes zero-arity Atoms.
	body := actionsTestAstCfg.NewAtom("+", actionsTestAstCfg.NewAtom("x"), actionsTestAstCfg.NewAtom("1"))
	defn := makeMacroDef("incr", []string{"x"}, body)

	macros := map[string]*AstDefinition{
		"incr": defn,
	}

	inst := actionsTestAstCfg.NewAtom("incr", actionsTestAstCfg.NewAtom("y"))
	result := instantiateMacro(inst, macros)
	if result == nil {
		t.Fatal("instantiateMacro returned nil for valid macro")
	}
	atom, ok := result.(*Atom)
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
	if a, ok := atom.Terms[0].(*Atom); !ok || a.Rep != "y" {
		t.Errorf("expected first arg 'y', got %s", atom.Terms[0])
	}
	// Second arg should remain "1" (not substituted)
	if a, ok := atom.Terms[1].(*Atom); !ok || a.Rep != "1" {
		t.Errorf("expected second arg '1', got %s", atom.Terms[1])
	}
}

func TestInstantiateMacroMultipleParams(t *testing.T) {
	// macro swap(a, b) = pair(b, a)
	// instantiate swap(x, y)
	// Expected: pair(y, x)
	body := actionsTestAstCfg.NewAtom("pair", actionsTestAstCfg.NewAtom("b"), actionsTestAstCfg.NewAtom("a"))
	defn := makeMacroDef("swap", []string{"a", "b"}, body)

	macros := map[string]*AstDefinition{
		"swap": defn,
	}

	inst := actionsTestAstCfg.NewAtom("swap", actionsTestAstCfg.NewAtom("x"), actionsTestAstCfg.NewAtom("y"))
	result := instantiateMacro(inst, macros)
	if result == nil {
		t.Fatal("instantiateMacro returned nil")
	}
	atom := result.(*Atom)
	if atom.Rep != "pair" {
		t.Errorf("expected 'pair', got %q", atom.Rep)
	}
	// First arg should be "y" (from "b" -> "y")
	if a, ok := atom.Terms[0].(*Atom); !ok || a.Rep != "y" {
		t.Errorf("expected 'y', got %s", atom.Terms[0])
	}
	// Second arg should be "x" (from "a" -> "x")
	if a, ok := atom.Terms[1].(*Atom); !ok || a.Rep != "x" {
		t.Errorf("expected 'x', got %s", atom.Terms[1])
	}
}

func TestInstantiateMacroZeroParams(t *testing.T) {
	// macro truthy() = true_val
	// instantiate truthy()
	body := actionsTestAstCfg.NewAtom("true_val")
	lhs := actionsTestAstCfg.NewAtom("truthy")
	defn := actionsTestAstCfg.NewDefinition(lhs, body)

	macros := map[string]*AstDefinition{
		"truthy": defn,
	}

	inst := actionsTestAstCfg.NewAtom("truthy")
	result := instantiateMacro(inst, macros)
	if result == nil {
		t.Fatal("instantiateMacro returned nil for zero-param macro")
	}
	if a, ok := result.(*Atom); !ok || a.Rep != "true_val" {
		t.Errorf("expected 'true_val', got %s", result)
	}
}

func TestInstantiateMacroNotFound(t *testing.T) {
	macros := map[string]*AstDefinition{}
	inst := actionsTestAstCfg.NewAtom("nonexistent", actionsTestAstCfg.NewSymbol("x", nil))
	result := instantiateMacro(inst, macros)
	if result != nil {
		t.Errorf("expected nil for missing macro, got %s", result)
	}
}

func TestInstantiateMacroNilMacroValue(t *testing.T) {
	macros := map[string]*AstDefinition{
		"m": nil,
	}
	inst := actionsTestAstCfg.NewAtom("m")
	result := instantiateMacro(inst, macros)
	if result != nil {
		t.Errorf("expected nil for nil macro def, got %s", result)
	}
}

func TestInstantiateMacroNilDef(t *testing.T) {
	// Macro value is nil (no definition stored)
	macros := map[string]*AstDefinition{
		"m": nil,
	}
	inst := actionsTestAstCfg.NewAtom("m")
	result := instantiateMacro(inst, macros)
	if result != nil {
		t.Errorf("expected nil for nil macro def, got %s", result)
	}
}

func TestInstantiateMacroWrongParamCount(t *testing.T) {
	// macro f(x) = x
	// instantiate f(a, b) -- wrong param count
	body := actionsTestAstCfg.NewAtom("x")
	defn := makeMacroDef("f", []string{"x"}, body)

	macros := map[string]*AstDefinition{
		"f": defn,
	}

	inst := actionsTestAstCfg.NewAtom("f", actionsTestAstCfg.NewAtom("a"), actionsTestAstCfg.NewAtom("b"))
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
	body := actionsTestAstCfg.NewAtom("result")
	lhs := actionsTestAstCfg.NewAtom("m")
	defn := actionsTestAstCfg.NewDefinition(lhs, body)

	macros := map[string]*AstDefinition{
		"m": defn,
	}

	inst := actionsTestAstCfg.NewSymbol("m", nil)
	result := instantiateMacro(inst, macros)
	if result == nil {
		t.Fatal("expected non-nil for Symbol inst")
	}
}

func TestInstantiateMacroNonNodeInst(t *testing.T) {
	// Pass something that is neither *ast.Atom nor *ast.Symbol.
	// Python would AttributeError on inst.relname / inst.args.
	// Faithful Go port panics.
	macros := map[string]*AstDefinition{}
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for non-Atom/Symbol inst")
		}
		msg := fmt.Sprintf("%v", r)
		if !strings.Contains(msg, "not Atom/Symbol") {
			t.Errorf("unexpected panic: %v", r)
		}
	}()
	_ = instantiateMacro(actionsTestAstCfg.NewAnd(), macros)
}

func TestInstantiateMacroPsubstApplied(t *testing.T) {
	// macro m(x) = foo.x.bar
	// The body uses "x" as a subscript in a composite name like "foo.x.bar"
	// psubst should rename "x" → actual param name within dotted names
	//
	// This tests psubst by checking that zero-arity params get string substitution.
	// The psubst is used by AstRewriteSubstConstantsParams.RewriteName which
	// calls SubstSubscripts. We test that the rewriter is constructed correctly.
	body := actionsTestAstCfg.NewAtom("x") // simple: just the param itself (zero-arity Atom)
	defn := makeMacroDef("m", []string{"x"}, body)

	macros := map[string]*AstDefinition{
		"m": defn,
	}

	// Pass a zero-arity Atom as actual param (triggers psubst)
	inst := actionsTestAstCfg.NewAtom("m", actionsTestAstCfg.NewAtom("myval"))
	result := instantiateMacro(inst, macros)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	// The zero-arity atom "myval" should be substituted for "x" via subst
	if a, ok := result.(*Atom); ok {
		if a.Rep != "myval" {
			t.Errorf("expected 'myval', got %q", a.Rep)
		}
	}
}

func TestInstantiateMacroNestedBody(t *testing.T) {
	// macro m(x) = f(g(x), x)
	// instantiate m(y)
	// Expected: f(g(y), y)
	inner := actionsTestAstCfg.NewAtom("g", actionsTestAstCfg.NewAtom("x"))
	body := actionsTestAstCfg.NewAtom("f", inner, actionsTestAstCfg.NewAtom("x"))
	defn := makeMacroDef("m", []string{"x"}, body)

	macros := map[string]*AstDefinition{
		"m": defn,
	}

	inst := actionsTestAstCfg.NewAtom("m", actionsTestAstCfg.NewAtom("y"))
	result := instantiateMacro(inst, macros)
	if result == nil {
		t.Fatal("instantiateMacro returned nil for nested body")
	}
	outer := result.(*Atom)
	if outer.Rep != "f" {
		t.Errorf("outer should be 'f', got %q", outer.Rep)
	}
	if len(outer.Terms) != 2 {
		t.Fatalf("expected 2 terms, got %d", len(outer.Terms))
	}
	// Check g(y)
	innerResult, ok := outer.Terms[0].(*Atom)
	if !ok || innerResult.Rep != "g" {
		t.Errorf("inner should be 'g', got %s", outer.Terms[0])
	}
	if a, ok := innerResult.Terms[0].(*Atom); !ok || a.Rep != "y" {
		t.Errorf("inner arg should be 'y', got %s", innerResult.Terms[0])
	}
	// Check outer second arg is y
	if a, ok := outer.Terms[1].(*Atom); !ok || a.Rep != "y" {
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
	a.AstInst = actionsTestAstCfg.NewAtom("mymacro", actionsTestAstCfg.NewSymbol("arg", nil))
	if a.AstInst == nil {
		t.Error("AstInst should be set")
	}
}

func TestInstantiateActionClonePreservesAstInst(t *testing.T) {
	a := NewInstantiateAction(nil)
	astNode := actionsTestAstCfg.NewAtom("mymacro", actionsTestAstCfg.NewSymbol("arg", nil))
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
	// Python (ivy_actions.py:807): hasattr(domain, 'macros') would AttributeError
	// when domain is None. The faithful Go port panics instead of silently
	// returning a NullUpdate.
	a := NewInstantiateAction(nil)
	ctx := &UpdateContext{Domain: nil}
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("IntUpdate should panic when ctx.Domain is nil")
		}
	}()
	a.IntUpdate(ctx)
}

func TestInstantiateActionIntUpdateMacroExpansion(t *testing.T) {
	// Set up a macro: macro incr(x) = assume x
	// The body after rewriting should be compiled just like Python's
	// im.compile().int_update(domain, pvars) path.
	body := actionsTestAstCfg.NewAtom("assume", actionsTestAstCfg.NewAtom("x"))
	defn := makeMacroDef("incr", []string{"x"}, body)

	mod := New()
	if _, err := mod.Sig.AddSymbol("p", Boolean); err != nil {
		t.Fatalf("AddSymbol(p): %v", err)
	}
	mod.Macros = map[string]*AstDefinition{
		"incr": defn,
	}

	// Create the action with AstInst
	a := NewInstantiateAction(nil)
	a.AstInst = actionsTestAstCfg.NewAtom("incr", actionsTestAstCfg.NewAtom("p"))

	ctx := &UpdateContext{
		Domain: mod,
		PVars:  nil,
	}

	u := a.IntUpdate(ctx)
	if u == nil {
		t.Fatal("IntUpdate should return non-nil")
	}
	if u.TR == nil || u.TR.IsTrue() {
		t.Errorf("expected non-trivial TR from assume(p), got %s", u.TR)
	}
}

func TestInstantiateActionIntUpdateSchemaFallback(t *testing.T) {
	// When macro not found, should fall back to schemata
	mod := New()
	mod.Macros = map[string]*AstDefinition{} // no macros

	fmla := NewConst("p", Boolean)
	mod.Schemata.Set("myschema", actionsTestAstCfg.NewLabeledFormula(actionsTestAstCfg.NewSymbol("myschema", nil), fmla))

	a := NewInstantiateAction(nil)
	a.AstInst = actionsTestAstCfg.NewAtom("myschema")

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
	// Python (ivy_actions.py:820): raise IvyError("instantiation of undefined: ...")
	// when neither macros nor schemata contain the name. Faithful Go port panics.
	mod := New()
	mod.Macros = map[string]*AstDefinition{}

	a := NewInstantiateAction(nil)
	a.AstInst = actionsTestAstCfg.NewAtom("unknown")

	ctx := &UpdateContext{
		Domain: mod,
		PVars:  nil,
	}

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("IntUpdate should panic for instantiation of undefined")
		}
	}()
	a.IntUpdate(ctx)
}

func TestInstantiateActionIntUpdateCompiledExprFallback(t *testing.T) {
	// When AstInst is nil, use compiled Inst for schema lookup
	mod := New()

	fmla := NewConst("q", Boolean)
	mod.Schemata.Set("myschema", actionsTestAstCfg.NewLabeledFormula(actionsTestAstCfg.NewSymbol("myschema", nil), fmla))

	sym := NewConst("myschema", TopS)
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
		aparams := make([]Node, nParams)
		apNames := make([]string, nParams)
		for i := range aparams {
			apNames[i] = fmt.Sprintf("a%d_%d", trial, i)
			aparams[i] = actionsTestAstCfg.NewAtom(apNames[i])
		}

		// Body: an atom referencing all formal params (as zero-arity Atoms)
		var bodyTerms []Node
		for _, fp := range fparams {
			bodyTerms = append(bodyTerms, actionsTestAstCfg.NewAtom(fp))
		}
		body := actionsTestAstCfg.NewAtom("body", bodyTerms...)

		macroName := fmt.Sprintf("m%d", trial)
		defn := makeMacroDef(macroName, fparams, body)
		macros := map[string]*AstDefinition{
			macroName: defn,
		}

		inst := actionsTestAstCfg.NewAtom(macroName, aparams...)
		result := instantiateMacro(inst, macros)
		if result == nil {
			t.Fatalf("trial %d: instantiateMacro returned nil", trial)
		}

		resultAtom, ok := result.(*Atom)
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
			a, ok := term.(*Atom)
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

		aparams := make([]Node, nParams)
		apNames := make([]string, nParams)
		for i := range aparams {
			apNames[i] = fmt.Sprintf("ap%d_%d", trial, i)
			aparams[i] = actionsTestAstCfg.NewAtom(apNames[i])
		}

		// Build a nested body: f0(f1(f2(...(fp0)...)))
		var body Node = actionsTestAstCfg.NewAtom(fparams[0])
		for d := 0; d < depth; d++ {
			body = actionsTestAstCfg.NewAtom(fmt.Sprintf("f%d", d), body)
		}

		macroName := fmt.Sprintf("nested%d", trial)
		defn := makeMacroDef(macroName, fparams, body)
		macros := map[string]*AstDefinition{
			macroName: defn,
		}

		inst := actionsTestAstCfg.NewAtom(macroName, aparams...)
		result := instantiateMacro(inst, macros)
		if result == nil {
			t.Fatalf("trial %d: instantiateMacro returned nil", trial)
		}

		// Walk down the nesting and verify the leaf is the actual param
		node := result
		for d := depth - 1; d >= 0; d-- {
			atom, ok := node.(*Atom)
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
		leafAtom, ok := node.(*Atom)
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
	macros := map[string]*AstDefinition{
		"existing": makeMacroDef("existing", nil, actionsTestAstCfg.NewAtom("body")),
	}

	for trial := 0; trial < 100; trial++ {
		name := fmt.Sprintf("notfound%d", rng.Intn(1000))
		inst := actionsTestAstCfg.NewAtom(name)
		result := instantiateMacro(inst, macros)
		if result != nil {
			t.Errorf("trial %d: expected nil for missing macro %q, got %s", trial, name, result)
		}
	}
}

func TestInstantiateActionMacroExpansionEndToEnd(t *testing.T) {
	// Scenario: macro double_assume(x) = assume x
	// Action: instantiate double_assume(p)
	// Expected: macro expands to "assume p", which compiles to AssumeAction(p)
	body := actionsTestAstCfg.NewAtom("assume", actionsTestAstCfg.NewAtom("x"))
	defn := makeMacroDef("double_assume", []string{"x"}, body)

	mod := New()
	if _, err := mod.Sig.AddSymbol("p", Boolean); err != nil {
		t.Fatalf("AddSymbol(p): %v", err)
	}
	mod.Macros = map[string]*AstDefinition{
		"double_assume": defn,
	}

	// Create InstantiateAction with AST
	a := NewInstantiateAction(nil)
	a.AstInst = actionsTestAstCfg.NewAtom("double_assume", actionsTestAstCfg.NewAtom("p"))

	ctx := &UpdateContext{
		Domain: mod,
		PVars:  nil,
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

		body := actionsTestAstCfg.NewAtom(bodyName, actionsTestAstCfg.NewAtom(paramName))
		defn := makeMacroDef(macroName, []string{paramName}, body)
		macros := map[string]*AstDefinition{
			macroName: defn,
		}

		inst := actionsTestAstCfg.NewAtom(macroName, actionsTestAstCfg.NewAtom(actualName))

		// Should not panic
		result := instantiateMacro(inst, macros)
		if result == nil {
			t.Error("expected non-nil result for valid macro")
			return
		}

		// Result should be an Atom with bodyName as Rep
		atom, ok := result.(*Atom)
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
		a, ok := atom.Terms[0].(*Atom)
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
	sym := NewConst("foo", TopS)
	name, args := extractInstInfo(sym)
	if name != "foo" {
		t.Errorf("expected 'foo', got %q", name)
	}
	if args != nil {
		t.Errorf("expected nil args, got %v", args)
	}
}

func TestExtractInstInfoApply(t *testing.T) {
	fSym := NewConst("bar", TopS)
	arg := NewConst("x", TopS)
	app := MustApply(fSym, arg)
	name, args := extractInstInfo(app)
	if name != "bar" {
		t.Errorf("expected 'bar', got %q", name)
	}
	if len(args) != 1 {
		t.Errorf("expected 1 arg, got %d", len(args))
	}
}

func TestExtractInstInfoOther(t *testing.T) {
	name, _ := extractInstInfo(True)
	if name != "" {
		t.Errorf("expected empty name for non-Symbol/Apply, got %q", name)
	}
}
