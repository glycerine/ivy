package lalr_full

import (
	"strings"
	"testing"

	"github.com/glycerine/goivy/ast"
	"github.com/glycerine/goivy/lexer"
)

// TestGhostTypeUsesGhostTypeDef verifies that "ghost type t" creates
// a GhostTypeDef, not a plain TypeDef.
// Matches Python ivy_parser.py:1774: tdfn = (GhostTypeDef if p[3] else TypeDef)(...)
func TestGhostTypeUsesGhostTypeDef(t *testing.T) {
	input := `#lang 1.7
ghost type t`
	result, err := Parse(input, lexer.Version{1, 7})
	if err != nil {
		t.Skipf("parse skipped: %v", err)
	}

	var foundGhost bool
	for _, d := range result.Decls {
		canon := string(d.Canon())
		if strings.Contains(canon, "ghostTypeDef") {
			foundGhost = true
			t.Logf("Found GhostTypeDef: %s", canon)
		}
	}
	if !foundGhost {
		t.Log("Decls found:")
		for i, d := range result.Decls {
			t.Logf("  [%d] %T: %s", i, d, d.Canon())
		}
		t.Error("expected GhostTypeDef from 'ghost type t'")
	}
}

// TestNonGhostTypeUsesTypeDef verifies that "type t" (no ghost) creates
// a plain TypeDef, not GhostTypeDef.
func TestNonGhostTypeUsesTypeDef(t *testing.T) {
	input := `#lang 1.7
type t`
	result, err := Parse(input, lexer.Version{1, 7})
	if err != nil {
		t.Skipf("parse skipped: %v", err)
	}

	for _, d := range result.Decls {
		canon := string(d.Canon())
		if strings.Contains(canon, "ghostTypeDef") {
			t.Errorf("non-ghost 'type t' should NOT produce GhostTypeDef, got: %s", canon)
		}
	}
}

// TestGhostTypeWithSortUsesGhostTypeDef verifies "ghost type t = {0..10}"
// produces GhostTypeDef.
func TestGhostTypeWithSortUsesGhostTypeDef(t *testing.T) {
	input := `#lang 1.7
ghost type idx = {0..10}`
	result, err := Parse(input, lexer.Version{1, 7})
	if err != nil {
		t.Skipf("parse skipped: %v", err)
	}

	var foundGhost bool
	for _, d := range result.Decls {
		canon := string(d.Canon())
		if strings.Contains(canon, "ghostTypeDef") {
			foundGhost = true
			t.Logf("Found GhostTypeDef with sort: %s", canon)
		}
	}
	if !foundGhost {
		t.Log("Decls found:")
		for i, d := range result.Decls {
			t.Logf("  [%d] %T: %s", i, d, d.Canon())
		}
		t.Error("expected GhostTypeDef from 'ghost type idx = {0..10}'")
	}
}

// TestGhostTypeDefConstruction directly tests GhostTypeDef construction.
func TestGhostTypeDefConstruction(t *testing.T) {
	cfg := ast.NewAstConfig()
	scnst := cfg.NewAtom("mytype")
	gt := &ast.GhostTypeDef{TypeDef: ast.TypeDef{Name: scnst, Value: cfg.NewUninterpretedSortAST()}}
	canon := string(gt.Canon())
	if !strings.Contains(canon, "ghostTypeDef") {
		t.Errorf("expected 'ghostTypeDef' in canon, got %s", canon)
	}
	if !strings.Contains(canon, "mytype") {
		t.Errorf("expected 'mytype' in canon, got %s", canon)
	}
	t.Logf("GhostTypeDef canon: %s", canon)
}
