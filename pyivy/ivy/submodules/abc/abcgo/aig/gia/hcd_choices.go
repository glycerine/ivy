package gia

// Hcd_ManEquivToChoices_rec copies a cone and installs acyclic choice links.
func Hcd_ManEquivToChoices_rec(dst, src *Gia_Man_t, obj *Gia_Obj_t) {
	if obj.Value != ^uint32(0) {
		return
	}
	repr := Gia_ObjReprObj(src, obj.id)
	if repr != nil {
		phase := uint32(Gia_ObjPhaseReal(repr) ^ Gia_ObjPhaseReal(obj))
		if Gia_ObjIsConst0(repr) != 0 {
			obj.Value = repr.Value ^ phase
			return
		}
		Hcd_ManEquivToChoices_rec(dst, src, repr)
		if Gia_ObjIsAnd(obj) == 0 {
			panic("gia: expected equivalent AND")
		}
		Hcd_ManEquivToChoices_rec(dst, src, Gia_ObjFanin0(obj))
		Hcd_ManEquivToChoices_rec(dst, src, Gia_ObjFanin1(obj))
		obj.Value = uint32(Gia_ManHashAnd(dst, Gia_ObjFanin0Copy(obj), Gia_ObjFanin1Copy(obj)))
		if obj.Value>>1 == repr.Value>>1 {
			if obj.Value != repr.Value^phase {
				panic("gia: inconsistent choice phase")
			}
			return
		}
		if repr.Value > obj.Value {
			return
		}
		r, o := Gia_ManObj(dst, int32(repr.Value>>1)), Gia_ManObj(dst, int32(obj.Value>>1))
		if existing := Gia_ObjReprObj(dst, o.id); existing != nil {
			if existing != r {
				panic("gia: conflicting choice representative")
			}
			obj.Value = repr.Value ^ phase
			return
		}
		if Hcd_ObjCheckTfi(dst, r, o) == 0 {
			if Gia_ObjNext(dst, o.id) != 0 {
				panic("gia: choice already linked")
			}
			Gia_ObjSetRepr(dst, o.id, r.id)
			Hcd_ManAddNextEntry_rec(dst, r, o)
		}
		obj.Value = repr.Value ^ phase
		return
	}
	if Gia_ObjIsAnd(obj) == 0 {
		panic("gia: expected AND in choice cone")
	}
	Hcd_ManEquivToChoices_rec(dst, src, Gia_ObjFanin0(obj))
	Hcd_ManEquivToChoices_rec(dst, src, Gia_ObjFanin1(obj))
	obj.Value = uint32(Gia_ManHashAnd(dst, Gia_ObjFanin0Copy(obj), Gia_ObjFanin1Copy(obj)))
}
func Hcd_ManRemoveBadChoices(p *Gia_Man_t) {
	Gia_ManRemoveBadChoices(p)
}
func Hcd_ManEquivToChoices(p *Gia_Man_t, snapshots int32) *Gia_Man_t {
	if snapshots <= 0 || Gia_ManCoNum(p)%snapshots != 0 || p.PReprs == nil {
		panic("gia: invalid choice snapshots or classes")
	}
	Gia_ManSetPhase(p)
	dst := Gia_ManStart(p.nObjs)
	dst.pName, dst.PSpec = p.pName, p.PSpec
	dst.PReprs = make([]Gia_Rpr_t, p.nObjs)
	dst.PNexts = make([]int32, p.nObjs)
	for i := range dst.PReprs {
		dst.PReprs[i].iRepr = GIA_VOID
	}
	Gia_ManFillValue(p)
	Gia_ManConst0(p).Value = 0
	for i := int32(0); i < Gia_ManCiNum(p); i++ {
		Gia_ManCi(p, i).Value = uint32(Gia_ManAppendCi(dst))
	}
	for i := int32(0); i < Gia_ManRegNum(p); i++ {
		obj := Gia_ManRo(p, i)
		if repr := Gia_ObjReprObj(p, obj.id); repr != nil {
			obj.Value = repr.Value
		}
	}
	Gia_ManHashAlloc(dst)
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		Hcd_ManEquivToChoices_rec(dst, p, Gia_ObjFanin0(Gia_ManCo(p, i)))
	}
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		if i%snapshots == 0 {
			Gia_ManAppendCo(dst, Gia_ObjFanin0Copy(Gia_ManCo(p, i)))
		}
	}
	Gia_ManHashStop(dst)
	Gia_ManSetRegNum(dst, Gia_ManRegNum(p))
	dst.PReprs = dst.PReprs[:dst.nObjs]
	dst.PNexts = dst.PNexts[:dst.nObjs]
	Hcd_ManRemoveBadChoices(dst)
	return Gia_ManCleanup(dst)
}
