package glucose2

// This file adapts the native Go solver to ABC's compact var*2+sign literal
// convention. It keeps the public Go boundary slice-based instead of carrying
// C pointer-plus-length pairs into callers.

// NewABCSolver creates an incremental solver with ABC-compatible literals.
func NewABCSolver() *SimpSolver {
	s := NewSimpSolver()
	s.SetIncrementalMode()
	return s
}

// EnsureABCVariables grows s until it contains n variables.
func (s *SimpSolver) EnsureABCVariables(n int) {
	if n < 0 {
		panic("glucose2: negative variable count")
	}
	for s.NVars() < n {
		s.NewVar(true, true)
	}
}

func abcLiterals(encoded []int32) []Lit {
	result := make([]Lit, len(encoded))
	for i, value := range encoded {
		if value < 0 {
			panic("glucose2: negative ABC literal")
		}
		result[i] = Lit(value)
	}
	return result
}

// AddABCClause adds a clause whose literals use ABC's var*2+sign encoding.
// It returns false if the clause makes the formula unsatisfiable.
func (s *SimpSolver) AddABCClause(encoded []int32) bool {
	maxVar := -1
	for _, value := range encoded {
		if value < 0 {
			panic("glucose2: negative ABC literal")
		}
		if int(value>>1) > maxVar {
			maxVar = int(value >> 1)
		}
	}
	s.EnsureABCVariables(maxVar + 1)
	return s.AddClause(abcLiterals(encoded))
}

// SolveABCLiterals solves under ABC-encoded assumptions and returns 1 for
// satisfiable, -1 for unsatisfiable, and 0 when a limit interrupts search.
func (s *SimpSolver) SolveABCLiterals(encoded []int32) int32 {
	maxVar := -1
	for _, value := range encoded {
		if value < 0 {
			panic("glucose2: negative ABC literal")
		}
		if int(value>>1) > maxVar {
			maxVar = int(value >> 1)
		}
	}
	s.EnsureABCVariables(maxVar + 1)
	switch s.SolveLimited(abcLiterals(encoded), false) {
	case LTrue:
		return GLUCOSE_SAT
	case LFalse:
		return GLUCOSE_UNSAT
	default:
		return GLUCOSE_UNDEC
	}
}

// ABCConflict returns an owned copy of the last failed-assumption core.
func (s *SimpSolver) ABCConflict() []int32 {
	result := make([]int32, len(s.Conflict))
	for i, p := range s.Conflict {
		result[i] = int32(p)
	}
	return result
}

// AddABCAnd encodes output = (a xor complA) AND (b xor complB), then xor
// complOut. The three clauses are the standard Tseitin encoding used by ABC.
func (s *SimpSolver) AddABCAnd(output, a, b int32, complA, complB, complOut bool) bool {
	for _, v := range []int32{output, a, b} {
		if v < 0 {
			panic("glucose2: negative ABC variable")
		}
	}
	s.EnsureABCVariables(int(max3(output, a, b)) + 1)
	return s.AddABCClause([]int32{
		abcLit(output, !complOut), abcLit(a, complA),
	}) && s.AddABCClause([]int32{
		abcLit(output, !complOut), abcLit(b, complB),
	}) && s.AddABCClause([]int32{
		abcLit(output, complOut), abcLit(a, !complA), abcLit(b, !complB),
	})
}

// AddABCXor encodes output = (a xor b) xor complement using four clauses.
func (s *SimpSolver) AddABCXor(output, a, b int32, complement bool) bool {
	for _, v := range []int32{output, a, b} {
		if v < 0 {
			panic("glucose2: negative ABC variable")
		}
	}
	s.EnsureABCVariables(int(max3(output, a, b)) + 1)
	for av := int32(0); av <= 1; av++ {
		for bv := int32(0); bv <= 1; bv++ {
			want := av ^ bv
			if complement {
				want ^= 1
			}
			// For the one forbidden output value, this clause excludes exactly
			// the current assignment to (output, a, b).
			badOutput := want ^ 1
			clause := []int32{abcLit(output, badOutput != 0), abcLit(a, av == 0), abcLit(b, bv == 0)}
			if !s.AddABCClause(clause) {
				return false
			}
		}
	}
	return true
}

func abcLit(variable int32, negative bool) int32 {
	lit := variable << 1
	if negative {
		lit++
	}
	return lit
}

func max3(a, b, c int32) int32 {
	if b > a {
		a = b
	}
	if c > a {
		a = c
	}
	return a
}

// MinimizeABCAssumptions returns an unsatisfiable subset of assumptions while
// retaining the first fixedPrefix assumptions. It performs deterministic
// chunk deletion and does not mutate the caller's slice. A satisfiable or
// interrupted input returns a copy of the original assumptions.
func (s *SimpSolver) MinimizeABCAssumptions(encoded []int32, fixedPrefix int) []int32 {
	if fixedPrefix < 0 || fixedPrefix > len(encoded) {
		panic("glucose2: invalid fixed assumption prefix")
	}
	original := append([]int32(nil), encoded...)
	if s.SolveABCLiterals(original) != GLUCOSE_UNSAT {
		return original
	}
	fixed := append([]int32(nil), original[:fixedPrefix]...)
	candidates := append([]int32(nil), original[fixedPrefix:]...)
	for len(candidates) > 1 {
		granularity := 2
		for granularity <= len(candidates) {
			chunk := (len(candidates) + granularity - 1) / granularity
			removed := false
			for start := 0; start < len(candidates); start += chunk {
				end := start + chunk
				if end > len(candidates) {
					end = len(candidates)
				}
				trial := make([]int32, 0, len(fixed)+len(candidates)-(end-start))
				trial = append(trial, fixed...)
				trial = append(trial, candidates[:start]...)
				trial = append(trial, candidates[end:]...)
				if s.SolveABCLiterals(trial) == GLUCOSE_UNSAT {
					candidates = append(candidates[:start], candidates[end:]...)
					if granularity > 2 {
						granularity--
					}
					removed = true
					break
				}
			}
			if !removed {
				if granularity == len(candidates) {
					break
				}
				granularity *= 2
				if granularity > len(candidates) {
					granularity = len(candidates)
				}
			}
		}
		break
	}
	if len(candidates) == 1 && s.SolveABCLiterals(fixed) == GLUCOSE_UNSAT {
		candidates = nil
	}
	return append(fixed, candidates...)
}
