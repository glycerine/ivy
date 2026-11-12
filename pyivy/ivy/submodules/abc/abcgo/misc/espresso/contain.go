package espresso

import "sort"

// Set lists borrow their rows. A nil final row is retained for compatibility
// with the source's pointer-list algorithms; rebuilding a family copies rows.
type setCompare func(pset, pset) int32

func sf_list(a pset_family) []pset {
	rows := make([]pset, a.count+1)
	for i := int32(0); i < a.count; i++ {
		rows[i] = GETSET(a, i)
	}
	return rows
}
func listLen(rows []pset) int {
	for i, p := range rows {
		if p == nil {
			return i
		}
	}
	return len(rows)
}
func sf_sort(a pset_family, compare setCompare) []pset {
	rows := sf_list(a)
	for _, p := range rows[:len(rows)-1] {
		PUTSIZE(p, set_ord(p))
	}
	sort.SliceStable(rows[:len(rows)-1], func(i, j int) bool { return compare(rows[i], rows[j]) < 0 })
	return rows
}
func sf_unlist(rows []pset, count, size int32) pset_family {
	if int(count) != listLen(rows) {
		panic("espresso: set-list count mismatch")
	}
	r := sf_new(count, size)
	for _, p := range rows[:count] {
		r = sf_addset(r, p)
	}
	return r
}
func rm_equal(rows []pset, compare setCompare) int32 {
	n := listLen(rows)
	dest := 0
	for i := 0; i < n; i++ {
		if i == n-1 || compare(rows[i], rows[i+1]) != 0 {
			rows[dest] = rows[i]
			dest++
		}
	}
	if dest < len(rows) {
		rows[dest] = nil
	}
	return int32(dest)
}
func removeContain(rows []pset, reverse bool) int32 {
	dest, check, last := 0, 0, int32(-1)
	for _, a := range rows[:listLen(rows)] {
		if SIZE(a) != last {
			last = SIZE(a)
			check = dest
		}
		contained := false
		for _, b := range rows[:check] {
			if (!reverse && setp_implies(a, b)) || (reverse && setp_implies(b, a)) {
				contained = true
				break
			}
		}
		if !contained {
			rows[dest] = a
			dest++
		}
	}
	if dest < len(rows) {
		rows[dest] = nil
	}
	return int32(dest)
}
func rm_contain(rows []pset) int32     { return removeContain(rows, false) }
func rm_rev_contain(rows []pset) int32 { return removeContain(rows, true) }
func sf_contain(a pset_family) pset_family {
	rows := sf_sort(a, descend)
	rm_equal(rows, descend)
	n := rm_contain(rows)
	return sf_unlist(rows, n, a.sf_size)
}
func sf_rev_contain(a pset_family) pset_family {
	rows := sf_sort(a, ascend)
	rm_equal(rows, ascend)
	n := rm_rev_contain(rows)
	return sf_unlist(rows, n, a.sf_size)
}
func sf_dupl(a pset_family) pset_family {
	rows := sf_sort(a, descend)
	n := rm_equal(rows, descend)
	return sf_unlist(rows, n, a.sf_size)
}
func sf_ind_contain(a pset_family, indices []int32) pset_family {
	rows := sf_sort(a, descend)
	rm_equal(rows, descend)
	n := rm_contain(rows)
	return sf_ind_unlist(rows, n, a.sf_size, indices, a.data)
}
func sf_ind_unlist(rows []pset, count, size int32, indices []int32, first pset) pset_family {
	r := sf_unlist(rows, count, size)
	mapped := make([]int32, count)
	for i, p := range rows[:count] {
		found := false
		for j := 0; j < len(first); j += int(r.wsize) {
			if &first[j] == &p[0] {
				mapped[i] = indices[j/int(r.wsize)]
				found = true
				break
			}
		}
		if !found {
			panic("espresso: row does not belong to indexed family")
		}
	}
	copy(indices, mapped)
	return r
}
func d1_rm_equal(rows []pset, compare setCompare) int32 {
	n := listLen(rows)
	if n == 0 {
		return 0
	}
	dest := 0
	current := rows[0]
	for _, p := range rows[1:n] {
		if compare(current, p) == 0 {
			set_or(current, current, p)
		} else {
			rows[dest] = current
			dest++
			current = p
		}
	}
	rows[dest] = current
	dest++
	if dest < len(rows) {
		rows[dest] = nil
	}
	return int32(dest)
}
func dist_merge(a pset_family, mask pset) pset_family {
	set_copy(cube.temp[0], mask)
	rows := sf_sort(a, d1_order)
	n := d1_rm_equal(rows, d1_order)
	return sf_unlist(rows, n, a.sf_size)
}
func d1merge(a pset_family, variable int32) pset_family {
	return dist_merge(a, cube.var_mask[variable])
}
func rm2_equal(a, b, e []pset, compare setCompare) int32 {
	na, nb := listLen(a), listLen(b)
	i, j, da, db, de := 0, 0, 0, 0, 0
	for i < na && j < nb {
		c := compare(a[i], b[j])
		switch {
		case c < 0:
			a[da] = a[i]
			da++
			i++
		case c > 0:
			b[db] = b[j]
			db++
			j++
		default:
			e[de] = a[i]
			de++
			i++
			j++
		}
	}
	for i < na {
		a[da] = a[i]
		da++
		i++
	}
	for j < nb {
		b[db] = b[j]
		db++
		j++
	}
	a[da] = nil
	b[db] = nil
	e[de] = nil
	return int32(de)
}
func rm2_contain(a, b []pset) int32 {
	dest := 0
	for _, p := range a[:listLen(a)] {
		keep := true
		for _, q := range b[:listLen(b)] {
			if SIZE(q) <= SIZE(p) {
				break
			}
			if setp_implies(p, q) {
				keep = false
				break
			}
		}
		if keep {
			a[dest] = p
			dest++
		}
	}
	a[dest] = nil
	return int32(dest)
}
func sf_merge(a, b, e []pset, count, size int32) pset_family {
	rows := make([]pset, 0, int(count)+1)
	rows = append(rows, a[:listLen(a)]...)
	rows = append(rows, b[:listLen(b)]...)
	rows = append(rows, e[:listLen(e)]...)
	sort.SliceStable(rows, func(i, j int) bool { return desc1(rows[i], rows[j]) < 0 })
	rows = append(rows, nil)
	return sf_unlist(rows, count, size)
}
func sf_union(a, b pset_family) pset_family {
	x, y := sf_list(a), sf_list(b)
	e := make([]pset, max(len(x), len(y)))
	n := rm2_equal(x, y, e, descend)
	n += rm2_contain(x, y) + rm2_contain(y, x)
	return sf_merge(x, y, e, n, a.sf_size)
}
