// Cube geometry from Espresso's setc.c. Binary variables retain their packed
// two-bit tests; multiple-valued variables use the precomputed per-variable masks.
package espresso

func full_row(p, cofactor pcube) bool {
	for i := LOOP(p); i > 0; i-- {
		if p[i]|cofactor[i] != cube.fullset[i] {
			return false
		}
	}
	return true
}
func cubeVarIntersects(a, b pcube, variable int32) bool {
	mask := cube.var_mask[variable]
	for w := cube.first_word[variable]; w <= cube.last_word[variable]; w++ {
		if a[w]&b[w]&mask[w] != 0 {
			return true
		}
	}
	return false
}
func cubeDistance(a, b pcube, limit int32) int32 {
	var distance int32
	for w := int32(1); w <= cube.inword; w++ {
		mask := uint32(DISJOINT)
		if w == cube.inword {
			mask = cube.inmask
		}
		x := a[w] & b[w]
		distance += count_ones(^(x | x>>1) & mask)
		if distance >= limit {
			return limit
		}
	}
	for variable := cube.num_binary_vars; variable < cube.num_vars; variable++ {
		if !cubeVarIntersects(a, b, variable) {
			distance++
			if distance >= limit {
				return limit
			}
		}
	}
	return distance
}
func cdist0(a, b pcube) bool   { return cubeDistance(a, b, 1) == 0 }
func cdist01(a, b pcube) int32 { return cubeDistance(a, b, 2) }
func cdist(a, b pcube) int32   { return cubeDistance(a, b, cube.num_vars+1) }
func force_lower(lower, a, b pcube) pset {
	for w := int32(1); w <= cube.inword; w++ {
		mask := uint32(DISJOINT)
		if w == cube.inword {
			mask = cube.inmask
		}
		x := a[w] & b[w]
		disjoint := ^(x | x>>1) & mask
		lower[w] |= (disjoint | disjoint<<1) & a[w]
	}
	for variable := cube.num_binary_vars; variable < cube.num_vars; variable++ {
		if cubeVarIntersects(a, b, variable) {
			continue
		}
		mask := cube.var_mask[variable]
		for w := cube.first_word[variable]; w <= cube.last_word[variable]; w++ {
			lower[w] |= a[w] & mask[w]
		}
	}
	return lower
}
func consensus(result, a, b pcube) {
	// Ordinary C call sites use distinct scratch. Preserve logical input values
	// for Go callers that deliberately reuse either input as the destination.
	if len(result) > 0 && len(a) > 0 && &result[0] == &a[0] {
		a = set_save(a)
	}
	if len(result) > 0 && len(b) > 0 && &result[0] == &b[0] {
		b = set_save(b)
	}
	set_clear(result, cube.size)
	for w := int32(1); w <= cube.inword; w++ {
		mask := uint32(DISJOINT)
		if w == cube.inword {
			mask = cube.inmask
		}
		x := a[w] & b[w]
		disjoint := ^(x | x>>1) & mask
		result[w] = x | ((disjoint | disjoint<<1) & (a[w] | b[w]))
	}
	for variable := cube.num_binary_vars; variable < cube.num_vars; variable++ {
		mask := cube.var_mask[variable]
		intersects := cubeVarIntersects(a, b, variable)
		for w := cube.first_word[variable]; w <= cube.last_word[variable]; w++ {
			if intersects {
				result[w] |= a[w] & b[w] & mask[w]
			} else {
				result[w] |= (a[w] | b[w]) & mask[w]
			}
		}
	}
}
func cactive(a pcube) int32 {
	active := int32(-1)
	var count int32
	for w := int32(1); w <= cube.inword; w++ {
		mask := uint32(DISJOINT)
		if w == cube.inword {
			mask = cube.inmask
		}
		x := a[w]
		missing := ^(x & (x >> 1)) & mask
		if missing == 0 {
			continue
		}
		count += count_ones(missing)
		if count > 1 {
			return -1
		}
		active = (w-1)*(BPI/2) + bit_index(missing)/2
	}
	for variable := cube.num_binary_vars; variable < cube.num_vars; variable++ {
		mask := cube.var_mask[variable]
		for w := cube.first_word[variable]; w <= cube.last_word[variable]; w++ {
			if mask[w]&^a[w] != 0 {
				count++
				if count > 1 {
					return -1
				}
				active = variable
				break
			}
		}
	}
	return active
}
func ccommon(a, b, cofactor pcube) bool {
	for w := int32(1); w <= cube.inword; w++ {
		mask := uint32(DISJOINT)
		if w == cube.inword {
			mask = cube.inmask
		}
		x, y := a[w]|cofactor[w], b[w]|cofactor[w]
		if ^(x&(x>>1))&^(y&(y>>1))&mask != 0 {
			return true
		}
	}
	for variable := cube.num_binary_vars; variable < cube.num_vars; variable++ {
		mask := cube.var_mask[variable]
		missingA, missingB := false, false
		for w := cube.first_word[variable]; w <= cube.last_word[variable]; w++ {
			missingA = missingA || mask[w]&^a[w]&^cofactor[w] != 0
			missingB = missingB || mask[w]&^b[w]&^cofactor[w] != 0
		}
		if missingA && missingB {
			return true
		}
	}
	return false
}
func lex_order(a, b pset) int32 {
	for i := LOOP(a); i > 0; i-- {
		if a[i] > b[i] {
			return -1
		}
		if a[i] < b[i] {
			return 1
		}
	}
	return 0
}
func descend(a, b pset) int32 {
	if SIZE(a) > SIZE(b) {
		return -1
	}
	if SIZE(a) < SIZE(b) {
		return 1
	}
	return lex_order(a, b)
}
func ascend(a, b pset) int32 { return -descend(a, b) }
func d1_order(a, b pset) int32 {
	mask := cube.temp[0]
	for i := LOOP(a); i > 0; i-- {
		x, y := a[i]|mask[i], b[i]|mask[i]
		if x > y {
			return -1
		}
		if x < y {
			return 1
		}
	}
	return 0
}
func desc1(a, b pset) int32 {
	if a == nil {
		if b == nil {
			return 0
		}
		return 1
	}
	if b == nil {
		return -1
	}
	return descend(a, b)
}
