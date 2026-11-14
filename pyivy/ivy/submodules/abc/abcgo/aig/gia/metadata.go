package gia

func Gia_ObjMark0(p *Gia_Obj_t) int32                { return int32(p.fMark0 & 1) }
func Gia_ObjMark1(p *Gia_Obj_t) int32                { return int32(p.fMark1 & 1) }
func Gia_ObjSetMark0(p *Gia_Obj_t, value int32)      { p.fMark0 = uint32(value) & 1 }
func Gia_ObjSetMark1(p *Gia_Obj_t, value int32)      { p.fMark1 = uint32(value) & 1 }
func Gia_ObjSetPhaseValue(p *Gia_Obj_t, value int32) { p.fPhase = uint32(value) & 1 }

func Gia_ObjSetPhase(p *Gia_Man_t, obj *Gia_Obj_t) {
	a, b, c, kind := p.Gate(Gia_ObjId(p, obj))
	phase := func(lit int32) uint32 { return (p.pObjs[lit>>1].fPhase ^ uint32(lit&1)) & 1 }
	var value uint32
	switch kind {
	case 2:
		value = phase(a) & phase(b)
	case 3:
		value = phase(a) ^ phase(b)
	case 4:
		if phase(a) != 0 {
			value = phase(b)
		} else {
			value = phase(c)
		}
	case 5, 6:
		value = phase(a)
	}
	obj.fPhase = value
}
func Gia_ManSetPhase(p *Gia_Man_t) {
	for _, obj := range p.pObjs {
		Gia_ObjSetPhase(p, obj)
	}
}
func Gia_ManSetPhasePattern(p *Gia_Man_t, values *Vec_Int_t) {
	if Vec_IntSize(values) != int(Gia_ManCiNum(p)) {
		panic("gia: phase pattern size mismatch")
	}
	for _, obj := range p.pObjs {
		if Gia_ObjIsCi(obj) != 0 {
			Gia_ObjSetPhaseValue(obj, Vec_IntEntry(values, int(Gia_ObjCioId(obj))))
		} else {
			Gia_ObjSetPhase(p, obj)
		}
	}
}
func Gia_ManSetPhase1(p *Gia_Man_t) {
	for _, obj := range p.pObjs {
		if Gia_ObjIsCi(obj) != 0 {
			obj.fPhase = 1
		} else {
			Gia_ObjSetPhase(p, obj)
		}
	}
}
func (p *Gia_Man_t) HasRepresentatives() bool { return p.PReprs != nil }
func (p *Gia_Man_t) StartRepresentatives() {
	p.PReprs = make([]Gia_Rpr_t, len(p.pObjs))
	p.PNexts = make([]int32, len(p.pObjs))
	for i := range p.PReprs {
		p.PReprs[i].iRepr = GIA_VOID
	}
}
func (p *Gia_Man_t) Next(id int32) int32 { return p.PNexts[id] }
func (p *Gia_Man_t) SetNext(id, next int32) {
	if next < -1 || int(next) >= len(p.pObjs) {
		panic("gia: invalid class link")
	}
	p.PNexts[id] = next
}
func (p *Gia_Man_t) SimulationWordCount() int32 { return p.NSimWords }

// SimulationInputs returns borrowed simulation storage.
func (p *Gia_Man_t) SimulationInputs() []uint64 {
	if p.VSimsPi == nil {
		return nil
	}
	return Vec_WrdArray(p.VSimsPi)
}

func Gia_ManCleanMark01(p *Gia_Man_t) {
	for _, obj := range p.pObjs {
		obj.fMark0 = 0
		obj.fMark1 = 0
	}
}
func Gia_ManCleanMark0(p *Gia_Man_t) {
	for _, obj := range p.pObjs {
		obj.fMark0 = 0
	}
}
func Gia_ManCleanMark1(p *Gia_Man_t) {
	for _, obj := range p.pObjs {
		obj.fMark1 = 0
	}
}
func Gia_ManCleanValue(p *Gia_Man_t) {
	for _, obj := range p.pObjs {
		obj.Value = 0
	}
}
func Gia_ManFillValue(p *Gia_Man_t) {
	for _, obj := range p.pObjs {
		obj.Value = ^uint32(0)
	}
}
func Gia_ManCleanPhase(p *Gia_Man_t) {
	for _, obj := range p.pObjs {
		obj.fPhase = 0
	}
}
