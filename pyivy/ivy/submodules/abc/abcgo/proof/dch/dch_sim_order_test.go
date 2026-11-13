package dch

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

// Dch_PerformRandomSimulation in dchSim.c visits the CI interface array,
// which can have a different order from object IDs.
func TestRandomSimulationUsesCIInterfaceOrder(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	output := g.AddOutput(g.AND(a, b))
	if err := aig.Aig_ManReorderCisCos(g, []aig.Edge{b, a}, []aig.Edge{output}, 0); err != nil {
		t.Fatal(err)
	}
	order, err := graphOrder(g)
	if err != nil {
		t.Fatal(err)
	}
	rng := aig.NewAigRandomSource()
	wantB, wantA := rng.Uint32()<<1, rng.Uint32()<<1
	got := simulateWords(g, order, 1, aig.NewAigRandomSource())
	if got[b.NodeID()][0] != wantB || got[a.NodeID()][0] != wantA {
		t.Fatalf("CI patterns a=%08x b=%08x, want a=%08x b=%08x", got[a.NodeID()][0], got[b.NodeID()][0], wantA, wantB)
	}
}
