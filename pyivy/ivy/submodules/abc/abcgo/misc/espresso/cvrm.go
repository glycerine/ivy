package espresso

import (
	"fmt"
	"sort"
)

// unravel_range enumerates selected multiple-valued coordinates in source
// order. Unselected variables retain their original parts.
func unravel_range(b pcover, start, end int32) pcover {
	if start < 0 || end >= cube.num_vars || start > end+1 {
		panic("espresso: invalid unravel range")
	}
	total := int64(0)
	for row := int32(0); row < b.count; row++ {
		p := GETSET(b, row)
		expansion := int64(1)
		for v := start; v <= end; v++ {
			size := int64(set_dist(p, cube.var_mask[v]))
			if size >= 2 {
				if expansion > 1000000/size {
					panic("espresso: unreasonable expansion in unravel")
				}
				expansion *= size
			}
		}
		total += expansion
		if total > 2147483647 {
			panic("espresso: unravel cover exceeds int32 capacity")
		}
	}
	result := new_cover(int32(total))
	base := new_cube()
	for v := int32(0); v < cube.num_vars; v++ {
		if v < start || v > end {
			set_or(base, base, cube.var_mask[v])
		}
	}
	for row := int32(0); row < b.count; row++ {
		cb_unravel(GETSET(b, row), start, end, base, result)
	}
	return result
}
func cb_unravel(c pcube, start, end int32, startbase pcube, result pcover) {
	base := set_save(startbase)
	expansion := int32(1)
	for v := start; v <= end; v++ {
		size := set_dist(c, cube.var_mask[v])
		if size < 2 {
			set_or(base, base, cube.var_mask[v])
		} else {
			expansion *= size
		}
	}
	set_and(base, c, base)
	offset := result.count
	for i := int32(0); i < expansion; i++ {
		sf_addset(result, base)
	}
	place := expansion
	for v := start; v <= end; v++ {
		size := set_dist(c, cube.var_mask[v])
		if size <= 1 {
			continue
		}
		skip := place
		place /= size
		n := int32(0)
		for i := cube.first_part[v]; i <= cube.last_part[v]; i++ {
			if is_in_set(c, i) {
				for j := n; j < expansion; j += skip {
					for k := int32(0); k < place; k++ {
						set_insert(GETSET(result, j+k+offset), i)
					}
				}
				n += place
			}
		}
	}
}
func unravel(b pcover, start int32) pcover { return unravel_range(b, start, cube.num_vars-1) }
func lex_sort(t pcover) pcover {
	rows := sf_sort(t, lex_order)
	return sf_unlist(rows, t.count, t.sf_size)
}
func size_sort(t pcover) pcover {
	rows := sf_sort(t, descend)
	return sf_unlist(rows, t.count, t.sf_size)
}
func mini_sort(f pcover, compare setCompare) pcover {
	counts := sf_count(f)
	rows := sf_list(f)
	for _, p := range rows[:f.count] {
		weight := int32(0)
		for i := int32(0); i < cube.size; i++ {
			if is_in_set(p, i) {
				weight += counts[i]
			}
		}
		PUTSIZE(p, weight)
	}
	sort.SliceStable(rows[:f.count], func(i, j int) bool { return compare(rows[i], rows[j]) < 0 })
	return sf_unlist(rows, f.count, f.sf_size)
}
func sort_reduce(t pcover) pcover {
	if t.count == 0 {
		return t
	}
	var largest pcube
	best := int32(-1)
	for i := int32(0); i < t.count; i++ {
		p := GETSET(t, i)
		if size := set_ord(p); size > best {
			largest = p
			best = size
		}
	}
	rows := sf_list(t)
	for _, p := range rows[:t.count] {
		PUTSIZE(p, ((cube.num_vars-cdist(largest, p))<<7)+min(set_ord(p), 127))
	}
	sort.SliceStable(rows[:t.count], func(i, j int) bool { return descend(rows[i], rows[j]) < 0 })
	return sf_unlist(rows, t.count, t.sf_size)
}

// random_order uses Espresso's portable deterministic fallback permutation.
func random_order(f pcover) pcover {
	temp := set_new(f.sf_size)
	for i := f.count - 1; i > 0; i-- {
		k := (int64(i)*23 + 997) % int64(i)
		set_copy(temp, GETSET(f, int32(k)))
		set_copy(GETSET(f, int32(k)), GETSET(f, i))
		set_copy(GETSET(f, i), temp)
	}
	return f
}
func cubelist_partition(t *cubeList, a, b **cubeList, componentDebug uint32) int32 {
	if len(t.cubes) == 0 {
		return 0
	}
	for _, p := range t.cubes {
		RESET(p, COVERED)
	}
	seed := set_save(t.cubes[0])
	SET(t.cubes[0], COVERED)
	count := 1
	for changed := true; changed; {
		changed = false
		for _, p := range t.cubes {
			if !TESTP(p, COVERED) && ccommon(p, seed, t.cofactor) {
				set_and(seed, seed, p)
				SET(p, COVERED)
				changed = true
				count++
			}
		}
	}
	remaining := len(t.cubes) - count
	if componentDebug != 0 {
		fmt.Printf("COMPONENT_REDUCTION: split into %d %d\n", count, remaining)
	}
	if remaining != 0 {
		*a = &cubeList{cofactor: set_save(t.cofactor)}
		*b = &cubeList{cofactor: set_save(t.cofactor)}
		for _, p := range t.cubes {
			if TESTP(p, COVERED) {
				(*a).cubes = append((*a).cubes, p)
			} else {
				(*b).cubes = append((*b).cubes, p)
			}
		}
	}
	return int32(remaining)
}
func cof_output(t pcover, i int32) pcover {
	r := new_cover(t.count)
	mask := cube.var_mask[cube.output]
	for j := int32(0); j < t.count; j++ {
		p := GETSET(t, j)
		if is_in_set(p, i) {
			q := new_cube()
			set_or(q, p, mask)
			RESET(q, PRIME)
			r = sf_addset(r, q)
		}
	}
	return r
}
func uncof_output(t pcover, i int32) pcover {
	if t == nil {
		return nil
	}
	mask := cube.var_mask[cube.output]
	for j := int32(0); j < t.count; j++ {
		p := GETSET(t, j)
		set_diff(p, p, mask)
		set_insert(p, i)
	}
	return t
}
func foreach_output_function(pla pPLA, operation, save func(pPLA, int32) bool) {
	for i := int32(0); i < cube.part_size[cube.output]; i++ {
		part := i + cube.first_part[cube.output]
		one := &PLA_t{F: cof_output(pla.F, part), R: cof_output(pla.R, part), D: cof_output(pla.D, part)}
		if !operation(one, i) {
			return
		}
		one.F = uncof_output(one.F, part)
		one.R = uncof_output(one.R, part)
		one.D = uncof_output(one.D, part)
		if !save(one, i) {
			return
		}
	}
}
