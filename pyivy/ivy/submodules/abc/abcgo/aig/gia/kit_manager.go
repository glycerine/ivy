package gia

// These methods satisfy kit's dependency-neutral GIA manager interface.
func (p *Gia_Man_t) Const1() int32                  { return 1 }
func (p *Gia_Man_t) InputLiteral(index int32) int32 { return Gia_ObjId(p, Gia_ManCi(p, index)) << 1 }
func (p *Gia_Man_t) AndLiteral(a, b int32, hash bool) int32 {
	if !hash {
		return Gia_ManAppendAnd(p, a, b)
	}
	if Vec_IntSize(&p.vHTable) != 0 {
		return Gia_ManHashAnd(p, a, b)
	}
	giaCheckFanin(p, a)
	giaCheckFanin(p, b)
	if a == 0 || b == 0 {
		return 0
	}
	if a == 1 {
		return b
	}
	if b == 1 {
		return a
	}
	if a == b {
		return a
	}
	if a == b^1 {
		return 0
	}
	if p.fGiaSimple != 0 {
		return Gia_ManAppendAnd(p, a, b)
	}
	if a > b {
		a, b = b, a
	}
	if p.literalHash == nil {
		p.literalHash = make(map[[2]int32]int32)
		for _, obj := range p.pObjs {
			if Gia_ObjIsAnd(obj) == 0 || Gia_ObjIsXor(obj) != 0 || (int(obj.id) < len(p.PMuxes) && p.PMuxes[obj.id] != 0) {
				continue
			}
			x, y := Gia_ObjFaninLit0(obj, obj.id), Gia_ObjFaninLit1(obj, obj.id)
			if x > y {
				x, y = y, x
			}
			p.literalHash[[2]int32{x, y}] = obj.id << 1
		}
	}
	key := [2]int32{a, b}
	if lit, exists := p.literalHash[key]; exists {
		return lit
	}
	lit := Gia_ManAppendAnd(p, a, b)
	p.literalHash[key] = lit
	return lit
}
