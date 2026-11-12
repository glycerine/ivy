package ivy

import "sort"

// Ivy_BalanceStore retains one stable scratch slice per recursion depth.
type Ivy_BalanceStore struct{ cones []*[]*Ivy_Obj_t }

func Ivy_ManBalance(p *Ivy_Man_t, update int32) *Ivy_Man_t {
	if Ivy_ManCheck(p) == 0 {
		panic("ivy: balancing invalid graph")
	}
	for _, n := range p.vObjs {
		if n != nil && Ivy_ObjIsLatch(n) != 0 {
			panic("ivy: balancing requires a combinational graph")
		}
	}
	Ivy_ManCleanTravId(p)
	result := Ivy_ManStart()
	p.pConst1.TravId = int32(Ivy_EdgeFromNode(result.pConst1))
	for _, n := range p.vPis {
		n.TravId = int32(Ivy_EdgeFromNode(Ivy_ObjCreatePi(result)))
	}
	store := new(Ivy_BalanceStore)
	for _, n := range p.vPos {
		edge := Ivy_ObjReal(Ivy_ObjChild0(n))
		mapped := Ivy_NodeBalance_rec(result, Ivy_Regular(edge), store, 0, update)
		mapped = Ivy_EdgeNotCond(mapped, Ivy_IsComplement(edge))
		Ivy_ObjCreatePo(result, Ivy_EdgeToNode(result, mapped))
	}
	Ivy_ManCleanup(result)
	if Ivy_ManCheck(result) == 0 {
		panic("ivy: balancing produced invalid graph")
	}
	return result
}

func Ivy_NodeCompareLevelsDecrease(a, b *Ivy_Obj_t) int {
	a, b = Ivy_Regular(a), Ivy_Regular(b)
	if a.Level > b.Level || a.Level == b.Level && a.Id > b.Id {
		return -1
	}
	if a.Level < b.Level || a.Id < b.Id {
		return 1
	}
	return 0
}

func Ivy_NodeBalance_rec(p *Ivy_Man_t, old *Ivy_Obj_t, store *Ivy_BalanceStore, depth int, update int32) Ivy_Edge_t {
	if old == nil || Ivy_IsComplement(old) != 0 || Ivy_ObjIsBuf(old) != 0 {
		panic("ivy: invalid balance leaf")
	}
	if Ivy_ObjIsConst1(old) != 0 || old.TravId != 0 {
		return Ivy_Edge_t(uint32(old.TravId))
	}
	if Ivy_ObjIsNode(old) == 0 {
		panic("ivy: unmapped balance leaf")
	}
	super := Ivy_NodeBalanceCone(old, store, depth)
	for i, edge := range *super {
		mapped := Ivy_NodeBalance_rec(p, Ivy_Regular(edge), store, depth+1, update)
		(*super)[i] = Ivy_EdgeToNode(p, Ivy_EdgeNotCond(mapped, Ivy_IsComplement(edge)))
	}
	result := Ivy_NodeBalanceBuildSuper(p, super, old.Type, update)
	*super = (*super)[:0]
	old.TravId = int32(Ivy_EdgeFromNode(result))
	return Ivy_Edge_t(uint32(old.TravId))
}

func Ivy_NodeBalanceBuildSuper(p *Ivy_Man_t, super *[]*Ivy_Obj_t, kind Ivy_Type_t, update int32) *Ivy_Obj_t {
	if len(*super) == 0 {
		return Ivy_ManConst0(p)
	}
	sort.SliceStable(*super, func(i, j int) bool { return Ivy_NodeCompareLevelsDecrease((*super)[i], (*super)[j]) < 0 })
	for len(*super) > 1 {
		left := 0
		if update != 0 {
			left = Ivy_NodeBalanceFindLeft(*super)
		}
		exor := int32(0)
		if kind == IVY_EXOR {
			exor = 1
		}
		Ivy_NodeBalancePermute(p, *super, left, exor)
		n := len(*super)
		a, b := (*super)[n-1], (*super)[n-2]
		*super = (*super)[:n-2]
		result := Ivy_Oper(p, a, b, kind)
		if kind == IVY_EXOR { // XOR operands must retain multiplicity: x XOR x cancels.
			*super = append(*super, result)
			sort.SliceStable(*super, func(i, j int) bool { return Ivy_NodeCompareLevelsDecrease((*super)[i], (*super)[j]) < 0 })
		} else {
			Ivy_NodeBalancePushUniqueOrderByLevel(super, result)
		}
	}
	return (*super)[0]
}

