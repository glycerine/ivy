package core

import (
	"context"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

// The original C first rewrite leaves 32 AND nodes in this shared MUX chain.
// Stable replacement IDs matter to subsequent cut enumeration in this pass.
func TestSourceRewritePreservesReplacementIdentity(t *testing.T) {
	g := aig.NewGraph()
	var inputs [8]aig.Edge
	var outputs [4]aig.Edge
	for i := range inputs {
		inputs[i] = g.NewInput()
	}
	parity := inputs[0]
	for i := 1; i <= 24; i++ {
		left := g.AND(inputs[(i+1)%8], parity)
		right := g.AND(inputs[(i+1)%8].Not(), inputs[(i+2)%8])
		parity = g.OR(left, right)
		if i%6 == 0 {
			outputs[i/6-1] = parity
		}
	}
	for _, output := range outputs {
		g.AddOutput(output)
	}
	g = aig.Aig_ManDupDfs(g)
	g, err := Balance(context.Background(), g, false)
	if err != nil {
		t.Fatal(err)
	}
	params := DefaultRewriteParams()
	params.UpdateLevel = false
	if err := Dar_ManRewrite(g, &params); err != nil {
		t.Fatal(err)
	}
	g = aig.Aig_ManDupDfs(g)
	if got := aig.Aig_ManNodeNum(g); got != 32 {
		t.Fatalf("first rewrite has %d nodes; C has 32", got)
	}
}
