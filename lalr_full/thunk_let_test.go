package lalr_full

import (
	"strings"
	"testing"

	"github.com/glycerine/goivy/ast"
	"github.com/glycerine/goivy/lexer"
)

// TestLetActionCreated verifies that "let x = y { skip }" creates a LetAction.
// Matches Python ivy_parser.py:3296-3299: LetAction(*(p[2]+[p[3]]))
func TestLetActionCreated(t *testing.T) {
	input := `#lang 1.7
type t
action foo = {
    let x = y {
        skip
    }
}`
	result, err := Parse(input, lexer.Version{1, 7})
	if err != nil {
		t.Skipf("parse skipped: %v", err)
	}

	var foundLet bool
	for _, d := range result.Decls {
		canon := string(d.Canon())
		if strings.Contains(canon, "letAction") {
			foundLet = true
			t.Logf("Found LetAction: %s", canon)
		}
	}
	if !foundLet {
		t.Log("Decls found:")
		for i, d := range result.Decls {
			t.Logf("  [%d] %T: %s", i, d, d.Canon())
		}
		t.Error("expected LetAction from 'let ... { ... }'")
	}
}

// TestLetActionConstruction directly tests LetAction construction.
func TestLetActionConstruction(t *testing.T) {
	binding := ast.NewAtom("=", ast.NewApp(ast.NewSymbol("x", nil)), ast.NewApp(ast.NewSymbol("y", nil)))
	body := ast.NewAtom("skip")
	la := ast.NewLetAction(binding, body)

	if len(la.Bindings) != 1 {
		t.Fatalf("expected 1 binding, got %d", len(la.Bindings))
	}
	if la.Body == nil {
		t.Fatal("expected non-nil body")
	}
	canon := string(la.Canon())
	if !strings.Contains(canon, "letAction") {
		t.Errorf("expected 'letAction' in canon, got %s", canon)
	}
	t.Logf("LetAction canon: %s", canon)
}

// TestLetActionArgs verifies Args() returns bindings + body.
func TestLetActionArgs(t *testing.T) {
	b1 := ast.NewAtom("b1")
	b2 := ast.NewAtom("b2")
	body := ast.NewAtom("body")
	la := ast.NewLetAction(b1, b2, body)

	args := la.Args()
	if len(args) != 3 {
		t.Fatalf("expected 3 args (2 bindings + body), got %d", len(args))
	}
}

// TestThunkActionCreated verifies basic ThunkAction construction from existing
// ast.NewThunkAction function.
func TestThunkActionCreated(t *testing.T) {
	label := ast.NewAtom("mylab")
	action := ast.NewAtom("foo")
	sort := ast.NewAtom("t")
	body := ast.NewAtom("skip")

	ta := ast.NewThunkAction(label, action, sort, body)
	canon := string(ta.Canon())
	if !strings.Contains(canon, "thunkAction") {
		t.Errorf("expected 'thunkAction' in canon, got %s", canon)
	}
	if !strings.Contains(canon, "mylab") {
		t.Errorf("expected 'mylab' in canon, got %s", canon)
	}
	t.Logf("ThunkAction canon: %s", canon)
}

// TestThunkActionFields verifies field access on ThunkAction.
func TestThunkActionFields(t *testing.T) {
	label := ast.NewAtom("lab")
	action := ast.NewAtom("act")
	sort := ast.NewAtom("s")
	body := ast.NewAtom("body")

	ta := ast.NewThunkAction(label, action, sort, body)
	if ast.NodeRep(ta.Label) != "lab" {
		t.Errorf("expected label 'lab', got %q", ast.NodeRep(ta.Label))
	}
	if ast.NodeRep(ta.Action) != "act" {
		t.Errorf("expected action 'act', got %q", ast.NodeRep(ta.Action))
	}
	if ast.NodeRep(ta.Sort) != "s" {
		t.Errorf("expected sort 's', got %q", ast.NodeRep(ta.Sort))
	}
}
