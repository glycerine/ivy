// Core object lifecycle ported from ABC ivyObj.c (Alan Mishchenko).
package ivy

func Ivy_ObjCreatePi(p *Ivy_Man_t) *Ivy_Obj_t {
	return Ivy_ObjCreate(p, Ivy_ObjCreateGhost(p, nil, nil, IVY_PI, IVY_INIT_NONE))
}
func Ivy_ObjCreatePo(p *Ivy_Man_t, driver *Ivy_Obj_t) *Ivy_Obj_t {
	return Ivy_ObjCreate(p, Ivy_ObjCreateGhost(p, driver, nil, IVY_PO, IVY_INIT_NONE))
}
func Ivy_ObjCreate(p *Ivy_Man_t, ghost *Ivy_Obj_t) *Ivy_Obj_t {
	if ghost == nil || Ivy_IsComplement(ghost) != 0 || ghost.Id >= 0 || ghost.owner != p {
		panic("manager-owned ghost required")
	}
	if Ivy_TableLookup(p, ghost) != nil {
		panic("object already exists")
	}
	typ, init, a, b := ghost.Type, ghost.Init, ghost.pFanin0, ghost.pFanin1
	node := Ivy_ManFetchMemory(p)
	node.Id = int32(len(p.vObjs))
	node.Type, node.Init = typ, init
	p.vObjs = append(p.vObjs, node)
	Ivy_ObjConnect(p, node, a, b)
	if Ivy_ObjIsNode(node) != 0 {
		node.Level = Ivy_ObjLevelNew(node)
		if typ == IVY_EXOR {
			node.fPhase = Ivy_ObjFaninPhase(a) ^ Ivy_ObjFaninPhase(b)
		} else {
			node.fPhase = Ivy_ObjFaninPhase(a) & Ivy_ObjFaninPhase(b)
		}
		node.fFailTfo = Ivy_Regular(a).fFailTfo | Ivy_Regular(b).fFailTfo
	} else if typ == IVY_LATCH {
		node.Level = 0
		node.fPhase = Ivy_ObjFaninPhase(a)
	} else if Ivy_ObjIsOneFanin(node) != 0 {
		node.Level = Ivy_ObjLevel(Ivy_Regular(a))
		node.fPhase = Ivy_ObjFaninPhase(a)
	}
	if typ == IVY_EXOR {
		Ivy_Regular(a).fExFan = 1
		Ivy_Regular(b).fExFan = 1
	}
	if typ == IVY_PI {
		p.vPis = append(p.vPis, node)
	} else if typ == IVY_PO {
		p.vPos = append(p.vPos, node)
	}
	if p.vRequired != nil && len(p.vRequired) <= int(node.Id) {
		n := max(2*len(p.vRequired), int(node.Id)+1)
		for len(p.vRequired) < n {
			p.vRequired = append(p.vRequired, 1000000)
		}
	}
	p.nObjs[typ]++
	p.nCreated++
	if p.pHaig != nil {
		Ivy_ManHaigCreateObj(p, node)
	}
	return node
}
func Ivy_ObjConnect(p *Ivy_Man_t, node, a, b *Ivy_Obj_t) {
	if Ivy_IsComplement(node) != 0 || node.owner != p {
		panic("invalid object connect")
	}
	node.pFanin0, node.pFanin1 = a, b
	for _, edge := range []*Ivy_Obj_t{a, b} {
		if edge != nil {
			fanin := Ivy_Regular(edge)
			if fanin.owner != p || fanin.recycled {
				panic("foreign or recycled fanin")
			}
			Ivy_ObjRefsInc(fanin)
			if p.fFanout != 0 {
				Ivy_ObjAddFanout(p, fanin, node)
			}
		}
	}
	Ivy_TableInsert(p, node)
}
func Ivy_ObjDisconnect(p *Ivy_Man_t, node *Ivy_Obj_t) {
	for _, edge := range []*Ivy_Obj_t{node.pFanin0, node.pFanin1} {
		if edge != nil {
			fanin := Ivy_Regular(edge)
			Ivy_ObjRefsDec(fanin)
			if p.fFanout != 0 {
				Ivy_ObjDeleteFanout(p, fanin, node)
			}
		}
	}
	Ivy_TableDelete(p, node)
	node.pFanin0, node.pFanin1 = nil, nil
}
func Ivy_ObjPatchFanin0(p *Ivy_Man_t, node, next *Ivy_Obj_t) {
	if next == nil || Ivy_Regular(next).owner != p || Ivy_Regular(next).recycled {
		panic("invalid replacement fanin")
	}
	old := Ivy_ObjFanin0(node)
	Ivy_ObjRefsDec(old)
	if p.fFanout != 0 {
		Ivy_ObjDeleteFanout(p, old, node)
	}
	node.pFanin0 = next
	Ivy_ObjRefsInc(Ivy_Regular(next))
	if p.fFanout != 0 {
		Ivy_ObjAddFanout(p, Ivy_Regular(next), node)
	}
	if Ivy_ObjIsPi(old) == 0 && Ivy_ObjIsConst1(old) == 0 && old.nRefs == 0 {
		Ivy_ObjDelete_rec(p, old, 1)
	}
}
func ivyRemoveObject(list []*Ivy_Obj_t, node *Ivy_Obj_t) []*Ivy_Obj_t {
	for i, old := range list {
		if old == node {
			copy(list[i:], list[i+1:])
			list[len(list)-1] = nil
			return list[:len(list)-1]
		}
	}
	return list
}
func Ivy_ObjDelete(p *Ivy_Man_t, node *Ivy_Obj_t, free int32) {
	if node == nil || Ivy_IsComplement(node) != 0 || node.owner != p || node == p.pConst1 || node.recycled || node.nRefs != 0 && free != 0 {
		panic("invalid object deletion")
	}
	typ := node.Type
	p.nObjs[typ]--
	p.nDeleted++
	Ivy_ObjDisconnect(p, node)
	if typ == IVY_PI {
		p.vPis = ivyRemoveObject(p.vPis, node)
	} else if typ == IVY_PO {
		p.vPos = ivyRemoveObject(p.vPos, node)
	} else if p.fFanout != 0 && typ == IVY_BUF {
		p.vBufs = ivyRemoveObject(p.vBufs, node)
	}
	if free != 0 {
		p.vObjs[node.Id] = nil
		Ivy_ManRecycleMemory(p, node)
	} else {
		refs, fanout := node.nRefs, node.pFanout
		Ivy_ObjClean(node)
		node.nRefs, node.pFanout = refs, fanout
	}
}
func Ivy_ObjDelete_rec(p *Ivy_Man_t, node *Ivy_Obj_t, free int32) {
	if Ivy_ObjIsConst1(node) != 0 || Ivy_ObjIsPi(node) != 0 {
		return
	}
	a, b := Ivy_ObjFanin0(node), Ivy_ObjFanin1(node)
	Ivy_ObjDelete(p, node, free)
	for _, fanin := range []*Ivy_Obj_t{a, b} {
		if fanin != nil && fanin.Type != IVY_NONE && fanin.nRefs == 0 {
			Ivy_ObjDelete_rec(p, fanin, 1)
		}
	}
}
