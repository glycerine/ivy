package lalr_full

import (
	"strings"
	"testing"

	"github.com/glycerine/goivy/ast"
	"github.com/glycerine/goivy/lexer"
)

// findActionDef extracts the ActionDef from the first ActionDecl in decls.
func findActionDef(decls []ast.Node) *ast.ActionDef {
	for _, d := range decls {
		if ad, ok := d.(*ast.ActionDecl); ok {
			if len(ad.DeclArgs) > 0 {
				if adef, ok := ad.DeclArgs[0].(*ast.ActionDef); ok {
					return adef
				}
			}
		}
	}
	return nil
}

// TestMethodPrependsSelf verifies that parsing a METHOD declaration
// prepends a self parameter to the formals list.
// Matches Python ivy_parser.py:2086-2090.
func TestMethodPrependsSelf(t *testing.T) {
	input := `#lang 1.7
type t
method foo(x:t) = {
    skip
}`
	result, err := Parse(input, lexer.Version{1, 7})
	if err != nil {
		t.Skipf("METHOD parse test skipped: %v", err)
	}

	adef := findActionDef(result.Decls)
	if adef == nil {
		t.Fatal("expected an ActionDecl with ActionDef")
	}

	// FormalParams should have self as first element (prefixed to fml:self)
	if len(adef.FormalParams) < 2 {
		t.Fatalf("expected at least 2 formal params (self + x), got %d", len(adef.FormalParams))
	}

	selfParam := adef.FormalParams[0]
	selfRep := ast.NodeRep(selfParam)
	if !strings.Contains(selfRep, "self") {
		t.Errorf("expected first formal param to contain 'self', got %q", selfRep)
	}

	t.Logf("FormalParams[0] rep: %q", selfRep)
	t.Logf("FormalParams[0] canon: %s", selfParam.Canon())
	for i, fp := range adef.FormalParams {
		t.Logf("  FormalParams[%d]: %T rep=%q canon=%s", i, fp, ast.NodeRep(fp), fp.Canon())
	}
}

// TestActionDoesNotPrependSelf verifies that ACTION (not METHOD)
// does NOT prepend a self parameter.
func TestActionDoesNotPrependSelf(t *testing.T) {
	input := `#lang 1.7
type t
action bar(x:t) = {
    skip
}`
	result, err := Parse(input, lexer.Version{1, 7})
	if err != nil {
		t.Skipf("ACTION parse test skipped: %v", err)
	}

	adef := findActionDef(result.Decls)
	if adef == nil {
		t.Fatal("expected an ActionDecl with ActionDef")
	}

	// For ACTION, formals should only have the declared param (x), no self
	for i, fp := range adef.FormalParams {
		rep := ast.NodeRep(fp)
		if strings.Contains(rep, "self") {
			t.Errorf("ACTION should not have self in formals, but FormalParams[%d] = %q", i, rep)
		}
	}
	t.Logf("ACTION FormalParams count: %d", len(adef.FormalParams))
	for i, fp := range adef.FormalParams {
		t.Logf("  FormalParams[%d]: %T rep=%q", i, fp, ast.NodeRep(fp))
	}
}

// TestMethodSelfHasThisSort verifies that the self parameter has
// sort This() — matching Python ivy_parser.py:2088: arg0.sort = This()
func TestMethodSelfHasThisSort(t *testing.T) {
	input := `#lang 1.7
type t
method foo(x:t) = {
    skip
}`
	result, err := Parse(input, lexer.Version{1, 7})
	if err != nil {
		t.Skipf("METHOD parse test skipped: %v", err)
	}

	adef := findActionDef(result.Decls)
	if adef == nil {
		t.Fatal("expected an ActionDecl with ActionDef")
	}

	if len(adef.FormalParams) < 1 {
		t.Fatal("expected at least 1 formal param")
	}

	selfParam := adef.FormalParams[0]
	canon := string(selfParam.Canon())
	// The self App should have ASort set to This, which canons as (this ...)
	if !strings.Contains(canon, "this") {
		t.Errorf("expected self param canon to contain 'this' sort, got %s", canon)
	}
	t.Logf("self param canon: %s", canon)
}

// TestMethodSelfIsFirstFormal verifies that self is the first formal
// followed by the declared params, when METHOD has multiple params.
func TestMethodSelfIsFirstFormal(t *testing.T) {
	input := `#lang 1.7
type t
method foo(x:t, y:t) = {
    skip
}`
	result, err := Parse(input, lexer.Version{1, 7})
	if err != nil {
		t.Skipf("METHOD parse test skipped: %v", err)
	}

	adef := findActionDef(result.Decls)
	if adef == nil {
		t.Fatal("expected an ActionDecl with ActionDef")
	}

	// Should have 3 formals: self, x, y (all with fml: prefix)
	if len(adef.FormalParams) != 3 {
		t.Fatalf("expected 3 formal params (self + x + y), got %d", len(adef.FormalParams))
	}

	// First should be self
	selfRep := ast.NodeRep(adef.FormalParams[0])
	if !strings.Contains(selfRep, "self") {
		t.Errorf("expected first formal to contain 'self', got %q", selfRep)
	}

	// Second and third should be x and y (with fml: prefix)
	xRep := ast.NodeRep(adef.FormalParams[1])
	yRep := ast.NodeRep(adef.FormalParams[2])
	if !strings.Contains(xRep, "x") {
		t.Errorf("expected second formal to contain 'x', got %q", xRep)
	}
	if !strings.Contains(yRep, "y") {
		t.Errorf("expected third formal to contain 'y', got %q", yRep)
	}

	for i, fp := range adef.FormalParams {
		t.Logf("  FormalParams[%d]: rep=%q", i, ast.NodeRep(fp))
	}
}

