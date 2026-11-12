// Graph checking ported from ABC ivyCheck.c/ivyDfs.c.
package ivy

func ivyAcyclicVisit(p *Ivy_Man_t, node *Ivy_Obj_t, state map[*Ivy_Obj_t]byte) bool {
	node = Ivy_Regular(node)
	if node == nil || node.owner != p || node.recycled {
		return false
	}
	if node == p.pConst1 || Ivy_ObjIsCi(node) != 0 {
		return true
	}
	if state[node] == 1 {
		return false
	}
	if state[node] == 2 {
		return true
	}
	state[node] = 1
	if !ivyAcyclicVisit(p, node.pFanin0, state) {
		return false
	}
	if Ivy_ObjIsNode(node) != 0 && !ivyAcyclicVisit(p, node.pFanin1, state) {
		return false
	}
	state[node] = 2
	return true
}
func Ivy_ManIsAcyclic_rec(p *Ivy_Man_t, start *Ivy_Obj_t) int32 {
	return ivyBool(ivyAcyclicVisit(p, start, map[*Ivy_Obj_t]byte{}))
}
func Ivy_ManIsAcyclic(p *Ivy_Man_t) int32 {
	state := map[*Ivy_Obj_t]byte{}
	for _, node := range p.vObjs {
		if node != nil && !ivyAcyclicVisit(p, node, state) {
			return 0
		}
	}
	return 1
}
func Ivy_ManCheck(p *Ivy_Man_t) int32 {
	if p == nil || p.memoryStopped || len(p.vObjs) == 0 || p.vObjs[0] != p.pConst1 || p.pConst1 == nil {
		return 0
	}
	var counts [IVY_VOID]int32
	refs := map[*Ivy_Obj_t]int32{}
	active := int32(0)
	for id, node := range p.vObjs {
		if node == nil {
			continue
		}
		if node.owner != p || node.recycled || node.Id != int32(id) || Ivy_IsComplement(node) != 0 {
			return 0
		}
		active++
		if id == 0 {
			if node.pFanin0 != nil || node.pFanin1 != nil || node.Level != 0 || node.fPhase != 1 {
				return 0
			}
			continue
		}
		if node.Type <= IVY_NONE || node.Type >= IVY_VOID {
			return 0
		}
		counts[node.Type]++
		if node.Type == IVY_PI {
			if node.pFanin0 != nil || node.pFanin1 != nil || node.Level != 0 {
				return 0
			}
			continue
		}
		if node.pFanin0 == nil {
			return 0
		}
		if Ivy_ObjIsNode(node) != 0 {
			if node.pFanin1 == nil || Ivy_ObjFaninId0(node) <= 0 || Ivy_ObjFaninId0(node) >= Ivy_ObjFaninId1(node) {
				return 0
			}
		} else if node.pFanin1 != nil {
			return 0
		}
		for _, edge := range []*Ivy_Obj_t{node.pFanin0, node.pFanin1} {
			if edge == nil {
				continue
			}
			fanin := Ivy_Regular(edge)
			if fanin.owner != p || fanin.recycled || fanin.Id < 0 || int(fanin.Id) >= len(p.vObjs) || p.vObjs[fanin.Id] != fanin {
				return 0
			}
			refs[fanin]++
		}
		if node.Type == IVY_LATCH {
			if node.Init < IVY_INIT_0 || node.Init > IVY_INIT_DC || node.Level != 0 || Ivy_IsComplement(node.pFanin0) != 0 {
				return 0
			}
		} else if node.Init != IVY_INIT_NONE {
			return 0
		}
		if Ivy_ObjIsHash(node) != 0 {
			key := ivyHashKey{node.pFanin0, node.pFanin1, node.Type, node.Init}
			if p.pTable[key] != node {
				return 0
			}
		}
		if Ivy_ObjIsNode(node) != 0 && node.Level != Ivy_ObjLevelNew(node) {
			return 0
		}
	}
	if counts != p.nObjs || active != Ivy_ManObjNum(p) || len(p.vPis) != int(counts[IVY_PI]) || len(p.vPos) != int(counts[IVY_PO]) || Ivy_TableCountEntries(p) != counts[IVY_AND]+counts[IVY_EXOR]+counts[IVY_LATCH] {
		return 0
	}
	for _, node := range p.vObjs {
		if node != nil && node.nRefs != refs[node] {
			return 0
		}
		if node != nil && p.fFanout != 0 && Ivy_ObjFanoutNum(p, node) != node.nRefs {
			return 0
		}
	}
	return Ivy_ManIsAcyclic(p)
}
