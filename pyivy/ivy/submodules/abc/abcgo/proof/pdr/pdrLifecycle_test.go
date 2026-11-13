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

func TestAbstractCounterexampleRejectsInconsistentPriorMap(t *testing.T) {
	g := lifecycleGraph()
	p := Pdr_ManStart(g, &Pdr_Par_t{fUseAbs: 1, fUseSimpleRef: 1}, nil)
	defer Pdr_ManStop(p)
	p.vMapPpi2Ff = vec.Vec_IntAllocArrayCopy([]int32{0}, 1)
	p.vMapFf2Ppi = vec.Vec_IntAllocArrayCopy([]int32{-1}, 1)
	defer func() {
		if recover() == nil {
			t.Fatal("expected the C source's inverse-map assertion")
		}
	}()
	Pdr_ManDeriveCexAbs(p)
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

// pdrMan.c refines only care PPIs, even when the abstract trace also
// assigns another PPI. Here the bad output depends on only the first flop.
func TestAbstractCounterexampleRefinesOnlyCareFlops(t *testing.T) {
	g := aig.NewGraph()
	_ = g.NewInput()
	first, second := g.NewInput(), g.NewInput()
	g.AddOutput(first)
	g.AddOutput(first)
	g.AddOutput(second)
	aig.Aig_ManSetRegNum(g, 2)
	p := Pdr_ManStart(g, &Pdr_Par_t{fUseAbs: 1}, nil)
	defer Pdr_ManStop(p)
	p.vMapPpi2Ff = vec.Vec_IntAlloc(2)
	p.vMapFf2Ppi = vec.Vec_IntStart(2)
	vec.Vec_IntWriteEntry(p.vMapFf2Ppi, 0, -1)
	vec.Vec_IntWriteEntry(p.vMapFf2Ppi, 1, -1)
	p.vAbsFlops = vec.Vec_IntStart(2)
	Pdr_QueuePush(p, Pdr_OblStart(0, 0, testPdrSet(nil, []int32{2, 4}), nil))
	if cex := Pdr_ManDeriveCexAbs(p); cex != nil {
		t.Fatal("abstract counterexample should require refinement")
	}
	if got := vec.Vec_IntArray(p.vAbsFlops); got[0] != 1 || got[1] != 0 {
		t.Fatalf("refined flops = %v, want only the first", got)
	}
	if p.nCexes != 1 || p.nCexesTotal != 1 {
		t.Fatalf("refinement counters = %d, %d", p.nCexes, p.nCexesTotal)
	}
}

func TestAbstractCounterexampleWithoutCarePPIIsConcrete(t *testing.T) {
	g := lifecycleGraph()
	p := Pdr_ManStart(g, &Pdr_Par_t{fUseAbs: 1}, nil)
	defer Pdr_ManStop(p)
	p.vMapPpi2Ff = vec.Vec_IntAlloc(1)
	p.vMapFf2Ppi = vec.Vec_IntStart(1)
	vec.Vec_IntWriteEntry(p.vMapFf2Ppi, 0, -1)
	p.vAbsFlops = vec.Vec_IntStart(1)
	Pdr_QueuePush(p, Pdr_OblStart(0, 0, testPdrSet(nil, []int32{0, 2}), nil))
	cex := Pdr_ManDeriveCexAbs(p)
	if cex == nil || saig.Saig_ManVerifyCex(g, cex) != 1 {
		t.Fatalf("expected concrete counterexample, got %+v", cex)
	}
	if vec.Vec_IntEntry(p.vAbsFlops, 0) != 0 || p.nCexes != 0 || p.nCexesTotal != 0 {
		t.Fatal("irrelevant abstract input caused refinement")
	}
}

func TestManagerStopsFanoutAsInC(t *testing.T) {
	for _, existing := range []bool{false, true} {
		g := lifecycleGraph()
		if existing {
			aig.Aig_ManFanoutStart(g)
		}
		p := Pdr_ManStart(g, &Pdr_Par_t{}, nil)
		if !aig.Aig_ManHasFanout(g) {
			t.Fatalf("existing=%v: missing fanout index", existing)
		}
		Pdr_ManStop(p)
		if aig.Aig_ManHasFanout(g) {
			t.Fatalf("existing=%v: manager retained fanout index", existing)
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
