package gia

import "fmt"

// Gia_ManFraigSelectReprs maps collapsed equivalence classes back to the
// original graph while keeping flop and carry-chain connections intact.
func Gia_ManFraigSelectReprs(p, collapsed *Gia_Man_t, verbose int32, flopTypes []int32) []int32 {
	if p.pManTime == nil || len(flopTypes) < 3 {
		panic("gia: invalid FRAIG selection arguments")
	}
	reprs := make([]int32, Gia_ManObjNum(p))
	for i := range reprs {
		reprs[i] = -1
	}
	mapping := make([]int32, Gia_ManObjNum(collapsed))
	for i := range mapping {
		mapping[i] = -1
	}
	var constants, equivalents int32
	for _, obj := range collapsed.pObjs[1:] {
		if Gia_ObjIsCo(obj) != 0 {
			continue
		}
		r := Gia_ObjReprSelf(collapsed, obj.id)
		if r == obj.id {
			continue
		}
		if r == 0 {
			constants++
		} else {
			equivalents++
		}
	}
	if verbose != 0 {
		fmt.Printf("Computed %d const objects and %d other objects.\n", constants, equivalents)
	}
	Gia_ManCleanMark0(p)
	for i := Gia_ManPoNum(p) - Gia_ManRegBoxNum(p); i < Gia_ManPoNum(p); i++ {
		obj := Gia_ObjFanin0(Gia_ManPo(p, i))
		if Gia_ObjIsCi(obj) == 0 {
			panic("gia: invalid flop driver")
		}
		obj.fMark0 = 1
	}
	carry := Gia_ManComputeCarryOuts(p)
	for _, id := range Vec_IntArray(carry) {
		Gia_ManObj(p, id).fMark0 = 1
	}
	if verbose != 0 {
		fmt.Printf("Fixed %d flop inputs and %d box/box connections (out of %d non-flop boxes).\n", Gia_ManRegBoxNum(p), Vec_IntSize(carry), Gia_ManNonRegBoxNum(p))
	}
	clear(flopTypes[:3])
	for i := int32(0); i < Gia_ManRegNum(collapsed); i++ {
		obj := Gia_ManRo(collapsed, i)
		repr := Gia_ObjReprObj(collapsed, obj.id)
		if repr == nil || repr == obj {
			continue
		}
		if repr == Gia_ManConst0(collapsed) {
			flopTypes[0]++
		} else if Gia_ObjIsRo(collapsed, repr) != 0 {
			flopTypes[1]++
		}
	}
	mapping[0] = 0
	Gia_ManSetPhase(collapsed)
	constants, equivalents = 0, 0
	for _, obj := range p.pObjs[1:] {
		if Gia_ObjIsCo(obj) != 0 || Gia_ObjIsCi(obj) != 0 && obj.fMark0 != 0 {
			continue
		}
		if Gia_ObjIsCi(obj) == 0 && Gia_ObjIsAnd(obj) == 0 {
			panic("gia: invalid FRAIG source object")
		}
		lit := Gia_ObjValue(obj)
		if lit == -1 {
			continue
		}
		id := lit >> 1
		r := Gia_ObjReprSelf(collapsed, id)
		if mapping[r] == -1 {
			mapping[r] = obj.id
			continue
		}
		other := mapping[r]
		otherLit := Gia_ObjValue(Gia_ManObj(p, other))
		phase := (lit & 1) ^ (otherLit & 1) ^ int32(Gia_ManObj(collapsed, id).fPhase) ^ int32(Gia_ManObj(collapsed, otherLit>>1).fPhase)
		reprs[obj.id] = other<<1 | phase
		if other >= obj.id {
			panic("gia: representative must precede object")
		}
		if other == 0 {
			constants++
		} else {
			equivalents++
		}
	}
	for i := int32(0); i < Gia_ManCiNum(p); i++ {
		Gia_ManCi(p, i).fMark0 = 0
	}
	if verbose != 0 {
		fmt.Printf("Found %d const objects and %d other objects.\n", constants, equivalents)
	}
	return reprs
}
