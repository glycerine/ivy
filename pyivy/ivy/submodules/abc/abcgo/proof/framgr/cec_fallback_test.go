package framgr

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/proof/fracomb"
	"testing"
)

func TestCECDriverUsesRealSweepAfterInitialSATLimit(t *testing.T) {
	const pigeons, holes = 8, 7
	g := aig.NewGraph()
	vars := make([][]aig.Edge, pigeons)
	for i := range vars {
		vars[i] = make([]aig.Edge, holes)
		for j := range vars[i] {
			vars[i][j] = g.NewInput()
		}
	}
	formula := g.Const1()
	for i := 0; i < pigeons; i++ {
		clause := g.Const0()
		for j := 0; j < holes; j++ {
			clause = g.OR(clause, vars[i][j])
		}
		formula = g.AND(formula, clause)
	}
	for hole := 0; hole < holes; hole++ {
		for i := 0; i < pigeons; i++ {
			for j := 0; j < i; j++ {
				formula = g.AND(formula, g.OR(vars[i][hole].Not(), vars[j][hole].Not()))
			}
		}
	}
	g.AddOutput(formula)
	calls := 0
	status := fracomb.Fra_FraigCec(&g, 100000, 0, func(g *aig.Graph, n, m int32) *aig.Graph { calls++; return SweepForCEC(g, n, m) })
	if calls == 0 {
		t.Fatal("fixture did not exercise the FRAIG sweep after the initial SAT limit")
	}
	if status != 1 {
		t.Fatalf("pigeonhole is unsatisfiable: status %d after %d sweeps", status, calls)
	}
}
