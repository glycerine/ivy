package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_ManBfsForCrossCut(p *Gia_Man_t) *Vec_Int_t {
	result := Vec_IntAlloc(int(Gia_ManObjNum(p)))
	levels := Gia_ManLevelize(p)
	for i := 0; i < vec.Vec_VecSize(levels); i++ {
		for _, entry := range Vec_PtrArray(vec.Vec_VecEntry(levels, i)) {
			Vec_IntPush(result, Gia_ObjId(p, entry.(*Gia_Obj_t)))
		}
	}
	return result
}
func Gia_ManPrint(p *Gia_Man_t) {
	fmt.Printf("GIA manager has %d ANDs, %d XORs, %d MUXes.\n", Gia_ManAndNum(p)-Gia_ManXorNum(p)-Gia_ManMuxNum(p), Gia_ManXorNum(p), Gia_ManMuxNum(p))
	for _, obj := range p.pObjs {
		Gia_ObjPrint(p, obj)
	}
}
func Gia_ManPrintCo_rec(p *Gia_Man_t, obj *Gia_Obj_t) {
	if Gia_ObjIsAnd(obj) != 0 {
		Gia_ManPrintCo_rec(p, Gia_ObjFanin0(obj))
		Gia_ManPrintCo_rec(p, Gia_ObjFanin1(obj))
		if Gia_ObjIsMux(p, obj) != 0 {
			Gia_ManPrintCo_rec(p, Gia_ObjFanin2(p, obj))
		}
	}
	Gia_ObjPrint(p, obj)
}
func Gia_ManPrintCo(p *Gia_Man_t, obj *Gia_Obj_t) {
	if Gia_ObjIsCo(obj) == 0 {
		panic("gia: expected CO")
	}
	fmt.Printf("TFI cone of CO number %d:\n", Gia_ObjCioId(obj))
	Gia_ManPrintCo_rec(p, Gia_ObjFanin0(obj))
	Gia_ObjPrint(p, obj)
}
func Gia_ManPrintCollect_rec(p *Gia_Man_t, obj *Gia_Obj_t, nodes *Vec_Int_t) {
	if Vec_IntFind(nodes, obj.id) >= 0 {
		return
	}
	if Gia_ObjIsAnd(obj) == 0 {
		panic("gia: missing cone leaf")
	}
	Gia_ManPrintCollect_rec(p, Gia_ObjFanin0(obj), nodes)
	Gia_ManPrintCollect_rec(p, Gia_ObjFanin1(obj), nodes)
	if Gia_ObjIsMux(p, obj) != 0 {
		Gia_ManPrintCollect_rec(p, Gia_ObjFanin2(p, obj), nodes)
	}
	Vec_IntPush(nodes, obj.id)
}
func Gia_ManPrintCone(p *Gia_Man_t, obj *Gia_Obj_t, leaves []int32, count int32, nodes *Vec_Int_t) {
	Vec_IntClear(nodes)
	for _, id := range leaves[:count] {
		Vec_IntPush(nodes, id)
	}
	Gia_ManPrintCollect_rec(p, obj, nodes)
	fmt.Printf("GIA logic cone for node %d:\n", obj.id)
	for _, id := range Vec_IntArray(nodes) {
		Gia_ObjPrint(p, Gia_ManObj(p, id))
	}
}
func Gia_ManPrintConeMulti(p *Gia_Man_t, objects, leaves, nodes *Vec_Int_t) {
	Vec_IntClear(nodes)
	Vec_IntAppend(nodes, leaves)
	for _, id := range Vec_IntArray(objects) {
		Gia_ManPrintCollect_rec(p, Gia_ManObj(p, id), nodes)
	}
	fmt.Printf("GIA logic cone for %d nodes:\n", Vec_IntSize(objects))
	for _, id := range Vec_IntArray(nodes) {
		Gia_ObjPrint(p, Gia_ManObj(p, id))
	}
}
func Gia_ManPrintCollect2_rec(p *Gia_Man_t, obj *Gia_Obj_t, nodes *Vec_Int_t) {
	if Vec_IntFind(nodes, obj.id) >= 0 {
		return
	}
	if Gia_ObjIsCo(obj) != 0 || Gia_ObjIsAnd(obj) != 0 {
		Gia_ManPrintCollect2_rec(p, Gia_ObjFanin0(obj), nodes)
	}
	if Gia_ObjIsAnd(obj) != 0 {
		Gia_ManPrintCollect2_rec(p, Gia_ObjFanin1(obj), nodes)
	}
	if Gia_ObjIsMux(p, obj) != 0 {
		Gia_ManPrintCollect2_rec(p, Gia_ObjFanin2(p, obj), nodes)
	}
	Vec_IntPush(nodes, obj.id)
}
func Gia_ManPrintCone2(p *Gia_Man_t, obj *Gia_Obj_t) {
	nodes := Vec_IntAlloc(100)
	Gia_ManPrintCollect2_rec(p, obj, nodes)
	fmt.Printf("GIA logic cone for node %d:\n", obj.id)
	for _, id := range Vec_IntArray(nodes) {
		Gia_ObjPrint(p, Gia_ManObj(p, id))
	}
}
