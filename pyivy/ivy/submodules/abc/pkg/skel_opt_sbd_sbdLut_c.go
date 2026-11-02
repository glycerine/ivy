package abc

// Hand-checked skeleton of src/opt/sbd/sbdLut.c.

// Sbd_ProblemCountParams
// C: int Sbd_ProblemCountParams( int nStrs, Sbd_Str_t * pStr0 )
func SbdProblemCountParams(nStrs int, pStr0 []SbdStr) int {
	panic("TODO, stubbed")
}

// Sbd_ProblemAddClauses
// C: int Sbd_ProblemAddClauses( sat_solver * pSat, int nVars, int nStrs, int * pVars, Sbd_Str_t * pStr0 )
func SbdProblemAddClauses(pSat *SatSolver, nVars int, nStrs int, pVars []int, pStr0 []SbdStr) int {
	panic("TODO, stubbed")
}

// Sbd_ProblemAddClausesInit
// C: void Sbd_ProblemAddClausesInit( sat_solver * pSat, int nVars, int nStrs, int * pVars, Sbd_Str_t * pStr0 )
func SbdProblemAddClausesInit(pSat *SatSolver, nVars int, nStrs int, pVars []int, pStr0 []SbdStr) {
	panic("TODO, stubbed")
}

// Sbd_ProblemPrintSolution
// C: void Sbd_ProblemPrintSolution( int nStrs, Sbd_Str_t * pStr0, Vec_Int_t * vLits )
func SbdProblemPrintSolution(nStrs int, pStr0 []SbdStr, vLits *IntVector) {
	panic("TODO, stubbed")
}

// Sbd_ProblemCollectSolution
// C: void Sbd_ProblemCollectSolution( int nStrs, Sbd_Str_t * pStr0, Vec_Int_t * vLits )
func SbdProblemCollectSolution(nStrs int, pStr0 []SbdStr, vLits *IntVector) {
	panic("TODO, stubbed")
}

// Sbd_ProblemSolve
// C: int Sbd_ProblemSolve( Gia_Man_t * p, Vec_Int_t * vMirrors, int Pivot, Vec_Int_t * vWinObjs, Vec_Int_t * vObj2Var, Vec_Int_t * vTfo, Vec_Int_t * vRoots, Vec_Int_t * vDivSet, int nStrs, Sbd_Str_t * pStr0 )
func (p *GiaManager) SbdProblemSolve(vMirrors *IntVector, Pivot int, vWinObjs *IntVector, vObj2Var *IntVector, vTfo *IntVector, vRoots *IntVector, vDivSet *IntVector, nStrs int, pStr0 []SbdStr) int {
	panic("TODO, stubbed")
}
