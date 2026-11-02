package abc

// Hand-checked skeleton of src/proof/ssw/sswPairs.c.

// Ssw_MiterStatus
// C: int Ssw_MiterStatus( Aig_Man_t * p, int fVerbose )
func (p *AigManager) SswMiterStatus(fVerbose int) int {
	panic("TODO, stubbed")
}

// Ssw_TransferSignalPairs
// C: Vec_Int_t * Ssw_TransferSignalPairs( Aig_Man_t * pMiter, Aig_Man_t * pAig1, Aig_Man_t * pAig2, Vec_Int_t * vIds1, Vec_Int_t * vIds2 )
func (pMiter *AigManager) SswTransferSignalPairs(pAig1 *AigManager, pAig2 *AigManager, vIds1 *IntVector, vIds2 *IntVector) *IntVector {
	panic("TODO, stubbed")
}

// Ssw_TransformPairsIntoTempClasses (returns an array of nObjNumMax Vec_Int_t *)
// C: Vec_Int_t ** Ssw_TransformPairsIntoTempClasses( Vec_Int_t * vPairs, int nObjNumMax )
func SswTransformPairsIntoTempClasses(vPairs *IntVector, nObjNumMax int) []*IntVector {
	panic("TODO, stubbed")
}

// Ssw_FreeTempClasses
// C: void Ssw_FreeTempClasses( Vec_Int_t ** pvClasses, int nObjNumMax )
func SswFreeTempClasses(pvClasses []*IntVector, nObjNumMax int) {
	panic("TODO, stubbed")
}

// Ssw_SignalCorrespondenceWithPairs
// C: Aig_Man_t * Ssw_SignalCorrespondenceWithPairs( Aig_Man_t * pAig1, Aig_Man_t * pAig2, Vec_Int_t * vIds1, Vec_Int_t * vIds2, Ssw_Pars_t * pPars )
func (pAig1 *AigManager) SswSignalCorrespondenceWithPairs(pAig2 *AigManager, vIds1 *IntVector, vIds2 *IntVector, pPars *SswPars) *AigManager {
	panic("TODO, stubbed")
}

// Ssw_SignalCorrespondeceTestPairs
// C: Aig_Man_t * Ssw_SignalCorrespondeceTestPairs( Aig_Man_t * pAig )
func (pAig *AigManager) SswSignalCorrespondeceTestPairs() *AigManager {
	panic("TODO, stubbed")
}

// Ssw_SecWithPairs
// C: int Ssw_SecWithPairs( Aig_Man_t * pAig1, Aig_Man_t * pAig2, Vec_Int_t * vIds1, Vec_Int_t * vIds2, Ssw_Pars_t * pPars )
func (pAig1 *AigManager) SswSecWithPairs(pAig2 *AigManager, vIds1 *IntVector, vIds2 *IntVector, pPars *SswPars) int {
	panic("TODO, stubbed")
}

// Ssw_SecGeneral
// C: int Ssw_SecGeneral( Aig_Man_t * pAig1, Aig_Man_t * pAig2, Ssw_Pars_t * pPars )
func (pAig1 *AigManager) SswSecGeneral(pAig2 *AigManager, pPars *SswPars) int {
	panic("TODO, stubbed")
}

// Ssw_SecGeneralMiter
// C: int Ssw_SecGeneralMiter( Aig_Man_t * pMiter, Ssw_Pars_t * pPars )
func (pMiter *AigManager) SswSecGeneralMiter(pPars *SswPars) int {
	panic("TODO, stubbed")
}
