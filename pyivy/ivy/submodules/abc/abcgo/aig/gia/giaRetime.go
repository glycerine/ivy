package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

func Gia_ManMarkAutonomous_rec(p *Gia_Man_t, obj *Gia_Obj_t) int32 {
	if Gia_ObjIsTravIdCurrent(p, obj) != 0 {
		return int32(obj.fMark0)
	}
	Gia_ObjSetTravIdCurrent(p, obj)
	if obj.fMark0 != 0 {
		panic("gia: autonomous mark already set")
	}
	if Gia_ObjIsPi(p, obj) != 0 || Gia_ObjIsConst0(obj) != 0 {
		obj.fMark0 = 1
		return 1
	}
	var value int32
	if Gia_ObjIsCo(obj) != 0 {
		value = Gia_ManMarkAutonomous_rec(p, Gia_ObjFanin0(obj))
	} else if Gia_ObjIsCi(obj) != 0 {
		value = Gia_ManMarkAutonomous_rec(p, Gia_ObjRoToRi(p, obj))
	} else {
		if Gia_ObjIsAnd(obj) == 0 {
			panic("gia: expected AND")
		}
		value = Gia_ManMarkAutonomous_rec(p, Gia_ObjFanin0(obj))
		if value == 0 {
			value = Gia_ManMarkAutonomous_rec(p, Gia_ObjFanin1(obj))
		}
	}
	obj.fMark0 = uint32(value)
	return value
}
func Gia_ManMarkAutonomous(p *Gia_Man_t) {
	Gia_ManCleanMark0(p)
	Gia_ManIncrementTravId(p)
	for i := int32(0); i < Gia_ManRegNum(p); i++ {
		Gia_ManMarkAutonomous_rec(p, Gia_ManRo(p, i))
	}
	Gia_ManIncrementTravId(p)
	for i := int32(0); i < Gia_ManRegNum(p); i++ {
		obj := Gia_ManRo(p, i)
		if obj.fMark0 != 0 {
			Gia_ObjSetTravIdCurrent(p, obj)
		}
	}
	Gia_ManCleanMark0(p)
}
func Gia_ManRetimeDup_rec(dst *Gia_Man_t, obj *Gia_Obj_t) {
	if obj.Value != ^uint32(0) {
		return
	}
	if Gia_ObjIsAnd(obj) == 0 {
		panic("gia: retiming encountered unassigned input")
	}
	Gia_ManRetimeDup_rec(dst, Gia_ObjFanin0(obj))
	Gia_ManRetimeDup_rec(dst, Gia_ObjFanin1(obj))
	obj.Value = uint32(Gia_ManHashAnd(dst, Gia_ObjFanin0Copy(obj), Gia_ObjFanin1Copy(obj)))
}
func Gia_ManRetimeDupForward(p *Gia_Man_t, cut []*Gia_Obj_t) *Gia_Man_t {
	dst := Gia_ManStart(p.nObjs)
	dst.CopyIdentityFrom(p)
	Gia_ManHashAlloc(dst)
	Gia_ManFillValue(p)
	Gia_ManSetPhase(p)
	p.pObjs[0].Value = 0
	for i := int32(0); i < Gia_ManPiNum(p); i++ {
		Gia_ManPi(p, i).Value = uint32(Gia_ManAppendCi(dst))
	}
	for _, obj := range cut {
		obj.Value = uint32(Gia_ManAppendCi(dst)) ^ obj.fPhase
	}
	for _, id := range vec.Vec_IntArray(p.vCos) {
		Gia_ManRetimeDup_rec(dst, Gia_ObjFanin0(Gia_ManObj(p, id)))
	}
	for i := int32(0); i < Gia_ManPoNum(p); i++ {
		Gia_ManAppendCo(dst, Gia_ObjFanin0Copy(Gia_ManPo(p, i)))
	}
	for i := int32(0); i < Gia_ManRegNum(p); i++ {
		ri := Gia_ManRi(p, i)
		ri.Value = uint32(Gia_ObjFanin0Copy(ri))
	}
	for i := int32(0); i < Gia_ManRegNum(p); i++ {
		Gia_ManRo(p, i).Value = Gia_ManRi(p, i).Value
	}
	for _, obj := range cut {
		if Gia_ObjIsAnd(obj) != 0 {
			obj.Value = ^uint32(0)
		}
	}
	for _, obj := range cut {
		Gia_ManRetimeDup_rec(dst, obj)
		Gia_ManAppendCo(dst, int32(obj.Value^obj.fPhase))
	}
	Gia_ManHashStop(dst)
	Gia_ManSetRegNum(dst, int32(len(cut)))
	result := Gia_ManCleanup(dst)
	Gia_ManStop(dst)
	return result
}
func Gia_ManRetimeForwardOne(p *Gia_Man_t, fixed, moves *int32) *Gia_Man_t {
	var classes []int32
	var newClasses *Vec_Int_t
	if p.VFlopClasses != nil {
		classes = make([]int32, p.nObjs)
		for i := range classes {
			classes[i] = -1
		}
		for i := int32(0); i < Gia_ManRegNum(p); i++ {
			classes[Gia_ManRo(p, i).id] = vec.Vec_IntEntry(p.VFlopClasses, int(i))
		}
		newClasses = vec.Vec_IntAlloc(int(Gia_ManRegNum(p)))
	}
	Gia_ManIncrementTravId(p)
	Gia_ManMarkAutonomous(p)
	Gia_ManIncrementTravId(p)
	*fixed = 0
	for i := int32(0); i < Gia_ManRegNum(p); i++ {
		obj := Gia_ManRo(p, i)
		if Gia_ObjIsTravIdPrevious(p, obj) != 0 {
			Gia_ObjSetTravIdCurrent(p, obj)
		} else {
			*fixed++
		}
	}
	*moves = 0
	for _, obj := range p.pObjs {
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		a, b := Gia_ObjFanin0(obj), Gia_ObjFanin1(obj)
		if Gia_ObjIsTravIdCurrent(p, a) == 0 || Gia_ObjIsTravIdCurrent(p, b) == 0 {
			continue
		}
		if classes != nil {
			if classes[a.id] != classes[b.id] {
				continue
			}
			classes[obj.id] = classes[a.id]
		}
		Gia_ObjSetTravIdCurrent(p, obj)
		*moves++
	}
	for i := int32(0); i < Gia_ManRegNum(p); i++ {
		Gia_ObjSetTravIdCurrent(p, Gia_ManRo(p, i))
	}
	Gia_ManIncrementTravId(p)
	var cut []*Gia_Obj_t
	add := func(obj *Gia_Obj_t) {
		if newClasses != nil {
			vec.Vec_IntPush(newClasses, classes[obj.id])
		}
		cut = append(cut, obj)
		Gia_ObjSetTravIdCurrent(p, obj)
	}
	for _, obj := range p.pObjs {
		if Gia_ObjIsTravIdPrevious(p, obj) != 0 {
			continue
		}
		if (Gia_ObjIsCo(obj) != 0 || Gia_ObjIsAnd(obj) != 0) && Gia_ObjIsTravIdPrevious(p, Gia_ObjFanin0(obj)) != 0 {
			add(Gia_ObjFanin0(obj))
		}
		if Gia_ObjIsAnd(obj) != 0 && Gia_ObjIsTravIdPrevious(p, Gia_ObjFanin1(obj)) != 0 {
			add(Gia_ObjFanin1(obj))
		}
	}
	result := Gia_ManRetimeDupForward(p, cut)
	result.VFlopClasses = newClasses
	return result
}
func Gia_ManRetimeForward(p *Gia_Man_t, maxIters, verbose int32) *Gia_Man_t {
	result := p
	var fixed int32
	moves := int32(1)
	for i := int32(0); i < maxIters && moves > 0; i++ {
		start := time.Now()
		previous := result
		result = Gia_ManRetimeForwardOne(previous, &fixed, &moves)
		if verbose != 0 {
			fmt.Printf("%2d : And = %6d. Reg = %5d. Unret = %5d. Move = %6d. Time = %.2f sec\n", i+1, Gia_ManAndNum(previous), Gia_ManRegNum(previous), fixed, moves, time.Since(start).Seconds())
		}
		if previous != p {
			Gia_ManStop(previous)
		}
	}
	return result
}
