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
// C: extern bmcg2_sat_solver * bmcg2_sat_solver_start()
func bmcg2_sat_solver_start() Bmcg2SatSolver {
	panic("TODO, stubbed")
}

// bmcg2_sat_solver_stop
// C: extern void bmcg2_sat_solver_stop(bmcg2_sat_solver* s)
func bmcg2_sat_solver_stop(s Bmcg2SatSolver) {
	panic("TODO, stubbed")
}

// bmcg2_sat_solver_reset
// C: extern void bmcg2_sat_solver_reset(bmcg2_sat_solver* s)
func bmcg2_sat_solver_reset(s Bmcg2SatSolver) {
	panic("TODO, stubbed")
}

// bmcg2_sat_solver_addclause
// C: extern int bmcg2_sat_solver_addclause(bmcg2_sat_solver* s, int * plits, int nlits)
func bmcg2_sat_solver_addclause(s Bmcg2SatSolver, plits []int, nlits int) int {
	panic("TODO, stubbed")
}

// bmcg2_sat_solver_setcallback
// C: extern void bmcg2_sat_solver_setcallback(bmcg2_sat_solver* s, void * pman, int(*pfunc)(void*, int, int*))
func bmcg2_sat_solver_setcallback(s Bmcg2SatSolver, pman any, pfunc func(any, int, []int) int) {
	panic("TODO, stubbed")
}

// bmcg2_sat_solver_solve
// C: extern int bmcg2_sat_solver_solve(bmcg2_sat_solver* s, int * plits, int nlits)
func bmcg2_sat_solver_solve(s Bmcg2SatSolver, plits []int, nlits int) int {
	panic("TODO, stubbed")
}

// bmcg2_sat_solver_final (ppArray is an out parameter)
// C: extern int bmcg2_sat_solver_final(bmcg2_sat_solver* s, int ** ppArray)
func bmcg2_sat_solver_final(s Bmcg2SatSolver, ppArray *[]int) int {
	panic("TODO, stubbed")
}

// bmcg2_sat_solver_addvar
// C: extern int bmcg2_sat_solver_addvar(bmcg2_sat_solver* s)
func bmcg2_sat_solver_addvar(s Bmcg2SatSolver) int {
	panic("TODO, stubbed")
}

// bmcg2_sat_solver_set_nvars
// C: extern void bmcg2_sat_solver_set_nvars(bmcg2_sat_solver* s, int nvars)
func bmcg2_sat_solver_set_nvars(s Bmcg2SatSolver, nvars int) {
	panic("TODO, stubbed")
}

// bmcg2_sat_solver_eliminate
// C: extern int bmcg2_sat_solver_eliminate(bmcg2_sat_solver* s, int turn_off_elim)
func bmcg2_sat_solver_eliminate(s Bmcg2SatSolver, turn_off_elim int) int {
	panic("TODO, stubbed")
}

// bmcg2_sat_solver_var_is_elim
// C: extern int bmcg2_sat_solver_var_is_elim(bmcg2_sat_solver* s, int v)
func bmcg2_sat_solver_var_is_elim(s Bmcg2SatSolver, v int) int {
	panic("TODO, stubbed")
}

// bmcg2_sat_solver_var_set_frozen
// C: extern void bmcg2_sat_solver_var_set_frozen(bmcg2_sat_solver* s, int v, int freeze)
func bmcg2_sat_solver_var_set_frozen(s Bmcg2SatSolver, v int, freeze int) {
	panic("TODO, stubbed")
}

// bmcg2_sat_solver_elim_varnum
// C: extern int bmcg2_sat_solver_elim_varnum(bmcg2_sat_solver* s)
func bmcg2_sat_solver_elim_varnum(s Bmcg2SatSolver) int {
	panic("TODO, stubbed")
}

// bmcg2_sat_solver_read_cex (returns the model array)
// C: extern int * bmcg2_sat_solver_read_cex(bmcg2_sat_solver* s)
func bmcg2_sat_solver_read_cex(s Bmcg2SatSolver) []int {
	panic("TODO, stubbed")
}

// bmcg2_sat_solver_read_cex_varvalue (the second C parameter is unnamed)
// C: extern int bmcg2_sat_solver_read_cex_varvalue(bmcg2_sat_solver* s, int)
func bmcg2_sat_solver_read_cex_varvalue(s Bmcg2SatSolver, a1 int) int {
	panic("TODO, stubbed")
}

// bmcg2_sat_solver_set_stop (pstop points at a single stop flag)
// C: extern void bmcg2_sat_solver_set_stop(bmcg2_sat_solver* s, int * pstop)
func bmcg2_sat_solver_set_stop(s Bmcg2SatSolver, pstop *int) {
	panic("TODO, stubbed")
}

