// Go slice adaptations of the declarations in src/sat/psat/m114p.h.
package psat

func M114p_SolverNew(recordProof int) M114p_Solver_t { return New(recordProof != 0) }
func M114p_SolverDelete(s M114p_Solver_t) {
	if s != nil {
		*s = Solver{deleted: true}
	}
}
func M114p_SolverPrintStats(s M114p_Solver_t, seconds float64) { s.PrintStats(seconds) }
func M114p_SolverSetVarNum(s M114p_Solver_t, n int)            { s.SetVariableCount(n) }

// A C [begin,end) pointer range becomes a slice and its end index in Go.
func M114p_SolverAddClause(s M114p_Solver_t, begin []lit, end int) int {
	return boolInt(s.AddClause(begin[:end]))
}
func M114p_SolverSimplify(s M114p_Solver_t) int { return boolInt(s.Simplify()) }
func M114p_SolverSolve(s M114p_Solver_t, begin []lit, end, conflictLimit int) int {
	return s.Solve(begin[:end], conflictLimit)
}
func M114p_SolverGetConflictNum(s M114p_Solver_t) int { return s.ConflictCount() }
func M114p_SolverProofIsReady(s M114p_Solver_t) int   { return boolInt(s.ProofReady()) }
func M114p_SolverProofSave(s M114p_Solver_t, path string) {
	if err := s.SaveProof(path); err != nil {
		panic(err)
	}
}
func M114p_SolverProofClauseNum(s M114p_Solver_t) int {
	if !s.ProofReady() {
		return 0
	}
	return len(s.proof.Roots) + len(s.proof.Chains)
}
func M114p_SolverGetFirstRoot(s M114p_Solver_t, out *[]int) int {
	s.rootCursor = 0
	return M114p_SolverGetNextRoot(s, out)
}
func M114p_SolverGetNextRoot(s M114p_Solver_t, out *[]int) int {
	if !s.ProofReady() || s.rootCursor >= len(s.proof.Roots) {
		*out = nil
		return 0
	}
	*out = append([]int(nil), s.proof.Roots[s.rootCursor]...)
	s.rootCursor++
	return len(*out)
}
func M114p_SolverGetFirstChain(s M114p_Solver_t, clauses, variables *[]int) int {
	s.chainCursor = 0
	return M114p_SolverGetNextChain(s, clauses, variables)
}
func M114p_SolverGetNextChain(s M114p_Solver_t, clauses, variables *[]int) int {
	if !s.ProofReady() || s.chainCursor >= len(s.proof.Chains) {
		*clauses = nil
		*variables = nil
		return 0
	}
	chain := s.proof.Chains[s.chainCursor]
	s.chainCursor++
	*clauses = append([]int(nil), chain.Clauses...)
	*variables = append([]int(nil), chain.Pivots...)
	return len(*variables)
}
func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
