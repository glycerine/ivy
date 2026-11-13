package framgr

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

// fraSat.c skips marked nodes at limits <= 10 only for non-speculative,
// combinational sweeping; other modes still issue a SAT query.
func TestEquivalenceFailedNodeBudgetPolicy(t *testing.T) {
	cases := []struct {
		name                     string
		limit, frames, speculate int32
		marked                   bool
		want, failures, calls    int32
	}{
		{"marked small", 10, 0, 0, true, -1, 1, 0},
		{"marked larger", 11, 0, 0, true, 0, 1, 1},
		{"unlimited", 0, 0, 0, true, 0, 0, 1},
		{"speculative", 10, 0, 1, true, 0, 0, 1},
		{"sequential", 10, 1, 0, true, 0, 0, 1},
		{"unmarked", 10, 0, 0, false, 0, 0, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := aig.NewGraph()
			a, b := g.NewInput(), g.NewInput()
			p := satTestManager(g)
			p.PPars.NBTLimitNode, p.PPars.NFramesK, p.PPars.FSpeculate = tc.limit, tc.frames, tc.speculate
			if tc.marked {
				aig.Aig_ObjSetMarkB(g, a)
			}
			if got := Fra_NodesAreEquiv(p, a, b); got != tc.want {
				t.Fatalf("result=%d want %d", got, tc.want)
			}
			if p.NSatFails != tc.failures || p.NSatCalls != tc.calls {
				t.Fatalf("failures=%d calls=%d", p.NSatFails, p.NSatCalls)
			}
		})
	}
}