// bmcg2_sat_solver_set_runtime_limit
// C: extern abctime bmcg2_sat_solver_set_runtime_limit(bmcg2_sat_solver* s, abctime Limit)
func bmcg2_sat_solver_set_runtime_limit(s Bmcg2SatSolver, Limit int64) int64 {
	panic("TODO, stubbed")
}

// bmcg2_sat_solver_set_conflict_budget
// C: extern void bmcg2_sat_solver_set_conflict_budget(bmcg2_sat_solver* s, int Limit)
func bmcg2_sat_solver_set_conflict_budget(s Bmcg2SatSolver, Limit int) {
	panic("TODO, stubbed")
}

// bmcg2_sat_solver_varnum
// C: extern int bmcg2_sat_solver_varnum(bmcg2_sat_solver* s)
func bmcg2_sat_solver_varnum(s Bmcg2SatSolver) int {
	panic("TODO, stubbed")
}

// bmcg2_sat_solver_clausenum
// C: extern int bmcg2_sat_solver_clausenum(bmcg2_sat_solver* s)
func bmcg2_sat_solver_clausenum(s Bmcg2SatSolver) int {
	panic("TODO, stubbed")
}

// bmcg2_sat_solver_learntnum
// C: extern int bmcg2_sat_solver_learntnum(bmcg2_sat_solver* s)
func bmcg2_sat_solver_learntnum(s Bmcg2SatSolver) int {
	panic("TODO, stubbed")
}

// bmcg2_sat_solver_conflictnum
// C: extern int bmcg2_sat_solver_conflictnum(bmcg2_sat_solver* s)
func bmcg2_sat_solver_conflictnum(s Bmcg2SatSolver) int {
	panic("TODO, stubbed")
}

// bmcg2_sat_solver_minimize_assumptions
// C: extern int bmcg2_sat_solver_minimize_assumptions(bmcg2_sat_solver * s, int * plits, int nlits, int pivot)
func bmcg2_sat_solver_minimize_assumptions(s Bmcg2SatSolver, plits []int, nlits int, pivot int) int {
	panic("TODO, stubbed")
}

// bmcg2_sat_solver_add_and
// C: extern int bmcg2_sat_solver_add_and(bmcg2_sat_solver * s, int iVar, int iVar0, int iVar1, int fCompl0, int fCompl1, int fCompl)
func bmcg2_sat_solver_add_and(s Bmcg2SatSolver, iVar int, iVar0 int, iVar1 int, fCompl0 int, fCompl1 int, fCompl int) int {
	panic("TODO, stubbed")
}

// bmcg2_sat_solver_add_xor
// C: extern int bmcg2_sat_solver_add_xor(bmcg2_sat_solver * s, int iVarA, int iVarB, int iVarC, int fCompl)
func bmcg2_sat_solver_add_xor(s Bmcg2SatSolver, iVarA int, iVarB int, iVarC int, fCompl int) int {
	panic("TODO, stubbed")
}

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
// C: extern int bmcg2_sat_solver_jftr(bmcg2_sat_solver * s)
func bmcg2_sat_solver_jftr(s Bmcg2SatSolver) int {
	panic("TODO, stubbed")
}

// bmcg2_sat_solver_set_jftr
// C: extern void bmcg2_sat_solver_set_jftr(bmcg2_sat_solver * s, int jftr)
func bmcg2_sat_solver_set_jftr(s Bmcg2SatSolver, jftr int) {
	panic("TODO, stubbed")
}

// bmcg2_sat_solver_set_var_fanin_lit
// C: extern void bmcg2_sat_solver_set_var_fanin_lit(bmcg2_sat_solver * s, int var, int lit0, int lit1)
func bmcg2_sat_solver_set_var_fanin_lit(s Bmcg2SatSolver, var_ int, lit0 int, lit1 int) {
	panic("TODO, stubbed")
}

// bmcg2_sat_solver_start_new_round
// C: extern void bmcg2_sat_solver_start_new_round(bmcg2_sat_solver * s)
func bmcg2_sat_solver_start_new_round(s Bmcg2SatSolver) {
	panic("TODO, stubbed")
}

// bmcg2_sat_solver_mark_cone
// C: extern void bmcg2_sat_solver_mark_cone(bmcg2_sat_solver * s, int var)
func bmcg2_sat_solver_mark_cone(s Bmcg2SatSolver, var_ int) {
	panic("TODO, stubbed")
}

// Glucose2_SolveCnf
// C: extern void Glucose2_SolveCnf( char * pFilename, Glucose2_Pars * pPars )
func Glucose2SolveCnf(pFilename string, pPars *Glucose2Pars) {
	panic("TODO, stubbed")
}

// Glucose2_SolveAig
// C: extern int Glucose2_SolveAig( Gia_Man_t * p, Glucose2_Pars * pPars )
func (p *GiaManager) Glucose2SolveAig(pPars *Glucose2Pars) int {
	panic("TODO, stubbed")
}
