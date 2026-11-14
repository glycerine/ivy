package gia

// MFFC reference walks share the manager's structural reference counts. The
// temporary dereference/reference pair restores every count before returning.
type giaRefFrame struct {
	fanins []int32
	next   int
}

func giaChangeConeReferences(p *Gia_Man_t, root *Gia_Obj_t, delta int32, mark bool) int32 {
	Gia_ObjId(p, root)
	if root.id == 0 || Gia_ObjIsCi(root) != 0 {
		return 0
	}
	if Gia_ObjIsAnd(root) == 0 {
		panic("gia: MFFC root must be an internal node")
	}
	if mark {
		Gia_ObjSetTravIdCurrent(p, root)
	}
	stack := []giaRefFrame{{fanins: giaObjectFanins(p, root)}}
	count := int32(1)
	for len(stack) > 0 {
		frame := &stack[len(stack)-1]
		if frame.next == len(frame.fanins) {
			stack = stack[:len(stack)-1]
			continue
		}
		id := frame.fanins[frame.next]
		frame.next++
		old := p.Refs[id]
		if delta < 0 && old <= 0 {
			panic("gia: MFFC reference underflow")
		}
		p.Refs[id] += delta
		descend := delta < 0 && p.Refs[id] == 0 || delta > 0 && old == 0
		if !descend {
			continue
		}
		obj := Gia_ManObj(p, id)
		if id == 0 || Gia_ObjIsCi(obj) != 0 {
			continue
		}
		if Gia_ObjIsAnd(obj) == 0 {
			panic("gia: invalid MFFC fanin")
		}
		if mark {
			Gia_ObjSetTravIdCurrent(p, obj)
		}
		count++
		stack = append(stack, giaRefFrame{fanins: giaObjectFanins(p, obj)})
	}
	return count
}
func Gia_NodeDeref_rec(p *Gia_Man_t, node *Gia_Obj_t) int32 {
	return giaChangeConeReferences(p, node, -1, false)
}
func Gia_NodeRef_rec(p *Gia_Man_t, node *Gia_Obj_t, mark int32) int32 {
	return giaChangeConeReferences(p, node, 1, mark != 0)
}
func Gia_ManPoMffcSize(p *Gia_Man_t) int32 {
	Gia_ManCreateRefs(p)
	return Gia_NodeDeref_rec(p, Gia_ObjFanin0(Gia_ManPo(p, 0)))
}
func Gia_NodeMffcSize(p *Gia_Man_t, node *Gia_Obj_t) int32 {
	if Gia_IsComplement(node) != 0 || Gia_ObjIsCand(node) == 0 && node.id != 0 {
		panic("gia: invalid MFFC root")
	}
	count := Gia_NodeDeref_rec(p, node)
	if restored := Gia_NodeRef_rec(p, node, 0); restored != count {
		panic("gia: inconsistent MFFC reference restoration")
	}
	return count
}
func Gia_NodeMffcSizeMark(p *Gia_Man_t, node *Gia_Obj_t) int32 {
	if Gia_IsComplement(node) != 0 || Gia_ObjIsCand(node) == 0 && node.id != 0 {
		panic("gia: invalid MFFC root")
	}
	count := Gia_NodeDeref_rec(p, node)
	Gia_ManIncrementTravId(p)
	if restored := Gia_NodeRef_rec(p, node, 1); restored != count {
		panic("gia: inconsistent marked MFFC restoration")
	}
	return count
}
func Gia_NodeCollect_rec(p *Gia_Man_t, node *Gia_Obj_t, support *Vec_Int_t) {
	stack := []int32{Gia_ObjId(p, node)}
	for len(stack) > 0 {
		id := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if Gia_ObjIsTravIdCurrentId(p, id) != 0 {
			continue
		}
		Gia_ObjSetTravIdCurrentId(p, id)
		obj := Gia_ManObj(p, id)
		if p.Refs[id] != 0 || Gia_ObjIsCi(obj) != 0 {
			Vec_IntPush(support, id)
			continue
		}
		if id == 0 {
			continue
		}
		if Gia_ObjIsAnd(obj) == 0 {
			panic("gia: invalid MFFC support object")
		}
		fanins := giaObjectFanins(p, obj)
		for i := len(fanins) - 1; i >= 0; i-- {
			stack = append(stack, fanins[i])
		}
	}
}
func Gia_NodeMffcSizeSupp(p *Gia_Man_t, node *Gia_Obj_t, support *Vec_Int_t) int32 {
	if Gia_IsComplement(node) != 0 || Gia_ObjIsAnd(node) == 0 {
		panic("gia: support MFFC requires an internal node")
	}
	Vec_IntClear(support)
	Gia_ManIncrementTravId(p)
	count := Gia_NodeDeref_rec(p, node)
	for _, id := range giaObjectFanins(p, node) {
		Gia_NodeCollect_rec(p, Gia_ManObj(p, id), support)
	}
	if restored := Gia_NodeRef_rec(p, node, 0); restored != count {
		panic("gia: inconsistent support MFFC restoration")
	}
	return count
}
func Gia_NodeMffcMapping_rec(p *Gia_Man_t, id int32, mapping, support *Vec_Int_t) int32 {
	if id == 0 || Vec_IntEntry(mapping, int(id)) != 0 {
		return 0
	}
	obj := Gia_ManObj(p, id)
	if Gia_ObjIsCi(obj) != 0 {
		return 0
	}
	Gia_NodeMffcSizeSupp(p, obj, support)
	Vec_IntSort(support, 0)
	Vec_IntWriteEntry(mapping, int(id), int32(Vec_IntSize(mapping)))
	Vec_IntPush(mapping, int32(Vec_IntSize(support)))
	Vec_IntAppend(mapping, support)
	Vec_IntPush(mapping, id)
	// Recursive calls reuse the support scratch vector; retain this boundary.
	leaves := append([]int32(nil), Vec_IntArray(support)...)
	count := int32(1)
	for _, leaf := range leaves {
		count += Gia_NodeMffcMapping_rec(p, leaf, mapping, support)
	}
	return count
}
func Gia_NodeMffcMapping(p *Gia_Man_t) int32 {
	old := p.Refs
	p.Refs = nil
	defer func() { p.Refs = old }()
	Gia_ManCreateRefs(p)
	mapping := Vec_IntAlloc(2 * len(p.pObjs))
	Vec_IntFill(mapping, len(p.pObjs), 0)
	support := Vec_IntAlloc(100)
	var count int32
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		count += Gia_NodeMffcMapping_rec(p, Gia_ObjFaninId0p(p, Gia_ManCo(p, i)), mapping, support)
	}
	p.VMapping = mapping
	return count
}
