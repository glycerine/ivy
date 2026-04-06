package parser

import (
	"testing"

	"github.com/glycerine/ivy/goivy/ast"
)

// TestLinenoAddRefWhenReferenceSet verifies that LinenoAddRef wraps the
// location with the current reference lineno.
func TestLinenoAddRefWhenReferenceSet(t *testing.T) {
	cfg := ast.NewAstConfig()
	orig := ast.Location{Filename: "orig.ivy", Line: 10}
	ref := ast.Location{Filename: "inst.ivy", Line: 50}

	cfg.SetReferenceLineno(ref)
	defer cfg.SetReferenceLineno(ast.Location{})

	wrapped := cfg.LinenoAddRef(orig)
	if wrapped.Filename != "inst.ivy" {
		t.Errorf("expected filename inst.ivy, got %s", wrapped.Filename)
	}
	if wrapped.Line != 50 {
		t.Errorf("expected line 50, got %d", wrapped.Line)
	}
	if wrapped.Reference == nil {
		t.Fatal("expected Reference to be set")
	}
	if wrapped.Reference.Filename != "orig.ivy" {
		t.Errorf("expected reference filename orig.ivy, got %s", wrapped.Reference.Filename)
	}
	if wrapped.Reference.Line != 10 {
		t.Errorf("expected reference line 10, got %d", wrapped.Reference.Line)
	}
}

// TestLinenoAddRefWhenNoReference verifies identity when referenceLineno is zero.
func TestLinenoAddRefWhenNoReference(t *testing.T) {
	cfg := ast.NewAstConfig()
	cfg.SetReferenceLineno(ast.Location{})

	orig := ast.Location{Filename: "test.ivy", Line: 42}
	result := cfg.LinenoAddRef(orig)
	if result != orig {
		t.Errorf("expected identity, got %v", result)
	}
}

// TestLinenoAddRefChaining verifies that references can chain (depth 2).
func TestLinenoAddRefChaining(t *testing.T) {
	cfg := ast.NewAstConfig()
	orig := ast.Location{Filename: "orig.ivy", Line: 1}

	// First wrap
	ref1 := ast.Location{Filename: "mod1.ivy", Line: 10}
	cfg.SetReferenceLineno(ref1)
	wrapped1 := cfg.LinenoAddRef(orig)

	// Second wrap
	ref2 := ast.Location{Filename: "mod2.ivy", Line: 20}
	cfg.SetReferenceLineno(ref2)
	wrapped2 := cfg.LinenoAddRef(wrapped1)

	cfg.SetReferenceLineno(ast.Location{})

	if wrapped2.Filename != "mod2.ivy" {
		t.Errorf("outer filename: got %s, want mod2.ivy", wrapped2.Filename)
	}
	if wrapped2.Reference == nil {
		t.Fatal("outer Reference nil")
	}
	if wrapped2.Reference.Filename != "mod1.ivy" {
		t.Errorf("middle filename: got %s, want mod1.ivy", wrapped2.Reference.Filename)
	}
	if wrapped2.Reference.Reference == nil {
		t.Fatal("inner Reference nil")
	}
	if wrapped2.Reference.Reference.Filename != "orig.ivy" {
		t.Errorf("inner filename: got %s, want orig.ivy", wrapped2.Reference.Reference.Filename)
	}
}

