package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

// Gia_ManRefSetAndPropFanout_rec propagates a marked value through its
// controlling fanouts, including register transitions and shared PPIs.
func Gia_ManRefSetAndPropFanout_rec(p *Gla_Man_t, object *gia.Gia_Obj_t, frame int32, selected *vec.Vec_Int_t, sign int32) {
	ref := Gla_ObjRef(p, object, frame)
	if int32(ref.Sign) != sign {
		return
	}
	if ref.fVisit != 0 {
		panic("GLA implication already visited")
	}
	ref.fVisit = 1
	if ref.fPPi != 0 {
		if ref.Prio == 0 {
			panic("GLA pseudo-input without priority")
		}
		for f := p.pPars.iFrame; f >= 0; f-- {
			if Gla_ObjRef(p, object, f).fVisit == 0 {
				Gia_ManRefSetAndPropFanout_rec(p, object, f, selected, sign)
			}
		}
		vec.Vec_IntPush(selected, gia.Gia_ObjId(p.pGia, object))
		return
	}
	if (gia.Gia_ObjIsCo(object) != 0 && frame == p.pPars.iFrame) || gia.Gia_ObjIsPo(p.pGia, object) != 0 {
		return
	}
	if gia.Gia_ObjIsRi(p.pGia, object) != 0 {
		ro := gia.Gia_ObjRiToRo(p.pGia, object)
		if Gla_ObjRef(p, ro, frame+1).fVisit == 0 {
			Gia_ManRefSetAndPropFanout_rec(p, ro, frame+1, selected, sign)
		}
		return
	}
	if gia.Gia_ObjIsRo(p.pGia, object) == 0 && gia.Gia_ObjIsAnd(object) == 0 {
		panic("invalid GLA implication node")
	}
	for k := int32(0); k < gia.Gia_ObjFanoutNum(p.pGia, object); k++ {
		fan := gia.Gia_ObjFanout(p.pGia, object, k)
		if Gla_ObjRef(p, fan, frame).fVisit != 0 {
			continue
		}
		if gia.Gia_ObjIsCo(fan) != 0 {
			Gia_ManRefSetAndPropFanout_rec(p, fan, frame, selected, sign)
			continue
		}
		if gia.Gia_ObjIsAnd(fan) == 0 {
			panic("invalid GLA fanout")
		}
		a, b := Gla_ObjRef(p, gia.Gia_ObjFanin0(fan), frame), Gla_ObjRef(p, gia.Gia_ObjFanin1(fan), frame)
		av, bv := a.Value^uint32(gia.Gia_ObjFaninC0(fan)), b.Value^uint32(gia.Gia_ObjFaninC1(fan))
		if (av == 0 && a.fVisit != 0) || (bv == 0 && b.fVisit != 0) || (av == 1 && a.fVisit != 0 && bv == 1 && b.fVisit != 0) {
			Gia_ManRefSetAndPropFanout_rec(p, fan, frame, selected, sign)
		}
	}
}

func Gla_ManRefSelect_rec(p *Gla_Man_t, object *gia.Gia_Obj_t, frame int32, selected *vec.Vec_Int_t, sign int32) {
	ref := Gla_ObjRef(p, object, frame)
	if ref.fVisit != 0 {
		return
	}
	if p.pPars.fPropFanout != 0 {
		Gia_ManRefSetAndPropFanout_rec(p, object, frame, selected, sign)
	} else {
		ref.fVisit = 1
	}
	if ref.fPPi != 0 {
		if ref.Prio == 0 {
			panic("GLA pseudo-input without priority")
		}
		if p.pPars.fPropFanout != 0 {
			for f := p.pPars.iFrame; f >= 0; f-- {
				if Gla_ObjRef(p, object, f).fVisit == 0 {
					Gia_ManRefSetAndPropFanout_rec(p, object, f, selected, sign)
				}
			}
		} else {
			vec.Vec_IntPush(selected, gia.Gia_ObjId(p.pGia, object))
			vec.Vec_IntAddToEntry(p.vObjCounts, int(frame), 1)
		}
		return
	}
	if gia.Gia_ObjIsPi(p.pGia, object) != 0 || gia.Gia_ObjIsConst0(object) != 0 {
		return
	}
	if gia.Gia_ObjIsRo(p.pGia, object) != 0 {
		if frame > 0 {
			Gla_ManRefSelect_rec(p, gia.Gia_ObjFanin0(gia.Gia_ObjRoToRi(p.pGia, object)), frame-1, selected, sign)
		}
		return
	}
	if gia.Gia_ObjIsAnd(object) == 0 {
		panic("invalid GLA justification node")
	}
	a, b := gia.Gia_ObjFanin0(object), gia.Gia_ObjFanin1(object)
	ar, br := Gla_ObjRef(p, a, frame), Gla_ObjRef(p, b, frame)
	visit := func(o *gia.Gia_Obj_t, r *Rfn_Obj_t) {
		if r.Prio > 0 {
			Gla_ManRefSelect_rec(p, o, frame, selected, sign)
		}
	}
	if ref.Value == 1 {
		visit(a, ar)
		visit(b, br)
		return
	}
	av, bv := ar.Value^uint32(gia.Gia_ObjFaninC0(object)), br.Value^uint32(gia.Gia_ObjFaninC1(object))
	if av == 0 && bv == 0 {
		if ar.Prio <= br.Prio {
			visit(a, ar)
		} else {
			visit(b, br)
		}
	} else if av == 0 {
		visit(a, ar)
	} else if bv == 0 {
		visit(b, br)
	} else {
		panic("invalid zero-valued GLA AND")
	}
}
