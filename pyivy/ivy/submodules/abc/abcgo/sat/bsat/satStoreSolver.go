package bsat

func solverClauseStore(s *sat_solver) *Sto_Man_t {
	if s.pStore == nil {
		return nil
	}
	p, ok := s.pStore.(*Sto_Man_t)
	if !ok {
		panic("invalid solver proof store")
	}
	return p
}
func sat_solver_store_alloc(s *sat_solver) {
	if s.pStore != nil {
		panic("proof store already allocated")
	}
	s.pStore = Sto_ManAlloc()
}
func sat_solver_store_write(s *sat_solver, name string) {
	if p := solverClauseStore(s); p != nil {
		Sto_ManDumpClauses(p, name)
	}
}
func sat_solver_store_free(s *sat_solver) {
	if p := solverClauseStore(s); p != nil {
		Sto_ManFree(p)
	}
	s.pStore = nil
}
func sat_solver_store_change_last(s *sat_solver) int32 {
	if p := solverClauseStore(s); p != nil {
		return Sto_ManChangeLastClause(p)
	}
	return -1
}
func sat_solver_store_mark_roots(s *sat_solver) {
	if p := solverClauseStore(s); p != nil {
		Sto_ManMarkRoots(p)
	}
}
func sat_solver_store_mark_clauses_a(s *sat_solver) {
	if p := solverClauseStore(s); p != nil {
		Sto_ManMarkClausesA(p)
	}
}
func sat_solver_store_release(s *sat_solver) any { p := s.pStore; s.pStore = nil; return p }
