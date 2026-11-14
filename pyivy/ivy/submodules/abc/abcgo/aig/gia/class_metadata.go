package gia

func Gia_ObjSibl(p *Gia_Man_t, id int32) int32 {
	if p.pSibls == nil {
		return 0
	}
	return p.pSibls[id]
}
func Gia_ObjSiblObj(p *Gia_Man_t, id int32) *Gia_Obj_t {
	if sibling := Gia_ObjSibl(p, id); sibling != 0 {
		return Gia_ManObj(p, sibling)
	}
	return nil
}
func Gia_ObjProved(p *Gia_Man_t, id int32) int32 { return int32(p.PReprs[id].fProved) }
func Gia_ObjSetProved(p *Gia_Man_t, id int32)    { p.PReprs[id].fProved = 1 }
func Gia_ObjUnsetProved(p *Gia_Man_t, id int32)  { p.PReprs[id].fProved = 0 }
func Gia_ObjFailed(p *Gia_Man_t, id int32) int32 { return int32(p.PReprs[id].fFailed) }
func Gia_ObjSetFailed(p *Gia_Man_t, id int32)    { p.PReprs[id].fFailed = 1 }
func Gia_ObjColor(p *Gia_Man_t, id, color int32) int32 {
	if color != 0 {
		return int32(p.PReprs[id].fColorB)
	}
	return int32(p.PReprs[id].fColorA)
}
func Gia_ObjColors(p *Gia_Man_t, id int32) int32 {
	return int32(p.PReprs[id].fColorB*2 + p.PReprs[id].fColorA)
}
func Gia_ObjSetColor(p *Gia_Man_t, id, color int32) {
	if color != 0 {
		p.PReprs[id].fColorB = 1
	} else {
		p.PReprs[id].fColorA = 1
	}
}
func Gia_ObjSetColors(p *Gia_Man_t, id int32) { p.PReprs[id].fColorA = 1; p.PReprs[id].fColorB = 1 }
func Gia_ObjVisitColor(p *Gia_Man_t, id, color int32) int32 {
	old := Gia_ObjColor(p, id, color)
	Gia_ObjSetColor(p, id, color)
	return old
}
func Gia_ObjDiffColors(p *Gia_Man_t, i, j int32) int32 {
	return giaBool(p.PReprs[i].fColorA != p.PReprs[j].fColorA && p.PReprs[i].fColorB != p.PReprs[j].fColorB)
}
func Gia_ObjDiffColors2(p *Gia_Man_t, i, j int32) int32 {
	return giaBool(p.PReprs[i].fColorA != p.PReprs[j].fColorA || p.PReprs[i].fColorB != p.PReprs[j].fColorB)
}
func Gia_ObjNextObj(p *Gia_Man_t, id int32) *Gia_Obj_t {
	if next := p.PNexts[id]; next != 0 {
		return Gia_ManObj(p, next)
	}
	return nil
}
func Gia_ObjNext(p *Gia_Man_t, id int32) int32    { return p.PNexts[id] }
func Gia_ObjSetNext(p *Gia_Man_t, id, next int32) { p.SetNext(id, next) }
func Gia_ObjIsConst(p *Gia_Man_t, id int32) int32 { return giaBool(Gia_ObjRepr(p, id) == 0) }
func Gia_ObjIsHead(p *Gia_Man_t, id int32) int32 {
	return giaBool(Gia_ObjRepr(p, id) == GIA_VOID && p.PNexts[id] > 0)
}
func Gia_ObjIsNone(p *Gia_Man_t, id int32) int32 {
	return giaBool(Gia_ObjRepr(p, id) == GIA_VOID && p.PNexts[id] <= 0)
}
func Gia_ObjIsTail(p *Gia_Man_t, id int32) int32 {
	repr := Gia_ObjRepr(p, id)
	return giaBool(repr > 0 && repr != GIA_VOID && p.PNexts[id] <= 0)
}
func Gia_ObjIsClass(p *Gia_Man_t, id int32) int32 {
	repr := Gia_ObjRepr(p, id)
	return giaBool((repr > 0 && repr != GIA_VOID) || p.PNexts[id] > 0)
}
func Gia_ObjHasSameRepr(p *Gia_Man_t, i, k int32) int32 {
	if k == 0 {
		panic("gia: second representative object must be nonzero")
	}
	if i == 0 {
		return Gia_ObjIsConst(p, k)
	}
	repr := Gia_ObjRepr(p, i)
	return giaBool(repr != GIA_VOID && repr == Gia_ObjRepr(p, k))
}
func Gia_ObjIsFailedPair(p *Gia_Man_t, i, k int32) int32 {
	if k == 0 {
		panic("gia: second pair object must be nonzero")
	}
	return giaBool(Gia_ObjFailed(p, k) != 0 || i != 0 && Gia_ObjFailed(p, i) != 0)
}
func Gia_ClassIsPair(p *Gia_Man_t, id int32) int32 {
	if Gia_ObjIsHead(p, id) == 0 || p.PNexts[id] == 0 {
		panic("gia: expected class head")
	}
	return giaBool(p.PNexts[p.PNexts[id]] <= 0)
}
func Gia_ClassUndoPair(p *Gia_Man_t, id int32) {
	if Gia_ClassIsPair(p, id) == 0 {
		panic("gia: expected pair class")
	}
	Gia_ObjSetRepr(p, p.PNexts[id], GIA_VOID)
	p.PNexts[id] = 0
}

// Sibling exposes the optional source choice chain through neutral algorithms.
func (p *Gia_Man_t) Sibling(id int32) int32 { return Gia_ObjSibl(p, id) }
func (p *Gia_Man_t) MuxesEnabled() bool     { return p.PMuxes != nil }
