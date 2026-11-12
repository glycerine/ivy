package espresso

import "sort"

// Cofactor lists borrow rows and own only their accumulated cofactor mask.
func sameCube(a, b pcube) bool { return len(a) > 0 && len(b) > 0 && &a[0] == &b[0] }
func cofactor(t *cubeList, c pcube) *cubeList {
	out := &cubeList{cofactor: set_or(new_cube(), t.cofactor, set_diff(new_cube(), cube.fullset, c))}
	for _, p := range t.cubes {
		if !sameCube(p, c) && cdist0(p, c) {
			out.cubes = append(out.cubes, p)
		}
	}
	return out
}
func scofactor(t *cubeList, c pcube, variable int32) *cubeList {
	out := &cubeList{cofactor: set_or(new_cube(), t.cofactor, set_diff(new_cube(), cube.fullset, c))}
	for _, p := range t.cubes {
		if !sameCube(p, c) && cubeVarIntersects(p, c, variable) {
			out.cubes = append(out.cubes, p)
		}
	}
	return out
}
func massive_count(t *cubeList) {
	clear(cdata.part_zeros)
	clear(cdata.var_zeros)
	clear(cdata.parts_active)
	clear(cdata.is_unate)
	for _, p := range t.cubes {
		for i := int32(0); i < cube.size; i++ {
			if !is_in_set(p, i) && !is_in_set(t.cofactor, i) {
				cdata.part_zeros[i]++
			}
		}
	}
	cdata.vars_active = 0
	cdata.vars_unate = 0
	cdata.best = -1
	mostactive, mostzero, balanced := int32(0), int32(0), int32(32000)
	for v := int32(0); v < cube.num_vars; v++ {
		active, maxzero := int32(0), int32(0)
		for i := cube.first_part[v]; i <= cube.last_part[v]; i++ {
			n := cdata.part_zeros[i]
			cdata.var_zeros[v] += n
			if n > 0 {
				active++
			}
			if n > maxzero {
				maxzero = n
			}
		}
		if active > mostactive || (active == mostactive && (cdata.var_zeros[v] > mostzero || (cdata.var_zeros[v] == mostzero && maxzero < balanced))) {
			cdata.best = v
			mostactive = active
			mostzero = cdata.var_zeros[v]
			balanced = maxzero
		}
		cdata.parts_active[v] = active
		cdata.is_unate[v] = active == 1
		if active > 0 {
			cdata.vars_active++
		}
		if active == 1 {
			cdata.vars_unate++
		}
	}
}
func binate_split_select(t *cubeList, left, right pcube, debugFlag uint32) int32 {
	best := cdata.best
	if best < 0 {
		panic("espresso: no active split variable")
	}
	set_diff(left, cube.fullset, cube.var_mask[best])
	set_copy(right, left)
	available := []int32{}
	for i := cube.first_part[best]; i <= cube.last_part[best]; i++ {
		if !is_in_set(t.cofactor, i) {
			available = append(available, i)
		}
	}
	if len(available) < 2 {
		panic("espresso: unsplittable variable")
	}
	for j, i := range available {
		if j < len(available)/2 {
			set_insert(left, i)
		} else {
			set_insert(right, i)
		}
	}
	return best
}
func cubeLists(covers ...pcover) *cubeList {
	t := &cubeList{cofactor: new_cube()}
	for _, f := range covers {
		if f != nil {
			for i := int32(0); i < f.count; i++ {
				t.cubes = append(t.cubes, GETSET(f, i))
			}
		}
	}
	return t
}
func cube1list(a pcover) *cubeList       { return cubeLists(a) }
func cube2list(a, b pcover) *cubeList    { return cubeLists(a, b) }
func cube3list(a, b, c pcover) *cubeList { return cubeLists(a, b, c) }
func cubeunlist(t *cubeList) pcover {
	f := new_cover(int32(len(t.cubes)))
	for _, p := range t.cubes {
		sf_addset(f, set_or(new_cube(), p, t.cofactor))
	}
	return f
}
func simplify_cubelist(t *cubeList) {
	// Sort normalized rows and remove duplicates, retaining the first row even for
	// singleton lists (the C pointer loop accidentally omitted it).
	sort.SliceStable(t.cubes, func(i, j int) bool {
		a, b := t.cubes[i], t.cubes[j]
		for w := LOOP(a); w > 0; w-- {
			x, y := a[w]|t.cofactor[w], b[w]|t.cofactor[w]
			if x != y {
				return x < y
			}
		}
		return false
	})
	out := t.cubes[:0]
	for _, p := range t.cubes {
		equal := len(out) > 0
		if equal {
			q := out[len(out)-1]
			for w := LOOP(p); w > 0; w-- {
				if p[w]|t.cofactor[w] != q[w]|t.cofactor[w] {
					equal = false
					break
				}
			}
		}
		if !equal {
			out = append(out, p)
		}
	}
	t.cubes = out
}
