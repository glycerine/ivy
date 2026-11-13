package fra

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/cnf"
	"testing"
)

func TestClausTopLevelTrivialInduction(t *testing.T) {
	graph := aig.NewGraph()
	graph.NewInput()
	state := graph.NewInput()
	graph.AddOutput(state)
	graph.AddOutput(graph.Const0())
	aig.Aig_ManSetRegNum(graph, 1)
	// The all-zero register is preserved and the property is one-step inductive.
	if got := Fra_Claus(graph, 1, 0, 32, 4, 8, 8, 1, 0, 1, 0, 1, 0, 0); got != 1 {
		t.Fatalf("got %d", got)
	}
}

func TestClausStrengthensTwoRegisterInvariant(t *testing.T) {
	g := aig.NewGraph()
	g.NewInput()
	first, second := g.NewInput(), g.NewInput()
	g.AddOutput(first)
	g.AddOutput(second)
	g.AddOutput(g.Const0())
	aig.Aig_ManSetRegNum(g, 2)
	p := Fra_ClausAlloc(g, 1, 0, 32, 4, 8, 8, 1, 0, 1, 0, 0)
	p.pCnf = cnf.Cnf_DeriveSimple(g, int32(g.OutputCount()))
	p.pSatBmc = cnf.Cnf_DataWriteIntoSolver(p.pCnf, 1, 1)
	p.pSatMain = cnf.Cnf_DataWriteIntoSolver(p.pCnf, 2, 0)
	if Fra_ClausRunSat(p) != 0 {
		t.Fatal("property should need strengthening")
	}
	if Fra_ClausProcessClauses2(p, 1) != 1 || p.nClauses == 0 {
		t.Fatal("no strengthening candidates")
	}
	p.nClauses -= Fra_ClausBmcClauses(p)
	for {
		removed := Fra_ClausInductiveClauses(p)
		if removed < 0 {
			t.Fatal("induction failed")
		}
		p.nClauses -= removed
		if removed == 0 {
			break
		}
	}
	if p.fFail != 0 {
		t.Fatal("strengthened property still fails")
	}
	Fra_ClausAddToStorage(p)
	if vec.Vec_IntSize(p.vClausesProven) == 0 || Fra_InvariantVerify(g, 1, p.vClausesProven, p.vLitsProven) != 1 {
		t.Fatal("strengthening is not a verified invariant")
	}
}
