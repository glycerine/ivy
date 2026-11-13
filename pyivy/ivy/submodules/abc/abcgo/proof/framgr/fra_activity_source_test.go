package framgr

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestActivityTraversalChecksSourcePreconditionsBeforeLevelCutoff(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	gate := g.AND(a, b)
	p := satTestManager(g)
	for _, tc := range []struct {
		name string
		node aig.Edge
	}{{"missing SAT variable", a}, {"complemented parent", gate.Not()}} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Error("source asserts before level cutoff")
				}
			}()
			Fra_SetActivityFactors_rec(p, tc.node, 1, 1, make(map[int]bool))
		})
	}
}
