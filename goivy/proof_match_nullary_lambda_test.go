package goivy

import "testing"

// Python match (ivy_proof.py): a pattern p whose symbol is free matches any
// instance via extract_terms, which for a nullary p builds
// il.Lambda([], inst) (Lambda accepts no variables). goivy's ExtractTerms
// returned nil for no terms, so "apply elimAndR with p = p & q"
// (test/ded2.ivy) failed with "Match is inconsistent".
func TestMatchNullaryFreeSymbolViaLambda(t *testing.T) {
	p, q := NewConst("p", Boolean), NewConst("q", Boolean)
	free := NewInsMap[NodeKey, Expr]()
	free.Set(Key(p), p)
	inst := &LogicAnd{Terms: []Expr{p, q}}
	m := Match(p, inst, free, nil)
	if m == nil {
		t.Fatal("p should match p & q via a nullary lambda")
	}
	lam, ok := m[Key(p)].(*Lambda)
	if !ok || len(lam.Variables) != 0 || Key(lam.Body) != Key(inst) {
		t.Fatalf("match[p] = %#v, want Lambda([], p & q)", m[Key(p)])
	}
}
