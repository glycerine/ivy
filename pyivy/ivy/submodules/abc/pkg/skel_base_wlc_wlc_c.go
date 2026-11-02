package abc

// Hand-checked skeleton of src/base/wlc/wlc.c.

// Wlc_GenerateCodeMax4
// C: void Wlc_GenerateCodeMax4( int nBits )
func WlcGenerateCodeMax4(nBits int) {
	panic("TODO, stubbed")
}

// Wlc_BlastFullAdderCtrlCnf (pc, ps and piVars are in/out scalars)
// C: void Wlc_BlastFullAdderCtrlCnf( sat_solver * pSat, int a, int ac, int b, int c, int * pc, int * ps, int * piVars )
func WlcBlastFullAdderCtrlCnf(pSat *SatSolver, a int, ac int, b int, c int, pc *int, ps *int, piVars *int) {
	panic("TODO, stubbed")
}

// Wlc_BlastMultiplierCnf (piVars is an in/out scalar)
// C: void Wlc_BlastMultiplierCnf( sat_solver * pSat, int * pArgA, int * pArgB, int nArgA, int nArgB, Vec_Int_t * vTemp, Vec_Int_t * vRes, int * piVars )
func WlcBlastMultiplierCnf(pSat *SatSolver, pArgA []int, pArgB []int, nArgA int, nArgB int, vTemp *IntVector, vRes *IntVector, piVars *int) {
	panic("TODO, stubbed")
}

// Wlc_BlastMultiplierCnfMain
// C: sat_solver * Wlc_BlastMultiplierCnfMain( int nBits )
func WlcBlastMultiplierCnfMain(nBits int) *SatSolver {
	panic("TODO, stubbed")
}

// Wlc_BlastMultiplierCnfTest
// C: void Wlc_BlastMultiplierCnfTest( int nBits )
func WlcBlastMultiplierCnfTest(nBits int) {
	panic("TODO, stubbed")
}