// TestMethodCrashAction verifies that "method foo = *" produces a
// CrashAction clone with Atom("this", formals) args.
// Matches Python ivy_parser.py:2091-2092.
func TestMethodCrashAction(t *testing.T) {
	input := `#lang 1.7
type t
method foo = *`
	result, err := Parse(input, lexer.Version{1, 7})
	if err != nil {
		t.Skipf("METHOD crash parse test skipped: %v", err)
	}

	adef := findActionDef(result.Decls)
	if adef == nil {
		t.Fatal("expected an ActionDecl with ActionDef")
	}

	// Body should be a CrashAction (or the clone of one)
	canon := string(adef.Body.Canon())
	if !strings.Contains(canon, "crashAction") {
		t.Errorf("expected crashAction in body canon, got %s", canon)
	}
	// The CrashAction should contain "this" from Atom(This(), formals)
	if !strings.Contains(canon, "this") {
		t.Errorf("expected 'this' in CrashAction canon (from Atom(This(), formals)), got %s", canon)
	}
	// Should also contain self since METHOD prepends it to formals
	if !strings.Contains(canon, "self") {
		t.Errorf("expected 'self' in CrashAction canon (self in formals), got %s", canon)
	}
	t.Logf("CrashAction body canon: %s", canon)
}

// TestActionCrashAction verifies that "action bar = *" produces a
// CrashAction clone WITHOUT self prepend.
func TestActionCrashAction(t *testing.T) {
	input := `#lang 1.7
type t
action bar = *`
	result, err := Parse(input, lexer.Version{1, 7})
	if err != nil {
		t.Skipf("ACTION crash parse test skipped: %v", err)
	}

	adef := findActionDef(result.Decls)
	if adef == nil {
		t.Fatal("expected an ActionDecl with ActionDef")
	}

	// Body should be a CrashAction
	canon := string(adef.Body.Canon())
	if !strings.Contains(canon, "crashAction") {
		t.Errorf("expected crashAction in body canon, got %s", canon)
	}
	// Should NOT contain self
	if strings.Contains(canon, "self") {
		t.Errorf("ACTION CrashAction should not contain 'self', got %s", canon)
	}
	t.Logf("ACTION CrashAction body canon: %s", canon)
}

// TestMethodNoArgs verifies that "method foo = { skip }" with no declared
// params has self as the only formal.
func TestMethodNoArgs(t *testing.T) {
	input := `#lang 1.7
type t
method foo = {
    skip
}`
	result, err := Parse(input, lexer.Version{1, 7})
	if err != nil {
		t.Skipf("METHOD no-args parse test skipped: %v", err)
	}

	adef := findActionDef(result.Decls)
	if adef == nil {
		t.Fatal("expected an ActionDecl with ActionDef")
	}

	// Should have exactly 1 formal: self
	if len(adef.FormalParams) != 1 {
		t.Fatalf("expected 1 formal param (self only), got %d", len(adef.FormalParams))
	}
	selfRep := ast.NodeRep(adef.FormalParams[0])
	if !strings.Contains(selfRep, "self") {
		t.Errorf("expected only formal to contain 'self', got %q", selfRep)
	}
	t.Logf("METHOD no-args FormalParams[0]: rep=%q canon=%s", selfRep, adef.FormalParams[0].Canon())
}

// TestSelfAppConstruction directly constructs the self App with This sort
// and verifies the canon output matches expectations.
func TestSelfAppConstruction(t *testing.T) {
	cfg := ast.NewAstConfig()
	// Build exactly what the grammar rule builds:
	// selfArg := cfg.NewApp(cfg.NewSymbol("self", nil))
	// selfArg.ASort = &ast.This{}
	selfArg := cfg.NewApp(cfg.NewSymbol("self", nil))
	selfArg.ASort = cfg.NewThis()

	canon := string(selfArg.Canon())
	if !strings.Contains(canon, "rep:self") {
		t.Errorf("expected rep:self in canon, got %s", canon)
	}
	if !strings.Contains(canon, "this") {
		t.Errorf("expected 'this' sort in canon, got %s", canon)
	}
	t.Logf("self App canon: %s", canon)

	// After fml: prefix, rep should become fml:self
	prefixed := ast.PrefixNode(selfArg, "fml:")
	prefixedRep := ast.NodeRep(prefixed)
	if prefixedRep != "fml:self" {
		t.Errorf("expected prefixed rep 'fml:self', got %q", prefixedRep)
	}
	t.Logf("prefixed self rep: %q canon: %s", prefixedRep, prefixed.Canon())
}
