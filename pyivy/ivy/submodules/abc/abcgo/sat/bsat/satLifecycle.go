package bsat

// Default activity initialization has only the integer branch: the C routine
// itself sets both representation selectors to zero before choosing a branch.
func solver_init_activities(s *sat_solver) {
	s.VarActType = 0
	s.ClaActType = 0
	s.var_inc = 1 << 5
	s.var_decay = ^uint64(0)
	s.cla_inc = 1 << 11
	s.cla_decay = ^uint32(0)
}
func sat_solver_new() *sat_solver                   { return newSatSolver(17, 91648253) }
func zsat_solver_new_seed(seed float64) *sat_solver { return newSatSolver(15, seed) }
func newSatSolver(pageBits int32, seed float64) *sat_solver {
	s := new(sat_solver)
	s.satAssignments = newSatAssignments(0)
	Sat_MemAlloc_(&s.Mem, pageBits)
	s.hLearnts = -1
	s.hBinary = Sat_MemAppend(&s.Mem, nil, 2, 0, 0)
	s.binary = Sat_MemClauseHand(&s.Mem, s.hBinary)
	s.nLearntStart = LEARNT_MAX_START_DEFAULT
	s.nLearntDelta = LEARNT_MAX_INCRE_DEFAULT
	s.nLearntRatio = LEARNT_MAX_RATIO_DEFAULT
	s.nLearntMax = s.nLearntStart
	for _, v := range []*veci{&s.order, &s.trail_lim, &s.tagged, &s.act_clas, &s.stack, &s.unit_lits, &s.temp_clause, &s.conf_final, &s.act_vars} {
		veci_new(v)
	}
	solver_init_activities(s)
	s.random_seed = seed
	return s
}
func satResize[T any](old []T, n, capacity int) []T {
	if n <= cap(old) {
		return old[:n]
	}
	if capacity < n {
		capacity = n
	}
	next := make([]T, n, capacity)
	copy(next, old)
	return next
}
func sat_solver_setnvars(s *sat_solver, n int32) {
	if n < 0 || n > 0x40000000 {
		panic("invalid SAT variable count")
	}
	if n <= s.size {
		return
	}
	if s.cap < n {
		for s.cap < n {
			next := int64(s.cap)*2 + 1
			if next > 0x40000000 {
				next = 0x40000000
			}
			s.cap = int32(next)
		}
		if s.cap < 50000 {
			s.cap = 50000
		}
	}
	length, capacity := int(n), int(s.cap)
	s.wlists = satResize(s.wlists, 2*length, 2*capacity)
	s.levels = satResize(s.levels, length, capacity)
	s.assigns = satResize(s.assigns, length, capacity)
	s.polarity = satResize(s.polarity, length, capacity)
	s.tags = satResize(s.tags, length, capacity)
	s.loads = satResize(s.loads, length, capacity)
	s.activity = satResize(s.activity, length, capacity)
	s.activity2 = satResize(s.activity2, length, capacity)
	s.pFreqs = satResize(s.pFreqs, length, capacity)
	s.orderpos = satResize(s.orderpos, length, capacity)
	s.reasons = satResize(s.reasons, length, capacity)
	s.trail = satResize(s.trail, length, capacity)
	s.model = satResize(s.model, length, capacity)
	if s.factors != nil {
		s.factors = satResize(s.factors, length, capacity)
	}
	if s.pGlobalVars != nil && len(s.pGlobalVars) < length {
		oldLen := len(s.pGlobalVars)
		s.pGlobalVars = satResize(s.pGlobalVars, length, capacity)
		clear(s.pGlobalVars[oldLen:])
	}
	for v := s.size; v < n; v++ {
		for _, w := range []*veci{&s.wlists[2*v], &s.wlists[2*v+1]} {
			if w.size != 0 {
				panic("new variable has watchers")
			}
			if w.ptr == nil {
				veci_new(w)
			}
		}
		switch s.VarActType {
		case 0:
			s.activity[v] = 1 << 10
		case 1, 2:
			s.activity[v] = 0
		default:
			panic("invalid activity representation")
		}
		s.pFreqs[v] = 0
		if s.factors != nil {
			s.factors[v] = 0
		}
		s.levels[v] = 0
		s.assigns[v] = satUnassigned
		s.polarity[v] = 0
		s.tags[v] = 0
		s.loads[v] = 0
		s.orderpos[v] = s.order.size
		s.reasons[v] = 0
		s.model[v] = 0
		veci_push(&s.order, v)
		satOrderUpdate(veci_begin(&s.order), s.orderpos, s.activity, v)
	}
	s.size = n
}
func sat_solver_addvar(s *sat_solver) int32 { sat_solver_setnvars(s, s.size+1); return s.size - 1 }

// All owned buffers and proof-store chunks are managed by Go's collector.
func sat_solver_delete(s *sat_solver) {
	if s != nil {
		*s = sat_solver{}
	}
}
func sat_solver_restart(s *sat_solver) { zsat_solver_restart_seed(s, 91648253) }
func zsat_solver_restart_seed(s *sat_solver, seed float64) {
	s.ClearActivityFactors()
	Sat_MemRestart(&s.Mem)
	s.hLearnts = -1
	s.hBinary = Sat_MemAppend(&s.Mem, nil, 2, 0, 0)
	s.binary = Sat_MemClauseHand(&s.Mem, s.hBinary)
	veci_resize(&s.trail_lim, 0)
	veci_resize(&s.order, 0)
	for i := range s.wlists {
		s.wlists[i].size = 0
	}
	s.nDBreduces = 0
	s.size = 0
	s.qhead = 0
	s.qtail = 0
	solver_init_activities(s)
	veci_resize(&s.act_clas, 0)
	s.root_level = 0
	s.random_seed = seed
	s.progress_estimate = 0
	s.verbosity = 0
	s.stats = stats_t{}
}
