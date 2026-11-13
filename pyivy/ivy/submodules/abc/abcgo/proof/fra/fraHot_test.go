package fra

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

func TestOneHotComputeAndConstraints(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	g.AddOutput(a)
	g.AddOutput(b)
	aig.Aig_ManSetRegNum(g, 2)
	sim := newFraSimulation(g, 0, 1, 1)
	sim.NodeWords(a.NodeID())[0] = 0x55555555
	sim.NodeWords(b.NodeID())[0] = 0xaaaaaaaa
	p := &Fra_Man_t{pManAig: g, pManFraig: g, pPars: &Fra_Par_t{NFramesK: 1}, pSml: sim, pSat: bsat.NewSolver(), pMemSatNums: make([]int32, len(g.Nodes()))}
	clauses := Fra_OneHotCompute(p, sim)
	if Fra_OneHotCount(p, clauses) != 1 {
		t.Fatal("expected one candidate")
	}
	exdc := Fra_OneHotCreateExdc(p, clauses)
	for mask := 0; mask < 4; mask++ {
		values := []bool{mask&1 != 0, mask&2 != 0}
		got, err := exdc.Evaluate(values)
		if err != nil {
			t.Fatal(err)
		}
		if got[0] != (mask != 3) {
			t.Fatalf("mask %d output %v", mask, got)
		}
	}
	p.pMemSatNums[a.NodeID()] = 1
	p.pMemSatNums[b.NodeID()] = 2
	Fra_OneHotAssume(p, clauses)
	if p.pSat.Solve(bsat.Positive(1), bsat.Positive(2)) != bsat.Unsatisfiable {
		t.Fatal("one-hot constraint not enforced")
	}
	if p.pSat.Solve(bsat.Positive(1), bsat.Negative(2)) != bsat.Satisfiable {
		t.Fatal("valid one-hot assignment rejected")
	}
}
