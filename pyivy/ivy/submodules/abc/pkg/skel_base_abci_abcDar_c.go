package abc

// Hand-checked skeleton of src/base/abci/abcDar.c.

// AmapPar is Amap_Par_t (src/map/amap/amap.h).
type AmapPar struct {
	nIterFlow int     // iterations of area flow
	nIterArea int     // iteratoins of exact area
	nCutsMax  int     // the maximum number of cuts at a node
	fUseMuxes int     // enables the use of MUXes
	fUseXors  int     // enables the use of XORs
	fFreeInvs int     // assume inverters are free (area = 0)
	fEpsilon  float32 // used to compare floating point numbers
	fADratio  float32 // ratio of area/delay improvement
	fVerbose  int     // verbosity flag
}

// Abc_ObjCompareById
// C: int Abc_ObjCompareById( Abc_Obj_t ** pp1, Abc_Obj_t ** pp2 )
func AbcObjCompareById(pp1 **ABCNetworkObject, pp2 **ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_CollectTopOr_rec
// C: void Abc_CollectTopOr_rec( Abc_Obj_t * pObj, Vec_Ptr_t * vSuper )
func AbcCollectTopOr_rec(pObj *ABCNetworkObject, vSuper *PointerVector[ABCNetworkObject]) {
	panic("TODO, stubbed")
}

// Abc_CollectTopOr
// C: void Abc_CollectTopOr( Abc_Obj_t * pObj, Vec_Ptr_t * vSuper )
func AbcCollectTopOr(pObj *ABCNetworkObject, vSuper *PointerVector[ABCNetworkObject]) {
	panic("TODO, stubbed")
}

// Abc_NtkToDarBmc
// C: Aig_Man_t * Abc_NtkToDarBmc( Abc_Ntk_t * pNtk, Vec_Int_t ** pvMap )
func (pNtk *ABCNetwork) AbcNtkToDarBmc(pvMap **IntVector) *AigManager {
	panic("TODO, stubbed")
}

// Abc_NtkFindDcLatches
// C: Vec_Int_t * Abc_NtkFindDcLatches( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkFindDcLatches() *IntVector {
	panic("TODO, stubbed")
}

// Abc_NtkToDar
// C: Aig_Man_t * Abc_NtkToDar( Abc_Ntk_t * pNtk, int fExors, int fRegisters )
func (pNtk *ABCNetwork) AbcNtkToDar(fExors int, fRegisters int) *AigManager {
	panic("TODO, stubbed")
}

// Abc_NtkToDarChoices
// C: Aig_Man_t * Abc_NtkToDarChoices( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkToDarChoices() *AigManager {
	panic("TODO, stubbed")
}

// Abc_NtkFromDar
// C: Abc_Ntk_t * Abc_NtkFromDar( Abc_Ntk_t * pNtkOld, Aig_Man_t * pMan )
func (pNtkOld *ABCNetwork) AbcNtkFromDar(pMan *AigManager) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkFromDarSeqSweep
// C: Abc_Ntk_t * Abc_NtkFromDarSeqSweep( Abc_Ntk_t * pNtkOld, Aig_Man_t * pMan )
func (pNtkOld *ABCNetwork) AbcNtkFromDarSeqSweep(pMan *AigManager) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkFromAigPhase
// C: Abc_Ntk_t * Abc_NtkFromAigPhase( Aig_Man_t * pMan )
func AbcNtkFromAigPhase(pMan *AigManager) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_ObjHopFromGia_rec
// C: Hop_Obj_t * Abc_ObjHopFromGia_rec( Hop_Man_t * pHopMan, Gia_Man_t * p, int Id, Vec_Ptr_t * vCopies )
func AbcObjHopFromGia_rec(pHopMan *HopManager, p *GiaManager, Id int, vCopies *PointerVector[HopObject]) *HopObject {
	panic("TODO, stubbed")
}

// Abc_ObjHopFromGia
// C: Hop_Obj_t * Abc_ObjHopFromGia( Hop_Man_t * pHopMan, Gia_Man_t * p, int GiaId, Vec_Ptr_t * vCopies )
func AbcObjHopFromGia(pHopMan *HopManager, p *GiaManager, GiaId int, vCopies *PointerVector[HopObject]) *HopObject {
	panic("TODO, stubbed")
}

