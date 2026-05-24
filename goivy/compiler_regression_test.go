package goivy

// Regression tests for bugs found while getting ord_live.ivy to compile.

import (
	"fmt"
	"strings"
	"testing"
)

func TestGetSymbolDependenciesIncludesAppliedFunctionDefinitions(t *testing.T) {
	x, err := NewVariable("X", TopS)
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}
	a := NewConst("a", TopS)
	derived := NewConst("derived", TopS)
	defMap := map[NodeKey]interface{}{
		Key(derived): MustApply(a, x),
	}
	deps := make(map[NodeKey]bool)

	GetSymbolDependencies(defMap, deps, MustApply(derived, x))

	if !deps[Key(derived)] {
		t.Fatalf("dependencies did not include applied function %s", derived)
	}
	if !deps[Key(a)] {
		t.Fatalf("dependencies did not follow %s definition to %s", derived, a)
	}
}

// TestRegression_ActionInsideIsolateBody verifies that actions declared inside
// an isolate body are registered in mod.Actions and visible to CreateIsolate.
//
// Bug: "undefined action: bar.a" when running goivy_check on action1.ivy.
// The action was compiled and registered but CreateIsolate couldn't find it.
//
// Root cause: either the action wasn't registered with the right key,
// or the module copy didn't carry it through.
func TestRegression_ActionInsideIsolateBody(t *testing.T) {
	src := `
isolate bar = {
    action a
    invariant true
}
export bar.a
`
	result, err := Parse(src, Version{1, 7})
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	mod := New()
	mod.Cfg = NewConfig()
	if err := IvyCompile(result.Decls, mod, false); err != nil {
		t.Fatalf("IvyCompile error: %v", err)
	}

	// Verify bar.a is registered in mod.Actions (without ext: prefix)
	if _, ok := mod.Actions.Get2("bar.a"); !ok {
		keys := make([]string, 0, mod.Actions.Len())
		for k := range mod.Actions.All() {
			keys = append(keys, k)
		}
		t.Errorf("bar.a not found in mod.Actions; have keys: %v", keys)
	}

	// Verify bar is registered in mod.Isolates
	if _, ok := mod.Isolates["bar"]; !ok {
		keys := make([]string, 0, len(mod.Isolates))
		for k := range mod.Isolates {
			keys = append(keys, k)
		}
		t.Errorf("bar not found in mod.Isolates; have keys: %v", keys)
	}

	// Verify CreateIsolate succeeds
	isoMod := mod.Copy()
	if err := CreateIsolate("bar", isoMod); err != nil {
		t.Errorf("CreateIsolate(bar) failed: %v", err)
	}
}

// TestRegression_ImmutableSymbolAssigned verifies that mutable symbols
// (like boolean vars assigned inside action bodies) are not incorrectly
// flagged as immutable during compilation.
//
// Bug: "immutable symbol assigned: cfabric.rd_pio_fair" when compiling
// ord_live.ivy with isolate=cf_live. The symbol rd_pio_fair is a var
// (mutable) inside object cfabric, but the compiler treats it as immutable.
//
// This test uses a minimal reproduction: an object with a boolean var
// that gets assigned in an action body.
func TestRegression_ImmutableSymbolAssigned(t *testing.T) {
	src := `
type proc

object cfabric = {
    var rd_pio_fair : bool

    action step = {
        rd_pio_fair := true
    }
}

export cfabric.step
`
	result, err := Parse(src, Version{1, 7})
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	mod := New()
	mod.Cfg = NewConfig()
	err = IvyCompile(result.Decls, mod, false)
	if err != nil {
		if strings.Contains(err.Error(), "immutable symbol assigned") {
			t.Errorf("IvyCompile incorrectly flags mutable symbol as immutable: %v", err)
		} else {
			// Other errors may be acceptable for this minimal test
			t.Logf("IvyCompile error (may be unrelated): %v", err)
		}
	}
}

