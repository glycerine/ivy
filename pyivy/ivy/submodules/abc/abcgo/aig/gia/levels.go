package gia

func Gia_ManCleanLevels(p *Gia_Man_t, size int32) {
	if p.vLevels == nil {
		p.vLevels = Vec_IntAlloc(int(size))
	}
	Vec_IntFill(p.vLevels, int(size), 0)
}
func Gia_ObjLevelId(p *Gia_Man_t, id int32) int32     { return Vec_IntGetEntry(p.vLevels, int(id)) }
func Gia_ObjLevel(p *Gia_Man_t, obj *Gia_Obj_t) int32 { return Gia_ObjLevelId(p, Gia_ObjId(p, obj)) }
func Gia_ObjSetLevelId(p *Gia_Man_t, id, level int32) { Vec_IntSetEntry(p.vLevels, int(id), level) }
func Gia_ObjUpdateLevelId(p *Gia_Man_t, id, level int32) {
	Gia_ObjSetLevelId(p, id, max(Vec_IntEntry(p.vLevels, int(id)), level))
}
func Gia_ObjSetLevel(p *Gia_Man_t, obj *Gia_Obj_t, level int32) {
	Gia_ObjSetLevelId(p, Gia_ObjId(p, obj), level)
}
func Gia_ObjSetCoLevel(p *Gia_Man_t, obj *Gia_Obj_t) {
	if Gia_ObjIsCo(obj) == 0 {
		panic("gia: expected CO")
	}
	Gia_ObjSetLevel(p, obj, Gia_ObjLevel(p, Gia_ObjFanin0(obj)))
}
func Gia_ObjSetBufLevel(p *Gia_Man_t, obj *Gia_Obj_t) {
	if Gia_ObjIsAnd(obj) == 0 {
		panic("gia: expected buffer")
	}
	Gia_ObjSetLevel(p, obj, Gia_ObjLevel(p, Gia_ObjFanin0(obj)))
}
func Gia_ObjSetAndLevel(p *Gia_Man_t, obj *Gia_Obj_t) {
	if Gia_ObjIsAnd(obj) == 0 {
		panic("gia: expected AND")
	}
	Gia_ObjSetLevel(p, obj, 1+max(Gia_ObjLevel(p, Gia_ObjFanin0(obj)), Gia_ObjLevel(p, Gia_ObjFanin1(obj))))
}
func Gia_ObjSetXorLevel(p *Gia_Man_t, obj *Gia_Obj_t) {
	if Gia_ObjIsXor(obj) == 0 {
		panic("gia: expected XOR")
	}
	Gia_ObjSetLevel(p, obj, 2+max(Gia_ObjLevel(p, Gia_ObjFanin0(obj)), Gia_ObjLevel(p, Gia_ObjFanin1(obj))))
}
func Gia_ObjSetMuxLevel(p *Gia_Man_t, obj *Gia_Obj_t) {
	if Gia_ObjIsMux(p, obj) == 0 {
		panic("gia: expected MUX")
	}
	Gia_ObjSetLevel(p, obj, 2+max(Gia_ObjLevel(p, Gia_ObjFanin0(obj)), Gia_ObjLevel(p, Gia_ObjFanin1(obj)), Gia_ObjLevel(p, Gia_ObjFanin2(p, obj))))
}
func Gia_ObjSetGateLevel(p *Gia_Man_t, obj *Gia_Obj_t) {
	switch {
	case p.fGiaSimple == 0 && Gia_ObjIsBuf(obj) != 0:
		Gia_ObjSetBufLevel(p, obj)
	case Gia_ObjIsMux(p, obj) != 0:
		Gia_ObjSetMuxLevel(p, obj)
	case Gia_ObjIsXor(obj) != 0:
		Gia_ObjSetXorLevel(p, obj)
	case Gia_ObjIsAnd(obj) != 0:
		Gia_ObjSetAndLevel(p, obj)
	}
}

func Gia_ManLevelNum(p *Gia_Man_t) int32 {
	Gia_ManCleanLevels(p, p.nObjs)
	p.nLevels = 0
	for _, obj := range p.pObjs {
		switch {
		case p.fGiaSimple == 0 && Gia_ObjIsBuf(obj) != 0:
			Gia_ObjSetBufLevel(p, obj)
		case Gia_ObjIsAnd(obj) != 0:
			Gia_ObjSetGateLevel(p, obj)
		case Gia_ObjIsCo(obj) != 0:
			Gia_ObjSetCoLevel(p, obj)
		default:
			Gia_ObjSetLevel(p, obj, 0)
		}
		p.nLevels = max(p.nLevels, Gia_ObjLevel(p, obj))
	}
	return p.nLevels
}

