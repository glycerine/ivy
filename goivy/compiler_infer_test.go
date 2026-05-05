package goivy

import (
	"strings"
	"testing"
)

// Helper: build an ActionDecl containing ActionDefs.
// Python: Decl("action", [actionDef1, actionDef2, ...])
func makeActionDecl(actionDefs ...Node) Node {
	cfg := NewAstConfig()
	return cfg.NewActionDecl(actionDefs...)
}

// Helper: build a MixinDecl containing mixin defs.
// Python: Decl("mixin", [mixinDef1, mixinDef2, ...])
func makeMixinDecl(mixinDefs ...Node) Node {
	cfg := NewAstConfig()
	return cfg.NewMixinDecl(mixinDefs...)
}

// Helper: build a MixinAfterDef (mixer after mixee).
func makeMixinAfter(mixerName, mixeeName string) *MixinAfterDef {
	cfg := NewAstConfig()
	return cfg.NewMixinAfterDef(cfg.NewAtom(mixerName), cfg.NewAtom(mixeeName))
}

// Helper: build a simple ActionDef with the given name, body, params, returns.
// Note: NewActionDef auto-prefixes params/returns with "fml:" and rewrites the body.
func makeActionDef(name string, body Node, params []Node, returns []Node) *ActionDef {
	cfg := NewAstConfig()
	nameAtom := cfg.NewAtom(name)
	if body == nil {
		body = cfg.NewAtom("skip")
	}
	return cfg.NewActionDef(nameAtom, body, params, returns)
}

// Helper: build an ActionDef whose Name atom has signature args (like action foo(x,y)).
func makeActionDefWithSigArgs(name string, sigArgs []Node, body Node, params []Node, returns []Node) *ActionDef {
	cfg := NewAstConfig()
	nameAtom := cfg.NewAtom(name, sigArgs...)
	if body == nil {
		body = cfg.NewAtom("skip")
	}
	return cfg.NewActionDef(nameAtom, body, params, returns)
}

// =============================================================================
// InferParameters Tests
// =============================================================================

// TestInferParameters_NoMixins: no mixin declarations, actions unchanged.
func TestInferParameters_NoMixins(t *testing.T) {
	cfg := NewAstConfig()
	action := makeActionDef("foo", nil, []Node{cfg.NewAtom("x")}, nil)
	decls := []Node{makeActionDecl(action)}

	err := InferParameters(decls)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// FormalParams should be unchanged (just the original "fml:x")
	if len(action.FormalParams) != 1 {
		t.Errorf("expected 1 formal param, got %d", len(action.FormalParams))
	}
}

// TestInferParameters_UndefinedAction: mixin references a non-existent action.
func TestInferParameters_UndefinedAction(t *testing.T) {
	action := makeActionDef("foo", nil, nil, nil)
	mixin := makeMixinAfter("foo", "nonexistent")

	decls := []Node{
		makeActionDecl(action),
		makeMixinDecl(mixin),
	}

	err := InferParameters(decls)
	if err == nil {
		t.Fatal("expected error for undefined action, got nil")
	}
	if !strings.Contains(err.Error(), "undefined action") {
		t.Errorf("expected 'undefined action' error, got: %v", err)
	}
}

// TestInferParameters_SkipInit: mixin to "init" is silently skipped.
func TestInferParameters_SkipInit(t *testing.T) {
	action := makeActionDef("foo", nil, nil, nil)
	mixin := makeMixinAfter("foo", "init")

	decls := []Node{
		makeActionDecl(action),
		makeMixinDecl(mixin),
	}

	err := InferParameters(decls)
	if err != nil {
		t.Fatalf("expected no error for init mixin, got: %v", err)
	}
}

// TestInferParameters_TooManyInputParams: monitor has more input params than mixee.
func TestInferParameters_TooManyInputParams(t *testing.T) {
	cfg := NewAstConfig()
	// mixer: action monitor(x, y) with formal params [a, b, c]
	// mixee: action target() with formal params [p]
	// total mixer inputs = 2 (sig) + 3 (formals) = 5
	// total mixee inputs = 0 (sig) + 1 (formals) = 1
	// 5 > 1 → error
	mixer := makeActionDefWithSigArgs("monitor",
		[]Node{cfg.NewAtom("x"), cfg.NewAtom("y")},
		nil,
		[]Node{cfg.NewAtom("a"), cfg.NewAtom("b"), cfg.NewAtom("c")},
		nil,
	)
	mixee := makeActionDef("target", nil, []Node{cfg.NewAtom("p")}, nil)
	mixin := makeMixinAfter("monitor", "target")

	decls := []Node{
		makeActionDecl(mixer, mixee),
		makeMixinDecl(mixin),
	}

	err := InferParameters(decls)
	if err == nil {
		t.Fatal("expected error for too many input parameters")
	}
	if !strings.Contains(err.Error(), "too many input parameters") {
		t.Errorf("expected 'too many input parameters' error, got: %v", err)
	}
}

