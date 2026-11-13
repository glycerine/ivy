package fra

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestClauseInductionSATQueries(t *testing.T) {
	g := aig.NewGraph()
	x, state := g.NewInput(), g.NewInput()
	g.AddOutput(state)
	g.AddOutput(x)
	aig.Aig_ManSetRegNum(g, 1)
	p := clauStart(g)
	if p == nil {
		t.Fatal("expected two-frame counterexample")
	}
	status, cube := p.checkProperty()
	if status != 0 || cube == nil || vec.Vec_IntSize(cube) != 1 {
		t.Fatalf("status=%d cube=%v", status, cube)
	}
	current := clauRemap(p.mainToTest, cube, false)
	for i := 0; i < 2; i++ {
		status, cex := p.checkClause(current)
		if status != 0 || cex == nil || vec.Vec_IntSize(cex) != 1 {
			t.Fatalf("attempt %d status=%d cex=%v", i, status, cex)
		}
		if vec.Vec_IntEntry(cex, 0) == vec.Vec_IntEntry(current, 0) {
			t.Fatal("antecedent was not excluded")
		}
	}
	next := clauRemap(p.testToBMC, current, false)
	if status := p.checkBMC(next); status != 0 {
		t.Fatalf("expected reachable next state, got %d", status)
	}
}
