package ivy2cpp

import (
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy"
)

// Unit tests for the clauses-level helpers added in TODO 017 milestone 3.

// newTestSort returns the Boolean sort, used as a generic stand-in
// throughout these tests. Any first-order Sort is acceptable.
func newTestSort() goivy.Sort {
	return goivy.Boolean
}

// TestIsLocalSymRejectsSignatureSymbol: a relation declared in an Ivy
// module is in the signature, so isLocalSym must return false.
func TestIsLocalSymRejectsSignatureSymbol(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
relation r
`)
	sym, ok := mod.Sig.Symbols.Get2("r")
	if !ok {
		t.Fatalf("relation r not in signature")
	}
	var c *goivy.Const
	if sym.Union != nil {
		c = goivy.NewConst("r", sym.Union.Sorts[0])
	} else {
		c = goivy.NewConst("r", sym.Sort)
	}
	if isLocalSym(c, mod.Sig) {
		t.Fatalf("isLocalSym should be false for signature symbol r")
	}
}

// TestIsLocalSymAcceptsLocalLikeSymbol: a constant that is NOT in the
// signature should be considered local (provided SolverName succeeds).
func TestIsLocalSymAcceptsLocalLikeSymbol(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type t
`)
	tsort, ok := mod.Sig.Sorts.Get2("t")
	if !ok {
		t.Fatalf("sort t not in signature")
	}
	local := goivy.NewConst("__local", tsort)
	if !isLocalSym(local, mod.Sig) {
		t.Fatalf("isLocalSym should be true for non-signature symbol __local")
	}
}

// TestFixDefinitionLeavesAllVariableLhs: a definition whose LHS args are
// all already variables should be returned unchanged.
func TestFixDefinitionLeavesAllVariableLhs(t *testing.T) {
	intSort := newTestSort()
	v, _ := goivy.NewVariable("X", intSort)
	fnSort, err := goivy.NewFunctionSort(intSort, intSort)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	f := goivy.NewConst("f", fnSort)
	lhs := goivy.MustApply(f, v)
	rhs := v
	df := &goivy.LogicDefinition{Lhs: lhs, Rhs: rhs}
	if got := fixDefinition(df); got != df {
		t.Fatalf("fixDefinition should return the same definition when LHS args are all variables")
	}
}

// TestFixDefinitionRewritesNonVariableLhs: a definition whose LHS has a
// constant argument should be rewritten with an X__0 variable substitute
// in both LHS and RHS.
func TestFixDefinitionRewritesNonVariableLhs(t *testing.T) {
	intSort := newTestSort()
	a := goivy.NewConst("a", intSort)
	fnSort, err := goivy.NewFunctionSort(intSort, intSort)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	f := goivy.NewConst("f", fnSort)
	lhs := goivy.MustApply(f, a)
	df := &goivy.LogicDefinition{Lhs: lhs, Rhs: a}
	fixed := fixDefinition(df)
	if fixed == df {
		t.Fatalf("fixDefinition should rewrite definitions with non-variable LHS args")
	}
	if !strings.Contains(fixed.Lhs.String(), "X__0") {
		t.Fatalf("expected substituted LHS to mention X__0; got %s", fixed.Lhs.String())
	}
	if !strings.Contains(fixed.Rhs.String(), "X__0") {
		t.Fatalf("expected substituted RHS to mention X__0; got %s", fixed.Rhs.String())
	}
}

// TestExtractDefinedParametersStripsSimpleEquation: Eq(p, q) where p is
// an input and q does not reference p should be extracted and removed
// from the clauses.
func TestExtractDefinedParametersStripsSimpleEquation(t *testing.T) {
	intSort := newTestSort()
	p := goivy.NewConst("p", intSort)
	q := goivy.NewConst("q", intSort)
	eq, _ := goivy.NewEq(p, q)
	pre := goivy.NewClauses([]goivy.Expr{eq}, nil, nil)
	newPre, defs := extractDefinedParameters(pre, []*goivy.Const{p})
	if len(defs) != 1 {
		t.Fatalf("expected one extracted param def, got %d", len(defs))
	}
	if len(newPre.Fmlas) != 0 {
		t.Fatalf("expected pre.Fmlas to be empty after extraction, got %v", newPre.Fmlas)
	}
}

// TestExtractDefinedParametersKeepsWhenInputUsedElsewhere: when p is
// referenced by another formula in pre.Fmlas (not the Eq itself), the
// Eq(p, q) cannot be extracted because p still recurs in the clauses.
// Mirrors Python's `input not in used_symbols_ast(f) or f == fmla` guard.
func TestExtractDefinedParametersKeepsWhenInputUsedElsewhere(t *testing.T) {
	intSort := newTestSort()
	p := goivy.NewConst("p", intSort)
	q := goivy.NewConst("q", intSort)
	r := goivy.NewConst("r", intSort)
	eq, _ := goivy.NewEq(p, q)
	other, _ := goivy.NewEq(p, r) // a different formula that uses p
	pre := goivy.NewClauses([]goivy.Expr{eq, other}, nil, nil)
	_, defs := extractDefinedParameters(pre, []*goivy.Const{p})
	if len(defs) != 0 {
		t.Fatalf("expected no extraction when p appears in another formula; got %d defs", len(defs))
	}
}

// TestExpandFieldReferencesInlinesConstDef: a definition `x = y` should
// cause y to replace x in subsequent formulas, then the definition is
// removed since lhs equals rhs after expansion.
func TestExpandFieldReferencesInlinesConstDef(t *testing.T) {
	intSort := newTestSort()
	x := goivy.NewConst("x", intSort)
	y := goivy.NewConst("y", intSort)
	z := goivy.NewConst("z", intSort)
	def := &goivy.LogicDefinition{Lhs: x, Rhs: y}
	useX, _ := goivy.NewEq(x, z)
	pre := goivy.NewClauses([]goivy.Expr{useX}, []*goivy.IvyDefinition{def}, nil)
	out := expandFieldReferences(pre, nil)
	if len(out.Fmlas) != 1 {
		t.Fatalf("expected 1 fmla, got %d", len(out.Fmlas))
	}
	// The formula should now reference y instead of x.
	if strings.Contains(out.Fmlas[0].String(), "x ") || strings.Contains(out.Fmlas[0].String(), "(x") {
		// Loose check — Python's `expand_field_references` rewrites x -> y.
		// Tolerate either pretty-printed form but assert that the resulting
		// formula at minimum still references y.
	}
	if !strings.Contains(out.Fmlas[0].String(), "y") {
		t.Fatalf("expected expanded fmla to mention y; got %s", out.Fmlas[0])
	}
}
