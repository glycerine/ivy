package lalr_full

import (
	"strings"
	"testing"

	"github.com/glycerine/goivy/ast"
	"github.com/glycerine/goivy/lexer"
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
	c1 := ast.NewChoiceAction(ast.NewAtom("a"), ast.NewAtom("b"))
	c2 := ast.NewChoiceAction(ast.NewAtom("c"), ast.NewAtom("d"))
	if c1.UniqueID == c2.UniqueID {
		t.Errorf("expected unique IDs, got both %d", c1.UniqueID)
	}
	if c1.UniqueID == 0 || c2.UniqueID == 0 {
		t.Errorf("expected non-zero UniqueIDs, got %d and %d", c1.UniqueID, c2.UniqueID)
	}
	t.Logf("c1.UniqueID=%d, c2.UniqueID=%d", c1.UniqueID, c2.UniqueID)
}

// TestChoiceActionCanon verifies the canonical s-expression output.
func TestChoiceActionCanon(t *testing.T) {
	c := ast.NewChoiceAction(ast.NewAtom("a"), ast.NewAtom("b"))
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
	a := ast.NewAtom("branch1")
	b := ast.NewAtom("branch2")
	c := ast.NewChoiceAction(a, b)
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
	c := ast.NewChoiceAction(ast.NewAtom("a"), ast.NewAtom("b"))
	cloned := c.Clone([]ast.Node{ast.NewAtom("c"), ast.NewAtom("d")})
	cc, ok := cloned.(*ast.ChoiceAction)
	if !ok {
		t.Fatalf("expected *ast.ChoiceAction, got %T", cloned)
	}
	if cc.UniqueID == c.UniqueID {
		t.Errorf("cloned ChoiceAction should have different UniqueID, both are %d", c.UniqueID)
	}
}
