package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/satoko"
)

type satoko_t = satoko.Solver

func Cec2_ObjSatId(p *gia.Gia_Man_t, obj *gia.Gia_Obj_t) int32 {
	return gia.Gia_ObjCopy2Array(p, gia.Gia_ObjId(p, gia.Gia_Regular(obj)))
}
func Cec2_ObjSetSatId(p *gia.Gia_Man_t, obj *gia.Gia_Obj_t, id int32) int32 {
	if Cec2_ObjSatId(p, obj) != -1 {
		panic("cec: SAT ID already assigned")
	}
	gia.Gia_ObjSetCopy2Array(p, gia.Gia_ObjId(p, gia.Gia_Regular(obj)), id)
	return id
}
func Cec2_ObjCleanSatId(p *gia.Gia_Man_t, obj *gia.Gia_Obj_t) {
	if Cec2_ObjSatId(p, obj) < 0 {
		panic("cec: SAT ID not assigned")
	}
	gia.Gia_ObjSetCopy2Array(p, gia.Gia_ObjId(p, gia.Gia_Regular(obj)), -1)
}

func Cec2_AddClausesMux(p *gia.Gia_Man_t, node *gia.Gia_Obj_t, s *satoko.Solver) {
	var t, e *gia.Gia_Obj_t
	control := gia.Gia_ObjRecognizeMux(node, &t, &e)
	lit := func(obj *gia.Gia_Obj_t) int {
		return int(Cec2_ObjSatId(p, gia.Gia_Regular(obj))*2 + gia.Gia_IsComplement(obj))
	}
	if !cec2AddMuxClauses(s, lit(node), lit(control), lit(t), lit(e)) {
		panic("cec: inconsistent MUX clauses")
	}
}
func Cec2_AddClausesSuper(p *gia.Gia_Man_t, node *gia.Gia_Obj_t, super *vec.Vec_Ptr_t, s *satoko.Solver) {
	lits := make([]int, vec.Vec_PtrSize(super))
	for i := range lits {
		obj := vec.Vec_PtrEntry(super, i).(*gia.Gia_Obj_t)
		lits[i] = int(2*Cec2_ObjSatId(p, gia.Gia_Regular(obj)) + gia.Gia_IsComplement(obj))
	}
	if !cec2AddSuperClauses(s, int(2*Cec2_ObjSatId(p, node)), lits) {
		panic("cec: inconsistent supergate clauses")
	}
}
func Cec2_CollectSuper_rec(p *gia.Gia_Man_t, obj *gia.Gia_Obj_t, super *vec.Vec_Ptr_t, first, useMuxes int32) {
	reg := gia.Gia_Regular(obj)
	id := gia.Gia_ObjId(p, reg)
	refs := gia.Gia_ObjValue(reg)
	if p.Refs != nil {
		refs = gia.Gia_ObjRefNumId(p, id)
	}
	if gia.Gia_IsComplement(obj) != 0 || gia.Gia_ObjIsCi(reg) != 0 || first == 0 && refs > 1 || useMuxes != 0 && gia.Gia_ObjMark0(reg) != 0 {
		vec.Vec_PtrPushUnique(super, obj)
		return
	}
	Cec2_CollectSuper_rec(p, gia.Gia_ObjChild0(reg), super, 0, useMuxes)
	Cec2_CollectSuper_rec(p, gia.Gia_ObjChild1(reg), super, 0, useMuxes)
}
func Cec2_CollectSuper(p *gia.Gia_Man_t, obj *gia.Gia_Obj_t, useMuxes int32, super *vec.Vec_Ptr_t) {
	if gia.Gia_IsComplement(obj) != 0 || gia.Gia_ObjIsCi(obj) != 0 {
		panic("cec: invalid supergate root")
	}
	vec.Vec_PtrClear(super)
	Cec2_CollectSuper_rec(p, obj, super, 1, useMuxes)
}
func Cec2_ObjAddToFrontier(p *gia.Gia_Man_t, obj *gia.Gia_Obj_t, frontier *vec.Vec_Ptr_t, s *satoko.Solver) {
	if Cec2_ObjSatId(p, obj) >= 0 {
		return
	}
	id := gia.Gia_ObjId(p, obj)
	if id == 0 {
		panic("cec: constant is not a CNF frontier")
	}
	v := s.AddVariable(false)
	gia.Gia_ManRecordSatVariable(p, v, id)
	Cec2_ObjSetSatId(p, obj, int32(v))
	if gia.Gia_ObjIsAnd(obj) != 0 {
		vec.Vec_PtrPush(frontier, obj)
	}
}
func Gia_ObjGetCnfVar(p *gia.Gia_Man_t, id int32, frontier, fanins *vec.Vec_Ptr_t, s *satoko.Solver) int32 {
	gia.Gia_ManExtendCopy2Array(p)
	obj := gia.Gia_ManObj(p, id)
	if v := Cec2_ObjSatId(p, obj); v >= 0 {
		return v
	}
	if id == 0 {
		panic("cec: constant requires a preallocated SAT variable")
	}
	if gia.Gia_ObjIsCi(obj) != 0 {
		v := s.AddVariable(false)
		gia.Gia_ManRecordSatVariable(p, v, id)
		return Cec2_ObjSetSatId(p, obj, int32(v))
	}
	if gia.Gia_ObjIsAnd(obj) == 0 {
		panic("cec: CNF root is not an AND or input")
	}
	vec.Vec_PtrClear(frontier)
	Cec2_ObjAddToFrontier(p, obj, frontier, s)
	for i := 0; i < vec.Vec_PtrSize(frontier); i++ {
		node := vec.Vec_PtrEntry(frontier, i).(*gia.Gia_Obj_t)
		if gia.Gia_ObjMark0(node) != 0 {
			vec.Vec_PtrClear(fanins)
			a, b := gia.Gia_ObjFanin0(node), gia.Gia_ObjFanin1(node)
			for _, f := range []*gia.Gia_Obj_t{gia.Gia_ObjFanin0(a), gia.Gia_ObjFanin0(b), gia.Gia_ObjFanin1(a), gia.Gia_ObjFanin1(b)} {
				vec.Vec_PtrPushUnique(fanins, f)
			}
			for j := 0; j < vec.Vec_PtrSize(fanins); j++ {
				Cec2_ObjAddToFrontier(p, gia.Gia_Regular(vec.Vec_PtrEntry(fanins, j).(*gia.Gia_Obj_t)), frontier, s)
			}
			Cec2_AddClausesMux(p, node, s)
		} else {
			Cec2_CollectSuper(p, node, 1, fanins)
			for j := 0; j < vec.Vec_PtrSize(fanins); j++ {
				Cec2_ObjAddToFrontier(p, gia.Gia_Regular(vec.Vec_PtrEntry(fanins, j).(*gia.Gia_Obj_t)), frontier, s)
			}
			Cec2_AddClausesSuper(p, node, fanins, s)
		}
	}
	return Cec2_ObjSatId(p, obj)
}