// Abc_NtkFromMappedGia_rec
// C: Abc_Obj_t * Abc_NtkFromMappedGia_rec( Abc_Ntk_t * pNtkNew, Gia_Man_t * p, int iObj, int fAddInv )
func (pNtkNew *ABCNetwork) AbcNtkFromMappedGia_rec(p *GiaManager, iObj int, fAddInv int) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_NtkFromMappedGia
// C: Abc_Ntk_t * Abc_NtkFromMappedGia( Gia_Man_t * p, int fFindEnables, int fUseBuffs )
func (p *GiaManager) AbcNtkFromMappedGia(fFindEnables int, fUseBuffs int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkFromCellWrite (static)
// C: void Abc_NtkFromCellWrite( Vec_Int_t * vCopyLits, int i, int c, int Id )
func abcNtkFromCellWrite(vCopyLits *IntVector, i int, c int, Id int) {
	panic("TODO, stubbed")
}

// Abc_NtkFromCellRead (static)
// C: Abc_Obj_t * Abc_NtkFromCellRead( Abc_Ntk_t * p, Vec_Int_t * vCopyLits, int i, int c )
func (p *ABCNetwork) abcNtkFromCellRead(vCopyLits *IntVector, i int, c int) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_NtkFromCellMappedGia
// C: Abc_Ntk_t * Abc_NtkFromCellMappedGia( Gia_Man_t * p, int fUseBuffs )
func (p *GiaManager) AbcNtkFromCellMappedGia(fUseBuffs int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkAfterTrim
// C: Abc_Ntk_t * Abc_NtkAfterTrim( Aig_Man_t * pMan, Abc_Ntk_t * pNtkOld )
func AbcNtkAfterTrim(pMan *AigManager, pNtkOld *ABCNetwork) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkFromDarChoices
// C: Abc_Ntk_t * Abc_NtkFromDarChoices( Abc_Ntk_t * pNtkOld, Aig_Man_t * pMan )
func (pNtkOld *ABCNetwork) AbcNtkFromDarChoices(pMan *AigManager) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkFromDarSeq
// C: Abc_Ntk_t * Abc_NtkFromDarSeq( Abc_Ntk_t * pNtkOld, Aig_Man_t * pMan )
func (pNtkOld *ABCNetwork) AbcNtkFromDarSeq(pMan *AigManager) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkCollectCiNames
// C: Vec_Ptr_t * Abc_NtkCollectCiNames( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkCollectCiNames() *PointerVector[string] {
	panic("TODO, stubbed")
}

// Abc_NtkCollectCoNames
// C: Vec_Ptr_t * Abc_NtkCollectCoNames( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkCollectCoNames() *PointerVector[string] {
	panic("TODO, stubbed")
}

// Abc_NtkGetLatchValues
// C: Vec_Int_t * Abc_NtkGetLatchValues( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkGetLatchValues() *IntVector {
	panic("TODO, stubbed")
}

