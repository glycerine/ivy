package bsat

// satAnalyzeFinal traces a conflict backwards to the decision literals supplied
// as assumptions. It does not change assignments or mutate the parent clauses.
func satAnalyzeFinal(a *satAssignments, m *Sat_Mem_t, conflict cla, rootLevel int32, skipFirst bool) []lit {
	if rootLevel == 0 {
		return nil
	}
	if a.tagged.size != 0 || a.trail_lim.size == 0 {
		panic("invalid final analysis state")
	}
	defer a.clearTags(0)
	literals := clause_begin(Sat_MemClauseHand(m, conflict))
	if skipFirst {
		if len(literals) == 0 {
			panic("skip first in empty conflict")
		}
		literals = literals[1:]
	}
	for _, l := range literals {
		v := lit_var(l)
		if a.levels[v] > 0 {
			a.tagVariable(v, 1)
		}
	}
	start := a.qtail - 1
	if rootLevel < a.trail_lim.size {
		start = a.trail_lim.ptr[rootLevel]
	}
	var result []lit
	for i := start; i >= a.trail_lim.ptr[0]; i-- {
		x := lit_var(a.trail[i])
		if a.tags[x] == 0 {
			continue
		}
		reason := a.reasons[x]
		if reason == 0 {
			if a.levels[x] <= 0 {
				panic("assumption at root level")
			}
			result = append(result, lit_neg(a.trail[i]))
			continue
		}
		var parents []lit
		if clause_is_lit(reason) != 0 {
			parents = []lit{clause_read_lit(reason)}
		} else {
			parents = clause_begin(Sat_MemClauseHand(m, reason))[1:]
		}
		for _, l := range parents {
			v := lit_var(l)
			if a.levels[v] > 0 {
				a.tagVariable(v, 1)
			}
		}
	}
	return result
}
