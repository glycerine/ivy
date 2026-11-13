package bmc

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
)

func TestCexCareMinimizeCompletions(t *testing.T) {
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
					util.Abc_InfoSetBit(cex.PData, int32(i+1))
				}
			}
			for real := int32(0); real <= 2; real++ {
				for tries := int32(1); tries <= 4; tries++ {
					care := Bmc_CexCareMinimizeAig(g, real, cex, tries, 0, 0)
					if care == nil || Bmc_CexVerify(g, cex, care) != 1 {
						t.Fatalf("mode=%d pattern=%b real=%d tries=%d: invalid care", mode, pattern, real, tries)
					}
					for candidate := 0; candidate < 1<<(2*frames); candidate++ {
						trial, valid := make([]bool, len(input)), true
						for i := range trial {
							trial[i] = candidate&(1<<i) != 0
							if util.Abc_InfoHasBit(care.PData, int32(i+1)) && trial[i] != input[i] {
								valid = false
								break
							}
						}
						if valid && !replayCareFixture(mode, trial) {
							t.Fatalf("mode=%d pattern=%b care=%v permits %b", mode, pattern, care.PData, candidate)
						}
					}
				}
			}
		}
	}
}

func TestCexCareMinimizeInvalid(t *testing.T) {
	g := gia.Gia_ManStart(8)
	x := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, x)
	cex := util.Abc_CexAlloc(0, 1, 1)
	if Bmc_CexCareMinimizeAig(g, 1, cex, 1, 0, 0) != nil {
		t.Fatal("accepted nonfailing trace")
	}
	util.Abc_InfoSetBit(cex.PData, 0)
	for _, tries := range []int32{0, 5} {
		if Bmc_CexCareMinimizeAig(g, 1, cex, tries, 0, 0) != nil {
			t.Fatal("accepted invalid try count")
		}
	}
	if got := Bmc_CexCareMinimize(gia.Gia_ManToAig(g, 0), 1, cex, 1, 0, 0); got == nil || !util.Abc_InfoHasBit(got.PData, 0) {
		t.Fatal("AIG wrapper lost required input")
	}
}

func TestCexCareMinimizeNativeXorMux(t *testing.T) {
	g := gia.Gia_ManStart(24)
	g.EnableMuxes()
	a, b, selectInput, r := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x := gia.Gia_ManAppendXorReal(g, a, b)
	m := gia.Gia_ManAppendMuxReal(g, selectInput, x, r)
	gia.Gia_ManAppendCo(g, m)
	gia.Gia_ManAppendCo(g, x)
	gia.Gia_ManSetRegNum(g, 1)
	for pattern := 0; pattern < 64; pattern++ {
		input := make([]bool, 6)
		for i := range input {
			input[i] = pattern&(1<<i) != 0
		}
		replay := func(v []bool) bool {
			reg := false
			for f := 0; f < 2; f++ {
				xor := v[3*f] != v[3*f+1]
				out := reg
				if v[3*f+2] {
					out = xor
				}
				if f == 1 {
					return out
				}
				reg = xor
			}
			return false
		}
		if !replay(input) {
			continue
		}
		cex := util.Abc_CexAlloc(1, 3, 2)
		cex.IFrame = 1
		for i, value := range input {
			if value {
				util.Abc_InfoSetBit(cex.PData, int32(i+1))
			}
		}
		care := Bmc_CexCareMinimizeAig(g, 2, cex, 4, 0, 0)
		if care == nil || Bmc_CexVerify(g, cex, care) == 0 {
			t.Fatalf("pattern %06b: invalid care", pattern)
		}
		for completion := 0; completion < 64; completion++ {
			v, matches := make([]bool, 6), true
			for i := range v {
				v[i] = completion&(1<<i) != 0
				if util.Abc_InfoHasBit(care.PData, int32(i+1)) && v[i] != input[i] {
					matches = false
					break
				}
			}
			if matches && !replay(v) {
				t.Fatalf("pattern %06b permits %06b", pattern, completion)
			}
		}
	}
}
