package reo

import cudd "github.com/glycerine/abcgo/bdd/cudd/core"

// Units represent complemented edges with wrappers. No integer/pointer casts
// are needed; Regular-unit identity remains stable for unique-table caching.
func unitRegular(u *reo_unit) *reo_unit {
	if u != nil && u.regular != nil {
		return u.regular
	}
	return u
}
func unitComplement(u *reo_unit) bool { return u != nil && u.regular != nil }
func unitNot(u *reo_unit) *reo_unit {
	if u == nil {
		return nil
	}
	if u.regular != nil {
		return u.regular
	}
	if u.inverse == nil {
		u.inverse = &reo_unit{regular: u}
	}
	return u.inverse
}
func unitNotCond(u *reo_unit, complement bool) *reo_unit {
	if complement {
		return unitNot(u)
	}
	return u
}

func reoTransferNodesToUnits_rec(p *Manager, edge *DdNode) *reo_unit {
	if edge == nil {
		panic("reo: nil transfer root")
	}
	if p.units == nil {
		p.units = map[*DdNode]*reo_unit{}
	}
	node := cudd.Cudd_Regular(edge)
	if old := p.units[node]; old != nil {
		old.n++
		return unitNotCond(old, cudd.Cudd_IsComplement(edge))
	}
	if p.nMemChunksAlloc <= p.nMemChunks {
		p.nMemChunksAlloc = p.nMemChunks + 1
	}
	u := p.get()
	*u = reo_unit{n: 1, origin: node, add: cudd.Cudd_IsAdd(node)}
	p.units[node] = u
	if cudd.Cudd_IsConstant(node) {
		u.lev = REO_CONST_LEVEL
		u.value = node.Value
	} else {
		level := -1
		for i, v := range p.slots {
			if v == int(node.Index) {
				level = i
				break
			}
		}
		if level < 0 || level >= REO_CONST_LEVEL {
			panic("reo: transfer variable outside support")
		}
		u.lev = int16(level)
		u.pE = reoTransferNodesToUnits_rec(p, node.E)
		u.pT = reoTransferNodesToUnits_rec(p, node.T)
	}
	level := int(u.lev)
	if level == REO_CONST_LEVEL {
		level = len(p.slots)
	}
	if len(p.planes) != len(p.slots)+1 {
		p.planes = make([]reo_plane, len(p.slots)+1)
	}
	addUnitToPlane(&p.planes[level], u)
	return unitNotCond(u, cudd.Cudd_IsComplement(edge))
}

func reoTransferUnitsToNodes_rec(p *Manager, edge *reo_unit) *DdNode {
	if edge == nil {
		return nil
	}
	if p.nodes == nil {
		p.nodes = map[*reo_unit]*DdNode{}
	}
	u := unitRegular(edge)
	if old := p.nodes[u]; old != nil {
		return cudd.Cudd_NotCond(old, unitComplement(edge))
	}
	var root *DdNode
	if u.lev == REO_CONST_LEVEL {
		root = u.origin
	} else {
		level := int(u.lev)
		if level < 0 || level >= len(p.slots) {
			panic("reo: invalid transfer level")
		}
		e := reoTransferUnitsToNodes_rec(p, u.pE)
		t := reoTransferUnitsToNodes_rec(p, u.pT)
		if u.add {
			root = cudd.Cudd_addIte(p.dd, cudd.Cudd_addIthVar(p.dd, p.slots[level]), t, e)
		} else {
			root = cudd.Cudd_bddIte(p.dd, cudd.Cudd_bddIthVar(p.dd, p.slots[level]), t, e)
		}
	}
	p.nodes[u] = root
	return cudd.Cudd_NotCond(root, unitComplement(edge))
}
