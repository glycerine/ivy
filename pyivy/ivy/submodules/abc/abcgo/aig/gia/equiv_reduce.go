package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_ManEquivSetColor_rec(p *Gia_Man_t, obj *Gia_Obj_t, odd int32) int32 {
	var count int32
	stack := []*Gia_Obj_t{obj}
	for len(stack) > 0 {
		n := len(stack) - 1
		current := stack[n]
		stack = stack[:n]
		if Gia_ObjVisitColor(p, current.id, odd) != 0 {
			continue
		}
		count++
		if Gia_ObjIsRo(p, current) != 0 {
			stack = append(stack, Gia_ObjFanin0(Gia_ObjRoToRi(p, current)))
			continue
		}
		if Gia_ObjIsAnd(current) == 0 {
			panic("gia: invalid color traversal")
		}
		stack = append(stack, Gia_ObjFanin1(current), Gia_ObjFanin0(current))
	}
	return count
}
func Gia_ManEquivSetColors(p *Gia_Man_t, verbose int32) int32 {
	if Gia_ManPoNum(p)&1 != 0 {
		panic("gia: dual outputs must be paired")
	}
	Gia_ObjSetColors(p, 0)
	for i := int32(0); i < Gia_ManPiNum(p); i++ {
		Gia_ObjSetColors(p, Gia_ManPi(p, i).id)
	}
	nodes := [2]int32{Gia_ManPiNum(p), Gia_ManPiNum(p)}
	for i := int32(0); i < Gia_ManPoNum(p); i++ {
		nodes[i&1] += Gia_ManEquivSetColor_rec(p, Gia_ObjFanin0(Gia_ManPo(p, i)), i&1)
	}
	a, b := Gia_ManCandNum(p)-nodes[0], Gia_ManCandNum(p)-nodes[1]
	if verbose != 0 {
		fmt.Printf("CI+AND = %7d  A = %7d  B = %7d  Ad = %7d  Bd = %7d  AB = %7d.\n", Gia_ManCandNum(p), nodes[0], nodes[1], a, b, Gia_ManCandNum(p)-a-b)
	}
	return (a + b) / 2
}
func Gia_ManEquivRepr(p *Gia_Man_t, obj *Gia_Obj_t, useAll, dual int32) *Gia_Obj_t {
	repr := Gia_ObjRepr(p, obj.id)
	if repr == GIA_VOID || useAll == 0 && Gia_ObjProved(p, obj.id) == 0 {
		return nil
	}
	if dual != 0 && Gia_ObjDiffColors2(p, obj.id, repr) == 0 {
		return nil
	}
	return Gia_ManObj(p, repr)
}
func Gia_ManEquivReduce_rec(q, p *Gia_Man_t, obj *Gia_Obj_t, useAll, dual int32) {
	if repr := Gia_ManEquivRepr(p, obj, useAll, dual); repr != nil {
		Gia_ManEquivReduce_rec(q, p, repr, useAll, dual)
		Gia_ObjSetValue(obj, Gia_ObjValue(repr)^Gia_ObjPhaseReal(repr)^Gia_ObjPhaseReal(obj))
		return
	}
	if Gia_ObjValue(obj) != -1 {
		return
	}
	if Gia_ObjIsAnd(obj) == 0 {
		panic("gia: unassigned reduction input")
	}
	Gia_ManEquivReduce_rec(q, p, Gia_ObjFanin0(obj), useAll, dual)
	Gia_ManEquivReduce_rec(q, p, Gia_ObjFanin1(obj), useAll, dual)
	Gia_ObjSetValue(obj, Gia_ManHashAnd(q, Gia_ObjFanin0Copy(obj), Gia_ObjFanin1Copy(obj)))
}
func giaEquivEnsureRepresentatives(p *Gia_Man_t) bool {
	if p.PReprs == nil && p.pSibls != nil {
		p.PReprs = make([]Gia_Rpr_t, p.nObjs)
		mapping := make([]int32, p.nObjs)
		for i := range mapping {
			mapping[i] = -1
			p.PReprs[i].iRepr = GIA_VOID
		}
		for i, sibling := range p.pSibls {
			if sibling > 0 {
				if mapping[sibling] < 0 {
					mapping[sibling] = sibling
				}
				mapping[i] = mapping[sibling]
			}
		}
		for i, sibling := range p.pSibls {
			if sibling > 0 {
				Gia_ObjSetRepr(p, int32(i), mapping[i])
			}
		}
		p.PNexts = Gia_ManDeriveNexts(p)
	}
	return p.PReprs != nil
}
func giaEquivReductionStart(p *Gia_Man_t) *Gia_Man_t {
	q := Gia_ManStart(p.nObjs)
	q.pName = p.pName
	q.PSpec = p.PSpec
	Gia_ManFillValue(p)
	Gia_ObjSetValue(Gia_ManConst0(p), 0)
	for i := int32(0); i < Gia_ManCiNum(p); i++ {
		Gia_ObjSetValue(Gia_ManCi(p, i), Gia_ManAppendCi(q))
	}
	Gia_ManHashAlloc(q)
	return q
}
func giaEquivReductionFinish(p, q *Gia_Man_t) *Gia_Man_t {
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		obj := Gia_ManCo(p, i)
		Gia_ObjSetValue(obj, Gia_ManAppendCo(q, Gia_ObjFanin0Copy(obj)))
	}
	Gia_ManHashStop(q)
	Gia_ManSetRegNum(q, p.nRegs)
	return q
}
func Gia_ManEquivReduce(p *Gia_Man_t, useAll, dual, skipPhase, verbose int32) *Gia_Man_t {
	if !giaEquivEnsureRepresentatives(p) {
		fmt.Println("Gia_ManEquivReduce(): Equivalence classes are not available.")
		return nil
	}
	if dual != 0 && Gia_ManPoNum(p)&1 != 0 {
		fmt.Println("Gia_ManEquivReduce(): Dual-output miter should have even number of POs.")
		return nil
	}
	any := false
	for i := int32(0); i < p.nObjs; i++ {
		if Gia_ObjRepr(p, i) != GIA_VOID {
			any = true
			break
		}
	}
	if !any {
		return Gia_ManDup(p)
	}
	if skipPhase == 0 {
		Gia_ManSetPhase(p)
	}
	if dual != 0 {
		Gia_ManEquivSetColors(p, verbose)
	}
	q := giaEquivReductionStart(p)
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		Gia_ManEquivReduce_rec(q, p, Gia_ObjFanin0(Gia_ManCo(p, i)), useAll, dual)
	}
	return giaEquivReductionFinish(p, q)
}
func Gia_ManEquivReduce2_rec(q, p *Gia_Man_t, obj *Gia_Obj_t, mapping *Vec_Int_t, dive int32) {
	if Gia_ObjValue(obj) != -1 {
		return
	}
	if Gia_ObjIsAnd(obj) == 0 {
		panic("gia: unassigned reduction input")
	}
	if repr := Gia_ManEquivRepr(p, obj, 1, 0); dive != 0 && repr != nil {
		if repr.id == 0 {
			Gia_ObjSetValue(obj, Gia_ObjPhaseReal(obj))
			return
		}
		head := repr.id
		best := Gia_ManObj(p, vec.Vec_IntEntry(mapping, int(head)))
		Gia_ManEquivReduce2_rec(q, p, best, mapping, 0)
		for id := head; id != 0; id = p.PNexts[id] {
			member := Gia_ManObj(p, id)
			Gia_ObjSetValue(member, Gia_ObjValue(best)^Gia_ObjPhaseReal(best)^Gia_ObjPhaseReal(member))
		}
		return
	}
	Gia_ManEquivReduce2_rec(q, p, Gia_ObjFanin0(obj), mapping, 1)
	Gia_ManEquivReduce2_rec(q, p, Gia_ObjFanin1(obj), mapping, 1)
	Gia_ObjSetValue(obj, Gia_ManHashAnd(q, Gia_ObjFanin0Copy(obj), Gia_ObjFanin1Copy(obj)))
}
func Gia_ManEquivReduce2(p *Gia_Man_t) *Gia_Man_t {
	if !giaEquivEnsureRepresentatives(p) {
		fmt.Println("Gia_ManEquivReduce(): Equivalence classes are not available.")
		return nil
	}
	any := false
	for i := int32(0); i < p.nObjs; i++ {
		if Gia_ObjRepr(p, i) != GIA_VOID {
			any = true
			break
		}
	}
	if !any {
		return Gia_ManDup(p)
	}
	mapping := Gia_ManChoiceMinLevel(p)
	Gia_ManSetPhase(p)
	q := giaEquivReductionStart(p)
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		Gia_ManEquivReduce2_rec(q, p, Gia_ObjFanin0(Gia_ManCo(p, i)), mapping, 1)
	}
	return giaEquivReductionFinish(p, q)
}
