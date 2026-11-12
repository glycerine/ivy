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
	return sfmFindAreaNodeMatchForced(p, pivot, pars, lib, -1)
}

func sfmFindAreaNodeMatchForced(p *Sfm_Ntk_t, pivot int, pars Sfm_Par_t, lib *Sfm_Lib_t, forced int) (gates []int32, fanins [][]int32, order, cone []int, ok bool) {
	session, window, ok := sfmPrepareAreaSession(p, pivot, pars)
	if !ok {
		return nil, nil, nil, nil, false
	}
	order = window.order
	forceIndex := -1
	if forced >= 0 {
		for i, id := range order[:window.divisors] {
			if id == forced {
				forceIndex = i
				break
			}
		}
		if forceIndex < 0 {
			return nil, nil, nil, nil, false
		}
	}
	bestGain, bestID := -1, int32(-1)
	var bestSupport []int32
	previous := 0
	for i := 0; i < max(1, pars.DecMax); i++ {
		if len(session.history) > previous {
			session.history = session.history[:previous]
		}
		previous = len(session.history) + 1
		truth, support, found := session.attempt(forceIndex)
		forceIndex = -1
		if !found {
			continue
		}
		cut := make([]int, len(support))
		for j, id := range support {
			cut[j] = order[id]
		}
		oldArea, candidateCone := sfmMappedMffcArea(p, pivot, cut)
		if len(support) < 2 {
			gates, fanins = Sfm_LibImplementSimple(lib, truth, support)
			return gates, fanins, order, candidateCone, true
		}
		area, id := Sfm_LibFindAreaMatch(lib, truth)
		if area < 0 || int(area) > oldArea || (!pars.ZeroCost && int(area) == oldArea) {
			continue
		}
		if gain := oldArea - int(area); gain > bestGain {
			bestGain, bestID, cone = gain, id, candidateCone
			bestSupport = append(bestSupport[:0], support...)
		}
	}
	if bestID < 0 {
		return nil, nil, nil, nil, false
	}
	gates, fanins = Sfm_LibImplementGatesArea(lib, bestSupport, bestID)
	return gates, fanins, order, cone, true
}
