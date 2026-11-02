package abc

// Hand-checked skeleton of src/opt/dar/darScript.c.

// Dar_ManRewriteDefault
// C: Aig_Man_t * Dar_ManRewriteDefault( Aig_Man_t * pAig )
func (pAig *AigManager) DarManRewriteDefault() *AigManager {
	panic("TODO, stubbed")
}

// Dar_ManRwsat
// C: Aig_Man_t * Dar_ManRwsat( Aig_Man_t * pAig, int fBalance, int fVerbose )
func (pAig *AigManager) DarManRwsat(fBalance int, fVerbose int) *AigManager {
	panic("TODO, stubbed")
}

// Dar_ManCompress
// C: Aig_Man_t * Dar_ManCompress( Aig_Man_t * pAig, int fBalance, int fUpdateLevel, int fPower, int fVerbose )
func (pAig *AigManager) DarManCompress(fBalance int, fUpdateLevel int, fPower int, fVerbose int) *AigManager {
	panic("TODO, stubbed")
}

// Dar_ManCompress2
// C: Aig_Man_t * Dar_ManCompress2( Aig_Man_t * pAig, int fBalance, int fUpdateLevel, int fFanout, int fPower, int fVerbose )
func (pAig *AigManager) DarManCompress2(fBalance int, fUpdateLevel int, fFanout int, fPower int, fVerbose int) *AigManager {
	panic("TODO, stubbed")
}

// Dar_ManChoiceSynthesis
// C: Vec_Ptr_t * Dar_ManChoiceSynthesis( Aig_Man_t * pAig, int fBalance, int fUpdateLevel, int fPower, int fVerbose )
func (pAig *AigManager) DarManChoiceSynthesis(fBalance int, fUpdateLevel int, fPower int, fVerbose int) *PointerVector[AigManager] {
	panic("TODO, stubbed")
}

// Dar_ManChoice
// C: Aig_Man_t * Dar_ManChoice( Aig_Man_t * pAig, int fBalance, int fUpdateLevel, int fConstruct, int nConfMax, int nLevelMax, int fVerbose )
func (pAig *AigManager) DarManChoice(fBalance int, fUpdateLevel int, fConstruct int, nConfMax int, nLevelMax int, fVerbose int) *AigManager {
	panic("TODO, stubbed")
}

// Dar_NewCompress
// C: Aig_Man_t * Dar_NewCompress( Aig_Man_t * pAig, int fBalance, int fUpdateLevel, int fPower, int fVerbose )
func (pAig *AigManager) DarNewCompress(fBalance int, fUpdateLevel int, fPower int, fVerbose int) *AigManager {
	panic("TODO, stubbed")
}

// Dar_NewCompress2
// C: Aig_Man_t * Dar_NewCompress2( Aig_Man_t * pAig, int fBalance, int fUpdateLevel, int fFanout, int fPower, int fLightSynth, int fVerbose )
func (pAig *AigManager) DarNewCompress2(fBalance int, fUpdateLevel int, fFanout int, fPower int, fLightSynth int, fVerbose int) *AigManager {
	panic("TODO, stubbed")
}

// Dar_NewChoiceSynthesisGuard
// C: int Dar_NewChoiceSynthesisGuard( Aig_Man_t * pAig )
func (pAig *AigManager) DarNewChoiceSynthesisGuard() int {
	panic("TODO, stubbed")
}

// Dar_NewChoiceSynthesis
// C: Gia_Man_t * Dar_NewChoiceSynthesis( Aig_Man_t * pAig, int fBalance, int fUpdateLevel, int fPower, int fLightSynth, int fVerbose )
func (pAig *AigManager) DarNewChoiceSynthesis(fBalance int, fUpdateLevel int, fPower int, fLightSynth int, fVerbose int) *GiaManager {
	panic("TODO, stubbed")
}

// Dar_ManChoiceNewAig
// C: Aig_Man_t * Dar_ManChoiceNewAig( Aig_Man_t * pAig, Dch_Pars_t * pPars )
func (pAig *AigManager) DarManChoiceNewAig(pPars *DchOptions) *AigManager {
	panic("TODO, stubbed")
}

// Dar_ManChoiceNew
// C: Aig_Man_t * Dar_ManChoiceNew( Aig_Man_t * pAig, Dch_Pars_t * pPars )
func (pAig *AigManager) DarManChoiceNew(pPars *DchOptions) *AigManager {
	panic("TODO, stubbed")
}
