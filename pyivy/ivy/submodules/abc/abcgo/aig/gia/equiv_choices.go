package gia

import "github.com/glycerine/abcgo/misc/vec"

func Gia_ObjCheckTfi_rec(p *Gia_Man_t, old, node *Gia_Obj_t, visited *Vec_Ptr_t) int32 {
	if node == nil || Gia_ObjIsCi(node) != 0 {
		return 0
	}
	if node == old {
		return 1
	}
	if node.fMark0 != 0 {
		return 0
	}
	node.fMark0 = 1
	vec.Vec_PtrPush(visited, node)
	if Gia_ObjCheckTfi_rec(p, old, Gia_ObjFanin0(node), visited) != 0 || Gia_ObjCheckTfi_rec(p, old, Gia_ObjFanin1(node), visited) != 0 {
		return 1
	}
	return Gia_ObjCheckTfi_rec(p, old, Gia_ObjNextObj(p, node.id), visited)
}
func Gia_ObjCheckTfi(p *Gia_Man_t, old, node *Gia_Obj_t) int32 {
	if Gia_IsComplement(old) != 0 || Gia_IsComplement(node) != 0 {
		panic("gia: complemented choice node")
	}
	visited := vec.Vec_PtrAlloc(100)
	result := Gia_ObjCheckTfi_rec(p, old, node, visited)
	for i := 0; i < vec.Vec_PtrSize(visited); i++ {
		vec.Vec_PtrEntry(visited, i).(*Gia_Obj_t).fMark0 = 0
	}
	return result
}
func Gia_ManAddNextEntry_rec(p *Gia_Man_t, old, node *Gia_Obj_t) {
	for p.PNexts[old.id] != 0 {
		old = Gia_ManObj(p, p.PNexts[old.id])
	}
	Gia_ObjSetNext(p, old.id, node.id)
}
func Gia_ManEquivToChoices_rec(q, p *Gia_Man_t, obj *Gia_Obj_t) {
	if Gia_ObjValue(obj) != -1 {
		return
	}
	repr := Gia_ObjReprObj(p, obj.id)
	if repr != nil {
		if Gia_ObjIsConst0(repr) != 0 {
			Gia_ObjSetValue(obj, Gia_ObjValue(repr)^Gia_ObjPhaseReal(repr)^Gia_ObjPhaseReal(obj))
			return
		}
		Gia_ManEquivToChoices_rec(q, p, repr)
	}
	if Gia_ObjIsAnd(obj) == 0 {
		panic("gia: choice object must be an AND")
	}
	Gia_ManEquivToChoices_rec(q, p, Gia_ObjFanin0(obj))
	Gia_ManEquivToChoices_rec(q, p, Gia_ObjFanin1(obj))
	Gia_ObjSetValue(obj, Gia_ManHashAnd(q, Gia_ObjFanin0Copy(obj), Gia_ObjFanin1Copy(obj)))
	if repr == nil {
		return
	}
	expected := Gia_ObjValue(repr) ^ Gia_ObjPhaseReal(repr) ^ Gia_ObjPhaseReal(obj)
	if Gia_ObjValue(obj)>>1 == Gia_ObjValue(repr)>>1 {
		if Gia_ObjValue(obj) != expected {
			panic("gia: inconsistent equivalent phases")
		}
		return
	}
	if Gia_ObjValue(repr) > Gia_ObjValue(obj) {
		return
	}
	old, newObj := Gia_ManObj(q, Gia_ObjValue(repr)>>1), Gia_ManObj(q, Gia_ObjValue(obj)>>1)
	if existing := Gia_ObjReprObj(q, newObj.id); existing != nil {
		if existing == old {
			Gia_ObjSetValue(obj, expected)
		}
		return
	}
	if Gia_ObjCheckTfi(q, old, newObj) == 0 {
		if Gia_ObjNext(q, newObj.id) != 0 {
			panic("gia: choice already linked")
		}
		Gia_ObjSetRepr(q, newObj.id, old.id)
		Gia_ManAddNextEntry_rec(q, old, newObj)
	}
	Gia_ObjSetValue(obj, expected)
}
func Gia_ManRemoveBadChoices(p *Gia_Man_t) {
	for _, obj := range p.pObjs {
		obj.fMark0 = 0
		if Gia_ObjIsAnd(obj) != 0 {
			Gia_ObjFanin0(obj).fMark0 = 1
			Gia_ObjFanin1(obj).fMark0 = 1
		} else if Gia_ObjIsCo(obj) != 0 {
			Gia_ObjFanin0(obj).fMark0 = 1
		}
	}
	for head := int32(1); head < p.nObjs; head++ {
		if Gia_ObjIsHead(p, head) == 0 {
			continue
		}
		prev := head
		for id := p.PNexts[prev]; id != 0; id = p.PNexts[prev] {
			if Gia_ManObj(p, id).fMark0 == 0 {
				prev = id
				continue
			}
			Gia_ObjUnsetRepr(p, id)
			Gia_ObjSetNext(p, prev, p.PNexts[id])
			Gia_ObjSetNext(p, id, 0)
		}
	}
	Gia_ManCleanMark0(p)
}
func Gia_ManEquivToChoices(p *Gia_Man_t, snapshots int32) *Gia_Man_t {
	if snapshots <= 0 || Gia_ManCoNum(p)%snapshots != 0 {
		panic("gia: invalid snapshot count")
	}
	Gia_ManSetPhase(p)
	q := Gia_ManStart(p.nObjs)
	q.pName = p.pName
	q.PSpec = p.PSpec
	q.PReprs = make([]Gia_Rpr_t, p.nObjs)
	q.PNexts = make([]int32, p.nObjs)
	for i := range q.PReprs {
		q.PReprs[i].iRepr = GIA_VOID
	}
	Gia_ManFillValue(p)
	Gia_ObjSetValue(Gia_ManConst0(p), 0)
	for i := int32(0); i < Gia_ManCiNum(p); i++ {
		Gia_ObjSetValue(Gia_ManCi(p, i), Gia_ManAppendCi(q))
	}
	for i := int32(0); i < p.nRegs; i++ {
		obj := Gia_ManRo(p, i)
		if repr := Gia_ObjReprObj(p, obj.id); repr != nil {
			if Gia_ObjIsConst0(repr) == 0 && Gia_ObjIsRo(p, repr) == 0 {
				panic("gia: invalid register representative")
			}
			Gia_ObjSetValue(obj, Gia_ObjValue(repr))
		}
	}
	Gia_ManHashAlloc(q)
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		Gia_ManEquivToChoices_rec(q, p, Gia_ObjFanin0(Gia_ManCo(p, i)))
	}
	for _, id := range vec.Vec_IntArray(Gia_ManGetDangling(p)) {
		Gia_ManEquivToChoices_rec(q, p, Gia_ManObj(p, id))
	}
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		if i%snapshots == 0 {
			Gia_ManAppendCo(q, Gia_ObjFanin0Copy(Gia_ManCo(p, i)))
		}
	}
	Gia_ManHashStop(q)
	Gia_ManSetRegNum(q, p.nRegs)
	Gia_ManRemoveBadChoices(q)
	q.PReprs = q.PReprs[:q.nObjs]
	q.PNexts = q.PNexts[:q.nObjs]
	result := Gia_ManCleanup(q)
	Gia_ManStop(q)
	return result
}
func Gia_ManCountChoiceNodes(p *Gia_Man_t) int32 {
	if p.PReprs == nil || p.PNexts == nil {
		return 0
	}
	var n int32
	for id := int32(0); id < p.nObjs; id++ {
		n += Gia_ObjIsHead(p, id)
	}
	return n
}
func Gia_ManCountChoices(p *Gia_Man_t) int32 {
	if p.PReprs == nil || p.PNexts == nil {
		return 0
	}
	var n int32
	for _, next := range p.PNexts {
		if next > 0 {
			n++
		}
	}
	return n
}
func Gia_ManHasNoEquivs(p *Gia_Man_t) int32 {
	if p.PReprs == nil {
		return 1
	}
	for i := int32(0); i < p.nObjs; i++ {
		if Gia_ObjRepr(p, i) != GIA_VOID {
			return 0
		}
	}
	return 1
}
