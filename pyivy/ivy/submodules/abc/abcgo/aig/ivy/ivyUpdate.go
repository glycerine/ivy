// Incremental level propagation ported from ABC ivyUtil.c.
package ivy

func Ivy_ObjUpdateLevel_rec(p *Ivy_Man_t, n *Ivy_Obj_t) {
	if p.fFanout == 0 {
		panic("ivy: level update requires fanouts")
	}
	var fanouts []*Ivy_Obj_t
	Ivy_ObjCollectFanouts(p, n, &fanouts)
	for _, out := range fanouts {
		if Ivy_ObjIsCo(out) != 0 {
			continue
		}
		level := Ivy_ObjFanin0(out).Level
		if Ivy_ObjIsNode(out) != 0 {
			level = Ivy_ObjLevelNew(out)
		}
		if out.Level != level {
			out.Level = level
			Ivy_ObjUpdateLevel_rec(p, out)
		}
	}
}
func Ivy_ObjLevelRNew(p *Ivy_Man_t, n *Ivy_Obj_t) int32 {
	if p.fFanout == 0 || p.vRequired == nil {
		panic("ivy: reverse level requires fanouts and required levels")
	}
	var fanouts []*Ivy_Obj_t
	Ivy_ObjCollectFanouts(p, n, &fanouts)
	level := int32(1000000)
	for _, out := range fanouts {
		level = min(level, p.vRequired[out.Id])
	}
	return level - 1
}
func Ivy_ObjUpdateLevelR_rec(p *Ivy_Man_t, n *Ivy_Obj_t, required int32) {
	if Ivy_ObjIsConst1(n) != 0 || Ivy_ObjIsCi(n) != 0 {
		return
	}
	if Ivy_ObjIsNode(n) == 0 && Ivy_ObjIsBuf(n) == 0 {
		panic("ivy: invalid reverse-level gate")
	}
	for _, edge := range []*Ivy_Obj_t{n.pFanin0, n.pFanin1} {
		if edge == nil {
			continue
		}
		fanin := Ivy_Regular(edge)
		if p.vRequired[fanin.Id] > required-1 {
			p.vRequired[fanin.Id] = required - 1
			Ivy_ObjUpdateLevelR_rec(p, fanin, required-1)
		}
	}
}

func Ivy_ManResetLevels_rec(n *Ivy_Obj_t) int32 {
	if n.Level != 0 || Ivy_ObjIsCi(n) != 0 || Ivy_ObjIsConst1(n) != 0 {
		return n.Level
	}
	if Ivy_ObjIsBuf(n) != 0 {
		n.Level = Ivy_ManResetLevels_rec(Ivy_ObjFanin0(n))
		return n.Level
	}
	if Ivy_ObjIsNode(n) == 0 {
		panic("ivy: invalid reset-level gate")
	}
	Ivy_ManResetLevels_rec(Ivy_ObjFanin0(n))
	Ivy_ManResetLevels_rec(Ivy_ObjFanin1(n))
	n.Level = Ivy_ObjLevelNew(n)
	return n.Level
}
func Ivy_ManResetLevels(p *Ivy_Man_t) {
	if Ivy_ManIsAcyclic(p) == 0 {
		panic("ivy: cannot reset cyclic levels")
	}
	for _, n := range p.vObjs {
		if n != nil {
			n.Level = 0
		}
	}
	for _, n := range p.vObjs {
		if n != nil && Ivy_ObjIsCo(n) != 0 {
			Ivy_ManResetLevels_rec(Ivy_ObjFanin0(n))
		}
	}
}
