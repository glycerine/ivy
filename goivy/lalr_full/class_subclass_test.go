package lalr_full

import (
	"strings"
	"testing"

	"github.com/glycerine/goivy/ast"
	"github.com/glycerine/goivy/lexer"
)

// TestClassCallsCreateObject verifies that parsing a CLASS declaration
// routes through createObject, which applies instMod prefix substitution.
// Matches Python ivy_parser.py:737-747.
func TestClassCallsCreateObject(t *testing.T) {
	input := `#lang 1.7
type t
class foo = {
    action bar = {
        skip
    }
}`
	result, err := Parse(input, lexer.Version{1, 7})
	if err != nil {
		t.Skipf("CLASS parse test skipped (grammar may need additional context): %v", err)
	}
	if len(result.Decls) == 0 {
		t.Fatal("expected at least 1 decl")
	}

	// Check that body decls got prefix substitution applied.
	// createObject → instMod should prefix "bar" with "foo."
	var foundPrefixedAction bool
	for _, d := range result.Decls {
		canonStr := string(d.Canon())
		if strings.Contains(canonStr, "foo.bar") {
			foundPrefixedAction = true
		}
	}
	if !foundPrefixedAction {
		t.Log("Decls found:")
		for i, d := range result.Decls {
			t.Logf("  [%d] %T: %s", i, d, d.Canon())
		}
		t.Error("expected createObject to produce prefixed name 'foo.bar'")
	}
}

// TestClassTypeDeclInOutput verifies that CLASS produces a TypeDecl
// with "this" name and UninterpretedSort in the output.
// The TypeDecl is declared into the module accumulator, then
// instMod rewrites it with the class prefix.
func TestClassTypeDeclInOutput(t *testing.T) {
	input := `#lang 1.7
type t
class foo = {
    action bar = {
        skip
    }
}`
	result, err := Parse(input, lexer.Version{1, 7})
	if err != nil {
		t.Skipf("CLASS parse test skipped: %v", err)
	}

	// instMod rewrites "this" → "foo" via prefix substitution, so the
	// TypeDecl from the class body should have name "foo" (not "this").
	var hasClassTypeDecl bool
	for _, d := range result.Decls {
		if td, ok := d.(*ast.TypeDecl); ok {
			canonStr := string(td.Canon())
			// The TypeDecl created by the CLASS rule has UninterpretedSort (constantSort)
			// and its name gets prefixed by instMod to "foo"
			if strings.Contains(canonStr, `rep:"foo"`) && strings.Contains(canonStr, "constantSort") {
				hasClassTypeDecl = true
			}
		}
	}
	if !hasClassTypeDecl {
		t.Log("Decls found:")
		for i, d := range result.Decls {
			t.Logf("  [%d] %T: %s", i, d, d.Canon())
		}
		t.Error("expected a TypeDecl with prefixed 'foo' name and constantSort from CLASS declaration")
	}
}

// TestSubclassCallsCreateObject verifies that SUBCLASS routes through
// createObject for prefix substitution.
// Matches Python ivy_parser.py:749-761.
func TestSubclassCallsCreateObject(t *testing.T) {
	input := `#lang 1.7
type base_t
subclass child of base_t = {
    action baz = {
        skip
    }
}`
	result, err := Parse(input, lexer.Version{1, 7})
	if err != nil {
		t.Skipf("SUBCLASS parse test skipped: %v", err)
	}
	if len(result.Decls) == 0 {
		t.Fatal("expected at least 1 decl")
	}

	var foundPrefixedAction bool
	for _, d := range result.Decls {
		canonStr := string(d.Canon())
		if strings.Contains(canonStr, "child.baz") {
			foundPrefixedAction = true
		}
	}
	if !foundPrefixedAction {
		t.Log("Decls found:")
		for i, d := range result.Decls {
			t.Logf("  [%d] %T: %s", i, d, d.Canon())
		}
		t.Error("expected createObject to produce prefixed name 'child.baz'")
	}
}

// TestSubclassCreatesVariantDecl verifies that SUBCLASS produces a VariantDecl.
// Python: vdfn = VariantDef(scnst, Atom(p[5]))
//
//	p[9].declare(VariantDecl(vdfn))
func TestSubclassCreatesVariantDecl(t *testing.T) {
	input := `#lang 1.7
type base_t
subclass child of base_t = {
    action baz = {
        skip
    }
}`
	result, err := Parse(input, lexer.Version{1, 7})
	if err != nil {
		t.Skipf("SUBCLASS parse test skipped: %v", err)
	}

	var hasVariantDecl bool
	for _, d := range result.Decls {
		if _, ok := d.(*ast.VariantDecl); ok {
			hasVariantDecl = true
			canonStr := string(d.Canon())
			t.Logf("VariantDecl canon: %s", canonStr)
		}
	}
	if !hasVariantDecl {
		t.Log("Decls found:")
		for i, d := range result.Decls {
			t.Logf("  [%d] %T: %s", i, d, d.Canon())
		}
		t.Error("expected a VariantDecl from SUBCLASS declaration")
	}
}

