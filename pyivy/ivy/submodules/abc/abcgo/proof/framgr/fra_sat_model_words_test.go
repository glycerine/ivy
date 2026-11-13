package framgr

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/bsat"
)

// The SAT query path must pack the solver model in CI order across the same
// 32-bit words used by Fra_SmlSavePattern.
func TestSATQuerySavesModelAcrossWords(t *testing.T) {
	g := aig.NewGraph()
	inputs := make([]aig.Edge, 40)
	for i := range inputs {
		inputs[i] = g.NewInput()
	}
	solver := bsat.NewSolver()
	vars := make([]int32, len(inputs))
	for i := range vars {
		vars[i] = solver.NewVariable()
	}
	p := &Fra_Man_t{
		PManAig:     g,
		PManFraig:   g,
		PSat:        solver,
		PMemSatNums: make([]int32, len(g.Nodes())),
		PPatWords:   make([]uint32, 2),
		NPatWords:   2,
		PPars:       &Fra_Par_t{},
	}
	assumptions := make([]bsat.Literal, len(vars))
	var want [2]uint32
	for i, input := range inputs {
		Fra_ObjSetSatNum(p, input, vars[i])
		value := i%3 == 0
		if value {
			assumptions[i] = bsat.Positive(vars[i])
			want[i/32] |= uint32(1) << uint(i%32)
		} else {
			assumptions[i] = bsat.Negative(vars[i])
		}
	}
	if got := fraSATQuery(p, aig.InvalidEdge, aig.InvalidEdge, 100, assumptions...); got != 0 {
		t.Fatalf("SAT query result = %d, want counterexample result 0", got)
	}
	if p.PPatWords[0] != want[0] || p.PPatWords[1] != want[1] {
		t.Fatalf("saved model = %#v, want %#v", p.PPatWords, want)
	}
}
