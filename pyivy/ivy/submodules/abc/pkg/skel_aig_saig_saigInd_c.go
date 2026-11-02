package abc

// Hand-checked skeleton of src/aig/saig/saigInd.c.

// Saig_ManStatesAreEqual
// C: int Saig_ManStatesAreEqual( sat_solver * pSat, Vec_Int_t * vState, int nRegs, int i, int k )
func (pSat *SatSolver) SaigManStatesAreEqual(vState *IntVector, nRegs int, i int, k int) int {
	panic("TODO, stubbed")
}

// Saig_ManAddUniqueness
// C: int Saig_ManAddUniqueness( sat_solver * pSat, Vec_Int_t * vState, int nRegs, int i, int k, int * pnSatVarNum, int * pnClauses, int fVerbose )
func (pSat *SatSolver) SaigManAddUniqueness(vState *IntVector, nRegs int, i int, k int, pnSatVarNum *int, pnClauses *int, fVerbose int) int {
	panic("TODO, stubbed")
}

// Saig_ManInduction
// C: int Saig_ManInduction( Aig_Man_t * p, int nTimeOut, int nFramesMax, int nConfMax, int fUnique, int fUniqueAll, int fGetCex, int fVerbose, int fVeryVerbose )
func (p *AigManager) SaigManInduction(nTimeOut int, nFramesMax int, nConfMax int, fUnique int, fUniqueAll int, fGetCex int, fVerbose int, fVeryVerbose int) int {
	panic("TODO, stubbed")
}
