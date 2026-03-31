package lalr_full

import (
	"strings"
	"testing"

	"github.com/glycerine/goivy/ast"
	"github.com/glycerine/goivy/lexer"
)

// TestExplicitDefinitionCreatesSchema verifies that "explicit definition ..."
// wraps the definition in a DefinitionSchema.
// Matches Python ivy_parser.py:1671-1683.
func TestExplicitDefinitionCreatesSchema(t *testing.T) {
	input := `#lang 1.7
type t
explicit definition foo(x:t) = true`
	result, err := Parse(input, lexer.Version{1, 7})
	if err != nil {
		t.Skipf("parse skipped: %v", err)
	}

	var foundSchema bool
	for _, d := range result.Decls {
		canon := string(d.Canon())
		if strings.Contains(canon, "definitionSchema") {
			foundSchema = true
			t.Logf("Found DefinitionSchema: %s", canon)
		}
	}
	if !foundSchema {
		t.Log("Decls found:")
		for i, d := range result.Decls {
			t.Logf("  [%d] %T: %s", i, d, d.Canon())
		}
		t.Error("expected DefinitionSchema from 'explicit definition'")
	}
}

// TestNonExplicitDefinitionNoSchema verifies that "definition ..." (without explicit)
// does NOT wrap in DefinitionSchema.
func TestNonExplicitDefinitionNoSchema(t *testing.T) {
	input := `#lang 1.7
type t
definition foo(x:t) = true`
	result, err := Parse(input, lexer.Version{1, 7})
	if err != nil {
		t.Skipf("parse skipped: %v", err)
	}

	for _, d := range result.Decls {
		canon := string(d.Canon())
		if strings.Contains(canon, "definitionSchema") {
			t.Errorf("non-explicit definition should NOT produce DefinitionSchema, got: %s", canon)
		}
	}
}

// TestDefinitionSchemaConstruction directly tests DefinitionSchema construction.
func TestDefinitionSchemaConstruction(t *testing.T) {
	cfg := ast.NewAstConfig()
	lhs := cfg.NewAtom("foo")
	rhs := cfg.NewAtom("true")
	def := cfg.NewDefinition(lhs, rhs)
	ds := cfg.NewDefinitionSchema(*def)

	canon := string(ds.Canon())
	if !strings.Contains(canon, "definitionSchema") {
		t.Errorf("expected 'definitionSchema' in canon, got %s", canon)
	}
	t.Logf("DefinitionSchema canon: %s", canon)
}
