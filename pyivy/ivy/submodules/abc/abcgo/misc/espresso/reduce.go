// Espresso's ordered reduction and recursive smallest cube containing a
// complement (SCCC). Cofactor lists borrow rows; results own their cube storage.
package espresso

import "fmt"

var reduceToggle = true

func reduce(f, d pcover) pcover {
	if use_random_order {
		f = random_order(f)
	} else {
		if reduceToggle {
			f = sort_reduce(f)
		} else {
			f = mini_sort(f, descend)
		}
		reduceToggle = !reduceToggle
	}
	fd := cube2list(f, d)
	for row := int32(0); row < f.count; row++ {
		original := GETSET(f, row)
		under := reduce_cube(fd, original)
		if setp_equal(under, original) {
			SET(original, ACTIVE)
			SET(original, PRIME)
		} else {
			if debug&REDUCE != 0 {
				fmt.Printf("REDUCE: %s to %s %s\n", pc1(original), pc2(under), print_time(ptime()))
			}
			set_copy(original, under)
			RESET(original, PRIME)
			if setp_empty(under) {
				RESET(original, ACTIVE)
			} else {
				SET(original, ACTIVE)
			}
		}
		free_cube(under)
	}
	free_cubelist(fd)
	return sf_inactive(f)
}
func reduce_cube(fd *cubeList, p pcube) pcube {
	under := sccc(cofactor(fd, p))
	return set_and(under, under, p)
}
func sccc(t *cubeList) pcube { return scccAt(t, 0) }
func scccAt(t *cubeList, level int32) pcube {
	if debug&REDUCE1 != 0 {
		debug_print(t, "SCCC", level)
	}
	var result pcube
	if sccc_special_cases(t, &result) == MAYBE {
		left, right := new_cube(), new_cube()
		best := binate_split_select(t, left, right, REDUCE1)
		l := scccAt(scofactor(t, left, best), level+1)
		r := scccAt(scofactor(t, right, best), level+1)
		result = sccc_merge(l, r, left, right)
		free_cubelist(t)
	}
	if debug&REDUCE1 != 0 {
		fmt.Printf("SCCC[%d]: result is %s\n", level, pc1(result))
	}
	return result
}
func sccc_merge(left, right, cl, cr pcube) pcube {
	set_and(left, left, cl)
	set_and(right, right, cr)
	set_or(left, left, right)
	free_cube(right)
	free_cube(cl)
	free_cube(cr)
	return left
}
func sccc_cube(result, p pcube) pcube {
	if variable := cactive(p); variable >= 0 {
		temp := set_xor(cube.temp[0], p, cube.var_mask[variable])
		set_and(result, result, temp)
	}
	return result
}
func sccc_special_cases(t *cubeList, result *pcube) int32 {
	if len(t.cubes) == 0 {
		*result = set_save(cube.fullset)
		free_cubelist(t)
		return 1
	}
	for _, p := range t.cubes {
		if full_row(p, t.cofactor) {
			*result = new_cube()
			free_cubelist(t)
			return 1
		}
	}
	massive_count(t)
	if cdata.vars_unate == cdata.vars_active || len(t.cubes) == 1 {
		*result = set_save(cube.fullset)
		for _, p := range t.cubes {
			sccc_cube(*result, set_or(cube.temp[1], p, t.cofactor))
		}
		free_cubelist(t)
		return 1
	}
	ceiling := set_save(t.cofactor)
	for _, p := range t.cubes {
		set_or(ceiling, ceiling, p)
	}
	if !setp_equal(ceiling, cube.fullset) {
		*result = sccc_cube(set_save(cube.fullset), ceiling)
		if setp_equal(*result, cube.fullset) {
			free_cube(ceiling)
		} else {
			// Finish creating the cofactor before recursive calls alter scratch data.
			child := cofactor(t, ceiling)
			*result = sccc_merge(sccc(child), set_save(cube.fullset), ceiling, *result)
		}
		free_cubelist(t)
		return 1
	}
	free_cube(ceiling)
	if cdata.vars_active == 1 {
		*result = new_cube()
		free_cubelist(t)
		return 1
	}
	if cdata.var_zeros[cdata.best] < int32(len(t.cubes))/2 {
		var a, b *cubeList
		if cubelist_partition(t, &a, &b, debug&REDUCE1) == 0 {
			return MAYBE
		}
		free_cubelist(t)
		*result = sccc(a)
		other := sccc(b)
		set_and(*result, *result, other)
		free_cube(other)
		return 1
	}
	return MAYBE
}
