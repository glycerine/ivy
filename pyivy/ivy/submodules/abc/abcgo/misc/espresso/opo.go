package espresso

import "fmt"

var opo_no_make_sparse, opo_repeated, opo_exact bool

func phase_assignment(pla pPLA, strategy int32) {
	opo_no_make_sparse = strategy%2 != 0
	skip_make_sparse = opo_no_make_sparse
	opo_repeated = (strategy/2)%2 != 0
	opo_exact = (strategy/4)%2 != 0
	if opo_repeated {
		pla.phase = set_save(cube.fullset)
		repeated_phase_assignment(pla)
	} else {
		pla.phase = find_phase(pla, 0, nil)
	}
	skip_make_sparse = false
	set_phase(pla)
	minimize(pla)
}
func repeated_phase_assignment(pla pPLA) {
	for i := int32(0); i < cube.part_size[cube.output]; i++ {
		phase := find_phase(pla, i, pla.phase)
		if !is_in_set(phase, cube.first_part[cube.output]+i) {
			set_remove(pla.phase, cube.first_part[cube.output]+i)
		}
		if trace || summary {
			fmt.Printf("\nOPO loop for output #%d\nPLA->phase is %s\nphase      is %s\n", i, pc1(pla.phase), pc1(phase))
		}
	}
}
func find_phase(pla pPLA, firstOutput int32, initial pcube) pcube {
	phase := set_save(cube.fullset)
	work := &PLA_t{F: sf_save(pla.F), D: sf_save(pla.D), R: sf_save(pla.R)}
	if initial != nil {
		work.phase = set_save(initial)
		set_phase(work)
	}
	output_phase_setup(work, firstOutput)
	minimize(work)
	work.F = opo(phase, work.F, work.D, work.R, firstOutput)
	free_PLA(work)
	setdown_cube()
	cube.part_size[cube.output] -= (cube.part_size[cube.output] - firstOutput) / 2
	cube_setup()
	return phase
}
func opo(phase pcube, t, d, r pcover, firstOutput int32) pcover {
	selectSet := set_full(t.count)
	for output := int32(0); output < firstOutput; output++ {
		index := cube.first_part[cube.output] + output
		for row := int32(0); row < t.count; row++ {
			if is_in_set(GETSET(t, row), index) {
				set_remove(selectSet, row)
			}
		}
	}
	offset := (cube.part_size[cube.output] - firstOutput) / 2
	lastOutput := firstOutput + offset - 1
	if offset <= 0 {
		fatal("output phase assignment requires an undecided output")
	}
	choices := opo_recur(t, d, selectSet, offset, firstOutput, lastOutput)
	if choices.count == 0 {
		fatal("output phase assignment has no feasible cover")
	}
	selected := GETSET(choices, 0)
	result := new_cover(t.count)
	for row := int32(0); row < t.count; row++ {
		if !is_in_set(selected, row) {
			result = sf_addset(result, GETSET(t, row))
		}
	}
	missing := complement(cube1list(result))
	notCovered, temp := new_cube(), new_cube()
	for i := int32(0); i < t.count; i++ {
		p := GETSET(t, i)
		for j := int32(0); j < missing.count; j++ {
			q := GETSET(missing, j)
			if cdist0(p, q) {
				set_and(temp, p, q)
				set_or(notCovered, notCovered, temp)
			}
		}
	}
	for output := firstOutput; output <= lastOutput; output++ {
		index := cube.first_part[cube.output] + output
		if is_in_set(notCovered, index) {
			if is_in_set(notCovered, index+offset) {
				fatal("error in output phase assignment")
			}
			set_remove(phase, index)
		}
	}
	return result
}
func opo_recur(t, d pcover, selectSet pcube, offset, first, last int32) pset_family {
	return opoRecur(t, d, selectSet, offset, first, last, 1)
}
func opoRecur(t, d pcover, selectSet pcube, offset, first, last, level int32) pset_family {
	if first > last {
		fatal("invalid output phase range")
	}
	if first == last {
		return opo_leaf(t, selectSet, first, first+offset)
	}
	middle := (first + last) / 2
	left := opoRecur(t, d, selectSet, offset, first, middle, level+1)
	right := opoRecur(t, d, selectSet, offset, middle+1, last, level+1)
	// These are complements of selected prime indices. An empty intersection
	// means every prime is needed, and remains a valid phase choice.
	result := unateIntersect(left, right, level == 1, true)
	if trace {
		fmt.Printf("# OPO[%d]: %4d = %4d x %4d, time = %s\n", level-1, result.count, left.count, right.count, print_time(ptime()))
	}
	return result
}
func opo_leaf(t pcover, selectSet pset, out1, out2 int32) pset_family {
	result := sf_new(2, t.count)
	for _, output := range []int32{out1, out2} {
		p := set_save(selectSet)
		index := output + cube.first_part[cube.output]
		for row := int32(0); row < t.count; row++ {
			if is_in_set(GETSET(t, row), index) {
				set_remove(p, row)
			}
		}
		result = sf_addset(result, p)
	}
	return result
}
func output_phase_setup(pla pPLA, firstOutput int32) {
	if cube.output < 0 {
		fatal("output_phase_setup: must have an output")
	}
	if firstOutput < 0 || firstOutput >= cube.part_size[cube.output] {
		fatal("output_phase_setup: invalid first output")
	}
	f, d, r := pla.F, pla.D, pla.R
	first := cube.first_part[cube.output] + firstOutput
	last := cube.last_part[cube.output]
	offset := cube.part_size[cube.output] - firstOutput
	setdown_cube()
	cube.part_size[cube.output] += offset
	cube_setup()
	mask := set_save(cube.fullset)
	for i := first; i < cube.size; i++ {
		set_remove(mask, i)
	}
	mask1 := set_save(mask)
	for i := cube.first_part[cube.output]; i < first; i++ {
		set_remove(mask1, i)
	}
	pla.F = new_cover(f.count + r.count)
	pla.R = new_cover(f.count + r.count)
	pla.D = new_cover(d.count)
	// Old rows may span fewer words after enlarging the cube; use membership
	// copying rather than C's unchecked INLINEset_and across unequal widths.
	intersectOld := func(old, keep pcube) pcube {
		p := new_cube()
		for i := int32(0); i < int32((len(old)-1))*BPI && i < cube.size; i++ {
			if is_in_set(old, i) && is_in_set(keep, i) {
				set_insert(p, i)
			}
		}
		return p
	}
	for row := int32(0); row < f.count; row++ {
		p := GETSET(f, row)
		pf, pr := intersectOld(p, mask), intersectOld(p, mask1)
		save := false
		for i := first; i <= last; i++ {
			if is_in_set(p, i) {
				set_insert(pf, i)
				set_insert(pr, i+offset)
				save = true
			}
		}
		pla.F = sf_addset(pla.F, pf)
		if save {
			pla.R = sf_addset(pla.R, pr)
		}
	}
	for row := int32(0); row < r.count; row++ {
		p := GETSET(r, row)
		pf, pr := intersectOld(p, mask1), intersectOld(p, mask)
		save := false
		for i := first; i <= last; i++ {
			if is_in_set(p, i) {
				set_insert(pr, i)
				set_insert(pf, i+offset)
				save = true
			}
		}
		if save {
			pla.F = sf_addset(pla.F, pf)
		}
		pla.R = sf_addset(pla.R, pr)
	}
	for row := int32(0); row < d.count; row++ {
		p := GETSET(d, row)
		q := intersectOld(p, mask)
		for i := first; i <= last; i++ {
			if is_in_set(p, i) {
				set_insert(q, i)
				set_insert(q, i+offset)
			}
		}
		pla.D = sf_addset(pla.D, q)
	}
}
func set_phase(pla pPLA) pPLA {
	if cube.output < 0 || pla.phase == nil {
		fatal("set_phase requires output phase mask")
	}
	mask := cube.var_mask[cube.output]
	inverse := set_diff(new_cube(), mask, pla.phase)
	inputs := set_diff(new_cube(), cube.fullset, mask)
	set_or(inverse, inputs, inverse)
	on, off := new_cover(pla.F.count+pla.R.count), new_cover(pla.F.count+pla.R.count)
	for source, family := range []pcover{pla.F, pla.R} {
		for row := int32(0); row < family.count; row++ {
			p := GETSET(family, row)
			for which, phase := range []pcube{pla.phase, inverse} {
				q := set_and(new_cube(), p, phase)
				if setp_disjoint(q, mask) {
					continue
				}
				if source == which {
					on = sf_addset(on, q)
				} else {
					off = sf_addset(off, q)
				}
			}
		}
	}
	pla.F, pla.R = on, off
	return pla
}
func opoall(pla pPLA, first, last, strategy int32) {
	if cube.output < 0 || first < 0 || last < first || last >= cube.part_size[cube.output] || last-first >= 31 {
		fatal("invalid exhaustive phase range")
	}
	opo_exact = strategy != 0
	bestPhase := set_save(cube.fullset)
	bestF, bestD, bestR := sf_save(pla.F), sf_save(pla.D), sf_save(pla.R)
	for choice := uint32(0); choice < uint32(1)<<(last-first+1); choice++ {
		f, d, r := sf_save(pla.F), sf_save(pla.D), sf_save(pla.R)
		pla.phase = set_save(cube.fullset)
		bits := choice
		for j := last; j >= first; j-- {
			if bits&1 == 0 {
				set_remove(pla.phase, cube.first_part[cube.output]+j)
			}
			bits >>= 1
		}
		set_phase(pla)
		fmt.Printf("# phase is %s\n", pc1(pla.phase))
		summary = true
		minimize(pla)
		if pla.F.count < bestF.count {
			set_copy(bestPhase, pla.phase)
			bestF, bestD, bestR = pla.F, pla.D, pla.R
		}
		pla.F, pla.D, pla.R = f, d, r
	}
	pla.phase = bestPhase
	pla.F, pla.D, pla.R = bestF, bestD, bestR
}
func minimize(pla pPLA) {
	start := ptime()
	name := "ESPRESSO  "
	if opo_exact {
		pla.F = minimize_exact(pla.F, pla.D, pla.R, true)
		name = "EXACT"
	} else {
		pla.F = espresso(pla.F, pla.D, pla.R)
	}
	if trace || summary {
		print_trace(pla.F, name, ptime()-start)
	}
}
