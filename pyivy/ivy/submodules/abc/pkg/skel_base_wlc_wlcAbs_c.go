package abc

// Hand-checked skeleton of src/base/wlc/wlcAbs.c.

// Wlc_NtkPrintNtk
// C: void Wlc_NtkPrintNtk( Wlc_Ntk_t * p )
func WlcNtkPrintNtk(p *WLCNetwork) {
	panic("TODO, stubbed")
}

// Wlc_NtkAbsGetSupp_rec
// C: void Wlc_NtkAbsGetSupp_rec( Wlc_Ntk_t * p, Wlc_Obj_t * pObj, Vec_Bit_t * vCiMarks, Vec_Int_t * vSuppRefs, Vec_Int_t * vSuppList )
func WlcNtkAbsGetSupp_rec(p *WLCNetwork, pObj *WLCObject, vCiMarks *BitVector, vSuppRefs *IntVector, vSuppList *IntVector) {
	panic("TODO, stubbed")
}

// Wlc_NtkAbsGetSupp
// C: void Wlc_NtkAbsGetSupp( Wlc_Ntk_t * p, Wlc_Obj_t * pObj, Vec_Bit_t * vCiMarks, Vec_Int_t * vSuppRefs, Vec_Int_t * vSuppList)
func WlcNtkAbsGetSupp(p *WLCNetwork, pObj *WLCObject, vCiMarks *BitVector, vSuppRefs *IntVector, vSuppList *IntVector) {
	panic("TODO, stubbed")
}

// Wlc_NtkNumPiBits
// C: int Wlc_NtkNumPiBits( Wlc_Ntk_t * pNtk )
func WlcNtkNumPiBits(pNtk *WLCNetwork) int {
	panic("TODO, stubbed")
}

// Wlc_NtkAbsAnalyzeRefine (nDisj and nNDisj are out parameters)
// C: void Wlc_NtkAbsAnalyzeRefine( Wlc_Ntk_t * p, Vec_Int_t * vBlacks, Vec_Bit_t * vUnmark, int * nDisj, int * nNDisj )
func WlcNtkAbsAnalyzeRefine(p *WLCNetwork, vBlacks *IntVector, vUnmark *BitVector, nDisj *int, nNDisj *int) {
	panic("TODO, stubbed")
}

// Wlc_NtkGetCoreSels (static)
// C: Vec_Int_t * Wlc_NtkGetCoreSels( Gia_Man_t * pFrames, int nFrames, int first_sel_pi, int num_sel_pis, Vec_Bit_t * vMark, int nConfLimit, Wlc_Par_t * pPars, int fSetPO, int RunId )
func wlcNtkGetCoreSels(pFrames *GiaManager, nFrames int, first_sel_pi int, num_sel_pis int, vMark *BitVector, nConfLimit int, pPars *WlcPar, fSetPO int, RunId int) *IntVector {
	panic("TODO, stubbed")
}

// Wlc_NtkUnrollWoCex (static)
// C: Gia_Man_t * Wlc_NtkUnrollWoCex(Wlc_Ntk_t * pChoice, int nFrames, int first_sel_pi, int num_sel_pis)
func wlcNtkUnrollWoCex(pChoice *WLCNetwork, nFrames int, first_sel_pi int, num_sel_pis int) *GiaManager {
	panic("TODO, stubbed")
}

// Wlc_NtkUnrollWithCex (static; p_num_ppis is an out parameter)
// C: Gia_Man_t * Wlc_NtkUnrollWithCex(Wlc_Ntk_t * pChoice, Abc_Cex_t * pCex, int nbits_old_pis, int num_sel_pis, int * p_num_ppis, int sel_pi_first, int fUsePPI)
func wlcNtkUnrollWithCex(pChoice *WLCNetwork, pCex *Counterexample, nbits_old_pis int, num_sel_pis int, p_num_ppis *int, sel_pi_first int, fUsePPI int) *GiaManager {
	panic("TODO, stubbed")
}