// Abc_NtkDar
// C: Abc_Ntk_t * Abc_NtkDar( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkDar() *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkDarFraig
// C: Abc_Ntk_t * Abc_NtkDarFraig( Abc_Ntk_t * pNtk, int nConfLimit, int fDoSparse, int fProve, int fTransfer, int fSpeculate, int fChoicing, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkDarFraig(nConfLimit int, fDoSparse int, fProve int, fTransfer int, fSpeculate int, fChoicing int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkDarFraigPart
// C: Abc_Ntk_t * Abc_NtkDarFraigPart( Abc_Ntk_t * pNtk, int nPartSize, int nConfLimit, int nLevelMax, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkDarFraigPart(nPartSize int, nConfLimit int, nLevelMax int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkCSweep
// C: Abc_Ntk_t * Abc_NtkCSweep( Abc_Ntk_t * pNtk, int nCutsMax, int nLeafMax, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkCSweep(nCutsMax int, nLeafMax int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkDRewrite
// C: Abc_Ntk_t * Abc_NtkDRewrite( Abc_Ntk_t * pNtk, Dar_RwrPar_t * pPars )
func (pNtk *ABCNetwork) AbcNtkDRewrite(pPars *DarRwrPar) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkDRefactor
// C: Abc_Ntk_t * Abc_NtkDRefactor( Abc_Ntk_t * pNtk, Dar_RefPar_t * pPars )
func (pNtk *ABCNetwork) AbcNtkDRefactor(pPars *DarRefPar) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkDC2
// C: Abc_Ntk_t * Abc_NtkDC2( Abc_Ntk_t * pNtk, int fBalance, int fUpdateLevel, int fFanout, int fPower, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkDC2(fBalance int, fUpdateLevel int, fFanout int, fPower int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkDChoice
// C: Abc_Ntk_t * Abc_NtkDChoice( Abc_Ntk_t * pNtk, int fBalance, int fUpdateLevel, int fConstruct, int nConfMax, int nLevelMax, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkDChoice(fBalance int, fUpdateLevel int, fConstruct int, nConfMax int, nLevelMax int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkDch
// C: Abc_Ntk_t * Abc_NtkDch( Abc_Ntk_t * pNtk, Dch_Pars_t * pPars )
func (pNtk *ABCNetwork) AbcNtkDch(pPars *DchOptions) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkDrwsat
// C: Abc_Ntk_t * Abc_NtkDrwsat( Abc_Ntk_t * pNtk, int fBalance, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkDrwsat(fBalance int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkConstructFromCnf
// C: Abc_Ntk_t * Abc_NtkConstructFromCnf( Abc_Ntk_t * pNtk, Cnf_Man_t * p, Vec_Ptr_t * vMapped )
func (pNtk *ABCNetwork) AbcNtkConstructFromCnf(p *CnfMan, vMapped *PointerVector[AigObject]) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkDarToCnf
// C: Abc_Ntk_t * Abc_NtkDarToCnf( Abc_Ntk_t * pNtk, char * pFileName, int fFastAlgo, int fChangePol, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkDarToCnf(pFileName string, fFastAlgo int, fChangePol int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkDSat
// C: int Abc_NtkDSat( Abc_Ntk_t * pNtk, ABC_INT64_T nConfLimit, ABC_INT64_T nInsLimit, int nLearnedStart, int nLearnedDelta, int nLearnedPerce, int fAlignPol, int fAndOuts, int fNewSolver, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkDSat(nConfLimit int64, nInsLimit int64, nLearnedStart int, nLearnedDelta int, nLearnedPerce int, fAlignPol int, fAndOuts int, fNewSolver int, fVerbose int) int {
	panic("TODO, stubbed")
}

// Abc_NtkPartitionedSat
// C: int Abc_NtkPartitionedSat( Abc_Ntk_t * pNtk, int nAlgo, int nPartSize, int nConfPart, int nConfTotal, int fAlignPol, int fSynthesize, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkPartitionedSat(nAlgo int, nPartSize int, nConfPart int, nConfTotal int, fAlignPol int, fSynthesize int, fVerbose int) int {
	panic("TODO, stubbed")
}

// Abc_NtkDarCec
// C: int Abc_NtkDarCec( Abc_Ntk_t * pNtk1, Abc_Ntk_t * pNtk2, int nConfLimit, int fPartition, int fVerbose )
func (pNtk1 *ABCNetwork) AbcNtkDarCec(pNtk2 *ABCNetwork, nConfLimit int, fPartition int, fVerbose int) int {
	panic("TODO, stubbed")
}

// Abc_NtkDarSeqSweep
// C: Abc_Ntk_t * Abc_NtkDarSeqSweep( Abc_Ntk_t * pNtk, Fra_Ssw_t * pPars )
func (pNtk *ABCNetwork) AbcNtkDarSeqSweep(pPars *FraSsw) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkPrintLatchEquivClasses
// C: void Abc_NtkPrintLatchEquivClasses( Abc_Ntk_t * pNtk, Aig_Man_t * pAig )
func (pNtk *ABCNetwork) AbcNtkPrintLatchEquivClasses(pAig *AigManager) {
	panic("TODO, stubbed")
}

