package abc

// Hand-checked skeleton of src/bdd/llb/llb1Hint.c.

// Llb_ManMaxFanoutCi
// C: int Llb_ManMaxFanoutCi( Aig_Man_t * pAig )
func (pAig *AigManager) LlbManMaxFanoutCi() int {
	panic("TODO, stubbed")
}

// Llb_ManPerformHints
// C: Aig_Man_t * Llb_ManPerformHints( Aig_Man_t * pAig, int nHintDepth )
func (pAig *AigManager) LlbManPerformHints(nHintDepth int) *AigManager {
	panic("TODO, stubbed")
}

// Llb_ManCollectHighFanoutObjects
// C: Vec_Int_t * Llb_ManCollectHighFanoutObjects( Aig_Man_t * pAig, int nCandMax, int fCisOnly )
func (pAig *AigManager) LlbManCollectHighFanoutObjects(nCandMax int, fCisOnly int) *IntVector {
	panic("TODO, stubbed")
}

// Llb_ManModelCheckAigWithHints
// C: int Llb_ManModelCheckAigWithHints( Aig_Man_t * pAigGlo, Gia_ParLlb_t * pPars )
func (pAigGlo *AigManager) LlbManModelCheckAigWithHints(pPars *GiaParLlb) int {
	panic("TODO, stubbed")
}
