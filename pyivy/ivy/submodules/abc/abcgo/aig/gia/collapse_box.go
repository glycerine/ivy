package gia

import "github.com/glycerine/abcgo/misc/tim"

func Gia_ManDupCollapse_rec(p *Gia_Man_t, obj *Gia_Obj_t, dst *Gia_Man_t) {
	if Gia_ObjIsTravIdCurrent(p, obj) != 0 {
		return
	}
	Gia_ObjSetTravIdCurrent(p, obj)
	if Gia_ObjIsAnd(obj) == 0 {
		panic("gia: uninitialized collapse input")
	}
	sibling := Gia_ObjSibl(p, obj.id)
	if sibling != 0 {
		Gia_ManDupCollapse_rec(p, Gia_ManObj(p, sibling), dst)
	}
	Gia_ManDupCollapse_rec(p, Gia_ObjFanin0(obj), dst)
	Gia_ManDupCollapse_rec(p, Gia_ObjFanin1(obj), dst)
	obj.Value = uint32(Gia_ManHashAnd(dst, Gia_ObjFanin0Copy(obj), Gia_ObjFanin1Copy(obj)))
	if sibling != 0 {
		id := int(obj.Value >> 1)
		if id >= len(dst.pSibls) {
			dst.pSibls = append(dst.pSibls, make([]int32, id+1-len(dst.pSibls))...)
		}
		dst.pSibls[id] = int32(Gia_ManObj(p, sibling).Value >> 1)
	}
}

// Gia_ManDupCollapse substitutes white-box implementations and exposes black
// box ports as inputs and outputs, preserving copy mappings through cleanup.
func Gia_ManDupCollapse(p, boxes *Gia_Man_t, present *Vec_Int_t, sequential int32) *Gia_Man_t {
	timing := p.pManTime.(*tim.Tim_Man_t)
	if sequential != 0 && p.vRegClasses == nil {
		panic("gia: missing register classes")
	}
	pi, po := tim.Tim_ManPiNum(timing), tim.Tim_ManPoNum(timing)
	if Gia_ManCiNum(p) != int32(pi)+Gia_ManCoNum(boxes) {
		panic("gia: invalid box implementation interface")
	}
	if present != nil && Vec_IntSize(present) != tim.Tim_ManBoxNum(timing) {
		panic("gia: invalid box presence mask")
	}
	dst := Gia_ManStart(Gia_ManObjNum(p))
	dst.pName, dst.pSpec = p.pName, p.pSpec
	if Gia_ManHasChoices(p) != 0 {
		dst.pSibls = make([]int32, Gia_ManObjNum(p))
	}
	Gia_ManHashAlloc(dst)
	Gia_ManFillValue(p)
	Gia_ManConst0(p).Value = 0
	Gia_ManIncrementTravId(p)
	Gia_ObjSetTravIdCurrent(p, Gia_ManConst0(p))
	for i := 0; i < pi; i++ {
		obj := Gia_ManCi(p, int32(i))
		obj.Value = uint32(Gia_ManAppendCi(dst))
		Gia_ObjSetTravIdCurrent(p, obj)
	}
	ci, co, newPIs, blackIns, blackOuts := pi, 0, 0, 0, 0
	for i := 0; i < tim.Tim_ManBoxNum(timing); i++ {
		Gia_ManIncrementTravId(boxes)
		Gia_ObjSetTravIdCurrent(boxes, Gia_ManConst0(boxes))
		Gia_ManConst0(boxes).Value = 0
		ins, outs := tim.Tim_ManBoxInputNum(timing, i), tim.Tim_ManBoxOutputNum(timing, i)
		if tim.Tim_ManBoxIsBlack(timing, i) {
			skip := present != nil && Vec_IntEntry(present, i) == 0
			for k := 0; k < ins; k++ {
				obj := Gia_ManCo(p, int32(co+k))
				Gia_ManDupCollapse_rec(p, Gia_ObjFanin0(obj), dst)
				if skip {
					obj.Value = ^uint32(0)
				} else {
					obj.Value = uint32(Gia_ManAppendCo(dst, Gia_ObjFanin0Copy(obj)))
				}
				blackOuts++
			}
			for k := 0; k < outs; k++ {
				obj := Gia_ManCi(p, int32(ci+k))
				if skip {
					obj.Value = 0
				} else {
					obj.Value = uint32(Gia_ManAppendCi(dst))
					newPIs++
				}
				Gia_ObjSetTravIdCurrent(p, obj)
				blackIns++
			}
		} else {
			for k := 0; k < ins; k++ {
				obj := Gia_ManCo(p, int32(co+k))
				Gia_ManDupCollapse_rec(p, Gia_ObjFanin0(obj), dst)
				input := Gia_ManCi(boxes, int32(k))
				input.Value = uint32(Gia_ObjFanin0Copy(obj))
				Gia_ObjSetTravIdCurrent(boxes, input)
			}
			for k := 0; k < outs; k++ {
				output := Gia_ManCo(boxes, int32(ci-pi+k))
				Gia_ManDupCollapse_rec(boxes, Gia_ObjFanin0(output), dst)
				obj := Gia_ManCi(p, int32(ci+k))
				obj.Value = uint32(Gia_ObjFanin0Copy(output))
				Gia_ObjSetTravIdCurrent(p, obj)
			}
		}
		co += ins
		ci += outs
	}
	for i := tim.Tim_ManCoNum(timing) - po; i < tim.Tim_ManCoNum(timing); i++ {
		obj := Gia_ManCo(p, int32(i))
		Gia_ManDupCollapse_rec(p, Gia_ObjFanin0(obj), dst)
		obj.Value = uint32(Gia_ManAppendCo(dst, Gia_ObjFanin0Copy(obj)))
	}
	co += po
	if ci != int(Gia_ManCiNum(p)) || co != int(Gia_ManCoNum(p)) {
		panic("gia: inconsistent box terminal count")
	}
	regs := Gia_ManRegNum(p)
	if sequential != 0 {
		regs = int32(Vec_IntSize(p.vRegClasses))
	}
	Gia_ManSetRegNum(dst, regs)
	Gia_ManHashStop(dst)
	previous := dst
	dst = Gia_ManCleanup(previous)
	Gia_ManCleanupRemap(p, previous)
	Gia_ManStop(previous)
	if newPIs != 0 {
		previous = dst
		dst = Gia_ManDupMoveLast(previous, int32(pi)-Gia_ManRegNum(previous), int32(newPIs))
		Gia_ManCleanupRemap(p, previous)
		Gia_ManStop(previous)
	}
	if present == nil && (po != int(Gia_ManCoNum(dst))-blackOuts || pi != int(Gia_ManCiNum(dst))-blackIns) {
		panic("gia: invalid collapsed terminal count")
	}
	if sequential != 0 && p.vRegInits != nil && Vec_IntSum(p.vRegInits) != 0 {
		if Vec_IntSize(p.vRegInits) != int(Gia_ManRegNum(dst)) {
			panic("gia: invalid register initialization")
		}
		initial := make([]byte, Vec_IntSize(p.vRegInits))
		for i, v := range Vec_IntArray(p.vRegInits) {
			switch v {
			case 0:
				initial[i] = '0'
			case 1:
				initial[i] = '1'
			default:
				initial[i] = 'X'
			}
		}
		previous = dst
		dst = Gia_ManDupZeroUndc(previous, string(initial), 0, 0, 1)
		dst.nConstrs = previous.nConstrs
		previous.nConstrs = 0
		Gia_ManStop(previous)
	}
	return dst
}
