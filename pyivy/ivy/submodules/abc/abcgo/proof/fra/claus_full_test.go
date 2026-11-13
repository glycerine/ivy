package fra

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/cnf"
	"testing"
)

func TestClausIntegratedPropertyChecks(t *testing.T) {
	for _, unsafe := range []bool{false, true} {
		g := aig.NewGraph()
		x, state := g.NewInput(), g.NewInput()
		g.AddOutput(state)
		if unsafe {
			g.AddOutput(x)
		} else {
			g.AddOutput(g.Const0())
		}
		aig.Aig_ManSetRegNum(g, 1)
		p := Fra_ClausAlloc(g, 2, 0, 20, 4, 10, 10, 1, 0, 1, 0, 0)
		p.pCnf = cnf.Cnf_DeriveSimple(g, int32(g.OutputCount()))
		p.pSatBmc = cnf.Cnf_DataWriteIntoSolver(p.pCnf, 2, 1)
		got := Fra_ClausRunBmc(p)
		want := int32(1)
		if unsafe {
			want = 0
		}
		if got != want {
			t.Fatalf("unsafe=%v BMC=%d want=%d", unsafe, got, want)
		}
		p.pSatMain = cnf.Cnf_DataWriteIntoSolver(p.pCnf, 3, 0)
		if !unsafe && Fra_ClausRunSat(p) != 1 {
			t.Fatal("safe property not inductive")
		}
	}
}
func TestClausIntegratedStorageAndRefinement(t *testing.T) {
	g := aig.NewGraph()
	state := g.NewInput()
	g.AddOutput(state)
	aig.Aig_ManSetRegNum(g, 1)
	p := Fra_ClausAlloc(g, 1, 0, 10, 4, 10, 10, 1, 0, 0, 0, 0)
	p.pCnf = cnf.Cnf_DeriveSimple(g, int32(g.OutputCount()))
	lit := 2*p.pCnf.VarNums[state.NodeID()] + 1
	vec.Vec_IntPush(p.vLits, lit)
	vec.Vec_IntPush(p.vClauses, 1)
	vec.Vec_IntPush(p.vCosts, 1)
	p.nClauses = 1
	if got := Fra_ClausInductiveClauses(p); got != 0 {
		t.Fatalf("identity invariant rejected: %d", got)
	}
	Fra_ClausAddToStorage(p)
	if p.nClauses != 0 || vec.Vec_IntSize(p.vClausesProven) != 1 || vec.Vec_IntEntry(p.vLitsProven, 0) != lit {
		t.Fatal("stored clause mismatch")
	}
}
