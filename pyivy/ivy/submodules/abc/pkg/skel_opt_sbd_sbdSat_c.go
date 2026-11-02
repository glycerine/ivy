package abc

// Hand-checked skeleton of src/opt/sbd/sbdSat.c.

// MAX_M is 8, MAX_N is 30, MAX_K is 6, MAX_D is 8, SBD_LUTS_MAX is 2,
// SBD_SIZE_MAX is 4 and SBD_DIV_MAX is 10. Multi-dimensional array parameters
// decay to pointers to arrays and are written as such.

// SbdPro is Sbd_Pro_t.
type SbdPro struct {
	nLuts  int            // LUT count
	nSize  int            // LUT size
	nDivs  int            // divisor count
	nVars  int            // intermediate variables (nLuts * nSize)
	nPars  int            // total parameter count (nLuts * (1 << nSize) + nLuts * nSize * nDivs)
	pPars1 [2][1 << 4]int // lut parameters
	pPars2 [2][4][10]int  // topo parameters
	pVars  [2][4 + 1]int  // internal variables
	pDivs  [10]int        // divisor variables (external)
}

// Sbd_ProblemSetup
// C: Vec_Int_t * Sbd_ProblemSetup( Sbd_Pro_t * p, int nLuts, int nSize, int nDivs )
func (p *SbdPro) SbdProblemSetup(nLuts int, nSize int, nDivs int) *IntVector {
	panic("TODO, stubbed")
}

// Sbd_ProblemLoad1
// C: void Sbd_ProblemLoad1( Sbd_Pro_t * p, Vec_Int_t * vCnf, int iStartVar, int * pDivVars, int iTopVar, sat_solver * pSat )
func (p *SbdPro) SbdProblemLoad1(vCnf *IntVector, iStartVar int, pDivVars []int, iTopVar int, pSat *SatSolver) {
	panic("TODO, stubbed")
}

// Sbd_ProblemLoad2
// C: void Sbd_ProblemLoad2( Sbd_Pro_t * p, Vec_Wec_t * vCnf, int iStartVar, int * pDivVarValues, int iTopVarValue, sat_solver * pSat )
func (p *SbdPro) SbdProblemLoad2(vCnf *IntLevelVector, iStartVar int, pDivVarValues []int, iTopVarValue int, pSat *SatSolver) {
	panic("TODO, stubbed")
}

// Sbd_SolverTopo (pnVars is an out parameter)
// C: sat_solver * Sbd_SolverTopo( int M, int N, int K, int pVars[MAX_N][MAX_M+MAX_N][MAX_K], int pVars2[MAX_M+MAX_N][MAX_D], int pDelays[], int Req, int * pnVars )
func SbdSolverTopo(M int, N int, K int, pVars *[30][38][6]int, pVars2 *[38][8]int, pDelays []int, Req int, pnVars *int) *SatSolver {
	panic("TODO, stubbed")
}

// Sbd_SolverTopoPrint
// C: void Sbd_SolverTopoPrint( sat_solver * pSat, int M, int N, int K, int pVars[MAX_N][MAX_M+MAX_N][MAX_K] )
func SbdSolverTopoPrint(pSat *SatSolver, M int, N int, K int, pVars *[30][38][6]int) {
	panic("TODO, stubbed")
}

// Sbd_SolverTopoTest
// C: void Sbd_SolverTopoTest()
func SbdSolverTopoTest() {
	panic("TODO, stubbed")
}

// Sbd_SolverSynth
// C: void Sbd_SolverSynth( int M, int N, int K, int pLuts[MAX_N][MAX_K] )
func SbdSolverSynth(M int, N int, K int, pLuts *[30][6]int) {
	panic("TODO, stubbed")
}

// Sbd_SolverTruth
// C: word Sbd_SolverTruth( int M, int N, int K, int pLuts[MAX_N][MAX_K], int pValues[MAX_N*((1<<MAX_K)-1)] )
func SbdSolverTruth(M int, N int, K int, pLuts *[30][6]int, pValues []int) uint64 {
	panic("TODO, stubbed")
}

// Sbd_SolverTruthWord (returns a pointer to a truth-table word array)
// C: word * Sbd_SolverTruthWord( int M, int N, int K, int pLuts[MAX_N][MAX_K], int pValues[MAX_N*((1<<MAX_K)-1)], word * pTruthsElem, int fCompl )
func SbdSolverTruthWord(M int, N int, K int, pLuts *[30][6]int, pValues []int, pTruthsElem []uint64, fCompl int) []uint64 {
	panic("TODO, stubbed")
}

// Sbd_SolverFunc
// C: int Sbd_SolverFunc( int M, int N, int K, int pLuts[MAX_N][MAX_K], word * pTruthInit, int * pValues )
func SbdSolverFunc(M int, N int, K int, pLuts *[30][6]int, pTruthInit []uint64, pValues []int) int {
	panic("TODO, stubbed")
}

// Sbd_SolverFuncTest
// C: void Sbd_SolverFuncTest()
func SbdSolverFuncTest() {
	panic("TODO, stubbed")
}