// TestInferParameters_TooManyOutputParams: monitor has more output params than mixee.
func TestInferParameters_TooManyOutputParams(t *testing.T) {
	cfg := NewAstConfig()
	mixer := makeActionDef("monitor", nil, nil,
		[]Node{cfg.NewAtom("r1"), cfg.NewAtom("r2")})
	mixee := makeActionDef("target", nil, nil,
		[]Node{cfg.NewAtom("r1")})
	mixin := makeMixinAfter("monitor", "target")

	decls := []Node{
		makeActionDecl(mixer, mixee),
		makeMixinDecl(mixin),
	}

	err := InferParameters(decls)
	if err == nil {
		t.Fatal("expected error for too many output parameters")
	}
	if !strings.Contains(err.Error(), "too many output parameters") {
		t.Errorf("expected 'too many output parameters' error, got: %v", err)
	}
}

// TestInferParameters_RequiredParamsNotMet: monitor doesn't supply enough explicit params.
func TestInferParameters_RequiredParamsNotMet(t *testing.T) {
	cfg := NewAstConfig()
	// mixee has 2 sig args, mixer has 0 sig args → required = 2
	// mixer has 1 formal param → 1 < 2 → error
	mixer := makeActionDef("monitor", nil,
		[]Node{cfg.NewAtom("a")}, nil)
	mixee := makeActionDefWithSigArgs("target",
		[]Node{cfg.NewAtom("x"), cfg.NewAtom("y")},
		nil,
		[]Node{cfg.NewAtom("p")}, nil)
	mixin := makeMixinAfter("monitor", "target")

	decls := []Node{
		makeActionDecl(mixer, mixee),
		makeMixinDecl(mixin),
	}

	err := InferParameters(decls)
	if err == nil {
		t.Fatal("expected error for insufficient explicit parameters")
	}
	if !strings.Contains(err.Error(), "must supply at least") {
		t.Errorf("expected 'must supply at least' error, got: %v", err)
	}
}

// TestInferParameters_ExtendsFormals: extra params/returns from mixee are added to mixer.
func TestInferParameters_ExtendsFormals(t *testing.T) {
	cfg := NewAstConfig()
	// mixer: action monitor() with formal params [] and returns []
	// mixee: action target() with formal params [p, q] and returns [r]
	// After inference, monitor should gain fml:p, fml:q as formals and fml:r as returns
	mixer := makeActionDef("monitor", nil, nil, nil)
	mixee := makeActionDef("target", nil,
		[]Node{cfg.NewAtom("p"), cfg.NewAtom("q")},
		[]Node{cfg.NewAtom("r")})
	mixin := makeMixinAfter("monitor", "target")

	decls := []Node{
		makeActionDecl(mixer, mixee),
		makeMixinDecl(mixin),
	}

	err := InferParameters(decls)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// mixer should now have 2 formal params (fml:p, fml:q) and 1 formal return (fml:r)
	if len(mixer.FormalParams) != 2 {
		t.Errorf("expected 2 formal params, got %d", len(mixer.FormalParams))
	}
	if len(mixer.FormalReturns) != 1 {
		t.Errorf("expected 1 formal return, got %d", len(mixer.FormalReturns))
	}

	// Verify the names are fml:-prefixed
	for i, fp := range mixer.FormalParams {
		if atom, ok := fp.(*Atom); ok {
			if !strings.HasPrefix(atom.Rep, "fml:") {
				t.Errorf("formal param %d: expected fml: prefix, got %q", i, atom.Rep)
			}
		} else {
			t.Errorf("formal param %d: expected *ast.Atom, got %T", i, fp)
		}
	}
	for i, fr := range mixer.FormalReturns {
		if atom, ok := fr.(*Atom); ok {
			if !strings.HasPrefix(atom.Rep, "fml:") {
				t.Errorf("formal return %d: expected fml: prefix, got %q", i, atom.Rep)
			}
		} else {
			t.Errorf("formal return %d: expected *ast.Atom, got %T", i, fr)
		}
	}
}