// Wlc_NtkIntroduceChoices
// C: Wlc_Ntk_t * Wlc_NtkIntroduceChoices( Wlc_Ntk_t * pNtk, Vec_Int_t * vBlacks, Vec_Int_t * vPPIs )
func WlcNtkIntroduceChoices(pNtk *WLCNetwork, vBlacks *IntVector, vPPIs *IntVector) *WLCNetwork {
	panic("TODO, stubbed")
}

// Wlc_NtkCexIsReal (static)
// C: Abc_Cex_t * Wlc_NtkCexIsReal( Wlc_Ntk_t * pOrig, Abc_Cex_t * pCex )
func wlcNtkCexIsReal(pOrig *WLCNetwork, pCex *Counterexample) *Counterexample {
	panic("TODO, stubbed")
}

// Wlc_NtkAbs2 (static; pvFlops is an out parameter)
// C: Wlc_Ntk_t * Wlc_NtkAbs2( Wlc_Ntk_t * pNtk, Vec_Int_t * vBlacks, Vec_Int_t ** pvFlops )
func wlcNtkAbs2(pNtk *WLCNetwork, vBlacks *IntVector, pvFlops **IntVector) *WLCNetwork {
	panic("TODO, stubbed")
}

// Wlc_NtkProofReduce (static)
// C: Vec_Bit_t * Wlc_NtkProofReduce( Wlc_Ntk_t * p, Wlc_Par_t * pPars, int nFrames, Vec_Int_t * vWhites, Vec_Int_t * vBlacks, int RunId )
func wlcNtkProofReduce(p *WLCNetwork, pPars *WlcPar, nFrames int, vWhites *IntVector, vBlacks *IntVector, RunId int) *BitVector {
	panic("TODO, stubbed")
}

// Wlc_NtkProofRefine (static; pvRefine is an out parameter)
// C: int Wlc_NtkProofRefine( Wlc_Ntk_t * p, Wlc_Par_t * pPars, Abc_Cex_t * pCex, Vec_Int_t * vBlacks, Vec_Int_t ** pvRefine )
func wlcNtkProofRefine(p *WLCNetwork, pPars *WlcPar, pCex *Counterexample, vBlacks *IntVector, pvRefine **IntVector) int {
	panic("TODO, stubbed")
}

// Wlc_NtkUpdateBlacks (static; pvBlacks is an in/out parameter)
// C: int Wlc_NtkUpdateBlacks( Wlc_Ntk_t * p, Wlc_Par_t * pPars, Vec_Int_t ** pvBlacks, Vec_Bit_t * vUnmark, Vec_Int_t * vSignals )
func wlcNtkUpdateBlacks(p *WLCNetwork, pPars *WlcPar, pvBlacks **IntVector, vUnmark *BitVector, vSignals *IntVector) int {
	panic("TODO, stubbed")
}

// Wlc_NtkMarkLimit (static)
// C: Vec_Bit_t * Wlc_NtkMarkLimit( Wlc_Ntk_t * p, Wlc_Par_t * pPars )
func wlcNtkMarkLimit(p *WLCNetwork, pPars *WlcPar) *BitVector {
	panic("TODO, stubbed")
}

// Wlc_NtkSetUnmark (static)
// C: void Wlc_NtkSetUnmark( Wlc_Ntk_t * p, Wlc_Par_t * pPars, Vec_Bit_t * vUnmark )
func wlcNtkSetUnmark(p *WLCNetwork, pPars *WlcPar, vUnmark *BitVector) {
	panic("TODO, stubbed")
}

// Wlc_NtkGetBlacks (static)
// C: Vec_Int_t * Wlc_NtkGetBlacks( Wlc_Ntk_t * p, Wlc_Par_t * pPars )
func wlcNtkGetBlacks(p *WLCNetwork, pPars *WlcPar) *IntVector {
	panic("TODO, stubbed")
}

