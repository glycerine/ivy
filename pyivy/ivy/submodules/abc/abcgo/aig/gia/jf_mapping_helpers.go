package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
)

func Jf_CutPrint(cut []int32) {
	fmt.Printf("%d {", Jf_CutSize(cut))
	for i := int32(1); i <= Jf_CutSize(cut); i++ {
		fmt.Printf(" %d", cut[i])
	}
	fmt.Printf(" } Func = %d\n", Jf_CutFunc(cut))
}
func Jf_ObjCutPrint(cuts []int32) {
	offset := 1
	for i := int32(0); i < cuts[0]; i++ {
		cut := cuts[offset:]
		Jf_CutPrint(cut)
		offset += int(Jf_CutSize(cut)) + 1
	}
	fmt.Println()
}
func Jf_CutArr(p *Jf_Man_t, cut []int32) int32 {
	var arrival int32
	for i := int32(1); i <= Jf_CutSize(cut); i++ {
		arrival = max(arrival, Jf_ObjArr(p, Jf_CutVar(cut, i)))
	}
	return arrival + 1
}
func Jf_CutRef(p *Jf_Man_t, cut []int32) {
	for i := int32(1); i <= Jf_CutSize(cut); i++ {
		Gia_ObjRefIncId(p.pGia, Jf_CutVar(cut, i))
	}
}
func Jf_CutDeref(p *Jf_Man_t, cut []int32) {
	for i := int32(1); i <= Jf_CutSize(cut); i++ {
		Gia_ObjRefDecId(p.pGia, Jf_CutVar(cut, i))
	}
}
func Jf_CutFlow(p *Jf_Man_t, cut []int32) float32 {
	var flow float32
	for i := int32(1); i <= Jf_CutSize(cut); i++ {
		flow += Jf_ObjFlow(p, Jf_CutVar(cut, i))
	}
	if flow < 0 {
		panic("JF negative cut flow")
	}
	return flow
}
func Jf_ObjCutFilterBoth(p *Jf_Man_t, store []*Jf_Cut_t, count int32) int32 {
	return jfObjCutFilterBoth(store, count)
}
func Jf_ObjCutFilter(p *Jf_Man_t, store []*Jf_Cut_t, count int32) int32 {
	if jfObjCutFilter(store, count, p.pPars.FCutMin != 0) {
		return 1
	}
	return 0
}
func Jf_ObjAddCutToStore(p *Jf_Man_t, store []*Jf_Cut_t, count, max int32) int32 {
	return jfObjAddCutToStore(store, count, max, p.pPars.FCutMin != 0, p.pCutCmp)
}
func jfAccounting(p *Jf_Man_t) *jfCutAccounting {
	return &jfCutAccounting{refs: p.pGia.Refs, best: func(id int32) []int32 { return Jf_ObjCutBest(p, id) }}
}
func Jf_CutRef_rec(p *Jf_Man_t, cut []int32) int32 {
	return jfAccounting(p).ref(cut, true, false, false)
}
func Jf_CutDeref_rec(p *Jf_Man_t, cut []int32) int32 {
	return jfAccounting(p).ref(cut, false, false, false)
}
func Jf_CutAreaOld(p *Jf_Man_t, cut []int32) int32 {
	a := Jf_CutRef_rec(p, cut)
	b := Jf_CutDeref_rec(p, cut)
	if a != b {
		panic("JF inconsistent area")
	}
	return a
}
func Jf_CutAreaRef_rec(p *Jf_Man_t, cut []int32) int32 {
	a := jfAccounting(p)
	n := a.ref(cut, true, false, true)
	for _, id := range a.touched {
		vec.Vec_IntPush(p.vTemp, id)
	}
	return n
}
func Jf_CutAreaRefEdge_rec(p *Jf_Man_t, cut []int32) int32 {
	a := jfAccounting(p)
	n := a.ref(cut, true, true, true)
	for _, id := range a.touched {
		vec.Vec_IntPush(p.vTemp, id)
	}
	return n
}
func Jf_CutArea(p *Jf_Man_t, cut []int32, edge int32) int32 {
	a := jfAccounting(p)
	n := a.ref(cut, true, edge != 0, true)
	vec.Vec_IntClear(p.vTemp)
	for _, id := range a.touched {
		vec.Vec_IntPush(p.vTemp, id)
		a.refs[id]--
	}
	return n
}
func Jf_CutCheckMffc_rec(p *Jf_Man_t, cut []int32, limit int32) int32 {
	a := jfAccounting(p)
	a.touched = append(a.touched, vec.Vec_IntArray(p.vTemp)...)
	prior := len(a.touched)
	ok := a.checkMffcRec(cut, int(limit))
	for _, id := range a.touched[prior:] {
		vec.Vec_IntPush(p.vTemp, id)
	}
	if ok {
		return 1
	}
	return 0
}
func Jf_CutCheckMffc(p *Jf_Man_t, cut []int32, limit int32) int32 {
	a := jfAccounting(p)
	ok := a.checkMffcRec(cut, int(limit))
	vec.Vec_IntClear(p.vTemp)
	for _, id := range a.touched {
		vec.Vec_IntPush(p.vTemp, id)
		a.refs[id]++
	}
	if ok {
		return 1
	}
	return 0
}
func Jf_ObjHasCutWithSize(store []*Jf_Cut_t, count, size int32) int32 {
	for i := int32(0); i < count; i++ {
		if store[i].pCut[0] <= size {
			return 1
		}
	}
	return 0
}
func Jf_ObjAssignCut(p *Jf_Man_t, obj *Gia_Obj_t) {
	id := Gia_ObjId(p.pGia, obj)
	if Gia_ObjIsCi(obj) == 0 && Gia_ObjIsBuf(obj) == 0 {
		panic("JF expected CI or buffer")
	}
	data := []int32{1, Jf_CutSetAll(2, 0, 1), Jf_ObjLit(id, 0)}
	vec.Vec_IntWriteEntry(&p.vCuts, int(id), p.pMem.append(data, len(data)))
}
