package abs

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"testing"
)

func TestLegacySequentialUnrollOracle(t *testing.T) {
	for transition := 0; transition < 16; transition++ {
		for trace := 0; trace < 8; trace++ {
			g := aig.NewGraph()
			x, lo := g.NewInput(), g.NewInput()
			g.AddOutput(g.OR(g.AND(x, lo.Not()), g.AND(x.Not(), lo)))
			g.AddOutput(oldTruth2(g, x, lo, transition))
			aig.Aig_ManSetRegNum(g, 1)
			cex := util.Abc_CexAlloc(1, 1, 2)
			cex.IFrame = 1
			for bit := 0; bit < 3; bit++ {
				if trace&(1<<bit) != 0 {
					util.Abc_InfoSetBit(cex.PData, int32(bit))
				}
			}
			p := Saig_RefManStart(g, cex, 0, 0)
			got := Saig_RefManSetPhases(p, nil, 0)
			init, x0, x1 := trace&1, (trace>>1)&1, (trace>>2)&1
			state := (transition >> (x0 + 2*init)) & 1
			want := int32(1)
			if state != x1 {
				want = 0
			}
			if got != want {
				t.Fatalf("transition %04b trace %03b: got %d want %d", transition, trace, got, want)
			}
		}
	}
}
