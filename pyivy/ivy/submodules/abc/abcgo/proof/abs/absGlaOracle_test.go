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

func glaOracleTruth2(g *gia.Gia_Man_t, x, y int32, truth int) int32 {
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
	return out
}
func TestGlaSequentialTruthOracle(t *testing.T) {
	for transition := 0; transition < 16; transition++ {
		for property := 0; property < 16; property++ {
			g := gia.Gia_ManStart(128)
			x := gia.Gia_ManAppendCi(g)
			q := gia.Gia_ManAppendCi(g)
			out := glaOracleTruth2(g, x, q, property)
			next := glaOracleTruth2(g, x, q, transition)
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
			pars.fUseSimple = 1
			pars.nFramesMax = 3
			pars.nRatioMin = 0
			pars.nRatioMin2 = 0
			pars.nFramesNoChangeLim = 100
			result := Gia_ManPerformGla(g, pars)
			if expected {
				if result != 0 || g.SequentialCounterexample() == nil || gia.Gia_ManVerifyCex(g, g.SequentialCounterexample(), 0) != 1 {
					t.Fatalf("trans=%x prop=%x result=%d witness=%+v", transition, property, result, g.SequentialCounterexample())
				}
			} else if result == 0 {
				t.Fatalf("trans=%x prop=%x false SAT", transition, property)
			}
		}
	}
}

// Exercise the non-default refinement and CNF options through the public
// driver, checking both results and witnesses against the transition table.
func TestGlaSequentialOracleWithAlternateOptions(t *testing.T) {
	for _, transition := range []int{0x0, 0x3, 0x6, 0xa, 0xc, 0xf} {
		for _, property := range []int{0x0, 0x2, 0x8, 0xe} {
			g := gia.Gia_ManStart(128)
			x, q := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
			out := glaOracleTruth2(g, x, q, property)
			next := glaOracleTruth2(g, x, q, transition)
			gia.Gia_ManAppendCo(g, out)
			gia.Gia_ManAppendCo(g, next)
			gia.Gia_ManSetRegNum(g, 1)
			expected := false
			for trace := 0; trace < 8; trace++ {
				state := 0
				for frame := 0; frame < 3; frame++ {
					index := ((trace >> frame) & 1) | state<<1
					if property&(1<<index) != 0 {
						expected = true
						break
					}
					state = (transition >> index) & 1
				}
			}
			pars := NewParameters()
			options := pars.Options()
			options.FramesMax = 3
			options.FramesNoChangeLimit = 100
			options.RatioMin, options.RatioMinRefine = 0, 0
			options.NewRefinement, options.SkipHash = true, true
			options.PropagateFanout = false
			pars.SetOptions(options)
			result := Gia_ManPerformGla(g, pars)
			if expected {
				if result != 0 || g.SequentialCounterexample() == nil || gia.Gia_ManVerifyCex(g, g.SequentialCounterexample(), 0) != 1 {
					t.Fatalf("transition=%x property=%x result=%d witness=%+v", transition, property, result, g.SequentialCounterexample())
				}
			} else if result == 0 {
				t.Fatalf("transition=%x property=%x: false SAT", transition, property)
			}
		}
	}
}
