package espresso

// do_sm_minimum_cover translates a dense family into the sparse covering
// problem. As in Espresso, it uses heuristic minimum cover (not exact search).
func do_sm_minimum_cover(a pset_family) pset {
	m := sm_alloc()
	for row := int32(0); row < a.count; row++ {
		p := GETSET(a, row)
		for column := int32(0); column < a.sf_size; column++ {
			if is_in_set(p, column) {
				sm_insert(m, int(row), int(column))
			}
		}
	}
	chosen := sm_minimum_cover(m, nil, 1, 0)
	result := set_new(a.sf_size)
	for e := chosen.first_col; e != nil; e = e.next_col {
		set_insert(result, int32(e.col_num))
	}
	return result
}
