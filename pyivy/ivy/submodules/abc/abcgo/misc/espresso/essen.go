package espresso

// essential moves essential primes from the onset to the don't-care cover.
// RELESSEN/NONESSEN are supplied by irredundancy/expansion respectively.
func essential(fp, dp *pcover) pcover {
	f, d := *fp, *dp
	sf_active(f)
	e := new_cover(10)
	for i := int32(0); i < f.count; i++ {
		p := GETSET(f, i)
		if !TESTP(p, NONESSEN) && TESTP(p, RELESSEN) && essen_cube(f, d, p) {
			sf_addset(e, p)
			RESET(p, ACTIVE)
			f.active_count--
		}
	}
	*fp = sf_inactive(f)
	*dp = sf_join(d, e)
	sf_free(d)
	return e
}
func essen_cube(f, d pcover, c pcube) bool {
	fd := sf_join(f, d)
	h := cb_consensus(fd, c)
	free_cover(fd)
	list := cube2list(h, d)
	result := !cube_is_covered(list, c)
	free_cubelist(list)
	free_cover(h)
	return result
}
func cb_consensus(t pcover, c pcube) pcover {
	result := new_cover(t.count * 2)
	temp := new_cube()
	for i := int32(0); i < t.count; i++ {
		p := GETSET(t, i)
		if sameCube(p, c) {
			continue
		}
		switch cdist01(p, c) {
		case 0:
			result = cb_consensus_dist0(result, p, c)
		case 1:
			consensus(temp, p, c)
			sf_addset(result, temp)
		}
	}
	return result
}
func cb_consensus_dist0(result pcover, p, c pcube) pcover {
	if setp_implies(p, c) {
		return result
	}
	diff, intersection := set_diff(new_cube(), p, c), set_and(new_cube(), p, c)
	gotOne := false
	for variable := cube.num_binary_vars; variable < cube.num_vars; variable++ {
		mask := cube.var_mask[variable]
		if !setp_disjoint(diff, mask) {
			sf_addset(result, set_merge(new_cube(), c, intersection, mask))
			gotOne = true
		}
	}
	if !gotOne && cube.num_binary_vars > 0 {
		sf_addset(result, intersection)
	}
	return result
}
