package mpm

import (
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// Mpm_TruthStretch aligns an input cut's truth variables to the merged cut.
func Mpm_TruthStretch(truth []word, cut, input *Mpm_Cut_t, limit int32) {
	for i, k := int(cut.nLeaves)-1, int(input.nLeaves)-1; i >= 0 && k >= 0; i-- {
		if input.pLeaves[k] < cut.pLeaves[i] {
			continue
		}
		assert(input.pLeaves[k] == cut.pLeaves[i], "cut leaf ordering")
		if k < i {
			util.Abc_TtSwapVars(truth, limit, int32(k), int32(i))
		}
		k--
	}
}
func Mpm_CutTruthMinimize6(p *Mpm_Man_t, cut *Mpm_Cut_t) int32 {
	t := Mpm_CutTruth(p, util.Abc_Lit2Var(int32(cut.iFunc)))[0]
	var size int32
	support := util.Abc_Tt6SupportAndSize(t, int32(cut.nLeaves), &size)
	if size == int32(cut.nLeaves) {
		return 0
	}
	if size < 2 {
		p.nSmallSupp++
	}
	var k int32
	for i := int32(0); i < int32(cut.nLeaves); i++ {
		if support>>i&1 != 0 {
			if k < i {
				cut.pLeaves[k] = cut.pLeaves[i]
				tmp := []word{t}
				util.Abc_TtSwapVars(tmp, p.nLutSize, k, i)
				t = tmp[0]
			}
			k++
		}
	}
	assert(k == size, "support size")
	cut.nLeaves = uint32(size)
	assert(size == util.Abc_TtSupportSize([]word{t}, 6), "minimized support")
	cut.iFunc = uint32(util.Abc_Var2Lit(int32(vec.Vec_MemHashInsert(p.vTtMem, []word{t})), util.Abc_LitIsCompl(int32(cut.iFunc))))
	return 1
}
func Mpm_CutTruthMinimize7(p *Mpm_Man_t, cut *Mpm_Cut_t) int32 {
	truth := Mpm_CutTruth(p, util.Abc_Lit2Var(int32(cut.iFunc)))
	var size int32
	support := util.Abc_TtSupportAndSize(truth, int32(cut.nLeaves), &size)
	if size == int32(cut.nLeaves) {
		return 0
	}
	if size < 2 {
		p.nSmallSupp++
	}
	util.Abc_TtCopy(p.Truth[:], truth, p.nTruWords, 0)
	var k int32
	for i := int32(0); i < int32(cut.nLeaves); i++ {
		if support>>i&1 != 0 {
			if k < i {
				cut.pLeaves[k] = cut.pLeaves[i]
				util.Abc_TtSwapVars(p.Truth[:], p.nLutSize, k, i)
			}
			k++
		}
	}
	assert(k == size, "support size")
	assert(size == util.Abc_TtSupportSize(p.Truth[:], int32(cut.nLeaves)), "minimized support")
	cut.nLeaves = uint32(size)
	cut.iFunc = uint32(util.Abc_Var2Lit(int32(vec.Vec_MemHashInsert(p.vTtMem, p.Truth[:])), util.Abc_LitIsCompl(int32(cut.iFunc))))
	return 1
}
func Mpm_CutComputeTruth6(p *Mpm_Man_t, cut, a, b, c *Mpm_Cut_t, ca, cb, cc, typ int32) int32 {
	t0 := Mpm_CutTruth(p, util.Abc_Lit2Var(int32(a.iFunc)))[0]
	t1 := Mpm_CutTruth(p, util.Abc_Lit2Var(int32(b.iFunc)))[0]
	if ca^int32(a.fCompl)^util.Abc_LitIsCompl(int32(a.iFunc)) != 0 {
		t0 = ^t0
	}
	if cb^int32(b.fCompl)^util.Abc_LitIsCompl(int32(b.iFunc)) != 0 {
		t1 = ^t1
	}
	x, y := []word{t0}, []word{t1}
	Mpm_TruthStretch(x, cut, a, p.nLutSize)
	Mpm_TruthStretch(y, cut, b, p.nLutSize)
	var control word
	if c != nil {
		z := []word{Mpm_CutTruth(p, util.Abc_Lit2Var(int32(c.iFunc)))[0]}
		if cc^int32(c.fCompl)^util.Abc_LitIsCompl(int32(c.iFunc)) != 0 {
			z[0] = ^z[0]
		}
		Mpm_TruthStretch(z, cut, c, p.nLutSize)
		control = z[0]
	}
	assert(p.nLutSize <= 6, "LUT size")
	var t word
	switch typ {
	case 1:
		t = x[0] & y[0]
	case 2:
		t = x[0] ^ y[0]
	case 3:
		t = control&y[0] | ^control&x[0]
	default:
		panic("invalid cut type")
	}
	complement := int32(t & 1)
	if complement != 0 {
		t = ^t
	}
	cut.iFunc = uint32(util.Abc_Var2Lit(int32(vec.Vec_MemHashInsert(p.vTtMem, []word{t})), complement))
	if p.pPars.fCutMin != 0 {
		return Mpm_CutTruthMinimize6(p, cut)
	}
	return 1
}
func Mpm_CutComputeTruth7(p *Mpm_Man_t, cut, a, b, c *Mpm_Cut_t, ca, cb, cc, typ int32) int32 {
	util.Abc_TtCopy(p.Truth0[:], Mpm_CutTruth(p, util.Abc_Lit2Var(int32(a.iFunc))), p.nTruWords, ca^int32(a.fCompl)^util.Abc_LitIsCompl(int32(a.iFunc)))
	util.Abc_TtCopy(p.Truth1[:], Mpm_CutTruth(p, util.Abc_Lit2Var(int32(b.iFunc))), p.nTruWords, cb^int32(b.fCompl)^util.Abc_LitIsCompl(int32(b.iFunc)))
	Mpm_TruthStretch(p.Truth0[:], cut, a, p.nLutSize)
	Mpm_TruthStretch(p.Truth1[:], cut, b, p.nLutSize)
	if c != nil {
		util.Abc_TtCopy(p.TruthC[:], Mpm_CutTruth(p, util.Abc_Lit2Var(int32(c.iFunc))), p.nTruWords, cc^int32(c.fCompl)^util.Abc_LitIsCompl(int32(c.iFunc)))
		Mpm_TruthStretch(p.TruthC[:], cut, c, p.nLutSize)
	}
	switch typ {
	case 1:
		util.Abc_TtAnd(p.Truth[:], p.Truth0[:], p.Truth1[:], p.nTruWords, 0)
	case 2:
		util.Abc_TtXor(p.Truth[:], p.Truth0[:], p.Truth1[:], p.nTruWords, 0)
	case 3:
		util.Abc_TtMux(p.Truth[:], p.TruthC[:], p.Truth1[:], p.Truth0[:], p.nTruWords)
	default:
		panic("invalid cut type")
	}
	complement := int32(p.Truth[0] & 1)
	if complement != 0 {
		util.Abc_TtNot(p.Truth[:], p.nTruWords)
	}
	cut.iFunc = uint32(util.Abc_Var2Lit(int32(vec.Vec_MemHashInsert(p.vTtMem, p.Truth[:])), complement))
	if p.pPars.fCutMin != 0 {
		return Mpm_CutTruthMinimize7(p, cut)
	}
	return 1
}
func Mpm_CutComputeTruth(p *Mpm_Man_t, cut, a, b, c *Mpm_Cut_t, ca, cb, cc, typ int32) int32 {
	if p.nLutSize <= 6 {
		return Mpm_CutComputeTruth6(p, cut, a, b, c, ca, cb, cc, typ)
	}
	return Mpm_CutComputeTruth7(p, cut, a, b, c, ca, cb, cc, typ)
}
