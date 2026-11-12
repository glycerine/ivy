package espresso

// Espresso-II expansion: raise each cube while keeping it orthogonal to the
// OFF-set, preferring expansions that absorb other ON-set cubes. The shared
// cube context and flags retain the original traversal and heuristic order.
import "fmt"

func expand(on, off pcover, nonsparse bool) pcover {
	if use_random_order {
		on = random_order(on)
	} else {
		on = mini_sort(on, ascend)
	}
	raise, free, initial, super, over := new_cube(), new_cube(), new_cube(), new_cube(), new_cube()
	if nonsparse {
		for variable := int32(0); variable < cube.num_vars; variable++ {
			if cube.sparse[variable] != 0 {
				set_or(initial, initial, cube.var_mask[variable])
			}
		}
	}
	for i := int32(0); i < on.count; i++ {
		row := GETSET(on, i)
		RESET(row, COVERED)
		RESET(row, NONESSEN)
	}
	for i := int32(0); i < on.count; i++ {
		row := GETSET(on, i)
		if TESTP(row, PRIME) || TESTP(row, COVERED) {
			continue
		}
		var covered int32
		expand1(off, on, raise, free, over, super, initial, &covered, row)
		if debug&EXPAND != 0 {
			fmt.Printf("EXPAND: %s (covered %d)\n", pc1(row), covered)
		}
		set_copy(row, raise)
		SET(row, PRIME)
		RESET(row, COVERED)
		if covered == 0 && !setp_equal(row, over) {
			SET(row, NONESSEN)
		}
	}
	on.active_count = 0
	changed := false
	for i := int32(0); i < on.count; i++ {
		row := GETSET(on, i)
		if TESTP(row, COVERED) {
			RESET(row, ACTIVE)
			changed = true
		} else {
			SET(row, ACTIVE)
			on.active_count++
		}
	}
	if changed {
		on = sf_inactive(on)
	}
	return on
}
func expand1(blocking, covering pcover, raise, free, over, super, initial pcube, covered *int32, c pcube) {
	if debug&EXPAND1 != 0 {
		fmt.Printf("\nEXPAND1:    \t%s\n", pc1(c))
	}
	SET(c, PRIME)
	setup_BB_CC(blocking, covering)
	*covered = 0
	set_copy(super, c)
	set_copy(raise, c)
	set_diff(free, cube.fullset, raise)
	if !setp_empty(initial) {
		set_diff(free, free, initial)
		elim_lowering(blocking, covering, raise, free)
	}
	essen_parts(blocking, covering, raise, free)
	set_or(over, raise, free)
	if covering != nil && covering.active_count > 0 {
		select_feasible(blocking, covering, raise, free, super, covered)
	}
	for covering != nil && covering.active_count > 0 {
		part := most_frequent(covering, free)
		set_insert(raise, part)
		set_remove(free, part)
		essen_parts(blocking, covering, raise, free)
	}
	for blocking.active_count > 0 {
		mincov(blocking, raise, free)
	}
	set_or(raise, raise, free)
}
func essen_parts(blocking, covering pcover, raise, free pcube) {
	lower := set_copy(cube.temp[0], cube.emptyset)
	for i := int32(0); i < blocking.count; i++ {
		row := GETSET(blocking, i)
		if !TESTP(row, ACTIVE) {
			continue
		}
		distance := cdist01(row, raise)
		if distance > 1 {
			continue
		}
		if distance == 0 {
			panic("espresso: ON-set and OFF-set are not orthogonal")
		}
		force_lower(lower, row, raise)
		blocking.active_count--
		RESET(row, ACTIVE)
	}
	if !setp_empty(lower) {
		set_diff(free, free, lower)
		elim_lowering(blocking, covering, raise, free)
	}
	if debug&EXPAND1 != 0 {
		fmt.Printf("ESSEN_PARTS:\tRAISE=%s FREESET=%s\n", pc1(raise), pc2(free))
	}
}
func essen_raising(blocking pcover, raise, free pcube) {
	raised := set_copy(cube.temp[0], cube.emptyset)
	for i := int32(0); i < blocking.count; i++ {
		row := GETSET(blocking, i)
		if TESTP(row, ACTIVE) {
			set_or(raised, raised, row)
		}
	}
	set_diff(raised, free, raised)
	set_or(raise, raise, raised)
	set_diff(free, free, raised)
	if debug&EXPAND1 != 0 {
		fmt.Printf("ESSEN_RAISING:\tRAISE=%s FREESET=%s\n", pc1(raise), pc2(free))
	}
}
func elim_lowering(blocking, covering pcover, raise, free pcube) {
	over := set_or(cube.temp[0], raise, free)
	for i := int32(0); i < blocking.count; i++ {
		row := GETSET(blocking, i)
		if TESTP(row, ACTIVE) && !cdist0(row, over) {
			blocking.active_count--
			RESET(row, ACTIVE)
		}
	}
	if covering != nil {
		for i := int32(0); i < covering.count; i++ {
			row := GETSET(covering, i)
			if TESTP(row, ACTIVE) && !setp_implies(row, over) {
				covering.active_count--
				RESET(row, ACTIVE)
			}
		}
	}
}
func most_frequent(covering pcover, free pcube) int32 {
	counts := make([]int32, cube.size)
	if covering != nil {
		for i := int32(0); i < covering.count; i++ {
			row := GETSET(covering, i)
			if TESTP(row, ACTIVE) {
				set_adjcnt(row, counts, 1)
			}
		}
	}
	best, count := int32(-1), int32(-1)
	for part := int32(0); part < cube.size; part++ {
		if is_in_set(free, part) && counts[part] > count {
			best, count = part, counts[part]
		}
	}
	if debug&EXPAND1 != 0 {
		fmt.Printf("MOST_FREQUENT:\tbest=%d FREESET=%s\n", best, pc2(free))
	}
	return best
}
func setup_BB_CC(blocking, covering pcover) {
	blocking.active_count = blocking.count
	for i := int32(0); i < blocking.count; i++ {
		SET(GETSET(blocking, i), ACTIVE)
	}
	if covering != nil {
		covering.active_count = covering.count
		for i := int32(0); i < covering.count; i++ {
			row := GETSET(covering, i)
			if TESTP(row, COVERED) || TESTP(row, PRIME) {
				covering.active_count--
				RESET(row, ACTIVE)
			} else {
				SET(row, ACTIVE)
			}
		}
	}
}
func select_feasible(blocking, covering pcover, raise, free, super pcube, covered *int32) {
	feasible := make([]pcube, 0, covering.active_count)
	for i := int32(0); i < covering.count; i++ {
		row := GETSET(covering, i)
		if TESTP(row, ACTIVE) {
			feasible = append(feasible, row)
		}
	}
	forced := new_cover(int32(len(feasible)))
	lowers := make([]pcube, len(feasible))
	for i := range lowers {
		lowers[i] = GETSET(forced, int32(i))
	}
	for {
		essen_raising(blocking, raise, free)
		previous := len(feasible)
		count := 0
		for i := 0; i < previous; i++ {
			row := feasible[i]
			if !TESTP(row, ACTIVE) {
				continue
			}
			if setp_implies(row, raise) {
				(*covered)++
				set_or(super, super, row)
				covering.active_count--
				RESET(row, ACTIVE)
				SET(row, COVERED)
			} else if feasibly_covered(blocking, row, raise, lowers[count]) {
				feasible[count] = row
				count++
			}
		}
		feasible = feasible[:count]
		if debug&EXPAND1 != 0 {
			fmt.Printf("SELECT_FEASIBLE: started with %d pfcc, ended with %d fcc\n", previous, count)
		}
		if count == 0 {
			sf_free(forced)
			return
		}
		var best pcube
		bestCount, bestSize := -1, int32(0)
		for i, row := range feasible {
			size := set_dist(row, free)
			count := 0
			for _, other := range feasible {
				if setp_disjoint(lowers[i], other) {
					count++
				}
			}
			if count > bestCount || count == bestCount && size < bestSize {
				best, bestCount, bestSize = row, count, size
			}
		}
		set_or(raise, raise, best)
		set_diff(free, free, raise)
		if debug&EXPAND1 != 0 {
			fmt.Printf("FEASIBLE:  \tRAISE=%s FREESET=%s\n", pc1(raise), pc2(free))
		}
		essen_parts(blocking, covering, raise, free)
	}
}
func feasibly_covered(blocking pcover, c, raise, lower pcube) bool {
	candidate := set_or(cube.temp[0], raise, c)
	set_copy(lower, cube.emptyset)
	for i := int32(0); i < blocking.count; i++ {
		row := GETSET(blocking, i)
		if !TESTP(row, ACTIVE) {
			continue
		}
		distance := cdist01(row, candidate)
		if distance > 1 {
			continue
		}
		if distance == 0 {
			return false
		}
		force_lower(lower, row, candidate)
	}
	return true
}
func mincov(blocking pcover, raise, free pcube) {
	required := new_cover(blocking.active_count)
	for i := int32(0); i < blocking.count; i++ {
		row := GETSET(blocking, i)
		if TESTP(row, ACTIVE) {
			lower := set_copy(GETSET(required, required.count), cube.emptyset)
			required.count++
			force_lower(lower, row, raise)
		}
	}
	total := int32(0)
	fallback := false
	for i := int32(0); i < required.count; i++ {
		row := GETSET(required, i)
		expansion := int32(1)
		for variable := cube.num_binary_vars; variable < cube.num_vars; variable++ {
			distance := set_dist(row, cube.var_mask[variable])
			if distance > 1 {
				if expansion > 500/distance {
					fallback = true
					break
				}
				expansion *= distance
			}
		}
		total += expansion
		if fallback || total > 500 {
			fallback = true
			break
		}
	}
	if fallback {
		sf_free(required)
		part := most_frequent(nil, free)
		set_insert(raise, part)
		set_diff(free, free, raise)
		essen_parts(blocking, nil, raise, free)
		return
	}
	required = unravel(required, cube.num_binary_vars)
	lower := do_sm_minimum_cover(required)
	raised := set_diff(cube.temp[0], free, lower)
	set_or(raise, raise, raised)
	set_copy(free, cube.emptyset)
	blocking.active_count = 0
	sf_free(required)
	set_free(lower)
	if debug&EXPAND1 != 0 {
		fmt.Printf("MINCOV:    \tRAISE=%s FREESET=%s\n", pc1(raise), pc2(free))
	}
}
func find_all_primes(blocking pcover, raise, free pcube) pcover {
	if blocking.active_count == 0 {
		result := new_cover(1)
		row := GETSET(result, 0)
		result.count = 1
		set_or(row, raise, free)
		SET(row, PRIME)
		return result
	}
	required := new_cover(blocking.active_count)
	for i := int32(0); i < blocking.count; i++ {
		row := GETSET(blocking, i)
		if TESTP(row, ACTIVE) {
			lower := set_copy(GETSET(required, required.count), cube.emptyset)
			required.count++
			force_lower(lower, row, raise)
		}
	}
	required = sf_rev_contain(unravel(required, cube.num_binary_vars))
	result := exact_minimum_cover(required)
	for i := int32(0); i < result.count; i++ {
		row := GETSET(result, i)
		set_diff(row, free, row)
		set_or(row, row, raise)
		SET(row, PRIME)
	}
	sf_free(required)
	return result
}
func all_primes(on, off pcover) pcover {
	free, raise := new_cube(), new_cube()
	result := new_cover(on.count)
	for i := int32(0); i < on.count; i++ {
		row := GETSET(on, i)
		if TESTP(row, PRIME) {
			result = sf_addset(result, row)
		} else {
			set_copy(raise, row)
			set_diff(free, cube.fullset, raise)
			setup_BB_CC(off, nil)
			essen_parts(off, nil, raise, free)
			result = sf_append(result, find_all_primes(off, raise, free))
		}
	}
	return result
}
func Expand(on, off *Family, nonsparse bool) *Family { return expand(on, off, nonsparse) }