// Gia_ManLevelRNum retains the source's unit reverse-edge weighting, which is
// deliberately different from forward real XOR/MUX gate weights.
func Gia_ManLevelRNum(p *Gia_Man_t) int32 {
	Gia_ManCleanLevels(p, p.nObjs)
	p.nLevels = 0
	for i := len(p.pObjs) - 1; i >= 0; i-- {
		obj := p.pObjs[i]
		switch {
		case p.fGiaSimple == 0 && Gia_ObjIsBuf(obj) != 0:
			Gia_ObjUpdateLevelId(p, Gia_ObjFaninId0(obj, int32(i)), Gia_ObjLevel(p, obj))
		case Gia_ObjIsAnd(obj) != 0:
			Gia_ObjUpdateLevelId(p, Gia_ObjFaninId0(obj, int32(i)), 1+Gia_ObjLevel(p, obj))
			Gia_ObjUpdateLevelId(p, Gia_ObjFaninId1(obj, int32(i)), 1+Gia_ObjLevel(p, obj))
		case Gia_ObjIsCo(obj) != 0:
			Gia_ObjUpdateLevelId(p, Gia_ObjFaninId0(obj, int32(i)), 1)
		default:
			p.nLevels = max(p.nLevels, Gia_ObjLevel(p, obj))
		}
	}
	return p.nLevels
}
func Gia_ManLevelAve(p *Gia_Man_t) float32 {
	if p.vLevels == nil {
		panic("gia: levels have not been computed")
	}
	var sum int32
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		sum += Gia_ObjLevel(p, Gia_ManCo(p, i))
	}
	return float32(sum) / float32(Gia_ManCoNum(p))
}

func Gia_ManGetCiLevels(p *Gia_Man_t) *Vec_Int_t {
	if p.vLevels == nil {
		return nil
	}
	levels := Vec_IntAlloc(int(Gia_ManCiNum(p)))
	for i := int32(0); i < Gia_ManCiNum(p); i++ {
		Vec_IntPush(levels, Gia_ObjLevel(p, Gia_ManCi(p, i)))
	}
	return levels
}
func Gia_ManSetLevels(p *Gia_Man_t, inputs *Vec_Int_t) int32 {
	if inputs == nil {
		return Gia_ManLevelNum(p)
	}
	if Vec_IntSize(inputs) != int(Gia_ManCiNum(p)) {
		panic("gia: input level count mismatch")
	}
	Gia_ManCleanLevels(p, p.nObjs)
	p.nLevels = 0
	for i := int32(0); i < Gia_ManCiNum(p); i++ {
		level := Vec_IntEntry(inputs, int(i))
		Gia_ObjSetLevel(p, Gia_ManCi(p, i), level)
		p.nLevels = max(p.nLevels, level)
	}
	for _, obj := range p.pObjs {
		if Gia_ObjIsAnd(obj) != 0 {
			Gia_ObjSetGateLevel(p, obj)
		} else if Gia_ObjIsCo(obj) != 0 {
			Gia_ObjSetCoLevel(p, obj)
		} else {
			continue
		}
		p.nLevels = max(p.nLevels, Gia_ObjLevel(p, obj))
	}
	return p.nLevels
}

// Gia_ManReverseLevel uses gate weights and includes dangling AND cones, as in C.
func Gia_ManReverseLevel(p *Gia_Man_t) *Vec_Int_t {
	result := Vec_IntStart(int(p.nObjs))
	for i := len(p.pObjs) - 1; i >= 0; i-- {
		obj := p.pObjs[i]
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		level := Vec_IntEntry(result, i)
		switch {
		case Gia_ObjIsMux(p, obj) != 0:
			Vec_IntUpdateEntry(result, int(Gia_ObjFaninId0(obj, int32(i))), level+2)
			Vec_IntUpdateEntry(result, int(Gia_ObjFaninId1(obj, int32(i))), level+2)
			Vec_IntUpdateEntry(result, int(Gia_ObjFaninId2(p, int32(i))), level+2)
		case Gia_ObjIsXor(obj) != 0:
			Vec_IntUpdateEntry(result, int(Gia_ObjFaninId0(obj, int32(i))), level+2)
			Vec_IntUpdateEntry(result, int(Gia_ObjFaninId1(obj, int32(i))), level+2)
		case Gia_ObjIsBuf(obj) != 0:
			Vec_IntUpdateEntry(result, int(Gia_ObjFaninId0(obj, int32(i))), level)
		default:
			Vec_IntUpdateEntry(result, int(Gia_ObjFaninId0(obj, int32(i))), level+1)
			Vec_IntUpdateEntry(result, int(Gia_ObjFaninId1(obj, int32(i))), level+1)
		}
	}
	return result
}
func Gia_ManRequiredLevel(p *Gia_Man_t) *Vec_Int_t {
	required := Gia_ManReverseLevel(p)
	var maximum int32
	for i := int32(0); i < Gia_ManCiNum(p); i++ {
		maximum = max(maximum, Vec_IntEntry(required, int(Gia_ObjId(p, Gia_ManCi(p, i)))))
	}
	for i := 0; i < int(p.nObjs); i++ {
		Vec_IntWriteEntry(required, i, maximum-Vec_IntEntry(required, i))
	}
	return required
}
func Gia_ManComputeSlacks(p *Gia_Man_t) *Vec_Int_t {
	maximum := Gia_ManLevelNum(p)
	reverse := Gia_ManReverseLevel(p)
	slacks := Vec_IntAlloc(int(p.nObjs))
	for i := int32(0); i < p.nObjs; i++ {
		Vec_IntPush(slacks, maximum-Gia_ObjLevelId(p, i)-Vec_IntEntry(reverse, int(i)))
	}
	return slacks
}
