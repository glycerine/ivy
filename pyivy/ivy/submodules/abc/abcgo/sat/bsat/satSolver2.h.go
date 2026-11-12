package bsat

/**************************************************************************************************
MiniSat -- Copyright (c) 2005, Niklas Sorensson
http://www.cs.chalmers.se/Cs/Research/FormalMethods/MiniSat/

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and
associated documentation files (the "Software"), to deal in the Software without restriction,
including without limitation the rights to use, copy, modify, merge, publish, distribute,
sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or
substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT
NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT
OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
**************************************************************************************************/

// Solver2 retains explicit unit clauses and proof state. Common state has the
// same layout-independent semantics as the primary solver.
type sat_solver2 = sat_solver2_t
type sat_solver2_t struct {
	sat_solver_t
	units                                                                          []cla
	variableA                                                                      []int32
	markedLevels                                                                   []bool
	nUnits, nof_learnts, fProofLogging                                             int32
	claProofs, temp_proof, mark_levels, min_lit_order, min_step_order, learnt_live veci
	pPrf1                                                                          *ProofGraph
	pPrf2                                                                          *Prf_Man_t
	pInt2                                                                          solver2Interpolation
	hLearntLast, hProofLast, tempInter                                             int32
	dPrfMemory                                                                     float64
}
type solver2Interpolation interface {
	chainStart(*clause) int32
	chainResolve(*clause, int32, int32) int32
}

func clause2_read(s *sat_solver2, h cla) *clause {
	if h == 0 {
		return nil
	}
	return Sat_MemClauseHand(&s.Mem, h)
}
func clause2_proofid(s *sat_solver2, c *clause, varA int32) int32 {
	if c.lrn != 0 {
		return s.claProofs.ptr[clause_id(c)]<<2 | varA<<1
	}
	return clause_id(c)<<2 | varA<<1 | 1
}
func clause2_is_partA(s *sat_solver2, h cla) int32     { return clause2_read(s, h).partA }
func clause2_set_partA(s *sat_solver2, h cla, a int32) { clause2_read(s, h).partA = a }
func clause2_id(s *sat_solver2, h cla) int32           { return clause_id(clause2_read(s, h)) }
func clause2_set_id(s *sat_solver2, h cla, id int32)   { clause_set_id(clause2_read(s, h), id) }
func solver2Bool(value bool) int32 {
	if value {
		return 1
	}
	return 0
}
func var_is_assigned(s *sat_solver2, v int32) int32 {
	return solver2Bool(s.assigns[v] != satUnassigned)
}
func var_is_partA(s *sat_solver2, v int32) int32          { return s.variableA[v] }
func var_set_partA(s *sat_solver2, v, a int32)            { s.variableA[v] = a }
func sat_solver2_nvars(s *sat_solver2) int32              { return s.size }
func sat_solver2_nclauses(s *sat_solver2) int32           { return int32(s.stats.clauses) }
func sat_solver2_nlearnts(s *sat_solver2) int32           { return int32(s.stats.learnts) }
func sat_solver2_nconflicts(s *sat_solver2) int32         { return int32(s.stats.conflicts) }
func sat_solver2_var_value(s *sat_solver2, v int32) int32 { return solver2Bool(s.model[v] == l_True) }
func sat_solver2_var_literal(s *sat_solver2, v int32) lit {
	return toLitCond(v, solver2Bool(s.model[v] != l_True))
}
func sat_solver2_act_var_clear(s *sat_solver2) { clear(s.activity); s.var_inc = 1 }
func sat_solver2_final(s *sat_solver2, out *[]lit) int32 {
	*out = veci_begin(&s.conf_final)
	return s.conf_final.size
}
func sat_solver2_set_runtime_limit(s *sat_solver2, limit abctime) abctime {
	old := s.nRuntimeLimit
	s.nRuntimeLimit = limit
	return old
}
func sat_solver2_set_learntmax(s *sat_solver2, n int32) int32 {
	old := s.nLearntMax
	s.nLearntMax = n
	return old
}
