package int

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestInterpolationNativeXorAndMux(t *testing.T) {
	for _, native := range []bool{false, true} {
		for mode := 0; mode < 3; mode++ {
			g := aig.NewGraph()
			x, a, b := g.NewInput(), g.NewInput(), g.NewInput()
			g.AddOutput(g.AND(a, b))
			var next aig.Edge
			if native {
				next = g.NewXor(x, a)
			} else {
				next = g.XOR(x, a)
			}
			g.AddOutput(next)
			g.AddOutput(g.Mux(x, a, b))
			aig.Aig_ManSetRegNum(g, 2)
			var p Inter_ManParams_t
			Inter_ManSetDefaultParams(&p)
			p.NFramesMax = 8
			if mode > 0 {
				p.FUseMiniSat = 1
				p.FUsePudlak = int32(mode - 1)
			}
			var frame int32
			got := Inter_ManPerformInterpolation(g, &p, &frame)
			// Input sequence 1,1,0 reaches state (0,1), then 1 reaches (1,0),
			// and another 1 reaches (0,1): (1,1) is unreachable.
			if got != 1 {
				t.Fatalf("native=%v mode=%d: got %d frame %d; want safe", native, mode, got, frame)
			}
		}
	}
}
