package gia

import "github.com/glycerine/abcgo/misc/vec"

func Gia_ManDeriveCiTfo_rec(p *Gia_Man_t, obj *Gia_Obj_t, result *Vec_Int_t) int32 {
	if Gia_ObjIsCi(obj) != 0 || Gia_ObjIsTravIdCurrent(p, obj) != 0 {
		return int32(obj.fMark0)
	}
	Gia_ObjSetTravIdCurrent(p, obj)
	if Gia_ObjIsAnd(obj) == 0 {
		panic("gia: CI fanout traversal expected AND")
	}
	Gia_ManDeriveCiTfo_rec(p, Gia_ObjFanin0(obj), result)
	Gia_ManDeriveCiTfo_rec(p, Gia_ObjFanin1(obj), result)
	obj.fMark0 = Gia_ObjFanin0(obj).fMark0 | Gia_ObjFanin1(obj).fMark0
	if obj.fMark0 != 0 {
		vec.Vec_IntPush(result, obj.id)
	}
	return int32(obj.fMark0)
}
func Gia_ManDeriveCiTfoOne(p *Gia_Man_t, pivot *Gia_Obj_t) *Vec_Int_t {
	if pivot.fMark0 != 0 {
		panic("gia: CI already marked")
	}
	pivot.fMark0 = 1
	result := vec.Vec_IntAlloc(100)
	vec.Vec_IntPush(result, pivot.id)
	Gia_ManIncrementTravId(p)
	Gia_ObjSetTravIdCurrent(p, Gia_ManConst0(p))
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		obj := Gia_ManCo(p, i)
		Gia_ManDeriveCiTfo_rec(p, Gia_ObjFanin0(obj), result)
		if Gia_ObjFanin0(obj).fMark0 != 0 {
			vec.Vec_IntPush(result, obj.id)
		}
	}
	pivot.fMark0 = 0
	return result
}
func Gia_ManDeriveCiTfo(p *Gia_Man_t) *Vec_Vec_t {
	Gia_ManCleanMark0(p)
	Gia_ManIncrementTravId(p)
	result := vec.Vec_VecAlloc(int(Gia_ManCiNum(p)))
	for i := int32(0); i < Gia_ManCiNum(p); i++ {
		vec.Vec_PtrPush(result, Gia_ManDeriveCiTfoOne(p, Gia_ManCi(p, i)))
	}
	Gia_ManCleanMark0(p)
	return result
}
