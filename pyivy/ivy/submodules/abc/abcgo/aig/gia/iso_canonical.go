package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"sort"
)

func Gia_ObjCompareByValue(a, b **Gia_Obj_t) int32 {
	av, bv := int32((*a).Value), int32((*b).Value)
	if av < bv {
		return -1
	}
	if av > bv {
		return 1
	}
	return 0
}
func Gia_ManFindCaninicalOrder_rec(p *Gia_Man_t, obj *Gia_Obj_t, ands *Vec_Int_t) {
	if Gia_ObjIsTravIdCurrent(p, obj) != 0 {
		return
	}
	Gia_ObjSetTravIdCurrent(p, obj)
	if Gia_ObjIsAnd(obj) == 0 {
		panic("gia: canonical traversal reached unmarked terminal")
	}
	a, b := Gia_ObjFanin0(obj), Gia_ObjFanin1(obj)
	if Gia_ObjIsAnd(a) != 0 && Gia_ObjIsAnd(b) != 0 {
		if a.Value == b.Value {
			panic("gia: canonical fanins do not have unique labels")
		}
		if a.Value > b.Value {
			a, b = b, a
		}
	}
	Gia_ManFindCaninicalOrder_rec(p, a, ands)
	Gia_ManFindCaninicalOrder_rec(p, b, ands)
	vec.Vec_IntPush(ands, Gia_ObjId(p, obj))
}

// Gia_ManFindCaninicalOrder derives PI, internal and output order from the
// unique labels assigned by isomorphism refinement.
func Gia_ManFindCaninicalOrder(p *Gia_Man_t, cis, ands, cos *Vec_Int_t, permutation **Vec_Int_t) {
	vec.Vec_IntClear(cis)
	vec.Vec_IntClear(ands)
	vec.Vec_IntClear(cos)
	inputs := make([]*Gia_Obj_t, Gia_ManPiNum(p))
	for i := range inputs {
		inputs[i] = Gia_ManCi(p, int32(i))
	}
	sort.SliceStable(inputs, func(i, j int) bool { return int32(inputs[i].Value) < int32(inputs[j].Value) })
	if permutation != nil {
		*permutation = vec.Vec_IntAlloc(len(inputs))
	}
	for _, obj := range inputs {
		vec.Vec_IntPush(cis, Gia_ObjId(p, obj))
		if permutation != nil {
			vec.Vec_IntPush(*permutation, Gia_ObjCioId(obj))
		}
	}
	outputs := make([]*Gia_Obj_t, Gia_ManPoNum(p))
	for i := range outputs {
		obj := Gia_ManPo(p, int32(i))
		outputs[i] = obj
		if len(outputs) > 1 {
			obj.Value = Gia_ObjFanin0(obj).Value*2 + uint32(Gia_ObjFaninC0(obj))
		}
	}
	sort.SliceStable(outputs, func(i, j int) bool { return int32(outputs[i].Value) < int32(outputs[j].Value) })
	for _, obj := range outputs {
		vec.Vec_IntPush(cos, Gia_ObjId(p, obj))
	}
	registers := make([]*Gia_Obj_t, Gia_ManRegNum(p))
	for i := range registers {
		registers[i] = Gia_ManRo(p, int32(i))
	}
	sort.SliceStable(registers, func(i, j int) bool { return int32(registers[i].Value) < int32(registers[j].Value) })
	for _, obj := range registers {
		vec.Vec_IntPush(cis, Gia_ObjId(p, obj))
		vec.Vec_IntPush(cos, Gia_ObjId(p, Gia_ObjRoToRi(p, obj)))
	}
	Gia_ManIncrementTravId(p)
	Gia_ObjSetTravIdCurrent(p, Gia_ManConst0(p))
	for _, id := range vec.Vec_IntArray(cis) {
		Gia_ObjSetTravIdCurrent(p, Gia_ManObj(p, id))
	}
	for _, id := range vec.Vec_IntArray(cos) {
		Gia_ManFindCaninicalOrder_rec(p, Gia_ObjFanin0(Gia_ManObj(p, id)), ands)
	}
}
func Gia_ManIsoCanonicize(p *Gia_Man_t, verbose int32) *Gia_Man_t {
	if Gia_ManCiNum(p) == 0 {
		if Gia_ManPoNum(p) != 1 || Gia_ManObjNum(p) != 2 {
			panic("gia: invalid constant AIG")
		}
		return Gia_ManDup(p)
	}
	Gia_IsoDeriveEquivPos(p, 0, verbose)
	cis, ands, cos := vec.Vec_IntAlloc(int(Gia_ManCiNum(p))), vec.Vec_IntAlloc(int(Gia_ManAndNum(p))), vec.Vec_IntAlloc(int(Gia_ManCoNum(p)))
	Gia_ManFindCaninicalOrder(p, cis, ands, cos, nil)
	return Gia_ManDupFromVecs(p, cis, ands, cos, Gia_ManRegNum(p))
}
func Gia_ManIsoFindString(p *Gia_Man_t, output, verbose int32, permutation **Vec_Int_t) *Vec_Str_t {
	part := Gia_ManDupCones(p, []int32{output}, 1, 1)
	if part == nil {
		return nil
	}
	defer Gia_ManStop(part)
	if Gia_ManPoNum(part) != 1 {
		panic("gia: canonical cone must have one output")
	}
	if Gia_ManCiNum(part) == 0 {
		if permutation != nil {
			*permutation = vec.Vec_IntAlloc(0)
		}
		encoded, err := Gia_AigerWriteIntoMemoryStr(part)
		if err != nil {
			return nil
		}
		return encoded
	}
	Gia_IsoDeriveEquivPos(part, 0, verbose)
	cis, ands, cos := vec.Vec_IntAlloc(int(Gia_ManCiNum(part))), vec.Vec_IntAlloc(int(Gia_ManAndNum(part))), vec.Vec_IntAlloc(int(Gia_ManCoNum(part)))
	Gia_ManFindCaninicalOrder(part, cis, ands, cos, permutation)
	encoded, err := Gia_AigerWriteIntoMemoryStrPart(part, cis, ands, cos, Gia_ManRegNum(part))
	if err != nil {
		return nil
	}
	return encoded
}
func Vec_IntCountNonTrivial(groups *Vec_Ptr_t, used *int32) int32 {
	*used = 0
	var count int32
	for _, entry := range vec.Vec_PtrArray(groups) {
		size := vec.Vec_IntSize(entry.(*Vec_Int_t))
		if size >= 2 {
			count++
			*used += int32(size)
		}
	}
	return count
}
