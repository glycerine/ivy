package abc

// Hand-checked skeleton of src/aig/aig/aigRepar.c.

// Aig_ManInterAddBuffer (static)
// C: void Aig_ManInterAddBuffer( sat_solver2 * pSat, int iVarA, int iVarB, int fCompl, int fMark )
func aigManInterAddBuffer(pSat *SatSolver2, iVarA int, iVarB int, fCompl int, fMark int) {
	panic("TODO, stubbed")
}

// Aig_ManInterAddXor (static)
// C: void Aig_ManInterAddXor( sat_solver2 * pSat, int iVarA, int iVarB, int iVarC, int fCompl, int fMark )
func aigManInterAddXor(pSat *SatSolver2, iVarA int, iVarB int, iVarC int, fCompl int, fMark int) {
	panic("TODO, stubbed")
}

// Aig_ManInterTest
// C: void Aig_ManInterTest( Aig_Man_t * pMan, int fVerbose )
func (pMan *AigManager) AigManInterTest(fVerbose int) {
	panic("TODO, stubbed")
}

// Aig_ManAppend
// C: void Aig_ManAppend( Aig_Man_t * pBase, Aig_Man_t * pNew )
func (pBase *AigManager) AigManAppend(pNew *AigManager) {
	panic("TODO, stubbed")
}

// Aig_ManInterRepar
// C: Aig_Man_t * Aig_ManInterRepar( Aig_Man_t * pMan, int fVerbose )
func (pMan *AigManager) AigManInterRepar(fVerbose int) *AigManager {
	panic("TODO, stubbed")
}
