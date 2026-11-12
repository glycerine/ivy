// Identity-preserving object replacement ported from ABC ivyObj.c.
package ivy

func Ivy_ObjReplace(p *Ivy_Man_t, old, next *Ivy_Obj_t, deleteOld, freeTop, update int32) {
	if old == nil || next == nil || Ivy_IsComplement(old) != 0 || old.owner != p || Ivy_Regular(next).owner != p || old.recycled || Ivy_Regular(next).recycled || Ivy_ObjIsPi(old) != 0 || old == p.pConst1 || old == Ivy_Regular(next) || Ivy_ObjIsBuf(next) != 0 || (Ivy_Regular(next).Type == IVY_PO || Ivy_Regular(next).Type == IVY_ASSERT) {
		panic("ivy: invalid object replacement")
	}
	// A replacement preserves old's identity; recycling it is incompatible.
	if freeTop != 0 {
		panic("ivy: replacement must retain the original node")
	}
	if Ivy_ObjIsInTfi_rec(Ivy_Regular(next), old, -1) != 0 {
		panic("ivy: replacement introduces a combinational cycle")
	}
	if p.pHaig != nil {
		Ivy_ManHaigCreateChoice(p, old, next)
	}
	if Ivy_IsComplement(next) != 0 || Ivy_ObjIsLatch(next) != 0 || Ivy_ObjRefs(next) > 0 || Ivy_ObjIsPi(next) != 0 || Ivy_ObjIsConst1(next) != 0 {
		next = Ivy_ObjCreate(p, Ivy_ObjCreateGhost(p, next, nil, IVY_BUF, IVY_INIT_NONE))
	}
	if deleteOld != 0 {
		Ivy_ObjDelete_rec(p, old, 0)
	} else if old.Type != IVY_NONE {
		panic("ivy: replacement target was not disconnected")
	}
	refs, fanout := old.nRefs, old.pFanout
	Ivy_ObjOverwrite(old, next)
	old.nRefs = refs
	old.pFanout = fanout
	if p.fFanout != 0 {
		if a := Ivy_ObjFanin0(old); a != nil {
			Ivy_ObjPatchFanout(p, a, next, old)
		}
		if b := Ivy_ObjFanin1(old); b != nil {
			Ivy_ObjPatchFanout(p, b, next, old)
		}
	}
	Ivy_TableUpdate(p, next, old.Id)
	p.vObjs[next.Id] = nil
	Ivy_ManRecycleMemory(p, next)
	if p.fFanout != 0 && Ivy_ObjIsBuf(old) != 0 {
		p.vBufs = append(p.vBufs, old)
	}
	if update != 0 {
		if p.fFanout != 0 && Ivy_ObjIsNode(old) != 0 {
			Ivy_ObjUpdateLevel_rec(p, old)
		}
		if p.vRequired != nil {
			Ivy_ObjUpdateLevelR_rec(p, old, p.vRequired[old.Id])
		}
	}
}
func Ivy_NodeFixBufferFanins(p *Ivy_Man_t, n *Ivy_Obj_t, update int32) {
	if Ivy_ObjIsPo(n) != 0 || n.Type == IVY_ASSERT {
		if Ivy_ObjIsBuf(Ivy_ObjFanin0(n)) != 0 {
			Ivy_ObjPatchFanin0(p, n, Ivy_ObjReal(Ivy_ObjChild0(n)))
		}
		return
	}
	a, b := Ivy_ObjFanin0(n), Ivy_ObjFanin1(n)
	if (a == nil || Ivy_ObjIsBuf(a) == 0) && (b == nil || Ivy_ObjIsBuf(b) == 0) {
		return
	}
	realA := Ivy_ObjReal(Ivy_ObjChild0(n))
	var result *Ivy_Obj_t
	if Ivy_ObjIsNode(n) != 0 {
		result = Ivy_Oper(p, realA, Ivy_ObjReal(Ivy_ObjChild1(n)), n.Type)
	} else if Ivy_ObjIsLatch(n) != 0 {
		result = Ivy_Latch(p, realA, n.Init)
	} else {
		panic("ivy: unexpected buffer consumer")
	}
	Ivy_ObjReplace(p, n, result, 1, 0, update)
}
