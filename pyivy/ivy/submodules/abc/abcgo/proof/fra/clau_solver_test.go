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

func TestClauseInductionTopLevel(t *testing.T) {
	for _, unsafe := range []bool{false, true} {
		g := aig.NewGraph()
		x, state := g.NewInput(), g.NewInput()
		g.AddOutput(state)
		if unsafe {
			g.AddOutput(x)
		} else {
			g.AddOutput(state)
		}
		aig.Aig_ManSetRegNum(g, 1)
		want := int32(1)
		if unsafe {
			want = 0
		}
		if got := Fra_Clau(g, 5, 0, 0); got != want {
			t.Fatalf("unsafe=%v got=%d want=%d", unsafe, got, want)
		}
	}
}