// TestLocationStringWithReference verifies that String() delegates to Reference.
func TestLocationStringWithReference(t *testing.T) {
	inner := ast.Location{Filename: "orig.ivy", Line: 5}
	outer := ast.Location{Filename: "inst.ivy", Line: 50, Reference: &inner}

	// Should delegate to inner (Reference)
	got := outer.String()
	want := "orig.ivy: line 5: "
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// TestCopyAttributesAstRefUsesLinenoAddRef verifies that CopyAttributesAstRef
// wraps lineno via LinenoAddRef when referenceLineno is set.
func TestCopyAttributesAstRefUsesLinenoAddRef(t *testing.T) {
	cfg := ast.NewAstConfig()
	src := cfg.NewAtom("x")
	src.SetLineno(ast.Location{Filename: "src.ivy", Line: 10})

	dst := cfg.NewAtom("y")

	ref := ast.Location{Filename: "inst.ivy", Line: 50}
	cfg.SetReferenceLineno(ref)
	defer cfg.SetReferenceLineno(ast.Location{})

	ast.CopyAttributesAstRef(src, dst)

	loc := dst.GetLineno()
	if loc.Filename != "inst.ivy" {
		t.Errorf("expected inst.ivy, got %s", loc.Filename)
	}
	if loc.Reference == nil {
		t.Fatal("expected Reference to be set on dst lineno")
	}
	if loc.Reference.Filename != "src.ivy" {
		t.Errorf("expected reference src.ivy, got %s", loc.Reference.Filename)
	}
}

// TestVariableResortUsesLinenoAddRef verifies that Variable.Resort wraps
// lineno via LinenoAddRef when referenceLineno is set.
func TestVariableResortUsesLinenoAddRef(t *testing.T) {
	cfg := ast.NewAstConfig()
	v := cfg.NewVariable("X", "t")
	v.SetLineno(ast.Location{Filename: "var.ivy", Line: 3})

	ref := ast.Location{Filename: "inst.ivy", Line: 50}
	cfg.SetReferenceLineno(ref)
	defer cfg.SetReferenceLineno(ast.Location{})

	resorted := v.Resort("u")
	loc := resorted.GetLineno()
	if loc.Reference == nil {
		t.Fatal("expected Reference to be set on resorted variable")
	}
	if loc.Reference.Filename != "var.ivy" {
		t.Errorf("expected reference var.ivy, got %s", loc.Reference.Filename)
	}
}

// TestWhenOperatorCloneUsesLinenoAddRef verifies that WhenOperator.Clone
// wraps lineno via LinenoAddRef when referenceLineno is set.
func TestWhenOperatorCloneUsesLinenoAddRef(t *testing.T) {
	cfg := ast.NewAstConfig()
	w := cfg.NewWhenOperator("when", cfg.NewAtom("a"), cfg.NewAtom("b"))
	w.SetLineno(ast.Location{Filename: "when.ivy", Line: 7})

	ref := ast.Location{Filename: "inst.ivy", Line: 50}
	cfg.SetReferenceLineno(ref)
	defer cfg.SetReferenceLineno(ast.Location{})

	cloned := w.Clone([]ast.Node{cfg.NewAtom("a"), cfg.NewAtom("b")})
	loc := cloned.GetLineno()
	if loc.Reference == nil {
		t.Fatal("expected Reference to be set on cloned WhenOperator")
	}
	if loc.Reference.Filename != "when.ivy" {
		t.Errorf("expected reference when.ivy, got %s", loc.Reference.Filename)
	}
}

// TestAstRewriteDefaultCaseUsesLinenoAddRef verifies that the default case
// in AstRewrite applies LinenoAddRef to cloned nodes.
func TestAstRewriteDefaultCaseUsesLinenoAddRef(t *testing.T) {
	cfg := ast.NewAstConfig()
	// Use a Sequence node which hits the default case in AstRewrite
	seq := cfg.NewSequence(cfg.NewAtom("a"), cfg.NewAtom("b"))
	seq.SetLineno(ast.Location{Filename: "seq.ivy", Line: 15})

	ref := ast.Location{Filename: "inst.ivy", Line: 50}
	cfg.SetReferenceLineno(ref)
	defer cfg.SetReferenceLineno(ast.Location{})

	// Use a no-op rewriter just to trigger the default case
	rw := ast.NewAstRewriteSubstPrefix(nil, nil)
	result := ast.AstRewrite(seq, rw)

	loc := result.GetLineno()
	if loc.Reference == nil {
		t.Fatal("expected Reference to be set on rewritten Sequence")
	}
	if loc.Reference.Filename != "seq.ivy" {
		t.Errorf("expected reference seq.ivy, got %s", loc.Reference.Filename)
	}
}

// TestIvyAccumImplementsNode is a compile-time check that ivyAccum
// implements ast.Node.
func TestIvyAccumImplementsNode(t *testing.T) {
	var _ ast.Node = (*ivyAccum)(nil)
}

// TestIvyAccumArgsReturnsNil verifies Args() returns nil, matching Python Ivy.args → [].
func TestIvyAccumArgsReturnsNil(t *testing.T) {
	m := newIvyAccum(nil, "")
	if m.Args() != nil {
		t.Error("expected nil Args")
	}
}

// TestIvyAccumCloneReturnsSelf verifies Clone returns self, matching Python Ivy.clone.
func TestIvyAccumCloneReturnsSelf(t *testing.T) {
	m := newIvyAccum(nil, "")
	cloned := m.Clone(nil)
	if cloned != m {
		t.Error("expected Clone to return self")
	}
}
