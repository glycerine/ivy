package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

// replayCareFixture is deliberately independent of GIA and ternary simulation.
// Input pairs are (a,b); the register starts at zero.
func replayCareFixture(mode int, inputs []bool) bool {
	state := false
	for frame := 0; frame < len(inputs)/2; frame++ {
		a, b := inputs[2*frame], inputs[2*frame+1]
		var out, next bool
		switch mode {
		case 0:
			out = state && a
			next = a || b
		case 1:
			out = state || (a && b)
			next = a && !b
		case 2:
			out = (state && a) || (!state && b)
			next = a != b
		}
		if frame == len(inputs)/2-1 {
			return out
		}
		state = next
	}
	return false
}

func careFixture(mode int) *gia.Gia_Man_t {
	p := gia.Gia_ManStart(32)
	a, b, r := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	gia.Gia_ManHashStart(p)
	var out, next int32
	switch mode {
	case 0:
		out = gia.Gia_ManHashAnd(p, r, a)
		next = gia.Gia_ManHashOr(p, a, b)
	case 1:
		out = gia.Gia_ManHashOr(p, r, gia.Gia_ManHashAnd(p, a, b))
		next = gia.Gia_ManHashAnd(p, a, b^1)
	case 2:
		out = gia.Gia_ManHashMux(p, r, a, b)
		next = gia.Gia_ManHashXor(p, a, b)
	}
	gia.Gia_ManAppendCo(p, out)
	gia.Gia_ManAppendCo(p, next)
	gia.Gia_ManHashStop(p)
	gia.Gia_ManSetRegNum(p, 1)
	return p
}

// Every completion of a computed care set must replay the failure; checking
// ternary verification alone would repeat the implementation's assumptions.
func TestCexCareExhaustiveCompletions(t *testing.T) {
	const frames = 3
	for mode := 0; mode < 3; mode++ {
		g := careFixture(mode)
		for pattern := 0; pattern < 1<<(2*frames); pattern++ {
			input := make([]bool, 2*frames)
			for i := range input {
				input[i] = pattern&(1<<i) != 0
			}
			if !replayCareFixture(mode, input) {
				continue
			}
			cex := util.Abc_CexAlloc(1, 2, frames)
			cex.IFrame = frames - 1
			for i, v := range input {
				if v {
					util.Abc_InfoSetBit(cex.PData, int32(1+i))
				}
			}
			var implied *util.Abc_Cex_t
			states := Bmc_CexInnerStates(g, cex, &implied, 0)
			for _, all := range []int32{0, 1} {
				care := Bmc_CexCareBits(g, states, implied, nil, all, 0)
				if Bmc_CexVerify(g, cex, care) != 1 {
					t.Fatalf("mode %d pattern %06b all %d: care rejected", mode, pattern, all)
				}
				for candidate := 0; candidate < 1<<(2*frames); candidate++ {
					trial := make([]bool, len(input))
					valid := true
					for i := range trial {
						trial[i] = candidate&(1<<i) != 0
						if util.Abc_InfoHasBit(care.PData, int32((i/2)*3+i%2)) && trial[i] != input[i] {
							valid = false
							break
						}
					}
					if valid && !replayCareFixture(mode, trial) {
						t.Fatalf("mode %d pattern %06b all %d permits nonfailing %06b, care=%v", mode, pattern, all, candidate, care.PData)
					}
				}
			}
		}
	}
}

func TestCexMinReasonsExhaustiveCompletions(t *testing.T) {
	const frames = 3
	for mode := 0; mode < 3; mode++ {
		graph := gia.Gia_ManToAig(careFixture(mode), 0)
		for pattern := 0; pattern < 1<<(2*frames); pattern++ {
			input := make([]bool, 2*frames)
			for i := range input {
				input[i] = pattern&(1<<i) != 0
			}
			if !replayCareFixture(mode, input) {
				continue
			}
			cex := util.Abc_CexAlloc(1, 2, frames)
			cex.IFrame = frames - 1
			for i, v := range input {
				if v {
					util.Abc_InfoSetBit(cex.PData, int32(1+i))
				}
			}
			reasons := Saig_ManCexMinComputeReason(graph, cex, 1)
			if reasons == nil {
				t.Fatalf("mode %d pattern %06b: no reasons", mode, pattern)
			}
			fixed := make([]bool, len(input))
			for frame := 0; frame < frames; frame++ {
				for _, lit := range vec.Vec_IntArray(vec.Vec_VecEntryInt(reasons, frame)) {
					idx := frame*2 + int(lit>>1)
					if idx < frame*2 || idx >= frame*2+2 {
						t.Fatalf("invalid literal %d", lit)
					}
					if (lit&1 == 0) != input[idx] {
						t.Fatalf("reason %d disagrees with trace", lit)
					}
					fixed[idx] = true
				}
			}
			for candidate := 0; candidate < 1<<(2*frames); candidate++ {
				trial := make([]bool, len(input))
				valid := true
				for i := range trial {
					trial[i] = candidate&(1<<i) != 0
					if fixed[i] && trial[i] != input[i] {
						valid = false
						break
					}
				}
				if valid && !replayCareFixture(mode, trial) {
					t.Fatalf("mode %d pattern %06b permits nonfailing %06b with reasons %v", mode, pattern, candidate, fixed)
				}
			}
		}
	}
}
