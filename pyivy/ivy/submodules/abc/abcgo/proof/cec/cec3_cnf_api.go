package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/glucose"
)

type bmcg_sat_solver = glucose.Solver

func Cec3_ObjSatId(p *gia.Gia_Man_t, obj *gia.Gia_Obj_t) int32 { return Cec2_ObjSatId(p, obj) }
func Cec3_ObjSetSatId(p *gia.Gia_Man_t, obj *gia.Gia_Obj_t, v int32) int32 {
	return Cec2_ObjSetSatId(p, obj, v)
}
func Cec3_ObjCleanSatId(p *gia.Gia_Man_t, obj *gia.Gia_Obj_t) { Cec2_ObjCleanSatId(p, obj) }

func Cec3_AddClausesMux(p *gia.Gia_Man_t, node *gia.Gia_Obj_t, s *glucose.Solver) {
	var t, e *gia.Gia_Obj_t
	control := gia.Gia_ObjRecognizeMux(node, &t, &e)
	lit := func(obj *gia.Gia_Obj_t) int32 {
		return 2*Cec3_ObjSatId(p, gia.Gia_Regular(obj)) + gia.Gia_IsComplement(obj)
	}
	out, i, a, b := lit(node), lit(control), lit(t), lit(e)
	clauses := [][]int32{{i ^ 1, a ^ 1, out}, {i ^ 1, a, out ^ 1}, {i, b ^ 1, out}, {i, b, out ^ 1}}
	if a>>1 != b>>1 {
		clauses = append(clauses, []int32{a, b, out ^ 1}, []int32{a ^ 1, b ^ 1, out})
	}
	for _, c := range clauses {
		if glucose.Bmcg_sat_solver_addclause(s, c, int32(len(c))) == 0 {
			panic("cec: inconsistent Glucose MUX clauses")
		}
	}
}
func Cec3_AddClausesSuper(p *gia.Gia_Man_t, node *gia.Gia_Obj_t, super *vec.Vec_Ptr_t, s *glucose.Solver) {
	out := 2 * Cec3_ObjSatId(p, node)
	reverse := make([]int32, vec.Vec_PtrSize(super)+1)
	for i := 0; i < vec.Vec_PtrSize(super); i++ {
		obj := vec.Vec_PtrEntry(super, i).(*gia.Gia_Obj_t)
		lit := 2*Cec3_ObjSatId(p, gia.Gia_Regular(obj)) + gia.Gia_IsComplement(obj)
		if glucose.Bmcg_sat_solver_addclause(s, []int32{lit, out ^ 1}, 2) == 0 {
			panic("cec: inconsistent Glucose AND clauses")
		}
		reverse[i] = lit ^ 1
	}
	reverse[len(reverse)-1] = out
	if glucose.Bmcg_sat_solver_addclause(s, reverse, int32(len(reverse))) == 0 {
		panic("cec: inconsistent Glucose AND clauses")
	}
}
func Cec3_CollectSuper_rec(obj *gia.Gia_Obj_t, super *vec.Vec_Ptr_t, first, useMuxes int32) {
	reg := gia.Gia_Regular(obj)
	if gia.Gia_IsComplement(obj) != 0 || gia.Gia_ObjIsCi(reg) != 0 || first == 0 && gia.Gia_ObjValue(reg) > 1 || useMuxes != 0 && gia.Gia_ObjMark0(reg) != 0 {
		vec.Vec_PtrPushUnique(super, obj)
		return
	}
	Cec3_CollectSuper_rec(gia.Gia_ObjChild0(reg), super, 0, useMuxes)
	Cec3_CollectSuper_rec(gia.Gia_ObjChild1(reg), super, 0, useMuxes)
}
func Cec3_CollectSuper(obj *gia.Gia_Obj_t, useMuxes int32, super *vec.Vec_Ptr_t) {
	if gia.Gia_IsComplement(obj) != 0 || gia.Gia_ObjIsCi(obj) != 0 {
		panic("cec: invalid supergate root")
	}
	vec.Vec_PtrClear(super)
	Cec3_CollectSuper_rec(obj, super, 1, useMuxes)
}
func Cec3_ObjAddToFrontier(p *gia.Gia_Man_t, obj *gia.Gia_Obj_t, frontier *vec.Vec_Ptr_t, s *glucose.Solver) {
	if Cec3_ObjSatId(p, obj) >= 0 {
		return
	}
	if gia.Gia_ObjId(p, obj) == 0 {
		panic("cec: constant is not a CNF frontier")
	}
	Cec3_ObjSetSatId(p, obj, glucose.Bmcg_sat_solver_addvar(s))
	if gia.Gia_ObjIsAnd(obj) != 0 {
		vec.Vec_PtrPush(frontier, obj)
	}
}
func Cec3_ObjGetCnfVar(p *Cec3_Man_t, id int32) int32 {
	g, s := p.pNew, p.pSat
	gia.Gia_ManExtendCopy2Array(g)
	obj := gia.Gia_ManObj(g, id)
	if v := Cec3_ObjSatId(g, obj); v >= 0 {
		return v
	}
	if id == 0 {
		panic("cec: constant requires a preallocated SAT variable")
	}
	if gia.Gia_ObjIsCi(obj) != 0 {
		return Cec3_ObjSetSatId(g, obj, glucose.Bmcg_sat_solver_addvar(s))
	}
	if gia.Gia_ObjIsAnd(obj) == 0 {
		panic("cec: CNF root is not an AND or input")
	}
	vec.Vec_PtrClear(p.vFrontier)
	Cec3_ObjAddToFrontier(g, obj, p.vFrontier, s)
	for i := 0; i < vec.Vec_PtrSize(p.vFrontier); i++ {
		node := vec.Vec_PtrEntry(p.vFrontier, i).(*gia.Gia_Obj_t)
		if gia.Gia_ObjMark0(node) != 0 {
			vec.Vec_PtrClear(p.vFanins)
			a, b := gia.Gia_ObjFanin0(node), gia.Gia_ObjFanin1(node)
			for _, f := range []*gia.Gia_Obj_t{gia.Gia_ObjFanin0(a), gia.Gia_ObjFanin0(b), gia.Gia_ObjFanin1(a), gia.Gia_ObjFanin1(b)} {
				vec.Vec_PtrPushUnique(p.vFanins, f)
			}
			for j := 0; j < vec.Vec_PtrSize(p.vFanins); j++ {
				Cec3_ObjAddToFrontier(g, gia.Gia_Regular(vec.Vec_PtrEntry(p.vFanins, j).(*gia.Gia_Obj_t)), p.vFrontier, s)
			}
			Cec3_AddClausesMux(g, node, s)
		} else {
			Cec3_CollectSuper(node, 1, p.vFanins)
			for j := 0; j < vec.Vec_PtrSize(p.vFanins); j++ {
				Cec3_ObjAddToFrontier(g, gia.Gia_Regular(vec.Vec_PtrEntry(p.vFanins, j).(*gia.Gia_Obj_t)), p.vFrontier, s)
			}
			Cec3_AddClausesSuper(g, node, p.vFanins, s)
		}
	}
	return Cec3_ObjSatId(g, obj)
}
