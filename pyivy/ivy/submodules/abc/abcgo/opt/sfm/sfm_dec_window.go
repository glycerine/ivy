package sfm

// sfmDecomposeWindow derives a local replacement from a prepared gate window.
func sfmDecomposeWindow(window *sfmDecWindow, gateCNFs [][][]int, divisors int, pars Sfm_Par_t, care, target uint64, sim []uint64, fallback []int) ([]uint64, []int32, bool) {
	return sfmDecomposeWindowForced(window, gateCNFs, divisors, pars, care, target, sim, fallback, -1)
}

func sfmDecomposeWindowForced(window *sfmDecWindow, gateCNFs [][][]int, divisors int, pars Sfm_Par_t, care, target uint64, sim []uint64, fallback []int, forced int) ([]uint64, []int32, bool) {
	if divisors < 0 || divisors > window.pivot || len(sim) != divisors || pars.VarMax < 1 || pars.VarMax > SFM_SUPP_MAX {
		panic("sfm: invalid decomposition parameters")
	}
	if forced < -1 || forced >= divisors {
		panic("sfm: invalid forced divisor")
	}
	solver, ok := window.prepareSolver(gateCNFs)
	if !ok {
		return nil, nil, false
	}
	s := sfmDecSearch{solver: solver, target: window.pivot, limit: pars.VarMax, conflictLimit: pars.BTLimit, useAndOr: pars.UseAndOr, fallback: fallback, forced: forced}
	s.patterns.setupSimulation(care, target, sim, pars.UseSim)
	var masks [2][SFM_SIM_WORDS]uint64
	for set := 0; set < 2; set++ {
		for i := 0; i < s.patterns.counts[set]; i++ {
			masks[set][i/64] |= uint64(1) << uint(i%64)
		}
	}
	return s.rec(nil, masks, nil, true, 0)
}

// sfmMatchAreaReplacement finds the prepared gate implementation for a window.
func sfmMatchAreaReplacement(window *sfmDecWindow, cnfs [][][]int, divisors int, pars Sfm_Par_t, library *Sfm_Lib_t, care, target uint64, sim []uint64, fallback []int, oldArea int32) ([]int32, [][]int32, bool) {
	truth, support, ok := sfmDecomposeWindow(window, cnfs, divisors, pars, care, target, sim, fallback)
	if !ok {
		return nil, nil, false
	}
	if len(support) < 2 {
		gates, fanins := Sfm_LibImplementSimple(library, truth, support)
		return gates, fanins, true
	}
	area, id := Sfm_LibFindAreaMatch(library, truth)
	if area < 0 || area > oldArea || (!pars.ZeroCost && area == oldArea) {
		return nil, nil, false
	}
	gates, fanins := Sfm_LibImplementGatesArea(library, support, id)
	return gates, fanins, true
}

// sfmDecWindowFromNetwork converts a caller-selected topological window to SAT
// object IDs and returns the corresponding original node IDs.
func sfmDecWindowFromNetwork(p *Sfm_Ntk_t, order, roots, boundary []int, pivot, divisors int) (*sfmDecWindow, [][][]int) {
	if divisors < 0 || divisors > len(order) {
		panic("sfm: invalid divisor count")
	}
	local := make(map[int]int, len(order))
	for i, id := range order {
		if id < 0 || id >= len(p.Fanins) || p.Deleted[id] {
			panic("sfm: invalid window node")
		}
		if _, exists := local[id]; exists {
			panic("sfm: duplicate window node")
		}
		local[id] = i
	}
	pivotID, ok := local[pivot]
	if !ok || pivotID < divisors {
		panic("sfm: invalid window pivot")
	}
	cut := make(map[int]bool, len(boundary))
	for _, id := range boundary {
		index, exists := local[id]
		if !exists || index >= pivotID {
			panic("sfm: invalid window boundary")
		}
		cut[id] = true
	}
	window := &sfmDecWindow{gates: make([]int, len(order)), fanins: make([][]int, len(order)), pivot: pivotID}
	var cnfs [][][]int
	for i, id := range order {
		if Sfm_ObjIsPi(p, id) || cut[id] {
			window.gates[i] = -1
			continue
		}
		for _, fan := range p.Fanins[id] {
			index, exists := local[fan]
			if !exists || index >= i {
				panic("sfm: window is not topological")
			}
			window.fanins[i] = append(window.fanins[i], index)
		}
		window.gates[i] = len(cnfs)
		cnfs = append(cnfs, Sfm_TruthToCnf(p.Truths[id][0], p.Truths[id], len(p.Fanins[id])))
	}
	for _, id := range roots {
		index, exists := local[id]
		if !exists || index < pivotID {
			panic("sfm: invalid window root")
		}
		window.roots = append(window.roots, index)
	}
	return window, cnfs
}
