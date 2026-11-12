package espresso

func map_cover_to_unate(t *cubeList) pcover {
	a := sf_new(int32(len(t.cubes)), cdata.vars_unate)
	for range t.cubes {
		a = sf_addset(a, set_new(a.sf_size))
	}
	column := int32(0)
	for i := int32(0); i < cube.size; i++ {
		if cdata.part_zeros[i] > 0 {
			if column >= a.sf_size {
				panic("espresso: unate column count exceeds configured size")
			}
			for j, p := range t.cubes {
				if !is_in_set(p, i) {
					set_insert(GETSET(a, int32(j)), column)
				}
			}
			column++
		}
	}
	return a
}
func map_unate_to_cover(a pset_family) pcover {
	b := sf_new(a.count, cube.size)
	var variables []int32
	for v := int32(0); v < cube.num_vars; v++ {
		if cdata.is_unate[v] {
			variables = append(variables, v)
		}
	}
	for j := int32(0); j < a.count; j++ {
		p := GETSET(a, j)
		q := set_save(cube.fullset)
		for col, v := range variables {
			if is_in_set(p, int32(col)) {
				for i := cube.first_part[v]; i <= cube.last_part[v]; i++ {
					if cdata.part_zeros[i] == 0 {
						set_remove(q, i)
					}
				}
			}
		}
		b = sf_addset(b, q)
	}
	return b
}
func unate_compl(a pset_family) pset_family {
	for i := int32(0); i < a.count; i++ {
		p := GETSET(a, i)
		PUTSIZE(p, set_ord(p))
	}
	return sf_rev_contain(unate_complement(a))
}
func unate_complement(a pset_family) pset_family {
	if a.count == 0 {
		return sf_addset(sf_new(1, a.sf_size), set_new(a.sf_size))
	}
	if a.count == 1 {
		r := sf_new(a.sf_size, a.sf_size)
		for i := int32(0); i < a.sf_size; i++ {
			if is_in_set(GETSET(a, 0), i) {
				p := set_new(a.sf_size)
				set_insert(p, i)
				r = sf_addset(r, p)
			}
		}
		return r
	}
	restrict := set_new(a.sf_size)
	smallest := a.sf_size + 1
	for i := int32(0); i < a.count; i++ {
		p := GETSET(a, i)
		if SIZE(p) < smallest {
			set_copy(restrict, p)
			smallest = SIZE(p)
		} else if SIZE(p) == smallest {
			set_or(restrict, restrict, p)
		}
	}
	if smallest == 0 {
		return sf_new(0, a.sf_size)
	}
	if smallest == 1 {
		r := unate_complement(abs_covered_many(a, restrict))
		for i := int32(0); i < r.count; i++ {
			p := GETSET(r, i)
			set_or(p, p, restrict)
		}
		return r
	}
	pick := abs_select_restricted(a, restrict)
	r := unate_complement(abs_covered(a, pick))
	for i := int32(0); i < r.count; i++ {
		set_insert(GETSET(r, i), pick)
	}
	for i := int32(0); i < a.count; i++ {
		p := GETSET(a, i)
		if is_in_set(p, pick) {
			set_remove(p, pick)
			PUTSIZE(p, SIZE(p)-1)
		}
	}
	return sf_append(r, unate_complement(a))
}

// exact_minimum_cover enumerates all inclusion-minimal hitting sets. It does
// not mean only the covers with minimum cardinality. The leveled intersections
// retain Espresso's balanced product and containment strategy.
func exact_minimum_cover(t pset_family) pset_family {
	if t.count <= 0 {
		return sf_addset(sf_new(1, t.sf_size), set_new(t.sf_size))
	}
	t = lex_sort(sf_save(t))
	level := 0
	for n := t.count; n != 0; n >>= 1 {
		level++
	}
	type frame struct {
		family pset_family
		level  int
	}
	stack := []frame{{sf_addset(sf_new(1, t.sf_size), set_full(t.sf_size)), level}}
	for row := int32(0); row < t.count; row++ {
		p := GETSET(t, row)
		temp := sf_new(set_ord(p), t.sf_size)
		for i := int32(0); i < t.sf_size; i++ {
			if is_in_set(p, i) {
				q := set_full(t.sf_size)
				set_remove(q, i)
				temp = sf_addset(temp, q)
			}
		}
		stack = append(stack, frame{temp, level})
		for len(stack) > 1 {
			n := len(stack)
			a, b := stack[n-2], stack[n-1]
			if a.level != b.level && row != t.count-1 {
				break
			}
			temp = unateIntersect(b.family, a.family, false, true)
			stack = stack[:n-2]
			stack = append(stack, frame{temp, min(a.level, b.level) - 1})
		}
	}
	r := stack[0].family
	full := set_full(t.sf_size)
	for i := int32(0); i < r.count; i++ {
		p := GETSET(r, i)
		set_diff(p, full, p)
	}
	return r
}
func unate_intersect(a, b pset_family, largestOnly bool) pset_family {
	return unateIntersect(a, b, largestOnly, false)
}

// Empty intersections represent the full hitting set in the dual encoding.
// Keep them in exact-cover enumeration, while ordinary unate intersection
// retains the source convention of omitting empty products.
func unateIntersect(a, b pset_family, largestOnly, keepEmpty bool) pset_family {
	const batch = 500
	t := sf_new(batch, a.sf_size)
	var saved pset_family
	largest := int32(0)
	flush := func() {
		t = sf_contain(t)
		if saved == nil {
			saved = t
		} else {
			saved = sf_union(saved, t)
		}
		t = sf_new(batch, a.sf_size)
	}
	for i := int32(0); i < a.count; i++ {
		for j := int32(0); j < b.count; j++ {
			p := set_new(a.sf_size)
			if nonempty := set_andp(p, GETSET(a, i), GETSET(b, j)); !nonempty && !keepEmpty {
				continue
			}
			if largestOnly {
				size := set_ord(p)
				if size > largest {
					saved = nil
					t = sf_new(batch, a.sf_size)
					largest = size
				} else if size < largest {
					continue
				}
			}
			t = sf_addset(t, p)
			if t.count >= batch {
				flush()
			}
		}
	}
	flush()
	return saved
}
func abs_covered(a pset_family, pick int32) pset_family {
	r := sf_new(a.count, a.sf_size)
	for i := int32(0); i < a.count; i++ {
		p := GETSET(a, i)
		if !is_in_set(p, pick) {
			r = sf_addset(r, p)
		}
	}
	return r
}
func abs_covered_many(a pset_family, pick pset) pset_family {
	r := sf_new(a.count, a.sf_size)
	for i := int32(0); i < a.count; i++ {
		p := GETSET(a, i)
		if setp_disjoint(p, pick) {
			r = sf_addset(r, p)
		}
	}
	return r
}
func abs_select_restricted(a pset_family, restrict pset) int32 {
	counts := sf_count_restricted(a, restrict)
	best, maxCount := int32(-1), int32(0)
	for i, c := range counts {
		if c > maxCount {
			best, maxCount = int32(i), c
		}
	}
	if best < 0 {
		panic("espresso: restricted column has no positive weight")
	}
	return best
}
