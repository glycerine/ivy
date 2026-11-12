package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
)

func Sat_ObjXValue(obj *Gia_Obj_t) int32 { return int32(obj.fMark1<<1 | obj.fMark0) }
func Sat_ObjSetXValue(obj *Gia_Obj_t, value int32) {
	obj.fMark0 = uint32(value) & 1
	obj.fMark1 = (uint32(value) >> 1) & 1
}

// Gia_SatCollectCone_rec appends the cone in fanin-first order and marks it
// unknown for ternary simulation.
func Gia_SatCollectCone_rec(p *Gia_Man_t, obj *Gia_Obj_t, visit *Vec_Int_t) {
	if Sat_ObjXValue(obj) == GIA_UND {
		return
	}
	if Gia_ObjIsAnd(obj) != 0 {
		Gia_SatCollectCone_rec(p, Gia_ObjFanin0(obj), visit)
		Gia_SatCollectCone_rec(p, Gia_ObjFanin1(obj), visit)
	}
	if Sat_ObjXValue(obj) != 0 {
		panic("gia: pattern cone marks are not clean")
	}
	Sat_ObjSetXValue(obj, GIA_UND)
	vec.Vec_IntPush(visit, Gia_ObjId(p, obj))
}
func Gia_SatCollectCone(p *Gia_Man_t, obj *Gia_Obj_t, visit *Vec_Int_t) {
	if Gia_IsComplement(obj) != 0 || Gia_ObjIsConst0(obj) != 0 || Sat_ObjXValue(obj) != 0 {
		panic("gia: invalid pattern cone root")
	}
	vec.Vec_IntClear(visit)
	Gia_SatCollectCone_rec(p, obj, visit)
}

// Gia_SatVerifyPattern checks whether a partial CI assignment asserts root.
// All marks set by the check are cleared, including CIs outside its cone.
func Gia_SatVerifyPattern(p *Gia_Man_t, root *Gia_Obj_t, cex, visit *Vec_Int_t) {
	if Gia_ObjIsCo(root) == 0 || Gia_ObjIsConst0(Gia_ObjFanin0(root)) != 0 {
		panic("gia: invalid pattern output")
	}
	Gia_SatCollectCone(p, Gia_ObjFanin0(root), visit)
	defer func() {
		for _, id := range vec.Vec_IntArray(visit) {
			Sat_ObjSetXValue(Gia_ManObj(p, id), 0)
		}
		for _, lit := range vec.Vec_IntArray(cex) {
			Sat_ObjSetXValue(Gia_ManCi(p, lit>>1), 0)
		}
	}()
	for _, lit := range vec.Vec_IntArray(cex) {
		value := int32(GIA_ONE)
		if lit&1 != 0 {
			value = GIA_ZER
		}
		Sat_ObjSetXValue(Gia_ManCi(p, lit>>1), value)
	}
	for _, id := range vec.Vec_IntArray(visit) {
		obj := Gia_ManObj(p, id)
		if Gia_ObjIsCi(obj) != 0 {
			continue
		}
		if Gia_ObjIsAnd(obj) == 0 {
			panic("gia: invalid pattern cone object")
		}
		value := Gia_XsimAndCond(Sat_ObjXValue(Gia_ObjFanin0(obj)), Gia_ObjFaninC0(obj), Sat_ObjXValue(Gia_ObjFanin1(obj)), Gia_ObjFaninC1(obj))
		Sat_ObjSetXValue(obj, value)
	}
	value := Gia_XsimNotCond(Sat_ObjXValue(Gia_ObjFanin0(root)), Gia_ObjFaninC0(root))
	if value != GIA_ONE {
		fmt.Println("Gia_SatVerifyPattern(): Verification FAILED.")
	}
}
