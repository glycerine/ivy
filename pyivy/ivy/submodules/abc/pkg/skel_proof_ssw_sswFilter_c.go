package abc

// Hand-checked skeleton of src/proof/ssw/sswFilter.c.

// Ssw_ManRefineByFilterSim
// C: void Ssw_ManRefineByFilterSim( Ssw_Man_t * p, int nFrames )
func (p *SswMan) SswManRefineByFilterSim(nFrames int) {
	panic("TODO, stubbed")
}

// Ssw_ManRollForward
// C: void Ssw_ManRollForward( Ssw_Man_t * p, int nFrames )
func (p *SswMan) SswManRollForward(nFrames int) {
	panic("TODO, stubbed")
}

// Ssw_ManFindStartingState
// C: void Ssw_ManFindStartingState( Ssw_Man_t * p, Abc_Cex_t * pCex )
func (p *SswMan) SswManFindStartingState(pCex *Counterexample) {
	panic("TODO, stubbed")
}

// Ssw_ManSweepNodeFilter
// C: int Ssw_ManSweepNodeFilter( Ssw_Man_t * p, Aig_Obj_t * pObj, int f )
func (p *SswMan) SswManSweepNodeFilter(pObj *AigObject, f int) int {
	panic("TODO, stubbed")
}

// Ssw_ManSweepBmcFilter_rec
// C: Aig_Obj_t * Ssw_ManSweepBmcFilter_rec( Ssw_Man_t * p, Aig_Obj_t * pObj, int f )
func (p *SswMan) SswManSweepBmcFilter_rec(pObj *AigObject, f int) *AigObject {
	panic("TODO, stubbed")
}

// Ssw_ManSweepBmcFilter
// C: int Ssw_ManSweepBmcFilter( Ssw_Man_t * p, int TimeLimit )
func (p *SswMan) SswManSweepBmcFilter(TimeLimit int) int {
	panic("TODO, stubbed")
}

// Ssw_SignalFilter
// C: void Ssw_SignalFilter( Aig_Man_t * pAig, int nFramesMax, int nConfMax, int nRounds, int TimeLimit, int TimeLimit2, Abc_Cex_t * pCex, int fLatchOnly, int fVerbose )
func (pAig *AigManager) SswSignalFilter(nFramesMax int, nConfMax int, nRounds int, TimeLimit int, TimeLimit2 int, pCex *Counterexample, fLatchOnly int, fVerbose int) {
	panic("TODO, stubbed")
}

// Ssw_SignalFilterGia
// C: void Ssw_SignalFilterGia( Gia_Man_t * p, int nFramesMax, int nConfMax, int nRounds, int TimeLimit, int TimeLimit2, Abc_Cex_t * pCex, int fLatchOnly, int fVerbose )
func (p *GiaManager) SswSignalFilterGia(nFramesMax int, nConfMax int, nRounds int, TimeLimit int, TimeLimit2 int, pCex *Counterexample, fLatchOnly int, fVerbose int) {
	panic("TODO, stubbed")
}
