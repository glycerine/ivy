package abs

import (
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

func Gia_GlaAddToCounters(p *Gla_Man_t, core *vec.Vec_Int_t) {
	for _, id := range vec.Vec_IntArray(core) {
		vec.Vec_IntAddToEntry(p.vCoreCounts, int(Gla_ManObj(p, id).iGiaObj), 1)
	}
}
func Gia_GlaAddToAbs(p *Gla_Man_t, add *vec.Vec_Int_t, check int32) {
	write := 0
	for _, id := range vec.Vec_IntArray(add) {
		obj := Gla_ManObj(p, id)
		if check != 0 {
			if bsat.Sat_solver2_Prf2(p.pSat) != nil {
				p.vProofIds[Gla_ObjId(p, obj)] = p.nProofIds
				p.nProofIds++
			}
		}
		if obj.fAbs != 0 {
			continue
		}
		obj.fAbs = 1
		vec.Vec_IntPush(p.vAbs, id)
		vec.Vec_IntWriteEntry(add, write, id)
		write++
	}
	vec.Vec_IntShrink(add, write)
}
func Gia_GlaAddTimeFrame(p *Gla_Man_t, frame int32) {
	for _, id := range vec.Vec_IntArray(p.vAbs) {
		Gla_ManAddClauses(p, id, frame, p.vTemp)
	}
	bsat.Sat_solver2_simplify(p.pSat)
}
func Gia_GlaAddOneSlice(p *Gla_Man_t, current int32, core *vec.Vec_Int_t) {
	for frame := current; frame >= 0; frame-- {
		for _, id := range vec.Vec_IntArray(core) {
			Gla_ManAddClauses(p, id, frame, p.vTemp)
		}
	}
	bsat.Sat_solver2_simplify(p.pSat)
}
func Gla_ManRollBack(p *Gla_Man_t) {
	added := vec.Vec_IntArray(p.vAddedNew)
	if len(added)%2 != 0 {
		panic("invalid old GLA rollback list")
	}
	for i := 0; i < len(added); i += 2 {
		obj := Gla_ManObj(p, added[i])
		if vec.Vec_IntEntry(&obj.vFrames, int(added[i+1])) <= 0 {
			panic("old GLA rollback variable is absent")
		}
		vec.Vec_IntWriteEntry(&obj.vFrames, int(added[i+1]), 0)
	}
	for _, id := range vec.Vec_IntArray(p.vAbs)[p.nAbsOld:] {
		obj := Gla_ManObj(p, id)
		if obj.fAbs == 0 {
			panic("old GLA rollback object is absent")
		}
		obj.fAbs = 0
	}
	vec.Vec_IntShrink(p.vAbs, int(p.nAbsOld))
}
func Gla_ManGetOutLit(p *Gla_Man_t, frame int32) int32 {
	fanin := Gla_ManObj(p, p.pObjRoot.Fanins[0])
	v := vec.Vec_IntEntry(&fanin.vFrames, int(frame))
	if frame == 0 && fanin.fRo != 0 && p.pObjRoot.fCompl0 == 0 {
		return -1
	}
	return v*2 + int32(p.pObjRoot.fCompl0)
}
func Gla_ManUnsatCore(p *Gla_Man_t, f int32, pSat *bsat.Sat_solver2, nConfMax, fVerbose int32, piRetValue, pnConfls *int32) *vec.Vec_Int_t {
	var vCore *vec.Vec_Int_t
	nConfPrev := bsat.Sat_solver2_nconflicts(pSat)
	iLit := Gla_ManGetOutLit(p, f)
	if piRetValue != nil {
		*piRetValue = 1
	}
	// consider special case when PO points to the flop
	// this leads to immediate conflict in the first timeframe
	if iLit == -1 {
		vCore = vec.Vec_IntAlloc(1)
		vec.Vec_IntPush(vCore, p.pObjRoot.Fanins[0])
		return vCore
	}
	// solve the problem
	RetValue := bsat.Sat_solver2_solve(pSat, []int32{iLit}, int64(nConfMax), 0, 0, 0)
	if pnConfls != nil {
		*pnConfls = bsat.Sat_solver2_nconflicts(pSat) - nConfPrev
	}
	if RetValue == int32(bsat.Unknown) {
		if piRetValue != nil {
			*piRetValue = -1
		}
		return nil
	}
	if RetValue == int32(bsat.Satisfiable) {
		if piRetValue != nil {
			*piRetValue = 0
		}
		return nil
	}
	// derive the UNSAT core
	if core := bsat.Sat_ProofCore(pSat); core != nil {
		vCore = vec.Vec_IntAllocArrayCopy(core, len(core))
		vec.Vec_IntSort(vCore, 1)
	}
	return vCore
}
