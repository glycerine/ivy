package cec

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/satoko"
	"time"
)

func Cec2_ManVerify_rec(g *gia.Gia_Man_t, id int32, s *satoko.Solver) int32 {
	if id == 0 {
		return 0
	}
	obj := gia.Gia_ManObj(g, id)
	if gia.Gia_ObjIsTravIdCurrentId(g, id) != 0 {
		return gia.Gia_ObjMark1(obj)
	}
	gia.Gia_ObjSetTravIdCurrentId(g, id)
	var value int32
	if gia.Gia_ObjIsCi(obj) != 0 {
		if s.ModelValue(int(Cec2_ObjSatId(g, obj))) == 1 {
			value = 1
		}
	} else {
		a := Cec2_ManVerify_rec(g, gia.Gia_ObjFaninId0(obj, id), s) ^ gia.Gia_ObjFaninC0(obj)
		b := Cec2_ManVerify_rec(g, gia.Gia_ObjFaninId1(obj, id), s) ^ gia.Gia_ObjFaninC1(obj)
		value = a & b
	}
	gia.Gia_ObjSetMark1(obj, value)
	return value
}
func Cec2_ManVerify(g *gia.Gia_Man_t, a, b, phase int32, s *satoko.Solver) {
	gia.Gia_ManIncrementTravId(g)
	if Cec2_ManVerify_rec(g, a, s)^Cec2_ManVerify_rec(g, b, s) == phase {
		fmt.Printf("CEX verification FAILED for obj %d and obj %d.\n", a, b)
	}
}
func Cec2_ManCollect_rec(p *Cec2_Man_t, id int32) {
	g := p.pNew
	if gia.Gia_ObjIsTravIdCurrentId(g, id) != 0 {
		return
	}
	gia.Gia_ObjSetTravIdCurrentId(g, id)
	obj := gia.Gia_ManObj(g, id)
	satID := Cec2_ObjSatId(g, obj)
	if satID >= 0 {
		vec.Vec_IntPush(p.vNodesNew, id)
		vec.Vec_IntPush(p.vSatVars, satID)
	}
	if id == 0 {
		return
	}
	if gia.Gia_ObjIsAnd(obj) != 0 {
		Cec2_ManCollect_rec(p, gia.Gia_ObjFaninId0(obj, id))
		Cec2_ManCollect_rec(p, gia.Gia_ObjFaninId1(obj, id))
	} else {
		if satID < 0 {
			panic("cec: cone input has no SAT variable")
		}
		vec.Vec_IntPushTwo(p.vObjSatPairs, gia.Gia_ManCiIdToId(p.pAig, gia.Gia_ObjCioId(obj)), satID)
	}
}
func Cec2_ManSolveTwo(p *Cec2_Man_t, a, b, phase int32) int32 {
	if b < a {
		a, b = b, a
	}
	if a >= b || phase < 0 || phase > 1 {
		panic("cec: invalid SAT comparison")
	}
	if p.pPars.fUseCones == 0 && p.pSat.VariableCount() != 0 {
		panic("cec: expected an empty SAT solver")
	}
	if a == 0 && Cec2_ObjSatId(p.pNew, gia.Gia_ManConst0(p.pNew)) < 0 {
		Cec2_ObjSetSatId(p.pNew, gia.Gia_ManConst0(p.pNew), int32(p.pSat.AddVariable(false)))
	}
	av, bv := Cec2_ObjGetCnfVar(p, a), Cec2_ObjGetCnfVar(p, b)
	vec.Vec_IntClear(p.vNodesNew)
	vec.Vec_IntClear(p.vSatVars)
	vec.Vec_IntClear(p.vObjSatPairs)
	gia.Gia_ManIncrementTravId(p.pNew)
	Cec2_ManCollect_rec(p, a)
	Cec2_ManCollect_rec(p, b)
	variables := make([]int, vec.Vec_IntSize(p.vSatVars))
	for i, v := range vec.Vec_IntArray(p.vSatVars) {
		variables[i] = int(v)
	}
	if p.pPars.fUseCones != 0 {
		p.pSat.MarkCone(variables, true)
	}
	status := p.pSat.SolveAssumptionsLimit([]int{int(2*av + 1), int(2*bv + phase)}, int64(p.pPars.nConfLimit))
	if status == satoko.SATOKO_UNSAT && a > 0 {
		status = p.pSat.SolveAssumptionsLimit([]int{int(2 * av), int(2*bv + (phase ^ 1))}, int64(p.pPars.nConfLimit))
	}
	if p.pPars.fUseCones != 0 {
		p.pSat.MarkCone(variables, false)
	} else {
		for _, id := range vec.Vec_IntArray(p.vNodesNew) {
			Cec2_ObjCleanSatId(p.pNew, gia.Gia_ManObj(p.pNew, id))
		}
	}
	return int32(status)
}
func Cec2_ManSweepNode(p *Cec2_Man_t, id int32) int32 {
	start := time.Now()
	obj := gia.Gia_ManObj(p.pAig, id)
	repr := gia.Gia_ObjReprObj(p.pAig, id)
	phase := (gia.Gia_ObjValue(obj)^gia.Gia_ObjValue(repr))&1 ^ gia.Gia_ObjPhase(obj) ^ gia.Gia_ObjPhase(repr)
	status := Cec2_ManSolveTwo(p, gia.Gia_ObjValue(repr)>>1, gia.Gia_ObjValue(obj)>>1, phase)
	result := int32(1)
	switch status {
	case satoko.SATOKO_SAT:
		p.nSatSat++
		p.nPatterns++
		pattern := gia.Gia_ManCecSimPattern(p.pAig) + 1
		if pattern == 64*p.pAig.SimulationWordCount() {
			pattern = 1
		}
		gia.Gia_ManCecSetSimPattern(p.pAig, pattern)
		pairs := vec.Vec_IntArray(p.vObjSatPairs)
		for i := 0; i+1 < len(pairs); i += 2 {
			value := int32(0)
			if p.pSat.ModelValue(int(pairs[i+1])) == 1 {
				value = 1
			}
			Cec2_ObjSimSetInputBit(p.pAig, pairs[i], value)
		}
		p.timeSatSat += time.Since(start).Nanoseconds()
		result = 0
	case satoko.SATOKO_UNSAT:
		p.nSatUnsat++
		gia.Gia_ObjSetValue(obj, gia.Gia_ObjValue(repr)^phase)
		gia.Gia_ObjSetProved(p.pAig, id)
		p.timeSatUnsat += time.Since(start).Nanoseconds()
	default:
		p.nSatUndec++
		gia.Gia_ObjSetFailed(p.pAig, id)
		p.timeSatUndec += time.Since(start).Nanoseconds()
		result = 2
	}
	if p.pPars.fUseCones == 0 {
		start = time.Now()
		p.pSat.Rollback()
		p.pSat.Stats().Conflicts = 0
		p.timeExtra += time.Since(start).Nanoseconds()
	}
	return result
}
func Cec2_ManPrintStats(g *gia.Gia_Man_t, pars *Cec2_Par_t, p *Cec2_Man_t) {
	if pars.fVerbose == 0 {
		return
	}
	var sat, unsat, undec int32
	if p != nil {
		sat, unsat, undec = p.nSatSat, p.nSatUnsat, p.nSatUndec
	}
	fmt.Printf("S =%5d U =%5d F =%5d ", sat, unsat, undec)
	gia.Gia_ManEquivPrintClasses(g, pars.fVeryVerbose, 0)
}
