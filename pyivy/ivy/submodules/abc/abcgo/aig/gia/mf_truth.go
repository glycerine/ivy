package gia

import (
	"github.com/glycerine/abcgo/bool/kit"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func mfCutTruth(p *Mf_Man_t, inputs []*Mf_Cut_t, complements []int32, out *Mf_Cut_t, operation byte) int32 {
	lut := p.pPars.NLutSize
	words := 1
	if lut > 6 {
		words = 1 << uint(lut-6)
	}
	expanded := make([][]uint64, len(inputs))
	for i, c := range inputs {
		t := append([]uint64(nil), vec.Vec_MemReadEntry(p.vTtMem, int(c.iFunc>>1))...)
		if int32(c.iFunc&1)^complements[i] != 0 {
			for k := range t {
				t[k] = ^t[k]
			}
		}
		util.Abc_TtExpand(t, max(6, lut), c.pLeaves[:], int32(c.nLeaves), out.pLeaves[:], int32(out.nLeaves))
		expanded[i] = t
	}
	truth := make([]uint64, words)
	for i := range truth {
		switch operation {
		case '&':
			truth[i] = expanded[0][i] & expanded[1][i]
		case '^':
			truth[i] = expanded[0][i] ^ expanded[1][i]
		case 'm':
			truth[i] = (expanded[2][i] & expanded[1][i]) | (^expanded[2][i] & expanded[0][i])
		}
	}
	complement := uint32(truth[0] & 1)
	if complement != 0 {
		for i := range truth {
			truth[i] = ^truth[i]
		}
	}
	old := out.nLeaves
	if lut > 6 || operation == 'm' || p.pPars.fCnfObjIds == 0 {
		out.nLeaves = uint32(util.Abc_TtMinBase(truth, out.pLeaves[:], int32(out.nLeaves), max(6, lut)))
	}
	id := vec.Vec_MemHashInsert(p.vTtMem, truth)
	out.iFunc = uint32(id<<1) | complement
	if (p.pPars.FGenCnf != 0 || p.pPars.FGenLit != 0) && id == vec.Vec_IntSize(&p.vCnfSizes) && lut <= 8 {
		var size int32
		if p.pPars.FGenCnf != 0 {
			if lut <= 6 {
				size = util.Abc_Tt6CnfSize(truth[0], int32(out.nLeaves))
			} else {
				size = util.Abc_Tt8CnfSize(truth, int32(out.nLeaves))
			}
		} else {
			words32 := make([]uint32, len(truth)*2)
			for i, v := range truth {
				words32[2*i], words32[2*i+1] = uint32(v), uint32(v>>32)
			}
			var scratch []uint32
			size = kit.Kit_TruthLitNum(words32, int32(out.nLeaves), &scratch)
		}
		vec.Vec_IntPush(&p.vCnfSizes, size)
	}
	return giaBool(out.nLeaves < old)
}
func Mf_CutComputeTruth6(p *Mf_Man_t, a, b *Mf_Cut_t, ca, cb int32, out *Mf_Cut_t, xor int32) int32 {
	op := byte('&')
	if xor != 0 {
		op = '^'
	}
	return mfCutTruth(p, []*Mf_Cut_t{a, b}, []int32{ca, cb}, out, op)
}
func Mf_CutComputeTruth(p *Mf_Man_t, a, b *Mf_Cut_t, ca, cb int32, out *Mf_Cut_t, xor int32) int32 {
	return Mf_CutComputeTruth6(p, a, b, ca, cb, out, xor)
}
func Mf_CutComputeTruthMux6(p *Mf_Man_t, a, b, c *Mf_Cut_t, ca, cb, cc int32, out *Mf_Cut_t) int32 {
	return mfCutTruth(p, []*Mf_Cut_t{a, b, c}, []int32{ca, cb, cc}, out, 'm')
}
func Mf_CutComputeTruthMux(p *Mf_Man_t, a, b, c *Mf_Cut_t, ca, cb, cc int32, out *Mf_Cut_t) int32 {
	return Mf_CutComputeTruthMux6(p, a, b, c, ca, cb, cc, out)
}
