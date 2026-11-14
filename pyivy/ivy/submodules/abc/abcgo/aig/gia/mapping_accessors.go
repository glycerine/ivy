package gia

func Gia_ObjSimWords(p *Gia_Man_t) int32 {
	inputs := Gia_ManPiNum(p)
	if inputs <= 0 {
		panic("gia: PI simulation requires primary inputs")
	}
	return int32(Vec_WrdSize(p.VSimsPi)) / inputs
}
func giaSimulationSlice(data *Vec_Wrd_t, index, words int32) []word {
	start, end := int(index*words), int((index+1)*words)
	return Vec_WrdArray(data)[start:end:end]
}

// Simulation views borrow exactly one input/object's word block.
func Gia_ObjSimPi(p *Gia_Man_t, index int32) []word {
	return giaSimulationSlice(p.VSimsPi, index, Gia_ObjSimWords(p))
}
func Gia_ObjSim(p *Gia_Man_t, id int32) []word {
	return giaSimulationSlice(p.vSims, id, Gia_ObjSimWords(p))
}
func Gia_ObjSimObj(p *Gia_Man_t, obj *Gia_Obj_t) []word { return Gia_ObjSim(p, Gia_ObjId(p, obj)) }
func Gia_ObjFoffsetId(p *Gia_Man_t, id int32) int32     { return Vec_IntEntry(p.vFanout, int(id)) }
func Gia_ObjFoffset(p *Gia_Man_t, obj *Gia_Obj_t) int32 {
	return Gia_ObjFoffsetId(p, Gia_ObjId(p, obj))
}
func Gia_ManHasMapping(p *Gia_Man_t) int32 { return giaBool(p.VMapping != nil) }
func Gia_ObjIsLut(p *Gia_Man_t, id int32) int32 {
	return giaBool(Vec_IntEntry(p.VMapping, int(id)) != 0)
}
func Gia_ObjLutSize(p *Gia_Man_t, id int32) int32 {
	return Vec_IntEntry(p.VMapping, int(Vec_IntEntry(p.VMapping, int(id))))
}

// giaMappingFanins returns a bounded borrowed record view, excluding its ID.
func giaMappingFanins(mapping *Vec_Int_t, id int32) []int32 {
	offset := int(Vec_IntEntry(mapping, int(id)))
	count := int(Vec_IntEntry(mapping, offset))
	start := offset + 1
	end := start + count
	return Vec_IntArray(mapping)[start:end:end]
}
func Gia_ObjLutFanins(p *Gia_Man_t, id int32) []int32     { return giaMappingFanins(p.VMapping, id) }
func Gia_ObjLutFanin(p *Gia_Man_t, id, index int32) int32 { return Gia_ObjLutFanins(p, id)[index] }
func Gia_ObjLutMuxId(p *Gia_Man_t, id int32) int32 {
	offset := Vec_IntEntry(p.VMapping, int(id))
	return Vec_IntEntry(p.VMapping, int(offset+1+Gia_ObjLutSize(p, id)))
}
func Gia_ObjLutIsMux(p *Gia_Man_t, id int32) int32 { return giaBool(Gia_ObjLutMuxId(p, id) < 0) }
func Gia_ManHasMapping2(p *Gia_Man_t) int32        { return giaBool(p.vMapping2 != nil) }
func Gia_ObjIsLut2(p *Gia_Man_t, id int32) int32 {
	return giaBool(Vec_IntSize(Vec_WecEntry(p.vMapping2, int(id))) != 0)
}
func Gia_ObjLutSize2(p *Gia_Man_t, id int32) int32 {
	return int32(Vec_IntSize(Vec_WecEntry(p.vMapping2, int(id))))
}
func Gia_ObjLutFanins2(p *Gia_Man_t, id int32) *Vec_Int_t { return Vec_WecEntry(p.vMapping2, int(id)) }
func Gia_ObjLutFanin2(p *Gia_Man_t, id, index int32) int32 {
	return Vec_IntEntry(Vec_WecEntry(p.vMapping2, int(id)), int(index))
}
func Gia_ObjLutFanoutNum2(p *Gia_Man_t, id int32) int32 {
	return int32(Vec_IntSize(Vec_WecEntry(p.vFanouts2, int(id))))
}
func Gia_ObjLutFanout2(p *Gia_Man_t, id, index int32) int32 {
	return Vec_IntEntry(Vec_WecEntry(p.vFanouts2, int(id)), int(index))
}
func Gia_ManHasCellMapping(p *Gia_Man_t) int32 { return giaBool(p.VCellMapping != nil) }
func Gia_ObjIsCell(p *Gia_Man_t, lit int32) int32 {
	return giaBool(Vec_IntEntry(p.VCellMapping, int(lit)) != 0)
}
func Gia_ObjIsCellInv(p *Gia_Man_t, lit int32) int32 {
	return giaBool(Vec_IntEntry(p.VCellMapping, int(lit)) == -1)
}
func Gia_ObjIsCellBuf(p *Gia_Man_t, lit int32) int32 {
	return giaBool(Vec_IntEntry(p.VCellMapping, int(lit)) == -2)
}
func Gia_ObjCellSize(p *Gia_Man_t, lit int32) int32 {
	return Vec_IntEntry(p.VCellMapping, int(Vec_IntEntry(p.VCellMapping, int(lit))))
}
func Gia_ObjCellFanins(p *Gia_Man_t, lit int32) []int32     { return giaMappingFanins(p.VCellMapping, lit) }
func Gia_ObjCellFanin(p *Gia_Man_t, lit, index int32) int32 { return Gia_ObjCellFanins(p, lit)[index] }
func Gia_ObjCellId(p *Gia_Man_t, lit int32) int32 {
	offset := Vec_IntEntry(p.VCellMapping, int(lit))
	return Vec_IntEntry(p.VCellMapping, int(offset+1+Gia_ObjCellSize(p, lit)))
}

// LUT implements the dependency-neutral mapped-network boundary. Returned
// leaves/truth are independently owned and include the actual output phase.
func (p *Gia_Man_t) LUT(id int32) (leaves []int32, truth []uint64, ok bool) {
	if id < 0 || id >= p.nObjs || p.VMapping == nil || Gia_ObjIsLut(p, id) == 0 {
		return nil, nil, false
	}
	leaves = append([]int32(nil), Gia_ObjLutFanins(p, id)...)
	truth = giaNewTruthEval(p, int32(len(leaves)), leaves).node(id)
	return leaves, truth, true
}

func Gia_ManLutSizeMax(p *Gia_Man_t) int32 {
	maximum := int32(-1)
	if p.VMapping == nil {
		return maximum
	}
	for id := int32(1); id < p.nObjs; id++ {
		if Gia_ObjIsLut(p, id) != 0 {
			maximum = max(maximum, Gia_ObjLutSize(p, id))
		}
	}
	return maximum
}