// Wlc_NtkAbsMarkOpers (static)
// C: Vec_Bit_t * Wlc_NtkAbsMarkOpers( Wlc_Ntk_t * p, Wlc_Par_t * pPars, Vec_Bit_t * vUnmark, int fVerbose )
func wlcNtkAbsMarkOpers(p *WLCNetwork, pPars *WlcPar, vUnmark *BitVector, fVerbose int) *BitVector {
	panic("TODO, stubbed")
}

// Wlc_NtkAbsMarkNodes_rec (static)
// C: void Wlc_NtkAbsMarkNodes_rec( Wlc_Ntk_t * p, Wlc_Obj_t * pObj, Vec_Bit_t * vLeaves, Vec_Int_t * vPisOld, Vec_Int_t * vPisNew, Vec_Int_t * vFlops )
func wlcNtkAbsMarkNodes_rec(p *WLCNetwork, pObj *WLCObject, vLeaves *BitVector, vPisOld *IntVector, vPisNew *IntVector, vFlops *IntVector) {
	panic("TODO, stubbed")
}

// Wlc_NtkAbsMarkNodes (static)
// C: void Wlc_NtkAbsMarkNodes( Wlc_Ntk_t * p, Vec_Bit_t * vLeaves, Vec_Int_t * vPisOld, Vec_Int_t * vPisNew, Vec_Int_t * vFlops )
func wlcNtkAbsMarkNodes(p *WLCNetwork, vLeaves *BitVector, vPisOld *IntVector, vPisNew *IntVector, vFlops *IntVector) {
	panic("TODO, stubbed")
}

// Wlc_NtkAbs (static; pvPisNew and pvFlops are out parameters)
// C: Wlc_Ntk_t * Wlc_NtkAbs( Wlc_Ntk_t * p, Wlc_Par_t * pPars, Vec_Bit_t * vUnmark, Vec_Int_t ** pvPisNew, Vec_Int_t ** pvFlops, int fVerbose )
func wlcNtkAbs(p *WLCNetwork, pPars *WlcPar, vUnmark *BitVector, pvPisNew **IntVector, pvFlops **IntVector, fVerbose int) *WLCNetwork {
	panic("TODO, stubbed")
}

// Wlc_NtkAbsRefinement (static)
// C: Vec_Int_t * Wlc_NtkAbsRefinement( Wlc_Ntk_t * p, Gia_Man_t * pGia, Abc_Cex_t * pCex, Vec_Int_t * vPisNew )
func wlcNtkAbsRefinement(p *WLCNetwork, pGia *GiaManager, pCex *Counterexample, vPisNew *IntVector) *IntVector {
	panic("TODO, stubbed")
}

// Wlc_NtkNodeDeref_rec (static)
// C: int Wlc_NtkNodeDeref_rec( Wlc_Ntk_t * p, Wlc_Obj_t * pNode, Vec_Bit_t * vUnmark )
func wlcNtkNodeDeref_rec(p *WLCNetwork, pNode *WLCObject, vUnmark *BitVector) int {
	panic("TODO, stubbed")
}

// Wlc_NtkNodeRef_rec (static)
// C: int Wlc_NtkNodeRef_rec( Wlc_Ntk_t * p, Wlc_Obj_t * pNode )
func wlcNtkNodeRef_rec(p *WLCNetwork, pNode *WLCObject) int {
	panic("TODO, stubbed")
}

// Wlc_NtkMarkMffc (static)
// C: int Wlc_NtkMarkMffc( Wlc_Ntk_t * p, Wlc_Obj_t * pNode, Vec_Bit_t * vUnmark )
func wlcNtkMarkMffc(p *WLCNetwork, pNode *WLCObject, vUnmark *BitVector) int {
	panic("TODO, stubbed")
}

// Wlc_NtkRemoveFromAbstraction (static)
// C: int Wlc_NtkRemoveFromAbstraction( Wlc_Ntk_t * p, Vec_Int_t * vRefine, Vec_Bit_t * vUnmark )
func wlcNtkRemoveFromAbstraction(p *WLCNetwork, vRefine *IntVector, vUnmark *BitVector) int {
	panic("TODO, stubbed")
}

