package abc

// Hand-checked skeleton of src/opt/sbd/sbdWin.c.

// Sbd_ManSatSolver
// C: sat_solver * Sbd_ManSatSolver( sat_solver * pSat, Gia_Man_t * p, Vec_Int_t * vMirrors, int Pivot, Vec_Int_t * vWinObjs, Vec_Int_t * vObj2Var, Vec_Int_t * vTfo, Vec_Int_t * vRoots, int fQbf )
func SbdManSatSolver(pSat *SatSolver, p *GiaManager, vMirrors *IntVector, Pivot int, vWinObjs *IntVector, vObj2Var *IntVector, vTfo *IntVector, vRoots *IntVector, fQbf int) *SatSolver {
	panic("TODO, stubbed")
}

// Sbd_ManSolve
// C: word Sbd_ManSolve( sat_solver * pSat, int PivotVar, int FreeVar, Vec_Int_t * vDivSet, Vec_Int_t * vDivVars, Vec_Int_t * vDivValues, Vec_Int_t * vTemp )
func SbdManSolve(pSat *SatSolver, PivotVar int, FreeVar int, vDivSet *IntVector, vDivVars *IntVector, vDivValues *IntVector, vTemp *IntVector) uint64 {
	panic("TODO, stubbed")
}

// Sbd_ManSolve2
// C: int Sbd_ManSolve2( sat_solver * pSat, int PivotVar, int FreeVar, Vec_Int_t * vDivVars, Vec_Int_t * vDivValues, Vec_Int_t * vTemp, Vec_Int_t * vSop )
func SbdManSolve2(pSat *SatSolver, PivotVar int, FreeVar int, vDivVars *IntVector, vDivValues *IntVector, vTemp *IntVector, vSop *IntVector) int {
	panic("TODO, stubbed")
}

// Sbd_ManSolverSupp (pnVars is an out parameter)
// C: word Sbd_ManSolverSupp( Vec_Int_t * vSop, int * pInds, int * pnVars )
func SbdManSolverSupp(vSop *IntVector, pInds []int, pnVars *int) uint64 {
	panic("TODO, stubbed")
}

// Sbd_ManSolverPrint
// C: void Sbd_ManSolverPrint( Vec_Int_t * vSop )
func SbdManSolverPrint(vSop *IntVector) {
	panic("TODO, stubbed")
}

// Sbd_ManSolveSelect
// C: void Sbd_ManSolveSelect( Gia_Man_t * p, Vec_Int_t * vMirrors, int Pivot, Vec_Int_t * vDivVars, Vec_Int_t * vDivValues, Vec_Int_t * vWinObjs, Vec_Int_t * vObj2Var, Vec_Int_t * vTfo, Vec_Int_t * vRoots )
func (p *GiaManager) SbdManSolveSelect(vMirrors *IntVector, Pivot int, vDivVars *IntVector, vDivValues *IntVector, vWinObjs *IntVector, vObj2Var *IntVector, vTfo *IntVector, vRoots *IntVector) {
	panic("TODO, stubbed")
}

// sat_solver_random_polarity (static)
// C: void sat_solver_random_polarity(sat_solver* s)
func satSolverRandomPolarity(s *SatSolver) {
	panic("TODO, stubbed")
}

// Sbd_ManCollectConstants
// C: int Sbd_ManCollectConstants( sat_solver * pSat, int nCareMints[2], int PivotVar, word * pVarSims[], Vec_Int_t * vInds )
func SbdManCollectConstants(pSat *SatSolver, nCareMints *[2]int, PivotVar int, pVarSims [][]uint64, vInds *IntVector) int {
	panic("TODO, stubbed")
}

// Sbd_ManCollectConstantsNew
// C: int Sbd_ManCollectConstantsNew( sat_solver * pSat, Vec_Int_t * vDivVars, int nConsts, int PivotVar, word * pOnset, word * pOffset )
func SbdManCollectConstantsNew(pSat *SatSolver, vDivVars *IntVector, nConsts int, PivotVar int, pOnset []uint64, pOffset []uint64) int {
	panic("TODO, stubbed")
}
