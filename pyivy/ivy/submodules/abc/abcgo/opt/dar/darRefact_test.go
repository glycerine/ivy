package dar

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestRefactorTruthAndBuild(t *testing.T) {
	g := aig.NewGraph()
	leaves := []aig.Edge{g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()}
	root := g.Mux(leaves[0], g.XOR(leaves[1], leaves[2]), g.AND(leaves[3], leaves[4].Not()))
	truth := refactorTruth(g, root, leaves)
	rebuilt := refactorBuild(g, leaves, truth)
	g.AddOutput(root)
	g.AddOutput(rebuilt)
	for n := 0; n < 32; n++ {
		values := make([]bool, 5)
		for i := range values {
			values[i] = n&(1<<i) != 0
		}
		out, err := g.Evaluate(values)
		if err != nil || out[0] != out[1] {
			t.Fatal(n, out, err)
		}
	}
	if err := Dar_ManRefactor(g, nil); err != nil {
		t.Fatal(err)
	}
	for n := 0; n < 32; n++ {
		values := make([]bool, 5)
		for i := range values {
			values[i] = n&(1<<i) != 0
		}
		out, err := g.Evaluate(values)
		if err != nil || out[0] != truth[n] || out[1] != truth[n] {
			t.Fatal(n, out, err)
		}
	}
}
func TestRefactorReduction(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	// Absorption is not a local structural-hash identity.
	g.AddOutput(g.OR(a, g.AND(a, g.OR(b, c))))
	before := g.NodeCount()
	if err := Dar_ManRefactor(g, nil); err != nil {
		t.Fatal(err)
	}
	if aig.Aig_ManDupDfs(g).NodeCount() >= before {
		t.Fatal("refactoring made no improvement")
	}
	for n := 0; n < 8; n++ {
		out, _ := g.Evaluate([]bool{n&1 != 0, n&2 != 0, n&4 != 0})
		if out[0] != (n&1 != 0) {
			t.Fatal(n, out)
		}
	}
}
func TestRefactorBoundaryExtension(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	shared := g.AND(a, b)
	root := g.AND(shared, c)
	g.AddOutput(shared)
	g.AddOutput(root)
	bounded, ok := refactorBoundary(g, root, 4, false)
	if !ok || len(bounded) != 2 {
		t.Fatal(bounded, ok)
	}
	extended, ok := refactorBoundary(g, root, 4, true)
	if !ok || len(extended) != 3 {
		t.Fatal(extended, ok)
	}
	if _, ok := refactorBoundary(g, root, 2, true); ok {
		t.Fatal("exceeded support limit")
	}
	p := DefaultRefactorParams()
	p.Extend = true
	if err := Dar_ManRefactor(g, &p); err != nil {
		t.Fatal(err)
	}
}
