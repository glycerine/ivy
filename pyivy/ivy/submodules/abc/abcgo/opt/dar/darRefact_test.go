package dar

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/bool/kit"
	"testing"
)

func TestRefactorTruthAndBuild(t *testing.T) {
	g := aig.NewGraph()
	leaves := []aig.Edge{g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()}
	root := g.Mux(leaves[0], g.XOR(leaves[1], leaves[2]), g.AND(leaves[3], leaves[4].Not()))
	truth := refactorTruth(g, root, leaves)
	words := []uint32{0}
	for i, bit := range truth {
		if bit {
			words[0] |= 1 << uint(i)
		}
	}
	var cover []uint32
	kit.Kit_TruthIsop(words, 5, &cover, 0)
	rebuilt := refactorBuildKit(g, leaves, kit.Kit_SopFactor(cover, 0, 5, &cover))
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
	var bounded, extended []aig.Edge
	aig.Aig_NodeMffcSupp(g, root, 0, &bounded)
	if len(bounded) != 2 {
		t.Fatal(bounded)
	}
	if !aig.Aig_NodeMffcExtendCut(g, root, bounded, &extended) || len(extended) != 3 {
		t.Fatal(extended)
	}
	p := DefaultRefactorParams()
	p.Extend = true
	if err := Dar_ManRefactor(g, &p); err != nil {
		t.Fatal(err)
	}
}

func TestRefactorConstantsAndMaximumLeafLimit(t *testing.T) {
	for _, negate := range []bool{false, true} {
		g := aig.NewGraph()
		a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
		root := g.AND(g.AND(a, b), g.AND(a.Not(), c))
		if negate {
			root = root.Not()
		}
		g.AddOutput(root)
		if err := Dar_ManRefactor(g, nil); err != nil {
			t.Fatal(err)
		}
		for mask := 0; mask < 8; mask++ {
			out, err := g.Evaluate([]bool{mask&1 != 0, mask&2 != 0, mask&4 != 0})
			if err != nil || out[0] != negate {
				t.Fatalf("negate=%v mask=%d output=%v err=%v", negate, mask, out, err)
			}
		}
		if compact := aig.Aig_ManDupDfs(g); compact.NodeCount() != compact.InputCount()+compact.OutputCount()+1 {
			t.Fatalf("constant cone retained gates; total nodes %d", compact.NodeCount())
		}
	}
	g := aig.NewGraph()
	e := g.NewInput()
	for i := 1; i < 15; i++ {
		e = g.AND(e, g.NewInput())
	}
	g.AddOutput(e)
	p := DefaultRefactorParams()
	p.LeafMax = 15
	if err := Dar_ManRefactor(g, &p); err != nil {
		t.Fatalf("source Kit supports 15 leaves: %v", err)
	}
	p.LeafMax = 16
	if err := Dar_ManRefactor(g, &p); err == nil {
		t.Fatal("Kit source rejects 16 leaves")
	}
}
