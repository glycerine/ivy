package abc

// Hand-checked skeleton of src/aig/saig/saigTempor.c.

// Saig_ManTemporFrames
// C: Aig_Man_t * Saig_ManTemporFrames( Aig_Man_t * pAig, int nFrames )
func (pAig *AigManager) SaigManTemporFrames(nFrames int) *AigManager {
	panic("TODO, stubbed")
}

// Saig_ManTemporDecompose
// C: Aig_Man_t * Saig_ManTemporDecompose( Aig_Man_t * pAig, int nFrames )
func (pAig *AigManager) SaigManTemporDecompose(nFrames int) *AigManager {
	panic("TODO, stubbed")
}

// Vec_IntLastNonZeroBeforeLimit
// C: int Vec_IntLastNonZeroBeforeLimit( Vec_Int_t * vTemp, int Limit )
func VecIntLastNonZeroBeforeLimit(vTemp *IntVector, Limit int) int {
	panic("TODO, stubbed")
}

// Saig_ManTempor
// C: Aig_Man_t * Saig_ManTempor( Aig_Man_t * pAig, int nFrames, int TimeOut, int nConfLimit, int fUseBmc, int fUseTransSigs, int fVerbose, int fVeryVerbose )
func (pAig *AigManager) SaigManTempor(nFrames int, TimeOut int, nConfLimit int, fUseBmc int, fUseTransSigs int, fVerbose int, fVeryVerbose int) *AigManager {
	panic("TODO, stubbed")
}
