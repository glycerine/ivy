package abc

// Hand-checked skeleton of src/sat/glucose/AbcGlucose.h.

// GlucosePars is Glucose_Pars.
type GlucosePars struct {
	pre     int // preprocessing
	verb    int // verbosity
	cust    int // customizable
	nConfls int // conflict limit (0 = no limit)
}

// BmcgSatSolver is `typedef void bmcg_sat_solver;`: the pointer type bmcg_sat_solver * is void *.
// The object behind it is the C++ SimpSolver of this directory.
type BmcgSatSolver = any

// Glucose_CreatePars (static inline)
// C: Glucose_Pars Glucose_CreatePars(int p, int v, int c, int nConfls)
func GlucoseCreatePars(p int, v int, c int, nConfls int) GlucosePars {
	panic("TODO, stubbed")
}

// bmcg_sat_solver_start
// bmcg_sat_solver_stop
// bmcg_sat_solver_reset
// bmcg_sat_solver_addclause
// bmcg_sat_solver_setcallback
// bmcg_sat_solver_solve
// bmcg_sat_solver_final (ppArray is an out parameter)
// bmcg_sat_solver_addvar
// bmcg_sat_solver_set_nvars
// bmcg_sat_solver_eliminate
// bmcg_sat_solver_var_is_elim
// bmcg_sat_solver_var_set_frozen
// bmcg_sat_solver_elim_varnum
// bmcg_sat_solver_read_cex (returns the model array)
// bmcg_sat_solver_read_cex_varvalue (the second C parameter is unnamed)
// bmcg_sat_solver_set_stop (pstop points at a single stop flag)
// bmcg_sat_solver_set_runtime_limit
// bmcg_sat_solver_set_conflict_budget
// bmcg_sat_solver_varnum
// bmcg_sat_solver_clausenum
// bmcg_sat_solver_learntnum
// bmcg_sat_solver_conflictnum
// bmcg_sat_solver_minimize_assumptions
// bmcg_sat_solver_add_and
// bmcg_sat_solver_add_xor
// bmcg_sat_solver_quantify
// C: extern int bmcg_sat_solver_quantify(bmcg_sat_solver * s[], Gia_Man_t * p, int iLit, int fHash, int(*pFuncCiToKeep)(void *, int), void * pData, Vec_Int_t * vDLits)
func bmcg_sat_solver_quantify(s []BmcgSatSolver, p *GiaManager, iLit int, fHash int, pFuncCiToKeep func(any, int) int, pData any, vDLits *IntVector) int {
	panic("TODO, stubbed")
}

// bmcg_sat_solver_equiv_overlap_check
// C: extern int bmcg_sat_solver_equiv_overlap_check(bmcg_sat_solver * s, Gia_Man_t * p, int iLit0, int iLit1, int fEquiv)
func bmcg_sat_solver_equiv_overlap_check(s BmcgSatSolver, p *GiaManager, iLit0 int, iLit1 int, fEquiv int) int {
	panic("TODO, stubbed")
}

// bmcg_sat_solver_sop
// C: extern Vec_Str_t * bmcg_sat_solver_sop(Gia_Man_t * p, int CubeLimit)
func bmcg_sat_solver_sop(p *GiaManager, CubeLimit int) *ByteVector {
	panic("TODO, stubbed")
}

// bmcg_sat_solver_jftr
// bmcg_sat_solver_set_jftr
// bmcg_sat_solver_set_var_fanin_lit
// bmcg_sat_solver_start_new_round
// bmcg_sat_solver_mark_cone
// Glucose_SolveCnf
// Glucose_SolveAig
// C: extern int Glucose_SolveAig( Gia_Man_t * p, Glucose_Pars * pPars )
func (p *GiaManager) GlucoseSolveAig(pPars *GlucosePars) int {
	panic("TODO, stubbed")
}
