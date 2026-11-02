package abc

// Hand-checked skeleton of src/base/abci/abcVerify.c.

// Abc_NtkCecSat
// C: void Abc_NtkCecSat( Abc_Ntk_t * pNtk1, Abc_Ntk_t * pNtk2, int nConfLimit, int nInsLimit )
func (pNtk1 *ABCNetwork) AbcNtkCecSat(pNtk2 *ABCNetwork, nConfLimit int, nInsLimit int) {
	panic("TODO, stubbed")
}

// Abc_NtkCecFraig
// C: void Abc_NtkCecFraig( Abc_Ntk_t * pNtk1, Abc_Ntk_t * pNtk2, int nSeconds, int fVerbose )
func (pNtk1 *ABCNetwork) AbcNtkCecFraig(pNtk2 *ABCNetwork, nSeconds int, fVerbose int) {
	panic("TODO, stubbed")
}

// Abc_NtkCecFraigPart
// C: void Abc_NtkCecFraigPart( Abc_Ntk_t * pNtk1, Abc_Ntk_t * pNtk2, int nSeconds, int nPartSize, int fVerbose )
func (pNtk1 *ABCNetwork) AbcNtkCecFraigPart(pNtk2 *ABCNetwork, nSeconds int, nPartSize int, fVerbose int) {
	panic("TODO, stubbed")
}

// Abc_NtkCecFraigPartAuto
// C: void Abc_NtkCecFraigPartAuto( Abc_Ntk_t * pNtk1, Abc_Ntk_t * pNtk2, int nSeconds, int fVerbose )
func (pNtk1 *ABCNetwork) AbcNtkCecFraigPartAuto(pNtk2 *ABCNetwork, nSeconds int, fVerbose int) {
	panic("TODO, stubbed")
}

// Abc_NtkSecSat
// C: void Abc_NtkSecSat( Abc_Ntk_t * pNtk1, Abc_Ntk_t * pNtk2, int nConfLimit, int nInsLimit, int nFrames )
func (pNtk1 *ABCNetwork) AbcNtkSecSat(pNtk2 *ABCNetwork, nConfLimit int, nInsLimit int, nFrames int) {
	panic("TODO, stubbed")
}

// Abc_NtkSecFraig
// C: int Abc_NtkSecFraig( Abc_Ntk_t * pNtk1, Abc_Ntk_t * pNtk2, int nSeconds, int nFrames, int fVerbose )
func (pNtk1 *ABCNetwork) AbcNtkSecFraig(pNtk2 *ABCNetwork, nSeconds int, nFrames int, fVerbose int) int {
	panic("TODO, stubbed")
}

// Abc_NtkVerifyGetCleanModel
// C: int * Abc_NtkVerifyGetCleanModel( Abc_Ntk_t * pNtk, int nFrames )
func (pNtk *ABCNetwork) AbcNtkVerifyGetCleanModel(nFrames int) []int {
	panic("TODO, stubbed")
}

// Abc_NtkVerifySimulatePattern
// C: int * Abc_NtkVerifySimulatePattern( Abc_Ntk_t * pNtk, int * pModel )
func (pNtk *ABCNetwork) AbcNtkVerifySimulatePattern(pModel []int) []int {
	panic("TODO, stubbed")
}

// Abc_NtkVerifyReportError (static)
// C: void Abc_NtkVerifyReportError( Abc_Ntk_t * pNtk1, Abc_Ntk_t * pNtk2, int * pModel )
func (pNtk1 *ABCNetwork) abcNtkVerifyReportError(pNtk2 *ABCNetwork, pModel []int) {
	panic("TODO, stubbed")
}

// Abc_NtkGetSeqPoSupp
// C: void Abc_NtkGetSeqPoSupp( Abc_Ntk_t * pNtk, int iFrame, int iNumPo )
func (pNtk *ABCNetwork) AbcNtkGetSeqPoSupp(iFrame int, iNumPo int) {
	panic("TODO, stubbed")
}

// Abc_NtkVerifyReportErrorSeq
// C: void Abc_NtkVerifyReportErrorSeq( Abc_Ntk_t * pNtk1, Abc_Ntk_t * pNtk2, int * pModel, int nFrames )
func (pNtk1 *ABCNetwork) AbcNtkVerifyReportErrorSeq(pNtk2 *ABCNetwork, pModel []int, nFrames int) {
	panic("TODO, stubbed")
}

// Abc_NtkSimulteBuggyMiter
// C: void Abc_NtkSimulteBuggyMiter( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkSimulteBuggyMiter() {
	panic("TODO, stubbed")
}

// Abc_NtkIsTrueCex
// C: int Abc_NtkIsTrueCex( Abc_Ntk_t * pNtk, Abc_Cex_t * pCex )
func (pNtk *ABCNetwork) AbcNtkIsTrueCex(pCex *Counterexample) int {
	panic("TODO, stubbed")
}

// Abc_NtkIsValidCex
// C: int Abc_NtkIsValidCex( Abc_Ntk_t * pNtk, Abc_Cex_t * pCex )
func (pNtk *ABCNetwork) AbcNtkIsValidCex(pCex *Counterexample) int {
	panic("TODO, stubbed")
}