// TestRegression_ActionInsideObjectRegistered verifies that actions defined
// inside objects are registered in mod.Actions after compilation.
//
// Bug: actions like `cfabric.step` (defined as `action step = { ... }` inside
// `object cfabric = { ... }`) were missing from mod.Actions because
// CompileAction was silently failing and the action was registered with an
// empty sequence but the original action body was lost.
//
// Python registers ~50 actions for ord_live.ivy; Go was registering only ~26.
// The missing actions caused the interference check to see fewer modifiers,
// producing a false positive "immutable symbol assigned" error.
func TestRegression_ActionInsideObjectRegistered(t *testing.T) {
	src := `
type bool
type loc_type

object cfabric = {
    individual rd_fair : bool
    individual wr_fair : bool
    individual rd_pio_fair : bool
    individual wr_pio_fair : bool

    after init {
        rd_fair := false;
        wr_fair := false;
        rd_pio_fair := false;
        wr_pio_fair := false
    }

    action step = {
        rd_fair := true;
        wr_fair := true;
        rd_pio_fair := true;
        wr_pio_fair := true
    }
}

export cfabric.step
`
	result, err := Parse(src, Version{1, 8})
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	mod := New()
	mod.Cfg = NewConfig()
	if err := IvyCompile(result.Decls, mod, false); err != nil {
		t.Fatalf("IvyCompile error: %v", err)
	}

	// Step 1: cfabric.step must be in mod.Actions after IvyCompile
	if _, ok := mod.Actions.Get2("cfabric.step"); !ok {
		keys := make([]string, 0, mod.Actions.Len())
		for k := range mod.Actions.All() {
			keys = append(keys, k)
		}
		t.Fatalf("cfabric.step not found in mod.Actions after IvyCompile; have: %v", keys)
	}

	// Step 2: cfabric.step must not be an empty action
	stepAction := mod.Actions.Get("cfabric.step")
	s := fmt.Sprintf("%v", stepAction)
	t.Logf("cfabric.step = %s", s)
	if s == "true" || s == "" || s == "Sequence()" {
		t.Errorf("cfabric.step is empty (CompileAction likely failed silently): %v", stepAction)
	}

	// Verify the interference check passes (rd_pio_fair is modified by
	// cfabric.step, so it should NOT be flagged as immutable)
	mod2 := New()
	mod2.Cfg = NewConfig()
	err = IvyCompile(result.Decls, mod2, false)
	if err != nil {
		if strings.Contains(err.Error(), "immutable symbol assigned") {
			t.Errorf("false positive: interference check incorrectly flagged a mutable symbol: %v", err)
		} else {
			t.Logf("IvyCompile error (may be unrelated): %v", err)
		}
	}
}

// TestRegression_ActionInsideObjectWithIsolate verifies that actions inside
// objects survive isolate creation. The full flow is:
//  1. Parser expands object cfabric → cfabric.step action
//  2. IvyCompile registers cfabric.step in mod.Actions
//  3. CheckModule copies module, calls CreateIsolate("live")
//  4. CreateIsolate must still find cfabric.step
//
// This matches the ord_live.ivy flow where cfabric.step was missing after
// isolate creation, causing the interference check to produce a false
// positive "immutable symbol assigned" error.
func TestRegression_ActionInsideObjectWithIsolate(t *testing.T) {
	src := `
type bool

object cfabric = {
    individual rd_pio_fair : bool

    after init {
        rd_pio_fair := false
    }

    action step = {
        rd_pio_fair := true
    }
}

export cfabric.step

isolate live = {
    function issued(T:bool) = cfabric.rd_pio_fair
} with this
`
	result, err := Parse(src, Version{1, 8})
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	mod := New()
	mod.Cfg = NewConfig()
	if err := IvyCompile(result.Decls, mod, false); err != nil {
		t.Fatalf("IvyCompile error: %v", err)
	}

	// Step 1: cfabric.step must be in mod.Actions after IvyCompile
	if _, ok := mod.Actions.Get2("cfabric.step"); !ok {
		keys := make([]string, 0, mod.Actions.Len())
		for k := range mod.Actions.All() {
			keys = append(keys, k)
		}
		t.Fatalf("cfabric.step not found in mod.Actions after IvyCompile; have: %v", keys)
	}

	// Step 2: cfabric.step must not be an empty action
	stepAction := mod.Actions.Get("cfabric.step")
	s := fmt.Sprintf("%v", stepAction)
	t.Logf("cfabric.step = %s", s)
	if s == "true" || s == "" || s == "Sequence()" {
		t.Errorf("cfabric.step is empty (CompileAction likely failed silently): %v", stepAction)
	}

	// Step 3: Copy the module (simulating CheckModule's mod.Copy())
	isoMod := mod.Copy()

	// Step 4: cfabric.step must survive the copy
	if _, ok := isoMod.Actions.Get2("cfabric.step"); !ok {
		keys := make([]string, 0, isoMod.Actions.Len())
		for k := range isoMod.Actions.All() {
			keys = append(keys, k)
		}
		t.Fatalf("cfabric.step not in copied mod.Actions; have: %v", keys)
	}

	// Step 5: CreateIsolate("live") must succeed
	if err := CreateIsolate("live", isoMod); err != nil {
		t.Errorf("CreateIsolate(live) failed: %v", err)
		for k := range isoMod.Actions.All() {
			t.Logf("  isoMod.Actions has: %s", k)
		}
	}
}

