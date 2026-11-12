package bsat

func clause_read(s *sat_solver, h cla) *clause   { return Sat_MemClauseHand(&s.Mem, h) }
func sat_solver_flip_print_clause(s *sat_solver) { s.fPrintClause ^= 1 }
func sat_solver_act_var_clear(s *sat_solver) {
	switch s.VarActType {
	case 0:
		for i := range s.activity {
			s.activity[i] = 1 << 10
		}
		s.var_inc = 1 << 5
	case 1:
		clear(s.activity)
		s.var_inc = 1
	case 2:
		for i := range s.activity {
			s.activity[i] = Xdbl_Const1()
		}
		s.var_inc = Xdbl_Const1()
	default:
		panic("invalid activity type")
	}
}
func sat_solver_compress(s *sat_solver) {
	if s.qtail != s.qhead && sat_solver_simplify(s) == 0 {
		panic("compress conflicting solver")
	}
}
func sat_solver_delete_p(p **sat_solver) {
	if *p != nil {
		sat_solver_delete(*p)
	}
	*p = nil
}
func sat_solver_clean_polarity(s *sat_solver, vars []int32, n int32) {
	for _, v := range vars[:n] {
		s.polarity[v] = 0
	}
}
func sat_solver_set_polarity(s *sat_solver, vars []int32, n int32) {
	clear(s.polarity)
	for _, v := range vars[:n] {
		s.polarity[v] = 1
	}
}
func sat_solver_set_literal_polarity(s *sat_solver, literals []lit, n int32) {
	for _, l := range literals[:n] {
		s.polarity[lit_var(l)] = 1 - lit_sign(l)
	}
}
func sat_solver_final(s *sat_solver, out *[]lit) int32 {
	*out = veci_begin(&s.conf_final)
	return s.conf_final.size
}
func sat_solver_set_runtime_limit(s *sat_solver, limit abctime) abctime {
	old := s.nRuntimeLimit
	s.nRuntimeLimit = limit
	return old
}
func sat_solver_set_random(s *sat_solver, disabled int32) int32 {
	old := s.fNotUseRandom
	s.fNotUseRandom = disabled
	return old
}
func sat_solver_bookmark(s *sat_solver) {
	if s.qhead != s.qtail {
		panic("bookmark before propagation completes")
	}
	s.iVarPivot = s.size
	s.iTrailPivot = s.qhead
	Sat_MemBookMark(&s.Mem)
	if s.activity2 != nil {
		s.var_inc2 = s.var_inc
		copy(s.activity2, s.activity[:s.iVarPivot])
	}
}
func sat_solver_set_pivot_variables(s *sat_solver, vars []int32, n int32) {
	s.pivot_vars.ptr = append([]int32(nil), vars[:n]...)
	s.pivot_vars.size = n
	s.pivot_vars.cap = n
}
func sat_solver_count_usedvars(s *sat_solver) int32 {
	var count int32
	for i, value := range s.pFreqs {
		if value != 0 {
			count++
			s.pFreqs[i] = 0
		}
	}
	return count
}
func sat_solver_set_runid(s *sat_solver, id int32)                   { s.RunId = id }
func sat_solver_set_stop_func(s *sat_solver, stop func(int32) int32) { s.pFuncStop = stop }
