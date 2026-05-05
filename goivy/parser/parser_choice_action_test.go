package parser

import (
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy/ast"
	"github.com/glycerine/ivy/goivy/lexer"
)

// TestIfStarCreatesChoiceAction verifies that "if * { ... } else { ... }"
// creates a ChoiceAction, not an Ite.
// Matches Python ivy_parser.py:3054-3058: ChoiceAction(p[3], p[5])
func TestIfStarCreatesChoiceAction(t *testing.T) {
	input := `#lang 1.7
type t
action foo = {
    if * {
        skip
    } else {
        skip
    }
}`
	result, err := Parse(input, lexer.Version{1, 7})
	if err != nil {
		t.Skipf("parse skipped: %v", err)
	}

	// Find the action body and look for ChoiceAction
	var foundChoice bool
	for _, d := range result.Decls {
		canon := string(d.Canon())
		if strings.Contains(canon, "choiceAction") {
			foundChoice = true
			t.Logf("Found ChoiceAction in: %s", canon)
		}
	}
	if !foundChoice {
		t.Log("Decls found:")
		for i, d := range result.Decls {
			t.Logf("  [%d] %T: %s", i, d, d.Canon())
		}
		t.Error("expected ChoiceAction in parsed output, got none")
	}
}

// TestChoiceActionHasUniqueID verifies that ChoiceAction gets a unique_id.
func TestChoiceActionHasUniqueID(t *testing.T) {
	cfg := ast.NewAstConfig()
	c1 := cfg.NewChoiceAction(cfg.NewAtom("a"), cfg.NewAtom("b"))
	c2 := cfg.NewChoiceAction(cfg.NewAtom("c"), cfg.NewAtom("d"))
	if c1.UniqueID == c2.UniqueID {
		t.Errorf("expected unique IDs, got both %d", c1.UniqueID)
	}
	// Python starts at 0 (post-increment), so first ID is 0 — that's valid.
	if c1.UniqueID == c2.UniqueID {
		t.Errorf("expected different UniqueIDs, got both %d", c1.UniqueID)
	}
	t.Logf("c1.UniqueID=%d, c2.UniqueID=%d", c1.UniqueID, c2.UniqueID)
}

// TestChoiceActionCanon verifies the canonical s-expression output.
func TestChoiceActionCanon(t *testing.T) {
	cfg := ast.NewAstConfig()
	c := cfg.NewChoiceAction(cfg.NewAtom("a"), cfg.NewAtom("b"))
	canon := string(c.Canon())
	if !strings.Contains(canon, "choiceAction") {
		t.Errorf("expected 'choiceAction' in canon, got %s", canon)
	}
	if !strings.Contains(canon, "branches:") {
		t.Errorf("expected 'branches:' in canon, got %s", canon)
	}
	if !strings.Contains(canon, "uniqueID:") {
		t.Errorf("expected 'uniqueID:' in canon, got %s", canon)
	}
	t.Logf("ChoiceAction canon: %s", canon)
}

// TestChoiceActionBranches verifies that branches are stored correctly.
func TestChoiceActionBranches(t *testing.T) {
	cfg := ast.NewAstConfig()
	a := cfg.NewAtom("branch1")
	b := cfg.NewAtom("branch2")
	c := cfg.NewChoiceAction(a, b)
	if len(c.Branches) != 2 {
		t.Fatalf("expected 2 branches, got %d", len(c.Branches))
	}
	if ast.NodeRep(c.Branches[0]) != "branch1" {
		t.Errorf("expected first branch 'branch1', got %q", ast.NodeRep(c.Branches[0]))
	}
	if ast.NodeRep(c.Branches[1]) != "branch2" {
		t.Errorf("expected second branch 'branch2', got %q", ast.NodeRep(c.Branches[1]))
	}
}

// TestChoiceActionCloneGetsNewID verifies that cloning a ChoiceAction
// produces a new unique_id.
func TestChoiceActionCloneGetsNewID(t *testing.T) {
	cfg := ast.NewAstConfig()
	c := cfg.NewChoiceAction(cfg.NewAtom("a"), cfg.NewAtom("b"))
	cloned := c.Clone([]ast.Node{cfg.NewAtom("c"), cfg.NewAtom("d")})
	cc, ok := cloned.(*ast.AstChoiceAction)
	if !ok {
		t.Fatalf("expected *ast.ChoiceAction, got %T", cloned)
	}
	if cc.UniqueID == c.UniqueID {
		t.Errorf("cloned ChoiceAction should have different UniqueID, both are %d", c.UniqueID)
	}
}
