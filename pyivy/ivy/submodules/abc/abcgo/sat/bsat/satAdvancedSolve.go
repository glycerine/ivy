package bsat

func sat_solver_solve_lexsat(s *sat_solver, literals []lit, n int32) int32 {
	if n <= 0 || int(n) > len(literals) {
		panic("invalid lex-SAT literals")
	}
	literals = literals[:n]
	sat_solver_set_literal_polarity(s, literals, n)
	status := sat_solver_solve_internal(s)
	if status != l_True {
		return status
	}
	mismatch := int32(0)
	for mismatch < n && literals[mismatch] == sat_solver_var_literal(s, lit_var(literals[mismatch])) {
		mismatch++
	}
	if mismatch == n {
		return l_True
	}
	pushed := int32(0)
	failed := false
	for pushed <= mismatch {
		ok := sat_solver_push(s, literals[pushed])
		pushed++
		if ok == 0 {
			failed = true
			break
		}
	}
	if failed {
		status = l_False
	} else {
		status = sat_solver_solve_internal(s)
	}
	if status == l_True {
		if mismatch+1 < n {
			status = sat_solver_solve_lexsat(s, literals[mismatch+1:], n-mismatch-1)
		}
	} else if status == l_False {
		if lit_sign(literals[mismatch]) == 0 {
			panic("lex-SAT lower-bound precondition violated")
		}
		sat_solver_pop(s)
		pushed--
		literals[mismatch] = lit_neg(literals[mismatch])
		if sat_solver_push(s, literals[mismatch]) == 0 {
			panic("lex-SAT expected satisfying branch")
		}
		pushed++
		for i := mismatch + 1; i < n; i++ {
			literals[i] = lit_neg(toLit(lit_var(literals[i])))
		}
		if mismatch+1 < n {
			status = sat_solver_solve_lexsat(s, literals[mismatch+1:], n-mismatch-1)
		} else {
			status = l_True
		}
	}
	for i := int32(0); i < pushed; i++ {
		sat_solver_pop(s)
	}
	return status
}
func sat_solver_minimize_assumptions(s *sat_solver, literals []lit, n int32, budget int32) int32 {
	if n <= 0 || int(n) > len(literals) {
		panic("invalid minimization literals")
	}
	literals = literals[:n]
	if n == 1 {
		if budget != 0 {
			s.nConfLimit = s.stats.conflicts + int64(budget)
		}
		if sat_solver_solve_internal(s) == l_False {
			return 0
		}
		return 1
	}
	left, right := n/2, n-n/2
	for i := int32(0); i < left; i++ {
		if sat_solver_push(s, literals[i]) == 0 {
			for k := int32(0); k <= i; k++ {
				sat_solver_pop(s)
			}
			return sat_solver_minimize_assumptions(s, literals, i+1, budget)
		}
	}
	if budget != 0 {
		s.nConfLimit = s.stats.conflicts + int64(budget)
	}
	if sat_solver_solve_internal(s) == l_False {
		for i := int32(0); i < left; i++ {
			sat_solver_pop(s)
		}
		return sat_solver_minimize_assumptions(s, literals, left, budget)
	}
	resultLeft := int32(1)
	if right != 1 {
		resultLeft = sat_solver_minimize_assumptions(s, literals[left:], right, budget)
	}
	for i := int32(0); i < left; i++ {
		sat_solver_pop(s)
	}
	saved := append([]lit(nil), literals[:left]...)
	copy(literals, literals[left:left+resultLeft])
	copy(literals[resultLeft:], saved)
	for i := int32(0); i < resultLeft; i++ {
		if sat_solver_push(s, literals[i]) == 0 {
			for k := int32(0); k <= i; k++ {
				sat_solver_pop(s)
			}
			return sat_solver_minimize_assumptions(s, literals, i+1, budget)
		}
	}
	if budget != 0 {
		s.nConfLimit = s.stats.conflicts + int64(budget)
	}
	if sat_solver_solve_internal(s) == l_False {
		for i := int32(0); i < resultLeft; i++ {
			sat_solver_pop(s)
		}
		return resultLeft
	}
	resultRight := int32(1)
	if left != 1 {
		resultRight = sat_solver_minimize_assumptions(s, literals[resultLeft:], left, budget)
	}
	for i := int32(0); i < resultLeft; i++ {
		sat_solver_pop(s)
	}
	return resultLeft + resultRight
}
func sat_solver_minimize_assumptions2(s *sat_solver, literals []lit, n int32, budget int32) int32 {
	if n <= 0 || int(n) > len(literals) {
		panic("invalid specialized minimization literals")
	}
	literals = literals[:n]
	if n == 1 {
		opposite := lit_neg(literals[0])
		old := s.nConfLimit
		s.nConfLimit = int64(budget)
		if sat_solver_push(s, opposite) == 0 {
			sat_solver_pop(s)
			s.nConfLimit = old
			return 1
		}
		status := sat_solver_solve_internal(s)
		sat_solver_pop(s)
		if status == l_False && sat_solver_addclause(s, []lit{opposite}) == 0 {
			panic("failed implied unit")
		}
		s.nConfLimit = old
		if status == l_False {
			return 0
		}
		return 1
	}
	left, right := n/2, n-n/2
	for i := int32(0); i < left; i++ {
		if sat_solver_push(s, literals[i]) == 0 {
			for k := int32(0); k <= i; k++ {
				sat_solver_pop(s)
			}
			return sat_solver_minimize_assumptions2(s, literals, i+1, budget)
		}
	}
	resultLeft := sat_solver_minimize_assumptions2(s, literals[left:], right, budget)
	for i := int32(0); i < left; i++ {
		sat_solver_pop(s)
	}
	saved := append([]lit(nil), literals[:left]...)
	copy(literals, literals[left:left+resultLeft])
	copy(literals[resultLeft:], saved)
	for i := int32(0); i < resultLeft; i++ {
		if sat_solver_push(s, literals[i]) == 0 {
			for k := int32(0); k <= i; k++ {
				sat_solver_pop(s)
			}
			return sat_solver_minimize_assumptions2(s, literals, i+1, budget)
		}
	}
	resultRight := sat_solver_minimize_assumptions2(s, literals[resultLeft:], left, budget)
	for i := int32(0); i < resultLeft; i++ {
		sat_solver_pop(s)
	}
	return resultLeft + resultRight
}
