package mpm

import (
	"github.com/glycerine/abcgo/misc/mem"
	"github.com/glycerine/abcgo/misc/util"
)

// Cut handles retain the C allocator's size and reuse behavior. The actual
// records live in Go-managed storage because their leaves cannot alias a C
// packed structure through Go's slice-backed word allocator.
func Mpm_CutAlloc(p *Mpm_Man_t, leaves int32, out **Mpm_Cut_t) int32 {
	h := mem.Mmr_StepFetch(p.pManCuts, Mpm_CutWordNum(leaves))
	if p.cuts == nil {
		p.cuts = make(map[int32]*Mpm_Cut_t)
	}
	cut := &Mpm_Cut_t{nLeaves: uint32(leaves), pLeaves: make([]int32, leaves)}
	p.cuts[h] = cut
	*out = cut
	return h
}
func Mpm_CutCreateZero(p *Mpm_Man_t) int32 {
	var cut *Mpm_Cut_t
	return Mpm_CutAlloc(p, 0, &cut)
}
func Mpm_CutCreateUnit(p *Mpm_Man_t, id int32) int32 {
	var cut *Mpm_Cut_t
	h := Mpm_CutAlloc(p, 1, &cut)
	cut.iFunc = uint32(util.Abc_Var2Lit(p.funcVar0, 0))
	cut.pLeaves[0] = util.Abc_Var2Lit(id, 0)
	return h
}
func Mpm_CutCreate(p *Mpm_Man_t, input *Mpm_Cut_t, out **Mpm_Cut_t) int32 {
	h := Mpm_CutAlloc(p, int32(input.nLeaves), out)
	(*out).iFunc = input.iFunc
	(*out).fCompl = input.fCompl
	(*out).fUseless = input.fUseless
	copy((*out).pLeaves, input.pLeaves[:input.nLeaves])
	return h
}
func Mpm_CutDup(p *Mpm_Man_t, input *Mpm_Cut_t, complement int32) int32 {
	var cut *Mpm_Cut_t
	h := Mpm_CutAlloc(p, int32(input.nLeaves), &cut)
	cut.iFunc = uint32(util.Abc_LitNotCond(int32(input.iFunc), complement))
	cut.fUseless = input.fUseless
	copy(cut.pLeaves, input.pLeaves[:input.nLeaves])
	return h
}
