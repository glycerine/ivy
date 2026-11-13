package dch

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

// dchSat.c solves old=1/new=0 before old=0/new=1, using the latter only
// after learning the implication from an UNSAT first query.
func TestSATCounterexampleInEachDirection(t *testing.T) {
	for _, flip := range []bool{false, true} {
		for _, reverse := range []bool{false, true} {
			g := aig.NewGraph()
			a, b := g.NewInput(), g.NewInput()
			and := g.AND(a, b)
			g.AddOutput(and)
			pars := DefaultParams()
			pars.PolarFlip = flip
			m, err := newManager(g, &pars)
			if err != nil {
				t.Fatal(err)
			}
			old, node := a, and
			if reverse {
				old, node = node, old
			}
			result, assignment, err := m.provePair(old, node)
			if err != nil || result != proofDifferent {
				t.Fatalf("flip=%v reverse=%v result=%v error=%v", flip, reverse, result, err)
			}
			if len(assignment) != 2 || !assignment[0] || assignment[1] {
				t.Fatalf("flip=%v reverse=%v model=%v, want [true false]", flip, reverse, assignment)
			}
		}
	}
}
