package bsat

func solver2_analyze_final(s *sat_solver2, conflict *clause, skip int32) int32 {
	veci_resize(&s.conf_final, 0)
	if s.root_level == 0 {
		return s.hProofLast
	}
	if s.tagged.size != 0 {
		panic("stale final-analysis tags")
	}
	proof_chain_start(s, conflict)
	for _, l := range conflict.lits[skip:conflict.size] {
		v := lit_var(l)
		if s.levels[v] != 0 {
			solver2_tag(s, v, 1)
		} else {
			proof_chain_resolve(s, nil, v)
		}
	}
	for i := s.qtail - 1; i >= s.trail_lim.ptr[0]; i-- {
		v := lit_var(s.trail[i])
		if s.tags[v] == 0 {
			continue
		}
		reason := clause2_read(s, s.reasons[v])
		if reason == nil {
			veci_push(&s.conf_final, lit_neg(s.trail[i]))
			continue
		}
		proof_chain_resolve(s, reason, v)
		for _, l := range reason.lits[1:reason.size] {
			x := lit_var(l)
			if s.levels[x] != 0 {
				solver2_tag(s, x, 1)
			} else {
				proof_chain_resolve(s, nil, x)
			}
		}
	}
	for _, v := range veci_begin(&s.tagged) {
		s.tags[v] = 0
	}
	veci_resize(&s.tagged, 0)
	return proof_chain_stop(s)
}
func solver2_record(s *sat_solver2, literals *veci, proof int32) {
	if literals.size == 0 {
		panic("empty solver2 asserting clause")
	}
	h := clause2_create_new(s, veci_begin(literals), 1, proof)
	if literals.size == 1 {
		if s.fProofLogging != 0 {
			s.units[lit_var(literals.ptr[0])] = h
		}
		h = 0
	}
	if solver2_enqueue(s, literals.ptr[0], h) == 0 {
		panic("solver2 learned clause not asserting")
	}
}
func solver2_search(s *sat_solver2, budget int64) lbool {
	frequency := float64(0.02)
	if s.fNotUseRandom != 0 {
		frequency = 0
	}
	s.stats.starts++
	var learned veci
	veci_new(&learned)
	conflicts := int64(0)
	for {
		conflict := solver2_propagate(s)
		if conflict != nil {
			s.stats.conflicts++
			conflicts++
			if solver2_dlevel(s) <= s.root_level {
				s.hProofLast = solver2_analyze_final(s, conflict, 0)
				return l_False
			}
			veci_resize(&learned, 0)
			proof := solver2_analyze(s, conflict, &learned)
			level := s.root_level
			if learned.size > 1 {
				level = max(level, s.levels[lit_var(learned.ptr[1])])
			}
			solver2_canceluntil(s, level)
			solver2_record(s, &learned, proof)
			if learned.size == 1 {
				s.levels[lit_var(learned.ptr[0])] = 0
			}
			act_var2_decay(s)
			act_clause2_decay(s)
			continue
		}
		exhausted := budget >= 0 && conflicts >= budget
		expired := s.stats.conflicts&63 == 0 && s.runtimeExpired()
		limited := s.nConfLimit != 0 && s.stats.conflicts > s.nConfLimit || s.nInsLimit != 0 && s.stats.propagations > s.nInsLimit
		if exhausted || expired || limited {
			s.progress_estimate = sat_solver_progress(&s.sat_solver_t)
			solver2_canceluntil(s, s.root_level)
			return l_Undef
		}
		s.stats.decisions++
		next := order_select(&s.sat_solver_t, frequency)
		if next == var_Undef {
			for v := int32(0); v < s.size; v++ {
				if s.assigns[v] == satUnassigned {
					panic("incomplete solver2 model")
				}
				if s.assigns[v] == 0 {
					s.model[v] = l_True
				} else {
					s.model[v] = l_False
				}
			}
			solver2_canceluntil(s, s.root_level)
			return l_True
		}
		solver2_assume(s, toLitCond(next, 1-s.polarity[next]))
	}
}
