package gia

import (
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func Lf_ManTtIsMux(t word) int32 {
	masks := [3]uint32{0xaaaaaaaa, 0xcccccccc, 0xf0f0f0f0}
	for ctrl := 0; ctrl < 3; ctrl++ {
		var others [2]uint32
		k := 0
		for i := 0; i < 3; i++ {
			if i != ctrl {
				others[k] = masks[i]
				k++
			}
		}
		for phase := 0; phase < 8; phase++ {
			c, a, b := masks[ctrl], others[0], others[1]
			if phase&1 != 0 {
				c = ^c
			}
			if phase&2 != 0 {
				a = ^a
			}
			if phase&4 != 0 {
				b = ^b
			}
			if uint32(t) == (^c&a | c&b) {
				return 1
			}
		}
	}
	return 0
}
func Lf_ManFindCofVar(truth []word, nWords, nVars int32) int32 {
	tmp := make([]word, nWords)
	for i := int32(0); i < nVars; i++ {
		util.Abc_TtCofactor0p(tmp, truth, nWords, i)
		if util.Abc_TtSupportSize(tmp, nVars) < 2 {
			continue
		}
		util.Abc_TtCofactor1p(tmp, truth, nWords, i)
		if util.Abc_TtSupportSize(tmp, nVars) >= 2 {
			return i
		}
	}
	return -1
}
func lfComputeTruth(p *Lf_Man_t, inputs []*Lf_Cut_t, phases []int32, out *Lf_Cut_t, xor, mux bool) int32 {
	lut := p.pPars.NLutSize
	words := 1
	if lut > 6 {
		words = 1 << uint(lut-6)
	}
	old := out.nLeaves
	tables := make([][]word, len(inputs))
	for i, c := range inputs {
		table := make([]word, words)
		util.Abc_TtCopy(table, Lf_CutTruth(p, c), int32(words), (c.iFunc&1)^phases[i])
		util.Abc_TtExpand(table, lut, c.pLeaves[:], int32(c.nLeaves), out.pLeaves[:], int32(out.nLeaves))
		tables[i] = table
	}
	result := make([]word, words)
	for i := range result {
		if mux {
			result[i] = tables[2][i]&tables[1][i] | ^tables[2][i]&tables[0][i]
		} else if xor {
			result[i] = tables[0][i] ^ tables[1][i]
		} else {
			result[i] = tables[0][i] & tables[1][i]
		}
	}
	phase := int32(result[0] & 1)
	if phase != 0 {
		for i := range result {
			result[i] = ^result[i]
		}
	}
	out.nLeaves = uint32(util.Abc_TtMinBase(result, out.pLeaves[:], int32(out.nLeaves), lut))
	if out.nLeaves > old {
		panic("LF truth increased support")
	}
	out.iFunc = int32(vec.Vec_MemHashInsert(p.vTtMem, result))<<1 | phase
	if out.nLeaves < old {
		return 1
	}
	return 0
}
func Lf_CutComputeTruth6(p *Lf_Man_t, a, b *Lf_Cut_t, pa, pb int32, out *Lf_Cut_t, xor int32) int32 {
	return lfComputeTruth(p, []*Lf_Cut_t{a, b}, []int32{pa, pb}, out, xor != 0, false)
}
func Lf_CutComputeTruth(p *Lf_Man_t, a, b *Lf_Cut_t, pa, pb int32, out *Lf_Cut_t, xor int32) int32 {
	return Lf_CutComputeTruth6(p, a, b, pa, pb, out, xor)
}
func Lf_CutComputeTruthMux6(p *Lf_Man_t, a, b, c *Lf_Cut_t, pa, pb, pc int32, out *Lf_Cut_t) int32 {
	return lfComputeTruth(p, []*Lf_Cut_t{a, b, c}, []int32{pa, pb, pc}, out, false, true)
}
func Lf_CutComputeTruthMux(p *Lf_Man_t, a, b, c *Lf_Cut_t, pa, pb, pc int32, out *Lf_Cut_t) int32 {
	return Lf_CutComputeTruthMux6(p, a, b, c, pa, pb, pc, out)
}
