package espresso

// espresso retains the original expand/reduce/irredundant improvement loops.
// Covers and cube flags are mutable; callers supply a disjoint ON/OFF pair.
func espresso(f, dontCare, off pcover) pcover {
	for {
		original, d := sf_save(f), sf_save(dontCare)
		if recompute_onset {
			start := ptime()
			f = simplify(cube1list(f))
			if trace {
				print_trace(f, "SIMPLIFY   ", ptime()-start)
			}
		}
		var cost, best cost_t
		execute := func(index int32, operation func() pcover) pcover {
			start := ptime()
			result := operation()
			totals(start, index, result, &cost)
			return result
		}
		cover_cost(f, &cost)
		if unwrap_onset && cube.part_size[cube.num_vars-1] > 1 && cost.out != cost.cubes*cube.part_size[cube.num_vars-1] && cost.out < 5000 {
			start := ptime()
			f = sf_contain(unravel(f, cube.num_vars-1))
			if trace {
				print_trace(f, "SETUP      ", ptime()-start)
			}
		}
		for i := int32(0); i < f.count; i++ {
			RESET(GETSET(f, i), PRIME)
		}
		f = execute(EXPAND_TIME, func() pcover { return expand(f, off, false) })
		cover_cost(f, &cost)
		f = execute(IRRED_TIME, func() pcover { return irredundant(f, d) })
		cover_cost(f, &cost)
		if !single_expand {
			e := new_cover(0)
			if remove_essential {
				e = execute(ESSEN_TIME, func() pcover { return essential(&f, &d) })
			}
			cover_cost(f, &cost)
			for {
				for {
					best = cost
					f = execute(REDUCE_TIME, func() pcover { return reduce(f, d) })
					f = execute(EXPAND_TIME, func() pcover { return expand(f, off, false) })
					f = execute(IRRED_TIME, func() pcover { return irredundant(f, d) })
					cover_cost(f, &cost)
					if cost.cubes >= best.cubes {
						break
					}
				}
				best = cost
				if use_super_gasp {
					f = super_gasp(f, d, off, &cost)
					if cost.cubes >= best.cubes {
						break
					}
				} else {
					f = last_gasp(f, d, off, &cost)
				}
				if !(cost.cubes < best.cubes || (cost.cubes == best.cubes && cost.total < best.total)) {
					break
				}
			}
			f = sf_append(f, e)
			if trace {
				size_stamp(f, "ADJUST     ")
			}
		}
		if !skip_make_sparse {
			f = make_sparse(f, dontCare, off)
		}
		if original.count < f.count {
			f = original
			unwrap_onset = false
			continue
		}
		return f
	}
}
