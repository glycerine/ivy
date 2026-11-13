package msat

import (
	"strings"
	"testing"
)

func TestSolverCleanKeepsStatisticsAndModelButEmptiesOrder(t *testing.T) {
	p := New(2)
	p.Stats.nDecisions = 7
	p.pModel[0] = 1
	p.nSeenId = 42
	Msat_SolverClean(p, 2)
	if p.Stats.nDecisions != 7 || p.pModel[0] != 1 {
		t.Fatal("C retains solver statistics and model on clean")
	}
	if len(p.pOrder.heap) != 0 || p.nSeenId != 1 {
		t.Fatalf("C empties decision order and resets seen id: order=%v seen=%d", p.pOrder.heap, p.nSeenId)
	}
}

func TestSourceDIMACSParserLeavesOrderUninitialized(t *testing.T) {
	var p *Msat_Solver_t
	if Msat_SolverParseDimacs(strings.NewReader("p cnf 2 1\n1 2 0\n"), &p, 0) == 0 {
		t.Fatal("parse")
	}
	if len(p.pOrder.heap) != 0 {
		t.Fatalf("C parser leaves order empty: %v", p.pOrder.heap)
	}
	q, err := ParseDIMACS(strings.NewReader("p cnf 2 1\n1 2 0\n"))
	if err != nil || len(q.pOrder.heap) != 2 {
		t.Fatal("Go parser should prepare its convenience order")
	}
}
