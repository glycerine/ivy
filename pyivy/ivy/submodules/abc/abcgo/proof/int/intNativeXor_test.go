package int

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

// The source's Aig_ManForEachNode copies each native EXOR through Aig_And.
// Confirmed against Inter_ManFramesInter in a compiled ABC source harness.
func TestNativeXorFrameCopyMatchesSourceAnd(t *testing.T) {
	g := aig.NewGraph()
	x, state := g.NewInput(), g.NewInput()
	g.AddOutput(g.NewXor(x, state))
	g.AddOutput(state)
	aig.Aig_ManSetRegNum(g, 1)
	frames := Inter_ManFramesInter(g, 1, 0, 0)
	for mask := 0; mask < 4; mask++ {
		initial, input := mask&1 != 0, mask&2 != 0
		out, err := frames.Evaluate([]bool{initial, input})
		if err != nil || out[0] != (initial && input) {
			t.Fatalf("state=%v input=%v output=%v error=%v", initial, input, out, err)
		}
	}
}

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
