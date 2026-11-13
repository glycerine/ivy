package fra

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/cnf"
	"testing"
)

func TestClausCutScoresAndClausePolarity(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	s := Fra_SmlStart(g, 0, 1, 8)
	for i := 0; i < 8; i++ {
		s.NodeWords(a.NodeID())[i] = 0xaaaaaaaa
		s.NodeWords(b.NodeID())[i] = 0xcccccccc
	}
	p := &Clu_Man_t{pCnf: &cnf.Cnf_Dat_t{VarNums: []int32{0, 2, 3}}, vLits: vec.Vec_IntAlloc(4), vClauses: vec.Vec_IntAlloc(2), vCosts: vec.Vec_IntAlloc(2)}
	scores := clausScores(p, s, []aig.Edge{a, b})
	if len(scores) != 4 {
		t.Fatal(scores)
	}
	for _, v := range scores {
		if v != 64 {
			t.Fatal(scores)
		}
	}
	clausRecord(p, []aig.Edge{a, b}, 1, 64)
	lits := vec.Vec_IntArray(p.vLits)
	if len(lits) != 2 || lits[0] != 5 || lits[1] != 6 {
		t.Fatalf("bad minterm clause %v", lits)
	}
}

func TestClausCandidatesExcludeUnreachableLatchStates(t *testing.T) {
	g := aig.NewGraph()
	input, state := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(input, state))
	g.AddOutput(state)
	aig.Aig_ManSetRegNum(g, 1)
	p := &Clu_Man_t{pAig: g, pCnf: cnf.Cnf_DeriveSimple(g, int32(g.OutputCount())), nSimWords: 16, nSimFrames: 2, nLutSize: 4, nLevels: 10, nCutsMax: 10, nClausesMax: 100, vLits: vec.Vec_IntAlloc(16), vClauses: vec.Vec_IntAlloc(16), vCosts: vec.Vec_IntAlloc(16), vLitsProven: vec.Vec_IntAlloc(4), vClausesProven: vec.Vec_IntAlloc(4)}
	if Fra_ClausProcessClauses2(p, 0) != 1 || p.nClauses == 0 {
		t.Fatalf("no candidates: cuts=%d clauses=%d", p.nCuts, p.nClauses)
	}
	lits, ends := vec.Vec_IntArray(p.vLits), vec.Vec_IntArray(p.vClauses)
	start := 0
	for _, end := range ends {
		for x := int32(0); x < 2; x++ {
			holds := false
			for _, lit := range lits[start:int(end)] {
				v := lit >> 1
				value := int32(0)
				if v == p.pCnf.VarNums[input.NodeID()] {
					value = x
				} else if v != p.pCnf.VarNums[state.NodeID()] {
					t.Fatalf("unexpected literal %d", lit)
				}
				if value != (lit & 1) {
					holds = true
				}
			}
			if !holds {
				t.Fatalf("clause %v rejects reachable x=%d, state=0", lits[start:int(end)], x)
			}
		}
		start = int(end)
	}
}

func TestClausCandidatesFilterPreviouslyProved(t *testing.T) {
	g := aig.NewGraph()
	x, state := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(x, state))
	g.AddOutput(state)
	aig.Aig_ManSetRegNum(g, 1)
	p := &Clu_Man_t{pAig: g, pCnf: cnf.Cnf_DeriveSimple(g, int32(g.OutputCount())), nSimWords: 16, nSimFrames: 2, nLutSize: 4, nLevels: 10, nCutsMax: 10, nClausesMax: 100, vLits: vec.Vec_IntAlloc(16), vClauses: vec.Vec_IntAlloc(16), vCosts: vec.Vec_IntAlloc(16), vLitsProven: vec.Vec_IntAlloc(4), vClausesProven: vec.Vec_IntAlloc(4)}
	if Fra_ClausProcessClauses2(p, 0) != 1 || p.nClauses == 0 {
		t.Fatal("no candidates")
	}
	firstEnd := vec.Vec_IntEntry(p.vClauses, 0)
	for _, lit := range vec.Vec_IntArray(p.vLits)[:firstEnd] {
		vec.Vec_IntPush(p.vLitsProven, lit)
	}
	vec.Vec_IntPush(p.vClausesProven, firstEnd)
	vec.Vec_IntClear(p.vLits)
	vec.Vec_IntClear(p.vClauses)
	vec.Vec_IntClear(p.vCosts)
	p.nClauses = 0
	if Fra_ClausProcessClauses2(p, 0) != 1 {
		t.Fatal("failed to filter candidates")
	}
	if vec.Vec_IntEntry(p.vCosts, 0) != -1 {
		t.Fatalf("previously proved candidate retained: %v", vec.Vec_IntArray(p.vCosts))
	}
}

// The source collects register clauses only when refs is enabled. Without it,
// this cut-free design has no strengthening candidates and remains unproved.
func TestClausLatchStrengtheningDisabledWithoutRefs(t *testing.T) {
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
	if Fra_ClausRunBmc(p) != 1 {
		t.Fatal("initial state violates property")
	}
	p.pSatMain = cnf.Cnf_DataWriteIntoSolver(p.pCnf, 2, 0)
	if Fra_ClausRunSat(p) != 0 {
		t.Fatal("property unexpectedly 1-inductive")
	}
	if Fra_ClausProcessClauses2(p, 0) != 1 || p.nClauses != 0 {
		t.Fatalf("unexpected candidates: %d", p.nClauses)
	}
	if removed := Fra_ClausInductiveClauses(p); removed != 0 || p.fFail != 1 {
		t.Fatalf("reported proof without strengthening: removed=%d fail=%d", removed, p.fFail)
	}
}
