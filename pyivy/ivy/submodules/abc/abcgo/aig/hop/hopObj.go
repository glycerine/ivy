// Object lifecycle, ported from ABC hopObj.c (Alan Mishchenko).
package hop

func Hop_ObjCreatePi(p *Hop_Man_t) *Hop_Obj_t {
	o := Hop_ManFetchMemory(p)
	o.Type = AIG_PI
	o.PioNum = Vec_PtrSize(p.vPis)
	Vec_PtrPush(p.vPis, o)
	p.nObjs[AIG_PI]++
	return o
}
func Hop_ObjCreatePo(p *Hop_Man_t, driver *Hop_Obj_t) *Hop_Obj_t {
	o := Hop_ManFetchMemory(p)
	o.Type = AIG_PO
	o.PioNum = Vec_PtrSize(p.vPos)
	Vec_PtrPush(p.vPos, o)
	o.pFanin0 = driver
	if p.fRefCount != 0 {
		Hop_ObjRef(Hop_Regular(driver))
	} else {
		o.nRefs = Hop_ObjLevel(Hop_Regular(driver))
	}
	o.fPhase = Hop_ObjPhaseCompl(driver)
	p.nObjs[AIG_PO]++
	return o
}
func Hop_ObjCreate(p *Hop_Man_t, ghost *Hop_Obj_t) *Hop_Obj_t {
	typ, a, b := ghost.Type, ghost.pFanin0, ghost.pFanin1
	o := Hop_ManFetchMemory(p)
	o.Type = typ
	Hop_ObjConnect(p, o, a, b)
	p.nObjs[typ]++
	return o
}
func Hop_ObjConnect(p *Hop_Man_t, o, a, b *Hop_Obj_t) {
	o.pFanin0, o.pFanin1 = a, b
	if p.fRefCount != 0 {
		if a != nil {
			Hop_ObjRef(Hop_Regular(a))
		}
		if b != nil {
			Hop_ObjRef(Hop_Regular(b))
		}
	} else {
		o.nRefs = Hop_ObjLevelNew(o)
	}
	o.fPhase = Hop_ObjPhaseCompl(a) & Hop_ObjPhaseCompl(b)
	Hop_TableInsert(p, o)
}
func Hop_ObjDisconnect(p *Hop_Man_t, o *Hop_Obj_t) {
	if o.pFanin0 != nil {
		Hop_ObjDeref(Hop_ObjFanin0(o))
	}
	if o.pFanin1 != nil {
		Hop_ObjDeref(Hop_ObjFanin1(o))
	}
	Hop_TableDelete(p, o)
	o.pFanin0, o.pFanin1 = nil, nil
}
func Hop_ObjDelete(p *Hop_Man_t, o *Hop_Obj_t) {
	assert(Hop_ObjIsNode(o) != 0 && o.nRefs == 0, "unreferenced internal node")
	p.nObjs[o.Type]--
	p.nDeleted++
	Hop_ObjDisconnect(p, o)
	Hop_ManRecycleMemory(p, o)
}
func Hop_ObjDelete_rec(p *Hop_Man_t, o *Hop_Obj_t) {
	if Hop_ObjIsConst1(o) != 0 || Hop_ObjIsPi(o) != 0 {
		return
	}
	a, b := Hop_ObjFanin0(o), Hop_ObjFanin1(o)
	Hop_ObjDelete(p, o)
	for _, f := range []*Hop_Obj_t{a, b} {
		if f != nil && f.Type != AIG_NONE && f.nRefs == 0 {
			Hop_ObjDelete_rec(p, f)
		}
	}
}
func Hop_ObjRepr(o *Hop_Obj_t) *Hop_Obj_t {
	for {
		next, ok := o.pData.(*Hop_Obj_t)
		if !ok || next == nil || next == o {
			return o
		}
		o = next
	}
}
func Hop_ObjCreateChoice(old, next *Hop_Obj_t) {
	a, b := Hop_ObjRepr(old), Hop_ObjRepr(next)
	if a != b {
		b.pData = a
	}
}
