package gia

// Gia_Min2ManSimulate returns one when the sampled patterns do not witness
// a satisfying assignment for the conjunction of the two root literals.
func Gia_Min2ManSimulate(p *Gia_Man2Min_t) int32 {
	p.nSims++
	Vec_WrdFillExtra(p.vSims, int(Gia_ManObjNum(p.pGia)), 0)
	for i := p.nCisOld; i < Gia_ManCiNum(p.pGia); i++ {
		Vec_WrdWriteEntry(p.vSims, int(Gia_ManCiIdToId(p.pGia, i)), Gia_ManRandomW(0)<<1)
	}
	p.nCisOld = Gia_ManCiNum(p.pGia)
	for _, nodes := range p.vObjs {
		for _, id := range Vec_IntArray(nodes) {
			obj := Gia_ManObj(p.pGia, id)
			a := Vec_WrdEntry(p.vSims, int(Gia_ObjFaninId0(obj, id)))
			b := Vec_WrdEntry(p.vSims, int(Gia_ObjFaninId1(obj, id)))
			if Gia_ObjFaninC0(obj) != 0 {
				a = ^a
			}
			if Gia_ObjFaninC1(obj) != 0 {
				b = ^b
			}
			Vec_WrdWriteEntry(p.vSims, int(id), a&b)
		}
	}
	var roots [2]uint64
	for i, lit := range p.iLits {
		roots[i] = Vec_WrdEntry(p.vSims, int(lit>>1))
		if lit&1 != 0 {
			roots[i] = ^roots[i]
		}
	}
	return giaBool(roots[0]&roots[1] == 0)
}

func Gia_Min2SimSetInputBit(p *Gia_Man2Min_t, id, bit, pattern int32) {
	if pattern <= 0 || pattern >= 64 {
		panic("gia: invalid simulation pattern")
	}
	sim := Vec_WrdEntryP(p.vSims, int(id))
	mask := uint64(1) << uint(pattern)
	if bit != 0 {
		*sim |= mask
	} else {
		*sim &^= mask
	}
}
