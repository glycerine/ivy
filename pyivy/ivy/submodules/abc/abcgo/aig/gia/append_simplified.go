package gia

func Gia_ManAppendAnd2(p *Gia_Man_t, a, b int32) int32 {
	if p.fGiaSimple == 0 {
		if a == 0 || b == 0 || a == b^1 {
			return 0
		}
		if a == 1 {
			return b
		}
		if b == 1 || a == b {
			return a
		}
	}
	return Gia_ManAppendAnd(p, a, b)
}
func Gia_ManAppendOr2(p *Gia_Man_t, a, b int32) int32 { return Gia_ManAppendAnd2(p, a^1, b^1) ^ 1 }
func Gia_ManAppendMux2(p *Gia_Man_t, control, then, otherwise int32) int32 {
	no := Gia_ManAppendAnd2(p, control^1, otherwise)
	yes := Gia_ManAppendAnd2(p, control, then)
	return Gia_ManAppendAnd2(p, no^1, yes^1) ^ 1
}
func Gia_ManAppendMaj2(p *Gia_Man_t, a, b, c int32) int32 {
	sum := Gia_ManAppendOr2(p, b, c)
	first := Gia_ManAppendAnd2(p, a, sum)
	second := Gia_ManAppendAnd2(p, b, c)
	return Gia_ManAppendOr2(p, first, second)
}
func Gia_ManAppendXor2(p *Gia_Man_t, a, b int32) int32 { return Gia_ManAppendMux2(p, a, b^1, b) }
func Gia_ManAppendXorReal2(p *Gia_Man_t, a, b int32) int32 {
	if p.fGiaSimple == 0 {
		if a < 2 {
			return b ^ a
		}
		if b < 2 {
			return a ^ b
		}
		if a == b {
			return 0
		}
		if a == b^1 {
			return 1
		}
	}
	return Gia_ManAppendXorReal(p, a, b)
}
func Gia_ManPatchCoDriver(p *Gia_Man_t, index, lit int32) {
	giaCheckFanin(p, lit)
	obj := Gia_ManCo(p, index)
	if lit>>1 >= obj.id {
		panic("gia: CO driver must precede output")
	}
	obj.iDiff0 = uint32(obj.id - (lit >> 1))
	obj.fCompl0 = uint32(lit & 1)
}
