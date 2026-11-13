// Go port of absDup.c by Alan Mishchenko, UC Berkeley (2005).
package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_ManDupAbsFlops_rec(dst *Gia_Man_t, obj *Gia_Obj_t) {
	if obj.Value != ^uint32(0) {
		return
	}
	if gia.Gia_ObjIsAnd(obj) == 0 {
		panic("abs: expected unmapped AND")
	}
	Gia_ManDupAbsFlops_rec(dst, gia.Gia_ObjFanin0(obj))
	Gia_ManDupAbsFlops_rec(dst, gia.Gia_ObjFanin1(obj))
	obj.Value = uint32(gia.Gia_ManHashAnd(dst, gia.Gia_ObjFanin0Copy(obj), gia.Gia_ObjFanin1Copy(obj)))
}

// Gia_ManDupAbsFlops exposes excluded registers as additional inputs.
func Gia_ManDupAbsFlops(p *Gia_Man_t, classes *Vec_Int_t) *Gia_Man_t {
	if vec.Vec_IntSize(classes) != int(gia.Gia_ManRegNum(p)) {
		panic("abs: invalid flop classes")
	}
	gia.Gia_ManFillValue(p)
	dst := gia.Gia_ManStart(5000)
	gia.Gia_ManCopyNameSpec(dst, p)
	gia.Gia_ManConst0(p).Value = 0
	for i := int32(0); i < gia.Gia_ManPiNum(p); i++ {
		gia.Gia_ManPi(p, i).Value = uint32(gia.Gia_ManAppendCi(dst))
	}
	for i := int32(0); i < gia.Gia_ManRegNum(p); i++ {
		if vec.Vec_IntEntry(classes, int(i)) == 0 {
			gia.Gia_ManRo(p, i).Value = uint32(gia.Gia_ManAppendCi(dst))
		}
	}
	var flops int32
	for i := int32(0); i < gia.Gia_ManRegNum(p); i++ {
		if vec.Vec_IntEntry(classes, int(i)) != 0 {
			gia.Gia_ManRo(p, i).Value = uint32(gia.Gia_ManAppendCi(dst))
			flops++
		}
	}
	gia.Gia_ManHashAlloc(dst)
	for i := int32(0); i < gia.Gia_ManPoNum(p); i++ {
		obj := gia.Gia_ManPo(p, i)
		Gia_ManDupAbsFlops_rec(dst, gia.Gia_ObjFanin0(obj))
		gia.Gia_ManAppendCo(dst, gia.Gia_ObjFanin0Copy(obj))
	}
	for i := int32(0); i < gia.Gia_ManRegNum(p); i++ {
		if vec.Vec_IntEntry(classes, int(i)) != 0 {
			obj := gia.Gia_ManRi(p, i)
			Gia_ManDupAbsFlops_rec(dst, gia.Gia_ObjFanin0(obj))
			gia.Gia_ManAppendCo(dst, gia.Gia_ObjFanin0Copy(obj))
		}
	}
	gia.Gia_ManHashStop(dst)
	gia.Gia_ManSetRegNum(dst, flops)
	gia.Gia_ManSetConstrNum(dst, gia.Gia_ManConstrNum(p))
	result := gia.Gia_ManSeqCleanup(dst)
	gia.Gia_ManStop(dst)
	return result
}

func Gia_GlaCollectAssigned(p *Gia_Man_t, classes *Vec_Int_t) *Vec_Int_t {
	return gia.Gia_GlaCollectAssigned(p, classes)
}
func Gia_ManGlaCollect(p *Gia_Man_t, classes *Vec_Int_t, pis, ppis, flops, nodes **Vec_Int_t) {
	gia.Gia_ManGlaCollect(p, classes, pis, ppis, flops, nodes)
}

func Gia_ManDupAbsGates_rec(dst *Gia_Man_t, obj *Gia_Obj_t) {
	if obj.Value != ^uint32(0) {
		return
	}
	if gia.Gia_ObjIsAnd(obj) == 0 {
		panic("abs: expected unmapped AND")
	}
	Gia_ManDupAbsGates_rec(dst, gia.Gia_ObjFanin0(obj))
	Gia_ManDupAbsGates_rec(dst, gia.Gia_ObjFanin1(obj))
	obj.Value = uint32(gia.Gia_ManAppendAnd(dst, gia.Gia_ObjFanin0Copy(obj), gia.Gia_ObjFanin1Copy(obj)))
}

