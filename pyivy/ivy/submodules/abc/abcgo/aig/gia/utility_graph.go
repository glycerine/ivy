package gia

func Gia_ManCollectPoIds(p *Gia_Man_t) *Vec_Int_t {
	result := Vec_IntAlloc(int(Gia_ManPoNum(p)))
	for i := int32(0); i < Gia_ManPoNum(p); i++ {
		Vec_IntPush(result, Gia_ObjId(p, Gia_ManPo(p, i)))
	}
	return result
}
func giaMarkSupport(p *Gia_Man_t, root *Gia_Obj_t, value uint32) {
	stack := []int32{Gia_ObjId(p, root)}
	for len(stack) > 0 {
		id := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		obj := Gia_ManObj(p, id)
		if id == 0 || obj.fMark0 == value {
			continue
		}
		obj.fMark0 = value
		fanins := giaObjectFanins(p, obj)
		for i := len(fanins) - 1; i >= 0; i-- {
			stack = append(stack, fanins[i])
		}
	}
}
func Gia_ManCheckSuppMark_rec(p *Gia_Man_t, obj *Gia_Obj_t)   { giaMarkSupport(p, obj, 1) }
func Gia_ManCheckSuppUnmark_rec(p *Gia_Man_t, obj *Gia_Obj_t) { giaMarkSupport(p, obj, 0) }
func Gia_ManCheckSupp_rec(p *Gia_Man_t, obj *Gia_Obj_t) int32 {
	stack := []int32{Gia_ObjId(p, obj)}
	seen := make(map[int32]bool)
	for len(stack) > 0 {
		id := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if id == 0 || seen[id] {
			continue
		}
		seen[id] = true
		node := Gia_ManObj(p, id)
		if node.fMark0 != 0 {
			return 1
		}
		stack = append(stack, giaObjectFanins(p, node)...)
	}
	return 0
}
func Gia_ManCheckSuppOverlap(p *Gia_Man_t, first, second int32) int32 {
	if first == 0 || second == 0 {
		return 0
	}
	root := Gia_ManObj(p, first)
	Gia_ManCheckSuppMark_rec(p, root)
	defer Gia_ManCheckSuppUnmark_rec(p, root)
	return Gia_ManCheckSupp_rec(p, Gia_ManObj(p, second))
}
func Gia_ManCountPisWithFanout(p *Gia_Man_t) int32 {
	Gia_ManMarkFanoutDrivers(p)
	var count int32
	for i := int32(0); i < Gia_ManCiNum(p); i++ {
		count += Gia_ObjMark0(Gia_ManCi(p, i))
	}
	Gia_ManCleanMark0(p)
	return count
}
func Gia_ManCountPosWithNonZeroDrivers(p *Gia_Man_t) int32 {
	var count int32
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		if Gia_ObjFaninLit0p(p, Gia_ManCo(p, i)) != 0 {
			count++
		}
	}
	return count
}
func Gia_ManUpdateCopy(copies *Vec_Int_t, p *Gia_Man_t) {
	for i, lit := range Vec_IntArray(copies) {
		if lit == -1 {
			continue
		}
		value := Gia_ManObj(p, lit>>1).Value
		if value == ^uint32(0) {
			Vec_IntWriteEntry(copies, i, -1)
		} else {
			Vec_IntWriteEntry(copies, i, int32(value)^(lit&1))
		}
	}
}
func Gia_ManMarkFanoutDrivers(p *Gia_Man_t) {
	Gia_ManCleanMark0(p)
	for _, obj := range p.pObjs {
		for _, id := range giaObjectFanins(p, obj) {
			p.pObjs[id].fMark0 = 1
		}
	}
}
func Gia_ManSwapPos(p *Gia_Man_t, index int32) {
	if index < 0 || index >= Gia_ManPoNum(p) {
		panic("gia: PO swap index out of range")
	}
	if index == 0 {
		return
	}
	first, other := Gia_ObjFaninLit0p(p, Gia_ManPo(p, 0)), Gia_ObjFaninLit0p(p, Gia_ManPo(p, index))
	Gia_ManPatchCoDriver(p, 0, other)
	Gia_ManPatchCoDriver(p, index, first)
}
func Gia_ManSaveValue(p *Gia_Man_t) *Vec_Int_t {
	result := Vec_IntAlloc(len(p.pObjs))
	for _, obj := range p.pObjs {
		Vec_IntPush(result, int32(obj.Value))
	}
	return result
}
func Gia_ManLoadValue(p *Gia_Man_t, values *Vec_Int_t) {
	if Vec_IntSize(values) != len(p.pObjs) {
		panic("gia: Value snapshot size mismatch")
	}
	for i, value := range Vec_IntArray(values) {
		p.pObjs[i].Value = uint32(value)
	}
}
func Gia_ManFirstFanouts(p *Gia_Man_t) *Vec_Int_t {
	result := Vec_IntStart(len(p.pObjs))
	for _, obj := range p.pObjs {
		for _, id := range giaObjectFanins(p, obj) {
			if Vec_IntEntry(result, int(id)) == 0 {
				Vec_IntWriteEntry(result, int(id), obj.id)
			}
		}
	}
	return result
}
