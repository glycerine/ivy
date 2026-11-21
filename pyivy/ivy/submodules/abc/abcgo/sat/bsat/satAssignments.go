package bsat

const satUnassigned = 3

// satAssignments owns the trail and variable metadata shared by CDCL enqueue,
// decisions and backtracking. A value is a literal sign, or 3 when unassigned.
type satAssignments struct {
	qhead, qtail             int32
	levels, reasons          []int32
	assigns, polarity, tags  []uint8
	trail                    []lit
	trail_lim, tagged, stack veci
}

func newSatAssignments(size int32) satAssignments {
	if size < 0 {
		panic("negative SAT variable count")
	}
	a := satAssignments{
		levels:   make([]int32, size),
		reasons:  make([]int32, size),
		assigns:  make([]uint8, size),
		polarity: make([]uint8, size),
		tags:     make([]uint8, size),
		trail:    make([]lit, size),
	}
	for i := range a.assigns {
		a.assigns[i] = satUnassigned
	}
	return a
}
func (a *satAssignments) enqueue(l lit, reason int32) bool {
	if lit_check(l, int32(len(a.assigns))) == 0 {
		panic("literal outside solver variables")
	}
	v := lit_var(l)
	sign := uint8(lit_sign(l))
	if a.assigns[v] != satUnassigned {
		return a.assigns[v] == sign
	}
	if a.qtail >= int32(len(a.trail)) {
		panic("SAT trail overflow")
	}
	a.assigns[v] = sign
	a.levels[v] = a.trail_lim.size
	a.reasons[v] = reason
	a.trail[a.qtail] = l
	a.qtail++
	return true
}
func (a *satAssignments) decision(l lit) bool {
	if a.qhead != a.qtail {
		panic("decision before propagation completes")
	}
	if lit_check(l, int32(len(a.assigns))) == 0 || a.assigns[lit_var(l)] != satUnassigned {
		panic("decision on assigned variable")
	}
	veci_push(&a.trail_lim, a.qtail)
	return a.enqueue(l, 0)
}

// cancelUntil returns the former trail of propagated literals. The solver
// consumes it in reverse before the next enqueue can overwrite it.
func (a *satAssignments) cancelUntil(level int32) []lit {
	if level < 0 {
		panic("negative decision level")
	}
	if a.trail_lim.size <= level {
		return nil
	}
	bound := a.trail_lim.ptr[level]
	lastLevel := a.trail_lim.ptr[a.trail_lim.size-1]
	for i := a.qtail - 1; i >= bound; i-- {
		v := lit_var(a.trail[i])
		a.assigns[v] = satUnassigned
		a.reasons[v] = 0
		if i < lastLevel {
			a.polarity[v] = uint8(1 - lit_sign(a.trail[i]))
		}
	}
	// A failed assumption can have a trail boundary beyond pending root units.
	restore := a.trail[bound:max(bound, a.qhead)]
	a.qhead = bound
	a.qtail = bound
	veci_resize(&a.trail_lim, level)
	return restore
}
func (a *satAssignments) rollback(bound int32) []lit {
	if a.trail_lim.size != 0 || bound < 0 || bound > a.qhead || a.qhead > a.qtail {
		panic("invalid top-level rollback")
	}
	restore := a.trail[bound:a.qhead]
	for i := a.qtail - 1; i >= bound; i-- {
		v := lit_var(a.trail[i])
		a.assigns[v] = satUnassigned
		a.reasons[v] = 0
	}
	a.qhead = bound
	a.qtail = bound
	return restore
}