// TestRegression_VarInActionBody verifies that `var` declarations inside
// action bodies are lowered to `local` scopes before compilation.
//
// Bug: Python's lower_var_stmts() (ivy_parser.py:2324-2350) transforms
// `var p : proc; stmts...` into `local p : proc { stmts... }`. Go's
// parseSequence() did NOT call any lowering, so `var p : proc` produced
// Atom("var", Variable("p")) with 1 term. The compiler expected
// Atom("local", varDecl, body) with 2+ terms → "local needs variables and body".
func TestRegression_VarInActionBody(t *testing.T) {
	src := `
type bool
type proc
type mem_type

object cfabric = {
    individual rd_pio_fair : bool
    individual wr_pio_fair : bool

    action step(ph:bool, sel_memc:bool) = {
        var p : proc;
        var m : mem_type;
        var wr : bool;
        var rd : bool;
        wr := false;
        rd := false;
        rd_pio_fair := ph & ~sel_memc;
        rd_pio_fair := false;
        wr_pio_fair := ph & sel_memc;
        wr_pio_fair := false
    }
}
export cfabric.step
`
	result, err := Parse(src, Version{1, 8})
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	mod := New()
	mod.Cfg = NewConfig()
	err = IvyCompile(result.Decls, mod, false)
	if err != nil {
		if strings.Contains(err.Error(), "local needs variables and body") {
			t.Errorf("var inside action body not lowered to local: %v", err)
		} else {
			t.Logf("IvyCompile error (may be unrelated): %v", err)
		}
	}
	// cfabric.step must be registered with a body that modifies symbols.
	// If CompileAction failed silently, it gets an empty Sequence with no modifies.
	if act, ok := mod.Actions.Get2("cfabric.step"); !ok {
		keys := make([]string, 0, mod.Actions.Len())
		for k := range mod.Actions.All() {
			keys = append(keys, k)
		}
		t.Errorf("cfabric.step missing from Actions; have: %v", keys)
	} else {
		s := fmt.Sprintf("%v", act)
		t.Logf("cfabric.step = %s", s)
		// The action body must contain assignments to rd_pio_fair.
		// If it's empty ({} or true), CompileAction failed silently.
		if !strings.Contains(s, "rd_pio_fair") {
			t.Errorf("cfabric.step body should assign rd_pio_fair but doesn't: %s", s)
		}
	}
}

// TestRegression_AliasTypeInModuleAction verifies that type aliases in
// module actions are properly rewritten during module expansion.
//
// Bug: When `module mymod = { type this; alias t = this; action next(x:t) }`
// is expanded as `instance idx : mymod`, the alias becomes `idx.t → idx`,
// but the action parameter `x:t` was NOT rewritten to `x:idx.t`.
// CmplSort("t") → ResolveAlias("t") → not found → "unknown type: t".
func TestRegression_AliasTypeInModuleAction(t *testing.T) {
	src := `
module mymod = {
    type this
    alias t = this
    action next(x:t) returns (y:t)
}
instance idx : mymod
`
	result, err := Parse(src, Version{1, 8})
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	mod := New()
	mod.Cfg = NewConfig()
	err = IvyCompile(result.Decls, mod, false)
	if err != nil {
		if strings.Contains(err.Error(), "unknown type: t") {
			t.Errorf("alias type 't' not resolved after module expansion: %v", err)
		} else {
			t.Logf("IvyCompile error (may be unrelated): %v", err)
		}
	}
	// idx.next must be registered (forward declaration → empty body is OK)
	if _, ok := mod.Actions.Get2("idx.next"); !ok {
		keys := make([]string, 0, mod.Actions.Len())
		for k := range mod.Actions.All() {
			keys = append(keys, k)
		}
		t.Errorf("idx.next missing from Actions; have: %v", keys)
	}
	// The alias idx.t must resolve to idx
	if alias, ok := mod.Aliases["idx.t"]; !ok || alias != "idx" {
		t.Errorf("alias idx.t should resolve to 'idx', got: %v (exists=%v)", mod.Aliases["idx.t"], ok)
	}
	// The sort idx must exist
	if _, err2 := mod.Sig.FindSort("idx", false); err2 != nil {
		t.Errorf("sort 'idx' not found: %v", err2)
	}
	// The error must NOT be "unknown type: t"
	if err != nil && strings.Contains(err.Error(), "unknown type: t") {
		t.Errorf("alias type 't' not resolved after module expansion: %v", err)
	}
}

func TestOtherThingSortInferRootPreservesAstHavocAction(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()
	idSort := &UninterpretedSort{Name: "id"}
	if err := c.Sig.AddSort(idSort); err != nil {
		t.Fatalf("add sort: %v", err)
	}
	if _, err := c.Sig.AddSymbol("src", idSort); err != nil {
		t.Fatalf("add symbol: %v", err)
	}

	seqNode := cfg.NewSequence(cfg.NewHavocAction(cfg.NewAtom("src")))
	result, err := c.CompileActionBody(seqNode)
	if err != nil {
		t.Fatalf("compile sequence with havoc: %v", err)
	}
	seq, ok := result.(*LogicSequence)
	if !ok {
		t.Fatalf("expected compiled sequence, got %T", result)
	}
	if len(seq.Elems) != 1 {
		t.Fatalf("expected one sequence element, got %d", len(seq.Elems))
	}
	havoc, ok := seq.Elems[0].(*LogicHavocAction)
	if !ok {
		t.Fatalf("expected havoc action to be preserved, got %T: %s", seq.Elems[0], seq.Elems[0])
	}
	target, ok := havoc.Target.(*Const)
	if !ok {
		t.Fatalf("expected havoc target const, got %T", havoc.Target)
	}
	if target.Name != "src" || !SortEqual(target.CSort, idSort) {
		t.Fatalf("unexpected havoc target: %s : %s", target.Name, target.CSort)
	}
}
