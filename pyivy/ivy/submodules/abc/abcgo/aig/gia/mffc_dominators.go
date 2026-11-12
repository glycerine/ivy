package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_ObjDom(p *Gia_Man_t, obj *Gia_Obj_t) int32 { return vec.Vec_IntEntry(p.vDoms, int(obj.id)) }
func Gia_ObjSetDom(p *Gia_Man_t, obj *Gia_Obj_t, d int32) {
	vec.Vec_IntWriteEntry(p.vDoms, int(obj.id), d)
}
func Gia_ManAddDom(p *Gia_Man_t, obj *Gia_Obj_t, d0 int32) {
	d1 := Gia_ObjDom(p, obj)
	if d1 == -1 {
		Gia_ObjSetDom(p, obj, d0)
		return
	}
	for d0 != d1 {
		if d0 > d1 {
			next := Gia_ObjDom(p, Gia_ManObj(p, d1))
			if next == d1 {
				Gia_ObjSetDom(p, obj, obj.id)
				return
			}
			d1 = next
		} else {
			next := Gia_ObjDom(p, Gia_ManObj(p, d0))
			if next == d0 {
				Gia_ObjSetDom(p, obj, obj.id)
				return
			}
			d0 = next
		}
	}
	Gia_ObjSetDom(p, obj, d0)
}
func Gia_ManComputeDoms(p *Gia_Man_t) {
	if p.vDoms == nil {
		p.vDoms = vec.Vec_IntAlloc(0)
	}
	vec.Vec_IntFill(p.vDoms, int(p.nObjs), -1)
	for i := p.nObjs - 1; i > 0; i-- {
		obj := Gia_ManObj(p, i)
		if Gia_ObjIsCi(obj) != 0 {
			continue
		}
		if Gia_ObjIsCo(obj) != 0 {
			Gia_ObjSetDom(p, obj, i)
			Gia_ManAddDom(p, Gia_ObjFanin0(obj), i)
			continue
		}
		if Gia_ObjIsAnd(obj) == 0 {
			panic("gia: expected AND")
		}
		Gia_ManAddDom(p, Gia_ObjFanin0(obj), i)
		Gia_ManAddDom(p, Gia_ObjFanin1(obj), i)
	}
}
func giaMffcSupportRefs(p *Gia_Man_t, obj *Gia_Obj_t, supp, refs *Vec_Int_t) {
	if Gia_ObjIsTravIdCurrent(p, obj) != 0 {
		return
	}
	Gia_ObjSetTravIdCurrent(p, obj)
	if Gia_ObjRefNum(p, obj) != 0 || Gia_ObjIsCi(obj) != 0 {
		vec.Vec_IntPush(supp, obj.id)
		vec.Vec_IntPush(refs, Gia_ObjRefNum(p, obj))
		return
	}
	if Gia_ObjIsAnd(obj) == 0 {
		panic("gia: expected AND")
	}
	giaMffcSupportRefs(p, Gia_ObjFanin0(obj), supp, refs)
	giaMffcSupportRefs(p, Gia_ObjFanin1(obj), supp, refs)
}
func Gia_NodeMffcSizeSuppRefs(p *Gia_Man_t, obj *Gia_Obj_t, supp, refs *Vec_Int_t) int32 {
	vec.Vec_IntClear(supp)
	vec.Vec_IntClear(refs)
	Gia_ManIncrementTravId(p)
	count := Gia_NodeDeref_rec(p, obj)
	giaMffcSupportRefs(p, Gia_ObjFanin0(obj), supp, refs)
	giaMffcSupportRefs(p, Gia_ObjFanin1(obj), supp, refs)
	if count != Gia_NodeRef_rec(p, obj, 0) {
		panic("gia: inconsistent MFFC references")
	}
	for i, id := range vec.Vec_IntArray(supp) {
		vec.Vec_IntAddToEntry(refs, i, -Gia_ObjRefNumId(p, id))
	}
	return count
}
func Gia_ManDomDerive_rec(dst, p *Gia_Man_t, obj *Gia_Obj_t) int32 {
	if Gia_ObjIsTravIdCurrent(p, obj) != 0 {
		return int32(obj.Value)
	}
	Gia_ObjSetTravIdCurrent(p, obj)
	if Gia_ObjIsAnd(obj) == 0 {
		panic("gia: expected AND")
	}
	Gia_ManDomDerive_rec(dst, p, Gia_ObjFanin0(obj))
	Gia_ManDomDerive_rec(dst, p, Gia_ObjFanin1(obj))
	lit := Gia_ManHashAnd(dst, Gia_ObjFanin0Copy(obj), Gia_ObjFanin1Copy(obj))
	obj.Value = uint32(lit)
	return lit
}
func Gia_ManDomDerive(p *Gia_Man_t, root *Gia_Obj_t, supp *Vec_Int_t, nVars int32) *Gia_Man_t {
	if nVars < 0 || nVars > 5 || int(nVars) > vec.Vec_IntSize(supp) {
		panic("gia: invalid dominator variables")
	}
	dst := Gia_ManStart(p.nObjs)
	dst.CopyIdentityFrom(p)
	p.pObjs[0].Value = 0
	Gia_ManHashAlloc(dst)
	for range vec.Vec_IntArray(supp) {
		Gia_ManAppendCi(dst)
	}
	for m := int32(0); m < 1<<nVars; m++ {
		Gia_ManIncrementTravId(p)
		for i, id := range vec.Vec_IntArray(supp) {
			obj := Gia_ManObj(p, id)
			if int32(i) < nVars {
				obj.Value = uint32(m >> i & 1)
			} else {
				obj.Value = uint32(Gia_ObjToLit(dst, Gia_ManCi(dst, int32(i))))
			}
			Gia_ObjSetTravIdCurrent(p, obj)
		}
		Gia_ManAppendCo(dst, Gia_ManDomDerive_rec(dst, p, root))
	}
	Gia_ManHashStop(dst)
	clean := Gia_ManCleanup(dst)
	Gia_ManStop(dst)
	return clean
}
func Gia_ManComputeDomsTry(p *Gia_Man_t) {
	start := util.Abc_Clock()
	p.Refs = nil
	Gia_ManLevelNum(p)
	Gia_ManCreateRefs(p)
	Gia_ManComputeDoms(p)
	util.Abc_PrintTime(1, "Time", util.Abc_Clock()-start)
	supp, refs := vec.Vec_IntAlloc(1000), vec.Vec_IntAlloc(1000)
	for _, obj := range p.pObjs {
		if Gia_ObjIsAnd(obj) == 0 && Gia_ObjIsCo(obj) == 0 {
			continue
		}
		id := obj.id
		if Gia_ObjDom(p, obj) != id {
			continue
		}
		if Gia_ObjIsCo(obj) != 0 {
			obj = Gia_ObjFanin0(obj)
		}
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		size := Gia_NodeMffcSizeSuppRefs(p, obj, supp, refs)
		if size < 10 {
			continue
		}
		vec.Vec_IntSelectSortCost2(vec.Vec_IntArray(supp), vec.Vec_IntSize(supp), vec.Vec_IntArray(refs))
		fmt.Printf("Obj %6d : Cone = %4d  Supp = %4d  \n", id, size, vec.Vec_IntSize(supp))
		derived := Gia_ManDomDerive(p, obj, supp, 0)
		_ = Gia_DumpAiger(derived, "mffc", id, 6)
		Gia_ManCollapseTestTest(derived)
		Gia_ManStop(derived)
	}
}
