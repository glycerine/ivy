package dch

import (
	"testing"

	aig "github.com/glycerine/abcgo/aig/aig"
)

// sat_solver_solve receives nBTLimit directly: zero means no conflict limit,
// whereas a positive bound may leave exactly the same proof undecided.
func TestSourceZeroConflictLimitIsUnlimited(t *testing.T) {
	g := aig.NewGraph()
	inputs := make([]aig.Edge, 8)
	for i := range inputs {
		inputs[i] = g.NewInput()
	}
	left := inputs[0]
	for _, input := range inputs[1:] {
		left = g.XOR(left, input)
	}
	var parity func([]aig.Edge) aig.Edge
	parity = func(v []aig.Edge) aig.Edge {
		if len(v) == 1 {
			return v[0]
		}
		mid := len(v) / 2
		return g.XOR(parity(v[:mid]), parity(v[mid:]))
	}
	right := parity(inputs)
	for _, tc := range []struct {
		limit int
		want  proofResult
	}{
		{-1, proofUnknown},
		{1, proofUnknown},
		{0, proofEquivalent},
	} {
		p := DefaultParams()
		p.ConflictLimit = tc.limit
		manager, err := newManager(g, &p)
		if err != nil {
			t.Fatal(err)
		}
		got, _, err := manager.provePair(left, right)
		if err != nil || got != tc.want {
			t.Fatalf("limit=%d result=%v want=%v error=%v", tc.limit, got, tc.want, err)
		}
	}
}

// C's signed limit is checked by the solver at search boundaries. A proof
// reached by propagation can still finish even with a negative budget.
func TestNegativeConflictLimitAllowsImmediateProof(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	zero := g.AND(g.AND(a, b), g.AND(a, b.Not()))
	params := DefaultParams()
	params.ConflictLimit = -1
	m, err := newManager(g, &params)
	if err != nil {
		t.Fatal(err)
	}
	result, _, err := m.provePair(g.Const1(), zero)
	if err != nil || result != proofEquivalent {
		t.Fatalf("result=%v error=%v, want immediate proof", result, err)
	}
}
