package gia

import "github.com/glycerine/abcgo/misc/tim"

// Gia_ManDupWithBoxes removes unmarked boxes and their unused logic while
// keeping the timing manager and box implementation outputs in sync.
func Gia_ManDupWithBoxes(src *Gia_Man_t, sequential int32) *Gia_Man_t {
	if Gia_ManBoxNum(src) == 0 {
		panic("gia: expected timing boxes")
	}
	timing := src.pManTime.(*tim.Tim_Man_t)
	Gia_ManMarkSeqGiaWithBoxes(src, sequential)
	dst := Gia_ManStart(Gia_ManObjNum(src))
	dst.pName, dst.pSpec = src.pName, src.pSpec
	Gia_ManConst0(src).Value = 0
	for _, obj := range src.pObjs[1:] {
		if Gia_ObjIsTravIdCurrent(src, obj) == 0 {
			continue
		}
		var lit int32
		switch {
		case Gia_ObjIsCi(obj) != 0:
			lit = Gia_ManAppendCi(dst)
		case Gia_ObjIsAnd(obj) != 0:
			lit = Gia_ManAppendAnd(dst, Gia_ObjFanin0Copy(obj), Gia_ObjFanin1Copy(obj))
		case Gia_ObjIsCo(obj) != 0:
			lit = Gia_ManAppendCo(dst, Gia_ObjFanin0Copy(obj))
		default:
			panic("gia: unsupported boxed object")
		}
		obj.Value = uint32(lit)
	}
	if Gia_ManHasDangling(dst) != 0 {
		panic("gia: dangling nodes in boxed duplicate")
	}
	if sequential != 0 {
		dst.vRegClasses = Vec_IntAlloc(int(Gia_ManRegBoxNum(src)))
		if src.vRegInits != nil {
			dst.vRegInits = Vec_IntAlloc(int(Gia_ManRegBoxNum(src)))
		}
		shift := Gia_ManPoNum(src) - Gia_ManRegBoxNum(src)
		for i := int32(0); i < Gia_ManRegBoxNum(src); i++ {
			if Gia_ObjIsTravIdCurrent(src, Gia_ManCo(src, shift+i)) == 0 {
				continue
			}
			Vec_IntPush(dst.vRegClasses, Vec_IntEntry(src.vRegClasses, int(i)))
			if src.vRegInits != nil {
				Vec_IntPush(dst.vRegInits, Vec_IntEntry(src.vRegInits, int(i)))
			}
		}
	} else {
		if src.vRegClasses != nil {
			dst.vRegClasses = Vec_IntDup(src.vRegClasses)
		}
		if src.vRegInits != nil {
			dst.vRegInits = Vec_IntDup(src.vRegInits)
		}
	}
	boxes := Vec_IntAlloc(int(Gia_ManBoxNum(src)))
	ci, co := tim.Tim_ManPiNum(timing), 0
	for i := 0; i < int(Gia_ManBoxNum(src)); i++ {
		ins, outs := tim.Tim_ManBoxInputNum(timing, i), tim.Tim_ManBoxOutputNum(timing, i)
		marked := 0
		for k := 0; k < ins; k++ {
			marked += int(Gia_ObjIsTravIdCurrent(src, Gia_ManCo(src, int32(co+k))))
		}
		for k := 0; k < outs; k++ {
			marked += int(Gia_ObjIsTravIdCurrent(src, Gia_ManCi(src, int32(ci+k))))
		}
		ci, co = ci+outs, co+ins
		if marked != 0 && marked != ins+outs {
			panic("gia: partly marked box")
		}
		if marked != 0 {
			Vec_IntPush(boxes, int32(i))
		}
	}
	co += tim.Tim_ManPoNum(timing)
	if ci != int(Gia_ManCiNum(src)) || co != int(Gia_ManCoNum(src)) {
		panic("gia: inconsistent box interface")
	}
	dst.pManTime = Gia_ManUpdateTimMan2(src, boxes, Gia_ManRegBoxNum(src)-Gia_ManRegBoxNum(dst))
	if src.pAigExtra == nil {
		panic("gia: missing box implementation graph")
	}
	dst.pAigExtra = Gia_ManUpdateExtraAig2(src.pManTime, src.pAigExtra, boxes)
	if Gia_ManCiNum(dst) != int32(tim.Tim_ManPiNum(dst.pManTime.(*tim.Tim_Man_t)))+Gia_ManCoNum(dst.pAigExtra) {
		panic("gia: inconsistent reduced box graph")
	}
	dst.nAnd2Delay = src.nAnd2Delay
	return dst
}
