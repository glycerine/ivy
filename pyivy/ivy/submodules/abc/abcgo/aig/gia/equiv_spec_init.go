package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_ManSpecBuildInit(q, p *Gia_Man_t, obj *Gia_Obj_t, xors *Vec_Int_t, frame, dual int32) {
	repr := Gia_ObjReprObj(p, obj.id)
	if repr == nil || dual != 0 && Gia_ObjDiffColors2(p, obj.id, repr.id) == 0 {
		return
	}
	expected := Gia_ObjCopyF(p, frame, repr) ^ Gia_ObjPhaseReal(repr) ^ Gia_ObjPhaseReal(obj)
	if Gia_ObjCopyF(p, frame, obj) != expected && Gia_ObjProved(p, obj.id) == 0 {
		vec.Vec_IntPush(xors, Gia_ManHashXor(q, Gia_ObjCopyF(p, frame, obj), expected))
	}
	Gia_ObjSetCopyF(p, frame, obj, expected)
}
func Gia_ManSpecReduceInit_rec(q, p *Gia_Man_t, obj *Gia_Obj_t, xors *Vec_Int_t, frame, dual int32) {
	if Gia_ObjCopyF(p, frame, obj) != -1 {
		return
	}
	if Gia_ObjIsAnd(obj) == 0 {
		panic("gia: unassigned initialized speculative input")
	}
	Gia_ManSpecReduceInit_rec(q, p, Gia_ObjFanin0(obj), xors, frame, dual)
	Gia_ManSpecReduceInit_rec(q, p, Gia_ObjFanin1(obj), xors, frame, dual)
	Gia_ObjSetCopyF(p, frame, obj, Gia_ManHashAnd(q, Gia_ObjFanin0CopyF(p, frame, obj), Gia_ObjFanin1CopyF(p, frame, obj)))
	Gia_ManSpecBuildInit(q, p, obj, xors, frame, dual)
}
func Gia_ManSpecReduceInit(p *Gia_Man_t, init *Abc_Cex_t, nFrames, dual int32) *Gia_Man_t {
	if p.pReprs == nil {
		fmt.Println("Gia_ManSpecReduceInit(): Equivalence classes are not available.")
		return nil
	}
	if p.nRegs == 0 {
		fmt.Println("Gia_ManSpecReduceInit(): Circuit is not sequential.")
		return nil
	}
	if init == nil || init.NRegs != p.nRegs {
		fmt.Println("Gia_ManSpecReduceInit(): Mismatch in the number of registers.")
		return nil
	}
	if dual != 0 && Gia_ManPoNum(p)&1 != 0 {
		fmt.Println("Gia_ManSpecReduceInit(): Dual-output miter should have even number of POs.")
		return nil
	}
	if init.NPis != 0 || nFrames <= 0 {
		panic("gia: invalid speculative initialization")
	}
	vec.Vec_IntFill(&p.vCopies, int(nFrames*p.nObjs), -1)
	xors := vec.Vec_IntAlloc(100)
	Gia_ManSetPhase(p)
	if dual != 0 {
		Gia_ManEquivSetColors(p, 0)
	}
	q := Gia_ManStart(nFrames * p.nObjs)
	q.pName = p.pName
	q.pSpec = p.pSpec
	Gia_ManHashAlloc(q)
	for i := int32(0); i < p.nRegs; i++ {
		bit := int32(0)
		if util.Abc_InfoHasBit(init.PData, i) {
			bit = 1
		}
		Gia_ObjSetCopyF(p, 0, Gia_ManRo(p, i), bit)
	}
	for f := int32(0); f < nFrames; f++ {
		Gia_ObjSetCopyF(p, f, Gia_ManConst0(p), 0)
		for i := int32(0); i < Gia_ManPiNum(p); i++ {
			Gia_ObjSetCopyF(p, f, Gia_ManPi(p, i), Gia_ManAppendCi(q))
		}
		for i := int32(0); i < p.nRegs; i++ {
			Gia_ManSpecBuildInit(q, p, Gia_ManRo(p, i), xors, f, dual)
		}
		for i := int32(0); i < Gia_ManCoNum(p); i++ {
			obj := Gia_ManCo(p, i)
			Gia_ManSpecReduceInit_rec(q, p, Gia_ObjFanin0(obj), xors, f, dual)
			Gia_ObjSetCopyF(p, f, obj, Gia_ObjFanin0CopyF(p, f, obj))
		}
		if f+1 < nFrames {
			for i := int32(0); i < p.nRegs; i++ {
				Gia_ObjSetCopyF(p, f+1, Gia_ManRo(p, i), Gia_ObjCopyF(p, f, Gia_ManRi(p, i)))
			}
		}
	}
	for _, lit := range vec.Vec_IntArray(xors) {
		Gia_ManAppendCo(q, lit)
	}
	if vec.Vec_IntSize(xors) == 0 {
		Gia_ManAppendCo(q, 0)
	}
	vec.Vec_IntErase(&p.vCopies)
	Gia_ManHashStop(q)
	result := Gia_ManCleanup(q)
	Gia_ManStop(q)
	return result
}
func Gia_ManSpecReduceInitFrames(p *Gia_Man_t, init *Abc_Cex_t, maxFrames int32, frames *int32, dual, minOutputs int32) *Gia_Man_t {
	nLits := Gia_ManEquivCountLits(p)
	for f := int32(1); ; f++ {
		result := Gia_ManSpecReduceInit(p, init, f, dual)
		if result == nil {
			return nil
		}
		if minOutputs == 0 && Gia_ManPoNum(result) >= nLits/2+1 || minOutputs != 0 && Gia_ManPoNum(result) >= minOutputs || f == maxFrames {
			if f == maxFrames {
				fmt.Printf("Stopped unrolling after %d frames.\n", maxFrames)
			}
			if frames != nil {
				*frames = f
			}
			return result
		}
		if Gia_ManAndNum(result) > 500000 {
			Gia_ManStop(result)
			return nil
		}
		Gia_ManStop(result)
	}
}
