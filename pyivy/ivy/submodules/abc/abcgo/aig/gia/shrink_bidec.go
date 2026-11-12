package gia

import (
	"github.com/glycerine/abcgo/bool/bdc"
	"github.com/glycerine/abcgo/misc/vec"
)

func Shr_ObjPerformBidec(p *Shr_Man_t, manager *bdc.Bdc_Man_t, dst *Gia_Man_t, leaves *Vec_Int_t, truth, care word) int32 {
	if truth == 0 || care == 0 {
		panic("gia: invalid bidecomposition truth")
	}
	words := []uint32{uint32(truth), uint32(truth >> 32)}
	careWords := []uint32{uint32(care), uint32(care >> 32)}
	n := int32(vec.Vec_IntSize(leaves))
	if bdc.Bdc_ManDecompose(manager, words, careWords, n, nil, 1000) < 0 {
		panic("gia: shrink bidecomposition failed")
	}
	bdc.Bdc_FuncSetCopyInt(bdc.Bdc_ManFunc(manager, 0), 1)
	for i, id := range vec.Vec_IntArray(leaves) {
		bdc.Bdc_FuncSetCopyInt(bdc.Bdc_ManFunc(manager, int32(i+1)), 2*id)
	}
	for i := n + 1; i < bdc.Bdc_ManNodeNum(manager); i++ {
		f := bdc.Bdc_ManFunc(manager, i)
		before := dst.nObjs
		lit := Gia_ManHashAnd(dst, bdc.Bdc_FunFanin0Copy(f), bdc.Bdc_FunFanin1Copy(f))
		bdc.Bdc_FuncSetCopyInt(f, lit)
		if dst.nObjs == before {
			continue
		}
		obj := Gia_ManObj(dst, lit>>1)
		Gia_ObjSetAndLevel(dst, obj)
		Shr_ManAddFanout(p, Gia_ObjFaninId0p(dst, obj), obj.id)
		Shr_ManAddFanout(p, Gia_ObjFaninId1p(dst, obj), obj.id)
		if dst.nObjs >= p.nNewSize {
			panic("gia: shrink manager capacity exceeded")
		}
	}
	return bdc.Bdc_FunObjCopy(bdc.Bdc_ManRoot(manager))
}
func Shr_ManComputeTruth6_rec(p *Gia_Man_t, id int32, truths *Vec_Wrd_t) word {
	if Gia_ObjIsTravIdCurrentId(p, id) != 0 {
		return vec.Vec_WrdEntry(truths, int(id))
	}
	Gia_ObjSetTravIdCurrentId(p, id)
	obj := Gia_ManObj(p, id)
	if Gia_ObjIsAnd(obj) == 0 {
		panic("gia: shrink cut does not cover cone")
	}
	a := Shr_ManComputeTruth6_rec(p, Gia_ObjFaninId0(obj, id), truths)
	b := Shr_ManComputeTruth6_rec(p, Gia_ObjFaninId1(obj, id), truths)
	if Gia_ObjFaninC0(obj) != 0 {
		a = ^a
	}
	if Gia_ObjFaninC1(obj) != 0 {
		b = ^b
	}
	value := a & b
	vec.Vec_WrdWriteEntry(truths, int(id), value)
	return value
}
func Shr_ManComputeTruth6(p *Gia_Man_t, obj *Gia_Obj_t, leaves *Vec_Int_t, truths *Vec_Wrd_t) word {
	patterns := [6]word{0xaaaaaaaaaaaaaaaa, 0xcccccccccccccccc, 0xf0f0f0f0f0f0f0f0, 0xff00ff00ff00ff00, 0xffff0000ffff0000, 0xffffffff00000000}
	if Gia_ObjIsAnd(obj) == 0 || vec.Vec_IntSize(leaves) > 6 {
		panic("gia: invalid shrink truth cut")
	}
	Gia_ManIncrementTravId(p)
	for i, id := range vec.Vec_IntArray(leaves) {
		Gia_ObjSetTravIdCurrentId(p, id)
		vec.Vec_WrdWriteEntry(truths, int(id), patterns[i])
	}
	return Shr_ManComputeTruth6_rec(p, obj.id, truths)
}
