package bsat

import (
	"reflect"
	"slices"
)

/**CFile****************************************************************

  FileName    [satProof2.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT solver.]

  Synopsis    [Proof logging.]

  Author      [Alan Mishchenko <alanmi@eecs.berkeley.edu>]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - January 1, 2004.]

  Revision    [$Id: satProof2.h,v 1.0 2004/01/01 1:00:00 alanmi Exp $]

***********************************************************************/

////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

// Prf_Man_t records each learned clause's root-clause dependencies as a bitset.
// vId2Pr is a borrowed problem-ID mapping; a negative entry excludes that root.
type Prf_Man_t struct {
	iFirst, iFirst2, nWords int32
	pInfo                   []word
	vInfo                   []word
	vSaved                  []int32
	vId2Pr                  *[]int32
}

func Prf_BitWordNum(width int32) int32 {
	if width < 0 {
		panic("negative proof width")
	}
	n := width / 64
	if width%64 != 0 {
		n++
	}
	return n
}
func Prf_ManSize(p *Prf_Man_t) int32 {
	if p.nWords <= 0 {
		if len(p.vInfo) == 0 {
			return 0
		}
		panic("proof width not initialized")
	}
	return int32(len(p.vInfo)) / p.nWords
}
func Prf_ManClearNewInfo(p *Prf_Man_t) {
	if p.nWords <= 0 {
		panic("proof width not initialized")
	}
	p.vInfo = append(p.vInfo, make([]word, p.nWords)...)
}
func Prf_ManClauseInfo(p *Prf_Man_t, id int32) []word {
	if id < 0 || id >= Prf_ManSize(p) {
		panic("proof clause index outside manager")
	}
	start := id * p.nWords
	return p.vInfo[start : start+p.nWords]
}
func Prf_ManAlloc() *Prf_Man_t {
	return &Prf_Man_t{iFirst: -1, iFirst2: -1, vInfo: make([]word, 0, 1000), vSaved: make([]int32, 0, 1000)}
}
func Prf_ManStop(p *Prf_Man_t) {
	if p != nil {
		*p = Prf_Man_t{}
	}
}
func Prf_ManStopP(p **Prf_Man_t) { Prf_ManStop(*p); *p = nil }
func Prf_ManMemory(p *Prf_Man_t) float64 {
	return float64(reflect.TypeOf(*p).Size()) + float64(cap(p.vInfo))*8 + float64(cap(p.vSaved))*4
}
func Prf_ManRestart(p *Prf_Man_t, idToProof *[]int32, first, width int32) {
	if p.iFirst != -1 || first < 0 || width <= 0 {
		panic("invalid proof restart")
	}
	p.iFirst = first
	p.nWords = Prf_BitWordNum(width)
	p.vId2Pr = idToProof
	p.pInfo = nil
	p.vInfo = p.vInfo[:0]
}
func Prf_ManGrow(p *Prf_Man_t, width int32) {
	if p.iFirst < 0 || p.pInfo != nil {
		panic("grow during proof chain")
	}
	if int64(width) < 64*int64(p.nWords) {
		return
	}
	words := max(Prf_BitWordNum(width), 2*p.nWords)
	rows := Prf_ManSize(p)
	// ABC reserves room for 1000 additional learned clauses on each growth.
	next := make([]word, rows*words, (rows+1000)*words)
	for i := int32(0); i < rows; i++ {
		copy(next[i*words:i*words+p.nWords], Prf_ManClauseInfo(p, i))
	}
	p.vInfo = next
	p.nWords = words
}
func Prf_ManShrink(p *Prf_Man_t, id int32) {
	row := id - p.iFirst
	if p.iFirst < 0 || row < 0 || row >= Prf_ManSize(p) || p.pInfo != nil {
		panic("invalid proof shrink")
	}
	p.vInfo = p.vInfo[:row*p.nWords]
}
func Prf_ManAddSaved(p *Prf_Man_t, id, newID int32) {
	if p.iFirst < 0 {
		panic("proof not initialized")
	}
	if id < p.iFirst {
		return
	}
	if len(p.vSaved) == 0 {
		if p.iFirst2 != -1 {
			panic("stale compacted proof base")
		}
		p.iFirst2 = newID
	}
	p.vSaved = append(p.vSaved, id)
}
func Prf_ManCompact(p *Prf_Man_t, newID int32) {
	if p.iFirst < 0 || p.pInfo != nil {
		panic("compact during proof chain")
	}
	for i, id := range p.vSaved {
		copy(p.vInfo[int32(i)*p.nWords:], Prf_ManClauseInfo(p, id-p.iFirst))
	}
	p.vInfo = p.vInfo[:int32(len(p.vSaved))*p.nWords]
	p.vSaved = p.vSaved[:0]
	p.iFirst = newID
	if p.iFirst2 != -1 {
		p.iFirst = p.iFirst2
	}
	p.iFirst2 = -1
}
func Prf_ManChainResolve(p *Prf_Man_t, c *clause) {
	if p.iFirst < 0 || p.pInfo == nil {
		panic("resolve outside proof chain")
	}
	id := clause_id(c)
	if c.lrn != 0 {
		if id >= p.iFirst {
			source := Prf_ManClauseInfo(p, id-p.iFirst)
			for i, value := range source {
				p.pInfo[i] |= value
			}
		}
		return
	}
	if id < 0 {
		return
	}
	entry := id
	if p.vId2Pr != nil {
		entry = (*p.vId2Pr)[id]
	}
	if entry < 0 {
		return
	}
	if int64(entry) >= 64*int64(p.nWords) {
		panic("problem proof ID outside width")
	}
	p.pInfo[entry/64] |= uint64(1) << uint(entry%64)
}
func Prf_ManChainStart(p *Prf_Man_t, c *clause) {
	if p.iFirst < 0 || p.pInfo != nil {
		panic("nested proof chain")
	}
	Prf_ManClearNewInfo(p)
	p.pInfo = Prf_ManClauseInfo(p, Prf_ManSize(p)-1)
	Prf_ManChainResolve(p, c)
}
func Prf_ManChainStop(p *Prf_Man_t) int32 {
	if p.pInfo == nil {
		panic("stop outside proof chain")
	}
	p.pInfo = nil
	return 0
}
func Prf_ManUnsatCore(p *Prf_Man_t) []int32 {
	if p.iFirst < 0 || p.pInfo != nil || Prf_ManSize(p) == 0 {
		panic("proof core unavailable")
	}
	info := Prf_ManClauseInfo(p, Prf_ManSize(p)-1)
	var core []int32
	if p.vId2Pr == nil {
		for i := int32(0); i < 64*p.nWords; i++ {
			if info[i/64]&(uint64(1)<<uint(i%64)) != 0 {
				core = append(core, i)
			}
		}
	} else {
		for id, entry := range *p.vId2Pr {
			if entry < 0 {
				continue
			}
			if int64(entry) >= 64*int64(p.nWords) {
				panic("mapped proof ID outside width")
			}
			if info[entry/64]&(uint64(1)<<uint(entry%64)) != 0 {
				core = append(core, int32(id))
			}
		}
	}
	slices.Sort(core)
	slices.Reverse(core)
	return core
}
