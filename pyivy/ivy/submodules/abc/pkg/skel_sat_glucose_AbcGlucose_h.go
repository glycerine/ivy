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
// C: extern bmcg_sat_solver * bmcg_sat_solver_start()
func bmcg_sat_solver_start() BmcgSatSolver {
	panic("TODO, stubbed")
}

// bmcg_sat_solver_stop
// C: extern void bmcg_sat_solver_stop(bmcg_sat_solver* s)
func bmcg_sat_solver_stop(s BmcgSatSolver) {
	panic("TODO, stubbed")
}

// bmcg_sat_solver_reset
// C: extern void bmcg_sat_solver_reset(bmcg_sat_solver* s)
func bmcg_sat_solver_reset(s BmcgSatSolver) {
	panic("TODO, stubbed")
}

// bmcg_sat_solver_addclause
// C: extern int bmcg_sat_solver_addclause(bmcg_sat_solver* s, int * plits, int nlits)
func bmcg_sat_solver_addclause(s BmcgSatSolver, plits []int, nlits int) int {
	panic("TODO, stubbed")
}

// bmcg_sat_solver_setcallback
// C: extern void bmcg_sat_solver_setcallback(bmcg_sat_solver* s, void * pman, int(*pfunc)(void*, int, int*))
func bmcg_sat_solver_setcallback(s BmcgSatSolver, pman any, pfunc func(any, int, []int) int) {
	panic("TODO, stubbed")
}

// bmcg_sat_solver_solve
// C: extern int bmcg_sat_solver_solve(bmcg_sat_solver* s, int * plits, int nlits)
func bmcg_sat_solver_solve(s BmcgSatSolver, plits []int, nlits int) int {
	panic("TODO, stubbed")
}

// bmcg_sat_solver_final (ppArray is an out parameter)
// C: extern int bmcg_sat_solver_final(bmcg_sat_solver* s, int ** ppArray)
func bmcg_sat_solver_final(s BmcgSatSolver, ppArray *[]int) int {
	panic("TODO, stubbed")
}

// bmcg_sat_solver_addvar
// C: extern int bmcg_sat_solver_addvar(bmcg_sat_solver* s)
func bmcg_sat_solver_addvar(s BmcgSatSolver) int {
	panic("TODO, stubbed")
}

// bmcg_sat_solver_set_nvars
// C: extern void bmcg_sat_solver_set_nvars(bmcg_sat_solver* s, int nvars)
func bmcg_sat_solver_set_nvars(s BmcgSatSolver, nvars int) {
	panic("TODO, stubbed")
}

// bmcg_sat_solver_eliminate
// C: extern int bmcg_sat_solver_eliminate(bmcg_sat_solver* s, int turn_off_elim)
func bmcg_sat_solver_eliminate(s BmcgSatSolver, turn_off_elim int) int {
	panic("TODO, stubbed")
}

// bmcg_sat_solver_var_is_elim
// C: extern int bmcg_sat_solver_var_is_elim(bmcg_sat_solver* s, int v)
func bmcg_sat_solver_var_is_elim(s BmcgSatSolver, v int) int {
	panic("TODO, stubbed")
}

// bmcg_sat_solver_var_set_frozen
// C: extern void bmcg_sat_solver_var_set_frozen(bmcg_sat_solver* s, int v, int freeze)
func bmcg_sat_solver_var_set_frozen(s BmcgSatSolver, v int, freeze int) {
	panic("TODO, stubbed")
}

// bmcg_sat_solver_elim_varnum
// C: extern int bmcg_sat_solver_elim_varnum(bmcg_sat_solver* s)
func bmcg_sat_solver_elim_varnum(s BmcgSatSolver) int {
	panic("TODO, stubbed")
}

// bmcg_sat_solver_read_cex (returns the model array)
// C: extern int * bmcg_sat_solver_read_cex(bmcg_sat_solver* s)
func bmcg_sat_solver_read_cex(s BmcgSatSolver) []int {
	panic("TODO, stubbed")
}

// bmcg_sat_solver_read_cex_varvalue (the second C parameter is unnamed)
// C: extern int bmcg_sat_solver_read_cex_varvalue(bmcg_sat_solver* s, int)
func bmcg_sat_solver_read_cex_varvalue(s BmcgSatSolver, a1 int) int {
	panic("TODO, stubbed")
}

// bmcg_sat_solver_set_stop (pstop points at a single stop flag)
// C: extern void bmcg_sat_solver_set_stop(bmcg_sat_solver* s, int * pstop)
func bmcg_sat_solver_set_stop(s BmcgSatSolver, pstop *int) {
	panic("TODO, stubbed")
}

