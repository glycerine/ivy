package abc

// Hand-checked skeleton of src/aig/gia/giaScript.c.

// Gia_ManAigPrintPiLevels
// C: void Gia_ManAigPrintPiLevels( Gia_Man_t * p )
func (p *GiaManager) GiaManAigPrintPiLevels() {
	panic("TODO, stubbed")
}

// Gia_ManAigSyn2
// C: Gia_Man_t * Gia_ManAigSyn2( Gia_Man_t * pInit, int fOldAlgo, int fCoarsen, int fCutMin, int nRelaxRatio, int fDelayMin, int fVerbose, int fVeryVerbose )
func (pInit *GiaManager) GiaManAigSyn2(fOldAlgo int, fCoarsen int, fCutMin int, nRelaxRatio int, fDelayMin int, fVerbose int, fVeryVerbose int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManAigSyn3
// C: Gia_Man_t * Gia_ManAigSyn3( Gia_Man_t * p, int fVerbose, int fVeryVerbose )
func (p *GiaManager) GiaManAigSyn3(fVerbose int, fVeryVerbose int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManAigSyn4
// C: Gia_Man_t * Gia_ManAigSyn4( Gia_Man_t * p, int fVerbose, int fVeryVerbose )
func (p *GiaManager) GiaManAigSyn4(fVerbose int, fVeryVerbose int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManOrderPios
// C: Vec_Ptr_t * Gia_ManOrderPios( Aig_Man_t * p, Gia_Man_t * pOrder )
func (p *AigManager) GiaManOrderPios(pOrder *GiaManager) *PointerVector[AigObject] {
	panic("TODO, stubbed")
}

// Gia_ManDupFromBarBufs
// C: Gia_Man_t * Gia_ManDupFromBarBufs( Gia_Man_t * p )
func (p *GiaManager) GiaManDupFromBarBufs() *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManDupToBarBufs
// C: Gia_Man_t * Gia_ManDupToBarBufs( Gia_Man_t * p, int nBarBufs )
func (p *GiaManager) GiaManDupToBarBufs(nBarBufs int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManAigSynch2Choices
// C: Gia_Man_t * Gia_ManAigSynch2Choices( Gia_Man_t * pGia1, Gia_Man_t * pGia2, Gia_Man_t * pGia3, Dch_Pars_t * pPars )
func (pGia1 *GiaManager) GiaManAigSynch2Choices(pGia2 *GiaManager, pGia3 *GiaManager, pPars *DchOptions) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManAigSynch2 (pPars0 is a void * that holds a Dch_Pars_t *)
// C: Gia_Man_t * Gia_ManAigSynch2( Gia_Man_t * pInit, void * pPars0, int nLutSize, int nRelaxRatio )
func (pInit *GiaManager) GiaManAigSynch2(pPars0 *DchOptions, nLutSize int, nRelaxRatio int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManPerformMap
// C: void Gia_ManPerformMap( int nAnds, int nLutSize, int nCutNum, int fMinAve, int fUseMfs, int fVerbose )
func GiaManPerformMap(nAnds int, nLutSize int, nCutNum int, fMinAve int, fUseMfs int, fVerbose int) {
	panic("TODO, stubbed")
}

// Gia_ManPerformRound
// C: void Gia_ManPerformRound( int fIsMapped, int nAnds, int nLevels, int nLutSize, int nCutNum, int fMinAve, int fUseMfs, int fVerbose )
func GiaManPerformRound(fIsMapped int, nAnds int, nLevels int, nLutSize int, nCutNum int, fMinAve int, fUseMfs int, fVerbose int) {
	panic("TODO, stubbed")
}

// Gia_ManPerformFlow
// C: void Gia_ManPerformFlow( int fIsMapped, int nAnds, int nLevels, int nLutSize, int nCutNum, int fMinAve, int fUseMfs, int fVerbose )
func GiaManPerformFlow(fIsMapped int, nAnds int, nLevels int, nLutSize int, nCutNum int, fMinAve int, fUseMfs int, fVerbose int) {
	panic("TODO, stubbed")
}

// Gia_ManPerformFlow2
// C: void Gia_ManPerformFlow2( int fIsMapped, int nAnds, int nLevels, int nLutSize, int nCutNum, int fBalance, int fMinAve, int fUseMfs, int fVerbose )
func GiaManPerformFlow2(fIsMapped int, nAnds int, nLevels int, nLutSize int, nCutNum int, fBalance int, fMinAve int, fUseMfs int, fVerbose int) {
	panic("TODO, stubbed")
}

// Gia_ManPerformFlow3
// C: void Gia_ManPerformFlow3( int nLutSize, int nCutNum, int fBalance, int fMinAve, int fUseMfs, int fUseLutLib, int fVerbose )
func GiaManPerformFlow3(nLutSize int, nCutNum int, fBalance int, fMinAve int, fUseMfs int, fUseLutLib int, fVerbose int) {
	panic("TODO, stubbed")
}
