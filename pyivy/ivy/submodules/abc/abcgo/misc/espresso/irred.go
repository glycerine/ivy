package espresso

// irredundant retains the original cover rows and selects a covering subset.
func irredundant(f, d pcover) pcover { mark_irredundant(f, d); return sf_inactive(f) }
func mark_irredundant(f, d pcover) {
	var e, rt, rp pcover
	irred_split_cover(f, d, &e, &rt, &rp)
	table := irred_derive_table(d, e, rp)
	selected := sm_minimum_cover(table, nil, 1, 0)
	for i := int32(0); i < f.count; i++ {
		p := GETSET(f, i)
		RESET(p, ACTIVE)
		RESET(p, RELESSEN)
	}
	for i := int32(0); i < e.count; i++ {
		p := GETSET(e, i)
		q := GETSET(f, SIZE(p))
		if !setp_equal(q, p) {
			panic("espresso: irredundant row identity lost")
		}
		SET(q, ACTIVE)
		SET(q, RELESSEN)
	}
	for pe := selected.first_col; pe != nil; pe = pe.next_col {
		SET(GETSET(f, int32(pe.col_num)), ACTIVE)
	}
	free_cover(e)
	free_cover(rt)
	free_cover(rp)
	sm_free(table)
	sm_row_free(selected)
}
func irred_split_cover(f, d pcover, e, rt, rp *pcover) {
	for i := int32(0); i < f.count; i++ {
		PUTSIZE(GETSET(f, i), i)
	}
	*e = new_cover(10)
	*rt = new_cover(10)
	*rp = new_cover(10)
	r := new_cover(10)
	fd := cube2list(f, d)
	for i := int32(0); i < f.count; i++ {
		p := GETSET(f, i)
		if cube_is_covered(fd, p) {
			sf_addset(r, p)
		} else {
			sf_addset(*e, p)
		}
	}
	free_cubelist(fd)
	ed := cube2list(*e, d)
	for i := int32(0); i < r.count; i++ {
		p := GETSET(r, i)
		if cube_is_covered(ed, p) {
			sf_addset(*rt, p)
		} else {
			sf_addset(*rp, p)
		}
	}
	free_cubelist(ed)
	free_cover(r)
}
func irred_derive_table(d, e, rp pcover) *sm_matrix {
	for _, f := range []pcover{d, e} {
		if f != nil {
			for i := int32(0); i < f.count; i++ {
				RESET(GETSET(f, i), REDUND)
			}
		}
	}
	for i := int32(0); i < rp.count; i++ {
		SET(GETSET(rp, i), REDUND)
	}
	list := cube3list(d, e, rp)
	table := sm_alloc()
	lastDominance := 0
	for i := int32(0); i < rp.count; i++ {
		p := GETSET(rp, i)
		irredFindTautology(cofactor(list, p), table, int(SIZE(p)))
		RESET(p, REDUND)
		if table.nrows-lastDominance > 1000 {
			sm_row_dominance(table)
			lastDominance = table.nrows
		}
	}
	free_cubelist(list)
	return table
}
func cube_is_covered(t *cubeList, c pcube) bool { return tautology(cofactor(t, c)) }
func tautology(t *cubeList) bool {
	result := taut_special_cases(t)
	if result != MAYBE {
		return result != 0
	}
	left, right := new_cube(), new_cube()
	best := binate_split_select(t, left, right, TAUT)
	// Both child lists must be formed before disposing the borrowed parent list.
	a, b := scofactor(t, left, best), scofactor(t, right, best)
	free_cubelist(t)
	if !tautology(a) {
		free_cubelist(b)
		return false
	}
	return tautology(b)
}
func taut_special_cases(t *cubeList) int32 {
	for _, p := range t.cubes {
		if full_row(p, t.cofactor) {
			free_cubelist(t)
			return 1
		}
	}
	for {
		ceil := set_save(t.cofactor)
		for _, p := range t.cubes {
			set_or(ceil, ceil, p)
		}
		if !setp_equal(ceil, cube.fullset) {
			free_cubelist(t)
			return 0
		}
		massive_count(t)
		if cdata.vars_unate == cdata.vars_active {
			free_cubelist(t)
			return 0
		}
		if cdata.vars_active == 1 {
			free_cubelist(t)
			return 1
		}
		if cdata.vars_unate != 0 {
			irredUnateReduce(t)
			continue
		}
		if cdata.var_zeros[cdata.best] < CUBELISTSIZE(t)/2 {
			var a, b *cubeList
			if cubelist_partition(t, &a, &b, debug&TAUT) != 0 {
				free_cubelist(t)
				if tautology(a) {
					free_cubelist(b)
					return 1
				}
				if tautology(b) {
					return 1
				}
				return 0
			}
		}
		return MAYBE
	}
}
func irredUnateReduce(t *cubeList) {
	mask := new_cube()
	for v := int32(0); v < cube.num_vars; v++ {
		if cdata.is_unate[v] {
			set_or(mask, mask, cube.var_mask[v])
		}
	}
	out := t.cubes[:0]
	temp := new_cube()
	for _, p := range t.cubes {
		if setp_implies(mask, set_or(temp, p, t.cofactor)) {
			out = append(out, p)
		}
	}
	t.cubes = out
}

// The current Rp column is passed explicitly instead of C's global Rp_current,
// so independent covering-table derivations do not share hidden recursion state.
func irredFindTautology(t *cubeList, table *sm_matrix, current int) {
	if irredTautSpecial(t, table, current) != MAYBE {
		return
	}
	left, right := new_cube(), new_cube()
	best := binate_split_select(t, left, right, TAUT)
	a, b := scofactor(t, left, best), scofactor(t, right, best)
	free_cubelist(t)
	irredFindTautology(a, table, current)
	irredFindTautology(b, table, current)
}
func irredTautSpecial(t *cubeList, table *sm_matrix, current int) int32 {
	for _, p := range t.cubes {
		if !TESTP(p, REDUND) && full_row(p, t.cofactor) {
			free_cubelist(t)
			return 1
		}
	}
	for {
		massive_count(t)
		if cdata.vars_unate == cdata.vars_active {
			row := 0
			if table.last_row != nil {
				row = table.last_row.row_num + 1
			}
			sm_insert(table, row, current)
			for _, p := range t.cubes {
				if TESTP(p, REDUND) && full_row(p, t.cofactor) {
					sm_insert(table, row, int(SIZE(p)))
				}
			}
			free_cubelist(t)
			return 1
		}
		if cdata.vars_unate != 0 {
			irredUnateReduce(t)
			continue
		}
		return MAYBE
	}
}
