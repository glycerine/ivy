package framgr

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestOneHotClauseSourceEmptySimulationStillChecksPolarity(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	s := Fra_SmlStart(g, 0, 1, 0)
	defer func() {
		if recover() == nil {
			t.Error("source asserts for two positive literals even when there are no words")
		}
	}()
	Fra_OneHotNodesAreClause(s, a, b, 0, 0)
}

func TestOneHotRefinementUsesLatchInputsAfterPrimaryInputs(t *testing.T) {
	g := aig.NewGraph()
	pi0, pi1 := g.NewInput(), g.NewInput()
	r0, r1 := g.NewInput(), g.NewInput()
	g.AddOutput(r0)
	g.AddOutput(r1)
	aig.Aig_ManSetRegNum(g, 2)
	s := Fra_SmlStart(g, 0, 1, 1)
	s.NodeWords(pi0.NodeID())[0], s.NodeWords(pi1.NodeID())[0] = 0, 0
	s.NodeWords(r0.NodeID())[0], s.NodeWords(r1.NodeID())[0] = 1, 1
	p := &Fra_Man_t{PManAig: g, PSml: s}
	clauses := vec.Vec_IntAlloc(4)
	vec.Vec_IntPushTwo(clauses, 0, 0) // already removed
	vec.Vec_IntPushTwo(clauses, Fra_RegToLit(0, 1), Fra_RegToLit(1, 1))
	if got := Fra_OneHotRefineUsingCex(p, clauses); got != 1 || Fra_OneHotCount(p, clauses) != 0 {
		t.Fatalf("latch violation was not removed: changed=%d clauses=%v", got, vec.Vec_IntArray(clauses))
	}
}

func TestOneHotClausesUseNonPrefixWords(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	s := Fra_SmlStart(g, 1, 1, 1)
	s.NodeWords(a.NodeID())[0], s.NodeWords(b.NodeID())[0] = ^uint32(0), ^uint32(0)
	tests := []struct {
		left, right      uint32
		negA, negB, want int32
	}{
		{3, 4, 1, 1, 1}, {3, 1, 1, 1, 0}, {3, 7, 1, 0, 1}, {3, 1, 1, 0, 0}, {7, 3, 0, 1, 1}, {1, 3, 0, 1, 0},
	}
	for _, tc := range tests {
		s.NodeWords(a.NodeID())[1] = tc.left
		s.NodeWords(b.NodeID())[1] = tc.right
		if got := Fra_OneHotNodesAreClause(s, a, b, tc.negA, tc.negB); got != tc.want {
			t.Errorf("%+v: got %d", tc, got)
		}
	}
}
