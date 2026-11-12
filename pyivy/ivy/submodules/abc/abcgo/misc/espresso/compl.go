package espresso

import "sort"

const (
	USE_COMPL_LIFT = iota
	USE_COMPL_LIFT_ONSET
	USE_COMPL_LIFT_ONSET_COMPLEX
	NO_LIFTING
)

// complement follows the unate recursive paradigm: factor constant columns,
// solve unate leaves, split a binate variable, then merge and lift both halves.
func complement(t *cubeList) pcover {
	var result pcover
	if compl_special_cases(t, &result) != MAYBE {
		return result
	}
	left, right := new_cube(), new_cube()
	variable := binate_split_select(t, left, right, COMPL)
	l, r := complement(scofactor(t, left, variable)), complement(scofactor(t, right, variable))
	lifting := USE_COMPL_LIFT
	if int64(l.count)*int64(r.count) > int64(l.count+r.count)*int64(len(t.cubes)) {
		lifting = USE_COMPL_LIFT_ONSET
	}
	result = compl_merge(t, l, r, left, right, variable, lifting)
	free_cubelist(t)
	return result
}
func compl_special_cases(t *cubeList, result *pcover) int32 {
	if len(t.cubes) == 0 {
		*result = sf_addset(new_cover(1), cube.fullset)
		free_cubelist(t)
		return 1
	}
	if len(t.cubes) == 1 {
		*result = compl_cube(set_or(t.cofactor, t.cofactor, t.cubes[0]))
		free_cubelist(t)
		return 1
	}
	for _, p := range t.cubes {
		if full_row(p, t.cofactor) {
			*result = new_cover(0)
			free_cubelist(t)
			return 1
		}
	}
	ceil := complCeiling(t)
	if !setp_equal(ceil, cube.fullset) {
		extra := compl_cube(ceil)
		set_or(t.cofactor, t.cofactor, set_diff(new_cube(), cube.fullset, ceil))
		*result = sf_append(complement(t), extra)
		return 1
	}
	massive_count(t)
	if cdata.vars_active == 1 {
		*result = new_cover(0)
		free_cubelist(t)
		return 1
	}
	if cdata.vars_unate == cdata.vars_active {
		a := map_cover_to_unate(t)
		free_cubelist(t)
		a = unate_compl(a)
		*result = map_unate_to_cover(a)
		sf_free(a)
		return 1
	}
	return MAYBE
}
func complCeiling(t *cubeList) pcube {
	p := set_save(t.cofactor)
	for _, row := range t.cubes {
		set_or(p, p, row)
	}
	return p
}
func compl_merge(t *cubeList, l, r pcover, left, right pcube, variable int32, lifting int) pcover {
	for _, branch := range []struct {
		f    pcover
		mask pcube
	}{{l, left}, {r, right}} {
		for i := int32(0); i < branch.f.count; i++ {
			p := GETSET(branch.f, i)
			set_and(p, p, branch.mask)
			SET(p, ACTIVE)
		}
	}
	a, b := sf_list(l), sf_list(r)
	a, b = a[:l.count], b[:r.count]
	set_copy(cube.temp[0], cube.var_mask[variable])
	sort.SliceStable(a, func(i, j int) bool { return d1_order(a[i], a[j]) < 0 })
	sort.SliceStable(b, func(i, j int) bool { return d1_order(b[i], b[j]) < 0 })
	compl_d1merge(a, b)
	switch lifting {
	case USE_COMPL_LIFT:
		compl_lift(a, b, right, variable)
		compl_lift(b, a, left, variable)
	case USE_COMPL_LIFT_ONSET, USE_COMPL_LIFT_ONSET_COMPLEX:
		onset := cubeunlist(t)
		if lifting == USE_COMPL_LIFT_ONSET {
			compl_lift_onset(a, onset, right, variable)
			compl_lift_onset(b, onset, left, variable)
		} else {
			compl_lift_onset_complex(a, onset, variable)
			compl_lift_onset_complex(b, onset, variable)
		}
		free_cover(onset)
	case NO_LIFTING:
	}
	result := new_cover(l.count + r.count)
	for i := int32(0); i < l.count; i++ {
		sf_addset(result, GETSET(l, i))
	}
	for i := int32(0); i < r.count; i++ {
		p := GETSET(r, i)
		if TESTP(p, ACTIVE) {
			sf_addset(result, p)
		}
	}
	free_cover(l)
	free_cover(r)
	return result
}
func compl_d1merge(a, b []pset) {
	a, b = a[:listLen(a)], b[:listLen(b)]
	for i, j := 0, 0; i < len(a) && j < len(b); {
		switch order := d1_order(a[i], b[j]); {
		case order > 0:
			j++
		case order < 0:
			i++
		default:
			RESET(b[j], ACTIVE)
			set_or(a[i], a[i], b[j])
			j++
		}
	}
}
func compl_lift(a, b []pset, bcube pcube, variable int32) {
	a, b = a[:listLen(a)], b[:listLen(b)]
	mask := cube.var_mask[variable]
	liftor := set_and(new_cube(), bcube, mask)
	lift := new_cube()
	for _, p := range a {
		if TESTP(p, ACTIVE) {
			set_merge(lift, bcube, p, mask)
			for _, q := range b {
				if setp_implies(lift, q) {
					set_or(p, p, liftor)
					break
				}
			}
		}
	}
}
func compl_lift_onset(a []pset, onset pcover, bcube pcube, variable int32) {
	a = a[:listLen(a)]
	for _, p := range a {
		if !TESTP(p, ACTIVE) {
			continue
		}
		lift := set_and(new_cube(), bcube, cube.var_mask[variable])
		set_or(lift, p, lift)
		intersects := false
		for i := int32(0); i < onset.count; i++ {
			if cdist0(GETSET(onset, i), lift) {
				intersects = true
				break
			}
		}
		if !intersects {
			set_copy(p, lift)
			SET(p, ACTIVE)
		}
	}
}
func compl_lift_onset_complex(a []pset, onset pcover, variable int32) {
	a = a[:listLen(a)]
	for _, p := range a {
		if !TESTP(p, ACTIVE) {
			continue
		}
		lower := new_cube()
		for i := int32(0); i < onset.count; i++ {
			q := GETSET(onset, i)
			distance := cdist01(q, p)
			if distance == 0 {
				panic("espresso: ON-set and OFF-set are not orthogonal")
			}
			if distance == 1 {
				force_lower(lower, q, p)
			}
		}
		set_diff(lower, cube.var_mask[variable], lower)
		set_or(p, p, lower)
	}
}
func compl_cube(p pcube) pcover {
	result := new_cover(cube.num_vars)
	diff := set_diff(new_cube(), cube.fullset, p)
	for _, mask := range cube.var_mask {
		if !setp_disjoint(diff, mask) {
			sf_addset(result, set_merge(new_cube(), diff, cube.fullset, mask))
		}
	}
	return result
}

