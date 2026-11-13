package pdr

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
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

func TestAbstractCounterexampleRefinesUsedFlop(t *testing.T) {
	g := lifecycleGraph()
	p := Pdr_ManStart(g, &Pdr_Par_t{fUseAbs: 1, fUseSimpleRef: 1}, nil)
	p.vMapPpi2Ff = vec.Vec_IntAlloc(1)
	p.vMapFf2Ppi = vec.Vec_IntStart(1)
	vec.Vec_IntWriteEntry(p.vMapFf2Ppi, 0, -1)
	p.vAbsFlops = vec.Vec_IntStart(1)
	Pdr_QueuePush(p, Pdr_OblStart(0, 0, testPdrSet(nil, []int32{2}), nil))
	if Pdr_ManDeriveCexAbs(p) != nil || vec.Vec_IntEntry(p.vAbsFlops, 0) != 1 || p.nCexes != 1 {
		t.Fatal("abstract input did not refine its flop")
	}
	Pdr_ManStop(p)
}

func TestAbstractCounterexampleUsesCareMinimization(t *testing.T) {
	g := aig.NewGraph()
	_ = g.NewInput()
	latch := g.NewInput()
	g.AddOutput(latch)
	g.AddOutput(latch)
	aig.Aig_ManSetRegNum(g, 1)
	p := Pdr_ManStart(g, &Pdr_Par_t{fUseAbs: 1}, nil)
	p.vMapPpi2Ff = vec.Vec_IntAlloc(1)
	p.vMapFf2Ppi = vec.Vec_IntStart(1)
	vec.Vec_IntWriteEntry(p.vMapFf2Ppi, 0, -1)
	p.vAbsFlops = vec.Vec_IntStart(1)
	Pdr_QueuePush(p, Pdr_OblStart(0, 0, testPdrSet(nil, []int32{2}), nil))
	if Pdr_ManDeriveCexAbs(p) != nil || vec.Vec_IntEntry(p.vAbsFlops, 0) != 1 {
		t.Fatal("care minimization did not refine the necessary flop")
	}
	Pdr_ManStop(p)
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
