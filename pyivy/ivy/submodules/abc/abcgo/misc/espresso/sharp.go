package espresso

// Sharp returns the Boolean difference while retaining the source's bounded
// per-variable cube decomposition. No minterm expansion is required.
func cv_sharp(a, b pcover) pcover {
	result := new_cover(0)
	for i := int32(0); i < a.count; i++ {
		result = sf_union(result, cb_sharp(GETSET(a, i), b))
	}
	return result
}
func cb_sharp(c pcube, t pcover) pcover {
	if t.count == 0 {
		return sf_addset(new_cover(1), c)
	}
	return cb_recur_sharp(c, t, 0, t.count-1, 0)
}
func cb_recur_sharp(c pcube, t pcover, first, last, level int32) pcover {
	if first < 0 || last >= t.count || first > last {
		panic("espresso: invalid recursive sharp range")
	}
	if first == last {
		return sharp(c, GETSET(t, first))
	}
	middle := first + (last-first)/2
	left, right := cb_recur_sharp(c, t, first, middle, level+1), cb_recur_sharp(c, t, middle+1, last, level+1)
	result := cv_intersect(left, right)
	free_cover(left)
	free_cover(right)
	return result
}
func sharp(a, b pcube) pcover {
	result := new_cover(cube.num_vars)
	if !cdist0(a, b) {
		return sf_addset(result, a)
	}
	diff := set_diff(new_cube(), a, b)
	for _, mask := range cube.var_mask[:cube.num_vars] {
		part := set_and(new_cube(), diff, mask)
		if !setp_empty(part) {
			sf_addset(result, set_or(part, part, set_diff(new_cube(), a, mask)))
		}
	}
	return result
}
func make_disjoint(a pcover) pcover {
	result := new_cover(0)
	for i := int32(0); i < a.count; i++ {
		result = sf_append(result, cb_dsharp(GETSET(a, i), result))
	}
	return result
}
func cv_dsharp(a, b pcover) pcover {
	result := new_cover(0)
	for i := int32(0); i < a.count; i++ {
		result = sf_union(result, cb_dsharp(GETSET(a, i), b))
	}
	return result
}
func cb1_dsharp(t pcover, c pcube) pcover {
	result := new_cover(t.count)
	for i := int32(0); i < t.count; i++ {
		result = sf_union(result, dsharp(GETSET(t, i), c))
	}
	return result
}
func cb_dsharp(c pcube, t pcover) pcover {
	result := sf_addset(new_cover(max(t.count, 1)), c)
	for i := int32(0); i < t.count; i++ {
		next := cb1_dsharp(result, GETSET(t, i))
		free_cover(result)
		result = next
	}
	return result
}
func dsharp(a, b pcube) pcover {
	result := new_cover(cube.num_vars)
	if !cdist0(a, b) {
		return sf_addset(result, a)
	}
	diff, intersection, mask := set_diff(new_cube(), a, b), set_and(new_cube(), a, b), new_cube()
	for _, variableMask := range cube.var_mask[:cube.num_vars] {
		if setp_disjoint(diff, variableMask) {
			continue
		}
		part := set_and(new_cube(), diff, variableMask)
		set_or(part, part, set_and(new_cube(), intersection, mask))
		set_or(mask, mask, variableMask)
		set_or(part, part, set_diff(new_cube(), a, mask))
		sf_addset(result, part)
	}
	return result
}

const sharpBatchSize int32 = 500

func cv_intersect(a, b pcover) pcover {
	batch := new_cover(sharpBatchSize)
	var saved pcover
	flush := func() {
		batch = sf_contain(batch)
		if saved == nil {
			saved = batch
		} else {
			saved = sf_union(saved, batch)
		}
	}
	for i := int32(0); i < a.count; i++ {
		p := GETSET(a, i)
		for j := int32(0); j < b.count; j++ {
			q := GETSET(b, j)
			if cdist0(p, q) {
				sf_addset(batch, set_and(new_cube(), p, q))
				if batch.count >= batch.capacity {
					flush()
					batch = new_cover(sharpBatchSize)
				}
			}
		}
	}
	flush()
	return saved
}
