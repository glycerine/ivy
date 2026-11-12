package sfm

// sfmDecomposeAreaNode extracts a window, simulates its free inputs and uses
// the SAT-backed decomposition search. The returned support contains local
// window indices; order translates them back to network IDs.
func sfmDecomposeAreaNode(p *Sfm_Ntk_t, pivot int, pars Sfm_Par_t) (truth []uint64, support []int32, order []int, ok bool) {
	return sfmDecomposeAreaNodeForced(p, pivot, pars, -1)
}

func sfmDecomposeAreaNodeForced(p *Sfm_Ntk_t, pivot int, pars Sfm_Par_t, forcedNode int) (truth []uint64, support []int32, order []int, ok bool) {
	w := sfmExtractAreaWindow(p, pivot, pars)
	if pars.WinSizeMax > 0 && len(w.order) > pars.WinSizeMax {
		return nil, nil, nil, false
	}
	window, cnfs := sfmDecWindowFromNetwork(p, w.order, w.roots, w.boundary, pivot, w.divisors)
	normal, flipped := make([]uint64, len(w.order)), make([]uint64, len(w.order))
	state := uint64(0x9e3779b97f4a7c15)
	for i, gate := range window.gates {
		if gate < 0 {
			state += 0x9e3779b97f4a7c15
			x := state
			x = (x ^ (x >> 30)) * 0xbf58476d1ce4e5b9
			x = (x ^ (x >> 27)) * 0x94d049bb133111eb
			normal[i] = x ^ (x >> 31)
			flipped[i] = normal[i]
			continue
		}
		for b := uint(0); b < 64; b++ {
			var a, af int
			for k, f := range window.fanins[i] {
				a |= int((normal[f]>>b)&1) << uint(k)
				af |= int((flipped[f]>>b)&1) << uint(k)
			}
			if bit(p.Truths[w.order[i]], a) {
				normal[i] |= uint64(1) << b
			}
			if bit(p.Truths[w.order[i]], af) {
				flipped[i] |= uint64(1) << b
			}
		}
		if i == window.pivot {
			flipped[i] = ^normal[i]
		}
	}
	var care uint64
	for _, root := range window.roots {
		care |= normal[root] ^ flipped[root]
	}
	local := make(map[int]int, len(w.order))
	for i, id := range w.order {
		local[id] = i
	}
	fallback := make([]int, len(w.mffcInputs))
	for i, id := range w.mffcInputs {
		fallback[i] = local[id]
	}
	forced := -1
	if forcedNode >= 0 {
		var found bool
		forced, found = local[forcedNode]
		if !found || forced >= w.divisors {
			return nil, nil, nil, false
		}
	}
	truth, support, ok = sfmDecomposeWindowForced(window, cnfs, w.divisors, pars, care, normal[window.pivot], normal[:w.divisors], fallback, forced)
	return truth, support, w.order, ok
}
