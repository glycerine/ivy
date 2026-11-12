// Final sparse-variable cleanup from Espresso's sparse.c. The reduction keeps
// a mapping to original rows while irredundancy is computed on each cofactor.
package espresso

func make_sparse(f, d, r pcover) pcover {
	var cost, best cost_t
	cover_cost(f, &best)
	for {
		start := ptime()
		f = mv_reduce(f, d)
		totals(start, MV_REDUCE_TIME, f, &cost)
		if cost.total == best.total {
			break
		}
		copy_cost(&cost, &best)
		start = ptime()
		f = expand(f, r, true)
		totals(start, RAISE_IN_TIME, f, &cost)
		if cost.total == best.total {
			break
		}
		copy_cost(&cost, &best)
		if !force_irredundant {
			break
		}
	}
	return f
}

func mv_reduce(f, d pcover) pcover {
	for variable := int32(0); variable < cube.num_vars; variable++ {
		if cube.sparse[variable] == 0 {
			continue
		}
		mask := cube.var_mask[variable]
		for part := cube.first_part[variable]; part <= cube.last_part[variable]; part++ {
			// Families remain distinct: marking cofactors must not overwrite the
			// original cover before its corresponding part is deliberately removed.
			originals := make([]pcube, 0, f.count)
			onset, dcset := sf_new(f.count, cube.size), sf_new(d.count, cube.size)
			for row := int32(0); row < f.count; row++ {
				original := GETSET(f, row)
				if !is_in_set(original, part) {
					continue
				}
				originals = append(originals, original)
				cofactor := GETSET(onset, onset.count)
				onset.count++
				set_diff(cofactor, original, mask)
				set_insert(cofactor, part)
			}
			for row := int32(0); row < d.count; row++ {
				original := GETSET(d, row)
				if !is_in_set(original, part) {
					continue
				}
				cofactor := GETSET(dcset, dcset.count)
				dcset.count++
				set_diff(cofactor, original, mask)
				set_insert(cofactor, part)
			}
			mark_irredundant(onset, dcset)
			for row := int32(0); row < onset.count; row++ {
				if TESTP(GETSET(onset, row), ACTIVE) {
					continue
				}
				original := originals[row]
				if variable == cube.num_vars-1 || !setp_implies(mask, original) {
					set_remove(original, part)
				}
				RESET(original, PRIME)
			}
			sf_free(onset)
			sf_free(dcset)
		}
	}
	sf_active(f)
	for variable := int32(0); variable < cube.num_vars; variable++ {
		if cube.sparse[variable] == 0 {
			continue
		}
		for row := int32(0); row < f.count; row++ {
			current := GETSET(f, row)
			if TESTP(current, ACTIVE) && setp_disjoint(current, cube.var_mask[variable]) {
				RESET(current, ACTIVE)
				f.active_count--
			}
		}
	}
	if f.count != f.active_count {
		sf_inactive(f)
	}
	return f
}
