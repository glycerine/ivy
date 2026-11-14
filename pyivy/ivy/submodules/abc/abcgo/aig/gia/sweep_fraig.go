package gia

import "github.com/glycerine/abcgo/misc/tim"

func Gia_ManUpdateTimMan(p *Gia_Man_t, present *Vec_Int_t) any {
	timing := p.PManTime.(*tim.Tim_Man_t)
	mask := make([]int, Vec_IntSize(present))
	for i, v := range Vec_IntArray(present) {
		mask[i] = int(v)
	}
	return tim.Tim_ManTrim(timing, mask)
}
func Gia_ManUpdateExtraAig(pTime any, p *Gia_Man_t, present *Vec_Int_t) *Gia_Man_t {
	timing := pTime.(*tim.Tim_Man_t)
	if Vec_IntSize(present) != tim.Tim_ManBoxNum(timing) || Gia_ManCoNum(p) != int32(tim.Tim_ManCiNum(timing)-tim.Tim_ManPiNum(timing)) {
		panic("gia: inconsistent extra graph")
	}
	outputs := Vec_IntAlloc(int(Gia_ManCoNum(p)))
	for i, keep := range Vec_IntArray(present) {
		for k := 0; k < tim.Tim_ManBoxOutputNum(timing, i); k++ {
			Vec_IntPush(outputs, keep)
		}
	}
	return Gia_ManDupOutputVec(p, outputs)
}

func Gia_ManFraigCheckCis(p *Gia_Man_t, obj *Gia_Obj_t) int32 {
	if Gia_ObjIsCi(obj) == 0 {
		panic("gia: expected CI block")
	}
	for i := obj.id; i >= 0 && Gia_ObjIsCi(p.pObjs[i]) != 0; i-- {
		if Gia_ObjIsTravIdCurrent(p, p.pObjs[i]) != 0 {
			return 1
		}
	}
	return 0
}
func Gia_ManFraigMarkCis(p *Gia_Man_t, obj *Gia_Obj_t, mark int32) *Gia_Obj_t {
	if Gia_ObjIsCi(obj) == 0 {
		panic("gia: expected CI block")
	}
	i := obj.id
	for Gia_ObjIsCi(p.pObjs[i]) != 0 {
		if mark != 0 {
			Gia_ObjSetTravIdCurrent(p, p.pObjs[i])
		}
		i--
	}
	return p.pObjs[i]
}
func Gia_ManFraigMarkCos(p *Gia_Man_t, obj *Gia_Obj_t, mark int32) *Gia_Obj_t {
	if Gia_ObjIsCo(obj) == 0 {
		panic("gia: expected CO block")
	}
	i := obj.id
	for Gia_ObjIsCo(p.pObjs[i]) != 0 {
		if mark != 0 {
			Gia_ObjSetTravIdCurrent(p, p.pObjs[i])
			Gia_ObjSetTravIdCurrent(p, Gia_ObjFanin0(p.pObjs[i]))
		}
		i--
	}
	return p.pObjs[i]
}
func Gia_ManFraigMarkAnd(p *Gia_Man_t, obj *Gia_Obj_t) *Gia_Obj_t {
	if Gia_ObjIsAnd(obj) == 0 {
		panic("gia: expected AND block")
	}
	i := obj.id
	for Gia_ObjIsAnd(p.pObjs[i]) != 0 {
		n := p.pObjs[i]
		if Gia_ObjIsTravIdCurrent(p, n) != 0 {
			Gia_ObjSetTravIdCurrent(p, Gia_ObjFanin0(n))
			Gia_ObjSetTravIdCurrent(p, Gia_ObjFanin1(n))
		}
		i--
	}
	return p.pObjs[i]
}

func Gia_ManFraigCreateGia(src *Gia_Man_t) *Gia_Man_t {
	if src.PManTime == nil || src.PAigExtra == nil {
		panic("gia: missing box metadata")
	}
	Gia_ManIncrementTravId(src)
	Gia_ObjSetTravIdCurrent(src, Gia_ManConst0(src))
	present := Vec_IntAlloc(1000)
	mark := int32(1)
	obj := Gia_ManObj(src, Gia_ManObjNum(src)-1)
	if Gia_ObjIsCo(obj) == 0 {
		panic("gia: expected final CO block")
	}
	for Gia_ObjIsCo(obj) != 0 {
		obj = Gia_ManFraigMarkCos(src, obj, mark)
		if Gia_ObjIsAnd(obj) != 0 {
			obj = Gia_ManFraigMarkAnd(src, obj)
		}
		if Gia_ObjIsCi(obj) == 0 {
			panic("gia: expected CI block")
		}
		mark = Gia_ManFraigCheckCis(src, obj)
		obj = Gia_ManFraigMarkCis(src, obj, mark)
		Vec_IntPush(present, mark)
	}
	Vec_IntPop(present)
	Vec_IntReverseOrder(present)
	if Gia_ObjIsConst0(obj) == 0 {
		panic("gia: invalid boxed ordering")
	}
	for _, o := range src.pObjs[1:] {
		if Gia_ObjIsCi(o) == 0 {
			break
		}
		Gia_ObjSetTravIdCurrent(src, o)
	}
	dst := Gia_ManStart(Gia_ManObjNum(src))
	dst.pName, dst.PSpec = src.pName, src.PSpec
	Gia_ManConst0(src).Value = 0
	for _, o := range src.pObjs[1:] {
		if Gia_ObjIsTravIdCurrent(src, o) == 0 {
			continue
		}
		var lit int32
		switch {
		case Gia_ObjIsCi(o) != 0:
			lit = Gia_ManAppendCi(dst)
		case Gia_ObjIsAnd(o) != 0:
			lit = Gia_ManAppendAnd(dst, Gia_ObjFanin0Copy(o), Gia_ObjFanin1Copy(o))
		case Gia_ObjIsCo(o) != 0:
			lit = Gia_ManAppendCo(dst, Gia_ObjFanin0Copy(o))
		default:
			panic("gia: invalid FRAIG object")
		}
		o.Value = uint32(lit)
	}
	dst.PManTime = Gia_ManUpdateTimMan(src, present)
	dst.PAigExtra = Gia_ManUpdateExtraAig(src.PManTime, src.PAigExtra, present)
	return dst
}
