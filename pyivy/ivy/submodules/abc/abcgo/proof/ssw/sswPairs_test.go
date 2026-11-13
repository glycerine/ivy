package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestSignalCorrespondenceWithPairsRunsOnSequentialMiter(t *testing.T) {
	makeGraph := func() (*aig.Graph, aig.Edge) {
		g := aig.NewGraph()
		pi, lo := g.NewInput(), g.NewInput()
		gate := g.AND(pi, lo)
		g.AddOutput(gate)
		g.AddOutput(pi)
		aig.Aig_ManSetRegNum(g, 1)
		return g, gate
	}
	a, x := makeGraph()
	b, y := makeGraph()
	ids1, ids2 := vec.Vec_IntAlloc(1), vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(ids1, int32(x.NodeID()))
	vec.Vec_IntPush(ids2, int32(y.NodeID()))
	var pars Ssw_Pars_t
	Ssw_ManSetDefaultParams(&pars)
	pars.nStepsMax = 1
	out := Ssw_SignalCorrespondenceWithPairs(a, b, ids1, ids2, &pars)
	if out == nil {
		t.Fatal("nil pair-guided result")
	}
	if got := Ssw_MiterStatus(out, 0); got != 1 {
		t.Fatalf("identical sequential graphs not proved: %d", got)
	}
}
