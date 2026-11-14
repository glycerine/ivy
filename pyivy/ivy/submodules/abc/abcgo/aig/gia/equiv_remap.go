package gia

import "github.com/glycerine/abcgo/misc/vec"

func Gia_ManEquivFixOutputPairs(p *Gia_Man_t) {
	if Gia_ManPoNum(p)&1 != 0 {
		panic("gia: output pairs must be complete")
	}
	for i := int32(0); i < Gia_ManPoNum(p); i += 2 {
		a, b := Gia_ManPo(p, i), Gia_ManPo(p, i+1)
		if Gia_ObjChild0(a) != Gia_ObjChild0(b) {
			continue
		}
		a.iDiff0 = uint32(a.id)
		a.fCompl0 = 0
		b.iDiff0 = uint32(b.id)
		b.fCompl0 = 0
	}
}
func Gia_ManEquivUpdatePointers(p, q *Gia_Man_t) {
	for _, obj := range p.pObjs {
		lit := Gia_ObjValue(obj)
		if lit >= 0 && Gia_ManObj(q, lit>>1).fMark0 != 0 {
			Gia_ObjSetValue(obj, -1)
		}
	}
}
func giaEquivInitReprs(p *Gia_Man_t) {
	p.PReprs = make([]Gia_Rpr_t, p.nObjs)
	for i := range p.PReprs {
		p.PReprs[i].iRepr = GIA_VOID
	}
}
func giaEquivRemapClasses(src, dst *Gia_Man_t, translate func(int32) (int32, bool)) {
	giaEquivInitReprs(dst)
	for id := int32(1); id < src.nObjs; id++ {
		if Gia_ObjIsConst(src, id) == 0 {
			continue
		}
		mapped, ok := translate(id)
		if ok && mapped != 0 {
			Gia_ObjSetRepr(dst, mapped, 0)
		}
	}
	for head := int32(1); head < src.nObjs; head++ {
		if Gia_ObjIsHead(src, head) == 0 {
			continue
		}
		class := vec.Vec_IntAlloc(4)
		for id := head; id != 0; id = src.PNexts[id] {
			if mapped, ok := translate(id); ok {
				vec.Vec_IntPushUnique(class, mapped)
			}
		}
		if vec.Vec_IntSize(class) < 2 {
			continue
		}
		vec.Vec_IntSort(class, 0)
		members := vec.Vec_IntArray(class)
		for _, id := range members[1:] {
			if id != 0 {
				Gia_ObjSetRepr(dst, id, members[0])
			}
		}
	}
	dst.PNexts = Gia_ManDeriveNexts(dst)
}
func Gia_ManEquivDeriveReprs(p, intermediate, final *Gia_Man_t) {
	giaEquivRemapClasses(p, final, func(id int32) (int32, bool) {
		lit := Gia_ObjValue(Gia_ManObj(p, id))
		if lit < 0 {
			return 0, false
		}
		mapped := Gia_ObjValue(Gia_ManObj(intermediate, lit>>1))
		if mapped < 0 {
			return 0, false
		}
		return mapped >> 1, true
	})
}
func Gia_ManEquivRemapDfs(p *Gia_Man_t) *Gia_Man_t {
	q := Gia_ManDupDfs(p)
	giaEquivRemapClasses(p, q, func(id int32) (int32, bool) { lit := Gia_ObjValue(Gia_ManObj(p, id)); return lit >> 1, lit >= 0 })
	return q
}
func Gia_ManEquivReduceAndRemap(p *Gia_Man_t, seq, miterPairs int32) *Gia_Man_t {
	q := Gia_ManEquivReduce(p, 0, 0, 0, 0)
	if q == nil {
		return nil
	}
	Gia_ManOrigIdsRemap(p, q)
	if miterPairs != 0 {
		Gia_ManEquivFixOutputPairs(q)
	}
	if seq != 0 {
		Gia_ManSeqMarkUsed(q)
	} else {
		Gia_ManCombMarkUsed(q)
	}
	Gia_ManEquivUpdatePointers(p, q)
	final := Gia_ManDupMarked(q)
	Gia_ManOrigIdsRemap(q, final)
	Gia_ManEquivDeriveReprs(p, q, final)
	Gia_ManStop(q)
	q = final
	final = Gia_ManEquivRemapDfs(q)
	Gia_ManOrigIdsRemap(q, final)
	Gia_ManStop(q)
	return final
}

func Gia_ManChangeOrder_rec(q, p *Gia_Man_t, obj *Gia_Obj_t) int32 {
	if lit := Gia_ObjValue(obj); lit != -1 {
		return lit
	}
	if Gia_ObjIsCi(obj) != 0 {
		lit := Gia_ManAppendCi(q)
		Gia_ObjSetValue(obj, lit)
		return lit
	}
	Gia_ManChangeOrder_rec(q, p, Gia_ObjFanin0(obj))
	var lit int32
	if Gia_ObjIsCo(obj) != 0 {
		lit = Gia_ManAppendCo(q, Gia_ObjFanin0Copy(obj))
	} else {
		Gia_ManChangeOrder_rec(q, p, Gia_ObjFanin1(obj))
		lit = Gia_ManAppendAnd(q, Gia_ObjFanin0Copy(obj), Gia_ObjFanin1Copy(obj))
	}
	Gia_ObjSetValue(obj, lit)
	return lit
}
func Gia_ManChangeOrder(p *Gia_Man_t) *Gia_Man_t {
	if p.PReprs == nil || p.PNexts == nil {
		panic("gia: missing equivalence classes")
	}
	Gia_ManFillValue(p)
	q := Gia_ManStart(p.nObjs)
	q.pName = p.pName
	q.PSpec = p.PSpec
	Gia_ObjSetValue(Gia_ManConst0(p), 0)
	for i := int32(0); i < Gia_ManCiNum(p); i++ {
		Gia_ObjSetValue(Gia_ManCi(p, i), Gia_ManAppendCi(q))
	}
	for head := int32(1); head < p.nObjs; head++ {
		if Gia_ObjIsHead(p, head) != 0 {
			for id := head; id != 0; id = p.PNexts[id] {
				Gia_ManChangeOrder_rec(q, p, Gia_ManObj(p, id))
			}
		}
	}
	for id := int32(1); id < p.nObjs; id++ {
		if Gia_ObjIsConst(p, id) != 0 {
			Gia_ManChangeOrder_rec(q, p, Gia_ManObj(p, id))
		}
	}
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		Gia_ManChangeOrder_rec(q, p, Gia_ObjFanin0(Gia_ManCo(p, i)))
	}
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		obj := Gia_ManCo(p, i)
		Gia_ObjSetValue(obj, Gia_ManAppendCo(q, Gia_ObjFanin0Copy(obj)))
	}
	Gia_ManSetRegNum(q, p.nRegs)
	if q.nObjs != p.nObjs {
		panic("gia: object count changed during reordering")
	}
	return q
}
func Gia_ManTransferEquivs(p, q *Gia_Man_t) {
	if p.nObjs != q.nObjs || p.PReprs == nil || p.PNexts == nil || q.PReprs != nil || q.PNexts != nil {
		panic("gia: invalid equivalence transfer")
	}
	giaEquivRemapClasses(p, q, func(id int32) (int32, bool) { lit := Gia_ObjValue(Gia_ManObj(p, id)); return lit >> 1, lit >= 0 })
}