// Abc_NtkDarSeqSweep2
// C: Abc_Ntk_t * Abc_NtkDarSeqSweep2( Abc_Ntk_t * pNtk, Ssw_Pars_t * pPars )
func (pNtk *ABCNetwork) AbcNtkDarSeqSweep2(pPars *SswPars) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkDarLcorr
// C: Abc_Ntk_t * Abc_NtkDarLcorr( Abc_Ntk_t * pNtk, int nFramesP, int nConfMax, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkDarLcorr(nFramesP int, nConfMax int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkDarLcorrNew
// C: Abc_Ntk_t * Abc_NtkDarLcorrNew( Abc_Ntk_t * pNtk, int nVarsMax, int nConfMax, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkDarLcorrNew(nVarsMax int, nConfMax int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkDarBmc
// C: int Abc_NtkDarBmc( Abc_Ntk_t * pNtk, int nStart, int nFrames, int nSizeMax, int nNodeDelta, int nTimeOut, int nBTLimit, int nBTLimitAll, int fRewrite, int fNewAlgo, int fOrDecomp, int nCofFanLit, int fVerbose, int * piFrames, int fUseSatoko )
func (pNtk *ABCNetwork) AbcNtkDarBmc(nStart int, nFrames int, nSizeMax int, nNodeDelta int, nTimeOut int, nBTLimit int, nBTLimitAll int, fRewrite int, fNewAlgo int, fOrDecomp int, nCofFanLit int, fVerbose int, piFrames *int, fUseSatoko int) int {
	panic("TODO, stubbed")
}

// Abc_NtkDarBmc3
// C: int Abc_NtkDarBmc3( Abc_Ntk_t * pNtk, Saig_ParBmc_t * pPars, int fOrDecomp )
func (pNtk *ABCNetwork) AbcNtkDarBmc3(pPars *AigScalableBMCParameters, fOrDecomp int) int {
	panic("TODO, stubbed")
}

// Abc_NtkDarBmcInter_int
// C: int Abc_NtkDarBmcInter_int( Aig_Man_t * pMan, Inter_ManParams_t * pPars, Aig_Man_t ** ppNtkRes )
func (pMan *AigManager) AbcNtkDarBmcInter_int(pPars *ProofIntParams, ppNtkRes **AigManager) int {
	panic("TODO, stubbed")
}

// Abc_NtkDarBmcInter
// C: int Abc_NtkDarBmcInter( Abc_Ntk_t * pNtk, Inter_ManParams_t * pPars, Abc_Ntk_t ** ppNtkRes )
func (pNtk *ABCNetwork) AbcNtkDarBmcInter(pPars *ProofIntParams, ppNtkRes **ABCNetwork) int {
	panic("TODO, stubbed")
}

// Abc_NtkDarDemiter
// C: int Abc_NtkDarDemiter( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkDarDemiter() int {
	panic("TODO, stubbed")
}

// Abc_NtkDarDemiterNew
// C: int Abc_NtkDarDemiterNew( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkDarDemiterNew() int {
	panic("TODO, stubbed")
}

// Abc_NtkDarDemiterDual
// C: int Abc_NtkDarDemiterDual( Abc_Ntk_t * pNtk, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkDarDemiterDual(fVerbose int) int {
	panic("TODO, stubbed")
}

// Abc_NtkDarProve
// C: int Abc_NtkDarProve( Abc_Ntk_t * pNtk, Fra_Sec_t * pSecPar, int nBmcFramesMax, int nBmcConfMax )
func (pNtk *ABCNetwork) AbcNtkDarProve(pSecPar *FraSec, nBmcFramesMax int, nBmcConfMax int) int {
	panic("TODO, stubbed")
}

// Abc_NtkDarSec
// C: int Abc_NtkDarSec( Abc_Ntk_t * pNtk1, Abc_Ntk_t * pNtk2, Fra_Sec_t * pSecPar )
func (pNtk1 *ABCNetwork) AbcNtkDarSec(pNtk2 *ABCNetwork, pSecPar *FraSec) int {
	panic("TODO, stubbed")
}

// Abc_NtkDarPdr
// C: int Abc_NtkDarPdr( Abc_Ntk_t * pNtk, Pdr_Par_t * pPars )
func (pNtk *ABCNetwork) AbcNtkDarPdr(pPars *PdrPar) int {
	panic("TODO, stubbed")
}

