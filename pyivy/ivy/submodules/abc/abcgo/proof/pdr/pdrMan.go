// Counterexample abstraction refinement, ported from pdrMan.c.
package pdr

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bmc"
)

func Pdr_ManDeriveCexAbs(p *Pdr_Man_t) *util.Abc_Cex_t {
	if p.pPars.fUseAbs == 0 || p.vMapPpi2Ff == nil {
		return Pdr_ManDeriveCex(p)
	}
	nPis := saig.Saig_ManPiNum(p.pAig)
	for i, flop := range vec.Vec_IntArray(p.vMapPpi2Ff) {
		if vec.Vec_IntEntry(p.vMapFf2Ppi, int(flop)) != int32(i) {
			panic("pdr: inconsistent abstraction maps")
		}
		vec.Vec_IntWriteEntry(p.vMapFf2Ppi, int(flop), -1)
	}
	vec.Vec_IntClear(p.vMapPpi2Ff)
	var frames int32
	for entry := p.pQueue; entry != nil; entry = entry.pNext {
		for _, lit := range entry.pState.Lits[entry.pState.nLits:entry.pState.nTotal] {
			flop := (lit >> 1) - nPis
			if flop < 0 || vec.Vec_IntEntry(p.vMapFf2Ppi, int(flop)) >= 0 {
				continue
			}
			vec.Vec_IntWriteEntry(p.vMapFf2Ppi, int(flop), int32(vec.Vec_IntSize(p.vMapPpi2Ff)))
			vec.Vec_IntPush(p.vMapPpi2Ff, flop)
		}
		frames++
	}
	nPpis := int32(vec.Vec_IntSize(p.vMapPpi2Ff))
	if nPpis == 0 {
		return Pdr_ManDeriveCex(p)
	}
	var refined int32
	if p.pPars.fUseSimpleRef != 0 {
		for _, flop := range vec.Vec_IntArray(p.vMapPpi2Ff) {
			if vec.Vec_IntEntry(p.vAbsFlops, int(flop)) != 0 {
				panic("pdr: expected abstract flop")
			}
			vec.Vec_IntWriteEntry(p.vAbsFlops, int(flop), 1)
			refined++
		}
	} else {
		cex := util.Abc_CexAlloc(aig.Aig_ManRegNum(p.pAig)-nPpis, nPis+nPpis, frames)
		cex.IPo, cex.IFrame = p.iOutCur, frames-1
		var frame int32
		for entry := p.pQueue; entry != nil; entry = entry.pNext {
			for _, lit := range entry.pState.Lits[entry.pState.nLits:entry.pState.nTotal] {
				if lit&1 != 0 {
					continue
				}
				input := lit >> 1
				if input >= nPis {
					input = nPis + vec.Vec_IntEntry(p.vMapFf2Ppi, int(input-nPis))
				}
				if input >= cex.NPis {
					panic("pdr: abstract input outside counterexample")
				}
				util.Abc_InfoSetBit(cex.PData, cex.NRegs+frame*cex.NPis+input)
			}
			frame++
		}
		abs := gia.Gia_ManDupAbs(p.pGia, p.vMapPpi2Ff, p.vMapFf2Ppi)
		care := bmc.Bmc_CexCareMinimizeAig(abs, nPis, cex, 1, 0, 0)
		if care.NPis != cex.NPis {
			panic("pdr: care counterexample input count mismatch")
		}
		for frame := int32(0); frame < frames; frame++ {
			for input := nPis; input < care.NPis; input++ {
				if !util.Abc_InfoHasBit(care.PData, care.NRegs+care.NPis*frame+input) {
					continue
				}
				flop := vec.Vec_IntEntry(p.vMapPpi2Ff, int(input-nPis))
				if vec.Vec_IntEntry(p.vAbsFlops, int(flop)) == 0 {
					vec.Vec_IntWriteEntry(p.vAbsFlops, int(flop), 1)
					refined++
				}
			}
		}
		if refined == 0 {
			return Pdr_ManDeriveCex(p)
		}
	}
	p.nCexesTotal++
	p.nCexes++
	return nil
}
