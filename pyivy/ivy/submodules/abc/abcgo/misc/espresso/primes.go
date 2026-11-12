package espresso

// primes_consensus uses recursive Shannon cofactors and consensus products to
// enumerate all prime implicants, with containment after each bounded batch.
func primes_consensus(t *cubeList) pcover {
	if result, done := primesConsensusSpecialCases(t); done {
		return result
	}
	left, right := new_cube(), new_cube()
	variable := binate_split_select(t, left, right, COMPL)
	l := primes_consensus(scofactor(t, left, variable))
	r := primes_consensus(scofactor(t, right, variable))
	result := primes_consensus_merge(l, r, left, right)
	free_cubelist(t)
	return result
}
func primesConsensusSpecialCases(t *cubeList) (pcover, bool) {
	if len(t.cubes) == 0 {
		free_cubelist(t)
		return new_cover(0), true
	}
	if len(t.cubes) == 1 {
		p := set_or(new_cube(), t.cofactor, t.cubes[0])
		free_cubelist(t)
		return sf_addset(new_cover(1), p), true
	}
	for _, p := range t.cubes {
		if full_row(p, t.cofactor) {
			free_cubelist(t)
			return sf_addset(new_cover(1), cube.fullset), true
		}
	}
	ceiling := set_save(t.cofactor)
	for _, p := range t.cubes {
		set_or(ceiling, ceiling, p)
	}
	if !setp_equal(ceiling, cube.fullset) {
		missing := set_diff(new_cube(), cube.fullset, ceiling)
		set_or(t.cofactor, t.cofactor, missing)
		result := primes_consensus(t)
		for i := int32(0); i < result.count; i++ {
			p := GETSET(result, i)
			set_and(p, p, ceiling)
		}
		return result, true
	}
	massive_count(t)
	if cdata.vars_active == 1 {
		free_cubelist(t)
		return sf_addset(new_cover(1), cube.fullset), true
	}
	if cdata.vars_unate == cdata.vars_active {
		result := sf_contain(cubeunlist(t))
		free_cubelist(t)
		return result, true
	}
	return nil, false
}
func primes_consensus_merge(left, right pcover, leftCofactor, rightCofactor pcube) pcover {
	left = and_with_cofactor(left, leftCofactor)
	right = and_with_cofactor(right, rightCofactor)
	batch := new_cover(500)
	result := sf_contain(sf_join(left, right))
	for i := int32(0); i < left.count; i++ {
		a := GETSET(left, i)
		for j := int32(0); j < right.count; j++ {
			b := GETSET(right, j)
			if cdist01(a, b) == 1 {
				p := new_cube()
				consensus(p, a, b)
				batch = sf_addset(batch, p)
				if batch.count >= 500 {
					result = sf_union(result, sf_contain(batch))
					batch = new_cover(500)
				}
			}
		}
	}
	return sf_union(result, sf_contain(batch))
}
func and_with_cofactor(a pset_family, cofactor pset) pcover {
	a.active_count = 0
	for i := int32(0); i < a.count; i++ {
		p := GETSET(a, i)
		set_and(p, p, cofactor)
		if cdist(p, cube.fullset) > 0 {
			RESET(p, ACTIVE)
		} else {
			SET(p, ACTIVE)
			a.active_count++
		}
	}
	return sf_inactive(a)
}
