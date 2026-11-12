package sfm

// sfmAreaWindow records the node order and cut boundaries used by the area
// decomposition solver. IDs refer to the original network.
type sfmAreaWindow struct {
	order, roots, boundary, mffc, mffcInputs []int
	divisors                                 int
}

// sfmExtractAreaWindow follows Sfm_DecExtract's area-oriented traversal.
// The timing-oriented ordering is handled separately by the timing manager.
func sfmExtractAreaWindow(p *Sfm_Ntk_t, pivot int, pars Sfm_Par_t) sfmAreaWindow {
	if pivot < p.Pis || pivot >= len(p.Fanins)-p.Pos || p.Deleted[pivot] || pars.MffcMax <= 0 {
		panic("sfm: invalid area window pivot or MFFC limit")
	}
	const (
		pi     = 1
		input  = 2
		fanin  = 4
		cone   = 8
		target = 16
	)
	n := len(p.Fanins)
	upper, lower := p.Levels[pivot]+pars.TfoLevMax, p.Levels[pivot]-pars.TfiLevMax
	seen := make([]bool, n)
	var tfo []int
	var reverse func(int)
	reverse = func(id int) {
		if seen[id] {
			return
		}
		seen[id] = true
		if Sfm_ObjIsPo(p, id) || p.Levels[id] > upper {
			return
		}
		if len(p.Fanouts[id]) <= pars.FanoutMax {
			inside := true
			for _, out := range p.Fanouts[id] {
				if Sfm_ObjIsPo(p, out) || p.Levels[out] > upper {
					inside = false
					break
				}
			}
			if inside {
				for _, out := range p.Fanouts[id] {
					reverse(out)
				}
			}
		}
		tfo = append(tfo, id)
	}
	reverse(pivot)
	fanoutsIn := make([]int, n)
	for _, id := range tfo {
		for _, f := range p.Fanins[id] {
			fanoutsIn[f]++
		}
	}
	var w sfmAreaWindow
	for _, id := range tfo {
		if fanoutsIn[id] != len(p.Fanouts[id]) {
			w.roots = append(w.roots, id)
		}
	}
	if len(w.roots) == 0 {
		panic("sfm: window has no observable root")
	}
	seen = make([]bool, n)
	labels := make([]int, n)
	var tfi []int
	var visit func(int, int) int
	visit = func(id, label int) int {
		if seen[id] {
			return labels[id]
		}
		seen[id] = true
		if Sfm_ObjIsPi(p, id) || (p.Levels[id] < lower && len(p.Fanins[id]) > 0) {
			labels[id] = label
			tfi = append(tfi, id)
			return label
		}
		if !Sfm_ObjIsNode(p, id) {
			panic("sfm: unexpected window output")
		}
		if len(p.Fanins[id]) == 0 {
			labels[id] = label
		}
		for _, f := range p.Fanins[id] {
			labels[id] |= visit(f, label)
		}
		tfi = append(tfi, id)
		return labels[id]
	}
	visit(pivot, pi)
	firstSize := len(tfi)
	w.mffc = append(w.mffc, pivot)
	eligible := func(id int) bool {
		return Sfm_ObjIsNode(p, id) && len(p.Fanouts[id]) == 1 && seen[id] && (p.Levels[id] >= lower || len(p.Fanins[id]) == 0)
	}
	add := func(id int) {
		for _, old := range w.mffc {
			if old == id {
				return
			}
		}
		w.mffc = append(w.mffc, id)
	}
	// The C traversal examines three successive fanin depths, in breadth order.
	front := []int{pivot}
	for depth := 0; depth < 3; depth++ {
		var next []int
		for _, id := range front {
			for _, f := range p.Fanins[id] {
				if eligible(f) && len(w.mffc) < pars.MffcMax {
					add(f)
					next = append(next, f)
				}
			}
		}
		front = next
	}
	for _, id := range w.mffc {
		labels[id] |= cone
	}
	labels[pivot] |= target
	for _, id := range w.mffc {
		for _, f := range p.Fanins[id] {
			if seen[f] && labels[f] == pi {
				found := false
				for _, x := range w.mffcInputs {
					if x == f {
						found = true
						break
					}
				}
				if !found {
					w.mffcInputs = append(w.mffcInputs, f)
				}
			}
		}
	}
	for _, id := range w.roots {
		visit(id, input)
	}
	for _, id := range tfi[firstSize:] {
		if labels[id] != input {
			for _, f := range p.Fanins[id] {
				if labels[f] == input {
					labels[f] = fanin
				}
			}
		}
	}
	appendNode := func(id int, skip bool) {
		w.order = append(w.order, id)
		if skip {
			w.boundary = append(w.boundary, id)
		}
	}
	for _, id := range tfi {
		if labels[id] == pi {
			appendNode(id, Sfm_ObjIsPi(p, id) || (p.Levels[id] < lower && len(p.Fanins[id]) > 0))
		}
	}
	w.divisors = len(w.order)
	for _, id := range tfi {
		if labels[id] == pi|input || labels[id] == fanin {
			appendNode(id, labels[id] == fanin)
		}
	}
	for _, id := range tfi {
		if labels[id] >= cone {
			appendNode(id, false)
		}
	}
	return w
}
