package sbd

import "github.com/glycerine/abcgo/sat/bsat"

// The sat_solver_* routines the SBD sources call, mapped onto the bsat
// C-API bridges so that a failed sat_solver_addclause does not poison the
// solver (C's sat_solver has no such flag).

type sat_solver = bsat.Solver

const (
	l_True  int32 = 1
	l_False int32 = -1
	l_Undef int32 = 0
)

func sat_solver_new() *sat_solver                { return bsat.NewSolver() }
func sat_solver_delete(s *sat_solver)            { s.Delete() }
func sat_solver_restart(s *sat_solver)           { s.RestartSearch() }
func sat_solver_setnvars(s *sat_solver, n int32) { s.SetNumVars(n) }
func sat_solver_nvars(s *sat_solver) int32       { return s.NumVars() }
func sat_solver_simplify(s *sat_solver) int32    { return s.SimplifyC() }
func sat_solver_addclause(s *sat_solver, lits []int32) int32 {
	return s.AddClauseC(lits)
}
func sat_solver_add_and(s *sat_solver, iVar, iVar0, iVar1, fCompl0, fCompl1, fCompl int32) int32 {
	return s.AddAndC(iVar, iVar0, iVar1, fCompl0, fCompl1, fCompl)
}
func sat_solver_add_xor(s *sat_solver, iVarA, iVarB, iVarC, fCompl int32) int32 {
	if s.AddXor(iVarA, iVarB, iVarC, fCompl != 0) {
		return 1
	}
	return 0
}

// sat_solver_solve( s, begin, end, nConfLimit, nInsLimit, nConfLimitGlobal, nInsLimitGlobal )
func sat_solver_solve(s *sat_solver, lits []int32, nConfLimit, nInsLimit, nConfLimitGlobal, nInsLimitGlobal int64) int32 {
	return int32(s.SolveWithResourceLimits(lits, nConfLimit, nInsLimit, nConfLimitGlobal, nInsLimitGlobal))
}
func sat_solver_var_value(s *sat_solver, v int32) int32 {
	if s.VarValue(v) {
		return 1
	}
	return 0
}
func sat_solver_var_literal(s *sat_solver, v int32) int32 {
	return Abc_Var2Lit(v, 1-sat_solver_var_value(s, v))
}

// sat_solver_final returns the conflict literals of the last UNSAT call.
func sat_solver_final(s *sat_solver) []int32 { return s.UnsatCore() }

// sat_solver_random_polarity is the static helper in sbdWin.c. Its loop
// steps i by 64 but indexes (i << 6) + k, so only variables 0..63 receive a
// random phase while Gia_ManRandomW is still called once per 64 variables.
func sat_solver_random_polarity(s *sat_solver) {
	size := s.NumVars()
	for i := int32(0); i < size; i += 64 {
		Polar := Gia_ManRandomW(0)
		for k := int32(0); k < 64 && (i<<6)+k < size; k++ {
			s.SetPolarity((i<<6)+k, (Polar>>uint(k))&1 != 0)
		}
	}
}

func Abc_Var2Lit(v, c int32) int32 { return v + v + c }
func Abc_Lit2Var(l int32) int32    { return l >> 1 }
func Abc_LitIsCompl(l int32) int32 { return l & 1 }
func Abc_LitNot(l int32) int32     { return l ^ 1 }
func Abc_LitNotCond(l, c int32) int32 {
	return l ^ b2i(c > 0)
}
func b2i(b bool) int32 {
	if b {
		return 1
	}
	return 0
}
