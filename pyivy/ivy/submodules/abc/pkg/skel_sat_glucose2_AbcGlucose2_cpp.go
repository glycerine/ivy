package abc

// Hand-checked skeleton of src/sat/glucose2/AbcGlucose2.cpp (namespace Gluco2).
//
// `#define USE_SIMP_SOLVER 1` is set at the top of the file, so only the SimpSolver variant of the
// glucose-solver wrappers is live (the `#else` branch is dead and not stubbed). The extern "C"
// bmcg2_sat_solver_* entry points are stubbed in skel_sat_glucose2_AbcGlucose2_h.go next to their
// declarations. Functions whose first parameter is a SimpSolver stay free functions (they are not
// class members).

// glucose2_solver_start
// glucose2_solver_stop
// glucose2_solver_reset
// glucose2_solver_addclause
// glucose2_solver_setcallback
// glucose2_solver_solve
// glucose2_solver_addvar
// glucose2_solver_read_cex
// glucose2_solver_read_cex_varvalue
// glucose2_solver_setstop
// glucose2_print_stats
// Glucose_ReadDimacs
// Glucose_SolverFromAig
// C: Vec_Int_t * Glucose_SolverFromAig( Gia_Man_t * p, SimpSolver& s )
func (p *GiaManager) GlucoseSolverFromAig_glucose2(s *Gluco2SimpSolver) *IntVector {
	panic("TODO, stubbed")
}

// Glucose_SolverFromAig2
// C: Vec_Int_t * Glucose_SolverFromAig2( Gia_Man_t * p, SimpSolver& S )
func (p *GiaManager) GlucoseSolverFromAig2_glucose2(S *Gluco2SimpSolver) *IntVector {
	panic("TODO, stubbed")
}

// Glucose2_GenerateCubes
// C: Vec_Str_t * Glucose2_GenerateCubes( bmcg2_sat_solver * pSat[2], Vec_Int_t * vCiSatVars, Vec_Int_t * vVar2Index, int CubeLimit )
func Glucose2GenerateCubes(pSat []Bmcg2SatSolver, vCiSatVars *IntVector, vVar2Index *IntVector, CubeLimit int) *ByteVector {
	panic("TODO, stubbed")
}

// bmcg2_sat_solver_print_sop
// C: void bmcg2_sat_solver_print_sop( Gia_Man_t * p )
func bmcg2_sat_solver_print_sop(p *GiaManager) {
	panic("TODO, stubbed")
}

// bmcg2_sat_solver_print_sop_lit
// C: void bmcg2_sat_solver_print_sop_lit( Gia_Man_t * p, int Lit )
func bmcg2_sat_solver_print_sop_lit(p *GiaManager, Lit int) {
	panic("TODO, stubbed")
}

// bmcg2_sat_generate_dvars
// C: void bmcg2_sat_generate_dvars( Vec_Int_t * vCiVars, Vec_Str_t * vSop, Vec_Int_t * vDLits )
func bmcg2_sat_generate_dvars(vCiVars *IntVector, vSop *ByteVector, vDLits *IntVector) {
	panic("TODO, stubbed")
}

// bmcg2_sat_solver_quantify2
// C: int bmcg2_sat_solver_quantify2( Gia_Man_t * p, int iLit, int fHash, int(*pFuncCiToKeep)(void *, int), void * pData, Vec_Int_t * vDLits )
func bmcg2_sat_solver_quantify2(p *GiaManager, iLit int, fHash int, pFuncCiToKeep func(any, int) int, pData any, vDLits *IntVector) int {
	panic("TODO, stubbed")
}

// Gia_ManSatAndCollect2_rec
// C: int Gia_ManSatAndCollect2_rec( Gia_Man_t * p, int iObj, Vec_Int_t * vObjsUsed, Vec_Int_t * vCiVars )
func (p *GiaManager) GiaManSatAndCollect2_rec(iObj int, vObjsUsed *IntVector, vCiVars *IntVector) int {
	panic("TODO, stubbed")
}

// Gia_ManQuantLoadCnf2
// C: void Gia_ManQuantLoadCnf2( Gia_Man_t * p, Vec_Int_t * vObjsUsed, bmcg2_sat_solver * pSats[] )
func (p *GiaManager) GiaManQuantLoadCnf2(vObjsUsed *IntVector, pSats []Bmcg2SatSolver) {
	panic("TODO, stubbed")
}

// Gia_ManFactorSop2
// C: int Gia_ManFactorSop2( Gia_Man_t * p, Vec_Int_t * vCiObjIds, Vec_Str_t * vSop, int fHash )
func (p *GiaManager) GiaManFactorSop2(vCiObjIds *IntVector, vSop *ByteVector, fHash int) int {
	panic("TODO, stubbed")
}

// Gia_ManCiIsToKeep2
// C: int Gia_ManCiIsToKeep2( void * pData, int i )
func GiaManCiIsToKeep2(pData any, i int) int {
	panic("TODO, stubbed")
}

// Glucose2_QuantifyAigTest
// C: void Glucose2_QuantifyAigTest( Gia_Man_t * p )
func (p *GiaManager) Glucose2QuantifyAigTest() {
	panic("TODO, stubbed")
}

// bmcg2_sat_solver_quantify_test
// C: int bmcg2_sat_solver_quantify_test( bmcg2_sat_solver * pSats[], Gia_Man_t * p, int iLit, int fHash, int(*pFuncCiToKeep)(void *, int), void * pData, Vec_Int_t * vDLits )
func bmcg2_sat_solver_quantify_test(pSats []Bmcg2SatSolver, p *GiaManager, iLit int, fHash int, pFuncCiToKeep func(any, int) int, pData any, vDLits *IntVector) int {
	panic("TODO, stubbed")
}

// Glucose2_CheckTwoNodesTest
// C: void Glucose2_CheckTwoNodesTest( Gia_Man_t * p )
func (p *GiaManager) Glucose2CheckTwoNodesTest() {
	panic("TODO, stubbed")
}
