// Gasp heuristics: generate primes covering maximally reduced cubes, then
// choose an irredundant subset together with the original cover.
package espresso

import "fmt"

func reduce_gasp(f, d pcover) pcover {
	reduced := sf_new(f.count, cube.size)
	fd := cube2list(f, d)
	for row := int32(0); row < f.count; row++ {
		original := GETSET(f, row)
		under := reduce_cube(fd, original)
		if setp_empty(under) {
			fatal("empty reduction in reduce_gasp, shouldn't happen")
		}
		if setp_equal(under, original) {
			SET(under, PRIME)
			// The source copies the original row here, including its PRIME flag.
			reduced = sf_addset(reduced, original)
		} else {
			RESET(under, PRIME)
			reduced = sf_addset(reduced, under)
		}
		if debug&GASP != 0 {
			fmt.Printf("REDUCE_GASP: %s reduced to %s\n", pc1(original), pc2(under))
		}
		free_cube(under)
	}
	free_cubelist(fd)
	return reduced
}

func expand_gasp(f, d, r, original pcover) pcover {
	candidates := sf_new(10, cube.size)
	for row := int32(0); row < f.count; row++ {
		expand1_gasp(f, d, r, original, row, &candidates)
	}
	return expand(sf_dupl(candidates), r, false)
}

func expand1_gasp(f, d, r, original pcover, row int32, candidates *pcover) {
	if debug&EXPAND1 != 0 {
		fmt.Printf("\nEXPAND1_GASP:    \t%s\n", pc1(GETSET(f, row)))
	}
	raise, free, temp := new_cube(), new_cube(), new_cube()
	r.active_count = r.count
	for i := int32(0); i < r.count; i++ {
		SET(GETSET(r, i), ACTIVE)
	}
	f.active_count = f.count
	for i := int32(0); i < f.count; i++ {
		current := GETSET(f, i)
		if i == row || TESTP(current, PRIME) {
			f.active_count--
			RESET(current, ACTIVE)
		} else {
			SET(current, ACTIVE)
		}
	}
	set_copy(raise, GETSET(f, row))
	set_diff(free, cube.fullset, raise)
	essen_parts(r, f, raise, free)
	essen_raising(r, raise, free)
	for other := int32(0); other < f.count; other++ {
		under := GETSET(f, other)
		if !TESTP(under, ACTIVE) {
			continue
		}
		if !setp_implies(under, raise) && !feasibly_covered(r, under, raise, temp) {
			continue
		}
		// Test against original cubes with only this expansion's seed reduced.
		replaced := sf_save(original)
		set_copy(GETSET(replaced, row), GETSET(f, row))
		fd := cube2list(replaced, d)
		essential := reduce_cube(fd, GETSET(replaced, other))
		free_cubelist(fd)
		sf_free(replaced)
		if feasibly_covered(r, essential, raise, temp) {
			set_or(temp, raise, essential)
			RESET(temp, PRIME)
			*candidates = sf_addset(*candidates, temp)
		}
		free_cube(essential)
	}
	free_cube(raise)
	free_cube(free)
	free_cube(temp)
}

func irred_gasp(f, d, candidates pcover) pcover {
	if candidates.count != 0 {
		return irredundant(sf_append(f, candidates), d)
	}
	free_cover(candidates)
	return f
}

func last_gasp(f, d, r pcover, cost pcost) pcover {
	start := ptime()
	reduced := reduce_gasp(f, d)
	totals(start, GREDUCE_TIME, reduced, cost)
	start = ptime()
	candidates := expand_gasp(reduced, d, r, f)
	totals(start, GEXPAND_TIME, candidates, cost)
	free_cover(reduced)
	start = ptime()
	f = irred_gasp(f, d, candidates)
	totals(start, GIRRED_TIME, f, cost)
	return f
}

func super_gasp(f, d, r pcover, cost pcost) pcover {
	start := ptime()
	reduced := reduce_gasp(f, d)
	totals(start, GREDUCE_TIME, reduced, cost)
	start = ptime()
	primes := all_primes(reduced, r)
	totals(start, GEXPAND_TIME, primes, cost)
	free_cover(reduced)
	start = ptime()
	combined := sf_dupl(sf_append(f, primes))
	if trace {
		print_trace(combined, "NEWPRIMES", ptime()-start)
	}
	start = ptime()
	result := irredundant(combined, d)
	totals(start, IRRED_TIME, result, cost)
	return result
}
