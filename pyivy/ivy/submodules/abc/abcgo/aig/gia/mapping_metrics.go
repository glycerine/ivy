package gia

func Gia_ManLutFaninCount(p *Gia_Man_t) int32 {
	var count int32
	for id := int32(1); id < p.nObjs; id++ {
		if Gia_ObjIsLut(p, id) != 0 {
			count += Gia_ObjLutSize(p, id)
		}
	}
	return count
}
func Gia_ManLutNum(p *Gia_Man_t) int32 {
	var count int32
	for id := int32(1); id < p.nObjs; id++ {
		if Gia_ObjIsLut(p, id) != 0 {
			count++
		}
	}
	return count
}

// Gia_ManLutLevel reports the mapped depth and optionally its object levels.
func Gia_ManLutLevel(p *Gia_Man_t, result *[]int32) int32 {
	levels := make([]int32, p.nObjs)
	for id := int32(1); id < p.nObjs; id++ {
		if Gia_ObjIsLut(p, id) == 0 {
			continue
		}
		var level int32
		for _, fanin := range Gia_ObjLutFanins(p, id) {
			level = max(level, levels[fanin])
		}
		levels[id] = level + 1
	}
	var maximum int32
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		obj := Gia_ManCo(p, i)
		levels[obj.id] = levels[Gia_ObjFaninId0(obj, obj.id)]
		maximum = max(maximum, levels[obj.id])
	}
	if result != nil {
		*result = levels
	}
	return maximum
}
func giaMappedRefs(p *Gia_Man_t) []int32 {
	refs := make([]int32, p.nObjs)
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		obj := Gia_ManCo(p, i)
		refs[Gia_ObjFaninId0(obj, obj.id)]++
	}
	for id := int32(1); id < p.nObjs; id++ {
		if Gia_ObjIsLut(p, id) != 0 {
			for _, fanin := range Gia_ObjLutFanins(p, id) {
				refs[fanin]++
			}
		}
	}
	return refs
}
func Gia_ManSetRefsMapped(p *Gia_Man_t) { p.Refs = giaMappedRefs(p) }
func Gia_ManSetLutRefs(p *Gia_Man_t)    { p.pLutRefs = giaMappedRefs(p) }

func Gia_ManComputeOverlap2One_rec(p *Gia_Man_t, id int32, label *Vec_Str_t, visited *Vec_Int_t) int32 {
	if Vec_StrEntry(label, int(id)) != 0 {
		return 0
	}
	Vec_StrWriteEntry(label, int(id), 1)
	obj := Gia_ManObj(p, id)
	if Gia_ObjIsAnd(obj) == 0 {
		panic("gia: LUT cone reaches non-cut input")
	}
	count := Gia_ManComputeOverlap2One_rec(p, Gia_ObjFaninId0(obj, id), label, visited) + Gia_ManComputeOverlap2One_rec(p, Gia_ObjFaninId1(obj, id), label, visited)
	Vec_IntPush(visited, id)
	return count + 1
}
func Gia_ManComputeOverlap2One(p *Gia_Man_t, id int32, label *Vec_Str_t, visited *Vec_Int_t) int32 {
	Vec_IntClear(visited)
	for _, fanin := range Gia_ObjLutFanins(p, id) {
		Vec_StrWriteEntry(label, int(fanin), 1)
	}
	count := Gia_ManComputeOverlap2One_rec(p, id, label, visited)
	for _, fanin := range Gia_ObjLutFanins(p, id) {
		Vec_StrWriteEntry(label, int(fanin), 0)
	}
	for _, item := range Vec_IntArray(visited) {
		Vec_StrWriteEntry(label, int(item), 0)
	}
	return count
}
func Gia_ManComputeOverlap2(p *Gia_Man_t) int32 {
	label := Vec_StrStart(int(p.nObjs))
	visited := Vec_IntAlloc(100)
	count := -Gia_ManAndNum(p)
	for id := int32(1); id < p.nObjs; id++ {
		if Gia_ObjIsLut(p, id) != 0 {
			count += Gia_ManComputeOverlap2One(p, id, label, visited)
		}
	}
	return count
}
func Gia_ManComputeOverlapOne_rec(p *Gia_Man_t, id int32) int32 {
	if Gia_ObjIsTravIdCurrentId(p, id) != 0 {
		return 0
	}
	Gia_ObjSetTravIdCurrentId(p, id)
	obj := Gia_ManObj(p, id)
	if Gia_ObjIsAnd(obj) == 0 {
		panic("gia: LUT cone reaches non-cut input")
	}
	return 1 + Gia_ManComputeOverlapOne_rec(p, Gia_ObjFaninId0(obj, id)) + Gia_ManComputeOverlapOne_rec(p, Gia_ObjFaninId1(obj, id))
}
func Gia_ManComputeOverlapOne(p *Gia_Man_t, id int32) int32 {
	Gia_ManIncrementTravId(p)
	for _, fanin := range Gia_ObjLutFanins(p, id) {
		Gia_ObjSetTravIdCurrentId(p, fanin)
	}
	return Gia_ManComputeOverlapOne_rec(p, id)
}
func Gia_ManComputeOverlap(p *Gia_Man_t) int32 {
	count := -Gia_ManAndNum(p)
	for id := int32(1); id < p.nObjs; id++ {
		if Gia_ObjIsLut(p, id) != 0 {
			count += Gia_ManComputeOverlapOne(p, id)
		}
	}
	return count
}
