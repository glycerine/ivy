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

func sat_solver3_addclause(s *sat_solver, input []lit) int32 { return sat_solver_addclause(s, input) }

func sat_solver3_record(s *sat_solver, learnt *veci) { sat_solver_record(s, learnt) }
func sat_solver3_solve_lexsat(s *sat_solver, literals []lit, n int32) int32 {
	return sat_solver_solve_lexsat(s, literals, n)
}

func sat_solver3_minimize_assumptions2(s *sat_solver, literals []lit, n int32, budget int32) int32 {
	return sat_solver_minimize_assumptions2(s, literals, n, budget)
}
func sat_solver3_lit_removable(s *sat_solver, x, minl int32) int32 {
	return sat_solver_lit_removable(s, x, minl)
}

func sat_solver3_analyze_final(s *sat_solver, h int32, skip int32) {
	sat_solver_analyze_final(s, h, skip)
}

func sat_solver3_analyze(s *sat_solver, h int32, learnt *veci) { sat_solver_analyze(s, h, learnt) }
func sat_solver3_act_var_clear(s *sat_solver)                  { sat_solver_act_var_clear(s) }

func sat_solver3_compress(s *sat_solver) { sat_solver_compress(s) }

func sat_solver3_delete_p(p **sat_solver) { sat_solver_delete_p(p) }

func sat_solver3_clean_polarity(s *sat_solver, vars []int32, n int32) {
	sat_solver_clean_polarity(s, vars, n)
}

func sat_solver3_set_polarity(s *sat_solver, vars []int32, n int32) {
	sat_solver_set_polarity(s, vars, n)
}

func sat_solver3_set_literal_polarity(s *sat_solver, literals []lit, n int32) {
	sat_solver_set_literal_polarity(s, literals, n)
}

func sat_solver3_final(s *sat_solver, out *[]lit) int32 { return sat_solver_final(s, out) }

func sat_solver3_set_runtime_limit(s *sat_solver, limit abctime) abctime {
	return sat_solver_set_runtime_limit(s, limit)
}

func sat_solver3_set_random(s *sat_solver, disabled int32) int32 {
	return sat_solver_set_random(s, disabled)
}

func sat_solver3_bookmark(s *sat_solver) { sat_solver_bookmark(s) }

func sat_solver3_set_pivot_variables(s *sat_solver, vars []int32, n int32) {
	sat_solver_set_pivot_variables(s, vars, n)
}

func sat_solver3_count_usedvars(s *sat_solver) int32 { return sat_solver_count_usedvars(s) }
func sat_solver3_enqueue(s *sat_solver, l lit, from int32) int32 {
	return sat_solver_enqueue(s, l, from)
}

func sat_solver3_decision(s *sat_solver, l lit) int32 { return sat_solver_decision(s, l) }

func sat_solver3_canceluntil(s *sat_solver, level int32) { sat_solver_canceluntil(s, level) }

func sat_solver3_canceluntil_rollback(s *sat_solver, bound int32) {
	sat_solver_canceluntil_rollback(s, bound)
}

func sat_solver3_propagate(s *sat_solver) int32 { return sat_solver_propagate(s) }

func sat_solver3_get_var_value(s *sat_solver, v int32) int32 { return sat_solver_get_var_value(s, v) }

func sat_solver3_dl(s *sat_solver) int32                { return sat_solver_dl(s) }
func sat_solver3_read_wlist(s *sat_solver, l lit) *veci { return sat_solver_read_wlist(s, l) }
func sat_solver3_new() *sat_solver                      { return sat_solver_new() }
func zsat_solver3_new_seed(seed float64) *sat_solver    { return zsat_solver_new_seed(seed) }

func sat_solver3_setnvars(s *sat_solver, n int32) { sat_solver_setnvars(s, n) }

func sat_solver3_delete(s *sat_solver) { sat_solver_delete(s) }

func sat_solver3_restart(s *sat_solver)                     { sat_solver_restart(s) }
func zsat_solver3_restart_seed(s *sat_solver, seed float64) { zsat_solver_restart_seed(s, seed) }
func sat_solver3_clause_new(s *sat_solver, literals []lit, learnt int32) cla {
	return sat_solver_clause_new(s, literals, learnt)
}

func sat_solver3_simplify(s *sat_solver) int32 { return sat_solver_simplify(s) }

func sat_solver3_reducedb(s *sat_solver) { sat_solver_reducedb(s) }
func sat_solver3_rollback(s *sat_solver) { sat_solver_rollback(s) }

func sat_solver3_count_assigned(s *sat_solver) int32 { return sat_solver_count_assigned(s) }

func sat_solver3_progress(s *sat_solver) float64 { return sat_solver_progress(s) }

func sat_solver3_search(s *sat_solver, budget int64) int32 { return sat_solver_search(s, budget) }

