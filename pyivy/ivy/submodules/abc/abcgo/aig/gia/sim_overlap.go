package gia

import "github.com/glycerine/abcgo/misc/vec"

// Gia_ManObjCheckSat_rec follows the original greedy implication search.
// It does not explore both branches of a complemented AND.
func Gia_ManObjCheckSat_rec(p *Gia_Man_t, lit int32, visited *Vec_Int_t) int32 {
	id := lit >> 1
	obj := Gia_ManObj(p, id)
	phase := uint32(lit & 1)
	if obj.fMark0 != 0 {
		return giaBool(obj.fMark1 == phase)
	}
	obj.fMark0 = 1
	obj.fMark1 = phase
	vec.Vec_IntPush(visited, id)
	if Gia_ObjIsAnd(obj) != 0 {
		child0 := 2*Gia_ObjFaninId0(obj, id) + Gia_ObjFaninC0(obj)
		if phase == 0 {
			child1 := 2*Gia_ObjFaninId1(obj, id) + Gia_ObjFaninC1(obj)
			if Gia_ManObjCheckSat_rec(p, child0, visited) == 0 || Gia_ManObjCheckSat_rec(p, child1, visited) == 0 {
				return 0
			}
		} else if Gia_ManObjCheckSat_rec(p, child0^1, visited) == 0 {
			return 0
		}
	}
	return 1
}
func Gia_ManObjCheckOverlap1(p *Gia_Man_t, a, b int32, visited *Vec_Int_t) int32 {
	vec.Vec_IntClear(visited)
	defer func() {
		for _, id := range vec.Vec_IntArray(visited) {
			obj := Gia_ManObj(p, id)
			obj.fMark0 = 0
			obj.fMark1 = 0
		}
	}()
	if Gia_ManObjCheckSat_rec(p, a, visited) == 0 {
		return 0
	}
	return Gia_ManObjCheckSat_rec(p, b, visited)
}
func Gia_ManObjCheckOverlap(p *Gia_Man_t, a, b int32, visited *Vec_Int_t) int32 {
	if Gia_ManObjCheckOverlap1(p, a, b, visited) != 0 {
		return 1
	}
	return Gia_ManObjCheckOverlap1(p, b, a, visited)
}
