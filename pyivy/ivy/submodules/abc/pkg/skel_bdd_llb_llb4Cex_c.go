package abc

// Hand-checked skeleton of src/bdd/llb/llb4Cex.c.

// Llb4_Nonlin4TransformCex
// C: Abc_Cex_t * Llb4_Nonlin4TransformCex( Aig_Man_t * pAig, Vec_Ptr_t * vStates, int iCexPo, int fVerbose )
func (pAig *AigManager) Llb4Nonlin4TransformCex(vStates *PointerVector[uint32], iCexPo int, fVerbose int) *Counterexample {
	panic("TODO, stubbed")
}

// Llb4_Nonlin4VerifyCex
// C: Vec_Ptr_t * Llb4_Nonlin4VerifyCex( Aig_Man_t * pAig, Abc_Cex_t * p )
func (pAig *AigManager) Llb4Nonlin4VerifyCex(p *Counterexample) *PointerVector[uint32] {
	panic("TODO, stubbed")
}

// Llb4_Nonlin4NormalizeCex
// C: Abc_Cex_t * Llb4_Nonlin4NormalizeCex( Aig_Man_t * pAigOrg, Aig_Man_t * pAigRpm, Abc_Cex_t * pCexRpm )
func (pAigOrg *AigManager) Llb4Nonlin4NormalizeCex(pAigRpm *AigManager, pCexRpm *Counterexample) *Counterexample {
	panic("TODO, stubbed")
}
