package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

// Rnm_ManJustifyPropFanout_rec propagates justified values towards outputs.
func Rnm_ManJustifyPropFanout_rec(p *Rnm_Man_t, obj *gia.Gia_Obj_t, frame int32, selected *vec.Vec_Int_t) {
	r := Rnm_ManObj(p, obj, frame)
	if r.fVisit != 0 { panic("refinement object visited twice") }
	r.fVisit = 1
	if Rnm_ManObj(p, obj, 0).fVisitJ == 0 { Rnm_ManObj(p, obj, 0).fVisitJ = 1; p.nVisited++ }
	if r.fPPi != 0 {
		if r.Prio == 0 { panic("pseudo-input has no priority") }
		for f := p.pCex.IFrame; f >= 0; f-- {
			if Rnm_ManObj(p, obj, f).fVisit == 0 { Rnm_ManJustifyPropFanout_rec(p, obj, f, selected) }
		}
		vec.Vec_IntPush(selected, gia.Gia_ObjId(p.pGia, obj))
		return
	}
	if (gia.Gia_ObjIsCo(obj) != 0 && frame == p.pCex.IFrame) || gia.Gia_ObjIsPo(p.pGia, obj) != 0 { return }
	if gia.Gia_ObjIsRi(p.pGia, obj) != 0 {
		fanout := gia.Gia_ObjRiToRo(p.pGia, obj)
		if Rnm_ManObj(p, fanout, frame+1).fVisit == 0 { Rnm_ManJustifyPropFanout_rec(p, fanout, frame+1, selected) }
		return
	}
	if gia.Gia_ObjIsRo(p.pGia, obj) == 0 && gia.Gia_ObjIsAnd(obj) == 0 { panic("invalid refinement fanout object") }
	for k := int32(0); k < gia.Gia_ObjFanoutNum(p.pGia, obj); k++ {
		fanout := gia.Gia_ObjFanout(p.pGia, obj, k)
		if fanout.Value == 0 { continue }
		out := Rnm_ManObj(p, fanout, frame)
		if out.fPPi != 0 || out.fVisit != 0 { continue }
		if gia.Gia_ObjIsCo(fanout) != 0 { Rnm_ManJustifyPropFanout_rec(p, fanout, frame, selected); continue }
		if gia.Gia_ObjIsAnd(fanout) == 0 { panic("invalid refinement fanout gate") }
		a := Rnm_ManObj(p, gia.Gia_ObjFanin0(fanout), frame)
		b := Rnm_ManObj(p, gia.Gia_ObjFanin1(fanout), frame)
		av, bv := a.Value^uint32(gia.Gia_ObjFaninC0(fanout)), b.Value^uint32(gia.Gia_ObjFaninC1(fanout))
		if av == 0 && a.fVisit != 0 || bv == 0 && b.fVisit != 0 || av == 1 && a.fVisit != 0 && bv == 1 && b.fVisit != 0 {
			Rnm_ManJustifyPropFanout_rec(p, fanout, frame, selected)
		}
	}
}

func Rnm_ManJustify_rec(p *Rnm_Man_t, obj *gia.Gia_Obj_t, frame int32, selected *vec.Vec_Int_t) {
	r := Rnm_ManObj(p, obj, frame)
	if r.fVisit != 0 { return }
	if p.fPropFanout != 0 {
		Rnm_ManJustifyPropFanout_rec(p, obj, frame, selected)
	} else {
		r.fVisit = 1
		if Rnm_ManObj(p, obj, 0).fVisitJ == 0 { Rnm_ManObj(p, obj, 0).fVisitJ = 1; p.nVisited++ }
	}
	if r.fPPi != 0 {
		if r.Prio == 0 { panic("pseudo-input has no priority") }
		if p.fPropFanout != 0 {
			for f := p.pCex.IFrame; f >= 0; f-- {
				if Rnm_ManObj(p, obj, f).fVisit == 0 { Rnm_ManJustifyPropFanout_rec(p, obj, f, selected) }
			}
		} else { vec.Vec_IntPush(selected, gia.Gia_ObjId(p.pGia, obj)) }
		return
	}
	if gia.Gia_ObjIsPi(p.pGia, obj) != 0 || gia.Gia_ObjIsConst0(obj) != 0 { return }
	if gia.Gia_ObjIsRo(p.pGia, obj) != 0 {
		if frame > 0 { Rnm_ManJustify_rec(p, gia.Gia_ObjFanin0(gia.Gia_ObjRoToRi(p.pGia, obj)), frame-1, selected) }
		return
	}
	if gia.Gia_ObjIsAnd(obj) == 0 { panic("invalid refinement justification object") }
	left, right := gia.Gia_ObjFanin0(obj), gia.Gia_ObjFanin1(obj)
	a, b := Rnm_ManObj(p, left, frame), Rnm_ManObj(p, right, frame)
	visit := func(child *gia.Gia_Obj_t, value *Rnm_Obj_t) { if value.Prio > 0 { Rnm_ManJustify_rec(p, child, frame, selected) } }
	if r.Value == 1 { visit(left, a); visit(right, b); return }
	av, bv := a.Value^uint32(gia.Gia_ObjFaninC0(obj)), b.Value^uint32(gia.Gia_ObjFaninC1(obj))
	switch {
	case av == 0 && bv == 0:
		if a.Prio <= b.Prio { visit(left, a) } else { visit(right, b) }
	case av == 0: visit(left, a)
	case bv == 0: visit(right, b)
	default: panic("false AND has no controlling input")
	}
}
