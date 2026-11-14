package gia

import "github.com/glycerine/abcgo/misc/tim"

// Gia_ManMarkSeqGiaWithBoxes_rec marks a cone, traversing all inputs of any
// timing box reached through one of its outputs.
func Gia_ManMarkSeqGiaWithBoxes_rec(p *Gia_Man_t, obj *Gia_Obj_t, roots *Vec_Int_t) {
	if Gia_ObjIsTravIdCurrent(p, obj) != 0 {
		return
	}
	Gia_ObjSetTravIdCurrent(p, obj)
	if Gia_ObjIsAnd(obj) != 0 {
		Gia_ManMarkSeqGiaWithBoxes_rec(p, Gia_ObjFanin0(obj), roots)
		Gia_ManMarkSeqGiaWithBoxes_rec(p, Gia_ObjFanin1(obj), roots)
		return
	}
	if Gia_ObjIsCi(obj) == 0 {
		panic("gia: invalid boxed cone")
	}
	timing := p.PManTime.(*tim.Tim_Man_t)
	realCis := int32(tim.Tim_ManPiNum(timing))
	if Gia_ObjCioId(obj) < realCis {
		regs := Gia_ManRegBoxNum(p)
		flop := Gia_ObjCioId(obj) - (realCis - regs)
		if flop < 0 || flop >= regs {
			panic("gia: unexpected unmarked primary input")
		}
		co := Gia_ManCo(p, Gia_ManPoNum(p)-regs+flop)
		Vec_IntPush(roots, Gia_ObjId(p, co))
		return
	}
	box := tim.Tim_ManBoxForCi(timing, int(Gia_ObjCioId(obj)))
	firstOut, outputs := tim.Tim_ManBoxOutputFirst(timing, box), tim.Tim_ManBoxOutputNum(timing, box)
	for i := 0; i < outputs; i++ {
		Gia_ObjSetTravIdCurrent(p, Gia_ManCi(p, int32(firstOut+i)))
	}
	firstIn, inputs := tim.Tim_ManBoxInputFirst(timing, box), tim.Tim_ManBoxInputNum(timing, box)
	for i := 0; i < inputs; i++ {
		Gia_ObjSetTravIdCurrent(p, Gia_ManCo(p, int32(firstIn+i)))
	}
	for i := 0; i < inputs; i++ {
		Gia_ManMarkSeqGiaWithBoxes_rec(p, Gia_ObjFanin0(Gia_ManCo(p, int32(firstIn+i))), roots)
	}
}
func Gia_ManMarkSeqGiaWithBoxes(p *Gia_Man_t, sequential int32) {
	timing := p.PManTime.(*tim.Tim_Man_t)
	realCis, realCos := int32(tim.Tim_ManPiNum(timing)), int32(tim.Tim_ManPoNum(timing))
	regs := int32(0)
	if sequential != 0 {
		regs = Gia_ManRegBoxNum(p)
	}
	if Gia_ManRegNum(p) != 0 {
		panic("gia: invalid boxed sequential graph")
	}
	Gia_ManIncrementTravId(p)
	Gia_ObjSetTravIdCurrent(p, Gia_ManConst0(p))
	for i := int32(0); i < realCis-regs; i++ {
		Gia_ObjSetTravIdCurrent(p, Gia_ManPi(p, i))
	}
	roots := Vec_IntAlloc(int(Gia_ManRegBoxNum(p)))
	for i := Gia_ManPoNum(p) - realCos; i < Gia_ManPoNum(p)-regs; i++ {
		co := Gia_ManPo(p, i)
		Gia_ObjSetTravIdCurrent(p, co)
		Gia_ManMarkSeqGiaWithBoxes_rec(p, Gia_ObjFanin0(co), roots)
	}
	if sequential != 0 {
		for i := 0; i < Vec_IntSize(roots); i++ {
			obj := Gia_ManObj(p, Vec_IntEntry(roots, i))
			if Gia_ObjIsCo(obj) == 0 {
				panic("gia: invalid register root")
			}
			Gia_ObjSetTravIdCurrent(p, obj)
			Gia_ManMarkSeqGiaWithBoxes_rec(p, Gia_ObjFanin0(obj), roots)
		}
	}
}
