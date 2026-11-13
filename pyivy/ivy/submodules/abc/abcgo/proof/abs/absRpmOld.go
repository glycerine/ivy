package abs

// Port of absRpmOld.c by Alan Mishchenko, UC Berkeley (2005).

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/aig/saig"
)

// Gia_ManDupIn2Ff replaces the primary inputs with register outputs while
// exposing the original register outputs as primary inputs.
func Gia_ManDupIn2Ff(p *gia.Gia_Man_t) *gia.Gia_Man_t {
	piCount := gia.Gia_ManPiNum(p)
	q := gia.Gia_ManStart(gia.Gia_ManObjNum(p) + 2*piCount)
	gia.Gia_ManCopyNameSpec(q, p)
	gia.Gia_ManFillValue(p)
	gia.Gia_ManConst0(p).Value = 0
	piOutputs := make([]int32, piCount)
	for i := int32(0); i < piCount; i++ {
		piOutputs[i] = gia.Gia_ManAppendCi(q)
	}
	for i := int32(0); i < gia.Gia_ManRegNum(p); i++ {
		gia.Gia_ManRo(p, i).Value = uint32(gia.Gia_ManAppendCi(q))
	}
	for i := int32(0); i < piCount; i++ {
		gia.Gia_ManPi(p, i).Value = uint32(gia.Gia_ManAppendCi(q))
	}
	for i := int32(1); i < gia.Gia_ManObjNum(p); i++ {
		obj := gia.Gia_ManObj(p, i)
		if gia.Gia_ObjIsAnd(obj) != 0 {
			obj.Value = uint32(gia.Gia_ManAppendAnd(q, gia.Gia_ObjFanin0Copy(obj), gia.Gia_ObjFanin1Copy(obj)))
		}
	}
	for i := int32(0); i < gia.Gia_ManPoNum(p); i++ {
		obj := gia.Gia_ManPo(p, i)
		obj.Value = uint32(gia.Gia_ManAppendCo(q, gia.Gia_ObjFanin0Copy(obj)))
	}
	for i := int32(0); i < gia.Gia_ManRegNum(p); i++ {
		obj := gia.Gia_ManRi(p, i)
		obj.Value = uint32(gia.Gia_ManAppendCo(q, gia.Gia_ObjFanin0Copy(obj)))
	}
	for _, lit := range piOutputs {
		gia.Gia_ManAppendCo(q, lit)
	}
	gia.Gia_ManSetRegNum(q, piCount)
	return q
}

func Gia_ManDupFf2In_rec(dst *gia.Gia_Man_t, obj *gia.Gia_Obj_t) int32 {
	if obj.Value != ^uint32(0) {
		return int32(obj.Value)
	}
	if gia.Gia_ObjIsAnd(obj) == 0 {
		panic("abs: expected AND in reverse input/register conversion")
	}
	Gia_ManDupFf2In_rec(dst, gia.Gia_ObjFanin0(obj))
	Gia_ManDupFf2In_rec(dst, gia.Gia_ObjFanin1(obj))
	obj.Value = uint32(gia.Gia_ManAppendAnd(dst, gia.Gia_ObjFanin0Copy(obj), gia.Gia_ObjFanin1Copy(obj)))
	return int32(obj.Value)
}

// Gia_ManDupFf2In reverses Gia_ManDupIn2Ff after sequential retiming.
func Gia_ManDupFf2In(p *gia.Gia_Man_t, oldFlops int32) *gia.Gia_Man_t {
	if oldFlops < 0 || oldFlops > gia.Gia_ManPiNum(p) {
		panic("abs: invalid original register count")
	}
	q := gia.Gia_ManStart(gia.Gia_ManObjNum(p))
	gia.Gia_ManCopyNameSpec(q, p)
	gia.Gia_ManFillValue(p)
	gia.Gia_ManConst0(p).Value = 0
	for i := int32(0); i < gia.Gia_ManRegNum(p); i++ {
		gia.Gia_ManRo(p, i).Value = uint32(gia.Gia_ManAppendCi(q))
	}
	for i := gia.Gia_ManPiNum(p) - oldFlops; i < gia.Gia_ManPiNum(p); i++ {
		gia.Gia_ManPi(p, i).Value = uint32(gia.Gia_ManAppendCi(q))
	}
	for i := int32(0); i < gia.Gia_ManPoNum(p); i++ {
		Gia_ManDupFf2In_rec(q, gia.Gia_ObjFanin0(gia.Gia_ManPo(p, i)))
	}
	for i := int32(0); i < gia.Gia_ManPoNum(p); i++ {
		gia.Gia_ManAppendCo(q, gia.Gia_ObjFanin0Copy(gia.Gia_ManPo(p, i)))
	}
	gia.Gia_ManSetRegNum(q, oldFlops)
	return q
}

func Abs_RpmPerformOld(p *gia.Gia_Man_t, verbose int32) *gia.Gia_Man_t {
	oldFlops := gia.Gia_ManRegNum(p)
	if verbose != 0 {
		fmt.Println("Original AIG:")
		gia.Gia_ManPrintStats(p, nil)
	}
	trimmed := gia.Gia_ManDupTrimmed(p, 1, 0, 0, -1)
	if verbose != 0 {
		fmt.Println("After PI trimming:")
		gia.Gia_ManPrintStats(trimmed, nil)
	}
	converted := Gia_ManDupIn2Ff(trimmed)
	gia.Gia_ManStop(trimmed)
	if verbose != 0 {
		fmt.Println("After PI-2-FF transformation:")
		gia.Gia_ManPrintStats(converted, nil)
	}
	graph := gia.Gia_ManToAigSimple(converted)
	gia.Gia_ManStop(converted)
	retimed := saig.Saig_ManRetimeMinArea(graph, 10, 0, 0, 1, 0)
	aig.Aig_ManStop(graph)
	converted = gia.Gia_ManFromAigSimple(retimed)
	aig.Aig_ManStop(retimed)
	if verbose != 0 {
		fmt.Println("After min-area retiming:")
		gia.Gia_ManPrintStats(converted, nil)
	}
	result := Gia_ManDupFf2In(converted, oldFlops)
	gia.Gia_ManStop(converted)
	if verbose != 0 {
		fmt.Println("After FF-2-PI transformation:")
		gia.Gia_ManPrintStats(result, nil)
	}
	return result
}
