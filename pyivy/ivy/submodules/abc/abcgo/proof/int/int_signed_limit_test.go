package int

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestInterpolationSignedConflictLimit(t *testing.T) {
	for _, tc := range []struct{ limit, want int32 }{{-1, -1}, {0, 0}, {1, 0}} {
		g := aig.NewGraph()
		x, a, b := g.NewInput(), g.NewInput(), g.NewInput()
		g.AddOutput(g.OR(g.AND(x, a), g.AND(x, b)))
		g.AddOutput(g.XOR(x, a))
		g.AddOutput(g.XOR(a, b))
		aig.Aig_ManSetRegNum(g, 2)
		var p Inter_ManParams_t
		Inter_ManSetDefaultParams(&p)
		p.NBTLimit = tc.limit
		var frame int32
		if got := Inter_ManPerformInterpolation(g, &p, &frame); got != tc.want {
			t.Errorf("limit=%d got=%d want=%d", tc.limit, got, tc.want)
		}
	}
}
