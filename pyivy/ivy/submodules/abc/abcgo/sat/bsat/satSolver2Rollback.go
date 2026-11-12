package bsat

func sat_solver2_bookmark(s *sat_solver2) {
	if s.qhead != s.qtail {
		panic("solver2 bookmark before propagation")
	}
	s.iVarPivot, s.iTrailPivot = s.size, s.qhead
	if s.pPrf1 != nil {
		s.hProofPivot = s.pPrf1.next
	}
	Sat_MemBookMark(&s.Mem)
	s.var_inc2 = s.var_inc
	copy(s.activity2, s.activity)
}
func sat_solver2_rollback(s *sat_solver2) {
	if s.iVarPivot < 0 || s.iVarPivot > s.size || s.iTrailPivot > s.qtail {
		panic("invalid solver2 bookmark")
	}
	if s.pPrf2 != nil {
		panic("solver2 bitset proofs do not support rollback")
	}
	s.fSolved = 0
	solver2_canceluntil_rollback(s, s.iTrailPivot)
	if s.iVarPivot < s.size {
		s.var_inc = s.var_inc2
		copy(s.activity[:s.iVarPivot], s.activity2[:s.iVarPivot])
		veci_resize(&s.order, 0)
		for v := int32(0); v < s.iVarPivot; v++ {
			s.orderpos[v] = -1
			if s.assigns[v] != satUnassigned {
				continue
			}
			s.orderpos[v] = s.order.size
			veci_push(&s.order, v)
			order_update(&s.sat_solver_t, v)
		}
	}
	for i := int32(0); i < 2*s.iVarPivot; i++ {
		w := &s.wlists[i]
		n := int32(0)
		for _, h := range veci_begin(w) {
			if Sat_MemClauseUsed(&s.Mem, h) != 0 {
				w.ptr[n] = h
				n++
			}
		}
		veci_resize(w, n)
	}
	for i := 2 * s.iVarPivot; i < 2*s.size; i++ {
		veci_resize(&s.wlists[i], 0)
	}
	s.stats.clauses, s.stats.learnts = uint32(s.Mem.BookMarkE[0]), uint32(s.Mem.BookMarkE[1])
	Sat_MemRollBack(&s.Mem)
	veci_resize(&s.act_clas, int32(s.stats.learnts))
	if s.claProofs.size != 0 {
		veci_resize(&s.claProofs, int32(s.stats.learnts))
	}
	if s.pPrf1 != nil {
		g := s.pPrf1
		pivot := s.hProofPivot
		for len(g.order) > 0 && g.order[len(g.order)-1] >= pivot {
			h := g.order[len(g.order)-1]
			delete(g.nodes, h)
			g.order = g.order[:len(g.order)-1]
		}
		g.next = max(pivot, 2)
	}
	s.size = s.iVarPivot
	// Slices retain capacity for subsequent growth, but live variable ranges must
	// shrink so common setnvars can reinitialize every reintroduced variable.
	if s.size == 0 {
		s.qhead, s.qtail, s.root_level = 0, 0, 0
		s.var_inc, s.cla_inc = 1<<5, 1<<11
		s.random_seed = 91648253
		s.progress_estimate = 0
		s.verbosity = 0
		s.stats = stats_t{}
		s.hLearntLast, s.hProofLast = -1, -1
		s.hProofPivot = 1
	}
}
