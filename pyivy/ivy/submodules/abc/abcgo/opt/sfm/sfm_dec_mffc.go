package sfm

// sfmMappedMffcArea calculates the area released by replacing pivot while
// retaining the cut nodes. Reference counts are copied, so the live network is
// unchanged even if validation fails. The returned cone is in fanin-first order.
func sfmMappedMffcArea(p *Sfm_Ntk_t, pivot int, cut []int) (int, []int) {
	if pivot < p.Pis || pivot >= len(p.Fanins)-p.Pos || p.Deleted[pivot] {
		panic("sfm: invalid MFFC pivot")
	}
	refs := make([]int, len(p.Fanins))
	for i := range refs {
		refs[i] = len(p.Fanouts[i])
	}
	for _, id := range cut {
		if id < 0 || id >= len(refs) || p.Deleted[id] || id == pivot {
			panic("sfm: invalid MFFC cut")
		}
		refs[id]++
	}
	var cone []int
	var visit func(int) int
	visit = func(id int) int {
		if p.Mapping[id] == nil {
			panic("sfm: unmapped MFFC node")
		}
		area := p.Mapping[id].Area
		for _, f := range p.Fanins[id] {
			if refs[f] <= 0 {
				panic("sfm: inconsistent MFFC references")
			}
			refs[f]--
			if refs[f] == 0 && Sfm_ObjIsNode(p, f) {
				area += visit(f)
			}
		}
		cone = append(cone, id)
		return area
	}
	return visit(pivot), cone
}

// sfmFindAreaNodeMatch returns a library implementation if decomposition reduces
// the true removable cone area. Gate fanins remain local window indices.
func sfmFindAreaNodeMatch(p *Sfm_Ntk_t, pivot int, pars Sfm_Par_t, lib *Sfm_Lib_t) (gates []int32, fanins [][]int32, order, cone []int, ok bool) {
	truth, support, order, ok := sfmDecomposeAreaNode(p, pivot, pars)
	if !ok {
		return nil, nil, nil, nil, false
	}
	cut := make([]int, len(support))
	for i, id := range support {
		cut[i] = order[id]
	}
	oldArea, cone := sfmMappedMffcArea(p, pivot, cut)
	if len(support) < 2 {
		gates, fanins = Sfm_LibImplementSimple(lib, truth, support)
		return gates, fanins, order, cone, true
	}
	area, id := Sfm_LibFindAreaMatch(lib, truth)
	if area < 0 || int(area) > oldArea || (!pars.ZeroCost && int(area) == oldArea) {
		return nil, nil, nil, nil, false
	}
	gates, fanins = Sfm_LibImplementGatesArea(lib, support, id)
	return gates, fanins, order, cone, true
}
