package pdr

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/util"
)

func lifecycleGraph() *aig.Graph {
	g := aig.NewGraph()
	input := g.NewInput()
	latch := g.NewInput()
	g.AddOutput(input)
	g.AddOutput(latch)
	aig.Aig_ManSetRegNum(g, 1)
	return g
}

func TestManagerFanoutOwnership(t *testing.T) {
	for _, existing := range []bool{false, true} {
		g := lifecycleGraph()
		if existing {
			aig.Aig_ManFanoutStart(g)
		}
		p := Pdr_ManStart(g, &Pdr_Par_t{}, nil)
		if !aig.Aig_ManHasFanout(g) || p.ownsFanout == existing {
			t.Fatalf("existing=%v: incorrect fanout ownership", existing)
		}
		Pdr_ManStop(p)
		if aig.Aig_ManHasFanout(g) != existing {
			t.Fatalf("existing=%v: manager changed caller's fanout ownership", existing)
		}
		if existing {
			aig.Aig_ManFanoutStop(g)
		}
	}
}

func TestManagerDerivesConcreteCounterexample(t *testing.T) {
	g := lifecycleGraph()
	p := Pdr_ManStart(g, &Pdr_Par_t{}, nil)
	state := testPdrSet(nil, []int32{0})
	entry := Pdr_OblStart(0, 0, state, nil)
	Pdr_QueuePush(p, entry)
	cex := Pdr_ManDeriveCex(p)
	if cex.IFrame != 0 || cex.IPo != 0 || !util.Abc_InfoHasBit(cex.PData, cex.NRegs) || saig.Saig_ManVerifyCex(g, cex) != 1 {
		t.Fatalf("invalid counterexample: %+v", cex)
	}
	Pdr_ManStop(p)
}
