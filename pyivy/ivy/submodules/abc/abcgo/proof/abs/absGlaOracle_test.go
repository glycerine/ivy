package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

// Each two-input truth table is checked in both refinement modes against an
// exhaustive combinational oracle; any reported counterexample is replayed.
func TestGlaCombinationalTruthOracle(t *testing.T) {
	for truth := 0; truth < 16; truth++ {
		for _, layer := range []int32{0, 1} {
			g := gia.Gia_ManStart(64)
			x := gia.Gia_ManAppendCi(g)
			y := gia.Gia_ManAppendCi(g)
			out := int32(0)
			for bits := 0; bits < 4; bits++ {
				if truth&(1<<bits) == 0 {
					continue
				}
				a, b := x, y
				if bits&1 == 0 {
					a ^= 1
				}
				if bits&2 == 0 {
					b ^= 1
				}
				cube := gia.Gia_ManAppendAnd(g, a, b)
				out = gia.Gia_ManAppendAnd(g, out^1, cube^1) ^ 1
			}
			gia.Gia_ManAppendCo(g, out)
			pars := &Abs_Par_t{}
			Abs_ParSetDefaults(pars)
			pars.fUseSimple = 1
			pars.fAddLayer = layer
			pars.nFramesMax = 1
			pars.nRatioMin = 0
			pars.nRatioMin2 = 0
			pars.nFramesNoChangeLim = 100
			got := Gia_ManPerformGla(g, pars)
			if truth == 0 {
				if got != 1 {
					t.Fatalf("truth=%x layer=%d result=%d", truth, layer, got)
				}
				continue
			}
			if got != 0 || g.SequentialCounterexample() == nil || gia.Gia_ManVerifyCex(g, g.SequentialCounterexample(), 0) != 1 {
				t.Fatalf("truth=%x layer=%d result=%d witness=%+v", truth, layer, got, g.SequentialCounterexample())
			}
		}
	}
}