// Repeated XOR leaves cancel; opposite leaves additionally toggle parity.
// AND duplicates are idempotent, while opposite leaves make the cone false.
func Ivy_NodeBalanceCone_rec(root, edge *Ivy_Obj_t, super *[]*Ivy_Obj_t) int32 {
	n := Ivy_Regular(edge)
	if edge != root && (Ivy_IsComplement(edge) != 0 || n.Type != root.Type || n.nRefs > 1 || len(*super) > 10000) {
		for i, prior := range *super {
			if prior != edge && prior != Ivy_Not(edge) {
				continue
			}
			if root.Type == IVY_AND {
				if prior != edge {
					return -1
				}
				return 1
			}
			*super = append((*super)[:i], (*super)[i+1:]...)
			if prior != edge {
				ivyBalanceToggleParity(root.owner.pConst1, super)
			}
			return 1
		}
		*super = append(*super, edge)
		return 0
	}
	if Ivy_IsComplement(edge) != 0 || Ivy_ObjIsNode(n) == 0 {
		panic("ivy: invalid supergate")
	}
	a := Ivy_NodeBalanceCone_rec(root, Ivy_ObjReal(Ivy_ObjChild0(n)), super)
	b := Ivy_NodeBalanceCone_rec(root, Ivy_ObjReal(Ivy_ObjChild1(n)), super)
	if a < 0 || b < 0 {
		return -1
	}
	if a != 0 || b != 0 {
		return 1
	}
	return 0
}
func ivyBalanceToggleParity(one *Ivy_Obj_t, super *[]*Ivy_Obj_t) {
	for i, n := range *super {
		if n == one {
			*super = append((*super)[:i], (*super)[i+1:]...)
			return
		}
	}
	*super = append(*super, one)
}
func Ivy_NodeBalanceCone(n *Ivy_Obj_t, store *Ivy_BalanceStore, depth int) *[]*Ivy_Obj_t {
	for len(store.cones) <= depth {
		row := new([]*Ivy_Obj_t)
		store.cones = append(store.cones, row)
	}
	row := store.cones[depth]
	*row = (*row)[:0]
	if Ivy_NodeBalanceCone_rec(n, n, row) < 0 {
		*row = (*row)[:0]
	}
	return row
}
func Ivy_NodeBalanceFindLeft(nodes []*Ivy_Obj_t) int {
	if len(nodes) < 3 {
		return 0
	}
	i := len(nodes) - 2
	level := Ivy_Regular(nodes[i]).Level
	for i > 0 && Ivy_Regular(nodes[i-1]).Level == level {
		i--
	}
	return i
}
func Ivy_NodeBalancePermute(p *Ivy_Man_t, nodes []*Ivy_Obj_t, left int, exor int32) {
	right := len(nodes) - 2
	if left < 0 || left > right {
		panic("ivy: invalid balance pairing")
	}
	if left == right {
		return
	}
	last := nodes[right+1]
	if Ivy_Regular(last) == p.pConst1 || Ivy_Regular(nodes[right]) == p.pConst1 {
		return
	}
	kind := IVY_AND
	if exor != 0 {
		kind = IVY_EXOR
	}
	for i := right; i >= left; i-- {
		candidate := nodes[i]
		if Ivy_Regular(candidate) == p.pConst1 || Ivy_Regular(last) == Ivy_Regular(candidate) || Ivy_TableLookup(p, Ivy_ObjCreateGhost(p, last, candidate, kind, IVY_INIT_NONE)) != nil {
			nodes[i], nodes[right] = nodes[right], nodes[i]
			return
		}
	}
}
func Ivy_NodeBalancePushUniqueOrderByLevel(nodes *[]*Ivy_Obj_t, n *Ivy_Obj_t) {
	for _, old := range *nodes {
		if old == n {
			return
		}
	}
	*nodes = append(*nodes, n)
	for i := len(*nodes) - 1; i > 0 && Ivy_Regular((*nodes)[i]).Level > Ivy_Regular((*nodes)[i-1]).Level; i-- {
		(*nodes)[i], (*nodes)[i-1] = (*nodes)[i-1], (*nodes)[i]
	}
}