// simp_comp returns both polarities from the same recursive decomposition.
func simp_comp(t *cubeList, on, off *pcover) {
	if simp_comp_special_cases(t, on, off) != MAYBE {
		return
	}
	left, right := new_cube(), new_cube()
	variable := binate_split_select(t, left, right, COMPL)
	var l, r, lb, rb pcover
	simp_comp(scofactor(t, left, variable), &l, &lb)
	simp_comp(scofactor(t, right, variable), &r, &rb)
	*on = compl_merge(t, l, r, left, right, variable, USE_COMPL_LIFT)
	*off = compl_merge(t, lb, rb, left, right, variable, USE_COMPL_LIFT)
	if int((*on).count) > len(t.cubes) {
		sf_free(*on)
		*on = cubeunlist(t)
	}
	free_cubelist(t)
}
func simp_comp_special_cases(t *cubeList, on, off *pcover) int32 {
	if len(t.cubes) == 0 {
		*on = new_cover(0)
		*off = sf_addset(new_cover(1), cube.fullset)
		free_cubelist(t)
		return 1
	}
	if len(t.cubes) == 1 {
		p := set_or(t.cofactor, t.cofactor, t.cubes[0])
		*on = sf_addset(new_cover(1), p)
		*off = compl_cube(p)
		free_cubelist(t)
		return 1
	}
	for _, p := range t.cubes {
		if full_row(p, t.cofactor) {
			*on = sf_addset(new_cover(1), cube.fullset)
			*off = new_cover(0)
			free_cubelist(t)
			return 1
		}
	}
	ceil := complCeiling(t)
	if !setp_equal(ceil, cube.fullset) {
		set_or(t.cofactor, t.cofactor, set_diff(new_cube(), cube.fullset, ceil))
		simp_comp(t, on, off)
		for i := int32(0); i < (*on).count; i++ {
			p := GETSET(*on, i)
			set_and(p, p, ceil)
		}
		*off = sf_append(*off, compl_cube(ceil))
		return 1
	}
	massive_count(t)
	if cdata.vars_active == 1 {
		*on = sf_addset(new_cover(1), cube.fullset)
		*off = new_cover(0)
		free_cubelist(t)
		return 1
	}
	if cdata.vars_unate == cdata.vars_active {
		*on = sf_contain(cubeunlist(t))
		a := unate_compl(map_cover_to_unate(t))
		*off = map_unate_to_cover(a)
		sf_free(a)
		free_cubelist(t)
		return 1
	}
	return MAYBE
}
func simplify(t *cubeList) pcover {
	var result pcover
	if simplify_special_cases(t, &result) != MAYBE {
		return result
	}
	left, right := new_cube(), new_cube()
	variable := binate_split_select(t, left, right, COMPL)
	l, r := simplify(scofactor(t, left, variable)), simplify(scofactor(t, right, variable))
	result = compl_merge(t, l, r, left, right, variable, USE_COMPL_LIFT)
	if int(result.count) > len(t.cubes) {
		sf_free(result)
		result = cubeunlist(t)
	}
	free_cubelist(t)
	return result
}
func simplify_special_cases(t *cubeList, result *pcover) int32 {
	if len(t.cubes) == 0 {
		*result = new_cover(0)
		free_cubelist(t)
		return 1
	}
	if len(t.cubes) == 1 {
		*result = sf_addset(new_cover(1), set_or(t.cofactor, t.cofactor, t.cubes[0]))
		free_cubelist(t)
		return 1
	}
	for _, p := range t.cubes {
		if full_row(p, t.cofactor) {
			*result = sf_addset(new_cover(1), cube.fullset)
			free_cubelist(t)
			return 1
		}
	}
	ceil := complCeiling(t)
	if !setp_equal(ceil, cube.fullset) {
		set_or(t.cofactor, t.cofactor, set_diff(new_cube(), cube.fullset, ceil))
		*result = simplify(t)
		for i := int32(0); i < (*result).count; i++ {
			p := GETSET(*result, i)
			set_and(p, p, ceil)
		}
		return 1
	}
	massive_count(t)
	if cdata.vars_active == 1 {
		*result = sf_addset(new_cover(1), cube.fullset)
		free_cubelist(t)
		return 1
	}
	if cdata.vars_unate == cdata.vars_active {
		*result = sf_contain(cubeunlist(t))
		free_cubelist(t)
		return 1
	}
	return MAYBE
}
