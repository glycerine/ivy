package gia

func Gia_ObjReprObj(p *Gia_Man_t, id int32) *Gia_Obj_t {
	if repr := p.pReprs[id].iRepr; repr != GIA_VOID {
		return Gia_ManObj(p, int32(repr))
	}
	return nil
}
func Gia_ObjRepr(p *Gia_Man_t, id int32) int32 { return int32(p.pReprs[id].iRepr) }
func Gia_ObjSetRepr(p *Gia_Man_t, id, repr int32) {
	if repr != GIA_VOID && (repr < 0 || repr >= id) {
		panic("gia: representative must precede object")
	}
	p.pReprs[id].iRepr = uint32(repr)
}
func Gia_ObjSetReprRev(p *Gia_Man_t, id, repr int32) {
	if repr != GIA_VOID && (repr <= id || repr >= p.nObjs) {
		panic("gia: reverse representative must follow object")
	}
	p.pReprs[id].iRepr = uint32(repr)
}
func Gia_ObjUnsetRepr(p *Gia_Man_t, id int32)     { p.pReprs[id].iRepr = GIA_VOID }
func Gia_ObjHasRepr(p *Gia_Man_t, id int32) int32 { return giaBool(p.pReprs[id].iRepr != GIA_VOID) }
func Gia_ObjReprSelf(p *Gia_Man_t, id int32) int32 {
	if Gia_ObjHasRepr(p, id) != 0 {
		return Gia_ObjRepr(p, id)
	}
	return id
}
func Gia_ManDeriveReprs(p *Gia_Man_t) {
	if p.pReprs != nil || len(p.pNexts) < int(p.nObjs) {
		panic("gia: invalid next links for representative construction")
	}
	p.pReprs = make([]Gia_Rpr_t, p.nObjs)
	for i := range p.pReprs {
		p.pReprs[i].iRepr = GIA_VOID
	}
	for head := int32(0); head < p.nObjs; head++ {
		if p.pNexts[head] == 0 || p.pReprs[head].iRepr != GIA_VOID {
			continue
		}
		for id := p.pNexts[head]; id != 0; id = p.pNexts[id] {
			if id < 0 || id >= p.nObjs {
				panic("gia: invalid or cyclic representative chain")
			}
			if id <= head || p.pReprs[id].iRepr != GIA_VOID {
				panic("gia: representative chain must start at its lowest object")
			}
			p.pReprs[id].iRepr = uint32(head)
		}
	}
}

func Gia_ManDeriveNexts(p *Gia_Man_t) []int32 {
	if len(p.pReprs) != int(p.nObjs) {
		panic("gia: invalid representative storage")
	}
	next := make([]int32, p.nObjs)
	for id := p.nObjs - 1; id > 0; id-- {
		repr := int32(p.pReprs[id].iRepr)
		if repr == GIA_VOID || repr == 0 {
			continue
		}
		if repr < 0 || repr >= id {
			panic("gia: noncanonical representative")
		}
		next[id] = next[repr]
		next[repr] = id
	}
	return next
}

func giaReprRoot(p *Gia_Man_t, id int32) int32 {
	if id < 0 || id >= p.nObjs {
		panic("gia: invalid representative object")
	}
	root := id
	for p.pReprs[root].iRepr != GIA_VOID {
		root = int32(p.pReprs[root].iRepr)
	}
	for id != root {
		next := int32(p.pReprs[id].iRepr)
		p.pReprs[id].iRepr = uint32(root)
		id = next
	}
	return root
}
func giaMergeRepr(p *Gia_Man_t, a, b int32) {
	a, b = giaReprRoot(p, a), giaReprRoot(p, b)
	if a == b {
		return
	}
	if a > b {
		a, b = b, a
	}
	p.pReprs[b].iRepr = uint32(a)
}
func giaFlattenReprs(p *Gia_Man_t) {
	for id := int32(0); id < p.nObjs; id++ {
		giaReprRoot(p, id)
	}
}
