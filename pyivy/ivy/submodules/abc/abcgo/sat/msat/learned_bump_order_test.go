package msat

import (
	"slices"
	"testing"
)

// msatClause.c bumps a new learned clause's variables in the order of the
// caller's literal vector (pLits), not in the clause's own order, where the
// literal with the highest decision level has been swapped into the second
// watch. With equal activities the bump order decides the heap layout, so
// the port diverged from ABC's variable order (fraig on i10: decision 99).
func TestLearnedClauseBumpsInCallerOrder(t *testing.T) {
	build := func() *Msat_Solver_t {
		p := Msat_SolverAlloc(16, 1, 1, 1, 1, 0)
		for i := 0; i < 8; i++ {
			Msat_SolverAddVar(p, 0)
		}
		cone := Msat_IntVecAlloc(8)
		for v := 0; v < 8; v++ {
			Msat_IntVecPush(cone, v)
		}
		Msat_SolverPrepare(p, cone)
		return p
	}
	lits := []int{2 * 5, 2 * 6, 2 * 7, 2 * 4} // the last literal has the highest level
	got := build()
	for i, lit := range lits {
		got.pLevel[lit>>1] = 1 + i/3*2
	}
	v := Msat_IntVecAlloc(4)
	for _, lit := range lits {
		Msat_IntVecPush(v, lit)
	}
	var c *Msat_Clause_t
	if Msat_ClauseCreate(got, v, 1, &c) == 0 || c == nil {
		t.Fatal("learned clause not created")
	}
	if c.lits[1] != lits[3] {
		t.Fatalf("second watch = %d, want the highest-level literal %d", c.lits[1], lits[3])
	}
	want := build()
	for _, lit := range lits { // ABC's order
		Msat_SolverVarBumpActivity(want, lit)
	}
	if !slices.Equal(got.pOrder.heap, want.pOrder.heap) {
		t.Fatalf("variable heap after learning %v, ABC's bump order gives %v", got.pOrder.heap, want.pOrder.heap)
	}
}
