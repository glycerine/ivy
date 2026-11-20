package goivy

import (
	"strings"
	"testing"
)

// Python skolemize_fmla gives a skolem function the outer universals in
// iu.unique(lu.variables_ast(fmla)) order. goivy iterated a Go map, so the
// argument order varied from run to run (test/fba.ivy).
func TestSkolemArgumentsFollowTraversalOrder(t *testing.T) {
	s := &UninterpretedSort{Name: "s"}
	q1 := &LogicVariable{Name: "Q1", VSort: s}
	q2 := &LogicVariable{Name: "Q2", VSort: s}
	x := &LogicVariable{Name: "X", VSort: s}
	p := NewConst("p", LogicRelationSort([]Sort{s, s, s}))
	body, _ := NewApply(p, q2, q1, x)
	// exists Q1 Q2. forall X. p(Q2,Q1,X), in positive position:
	// Q1, Q2 are universalized (outer) and X is skolemized.
	fmla := IvyExists([]*LogicVariable{q1, q2}, &ForAll{Variables: []*LogicVariable{x}, Body: body})
	for i := 0; i < 20; i++ {
		var skfuns []*Const
		res := SkolemizeFmla(fmla, true, NewUniqueRenamer("", nil), &skfuns, true)
		got := res.(Expr).String()
		if !strings.Contains(got, "_X(Q2,Q1)") {
			t.Fatalf("run %d: %s, want _X(Q2,Q1)", i, got)
		}
	}
}