// Abc_NtkDarAbSec
// C: int Abc_NtkDarAbSec( Abc_Ntk_t * pNtk1, Abc_Ntk_t * pNtk2, int nFrames, int fVerbose )
func (pNtk1 *ABCNetwork) AbcNtkDarAbSec(pNtk2 *ABCNetwork, nFrames int, fVerbose int) int {
	panic("TODO, stubbed")
}

// Abc_NtkDarSimSec
// C: int Abc_NtkDarSimSec( Abc_Ntk_t * pNtk1, Abc_Ntk_t * pNtk2, Ssw_Pars_t * pPars )
func (pNtk1 *ABCNetwork) AbcNtkDarSimSec(pNtk2 *ABCNetwork, pPars *SswPars) int {
	panic("TODO, stubbed")
}

// Abc_NtkDarMatch
// C: Abc_Ntk_t * Abc_NtkDarMatch( Abc_Ntk_t * pNtk1, Abc_Ntk_t * pNtk2, int nDist, int fVerbose )
func (pNtk1 *ABCNetwork) AbcNtkDarMatch(pNtk2 *ABCNetwork, nDist int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkDarLatchSweep
// C: Abc_Ntk_t * Abc_NtkDarLatchSweep( Abc_Ntk_t * pNtk, int fLatchConst, int fLatchEqual, int fSaveNames, int fUseMvSweep, int nFramesSymb, int nFramesSatur, int fVerbose, int fVeryVerbose )
func (pNtk *ABCNetwork) AbcNtkDarLatchSweep(fLatchConst int, fLatchEqual int, fSaveNames int, fUseMvSweep int, nFramesSymb int, nFramesSatur int, fVerbose int, fVeryVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkDarRetime
// C: Abc_Ntk_t * Abc_NtkDarRetime( Abc_Ntk_t * pNtk, int nStepsMax, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkDarRetime(nStepsMax int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkDarRetimeF
// C: Abc_Ntk_t * Abc_NtkDarRetimeF( Abc_Ntk_t * pNtk, int nStepsMax, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkDarRetimeF(nStepsMax int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkDarRetimeMostFwd
// C: Abc_Ntk_t * Abc_NtkDarRetimeMostFwd( Abc_Ntk_t * pNtk, int nMaxIters, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkDarRetimeMostFwd(nMaxIters int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkDarRetimeMinArea
// C: Abc_Ntk_t * Abc_NtkDarRetimeMinArea( Abc_Ntk_t * pNtk, int nMaxIters, int fForwardOnly, int fBackwardOnly, int fInitial, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkDarRetimeMinArea(nMaxIters int, fForwardOnly int, fBackwardOnly int, fInitial int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkDarRetimeStep
// C: Abc_Ntk_t * Abc_NtkDarRetimeStep( Abc_Ntk_t * pNtk, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkDarRetimeStep(fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkDarSeqSim
// C: int Abc_NtkDarSeqSim( Abc_Ntk_t * pNtk, int nFrames, int nWords, int TimeOut, int fNew, int fMiter, int fVerbose, char * pFileSim )
func (pNtk *ABCNetwork) AbcNtkDarSeqSim(nFrames int, nWords int, TimeOut int, fNew int, fMiter int, fVerbose int, pFileSim string) int {
	panic("TODO, stubbed")
}

// Abc_NtkDarSeqSim3
// C: int Abc_NtkDarSeqSim3( Abc_Ntk_t * pNtk, Ssw_RarPars_t * pPars )
func (pNtk *ABCNetwork) AbcNtkDarSeqSim3(pPars *SswRarPars) int {
	panic("TODO, stubbed")
}

// Abc_NtkDarClau
// C: int Abc_NtkDarClau( Abc_Ntk_t * pNtk, int nFrames, int nPref, int nClauses, int nLutSize, int nLevels, int nCutsMax, int nBatches, int fStepUp, int fBmc, int fRefs, int fTarget, int fVerbose, int fVeryVerbose )
func (pNtk *ABCNetwork) AbcNtkDarClau(nFrames int, nPref int, nClauses int, nLutSize int, nLevels int, nCutsMax int, nBatches int, fStepUp int, fBmc int, fRefs int, fTarget int, fVerbose int, fVeryVerbose int) int {
	panic("TODO, stubbed")
}

// Abc_NtkDarEnlarge
// C: Abc_Ntk_t * Abc_NtkDarEnlarge( Abc_Ntk_t * pNtk, int nFrames, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkDarEnlarge(nFrames int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkDarTempor
// C: Abc_Ntk_t * Abc_NtkDarTempor( Abc_Ntk_t * pNtk, int nFrames, int TimeOut, int nConfLimit, int fUseBmc, int fUseTransSigs, int fVerbose, int fVeryVerbose )
func (pNtk *ABCNetwork) AbcNtkDarTempor(nFrames int, TimeOut int, nConfLimit int, fUseBmc int, fUseTransSigs int, fVerbose int, fVeryVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkDarInduction
// C: int Abc_NtkDarInduction( Abc_Ntk_t * pNtk, int nTimeOut, int nFramesMax, int nConfMax, int fUnique, int fUniqueAll, int fGetCex, int fVerbose, int fVeryVerbose )
func (pNtk *ABCNetwork) AbcNtkDarInduction(nTimeOut int, nFramesMax int, nConfMax int, fUnique int, fUniqueAll int, fGetCex int, fVerbose int, fVeryVerbose int) int {
	panic("TODO, stubbed")
}

// Abc_NtkInterOne
// C: Abc_Ntk_t * Abc_NtkInterOne( Abc_Ntk_t * pNtkOn, Abc_Ntk_t * pNtkOff, int fRelation, int fVerbose )
func (pNtkOn *ABCNetwork) AbcNtkInterOne(pNtkOff *ABCNetwork, fRelation int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Gia_ManInterOne
// C: Gia_Man_t * Gia_ManInterOne( Gia_Man_t * pNtkOn, Gia_Man_t * pNtkOff, int fVerbose )
func (pNtkOn *GiaManager) GiaManInterOne(pNtkOff *GiaManager, fVerbose int) *GiaManager {
	panic("TODO, stubbed")
}

// Abc_NtkInterFast
// C: void Abc_NtkInterFast( Abc_Ntk_t * pNtkOn, Abc_Ntk_t * pNtkOff, int fVerbose )
func (pNtkOn *ABCNetwork) AbcNtkInterFast(pNtkOff *ABCNetwork, fVerbose int) {
	panic("TODO, stubbed")
}

// Abc_NtkInter
// C: Abc_Ntk_t * Abc_NtkInter( Abc_Ntk_t * pNtkOn, Abc_Ntk_t * pNtkOff, int fRelation, int fVerbose )
func (pNtkOn *ABCNetwork) AbcNtkInter(pNtkOff *ABCNetwork, fRelation int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkPrintSccs
// C: void Abc_NtkPrintSccs( Abc_Ntk_t * pNtk, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkPrintSccs(fVerbose int) {
	panic("TODO, stubbed")
}

// Abc_NtkDarPrintCone
// C: int Abc_NtkDarPrintCone( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkDarPrintCone() int {
	panic("TODO, stubbed")
}

// Abc_NtkBalanceExor
// C: Abc_Ntk_t * Abc_NtkBalanceExor( Abc_Ntk_t * pNtk, int fUpdateLevel, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkBalanceExor(fUpdateLevel int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkPhaseAbstract
// C: Abc_Ntk_t * Abc_NtkPhaseAbstract( Abc_Ntk_t * pNtk, int nFrames, int nPref, int fIgnore, int fPrint, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkPhaseAbstract(nFrames int, nPref int, fIgnore int, fPrint int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkPhaseFrameNum
// C: int Abc_NtkPhaseFrameNum( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkPhaseFrameNum() int {
	panic("TODO, stubbed")
}

// Abc_NtkDarSynchOne
// C: Abc_Ntk_t * Abc_NtkDarSynchOne( Abc_Ntk_t * pNtk, int nWords, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkDarSynchOne(nWords int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkDarSynch
// C: Abc_Ntk_t * Abc_NtkDarSynch( Abc_Ntk_t * pNtk1, Abc_Ntk_t * pNtk2, int nWords, int fVerbose )
func (pNtk1 *ABCNetwork) AbcNtkDarSynch(pNtk2 *ABCNetwork, nWords int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkDarClockGate
// C: Abc_Ntk_t * Abc_NtkDarClockGate( Abc_Ntk_t * pNtk, Abc_Ntk_t * pCare, Cgt_Par_t * pPars )
func (pNtk *ABCNetwork) AbcNtkDarClockGate(pCare *ABCNetwork, pPars *ClockGatingParameters) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkDarExtWin
// C: Abc_Ntk_t * Abc_NtkDarExtWin( Abc_Ntk_t * pNtk, int nObjId, int nDist, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkDarExtWin(nObjId int, nDist int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkDarInsWin
// C: Abc_Ntk_t * Abc_NtkDarInsWin( Abc_Ntk_t * pNtk, Abc_Ntk_t * pCare, int nObjId, int nDist, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkDarInsWin(pCare *ABCNetwork, nObjId int, nDist int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkDarFrames
// C: Abc_Ntk_t * Abc_NtkDarFrames( Abc_Ntk_t * pNtk, int nPrefix, int nFrames, int fInit, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkDarFrames(nPrefix int, nFrames int, fInit int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkDarCleanupAig
// C: Abc_Ntk_t * Abc_NtkDarCleanupAig( Abc_Ntk_t * pNtk, int fCleanupPis, int fCleanupPos, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkDarCleanupAig(fCleanupPis int, fCleanupPos int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkDarReach
// C: int Abc_NtkDarReach( Abc_Ntk_t * pNtk, Saig_ParBbr_t * pPars )
func (pNtk *ABCNetwork) AbcNtkDarReach(pPars *SaigBDDParameters) int {
	panic("TODO, stubbed")
}

// Amap_ManProduceNetwork (vMapping holds Amap_Out_t * entries (no Go port of Amap_Out_t))
// C: Abc_Ntk_t * Amap_ManProduceNetwork( Abc_Ntk_t * pNtk, Vec_Ptr_t * vMapping )
func (pNtk *ABCNetwork) AmapManProduceNetwork(vMapping *PointerVector[AmapOut]) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkDarAmap
// C: Abc_Ntk_t * Abc_NtkDarAmap( Abc_Ntk_t * pNtk, Amap_Par_t * pPars )
func (pNtk *ABCNetwork) AbcNtkDarAmap(pPars *AmapPar) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkDarConstr
// C: void Abc_NtkDarConstr( Abc_Ntk_t * pNtk, int nFrames, int nConfs, int nProps, int fStruct, int fOldAlgo, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkDarConstr(nFrames int, nConfs int, nProps int, fStruct int, fOldAlgo int, fVerbose int) {
	panic("TODO, stubbed")
}

// Abc_NtkDarOutdec
// C: Abc_Ntk_t * Abc_NtkDarOutdec( Abc_Ntk_t * pNtk, int nLits, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkDarOutdec(nLits int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkDarUnfold
// C: Abc_Ntk_t * Abc_NtkDarUnfold( Abc_Ntk_t * pNtk, int nFrames, int nConfs, int nProps, int fStruct, int fOldAlgo, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkDarUnfold(nFrames int, nConfs int, nProps int, fStruct int, fOldAlgo int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkDarFold
// C: Abc_Ntk_t * Abc_NtkDarFold( Abc_Ntk_t * pNtk, int fCompl, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkDarFold(fCompl int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkDarConstrProfile
// C: void Abc_NtkDarConstrProfile( Abc_Ntk_t * pNtk, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkDarConstrProfile(fVerbose int) {
	panic("TODO, stubbed")
}

// Abc_NtkDarTest
// C: void Abc_NtkDarTest( Abc_Ntk_t * pNtk, int Num )
func (pNtk *ABCNetwork) AbcNtkDarTest(Num int) {
	panic("TODO, stubbed")
}

// Abc_NtkDarTestNtk
// C: Abc_Ntk_t * Abc_NtkDarTestNtk( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkDarTestNtk() *ABCNetwork {
	panic("TODO, stubbed")
}
