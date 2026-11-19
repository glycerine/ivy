package goivy

import "testing"

// Python ASTContext.__exit__ (ivy_ast.py), which thing() wraps around the
// compilation of every node: a logic-level error (SortError) becomes
// IvyError(node, str(err)); an IvyError without a line number gets the
// node's lineno, if the node has one; other errors pass through. goivy had
// no ASTContext, so "unknown symbol: pout" (test/ite1.ivy) lost its line.
func TestASTContextErrorLikePython(t *testing.T) {
	cfg := NewAstConfig()
	node := cfg.NewAtom("pout")
	node.SetLineno(Location{Filename: "f.ivy", Line: 7})

	e1 := &IvyError{Msg: "unknown symbol: pout"}
	got := astContextError(node, e1)
	if got.Error() != "f.ivy: line 7: error: unknown symbol: pout" {
		t.Fatalf("unlocated IvyError: %q", got.Error())
	}

	e2 := &IvyError{Msg: "m", Loc: Location{Filename: "g.ivy", Line: 3}, HasLoc: true}
	if got := astContextError(node, e2); got.Error() != "g.ivy: line 3: error: m" {
		t.Fatalf("located IvyError must keep its line: %q", got.Error())
	}

	if got := astContextError(node, &SortError{Msg: "bad sort"}); got.Error() != "f.ivy: line 7: error: bad sort" {
		t.Fatalf("SortError: %q", got.Error())
	}
	if _, ok := astContextError(node, &SortError{Msg: "bad sort"}).(*IvyError); !ok {
		t.Fatal("SortError must become an IvyError")
	}

	bare := cfg.NewAtom("x") // no lineno
	if got := astContextError(bare, &IvyError{Msg: "m"}); got.Error() != "m" {
		t.Fatalf("node without lineno must leave the error unlocated: %q", got.Error())
	}
	if astContextError(node, nil) != nil {
		t.Fatal("nil stays nil")
	}
}
