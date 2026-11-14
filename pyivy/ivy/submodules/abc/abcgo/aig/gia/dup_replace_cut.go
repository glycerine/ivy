package gia

// Gia_ManHighLightFlopLogic marks nodes in the transitive fanout of registers.
func Gia_ManHighLightFlopLogic(p *Gia_Man_t) {
	p.pObjs[0].fMark0 = 0
	for i := int32(0); i < Gia_ManPiNum(p); i++ {
		Gia_ManPi(p, i).fMark0 = 0
	}
	for i := int32(0); i < Gia_ManRegNum(p); i++ {
		Gia_ManRo(p, i).fMark0 = 1
	}
	for _, obj := range p.pObjs {
		if Gia_ObjIsCi(obj) != 0 || obj.id == 0 {
			continue
		}
		obj.fMark0 = 0
		for _, id := range giaObjectFanins(p, obj) {
			obj.fMark0 |= p.pObjs[id].fMark0
		}
	}
}
func Gia_ManDupReplaceCut(src *Gia_Man_t) *Gia_Man_t {
	Gia_ManHighLightFlopLogic(src)
	defer Gia_ManCleanMark0(src)
	out := Gia_ManStart(src.nObjs)
	out.CopyIdentityFrom(src)
	if src.PMuxes != nil {
		out.EnableMuxes()
	}
	Gia_ManFillValue(src)
	src.pObjs[0].Value = 0
	for _, obj := range src.pObjs {
		if Gia_ObjIsAnd(obj) == 0 || obj.fMark0 == 0 {
			continue
		}
		for _, id := range giaObjectFanins(src, obj) {
			fanin := src.pObjs[id]
			if fanin.fMark0 == 0 && fanin.Value == ^uint32(0) {
				fanin.Value = uint32(Gia_ManAppendCi(out))
			}
		}
	}
	// An output driven directly by PI-only logic is also a boundary signal.
	for i := int32(0); i < Gia_ManCoNum(src); i++ {
		fanin := Gia_ObjFanin0(Gia_ManCo(src, i))
		if fanin.fMark0 == 0 && fanin.Value == ^uint32(0) {
			fanin.Value = uint32(Gia_ManAppendCi(out))
		}
	}
	for i := int32(0); i < Gia_ManRegNum(src); i++ {
		Gia_ManRo(src, i).Value = uint32(Gia_ManAppendCi(out))
	}
	for i := int32(0); i < Gia_ManCoNum(src); i++ {
		Gia_ManDupOrderDfs_rec(out, src, Gia_ManCo(src, i))
	}
	Gia_ManSetRegNum(out, Gia_ManRegNum(src))
	return out
}
