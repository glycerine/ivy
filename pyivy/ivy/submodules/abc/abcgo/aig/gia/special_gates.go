package gia

func Gia_ManAppendXorReal(p *Gia_Man_t, a, b int32) int32 {
	giaCheckFanin(p, a)
	giaCheckFanin(p, b)
	if a>>1 == b>>1 {
		panic("gia: real XOR requires distinct fanin objects")
	}
	if a>>1 < b>>1 {
		a, b = b, a
	}
	obj := Gia_ManAppendObj(p)
	obj.iDiff0 = uint32(obj.id - (a >> 1))
	obj.fCompl0 = uint32(a & 1)
	obj.iDiff1 = uint32(obj.id - (b >> 1))
	obj.fCompl1 = uint32(b & 1)
	p.nXors++
	return obj.id << 1
}
func Gia_ManAppendMuxReal(p *Gia_Man_t, control, then, otherwise int32) int32 {
	giaCheckFanin(p, control)
	giaCheckFanin(p, then)
	giaCheckFanin(p, otherwise)
	if p.PMuxes == nil {
		panic("gia: real MUX storage is not enabled")
	}
	if then>>1 == otherwise>>1 || control>>1 == then>>1 || control>>1 == otherwise>>1 {
		panic("gia: real MUX requires distinct fanin objects")
	}
	if Vec_IntSize(&p.vHTable) != 0 && then&1 != 0 {
		panic("gia: hashed real MUX then-fanin must be regular")
	}
	if otherwise>>1 > then>>1 {
		then, otherwise = otherwise, then
		control ^= 1
	}
	obj := Gia_ManAppendObj(p)
	obj.iDiff0 = uint32(obj.id - (otherwise >> 1))
	obj.fCompl0 = uint32(otherwise & 1)
	obj.iDiff1 = uint32(obj.id - (then >> 1))
	obj.fCompl1 = uint32(then & 1)
	p.PMuxes[obj.id] = uint32(control)
	p.nMuxes++
	return obj.id << 1
}
func Gia_ManAppendBuf(p *Gia_Man_t, lit int32) int32 {
	giaCheckFanin(p, lit)
	obj := Gia_ManAppendObj(p)
	obj.iDiff0 = uint32(obj.id - (lit >> 1))
	obj.iDiff1 = obj.iDiff0
	obj.fCompl0 = uint32(lit & 1)
	obj.fCompl1 = obj.fCompl0
	p.nBufs++
	return obj.id << 1
}

// EnableMuxes enables storage for real MUX control literals. Existing metadata
// remains intact; subsequent object appends grow the storage with the manager.
func (p *Gia_Man_t) EnableMuxes() {
	if p.PMuxes == nil {
		p.PMuxes = make([]uint32, len(p.pObjs))
	}
}
