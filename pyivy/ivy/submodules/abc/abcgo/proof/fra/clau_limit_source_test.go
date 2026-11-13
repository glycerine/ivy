package fra

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

// ABC fraClau.c checks the property before comparing the iteration bound,
// then declares success unless Iter equals nIters, even for negative limits.
func TestClauIterationLimitSourceBehavior(t *testing.T) {
	for _, tc := range []struct{ limit, want int32 }{{-1, 1}, {0, 0}, {1, 0}} {
		g := aig.NewGraph()
		input, state := g.NewInput(), g.NewInput()
		g.AddOutput(state)
		g.AddOutput(input)
		aig.Aig_ManSetRegNum(g, 1)
		if got := Fra_Clau(g, tc.limit, 0, 0); got != tc.want {
			t.Errorf("limit %d: got %d, want %d", tc.limit, got, tc.want)
		}
	}
}
