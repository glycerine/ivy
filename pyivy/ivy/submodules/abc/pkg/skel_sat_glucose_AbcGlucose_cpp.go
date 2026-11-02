package abc

// Hand-checked skeleton of src/sat/glucose/AbcGlucose.cpp (namespace Gluco).
//
// `#define USE_SIMP_SOLVER 1` is set at the top of the file, so only the SimpSolver variant of the
// glucose-solver wrappers is live (the `#else` branch is dead and not stubbed). The extern "C"
// bmcg_sat_solver_* entry points are stubbed in skel_sat_glucose_AbcGlucose_h.go next to their
// declarations. Functions whose first parameter is a SimpSolver stay free functions (they are not
// class members).

// glucose_solver_start
// C: SimpSolver * glucose_solver_start()
func glucose_solver_start() *GlucoSimpSolver {
	panic("TODO, stubbed")
}

// glucose_solver_stop
// C: void glucose_solver_stop(Gluco::SimpSolver* S)
func glucose_solver_stop(S *GlucoSimpSolver) {
	panic("TODO, stubbed")
}

// glucose_solver_reset
// C: void glucose_solver_reset(Gluco::SimpSolver* S)
func glucose_solver_reset(S *GlucoSimpSolver) {
	panic("TODO, stubbed")
}

// glucose_solver_addclause
// C: int glucose_solver_addclause(Gluco::SimpSolver* S, int * plits, int nlits)
func glucose_solver_addclause(S *GlucoSimpSolver, plits []int, nlits int) int {
	panic("TODO, stubbed")
}

// glucose_solver_setcallback
// C: void glucose_solver_setcallback(Gluco::SimpSolver* S, void * pman, int(*pfunc)(void*, int, int*))
func glucose_solver_setcallback(S *GlucoSimpSolver, pman any, pfunc func(any, int, []int) int) {
	panic("TODO, stubbed")
}

// glucose_solver_solve
// C: int glucose_solver_solve(Gluco::SimpSolver* S, int * plits, int nlits)
func glucose_solver_solve(S *GlucoSimpSolver, plits []int, nlits int) int {
	panic("TODO, stubbed")
}

// glucose_solver_addvar
// C: int glucose_solver_addvar(Gluco::SimpSolver* S)
func glucose_solver_addvar(S *GlucoSimpSolver) int {
	panic("TODO, stubbed")
}

// glucose_solver_read_cex
// C: int * glucose_solver_read_cex(Gluco::SimpSolver* S )
func glucose_solver_read_cex(S *GlucoSimpSolver) []int {
	panic("TODO, stubbed")
}

// glucose_solver_read_cex_varvalue
// C: int glucose_solver_read_cex_varvalue(Gluco::SimpSolver* S, int ivar)
func glucose_solver_read_cex_varvalue(S *GlucoSimpSolver, ivar int) int {
	panic("TODO, stubbed")
}

// glucose_solver_setstop
// C: void glucose_solver_setstop(Gluco::SimpSolver* S, int * pstop)
func glucose_solver_setstop(S *GlucoSimpSolver, pstop *int) {
	panic("TODO, stubbed")
}

// bmcg_solver_add_xor (the C file defines this name; the header declares bmcg_sat_solver_add_xor)
// C: int bmcg_solver_add_xor( bmcg_sat_solver * pSat, int iVarA, int iVarB, int iVarC, int fCompl )
func bmcg_solver_add_xor(pSat BmcgSatSolver, iVarA int, iVarB int, iVarC int, fCompl int) int {
	panic("TODO, stubbed")
}

// glucose_print_stats
// C: void glucose_print_stats(SimpSolver& s, abctime clk)
func glucose_print_stats(s *GlucoSimpSolver, clk int64) {
	panic("TODO, stubbed")
}

// Glucose_ReadDimacs
// C: void Glucose_ReadDimacs( char * pFileName, SimpSolver& s )
func GlucoseReadDimacs(pFileName string, s *GlucoSimpSolver) {
	panic("TODO, stubbed")
}

// Glucose_SolverFromAig
// C: Vec_Int_t * Glucose_SolverFromAig( Gia_Man_t * p, SimpSolver& s )
func (p *GiaManager) GlucoseSolverFromAig(s *GlucoSimpSolver) *IntVector {
	panic("TODO, stubbed")
}

