package sfm

// sfmDecomposeWindow derives a local replacement from a prepared gate window.
func sfmDecomposeWindow(window *sfmDecWindow, gateCNFs [][][]int, divisors int, pars Sfm_Par_t, care, target uint64, sim []uint64, fallback []int) ([]uint64, []int32, bool) {
	if divisors < 0 || divisors > window.pivot || len(sim) != divisors || pars.VarMax < 1 || pars.VarMax > SFM_SUPP_MAX {
		panic("sfm: invalid decomposition parameters")
	}
	solver, ok := window.prepareSolver(gateCNFs)
	if !ok {
		return nil, nil, false
	}
	s := sfmDecSearch{solver: solver, target: window.pivot, limit: pars.VarMax, conflictLimit: pars.BTLimit, useAndOr: pars.UseAndOr, fallback: fallback, forced: -1}
	s.patterns.setupSimulation(care, target, sim, pars.UseSim)
	var masks [2][SFM_SIM_WORDS]uint64
	for set := 0; set < 2; set++ {
		for i := 0; i < s.patterns.counts[set]; i++ {
			masks[set][i/64] |= uint64(1) << uint(i%64)
		}
	}
	return s.rec(nil, masks, nil, true, 0)
}