func sat_solver3_solve_internal(s *sat_solver) int32     { return sat_solver_solve_internal(s) }
func sat_solver3_var_value(s *sat_solver, v int32) int32 { return sat_solver_var_value(s, v) }

func sat_solver3_var_literal(s *sat_solver, v int32) lit { return sat_solver_var_literal(s, v) }

func sat_solver3_push(s *sat_solver, p lit) int32 { return sat_solver_push(s, p) }

func sat_solver3_pop(s *sat_solver) { sat_solver_pop(s) }

func sat_solver3_set_resource_limits(s *sat_solver, conflicts, propagations, globalConflicts, globalPropagations int64) {
	sat_solver_set_resource_limits(s, conflicts, propagations, globalConflicts, globalPropagations)
}

func sat_solver3_solve(s *sat_solver, assumptions []lit, conflicts, propagations, globalConflicts, globalPropagations int64) int32 {
	return sat_solver_solve(s, assumptions, conflicts, propagations, globalConflicts, globalPropagations)
}

func sat_solver3_memory(s *sat_solver) float64 { return sat_solver_memory(s) }

func sat_solver3_nvars(s *sat_solver) int32      { return sat_solver_nvars(s) }
func sat_solver3_nclauses(s *sat_solver) int32   { return sat_solver_nclauses(s) }
func sat_solver3_nconflicts(s *sat_solver) int32 { return sat_solver_nconflicts(s) }
func sat_solver3_add_const(pSat *sat_solver, iVar int32, fCompl int32) int32 {
	return sat_solver_add_const(pSat, iVar, fCompl)
}

func sat_solver3_add_buffer(pSat *sat_solver, iVarA int32, iVarB int32, fCompl int32) int32 {
	return sat_solver_add_buffer(pSat, iVarA, iVarB, fCompl)
}

func sat_solver3_add_buffer_enable(pSat *sat_solver, iVarA int32, iVarB int32, iVarEn int32, fCompl int32) int32 {
	return sat_solver_add_buffer_enable(pSat, iVarA, iVarB, iVarEn, fCompl)
}

func sat_solver3_add_and(pSat *sat_solver, iVar int32, iVar0 int32, iVar1 int32, fCompl0 int32, fCompl1 int32, fCompl int32) int32 {
	return sat_solver_add_and(pSat, iVar, iVar0, iVar1, fCompl0, fCompl1, fCompl)
}

func sat_solver3_add_xor(pSat *sat_solver, iVarA int32, iVarB int32, iVarC int32, fCompl int32) int32 {
	return sat_solver_add_xor(pSat, iVarA, iVarB, iVarC, fCompl)
}

func sat_solver3_add_mux(pSat *sat_solver, iVarZ int32, iVarC int32, iVarT int32, iVarE int32, iComplC int32, iComplT int32, iComplE int32, iComplZ int32) int32 {
	return sat_solver_add_mux(pSat, iVarZ, iVarC, iVarT, iVarE, iComplC, iComplT, iComplE, iComplZ)
}

func sat_solver3_add_mux41(pSat *sat_solver, iVarZ int32, iVarC0 int32, iVarC1 int32, iVarD0 int32, iVarD1 int32, iVarD2 int32, iVarD3 int32) int32 {
	return sat_solver_add_mux41(pSat, iVarZ, iVarC0, iVarC1, iVarD0, iVarD1, iVarD2, iVarD3)
}

func sat_solver3_add_xor_and(pSat *sat_solver, iVarF int32, iVarA int32, iVarB int32, iVarC int32) int32 {
	return sat_solver_add_xor_and(pSat, iVarF, iVarA, iVarB, iVarC)
}

func sat_solver3_add_constraint(pSat *sat_solver, iVar int32, iVar2 int32, fCompl int32) int32 {
	return sat_solver_add_constraint(pSat, iVar, iVar2, fCompl)
}

func sat_solver3_add_half_sorter(pSat *sat_solver, iVarA int32, iVarB int32, iVar0 int32, iVar1 int32) int32 {
	return sat_solver_add_half_sorter(pSat, iVarA, iVarB, iVar0, iVar1)
}
func sat_solver3_store_alloc(s *sat_solver) { sat_solver_store_alloc(s) }

func sat_solver3_store_write(s *sat_solver, name string) { sat_solver_store_write(s, name) }

func sat_solver3_store_free(s *sat_solver) { sat_solver_store_free(s) }

func sat_solver3_store_mark_roots(s *sat_solver) { sat_solver_store_mark_roots(s) }

func sat_solver3_store_mark_clauses_a(s *sat_solver) { sat_solver_store_mark_clauses_a(s) }

func sat_solver3_store_release(s *sat_solver) any { return sat_solver_store_release(s) }
