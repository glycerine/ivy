package bsat

func solver2_tag(s *sat_solver2, v, tag int32) {
	if s.tags[v] == 0 {
		veci_push(&s.tagged, v)
	}
	s.tags[v] |= tag
}
func solver2_lit_removable_rec(s *sat_solver2, v int32) bool {
	if s.tags[v]&2 != 0 {
		return s.tags[v]&4 != 0
	}
	c := clause2_read(s, s.reasons[v])
	if c == nil {
		solver2_tag(s, v, 2)
		return false
	}
	for _, l := range c.lits[1:c.size] {
		x := lit_var(l)
		if s.tags[x]&1 != 0 {
			solver2_lit_removable_rec(s, x)
			continue
		}
		if s.levels[x] == 0 || s.tags[x] == 6 {
			continue
		}
		if s.tags[x] == 2 || !s.markedLevels[s.levels[x]] || !solver2_lit_removable_rec(s, x) {
			solver2_tag(s, v, 2)
			return false
		}
	}
	if s.fProofLogging != 0 && s.tags[v]&1 != 0 {
		veci_push(&s.min_lit_order, v)
	}
	solver2_tag(s, v, 6)
	return true
}
func solver2_logging_order_rec(s *sat_solver2, v int32) {
	if s.tags[v]&8 != 0 {
		return
	}
	c := clause2_read(s, s.reasons[v])
	if c == nil {
		panic("missing minimization proof reason")
	}
	for _, l := range c.lits[1:c.size] {
		x := lit_var(l)
		if s.levels[x] != 0 && s.tags[x]&1 == 0 {
			solver2_logging_order_rec(s, x)
		}
	}
	solver2_tag(s, v, 8)
	veci_push(&s.min_step_order, v)
}
func solver2_analyze(s *sat_solver2, c *clause, learnt *veci) int32 {
	if s.tagged.size != 0 {
		panic("stale solver2 conflict tags")
	}
	proof_chain_start(s, c)
	veci_push(learnt, lit_Undef)
	pending := int32(0)
	index := s.qtail - 1
	var pivot lit
	skip := int32(0)
	for {
		if c == nil {
			panic("missing first-UIP reason")
		}
		if c.lrn != 0 {
			act_clause2_bump(s, c)
		}
		for _, l := range c.lits[skip:c.size] {
			v := lit_var(l)
			if s.tags[v] != 0 {
				continue
			}
			if s.levels[v] == 0 {
				proof_chain_resolve(s, nil, v)
				continue
			}
			solver2_tag(s, v, 1)
			act_var2_bump(s, v)
			if s.levels[v] == solver2_dlevel(s) {
				pending++
			} else {
				veci_push(learnt, l)
			}
		}
		for index >= 0 && s.tags[lit_var(s.trail[index])] == 0 {
			index--
		}
		if index < 0 {
			panic("conflict has no current-level pivot")
		}
		pivot = s.trail[index]
		index--
		c = clause2_read(s, s.reasons[lit_var(pivot)])
		pending--
		if pending == 0 {
			break
		}
		proof_chain_resolve(s, c, lit_var(pivot))
		skip = 1
	}
	learnt.ptr[0] = lit_neg(pivot)
	for _, l := range learnt.ptr[1:learnt.size] {
		level := s.levels[lit_var(l)]
		if int(level) >= len(s.markedLevels) {
			s.markedLevels = satResize(s.markedLevels, int(level)+1, 2*int(level)+1)
		}
		if !s.markedLevels[level] {
			s.markedLevels[level] = true
			veci_push(&s.mark_levels, level)
		}
	}
	veci_resize(&s.min_lit_order, 0)
	kept := int32(1)
	for _, l := range learnt.ptr[1:learnt.size] {
		if !solver2_lit_removable_rec(s, lit_var(l)) {
			learnt.ptr[kept] = l
			kept++
		}
	}
	proof := int32(0)
	if s.fProofLogging != 0 {
		veci_resize(&s.min_step_order, 0)
		for _, v := range veci_begin(&s.min_lit_order) {
			solver2_logging_order_rec(s, v)
		}
		for i := s.min_step_order.size - 1; i >= 0; i-- {
			v := s.min_step_order.ptr[i]
			reason := clause2_read(s, s.reasons[v])
			proof_chain_resolve(s, reason, v)
			for _, l := range reason.lits[1:reason.size] {
				if s.levels[lit_var(l)] == 0 {
					proof_chain_resolve(s, nil, lit_var(l))
				}
			}
		}
		proof = proof_chain_stop(s)
	}
	for _, level := range veci_begin(&s.mark_levels) {
		s.markedLevels[level] = false
	}
	veci_resize(&s.mark_levels, 0)
	veci_resize(learnt, kept)
	s.stats.tot_literals += int64(kept)
	for _, v := range veci_begin(&s.tagged) {
		s.tags[v] = 0
	}
	veci_resize(&s.tagged, 0)
	if kept > 1 {
		highest := int32(1)
		for i := int32(2); i < kept; i++ {
			if s.levels[lit_var(learnt.ptr[i])] > s.levels[lit_var(learnt.ptr[highest])] {
				highest = i
			}
		}
		learnt.ptr[1], learnt.ptr[highest] = learnt.ptr[highest], learnt.ptr[1]
	}
	return proof
}
