package bsat

// NewSolver3 shares the source-identical CDCL implementation. Solver3's original
// separate minimizer and type-1-only priority function remain available below.
func NewSolver3() *Solver { return sat_solver3_new() }
func sat_solver3_set_var_activity(s *sat_solver3, variables []int32, n int32) {
	if s.VarActType != 1 {
		panic("solver3 priority requires double activity")
	}
	for i := int32(0); i < n; i++ {
		v := i
		if variables != nil {
			v = variables[i]
		}
		if s.orderpos[v] < 0 {
			panic("priority variable not in decision heap")
		}
	}
	sat_solver_set_var_activity(s, variables, n)
}

// This is the solver3 divide-and-conquer minimizer, not solver1's extra-probing
// variant. Each base case uses an absolute conflict limit and restores it.
func sat_solver3_minimize_assumptions(s *sat_solver3, literals []lit, n, budget int32) int32 {
	if n <= 0 || int(n) > len(literals) {
		panic("invalid solver3 minimization input")
	}
	literals = literals[:n]
	if n == 1 {
		old := s.nConfLimit
		s.nConfLimit = int64(budget)
		status := sat_solver3_solve_internal(s)
		s.nConfLimit = old
		if status == l_False {
			return 0
		}
		return 1
	}
	left, right := n/2, n-n/2
	for i := int32(0); i < left; i++ {
		if sat_solver3_push(s, literals[i]) == 0 {
			for k := int32(0); k <= i; k++ {
				sat_solver3_pop(s)
			}
			return sat_solver3_minimize_assumptions(s, literals, i+1, budget)
		}
	}
	resultLeft := sat_solver3_minimize_assumptions(s, literals[left:], right, budget)
	for i := int32(0); i < left; i++ {
		sat_solver3_pop(s)
	}
	saved := append([]lit(nil), literals[:left]...)
	copy(literals, literals[left:left+resultLeft])
	copy(literals[resultLeft:], saved)
	for i := int32(0); i < resultLeft; i++ {
		if sat_solver3_push(s, literals[i]) == 0 {
			for k := int32(0); k <= i; k++ {
				sat_solver3_pop(s)
			}
			return sat_solver3_minimize_assumptions(s, literals, i+1, budget)
		}
	}
	resultRight := sat_solver3_minimize_assumptions(s, literals[resultLeft:], left, budget)
	for i := int32(0); i < resultLeft; i++ {
		sat_solver3_pop(s)
	}
	return resultLeft + resultRight
}
