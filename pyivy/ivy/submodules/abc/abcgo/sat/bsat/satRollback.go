package bsat

// Rollback uses the original ABC bookmark contract: added problem clauses refer
// to newly introduced variables. Tagged binary watches carry no timestamp.
func sat_solver_rollback(s *sat_solver) {
	m := &s.Mem
	if s.iVarPivot < 0 || s.iVarPivot > s.size || s.iTrailPivot < 0 || s.iTrailPivot > s.qtail {
		panic("invalid solver bookmark")
	}
	sat_solver_canceluntil_rollback(s, s.iTrailPivot)
	if s.iVarPivot < s.size {
		if s.activity2 != nil {
			s.var_inc = s.var_inc2
			copy(s.activity, s.activity2[:s.iVarPivot])
		}
		veci_resize(&s.order, 0)
		for v := int32(0); v < s.iVarPivot; v++ {
			if s.assigns[v] != varX {
				continue
			}
			s.orderpos[v] = s.order.size
			veci_push(&s.order, v)
			order_update(s, v)
		}
	}
	for i := int32(0); i < s.iVarPivot*2; i++ {
		w := &s.wlists[i]
		write := int32(0)
		for _, h := range veci_begin(w) {
			keep := false
			if clause_is_lit(h) != 0 {
				keep = clause_read_lit(h) < s.iVarPivot*2
			} else {
				keep = Sat_MemClauseUsed(m, h) != 0
			}
			if keep {
				w.ptr[write] = h
				write++
			}
		}
		veci_resize(w, write)
	}
	for i := s.iVarPivot * 2; i < s.size*2; i++ {
		s.wlists[i].size = 0
	}
	s.stats.clauses = uint32(m.BookMarkE[0])
	s.stats.learnts = uint32(m.BookMarkE[1])
	Sat_MemRollBack(m)
	veci_resize(&s.act_clas, int32(s.stats.learnts))
	s.size = s.iVarPivot
	if s.size == 0 {
		s.qhead = 0
		s.qtail = 0
		solver_init_activities(s)
		s.root_level = 0
		s.random_seed = 91648253
		s.progress_estimate = 0
		s.verbosity = 0
		s.stats = stats_t{}
		s.iVarPivot = 0
		s.iTrailPivot = 0
		s.hProofPivot = 1
	}
}
