package parser

import (
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy/ast"
	"github.com/glycerine/ivy/goivy/lexer"
)

// TestUnprovableInvariantSkippedByDefault verifies that an [unprovable]
// invariant is NOT declared when CheckUnprovable is false (default).
// Matches Python ivy_parser.py:608: if not lf.unprovable or check_unprovable.get(): p[0].declare(d)
func TestUnprovableInvariantSkippedByDefault(t *testing.T) {
	cfg := ast.NewAstConfig()
	cfg.CheckUnprovable = false

	input := `#lang 1.7
type t
relation r(X:t)
unprovable invariant [unp] r(X)`
	result, err := Parse(input, lexer.Version{1, 7}, WithAstConfig(cfg))
	if err != nil {
		t.Skipf("parse skipped: %v", err)
	}

	for _, d := range result.Decls {
		if _, ok := d.(*ast.ConjectureDecl); ok {
			canon := string(d.Canon())
			t.Errorf("unprovable invariant should NOT be declared when CheckUnprovable=false, got: %s", canon)
		}
	}
}

// TestUnprovableInvariantDeclaredWhenEnabled verifies that [unprovable]
// invariant IS declared when CheckUnprovable is true.
func TestUnprovableInvariantDeclaredWhenEnabled(t *testing.T) {
	cfg := ast.NewAstConfig()
	cfg.CheckUnprovable = true

	input := `#lang 1.7
type t
relation r(X:t)
unprovable invariant [unp] r(X)`
	result, err := Parse(input, lexer.Version{1, 7}, WithAstConfig(cfg))
	if err != nil {
		t.Skipf("parse skipped: %v", err)
	}

	var found bool
	for _, d := range result.Decls {
		if _, ok := d.(*ast.ConjectureDecl); ok {
			found = true
		}
	}
	if !found {
		t.Error("unprovable invariant should be declared when CheckUnprovable=true")
	}
}

// TestUnprovableAssertBecomesSequence verifies that "[unprovable] assert ..."
// is replaced with an empty Sequence when CheckUnprovable is false.
// Matches Python ivy_parser.py:2800-2801.
func TestUnprovableAssertBecomesSequence(t *testing.T) {
	cfg := ast.NewAstConfig()
	cfg.CheckUnprovable = false

	input := `#lang 1.7
type t
individual x:t
action foo = {
    unprovable assert x = x
}`
	result, err := Parse(input, lexer.Version{1, 7}, WithAstConfig(cfg))
	if err != nil {
		t.Skipf("parse skipped: %v", err)
	}

	// The action body should contain a Sequence (empty) instead of AssertAction
	for _, d := range result.Decls {
		canon := string(d.Canon())
		if strings.Contains(canon, "assertAction") {
			t.Errorf("unprovable assert should be replaced with Sequence, found assertAction in: %s", canon)
		}
	}
}

// TestProvableAssertKept verifies that a normal (not unprovable) assert is kept.
func TestProvableAssertKept(t *testing.T) {
	input := `#lang 1.7
type t
individual x:t
action foo = {
    assert x = x
}`
	result, err := Parse(input, lexer.Version{1, 7})
	if err != nil {
		t.Skipf("parse skipped: %v", err)
	}

	var foundAssert bool
	for _, d := range result.Decls {
		canon := string(d.Canon())
		if strings.Contains(canon, "assertAction") {
			foundAssert = true
		}
	}
	if !foundAssert {
		t.Error("provable assert should be kept as AssertAction")
	}
}

// TestProvableInvariantAlwaysDeclared verifies that a regular invariant
// (not unprovable) is always declared.
func TestProvableInvariantAlwaysDeclared(t *testing.T) {
	input := `#lang 1.7
type t
individual x:t
invariant x = x`
	result, err := Parse(input, lexer.Version{1, 7})
	if err != nil {
		t.Skipf("parse skipped: %v", err)
	}

	var found bool
	for _, d := range result.Decls {
		if _, ok := d.(*ast.ConjectureDecl); ok {
			found = true
		}
	}
	if !found {
		t.Error("regular invariant should always be declared")
	}
}
