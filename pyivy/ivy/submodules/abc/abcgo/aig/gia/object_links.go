package gia

func Gia_ObjPhaseReal(obj *Gia_Obj_t) int32 {
	return Gia_ObjPhase(Gia_Regular(obj)) ^ Gia_IsComplement(obj)
}
func Gia_ObjPhaseDiff(p *Gia_Man_t, i, k int32) int32 {
	return Gia_ObjPhase(Gia_ManObj(p, i)) ^ Gia_ObjPhase(Gia_ManObj(p, k))
}
func Gia_ObjIsMuxId(p *Gia_Man_t, id int32) int32 {
	return giaBool(id >= 0 && int(id) < len(p.PMuxes) && p.PMuxes[id] != 0)
}
func Gia_ObjIsMux(p *Gia_Man_t, obj *Gia_Obj_t) int32 { return Gia_ObjIsMuxId(p, Gia_ObjId(p, obj)) }
func Gia_ObjIsAndReal(p *Gia_Man_t, obj *Gia_Obj_t) int32 {
	return giaBool(Gia_ObjIsAnd(obj) != 0 && obj.iDiff0 > obj.iDiff1 && Gia_ObjIsMux(p, obj) == 0)
}
func Gia_Obj2Lit(p *Gia_Man_t, obj *Gia_Obj_t) int32 {
	return Gia_ObjId(p, Gia_Regular(obj))<<1 | Gia_IsComplement(obj)
}
func Gia_Lit2Obj(p *Gia_Man_t, lit int32) *Gia_Obj_t {
	return Gia_NotCond(Gia_ManObj(p, lit>>1), lit&1)
}
func Gia_ManCiLit(p *Gia_Man_t, id int32) int32     { return Gia_Obj2Lit(p, Gia_ManCi(p, id)) }
func Gia_ManIdToCioId(p *Gia_Man_t, id int32) int32 { return Gia_ObjCioId(Gia_ManObj(p, id)) }
func Gia_ManCiIdToId(p *Gia_Man_t, id int32) int32  { return Gia_ObjId(p, Gia_ManCi(p, id)) }
func Gia_ManCoIdToId(p *Gia_Man_t, id int32) int32  { return Gia_ObjId(p, Gia_ManCo(p, id)) }
func Gia_ObjIsPi(p *Gia_Man_t, obj *Gia_Obj_t) int32 {
	return giaBool(Gia_ObjIsCi(obj) != 0 && Gia_ObjCioId(obj) < Gia_ManPiNum(p))
}
func Gia_ObjIsPo(p *Gia_Man_t, obj *Gia_Obj_t) int32 {
	return giaBool(Gia_ObjIsCo(obj) != 0 && Gia_ObjCioId(obj) < Gia_ManPoNum(p))
}
func Gia_ObjIsRo(p *Gia_Man_t, obj *Gia_Obj_t) int32 {
	return giaBool(Gia_ObjIsCi(obj) != 0 && Gia_ObjCioId(obj) >= Gia_ManPiNum(p))
}
func Gia_ObjIsRi(p *Gia_Man_t, obj *Gia_Obj_t) int32 {
	return giaBool(Gia_ObjIsCo(obj) != 0 && Gia_ObjCioId(obj) >= Gia_ManPoNum(p))
}
func Gia_ObjRoToRi(p *Gia_Man_t, obj *Gia_Obj_t) *Gia_Obj_t {
	if Gia_ObjIsRo(p, obj) == 0 {
		panic("gia: expected register output")
	}
	return Gia_ManCo(p, Gia_ManCoNum(p)-Gia_ManCiNum(p)+Gia_ObjCioId(obj))
}
func Gia_ObjRiToRo(p *Gia_Man_t, obj *Gia_Obj_t) *Gia_Obj_t {
	if Gia_ObjIsRi(p, obj) == 0 {
		panic("gia: expected register input")
	}
	return Gia_ManCi(p, Gia_ManCiNum(p)-Gia_ManCoNum(p)+Gia_ObjCioId(obj))
}
func Gia_ObjRoToRiId(p *Gia_Man_t, id int32) int32 {
	return Gia_ObjId(p, Gia_ObjRoToRi(p, Gia_ManObj(p, id)))
}
func Gia_ObjRiToRoId(p *Gia_Man_t, id int32) int32 {
	return Gia_ObjId(p, Gia_ObjRiToRo(p, Gia_ManObj(p, id)))
}
func Gia_ObjDiff0(obj *Gia_Obj_t) int32 { return int32(obj.iDiff0) }
func Gia_ObjDiff1(obj *Gia_Obj_t) int32 { return int32(obj.iDiff1) }
func Gia_ObjFaninC2(p *Gia_Man_t, obj *Gia_Obj_t) int32 {
	if len(p.PMuxes) == 0 {
		return 0
	}
	return int32(p.PMuxes[Gia_ObjId(p, obj)] & 1)
}
func Gia_ObjFanin2(p *Gia_Man_t, obj *Gia_Obj_t) *Gia_Obj_t {
	if len(p.PMuxes) == 0 {
		return nil
	}
	return Gia_ManObj(p, int32(p.PMuxes[Gia_ObjId(p, obj)]>>1))
}
func Gia_ObjChild0(obj *Gia_Obj_t) *Gia_Obj_t {
	return Gia_NotCond(Gia_ObjFanin0(obj), Gia_ObjFaninC0(obj))
}
func Gia_ObjChild1(obj *Gia_Obj_t) *Gia_Obj_t {
	return Gia_NotCond(Gia_ObjFanin1(obj), Gia_ObjFaninC1(obj))
}
func Gia_ObjChild2(p *Gia_Man_t, obj *Gia_Obj_t) *Gia_Obj_t {
	return Gia_NotCond(Gia_ObjFanin2(p, obj), Gia_ObjFaninC2(p, obj))
}
func Gia_ObjFaninId2(p *Gia_Man_t, id int32) int32 {
	if Gia_ObjIsMuxId(p, id) == 0 {
		return -1
	}
	return int32(p.PMuxes[id] >> 1)
}
func Gia_ObjFaninId0p(p *Gia_Man_t, obj *Gia_Obj_t) int32 {
	return Gia_ObjFaninId0(obj, Gia_ObjId(p, obj))
}
func Gia_ObjFaninId1p(p *Gia_Man_t, obj *Gia_Obj_t) int32 {
	return Gia_ObjFaninId1(obj, Gia_ObjId(p, obj))
}
func Gia_ObjFaninId2p(p *Gia_Man_t, obj *Gia_Obj_t) int32 {
	return Gia_ObjFaninId2(p, Gia_ObjId(p, obj))
}
func Gia_ObjFaninLit0p(p *Gia_Man_t, obj *Gia_Obj_t) int32 {
	return Gia_ObjFaninLit0(obj, Gia_ObjId(p, obj))
}
func Gia_ObjFaninLit1p(p *Gia_Man_t, obj *Gia_Obj_t) int32 {
	return Gia_ObjFaninLit1(obj, Gia_ObjId(p, obj))
}
func Gia_ObjFaninLit2(p *Gia_Man_t, id int32) int32 {
	if Gia_ObjIsMuxId(p, id) == 0 {
		return -1
	}
	return int32(p.PMuxes[id])
}
func Gia_ObjFaninLit2p(p *Gia_Man_t, obj *Gia_Obj_t) int32 {
	return Gia_ObjFaninLit2(p, Gia_ObjId(p, obj))
}
func Gia_ObjFaninNum(p *Gia_Man_t, obj *Gia_Obj_t) int32 {
	if Gia_ObjIsMux(p, obj) != 0 {
		return 3
	}
	if Gia_ObjIsAnd(obj) != 0 {
		return 2
	}
	if Gia_ObjIsCo(obj) != 0 {
		return 1
	}
	return 0
}
func Gia_ObjWhatFanin(p *Gia_Man_t, obj, fanin *Gia_Obj_t) int32 {
	if Gia_ObjFanin0(obj) == fanin {
		return 0
	}
	if Gia_ObjFanin1(obj) == fanin {
		return 1
	}
	if Gia_ObjFanin2(p, obj) == fanin {
		return 2
	}
	panic("gia: object is not a fanin")
}
func Gia_ObjCopy(p *Gia_Man_t, obj *Gia_Obj_t) *Gia_Obj_t { return Gia_ManObj(p, int32(obj.Value>>1)) }
func Gia_ObjLitCopy(p *Gia_Man_t, lit int32) int32 {
	return int32(Gia_ManObj(p, lit>>1).Value) ^ (lit & 1)
}
func Gia_ObjFanin2Copy(p *Gia_Man_t, obj *Gia_Obj_t) int32 {
	return int32(Gia_ObjFanin2(p, obj).Value) ^ Gia_ObjFaninC2(p, obj)
}
func Gia_ManPoIsConst(p *Gia_Man_t, index int32) int32 {
	return giaBool(Gia_ObjFaninId0p(p, Gia_ManPo(p, index)) == 0)
}
func Gia_ManPoIsConst0(p *Gia_Man_t, index int32) int32 {
	return giaBool(Gia_ObjFaninLit0p(p, Gia_ManPo(p, index)) == 0)
}
func Gia_ManPoIsConst1(p *Gia_Man_t, index int32) int32 {
	return giaBool(Gia_ObjFaninLit0p(p, Gia_ManPo(p, index)) == 1)
}