func Gia_ManDupAbsGates(p *Gia_Man_t, classes *Vec_Int_t) *Gia_Man_t {
	var pis, ppis, flops, nodes *Vec_Int_t
	Gia_ManGlaCollect(p, classes, &pis, &ppis, &flops, &nodes)
	// A primary output may be driven directly by a primary input. It has no
	// classified gate to pull it into the collected input set.
	for i := int32(0); i < gia.Gia_ManPoNum(p); i++ {
		driver := gia.Gia_ObjFanin0(gia.Gia_ManPo(p, i))
		if gia.Gia_ObjIsPi(p, driver) != 0 {
			id := gia.Gia_ObjId(p, driver)
			if vec.Vec_IntFind(pis, id) < 0 {
				vec.Vec_IntPush(pis, id)
			}
		}
	}
	vec.Vec_IntSort(pis, 0)
	dst := gia.Gia_ManStart(5000)
	gia.Gia_ManCopyNameSpec(dst, p)
	gia.Gia_ManFillValue(p)
	gia.Gia_ManConst0(p).Value = 0
	for _, group := range []*Vec_Int_t{pis, ppis, flops} {
		for _, id := range vec.Vec_IntArray(group) {
			gia.Gia_ManObj(p, id).Value = uint32(gia.Gia_ManAppendCi(dst))
		}
	}
	for _, id := range vec.Vec_IntArray(nodes) {
		obj := gia.Gia_ManObj(p, id)
		a, b, c, kind := p.Gate(id)
		copyLit := func(lit int32) int32 {
			return int32(gia.Gia_ManObj(p, lit>>1).Value) ^ (lit & 1)
		}
		switch kind {
		case 2:
			obj.Value = uint32(gia.Gia_ManAppendAnd(dst, copyLit(a), copyLit(b)))
		case 3:
			obj.Value = uint32(gia.Gia_ManAppendXorReal(dst, copyLit(a), copyLit(b)))
		case 4:
			dst.EnableMuxes()
			obj.Value = uint32(gia.Gia_ManAppendMuxReal(dst, copyLit(a), copyLit(b), copyLit(c)))
		case 5:
			obj.Value = uint32(gia.Gia_ManAppendBuf(dst, copyLit(a)))
		default:
			panic("abs: invalid gate abstraction node")
		}
	}
	for i := int32(0); i < gia.Gia_ManPoNum(p); i++ {
		obj := gia.Gia_ManPo(p, i)
		obj.Value = uint32(gia.Gia_ManAppendCo(dst, gia.Gia_ObjFanin0Copy(obj)))
	}
	for _, id := range vec.Vec_IntArray(flops) {
		ri := gia.Gia_ObjRoToRi(p, gia.Gia_ManObj(p, id))
		ri.Value = uint32(gia.Gia_ManAppendCo(dst, gia.Gia_ObjFanin0Copy(ri)))
	}
	gia.Gia_ManSetRegNum(dst, int32(vec.Vec_IntSize(flops)))
	gia.Gia_ManSetConstrNum(dst, gia.Gia_ManConstrNum(p))
	result := gia.Gia_ManSeqCleanup(dst)
	if gia.Gia_ManObjNum(dst) != gia.Gia_ManObjNum(result) {
		for id := int32(0); id < gia.Gia_ManObjNum(p); id++ {
			obj := gia.Gia_ManObj(p, id)
			if obj.Value == ^uint32(0) {
				continue
			}
			copyObj := gia.Gia_ObjCopy(dst, obj)
			if copyObj.Value == ^uint32(0) {
				vec.Vec_IntWriteEntry(classes, int(id), 0)
				obj.Value = ^uint32(0)
				continue
			}
			obj.Value = copyObj.Value
		}
	}
	gia.Gia_ManStop(dst)
	return result
}
func Gia_ManPrintFlopClasses(p *Gia_Man_t) { gia.Gia_ManPrintFlopClasses(p) }
func Gia_ManPrintGateClasses(p *Gia_Man_t) { gia.Gia_ManPrintGateClasses(p) }
func Gia_ManPrintObjClasses(p *Gia_Man_t)  { gia.Gia_ManPrintObjClasses(p) }
