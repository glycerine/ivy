package gia

func Tas_ManPropagateOne(p *Tas_Man_t, node *Gia_Obj_t, level int32) int32 {
	if Gia_IsComplement(node) != 0 || Tas_VarIsAssigned(node) == 0 {
		panic("gia: invalid propagation variable")
	}
	if conflict := Tas_ManPropagateWatch(p, level, Tas_VarToLit(p, node)); conflict != 0 {
		return conflict
	}
	if Gia_ObjIsCi(node) != 0 {
		return 0
	}
	if Gia_ObjIsAnd(node) == 0 {
		panic("gia: propagation requires AND or input")
	}
	value0, value1 := Tas_VarFanin0Value(node), Tas_VarFanin1Value(node)
	if Tas_VarValue(node) != 0 {
		if value0 == 0 || value1 == 0 {
			if value0 == 0 && value1 != 0 {
				return Tas_ManAnalyze(p, level, node, Gia_ObjFanin0(node), nil)
			}
			if value0 != 0 && value1 == 0 {
				return Tas_ManAnalyze(p, level, node, Gia_ObjFanin1(node), nil)
			}
			return Tas_ManAnalyze(p, level, node, Gia_ObjFanin0(node), Gia_ObjFanin1(node))
		}
		if value0 == 2 {
			Tas_ManAssign(p, Gia_ObjChild0(node), level, node, nil)
		}
		if value1 == 2 {
			Tas_ManAssign(p, Gia_ObjChild1(node), level, node, nil)
		}
		return 0
	}
	if value0 == 0 || value1 == 0 {
		return 0
	}
	if value0 == 1 && value1 == 1 {
		return Tas_ManAnalyze(p, level, node, Gia_ObjFanin0(node), Gia_ObjFanin1(node))
	}
	if value0 == 1 || value1 == 1 {
		if value0 == 2 {
			Tas_ManAssign(p, Gia_Not(Gia_ObjChild0(node)), level, node, Gia_ObjFanin1(node))
		}
		if value1 == 2 {
			Tas_ManAssign(p, Gia_Not(Gia_ObjChild1(node)), level, node, Gia_ObjFanin0(node))
		}
		return 0
	}
	if Tas_VarIsJust(node) == 0 || Tas_QueHasNode(&p.pJust, node) != 0 {
		panic("gia: invalid justification queue")
	}
	Tas_QuePush(&p.pJust, node)
	return 0
}

func Tas_ManPropagateTwo(p *Tas_Man_t, node *Gia_Obj_t, level int32) int32 {
	if Gia_IsComplement(node) != 0 || Gia_ObjIsAnd(node) == 0 || Tas_VarIsAssigned(node) == 0 || Tas_VarValue(node) != 0 {
		panic("gia: invalid justification variable")
	}
	value0, value1 := Tas_VarFanin0Value(node), Tas_VarFanin1Value(node)
	if value0 == 0 || value1 == 0 {
		return 0
	}
	if value0 == 1 && value1 == 1 {
		return Tas_ManAnalyze(p, level, node, Gia_ObjFanin0(node), Gia_ObjFanin1(node))
	}
	if value0 != 1 && value1 != 1 {
		panic("gia: unresolved justification")
	}
	if value0 == 2 {
		Tas_ManAssign(p, Gia_Not(Gia_ObjChild0(node)), level, node, Gia_ObjFanin1(node))
	}
	if value1 == 2 {
		Tas_ManAssign(p, Gia_Not(Gia_ObjChild1(node)), level, node, Gia_ObjFanin0(node))
	}
	return 0
}

func Tas_ManPropagate(p *Tas_Man_t, level int32) int32 {
	for {
		for i := p.pProp.iHead; i < p.pProp.iTail && p.pProp.pData[i] != nil; i++ {
			if conflict := Tas_ManPropagateOne(p, p.pProp.pData[i], level); conflict != 0 {
				return conflict
			}
		}
		p.pProp.iHead = p.pProp.iTail
		k := p.pJust.iHead
		for i := p.pJust.iHead; i < p.pJust.iTail && p.pJust.pData[i] != nil; i++ {
			node := p.pJust.pData[i]
			if Tas_VarIsJust(node) != 0 {
				p.pJust.pData[k] = node
				k++
			} else if conflict := Tas_ManPropagateTwo(p, node, level); conflict != 0 {
				return conflict
			}
		}
		if k == p.pJust.iTail {
			break
		}
		p.pJust.iTail = k
	}
	return 0
}
