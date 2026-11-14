package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_ManSpecBuild(q, p *Gia_Man_t, obj *Gia_Obj_t, xors *Vec_Int_t, dual, speculate int32, trace, guide, mapping *Vec_Int_t) {
	repr := Gia_ObjReprObj(p, obj.id)
	if repr == nil || dual != 0 && Gia_ObjDiffColors2(p, obj.id, repr.id) == 0 {
		return
	}
	expected := Gia_ObjValue(repr) ^ Gia_ObjPhaseReal(repr) ^ Gia_ObjPhaseReal(obj)
	if Gia_ObjValue(obj) != expected && Gia_ObjProved(p, obj.id) == 0 {
		if trace != nil {
			vec.Vec_IntPush(trace, 1)
		}
		if guide == nil || vec.Vec_IntEntry(guide, vec.Vec_IntSize(trace)-1) != 0 {
			if mapping != nil {
				vec.Vec_IntPush(mapping, obj.id)
			}
			vec.Vec_IntPush(xors, Gia_ManHashXor(q, Gia_ObjValue(obj), expected))
		}
	} else if trace != nil {
		vec.Vec_IntPush(trace, 0)
	}
	if speculate != 0 {
		Gia_ObjSetValue(obj, expected)
	}
}
func Gia_ManSpecReduce_rec(q, p *Gia_Man_t, obj *Gia_Obj_t, xors *Vec_Int_t, dual, speculate int32, trace, guide, mapping *Vec_Int_t) {
	if Gia_ObjValue(obj) != -1 {
		return
	}
	if Gia_ObjIsAnd(obj) == 0 {
		panic("gia: unassigned speculative input")
	}
	Gia_ManSpecReduce_rec(q, p, Gia_ObjFanin0(obj), xors, dual, speculate, trace, guide, mapping)
	Gia_ManSpecReduce_rec(q, p, Gia_ObjFanin1(obj), xors, dual, speculate, trace, guide, mapping)
	Gia_ObjSetValue(obj, Gia_ManHashAnd(q, Gia_ObjFanin0Copy(obj), Gia_ObjFanin1Copy(obj)))
	Gia_ManSpecBuild(q, p, obj, xors, dual, speculate, trace, guide, mapping)
}
func giaSpecReduction(p *Gia_Man_t, dual, synthesis, speculate int32, trace, guide, mapping *Vec_Int_t) *Gia_Man_t {
	Gia_ManSetPhase(p)
	q := giaEquivReductionStart(p)
	xors := vec.Vec_IntAlloc(100)
	for i := int32(0); i < p.nRegs; i++ {
		Gia_ManSpecBuild(q, p, Gia_ManRo(p, i), xors, dual, speculate, trace, guide, mapping)
	}
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		Gia_ManSpecReduce_rec(q, p, Gia_ObjFanin0(Gia_ManCo(p, i)), xors, dual, speculate, trace, guide, mapping)
	}
	if synthesis == 0 {
		for i := int32(0); i < Gia_ManPoNum(p); i++ {
			Gia_ManAppendCo(q, Gia_ObjFanin0Copy(Gia_ManPo(p, i)))
		}
	}
	for _, lit := range vec.Vec_IntArray(xors) {
		Gia_ManAppendCo(q, lit)
	}
	if vec.Vec_IntSize(xors) == 0 {
		fmt.Println("Speculatively reduced model has no primary outputs.")
		Gia_ManAppendCo(q, 0)
	}
	for i := int32(0); i < p.nRegs; i++ {
		Gia_ManAppendCo(q, Gia_ObjFanin0Copy(Gia_ManRi(p, i)))
	}
	Gia_ManHashStop(q)
	Gia_ManSetRegNum(q, p.nRegs)
	result := Gia_ManCleanup(q)
	Gia_ManStop(q)
	return result
}
func Gia_ManSpecReduceTrace(p *Gia_Man_t, trace, mapping *Vec_Int_t) *Gia_Man_t {
	if p.PReprs == nil {
		fmt.Println("Gia_ManSpecReduce(): Equivalence classes are not available.")
		return nil
	}
	vec.Vec_IntClear(trace)
	return giaSpecReduction(p, 0, 0, 1, trace, nil, mapping)
}
func Gia_ManSpecReduce(p *Gia_Man_t, dual, synthesis, speculate, skipSome, verbose int32) *Gia_Man_t {
	if p.PReprs == nil {
		fmt.Println("Gia_ManSpecReduce(): Equivalence classes are not available.")
		return nil
	}
	if dual != 0 && Gia_ManPoNum(p)&1 != 0 {
		fmt.Println("Gia_ManSpecReduce(): Dual-output miter should have even number of POs.")
		return nil
	}
	var trace, guide *Vec_Int_t
	if skipSome != 0 {
		guide = vec.Vec_IntAlloc(100)
		tmp := Gia_ManSpecReduceTrace(p, guide, nil)
		Gia_ManStop(tmp)
		if int32(vec.Vec_IntSize(guide)) != Gia_ManEquivCountLitsAll(p) {
			panic("gia: incomplete speculative trace")
		}
		trace = vec.Vec_IntAlloc(100)
	}
	if dual != 0 {
		Gia_ManEquivSetColors(p, verbose)
	}
	return giaSpecReduction(p, dual, synthesis, speculate, trace, guide, nil)
}

func Gia_ManCombSpecReduce(p *Gia_Man_t) *Gia_Man_t {
	if p.PReprs == nil || p.PNexts == nil {
		panic("gia: missing equivalence classes")
	}
	Gia_ManLevelNum(p)
	Gia_ManSetPhase(p)
	q := giaEquivReductionStart(p)
	xors := vec.Vec_IntAlloc(100)
	for _, obj := range p.pObjs {
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		Gia_ObjSetValue(obj, Gia_ManHashAnd(q, Gia_ObjFanin0Copy(obj), Gia_ObjFanin1Copy(obj)))
		repr := Gia_ObjReprObj(p, obj.id)
		if repr != nil && Gia_ObjValue(obj)>>1 != Gia_ObjValue(repr)>>1 {
			lit := Gia_ObjValue(repr) ^ int32(obj.fPhase^repr.fPhase)
			vec.Vec_IntPush(xors, Gia_ManHashXor(q, Gia_ObjValue(obj), lit))
			Gia_ObjSetValue(obj, lit)
		}
	}
	Gia_ManHashStop(q)
	if vec.Vec_IntSize(xors) == 0 {
		vec.Vec_IntPush(xors, 0)
	}
	for _, lit := range vec.Vec_IntArray(xors) {
		Gia_ManAppendCo(q, lit)
	}
	Gia_ManSetRegNum(q, p.nRegs)
	result := Gia_ManCleanup(q)
	Gia_ManStop(q)
	return result
}
func Gia_ManCombSpecReduceTest(p *Gia_Man_t, filename string) {
	q := Gia_ManCombSpecReduce(p)
	if filename == "" {
		filename = "test.aig"
	}
	if err := Gia_AigerWrite(q, filename, 0, 0, 0); err != nil {
		fmt.Printf("Cannot write %q: %v\n", filename, err)
	} else {
		fmt.Printf("Speculatively reduced model was written into file %q.\n", filename)
	}
	Gia_ManStop(q)
}
