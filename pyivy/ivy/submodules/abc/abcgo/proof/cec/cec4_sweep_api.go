package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/glucose2"
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
	pack := cec4ManagerPacker(p)
	// The source checks the manager cursor, even when bit is different.
	if p.pattern <= 0 || p.pattern >= 64*pack.values.words {
		panic("cec4: pattern cursor out of range")
	}
	if pack.try(int(bit), vec.Vec_IntArray(literals)) {
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
	unpin := cecPinClockThread()
	defer unpin()
	if p.candidates == nil {
		panic("cec4: candidate iteration not started")
	}
	start := cecCPUClock()
	pack := cec4ManagerPacker(p)
	count, more := cec4GeneratePatterns(p.pAig, pack, p.candidates, p.pPars.NItersMax)
	p.timeGenPats += cecCPUClock() - start
	p.nSatSat += int32(count)
	p.pattern = pack.cursor
	if more {
		return 1
	}
	return 0
}
func Cec4_ManFindRepr(graph *Gia_Man_t, p *Cec4_Man_t, id int32) *Gia_Obj_t {
	unpin := cecPinClockThread()
	defer unpin()
	if p.simState == nil {
		panic("cec4: simulation is not allocated")
	}
	start := cecCPUClock()
	repr := cec4FindReprAtPattern(p.simState.sim, graph, id, int32(p.pattern), vec.Vec_IntArray(p.vCexStamps))
	p.timeResimLoc += cecCPUClock() - start
	if repr < 0 {
		return nil
	}
	return gia.Gia_ManObj(graph, repr)
}
func Cec4_ManSweepNode(p *Cec4_Man_t, id, repr int32) int32 {
	unpin := cecPinClockThread()
	defer unpin()
	if p.pNew == nil || p.simState == nil {
		panic("cec4: sweep is not initialized")
	}
	if p.cnf == nil {
		p.cnf = newCec4CNF(p.pNew, p.pPars.JType)
		p.pSat = p.cnf.solver
	}
	if p.copies == nil {
		p.copies = make([]int32, gia.Gia_ManObjNum(p.pAig))
		for i := range p.copies {
			p.copies[i] = gia.Gia_ObjValue(gia.Gia_ManObj(p.pAig, int32(i)))
		}
	}
	// The reduced graph is built incrementally between calls.
	p.copies[id] = gia.Gia_ObjValue(gia.Gia_ManObj(p.pAig, id))
	p.copies[repr] = gia.Gia_ObjValue(gia.Gia_ManObj(p.pAig, repr))
	start := cecCPUClock()
	obj := gia.Gia_ManObj(p.pAig, id)
	representative := gia.Gia_ManObj(p.pAig, repr)
	phase := (p.copies[id]^p.copies[repr])&1 ^ gia.Gia_ObjPhase(obj) ^ gia.Gia_ObjPhase(representative)
	var easy int32
	status := Cec4_ManSolveTwo(p, p.copies[repr]>>1, p.copies[id]>>1, phase, &easy, p.pPars.FVerbose)
	cec4ApplySweepResult(p.pAig, p.pNew, p.simState.sim, p.cnf, p.copies, id, repr, &p.pattern, status)
	elapsed := cecCPUClock() - start

	switch status {
	case glucose2.GLUCOSE_SAT:
		vec.Vec_IntClear(p.vPat)
		if p.pPars.JType == 0 {
			// Source iterates the reduced graph's CI/SAT-ID pairs in allocation order.
			for variable, object := range p.cnf.objects {
				if mapped, ok := p.cnf.inputs[object]; ok && int(mapped) == variable {
					value := int32(0)
					if p.cnf.solver.Model[variable] == glucose2.LTrue {
						value = 1
					}
					vec.Vec_IntPush(p.vPat, object<<1|value)
				}
			}
		} else {
			// Glucose's justification model contains a count followed by SAT
			// literals; ABC maps their complements back to GIA object IDs.
			for _, lit := range p.cnf.solver.JustModel[1:] {
				v := int(lit >> 1)
				if v < len(p.cnf.objects) && p.cnf.objects[v] >= 0 {
					vec.Vec_IntPush(p.vPat, p.cnf.objects[v]<<1|(int32(lit)^1)&1)
				}
			}
		}
		p.nSatSat++
		p.nPatterns++
		if easy != 0 {
			p.timeSatSat0 += elapsed
		} else {
			p.timeSatSat += elapsed
		}
		if p.pattern == 64*p.simState.sim.words-2 {
			cec4SimulationRound(p.simState.sim, p.pAig, p.pPars.NLevelMax, true)
			vec.Vec_IntFill(p.vCexStamps, int(gia.Gia_ManObjNum(p.pAig)), 0)
			p.pattern = 0
			if p.packer != nil {
				clear(p.packer.reserved.data)
			}
		}
		return 0
	case glucose2.GLUCOSE_UNSAT:
		// The public source helper uses the relation between reduced copies;
		// its caller may subsequently replace this with the original phase.
		p.copies[id] = p.copies[repr] ^ phase
		gia.Gia_ObjSetValue(obj, p.copies[id])
		p.nSatUnsat++
		if easy != 0 {
			p.timeSatUnsat0 += elapsed
		} else {
			p.timeSatUnsat += elapsed
		}
		if repr == 0 {
			p.iLastConst = id
		}
		return 1
	default:
		p.nSatUndec++
		p.timeSatUndec += elapsed
		vec.Vec_BitSetEntry(p.vFails, int(id), 1)
		return 2
	}
}
