package bsat

// Activity hooks preserve solver-specific numeric activity representations while
// keeping conflict analysis independent of the solver's allocation machinery.
type satActivityBumper interface {
	bumpVariable(int32)
	bumpClause(*clause)
}

func (a *satAssignments) tagVariable(v, tag int32) {
	if tag <= 0 || tag >= 16 {
		panic("invalid analysis tag")
	}
	if a.tags[v] == 0 {
		veci_push(&a.tagged, v)
	}
	a.tags[v] = tag
}
func (a *satAssignments) clearTags(start int32) {
	for _, v := range veci_begin(&a.tagged)[start:] {
		a.tags[v] = 0
	}
	veci_resize(&a.tagged, start)
}
func satLiteralRemovable(a *satAssignments, m *Sat_Mem_t, x int32, minLevels uint32) bool {
	top := a.tagged.size
	stack := []int32{x}
	for len(stack) > 0 {
		v := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		reason := a.reasons[v]
		if reason == 0 {
			panic("minimization without reason")
		}
		var literals []lit
		if clause_is_lit(reason) != 0 {
			literals = []lit{clause_read_lit(reason)}
		} else {
			literals = clause_begin(Sat_MemClauseHand(m, reason))[1:]
		}
		for _, l := range literals {
			u := lit_var(l)
			if a.tags[u] != 0 || a.levels[u] == 0 {
				continue
			}
			if a.reasons[u] == 0 || minLevels&(uint32(1)<<uint(a.levels[u]&31)) == 0 {
				a.clearTags(top)
				return false
			}
			stack = append(stack, u)
			a.tagVariable(u, 1)
		}
	}
	return true
}
func satAnalyze(a *satAssignments, m *Sat_Mem_t, h cla, hooks satActivityBumper, stats *stats_t) []lit {
	if a.tagged.size != 0 {
		panic("analysis with stale tags")
	}
	learnt := []lit{lit_Undef}
	pending := int32(0)
	p := lit_Undef
	index := a.qtail - 1
	for {
		if h == 0 {
			panic("missing conflict reason")
		}
		var literals []lit
		if clause_is_lit(h) != 0 {
			literals = []lit{clause_read_lit(h)}
		} else {
			c := Sat_MemClauseHand(m, h)
			if c.lrn != 0 && hooks != nil {
				hooks.bumpClause(c)
			}
			literals = clause_begin(c)
			if p != lit_Undef {
				literals = literals[1:]
			}
		}
		for _, l := range literals {
			v := lit_var(l)
			if a.tags[v] != 0 || a.levels[v] == 0 {
				continue
			}
			a.tagVariable(v, 1)
			if hooks != nil {
				hooks.bumpVariable(v)
			}
			if a.levels[v] == a.trail_lim.size {
				pending++
			} else {
				learnt = append(learnt, l)
			}
		}
		for index >= 0 && a.tags[lit_var(a.trail[index])] == 0 {
			index--
		}
		if index < 0 || pending <= 0 {
			panic("conflict does not reach current decision level")
		}
		p = a.trail[index]
		index--
		h = a.reasons[lit_var(p)]
		pending--
		if pending == 0 {
			break
		}
	}
	learnt[0] = lit_neg(p)
	var levels uint32
	for _, l := range learnt[1:] {
		levels |= uint32(1) << uint(a.levels[lit_var(l)]&31)
	}
	write := 1
	for _, l := range learnt[1:] {
		v := lit_var(l)
		if a.reasons[v] == 0 || !satLiteralRemovable(a, m, v, levels) {
			learnt[write] = l
			write++
		}
	}
	learnt = learnt[:write]
	stats.tot_literals += int64(write)
	a.clearTags(0)
	if len(learnt) > 1 {
		best := 1
		for i := 2; i < len(learnt); i++ {
			if a.levels[lit_var(learnt[i])] > a.levels[lit_var(learnt[best])] {
				best = i
			}
		}
		learnt[1], learnt[best] = learnt[best], learnt[1]
	}
	return learnt
}
