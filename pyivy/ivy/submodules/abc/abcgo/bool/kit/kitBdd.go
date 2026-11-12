package kit

// The Go kit uses its local Cloud ROBDD engine for the operations that used
// optional CUDD in C. Automatic memory management replaces reference counting.
func Kit_SopToBdd(dd *CloudManager, sop *Kit_Sop_t, nVars int32) *CloudNode {
	if nVars < 0 || nVars >= 16 || nVars > dd.nVars {
		panic("kit: invalid SOP variable count")
	}
	sum := dd.zero
	for _, cube := range sop.pCubes[:sop.cubeNum()] {
		product := dd.one
		for v := int32(0); v < nVars; v++ {
			value := cube >> uint(2*v) & 3
			if value == 0 {
				continue
			}
			if value == 3 {
				product = dd.zero
				break
			}
			variable := dd.vars[v]
			if value == 1 {
				variable = Cloud_Not(variable)
			}
			product = Cloud_bddAnd(dd, product, variable)
			if product == nil {
				return nil
			}
		}
		sum = Cloud_bddOr(dd, sum, product)
		if sum == nil {
			return nil
		}
	}
	return sum
}
func Kit_GraphToBdd(dd *CloudManager, g *Kit_Graph_t) *CloudNode {
	if g.fConst != 0 {
		return Cloud_NotCond(dd.one, g.eRoot.fCompl != 0)
	}
	if g.nLeaves > dd.nVars {
		panic("kit: insufficient BDD variables")
	}
	nodes := make([]*CloudNode, int32(len(g.pNodes)))
	copy(nodes, dd.vars[:g.nLeaves])
	for i := g.nLeaves; i < int32(len(g.pNodes)); i++ {
		n := g.pNodes[i]
		a := Cloud_NotCond(nodes[n.eEdge0.Node], n.eEdge0.fCompl != 0)
		b := Cloud_NotCond(nodes[n.eEdge1.Node], n.eEdge1.fCompl != 0)
		nodes[i] = Cloud_bddAnd(dd, a, b)
		if nodes[i] == nil {
			return nil
		}
		n.pFunc = nodes[i]
	}
	return Cloud_NotCond(nodes[g.eRoot.Node], g.eRoot.fCompl != 0)
}
func kitCloudIte(dd *CloudManager, v, t, e *CloudNode) *CloudNode {
	a := Cloud_bddAnd(dd, v, t)
	if a == nil {
		return nil
	}
	b := Cloud_bddAnd(dd, Cloud_Not(v), e)
	if b == nil {
		return nil
	}
	return Cloud_bddOr(dd, a, b)
}
func Kit_TruthToBdd_rec(dd *CloudManager, truth []uint32, bit, nVars, total, msbOnTop int32) *CloudNode {
	if nVars == 0 {
		if Kit_TruthHasBit(truth, bit) != 0 {
			return dd.one
		}
		return dd.zero
	}
	low := Kit_TruthToBdd_rec(dd, truth, bit, nVars-1, total, msbOnTop)
	high := Kit_TruthToBdd_rec(dd, truth, bit+(1<<uint(nVars-1)), nVars-1, total, msbOnTop)
	if low == nil || high == nil {
		return nil
	}
	variable := nVars - 1
	if msbOnTop != 0 {
		variable = total - nVars
	}
	return kitCloudIte(dd, dd.vars[variable], high, low)
}
func Kit_TruthToBdd(dd *CloudManager, truth []uint32, nVars, msbOnTop int32) *CloudNode {
	if nVars < 0 || nVars > dd.nVars {
		panic("kit: BDD variable count out of bounds")
	}
	return Kit_TruthToBdd_rec(dd, truth, 0, nVars, nVars, msbOnTop)
}
func Kit_SopFactorVerify(cover []uint32, g *Kit_Graph_t, nVars int32) int32 {
	dd := Cloud_Init(nVars, 20)
	sop := Kit_Sop_t{pCubes: cover}
	a, b := Kit_SopToBdd(dd, &sop, nVars), Kit_GraphToBdd(dd, g)
	if a != nil && a == b {
		return 1
	}
	return 0
}