// TestInferParameters_BodyRewritten: body AST is rewritten with substituted param names.
func TestInferParameters_BodyRewritten(t *testing.T) {
	cfg := NewAstConfig()
	// mixer: action monitor() with body referencing "p"
	// mixee: action target() with formal param [p]
	// After inference, the body's reference to "p" should become "fml:p"
	body := cfg.NewAtom("p") // references the param by unprefixed name
	mixer := makeActionDef("monitor", body, nil, nil)
	mixee := makeActionDef("target", nil,
		[]Node{cfg.NewAtom("p")}, nil)
	mixin := makeMixinAfter("monitor", "target")

	decls := []Node{
		makeActionDecl(mixer, mixee),
		makeMixinDecl(mixin),
	}

	err := InferParameters(decls)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The body should have been rewritten: "p" → "fml:p"
	bodyAtom, ok := mixer.Body.(*Atom)
	if !ok {
		t.Fatalf("expected body to be *ast.Atom, got %T", mixer.Body)
	}
	if bodyAtom.Rep != "fml:p" {
		t.Errorf("expected body atom to be 'fml:p', got %q", bodyAtom.Rep)
	}
}

// TestInferParameters_MultipleMixees: action with >1 mixee is skipped.
func TestInferParameters_MultipleMixees(t *testing.T) {
	cfg := NewAstConfig()
	monitor := makeActionDef("monitor", nil, nil, nil)
	target1 := makeActionDef("target1", nil,
		[]Node{cfg.NewAtom("p")}, nil)
	target2 := makeActionDef("target2", nil,
		[]Node{cfg.NewAtom("q")}, nil)
	mixin1 := makeMixinAfter("monitor", "target1")
	mixin2 := makeMixinAfter("monitor", "target2")

	decls := []Node{
		makeActionDecl(monitor, target1, target2),
		makeMixinDecl(mixin1, mixin2),
	}

	err := InferParameters(decls)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// monitor should NOT have been extended (multiple mixees → skip)
	if len(monitor.FormalParams) != 0 {
		t.Errorf("expected 0 formal params (skipped), got %d", len(monitor.FormalParams))
	}
}

// TestInferParameters_NoExtrasNeeded: all params already supplied, no changes.
func TestInferParameters_NoExtrasNeeded(t *testing.T) {
	cfg := NewAstConfig()
	// mixer and mixee have same number of params/returns → no extension
	mixer := makeActionDef("monitor", nil,
		[]Node{cfg.NewAtom("a")},
		[]Node{cfg.NewAtom("r")})
	mixee := makeActionDef("target", nil,
		[]Node{cfg.NewAtom("p")},
		[]Node{cfg.NewAtom("q")})
	mixin := makeMixinAfter("monitor", "target")

	decls := []Node{
		makeActionDecl(mixer, mixee),
		makeMixinDecl(mixin),
	}

	err := InferParameters(decls)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should remain 1 param and 1 return (no extension needed)
	if len(mixer.FormalParams) != 1 {
		t.Errorf("expected 1 formal param (unchanged), got %d", len(mixer.FormalParams))
	}
	if len(mixer.FormalReturns) != 1 {
		t.Errorf("expected 1 formal return (unchanged), got %d", len(mixer.FormalReturns))
	}
}

// TestInferParameters_PartialExtension: mixer supplies some but not all params.
func TestInferParameters_PartialExtension(t *testing.T) {
	cfg := NewAstConfig()
	// mixee: target(x) with formal params [p, q, r]
	// mixer: monitor() with formal params [a]
	// nparms=0 (mixer has no sig args), mnparms=1 (mixee has 1 sig arg)
	// required = 1 - 0 = 1, mixer has 1 formal → ok
	// combined = [x, fml:p, fml:q, fml:r]  (mixee sig args + mixee formals)
	// skip = len(mixer.formals) + nparms = 1 + 0 = 1
	// xtraps = combined[1:] = [fml:p, fml:q, fml:r]
	mixer := makeActionDef("monitor", nil,
		[]Node{cfg.NewAtom("a")}, nil)
	mixee := makeActionDefWithSigArgs("target",
		[]Node{cfg.NewAtom("x")},
		nil,
		[]Node{cfg.NewAtom("p"), cfg.NewAtom("q"), cfg.NewAtom("r")},
		nil)
	mixin := makeMixinAfter("monitor", "target")

	decls := []Node{
		makeActionDecl(mixer, mixee),
		makeMixinDecl(mixin),
	}

	err := InferParameters(decls)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// mixer started with 1 formal, should gain 3 more from mixee
	if len(mixer.FormalParams) != 4 {
		t.Errorf("expected 4 formal params after extension, got %d", len(mixer.FormalParams))
	}
}
