package gia

import (
	"github.com/glycerine/abcgo/sat/bsat"
	"time"
)

func Gia_ManAddClausesMux(p *Swp_Man_t, node *Gia_Obj_t) {
	if Gia_IsComplement(node) != 0 || Gia_ObjIsMuxType(node) == 0 {
		panic("gia: invalid sweeper MUX")
	}
	var then, otherwise *Gia_Obj_t
	control := Gia_ObjRecognizeMux(node, &then, &otherwise)
	f := Swp_ManLit2Lit(p, Gia_Obj2Lit(p.pGia, node))
	i := Swp_ManLit2Lit(p, Gia_Obj2Lit(p.pGia, control))
	t := Swp_ManLit2Lit(p, Gia_Obj2Lit(p.pGia, then))
	e := Swp_ManLit2Lit(p, Gia_Obj2Lit(p.pGia, otherwise))
	clauses := [][]int32{{i ^ 1, t ^ 1, f}, {i ^ 1, t, f ^ 1}, {i, e ^ 1, f}, {i, e, f ^ 1}}
	if t != e {
		clauses = append(clauses, []int32{t, e, f ^ 1}, []int32{t ^ 1, e ^ 1, f})
	}
	for _, clause := range clauses {
		if !p.pSat.AddClause(clause...) {
			panic("gia: inconsistent sweeper MUX")
		}
	}
}
func Gia_ManAddClausesSuper(p *Swp_Man_t, node *Gia_Obj_t, super *Vec_Int_t) {
	if Gia_IsComplement(node) != 0 || Gia_ObjIsAnd(node) == 0 {
		panic("gia: invalid sweeper supergate")
	}
	f := Swp_ManLit2Lit(p, Gia_Obj2Lit(p.pGia, node))
	for i, lit := range Vec_IntArray(super) {
		mapped := Swp_ManLit2Lit(p, lit)
		if !p.pSat.AddClause(mapped, f^1) {
			panic("gia: inconsistent sweeper supergate")
		}
		Vec_IntWriteEntry(super, i, mapped^1)
	}
	Vec_IntPush(super, f)
	if !p.pSat.AddClause(Vec_IntArray(super)...) {
		panic("gia: inconsistent sweeper supergate")
	}
}
func Gia_ManCollectSuper_rec(p *Gia_Man_t, obj *Gia_Obj_t, super *Vec_Int_t) {
	if Gia_IsComplement(obj) != 0 || obj.fMark1 != 0 || Gia_ObjIsCi(obj) != 0 || Gia_ObjIsMuxType(obj) != 0 {
		Vec_IntPushUnique(super, Gia_Obj2Lit(p, obj))
		return
	}
	Gia_ManCollectSuper_rec(p, Gia_ObjChild0(obj), super)
	Gia_ManCollectSuper_rec(p, Gia_ObjChild1(obj), super)
}
func Gia_ManCollectSuper(p *Gia_Man_t, obj *Gia_Obj_t, super *Vec_Int_t) {
	if Gia_IsComplement(obj) != 0 || Gia_ObjIsAnd(obj) == 0 {
		panic("gia: invalid sweeper supergate root")
	}
	Vec_IntClear(super)
	Gia_ManCollectSuper_rec(p, Gia_ObjChild0(obj), super)
	Gia_ManCollectSuper_rec(p, Gia_ObjChild1(obj), super)
}
func Gia_ManObjAddToFrontier(p *Swp_Man_t, id int32, front *Vec_Int_t) {
	if id == 0 || Swp_ManObj2Lit(p, id) != 0 {
		return
	}
	obj := Gia_ManObj(p.pGia, id)
	v := p.pSat.NewVariable()
	if v != p.nSatVars {
		panic("gia: SAT variable allocation out of sync")
	}
	Swp_ManSetObj2Lit(p, id, v<<1|int32(obj.fPhase))
	p.nSatVars++
	if Gia_ObjIsAnd(obj) != 0 {
		Vec_IntPush(front, id)
	}
}
func Gia_ManCnfNodeAddToSolver(p *Swp_Man_t, id int32) {
	if id == 0 || Swp_ManObj2Lit(p, id) != 0 {
		return
	}
	start := time.Now()
	defer func() { p.timeCnf += time.Since(start) }()
	Vec_IntClear(p.vFront)
	Gia_ManObjAddToFrontier(p, id, p.vFront)
	for i := 0; i < Vec_IntSize(p.vFront); i++ {
		node := Gia_ManObj(p.pGia, Vec_IntEntry(p.vFront, i))
		if Gia_ObjIsMuxType(node) != 0 {
			Vec_IntClear(p.vFanins)
			for _, child := range []*Gia_Obj_t{Gia_ObjFanin0(node), Gia_ObjFanin1(node)} {
				Vec_IntPushUnique(p.vFanins, Gia_ObjFaninId0p(p.pGia, child))
				Vec_IntPushUnique(p.vFanins, Gia_ObjFaninId1p(p.pGia, child))
			}
			for _, fid := range Vec_IntArray(p.vFanins) {
				Gia_ManObjAddToFrontier(p, fid, p.vFront)
			}
			Gia_ManAddClausesMux(p, node)
		} else {
			Gia_ManCollectSuper(p.pGia, node, p.vFanins)
			for _, lit := range Vec_IntArray(p.vFanins) {
				Gia_ManObjAddToFrontier(p, lit>>1, p.vFront)
			}
			Gia_ManAddClausesSuper(p, node, p.vFanins)
		}
		if Vec_IntSize(p.vFanins) <= 1 {
			panic("gia: invalid sweeper fanins")
		}
	}
}
func Gia_ManGetCex(graph *Gia_Man_t, mapping *Vec_Int_t, solver *bsat.Solver, cex *Vec_Int_t) *Vec_Int_t {
	Vec_IntClear(cex)
	for i := int32(0); i < Gia_ManPiNum(graph); i++ {
		id := Gia_ManPi(graph, i).id
		if int(id) >= Vec_IntSize(mapping) || Vec_IntEntry(mapping, int(id)) == 0 {
			Vec_IntPush(cex, 2)
			continue
		}
		lit := Vec_IntEntry(mapping, int(id))
		value, ok := solver.Value(lit >> 1)
		if !ok {
			panic("gia: missing SAT model")
		}
		Vec_IntPush(cex, giaBool(value)^lit&1)
	}
	return cex
}
