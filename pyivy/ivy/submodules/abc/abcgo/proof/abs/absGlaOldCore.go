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
		if check != 0 && obj.fAbs != 0 {
			panic("old GLA object already abstracted")
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
	p.pSat.Simplify()
}
func Gia_GlaAddOneSlice(p *Gla_Man_t, current int32, core *vec.Vec_Int_t) {
	for frame := current; frame >= 0; frame-- {
		for _, id := range vec.Vec_IntArray(core) {
			Gla_ManAddClauses(p, id, frame, p.vTemp)
		}
	}
	p.pSat.Simplify()
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
	if v <= 0 {
		panic("old GLA output SAT variable is absent")
	}
	if frame == 0 && fanin.fRo != 0 && p.pObjRoot.fCompl0 == 0 {
		return -1
	}
	return v*2 + int32(p.pObjRoot.fCompl0)
}
func Gla_ManUnsatCore(p *Gla_Man_t, frame int32, solver *bsat.LabeledSolver2, confLimit, verbose int32, status, conflicts *int32) *vec.Vec_Int_t {
	_ = verbose
	if status != nil {
		*status = 1
	}
	lit := Gla_ManGetOutLit(p, frame)
	if lit == -1 {
		v := vec.Vec_IntAlloc(1)
		vec.Vec_IntPush(v, p.pObjRoot.Fanins[0])
		return v
	}
	before := solver.Statistics().Conflicts
	solver.SetLimits(int64(confLimit), 0)
	result := solver.Solve(bsat.Literal(lit))
	if conflicts != nil {
		*conflicts = int32(solver.Statistics().Conflicts - before)
	}
	if result == bsat.Unknown {
		if status != nil {
			*status = -1
		}
		return nil
	}
	if result == bsat.Satisfiable {
		if status != nil {
			*status = 0
		}
		return nil
	}
	ids := solver.ProofCore()
	core := vec.Vec_IntAlloc(len(ids))
	for _, id := range ids {
		vec.Vec_IntPush(core, id)
	}
	vec.Vec_IntSort(core, 1)
	return core
}
