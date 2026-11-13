package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/glucose2"
	"time"
)

func cec4ManagerPacker(p *Cec4_Man_t) *cec4PatternPacker {
	if p.simState == nil || p.simState.sim == nil {
		panic("cec4: simulation is not allocated")
	}
	if p.packer == nil {
		p.packer = newCec4PatternPacker(int(gia.Gia_ManObjNum(p.pAig)), p.simState.sim.words)
		p.packer.values = p.simState.sim
	}
	return p.packer
}
func Cec4_ManPackAddPatterns(p *Cec4_Man_t, bit int32, literals *Vec_Int_t) {
	cec4ManagerPacker(p).extend(int(bit), vec.Vec_IntArray(literals))
}
func Cec4_ManPackAddPatternTry(p *Cec4_Man_t, bit int32, literals *Vec_Int_t) int32 {
	if cec4ManagerPacker(p).try(int(bit), vec.Vec_IntArray(literals)) {
		return 1
	}
	return 0
}
func Cec4_ManPackAddPattern(p *Cec4_Man_t, literals *Vec_Int_t, extend int32) int32 {
	pack := cec4ManagerPacker(p)
	pack.cursor = p.pattern
	count := pack.add(vec.Vec_IntArray(literals), extend != 0)
	p.pattern = pack.cursor
	return int32(count)
}
func Cec4_ManCandIterStart(p *Cec4_Man_t) { p.candidates = newCec4Candidates(p.pAig) }
func Cec4_ManCandIterNext(p *Cec4_Man_t) int32 {
	if p.candidates == nil {
		panic("cec4: candidate iteration not started")
	}
	return p.candidates.next()
}
func Cec4_ManGeneratePatterns(p *Cec4_Man_t) int32 {
	if p.candidates == nil {
		panic("cec4: candidate iteration not started")
	}
	start := time.Now()
	pack := cec4ManagerPacker(p)
	count, more := cec4GeneratePatterns(p.pAig, pack, p.candidates, p.pPars.nItersMax)
	p.timeGenPats += time.Since(start).Nanoseconds()
	p.nSatSat += int32(count)
	p.pattern = pack.cursor
	if more {
		return 1
	}
	return 0
}
func Cec4_ManFindRepr(graph *Gia_Man_t, p *Cec4_Man_t, id int32) *Gia_Obj_t {
	if p.simState == nil {
		panic("cec4: simulation is not allocated")
	}
	start := time.Now()
	cec4SimulateGraph(p.simState.sim, graph)
	repr := cec4FindRepr(p.simState.sim, graph, id)
	p.timeResimLoc += time.Since(start).Nanoseconds()
	if repr < 0 {
		return nil
	}
	return gia.Gia_ManObj(graph, repr)
}
func Cec4_ManSweepNode(p *Cec4_Man_t, id, repr int32) int32 {
	if p.pNew == nil || p.simState == nil {
		panic("cec4: sweep is not initialized")
	}
	if p.cnf == nil {
		p.cnf = newCec4CNF(p.pNew, p.pPars.jType)
		p.pSat = p.cnf.solver
	}
	if p.copies == nil {
		p.copies = make([]int32, gia.Gia_ManObjNum(p.pAig))
		for i := range p.copies {
			p.copies[i] = gia.Gia_ObjValue(gia.Gia_ManObj(p.pAig, int32(i)))
		}
	}
	p.nCallsSince++
	if p.pPars.nSatVarMax != 0 && p.nCallsSince > p.pPars.nCallsRecycle && int32(len(p.cnf.ids)) > p.pPars.nSatVarMax {
		p.cnf.recycle()
		p.pSat = p.cnf.solver
		p.nCallsSince = 0
		p.nRecycles++
	}
	limit := p.pPars.nBTLimit
	if vec.Vec_BitEntry(p.vFails, int(p.copies[id]>>1)) != 0 || vec.Vec_BitEntry(p.vFails, int(p.copies[repr]>>1)) != 0 {
		limit = max(int32(1), limit/10)
	}
	status := cec4SweepNode(p.pAig, p.pNew, p.simState.sim, p.cnf, p.copies, id, repr, limit, &p.pattern)
	switch status {
	case glucose2.GLUCOSE_SAT:
		p.nSatSat++
		p.nPatterns++
		if p.pattern == 64*p.simState.sim.words-2 {
			cec4SimulationRound(p.simState.sim, p.pAig, p.pPars.nLevelMax, true)
			p.pattern = 0
			if p.packer != nil {
				clear(p.packer.reserved.data)
			}
		}
		return 0
	case glucose2.GLUCOSE_UNSAT:
		p.nSatUnsat++
		if repr == 0 {
			p.iLastConst = id
		}
		return 1
	default:
		p.nSatUndec++
		vec.Vec_BitWriteEntry(p.vFails, int(p.copies[id]>>1), 1)
		return 2
	}
}
