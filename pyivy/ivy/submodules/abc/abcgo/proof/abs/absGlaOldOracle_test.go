package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

// Exhaustively compare old GLA against concrete three-frame traces of every
// one-input, one-register transition/property pair.
func TestOldGlaSequentialTruthOracle(t *testing.T) {
	for transition := 0; transition < 16; transition++ {
		for property := 0; property < 16; property++ {
			for _, layer := range []int32{0, 1} {
				g := gia.Gia_ManStart(128)
				x := gia.Gia_ManAppendCi(g)
				q := gia.Gia_ManAppendCi(g)
				gia.Gia_ManHashAlloc(g)
				makeTruth := func(truth int) int32 {
					out := int32(0)
					for bits := 0; bits < 4; bits++ {
						if truth&(1<<bits) == 0 {
							continue
						}
						a, b := x, q
						if bits&1 == 0 {
							a ^= 1
						}
						if bits&2 == 0 {
							b ^= 1
						}
						out = gia.Gia_ManHashOr(g, out, gia.Gia_ManHashAnd(g, a, b))
					}
					return out
				}
				out := makeTruth(property)
				next := makeTruth(transition)
				gia.Gia_ManHashStop(g)
				gia.Gia_ManAppendCo(g, out)
				gia.Gia_ManAppendCo(g, next)
				gia.Gia_ManSetRegNum(g, 1)
				expected := false
				for trace := 0; trace < 8; trace++ {
					state := 0
					for frame := 0; frame < 3; frame++ {
						input := (trace >> frame) & 1
						index := input | (state << 1)
						if property&(1<<index) != 0 {
							expected = true
							break
						}
						state = (transition >> index) & 1
					}
				}
				pars := &Abs_Par_t{}
				Abs_ParSetDefaults(pars)
				pars.fAddLayer = layer
				pars.nFramesMax = 3
				pars.nRatioMin = 0
				pars.nRatioMin2 = 0
				pars.nFramesNoChangeLim = 100
				result := Gia_ManPerformGlaOld(g, pars, 0)
				if expected {
					if result != 0 || g.SequentialCounterexample() == nil || gia.Gia_ManVerifyCex(g, g.SequentialCounterexample(), 0) != 1 {
						t.Fatalf("transition=%x property=%x layer=%d result=%d witness=%+v", transition, property, layer, result, g.SequentialCounterexample())
					}
				} else if result == 0 {
					t.Fatalf("transition=%x property=%x layer=%d false SAT", transition, property, layer)
				}
			}
		}
	}
}
