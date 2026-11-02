package abc

// Hand-checked skeleton of src/sat/glucose2/AbcGlucose2.h.

// Glucose2Pars is Glucose2_Pars.
type Glucose2Pars struct {
	pre     int // preprocessing
	verb    int // verbosity
	cust    int // customizable
	nConfls int // conflict limit (0 = no limit)
}

// Bmcg2SatSolver is `typedef void bmcg2_sat_solver;`: the pointer type bmcg2_sat_solver * is void *.
// The object behind it is the C++ SimpSolver of this directory.
type Bmcg2SatSolver = any

// Glucose2_CreatePars (static inline)
// C: Glucose2_Pars Glucose2_CreatePars(int p, int v, int c, int nConfls)
func Glucose2CreatePars(p int, v int, c int, nConfls int) Glucose2Pars {
	panic("TODO, stubbed")
}

// bmcg2_sat_solver_start
// bmcg2_sat_solver_stop
// bmcg2_sat_solver_reset
// bmcg2_sat_solver_addclause
// bmcg2_sat_solver_setcallback
// bmcg2_sat_solver_solve
// bmcg2_sat_solver_final (ppArray is an out parameter)
// bmcg2_sat_solver_addvar
// bmcg2_sat_solver_set_nvars
// bmcg2_sat_solver_eliminate
// bmcg2_sat_solver_var_is_elim
// bmcg2_sat_solver_var_set_frozen
// bmcg2_sat_solver_elim_varnum
// bmcg2_sat_solver_read_cex (returns the model array)
// bmcg2_sat_solver_read_cex_varvalue (the second C parameter is unnamed)
// bmcg2_sat_solver_set_stop (pstop points at a single stop flag)
// bmcg2_sat_solver_set_runtime_limit
// bmcg2_sat_solver_set_conflict_budget
// bmcg2_sat_solver_varnum
// bmcg2_sat_solver_clausenum
// bmcg2_sat_solver_learntnum
// bmcg2_sat_solver_conflictnum
// bmcg2_sat_solver_minimize_assumptions
// bmcg2_sat_solver_add_and
// bmcg2_sat_solver_add_xor
// bmcg2_sat_solver_quantify
// C: extern int bmcg2_sat_solver_quantify(bmcg2_sat_solver * s[], Gia_Man_t * p, int iLit, int fHash, int(*pFuncCiToKeep)(void *, int), void * pData, Vec_Int_t * vDLits)
func bmcg2_sat_solver_quantify(s []Bmcg2SatSolver, p *GiaManager, iLit int, fHash int, pFuncCiToKeep func(any, int) int, pData any, vDLits *IntVector) int {
	panic("TODO, stubbed")
}

// bmcg2_sat_solver_equiv_overlap_check
// C: extern int bmcg2_sat_solver_equiv_overlap_check(bmcg2_sat_solver * s, Gia_Man_t * p, int iLit0, int iLit1, int fEquiv)
func bmcg2_sat_solver_equiv_overlap_check(s Bmcg2SatSolver, p *GiaManager, iLit0 int, iLit1 int, fEquiv int) int {
	panic("TODO, stubbed")
}

// bmcg2_sat_solver_sop
// C: extern Vec_Str_t * bmcg2_sat_solver_sop(Gia_Man_t * p, int CubeLimit)
func bmcg2_sat_solver_sop(p *GiaManager, CubeLimit int) *ByteVector {
	panic("TODO, stubbed")
}

// bmcg2_sat_solver_jftr
// bmcg2_sat_solver_set_jftr
// bmcg2_sat_solver_set_var_fanin_lit
// bmcg2_sat_solver_start_new_round
// bmcg2_sat_solver_mark_cone
// Glucose2_SolveCnf
// Glucose2_SolveAig
// C: extern int Glucose2_SolveAig( Gia_Man_t * p, Glucose2_Pars * pPars )
func (p *GiaManager) Glucose2SolveAig(pPars *Glucose2Pars) int {
	panic("TODO, stubbed")
}