// Glucose_SolverFromAig2
// C: Vec_Int_t * Glucose_SolverFromAig2( Gia_Man_t * p, SimpSolver& S )
func (p *GiaManager) GlucoseSolverFromAig2(S *GlucoSimpSolver) *IntVector {
	panic("TODO, stubbed")
}

// Glucose_GenerateCubes
// C: Vec_Str_t * Glucose_GenerateCubes( bmcg_sat_solver * pSat[2], Vec_Int_t * vCiSatVars, Vec_Int_t * vVar2Index, int CubeLimit )
func GlucoseGenerateCubes(pSat []BmcgSatSolver, vCiSatVars *IntVector, vVar2Index *IntVector, CubeLimit int) *ByteVector {
	panic("TODO, stubbed")
}

// bmcg_sat_solver_print_sop
// C: void bmcg_sat_solver_print_sop( Gia_Man_t * p )
func bmcg_sat_solver_print_sop(p *GiaManager) {
	panic("TODO, stubbed")
}

// bmcg_sat_solver_print_sop_lit
// C: void bmcg_sat_solver_print_sop_lit( Gia_Man_t * p, int Lit )
func bmcg_sat_solver_print_sop_lit(p *GiaManager, Lit int) {
	panic("TODO, stubbed")
}

// bmcg_sat_generate_dvars
// C: void bmcg_sat_generate_dvars( Vec_Int_t * vCiVars, Vec_Str_t * vSop, Vec_Int_t * vDLits )
func bmcg_sat_generate_dvars(vCiVars *IntVector, vSop *ByteVector, vDLits *IntVector) {
	panic("TODO, stubbed")
}

// bmcg_sat_solver_quantify2
// C: int bmcg_sat_solver_quantify2( Gia_Man_t * p, int iLit, int fHash, int(*pFuncCiToKeep)(void *, int), void * pData, Vec_Int_t * vDLits )
func bmcg_sat_solver_quantify2(p *GiaManager, iLit int, fHash int, pFuncCiToKeep func(any, int) int, pData any, vDLits *IntVector) int {
	panic("TODO, stubbed")
}

// Gia_ManSatAndCollect_rec
// C: int Gia_ManSatAndCollect_rec( Gia_Man_t * p, int iObj, Vec_Int_t * vObjsUsed, Vec_Int_t * vCiVars )
func (p *GiaManager) GiaManSatAndCollect_rec(iObj int, vObjsUsed *IntVector, vCiVars *IntVector) int {
	panic("TODO, stubbed")
}

// Gia_ManQuantLoadCnf
// C: void Gia_ManQuantLoadCnf( Gia_Man_t * p, Vec_Int_t * vObjsUsed, bmcg_sat_solver * pSats[] )
func (p *GiaManager) GiaManQuantLoadCnf(vObjsUsed *IntVector, pSats []BmcgSatSolver) {
	panic("TODO, stubbed")
}

// Gia_ManFactorSop
// C: int Gia_ManFactorSop( Gia_Man_t * p, Vec_Int_t * vCiObjIds, Vec_Str_t * vSop, int fHash )
func (p *GiaManager) GiaManFactorSop(vCiObjIds *IntVector, vSop *ByteVector, fHash int) int {
	panic("TODO, stubbed")
}

// Gia_ManCiIsToKeep
// C: int Gia_ManCiIsToKeep( void * pData, int i )
func GiaManCiIsToKeep(pData any, i int) int {
	panic("TODO, stubbed")
}

// Glucose_QuantifyAigTest
// C: void Glucose_QuantifyAigTest( Gia_Man_t * p )
func (p *GiaManager) GlucoseQuantifyAigTest() {
	panic("TODO, stubbed")
}

// bmcg_sat_solver_quantify_test
// C: int bmcg_sat_solver_quantify_test( bmcg_sat_solver * pSats[], Gia_Man_t * p, int iLit, int fHash, int(*pFuncCiToKeep)(void *, int), void * pData, Vec_Int_t * vDLits )
func bmcg_sat_solver_quantify_test(pSats []BmcgSatSolver, p *GiaManager, iLit int, fHash int, pFuncCiToKeep func(any, int) int, pData any, vDLits *IntVector) int {
	panic("TODO, stubbed")
}

// Glucose_CheckTwoNodesTest
// C: void Glucose_CheckTwoNodesTest( Gia_Man_t * p )
func (p *GiaManager) GlucoseCheckTwoNodesTest() {
	panic("TODO, stubbed")
}
