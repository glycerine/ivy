package gia

// False-path certificates and controlling values are defined on ordinary ANDs.
func giaFalseNeedsLowering(p *Gia_Man_t) bool {
	return p.nXors != 0 || p.nMuxes != 0 || p.nBufs != 0 || p.fGiaSimple != 0
}

// giaFalseLower builds the ordinary representation without cleanup so the
// selected internal root can be mapped to its corresponding node. Output
// polarity stays in the CO literals; path sensitivity is unchanged by it.
func giaFalseLower(p *Gia_Man_t, root int32) (*Gia_Man_t, int32) {
	q := Gia_ManStart(p.nObjs)
	q.CopyIdentityFrom(p)
	Gia_ManHashAlloc(q)
	copies := make([]int32, p.nObjs)
	for i := int32(0); i < Gia_ManCiNum(p); i++ {
		copies[Gia_ManCi(p, i).id] = Gia_ManAppendCi(q)
	}
	read := func(lit int32) int32 { return copies[lit>>1] ^ (lit & 1) }
	for _, obj := range p.pObjs[1:] {
		if Gia_ObjIsAnd(obj) != 0 {
			copies[obj.id] = giaHashMappedGate(q, p, obj, read)
		}
	}
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		obj := Gia_ManCo(p, i)
		copies[obj.id] = Gia_ManAppendCo(q, read(Gia_ObjFaninLit0(obj, obj.id)))
	}
	Gia_ManHashStop(q)
	Gia_ManSetRegNum(q, p.nRegs)
	giaCopyInterfaceMetadata(q, p)
	Gia_ManLevelNum(q)
	return q, copies[root] >> 1
}
func giaFalseDupOrdinary(p *Gia_Man_t) *Gia_Man_t {
	if giaFalseNeedsLowering(p) {
		q, _ := giaFalseLower(p, 0)
		return q
	}
	return Gia_ManDup(p)
}
