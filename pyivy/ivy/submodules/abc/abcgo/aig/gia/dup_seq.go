package gia

// Gia_ManDupFlip complements each selected RO and its RI, preserving the
// transition relation under the corresponding state-coordinate change. Packed
// bits use the source Vec_Bit int32-word layout, least significant bit first.
func Gia_ManDupFlip(src *Gia_Man_t, initial []int32) *Gia_Man_t {
	if len(initial) < int((src.nRegs+31)/32) {
		panic("gia: short packed register initialization")
	}
	dst := giaDupMappedStart(src)
	bit := func(index int32) int32 { return int32(uint32(initial[index>>5]) >> uint(index&31) & 1) }
	for i := int32(0); i < Gia_ManCiNum(src); i++ {
		literal := Gia_ManAppendCi(dst)
		if i >= Gia_ManPiNum(src) {
			literal ^= bit(i - Gia_ManPiNum(src))
		}
		Gia_ManCi(src, i).Value = uint32(literal)
	}
	// CI copy literals retain their state-coordinate polarity; they cannot
	// use the regular-literal CI remapper after a physical-order copy.
	giaDupMappedNodes(dst, src)
	for i := int32(0); i < Gia_ManCoNum(src); i++ {
		obj := Gia_ManCo(src, i)
		literal := Gia_ObjFanin0Copy(obj)
		if i >= Gia_ManPoNum(src) {
			literal ^= bit(i - Gia_ManPoNum(src))
		}
		obj.Value = uint32(Gia_ManAppendCo(dst, literal))
	}
	Gia_ManSetRegNum(dst, src.nRegs)
	giaCopyOutputRoles(dst, src)
	return dst
}
func Gia_ManCycle(p *Gia_Man_t, cex *Abc_Cex_t, frames int32) {
	Gia_ManRandom(1)
	if frames < 0 || cex != nil && (frames > cex.IFrame || cex.NPis != Gia_ManPiNum(p) || cex.NRegs != p.nRegs) {
		panic("gia: invalid simulation frame/counterexample dimensions")
	}
	p.pObjs[0].fMark0 = 0
	value := func(lit int32) uint8 { return (p.pObjs[lit>>1].fMark0 ^ uint8(lit&1)) & 1 }
	for frame := int32(0); frame < frames; frame++ {
		for i := int32(0); i < Gia_ManPiNum(p); i++ {
			var mark uint8
			if cex != nil {
				index := p.nRegs + frame*cex.NPis + i
				mark = uint8(cex.PData[index>>5] >> uint(index&31) & 1)
			} else {
				mark = uint8(Gia_ManRandom(0) & 1)
			}
			Gia_ManPi(p, i).fMark0 = mark
		}
		for _, obj := range p.pObjs {
			a, b, c, kind := p.Gate(obj.id)
			switch kind {
			case 2:
				obj.fMark0 = value(a) & value(b)
			case 3:
				obj.fMark0 = value(a) ^ value(b)
			case 4:
				if value(a) != 0 {
					obj.fMark0 = value(b)
				} else {
					obj.fMark0 = value(c)
				}
			case 5:
				obj.fMark0 = value(a)
			}
		}
		for i := int32(0); i < Gia_ManCoNum(p); i++ {
			obj := Gia_ManCo(p, i)
			obj.fMark0 = value(Gia_ObjFaninLit0p(p, obj))
		}
		for i := int32(0); i < p.nRegs; i++ {
			Gia_ManRo(p, i).fMark0 = Gia_ManRi(p, i).fMark0
		}
	}
}
func Gia_ManDupCycled(p *Gia_Man_t, cex *Abc_Cex_t, frames int32) *Gia_Man_t {
	Gia_ManCleanMark0(p)
	Gia_ManCycle(p, cex, frames)
	bits := make([]int32, (p.nRegs+31)/32)
	for i := int32(0); i < p.nRegs; i++ {
		if Gia_ObjMark0(Gia_ManRo(p, i)) != 0 {
			bits[i>>5] |= int32(uint32(1) << uint(i&31))
		}
	}
	result := Gia_ManDupFlip(p, bits)
	Gia_ManCleanMark0(p)
	return result
}
