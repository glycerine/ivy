package gia

func Gia_ManCombMarkUsed_rec(p *Gia_Man_t, obj *Gia_Obj_t) int32 {
	if obj == nil {
		return 0
	}
	stack := []int32{Gia_ObjId(p, obj)}
	var count int32
	for len(stack) > 0 {
		id := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		node := Gia_ManObj(p, id)
		if node.fMark0 == 0 {
			continue
		}
		node.fMark0 = 0
		if Gia_ObjIsAnd(node) == 0 || Gia_ObjIsBuf(node) != 0 {
			panic("gia: combinational marking expected a non-buffer node")
		}
		count++
		fanins := giaObjectFanins(p, node)
		// Follow source equivalence links as reachability edges as well.
		if p.PNexts != nil && p.PNexts[id] > 0 {
			fanins = append(fanins, p.PNexts[id])
		}
		if sibling := Gia_ObjSibl(p, id); sibling != 0 {
			fanins = append(fanins, sibling)
		}
		for i := len(fanins) - 1; i >= 0; i-- {
			stack = append(stack, fanins[i])
		}
	}
	return count
}
func Gia_ManCombMarkUsed(p *Gia_Man_t) int32 {
	for _, obj := range p.pObjs {
		obj.fMark0 = uint32(giaBool(Gia_ObjIsAnd(obj) != 0 && Gia_ObjIsBuf(obj) == 0))
	}
	var count int32
	for _, obj := range p.pObjs {
		if Gia_ObjIsBuf(obj) != 0 {
			count += Gia_ManCombMarkUsed_rec(p, Gia_ObjFanin0(obj))
		}
	}
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		count += Gia_ManCombMarkUsed_rec(p, Gia_ObjFanin0(Gia_ManCo(p, i)))
	}
	return count
}
func Gia_ManCleanup(p *Gia_Man_t) *Gia_Man_t { Gia_ManCombMarkUsed(p); return Gia_ManDupMarked(p) }
func Gia_ManCleanupOutputs(p *Gia_Man_t, count int32) *Gia_Man_t {
	if p.nRegs != 0 || count < 0 || count >= Gia_ManCoNum(p) {
		panic("gia: invalid cleanup output prefix")
	}
	Gia_ManCombMarkUsed(p)
	for i := int32(0); i < count; i++ {
		Gia_ManCo(p, i).fMark0 = 1
	}
	return Gia_ManDupMarked(p)
}
func Gia_ManSeqMarkUsed_rec(p *Gia_Man_t, obj *Gia_Obj_t, roots *Vec_Int_t) int32 {
	stack := []int32{Gia_ObjId(p, obj)}
	var count int32
	for len(stack) > 0 {
		id := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		node := Gia_ManObj(p, id)
		if node.fMark0 == 0 {
			continue
		}
		node.fMark0 = 0
		if Gia_ObjIsRo(p, node) != 0 {
			Vec_IntPush(roots, Gia_ObjId(p, Gia_ObjRoToRi(p, node)))
			continue
		}
		if Gia_ObjIsCo(node) == 0 && Gia_ObjIsAnd(node) == 0 {
			panic("gia: invalid sequential reachability object")
		}
		if Gia_ObjIsAnd(node) != 0 {
			count++
		}
		fanins := giaObjectFanins(p, node)
		for i := len(fanins) - 1; i >= 0; i-- {
			stack = append(stack, fanins[i])
		}
	}
	return count
}
func Gia_ManSeqMarkUsed(p *Gia_Man_t) int32 {
	Gia_ManSetMark0(p)
	p.pObjs[0].fMark0 = 0
	for i := int32(0); i < Gia_ManPiNum(p); i++ {
		Gia_ManPi(p, i).fMark0 = 0
	}
	roots := Gia_ManCollectPoIds(p)
	var count int32
	for i := 0; i < Vec_IntSize(roots); i++ {
		count += Gia_ManSeqMarkUsed_rec(p, Gia_ManObj(p, Vec_IntEntry(roots, i)), roots)
	}
	return count
}
func Gia_ManSeqCleanup(p *Gia_Man_t) *Gia_Man_t { Gia_ManSeqMarkUsed(p); return Gia_ManDupMarked(p) }
