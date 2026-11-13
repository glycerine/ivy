package mpm

import (
	"slices"

	"github.com/glycerine/abcgo/misc/vec"
)

// Mpm_CutFindLeaf compares object IDs while ignoring literal complement.
func Mpm_CutFindLeaf(cut *Mpm_Cut_t, id int32) int32 {
	for i, lit := range cut.pLeaves[:cut.nLeaves] {
		if lit>>1 == id {
			return int32(i)
		}
	}
	return int32(cut.nLeaves)
}

// Mpm_ManMergeCuts uses the source's free cut unit and leaves it in the pool;
// the insertion routine consumes the unit only after accepting this cut.
func Mpm_ManMergeCuts(p *Mpm_Man_t, first, second, third *Mpm_Cut_t) *Mpm_Cut_t {
	cut := &vec.Vec_PtrEntryLast(&p.vFreeUnits).(*Mpm_Uni_t).pCut
	cut.pLeaves = append(cut.pLeaves[:0], first.pLeaves[:first.nLeaves]...)
	cut.nLeaves = first.nLeaves
	for c, other := range []*Mpm_Cut_t{second, third} {
		if other == nil {
			break
		}
		slot := c + 1
		if p.pPars.fUseDsd != 0 {
			p.uPermMask[slot] = 0x3ffff
			p.uComplMask[slot] = 0
		}
		for i, lit := range other.pLeaves[:other.nLeaves] {
			place := Mpm_CutFindLeaf(cut, lit>>1)
			if place == int32(cut.nLeaves) {
				if int32(cut.nLeaves) == p.nLutSize {
					return nil
				}
				cut.pLeaves = append(cut.pLeaves, lit)
				cut.nLeaves++
			}
			if p.pPars.fUseDsd != 0 {
				p.uPermMask[slot] ^= uint32((i&7)^7) << uint(3*place)
				if lit != cut.pLeaves[place] {
					p.uComplMask[slot] |= 1 << uint(place)
				}
			}
		}
	}
	cut.hNext = 0
	if second == nil {
		cut.iFunc = first.iFunc
		cut.fUseless = first.fUseless
		cut.fCompl = first.fCompl
	} else {
		cut.iFunc = (1 << 25) - 1
		cut.fUseless = 0
		cut.fCompl = 0
	}
	p.nCutsMerged++
	p.nCutsMergedAll++
	if p.pPars.fUseTruth != 0 {
		slices.Sort(cut.pLeaves[:cut.nLeaves])
	}
	return cut
}
