package bsat

// sat_solver_clause_new expects at least two normalized literals. Binary input
// clauses use literal-tagged watches without allocating a clause; learned binary
// clauses still have storage for activity/proof identity, as in ABC.
func sat_solver_clause_new(s *sat_solver, literals []lit, learnt int32) cla {
	if len(literals) < 2 || learnt < 0 || learnt > 1 {
		panic("invalid new clause")
	}
	for _, l := range literals {
		if lit_check(l, s.size) == 0 {
			panic("clause literal outside solver")
		}
	}
	size := int32(len(literals))
	if size == 2 && learnt == 0 {
		veci_push(&s.wlists[lit_neg(literals[0])], clause_from_lit(literals[1]))
		veci_push(&s.wlists[lit_neg(literals[1])], clause_from_lit(literals[0]))
		s.stats.clauses++
		s.stats.clauses_literals += int64(size)
		return 0
	}
	h := Sat_MemAppend(&s.Mem, literals, size, learnt, 0)
	if learnt != 0 {
		if s.hLearnts == -1 {
			s.hLearnts = h
		}
		c := Sat_MemClauseHand(&s.Mem, h)
		c.lbd = satClauseLBD(&s.satAssignments, c)
		if clause_id(c) != s.act_clas.size {
			panic("learned activity ID mismatch")
		}
		activity := int32(s.cla_inc)
		if s.ClaActType == 0 {
			activity = 1 << 10
		}
		veci_push(&s.act_clas, activity)
		s.stats.learnts++
		s.stats.learnts_literals += int64(size)
	} else {
		s.stats.clauses++
		s.stats.clauses_literals += int64(size)
	}
	left, right := h, h
	if size == 2 {
		left = clause_from_lit(literals[1])
		right = clause_from_lit(literals[0])
	}
	veci_push(&s.wlists[lit_neg(literals[0])], left)
	veci_push(&s.wlists[lit_neg(literals[1])], right)
	return h
}