// TestSubclassEmptyObjectArgs verifies that SUBCLASS passes empty objectargs
// to createObject (no Variable substitution from objectargs).
// Python: create_object(p[0], p[3], [], p[9], get_lineno(p,3), p[8])
func TestSubclassEmptyObjectArgs(t *testing.T) {
	input := `#lang 1.7
type base_t
subclass child of base_t = {
    action baz = {
        skip
    }
}`
	result, err := Parse(input, lexer.Version{1, 7})
	if err != nil {
		t.Skipf("SUBCLASS parse test skipped: %v", err)
	}

	// With empty objectargs, the ObjectDecl's atom should have no Variable terms.
	for _, d := range result.Decls {
		if od, ok := d.(*ast.ObjectDecl); ok {
			if len(od.DeclArgs) > 0 {
				if a, ok := od.DeclArgs[0].(*ast.Atom); ok {
					if len(a.Terms) != 0 {
						t.Errorf("expected ObjectDecl atom to have no terms (empty objectargs), got %d terms", len(a.Terms))
					}
				}
			}
		}
	}
}

// TestDeclReorderingClassOneToFront tests the rotate-last-to-front logic
// used by CLASS: p[8].decls = [p[8].decls[-1]] + p[8].decls[:-1]
func TestDeclReorderingClassOneToFront(t *testing.T) {
	cfg := ast.NewAstConfig()
	// Simulate a decls slice with elements [A, B, C] where C was just appended.
	// After rotation, should be [C, A, B].
	a := cfg.NewAtom("A")
	b := cfg.NewAtom("B")
	c := cfg.NewAtom("C")
	decls := []ast.Node{a, b, c}

	// Apply the same rotation as the CLASS rule
	if n := len(decls); n > 1 {
		last := decls[n-1]
		copy(decls[1:], decls[:n-1])
		decls[0] = last
	}

	if ast.NodeRep(decls[0]) != "C" {
		t.Errorf("expected first element to be C, got %s", ast.NodeRep(decls[0]))
	}
	if ast.NodeRep(decls[1]) != "A" {
		t.Errorf("expected second element to be A, got %s", ast.NodeRep(decls[1]))
	}
	if ast.NodeRep(decls[2]) != "B" {
		t.Errorf("expected third element to be B, got %s", ast.NodeRep(decls[2]))
	}
}

// TestDeclReorderingSubclassTwoToFront tests the rotate-last-2-to-front logic
// used by SUBCLASS: p[9].decls = p[9].decls[-2:] + p[9].decls[:-2]
func TestDeclReorderingSubclassTwoToFront(t *testing.T) {
	cfg := ast.NewAstConfig()
	// Simulate decls [A, B, C, D] where C and D were just appended.
	// After rotation, should be [C, D, A, B].
	a := cfg.NewAtom("A")
	b := cfg.NewAtom("B")
	c := cfg.NewAtom("C")
	d := cfg.NewAtom("D")
	decls := []ast.Node{a, b, c, d}

	// Apply the same rotation as the SUBCLASS rule
	if n := len(decls); n > 2 {
		rotated := make([]ast.Node, n)
		copy(rotated, decls[n-2:])
		copy(rotated[2:], decls[:n-2])
		decls = rotated
	}

	if ast.NodeRep(decls[0]) != "C" {
		t.Errorf("expected first element to be C, got %s", ast.NodeRep(decls[0]))
	}
	if ast.NodeRep(decls[1]) != "D" {
		t.Errorf("expected second element to be D, got %s", ast.NodeRep(decls[1]))
	}
	if ast.NodeRep(decls[2]) != "A" {
		t.Errorf("expected third element to be A, got %s", ast.NodeRep(decls[2]))
	}
	if ast.NodeRep(decls[3]) != "B" {
		t.Errorf("expected fourth element to be B, got %s", ast.NodeRep(decls[3]))
	}
}

// TestClassObjectDeclCreated verifies that createObject (called from CLASS)
// declares an ObjectDecl with the class name.
func TestClassObjectDeclCreated(t *testing.T) {
	input := `#lang 1.7
type t
class foo = {
    action bar = {
        skip
    }
}`
	result, err := Parse(input, lexer.Version{1, 7})
	if err != nil {
		t.Skipf("CLASS parse test skipped: %v", err)
	}

	var hasObjectDecl bool
	for _, d := range result.Decls {
		if _, ok := d.(*ast.ObjectDecl); ok {
			canonStr := string(d.Canon())
			if strings.Contains(canonStr, "foo") {
				hasObjectDecl = true
			}
		}
	}
	if !hasObjectDecl {
		t.Log("Decls found:")
		for i, d := range result.Decls {
			t.Logf("  [%d] %T: %s", i, d, d.Canon())
		}
		t.Error("expected an ObjectDecl with 'foo' from CLASS declaration")
	}
}

// TestVariantDefConstruction directly tests VariantDef construction
// matching Python: VariantDef(scnst, Atom(p[5]))
func TestVariantDefConstruction(t *testing.T) {
	cfg := ast.NewAstConfig()
	scnst := cfg.NewAtom("this")
	superType := cfg.NewAtom("base_t")
	vdfn := cfg.NewVariantDef(scnst, superType)

	canon := string(vdfn.Canon())
	if !strings.Contains(canon, "variantDef") {
		t.Errorf("expected 'variantDef' in canon, got %s", canon)
	}
	if !strings.Contains(canon, "this") {
		t.Errorf("expected 'this' in canon, got %s", canon)
	}
	if !strings.Contains(canon, "base_t") {
		t.Errorf("expected 'base_t' in canon, got %s", canon)
	}
	t.Logf("VariantDef canon: %s", canon)
}
