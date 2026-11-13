// Go port of cecChoice.c, by Alan Mishchenko, UC Berkeley (2005).
package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

// Cec_ManCombSpecReal constructs the real function of an AND whose fanins
// may themselves have been replaced by speculative representatives.
func Cec_ManCombSpecReal(dst, src *Gia_Man_t, obj *Gia_Obj_t) int32 {
	if gia.Gia_ObjIsAnd(obj) == 0 {
		panic("cec: expected AND node")
	}
	Cec_ManCombSpecReduce_rec(dst, src, gia.Gia_ObjFanin0(obj))
	Cec_ManCombSpecReduce_rec(dst, src, gia.Gia_ObjFanin1(obj))
	return gia.Gia_ManHashAnd(dst, gia.Gia_ObjFanin0Copy(obj), gia.Gia_ObjFanin1Copy(obj))
}

func Cec_ManCombSpecReduce_rec(dst, src *Gia_Man_t, obj *Gia_Obj_t) {
	if gia.Gia_ObjValue(obj) != -1 {
		return
	}
	if repr := gia.Gia_ObjReprObj(src, gia.Gia_ObjId(src, obj)); repr != nil {
		Cec_ManCombSpecReduce_rec(dst, src, repr)
		gia.Gia_ObjSetValue(obj, gia.Gia_ObjValue(repr)^gia.Gia_ObjPhase(repr)^gia.Gia_ObjPhase(obj))
		return
	}
	gia.Gia_ObjSetValue(obj, Cec_ManCombSpecReal(dst, src, obj))
}

// Cec_ManCombSpecReduce constructs the comparison outputs for equivalence
// candidates. outputs records the original object pair for each output.
func Cec_ManCombSpecReduce(src *Gia_Man_t, outputs **Vec_Int_t, rings int32) *Gia_Man_t {
	if src == nil || outputs == nil || !src.HasRepresentatives() {
		panic("cec: missing representatives or output vector")
	}
	gia.Gia_ManSetPhase(src)
	gia.Gia_ManFillValue(src)
	dst := gia.Gia_ManStart(gia.Gia_ManObjNum(src))
	dst.CopyIdentityFrom(src)
	gia.Gia_ManHashAlloc(dst)
	gia.Gia_ObjSetValue(gia.Gia_ManConst0(src), 0)
	for i := int32(0); i < gia.Gia_ManCiNum(src); i++ {
		gia.Gia_ObjSetValue(gia.Gia_ManCi(src, i), gia.Gia_ManAppendCi(dst))
	}
	*outputs = vec.Vec_IntAlloc(1000)
	xors := vec.Vec_IntAlloc(1000)
	add := func(a, b, lit int32) {
		vec.Vec_IntPush(*outputs, a)
		vec.Vec_IntPush(*outputs, b)
		vec.Vec_IntPush(xors, lit)
	}
	for id := int32(1); id < gia.Gia_ManObjNum(src); id++ {
		obj := gia.Gia_ManObj(src, id)
		if rings != 0 {
			if gia.Gia_ObjIsConst(src, id) != 0 {
				lit := Cec_ManCombSpecReal(dst, src, obj) ^ gia.Gia_ObjPhase(obj)
				if lit != 0 {
					add(0, id, lit)
				}
			} else if gia.Gia_ObjIsHead(src, id) != 0 {
				prev := id
				for next := gia.Gia_ObjNext(src, id); next > 0; next = gia.Gia_ObjNext(src, next) {
					a := Cec_ManCombSpecReal(dst, src, gia.Gia_ManObj(src, prev)) ^ gia.Gia_ObjPhase(obj) ^ gia.Gia_ObjPhase(gia.Gia_ManObj(src, prev))
					b := Cec_ManCombSpecReal(dst, src, gia.Gia_ManObj(src, next)) ^ gia.Gia_ObjPhase(obj) ^ gia.Gia_ObjPhase(gia.Gia_ManObj(src, next))
					if a != b && a != 0 && b != 1 {
						add(prev, next, gia.Gia_ManHashAnd(dst, a, b^1))
					}
					prev = next
				}
				a := Cec_ManCombSpecReal(dst, src, gia.Gia_ManObj(src, prev)) ^ gia.Gia_ObjPhase(obj) ^ gia.Gia_ObjPhase(gia.Gia_ManObj(src, prev))
				b := Cec_ManCombSpecReal(dst, src, obj)
				if a != b && a != 0 && b != 1 {
					add(prev, id, gia.Gia_ManHashAnd(dst, a, b^1))
				}
			}
			continue
		}
		repr := gia.Gia_ObjReprObj(src, id)
		if repr == nil {
			continue
		}
		var a int32
		if gia.Gia_ObjIsConst(src, id) == 0 {
			a = Cec_ManCombSpecReal(dst, src, repr)
		}
		b := Cec_ManCombSpecReal(dst, src, obj) ^ gia.Gia_ObjPhase(repr) ^ gia.Gia_ObjPhase(obj)
		if a != b {
			add(gia.Gia_ObjId(src, repr), id, gia.Gia_ManHashXor(dst, a, b))
		}
	}
	for _, lit := range vec.Vec_IntArray(xors) {
		gia.Gia_ManAppendCo(dst, lit)
	}
	gia.Gia_ManHashStop(dst)
	result := gia.Gia_ManCleanup(dst)
	gia.Gia_ManStop(dst)
	return result
}