// bmcg_sat_solver_set_runtime_limit
// C: extern abctime bmcg_sat_solver_set_runtime_limit(bmcg_sat_solver* s, abctime Limit)
func bmcg_sat_solver_set_runtime_limit(s BmcgSatSolver, Limit int64) int64 {
	panic("TODO, stubbed")
}

// bmcg_sat_solver_set_conflict_budget
// C: extern void bmcg_sat_solver_set_conflict_budget(bmcg_sat_solver* s, int Limit)
func bmcg_sat_solver_set_conflict_budget(s BmcgSatSolver, Limit int) {
	panic("TODO, stubbed")
}

// bmcg_sat_solver_varnum
// C: extern int bmcg_sat_solver_varnum(bmcg_sat_solver* s)
func bmcg_sat_solver_varnum(s BmcgSatSolver) int {
	panic("TODO, stubbed")
}

// bmcg_sat_solver_clausenum
// C: extern int bmcg_sat_solver_clausenum(bmcg_sat_solver* s)
func bmcg_sat_solver_clausenum(s BmcgSatSolver) int {
	panic("TODO, stubbed")
}

// bmcg_sat_solver_learntnum
// C: extern int bmcg_sat_solver_learntnum(bmcg_sat_solver* s)
func bmcg_sat_solver_learntnum(s BmcgSatSolver) int {
	panic("TODO, stubbed")
}

// bmcg_sat_solver_conflictnum
// C: extern int bmcg_sat_solver_conflictnum(bmcg_sat_solver* s)
func bmcg_sat_solver_conflictnum(s BmcgSatSolver) int {
	panic("TODO, stubbed")
}

// bmcg_sat_solver_minimize_assumptions
// C: extern int bmcg_sat_solver_minimize_assumptions(bmcg_sat_solver * s, int * plits, int nlits, int pivot)
func bmcg_sat_solver_minimize_assumptions(s BmcgSatSolver, plits []int, nlits int, pivot int) int {
	panic("TODO, stubbed")
}

// bmcg_sat_solver_add_and
// C: extern int bmcg_sat_solver_add_and(bmcg_sat_solver * s, int iVar, int iVar0, int iVar1, int fCompl0, int fCompl1, int fCompl)
func bmcg_sat_solver_add_and(s BmcgSatSolver, iVar int, iVar0 int, iVar1 int, fCompl0 int, fCompl1 int, fCompl int) int {
	panic("TODO, stubbed")
}

// bmcg_sat_solver_add_xor
// C: extern int bmcg_sat_solver_add_xor(bmcg_sat_solver * s, int iVarA, int iVarB, int iVarC, int fCompl)
func bmcg_sat_solver_add_xor(s BmcgSatSolver, iVarA int, iVarB int, iVarC int, fCompl int) int {
	panic("TODO, stubbed")
}

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
// C: extern int bmcg_sat_solver_jftr(bmcg_sat_solver * s)
func bmcg_sat_solver_jftr(s BmcgSatSolver) int {
	panic("TODO, stubbed")
}

// bmcg_sat_solver_set_jftr
// C: extern void bmcg_sat_solver_set_jftr(bmcg_sat_solver * s, int jftr)
func bmcg_sat_solver_set_jftr(s BmcgSatSolver, jftr int) {
	panic("TODO, stubbed")
}

// bmcg_sat_solver_set_var_fanin_lit
// C: extern void bmcg_sat_solver_set_var_fanin_lit(bmcg_sat_solver * s, int var, int lit0, int lit1)
func bmcg_sat_solver_set_var_fanin_lit(s BmcgSatSolver, var_ int, lit0 int, lit1 int) {
	panic("TODO, stubbed")
}

// bmcg_sat_solver_start_new_round
// C: extern void bmcg_sat_solver_start_new_round(bmcg_sat_solver * s)
func bmcg_sat_solver_start_new_round(s BmcgSatSolver) {
	panic("TODO, stubbed")
}

// bmcg_sat_solver_mark_cone
// C: extern void bmcg_sat_solver_mark_cone(bmcg_sat_solver * s, int var)
func bmcg_sat_solver_mark_cone(s BmcgSatSolver, var_ int) {
	panic("TODO, stubbed")
}

// Glucose_SolveCnf
// C: extern void Glucose_SolveCnf( char * pFilename, Glucose_Pars * pPars )
func GlucoseSolveCnf(pFilename string, pPars *GlucosePars) {
	panic("TODO, stubbed")
}

// Glucose_SolveAig
// C: extern int Glucose_SolveAig( Gia_Man_t * p, Glucose_Pars * pPars )
func (p *GiaManager) GlucoseSolveAig(pPars *GlucosePars) int {
	panic("TODO, stubbed")
}
