package gia

func Gia_ManAreMostUsedPi_rec(p *Gia_Man_t, obj *Gia_Obj_t) {
	if Gia_ObjIsTravIdCurrent(p, obj) != 0 {
		return
	}
	Gia_ObjSetTravIdCurrent(p, obj)
	if Gia_ObjIsCi(obj) != 0 {
		obj.Value++
		return
	}
	if Gia_ObjIsAnd(obj) == 0 {
		panic("gia: expected AND during support traversal")
	}
	Gia_ManAreMostUsedPi_rec(p, Gia_ObjFanin0(obj))
	Gia_ManAreMostUsedPi_rec(p, Gia_ObjFanin1(obj))
}
func Gia_ManAreMostUsedPi(p *Gia_ManAre_t) *Gia_Obj_t {
	for i := int32(0); i < Gia_ManCiNum(p.pNew); i++ {
		Gia_ManCi(p.pNew, i).Value = 0
	}
	for i := int32(0); i < p.pAig.nRegs; i++ {
		lit := Gia_ObjValue(Gia_ManRi(p.pAig, i))
		if lit <= 1 {
			continue
		}
		Gia_ManIncrementTravId(p.pNew)
		Gia_ManAreMostUsedPi_rec(p.pNew, Gia_ManObj(p.pNew, lit>>1))
	}
	var best *Gia_Obj_t
	for i := int32(0); i < Gia_ManCiNum(p.pNew); i++ {
		obj := Gia_ManCi(p.pNew, i)
		if best == nil || best.Value < obj.Value {
			best = obj
		}
	}
	if best != nil && best.Value > 1 {
		return best
	}
	return nil
}
func Gia_ManCheckPOs_rec(p *Gia_Man_t, obj *Gia_Obj_t) int32 {
	if Gia_ObjIsTravIdCurrent(p, obj) != 0 {
		return 0
	}
	Gia_ObjSetTravIdCurrent(p, obj)
	if Gia_ObjIsCi(obj) != 0 {
		return 1
	}
	if Gia_ObjIsAnd(obj) == 0 {
		panic("gia: expected AND in output support")
	}
	return Gia_ManCheckPOs_rec(p, Gia_ObjFanin0(obj)) + Gia_ManCheckPOs_rec(p, Gia_ObjFanin1(obj))
}
func Gia_ManCheckPOs(p *Gia_ManAre_t) int32 {
	var result int32
	for i := int32(0); i < Gia_ManPoNum(p.pAig); i++ {
		obj := Gia_ManObj(p.pNew, Gia_ObjValue(Gia_ManPo(p.pAig, i))>>1)
		var count int32
		if Gia_ObjIsConst0(obj) == 0 {
			Gia_ManIncrementTravId(p.pNew)
			count = Gia_ManCheckPOs_rec(p.pNew, obj)
		}
		result = max(result, count)
	}
	return result
}
func Gia_ManCheckPOstatus(p *Gia_ManAre_t) int32 {
	for i := int32(0); i < Gia_ManPoNum(p.pAig); i++ {
		lit := Gia_ObjValue(Gia_ManPo(p.pAig, i))
		obj := Gia_ManObj(p.pNew, lit>>1)
		if Gia_ObjIsConst0(obj) == 0 || lit&1 != 0 {
			p.iOutFail = i
			return 1
		}
	}
	return 0
}