// Wlc_NtkUnmarkRefinement (static)
// C: int Wlc_NtkUnmarkRefinement( Wlc_Ntk_t * p, Vec_Int_t * vRefine, Vec_Bit_t * vUnmark )
func wlcNtkUnmarkRefinement(p *WLCNetwork, vRefine *IntVector, vUnmark *BitVector) int {
	panic("TODO, stubbed")
}

// Wlc_NtkFlopsRemap
// C: Vec_Int_t * Wlc_NtkFlopsRemap( Wlc_Ntk_t * p, Vec_Int_t * vFfOld, Vec_Int_t * vFfNew )
func WlcNtkFlopsRemap(p *WLCNetwork, vFfOld *IntVector, vFfNew *IntVector) *IntVector {
	panic("TODO, stubbed")
}

// Wla_ManCollectNodes
// C: Vec_Int_t * Wla_ManCollectNodes( Wla_Man_t * pWla, int fBlack )
func (pWla *WlaMan) WlaManCollectNodes(fBlack int) *IntVector {
	panic("TODO, stubbed")
}

// Wla_ManShrinkAbs
// C: int Wla_ManShrinkAbs( Wla_Man_t * pWla, int nFrames, int RunId )
func (pWla *WlaMan) WlaManShrinkAbs(nFrames int, RunId int) int {
	panic("TODO, stubbed")
}

// Wla_ManCreateAbs
// C: Wlc_Ntk_t * Wla_ManCreateAbs( Wla_Man_t * pWla )
func (pWla *WlaMan) WlaManCreateAbs() *WLCNetwork {
	panic("TODO, stubbed")
}

// Wla_ManBitBlast
// C: Aig_Man_t * Wla_ManBitBlast( Wla_Man_t * pWla, Wlc_Ntk_t * pAbs )
func (pWla *WlaMan) WlaManBitBlast(pAbs *WLCNetwork) *AigManager {
	panic("TODO, stubbed")
}

// Wla_ManCheckCombUnsat
// C: int Wla_ManCheckCombUnsat( Wla_Man_t * pWla, Aig_Man_t * pAig )
func (pWla *WlaMan) WlaManCheckCombUnsat(pAig *AigManager) int {
	panic("TODO, stubbed")
}

// Wla_ManSolveInt
// C: int Wla_ManSolveInt( Wla_Man_t * pWla, Aig_Man_t * pAig )
func (pWla *WlaMan) WlaManSolveInt(pAig *AigManager) int {
	panic("TODO, stubbed")
}

// Wla_ManRefine
// C: void Wla_ManRefine( Wla_Man_t * pWla )
func (pWla *WlaMan) WlaManRefine() {
	panic("TODO, stubbed")
}

// Wla_ManStart
// C: Wla_Man_t * Wla_ManStart( Wlc_Ntk_t * pNtk, Wlc_Par_t * pPars )
func WlaManStart(pNtk *WLCNetwork, pPars *WlcPar) *WlaMan {
	panic("TODO, stubbed")
}

// Wla_ManStop
// C: void Wla_ManStop( Wla_Man_t * p )
func (p *WlaMan) WlaManStop() {
	panic("TODO, stubbed")
}

// Wla_ManSolve
// C: int Wla_ManSolve( Wla_Man_t * pWla, Wlc_Par_t * pPars )
func (pWla *WlaMan) WlaManSolve(pPars *WlcPar) int {
	panic("TODO, stubbed")
}

// Wlc_NtkPdrAbs
// C: int Wlc_NtkPdrAbs( Wlc_Ntk_t * p, Wlc_Par_t * pPars )
func WlcNtkPdrAbs(p *WLCNetwork, pPars *WlcPar) int {
	panic("TODO, stubbed")
}

// Wlc_NtkAbsCore
// C: int Wlc_NtkAbsCore( Wlc_Ntk_t * p, Wlc_Par_t * pPars )
func WlcNtkAbsCore(p *WLCNetwork, pPars *WlcPar) int {
	panic("TODO, stubbed")
}
