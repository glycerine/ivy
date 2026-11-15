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
		a, b, c, kind := p.Gate(obj.id)
		mark := p.pObjs[a>>1].fMark0
		if kind == 2 || kind == 3 || kind == 4 {
			mark |= p.pObjs[b>>1].fMark0
		}
		if kind == 4 {
			mark |= p.pObjs[c>>1].fMark0
		}
		obj.fMark0 = mark
	}
}
func Gia_ManDupReplaceCut(src *Gia_Man_t) *Gia_Man_t {
	Gia_ManHighLightFlopLogic(src)
	defer Gia_ManCleanMark0(src)
	out := giaDupMappedStart(src)
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
	// Copy uncut PI-only output cones before ROs, keeping registers last.
	for i := int32(0); i < Gia_ManCoNum(src); i++ {
		fanin := Gia_ObjFanin0(Gia_ManCo(src, i))
		if fanin.fMark0 == 0 && fanin.Value == ^uint32(0) {
			Gia_ManDupOrderDfs_rec(out, src, fanin)
		}
	}
	for i := int32(0); i < Gia_ManRegNum(src); i++ {
		Gia_ManRo(src, i).Value = uint32(Gia_ManAppendCi(out))
	}
	for i := int32(0); i < Gia_ManCoNum(src); i++ {
		Gia_ManDupOrderDfs_rec(out, src, Gia_ManCo(src, i))
	}
	Gia_ManSetRegNum(out, Gia_ManRegNum(src))
	giaCopyOutputRoles(out, src)
	return out
}
