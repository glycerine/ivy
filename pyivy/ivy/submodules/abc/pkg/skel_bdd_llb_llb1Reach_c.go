package abc

// Hand-checked skeleton of src/bdd/llb/llb1Reach.c.

// Llb_ManConstructOutBdd
// C: DdNode * Llb_ManConstructOutBdd( Aig_Man_t * pAig, Aig_Obj_t * pNode, DdManager * dd )
func (pAig *AigManager) LlbManConstructOutBdd(pNode *AigObject, dd *CuddFoundation) *DdNode {
	panic("TODO, stubbed")
}

// Llb_ManConstructGroupBdd
// C: DdNode * Llb_ManConstructGroupBdd( Llb_Man_t * p, Llb_Grp_t * pGroup )
func (p *LlbMan) LlbManConstructGroupBdd(pGroup *LlbGrp) *DdNode {
	panic("TODO, stubbed")
}

// Llb_ManConstructQuantCubeIntern
// C: DdNode * Llb_ManConstructQuantCubeIntern( Llb_Man_t * p, Llb_Grp_t * pGroup, int iGrpPlace, int fBackward )
func (p *LlbMan) LlbManConstructQuantCubeIntern(pGroup *LlbGrp, iGrpPlace int, fBackward int) *DdNode {
	panic("TODO, stubbed")
}

// Llb_ManConstructQuantCubeFwd
// C: DdNode * Llb_ManConstructQuantCubeFwd( Llb_Man_t * p, Llb_Grp_t * pGroup, int iGrpPlace )
func (p *LlbMan) LlbManConstructQuantCubeFwd(pGroup *LlbGrp, iGrpPlace int) *DdNode {
	panic("TODO, stubbed")
}

// Llb_ManConstructQuantCubeBwd
// C: DdNode * Llb_ManConstructQuantCubeBwd( Llb_Man_t * p, Llb_Grp_t * pGroup, int iGrpPlace )
func (p *LlbMan) LlbManConstructQuantCubeBwd(pGroup *LlbGrp, iGrpPlace int) *DdNode {
	panic("TODO, stubbed")
}

// Llb_ManComputeInitState
// C: DdNode * Llb_ManComputeInitState( Llb_Man_t * p, DdManager * dd )
func (p *LlbMan) LlbManComputeInitState(dd *CuddFoundation) *DdNode {
	panic("TODO, stubbed")
}

// Llb_ManComputeImage
// C: DdNode * Llb_ManComputeImage( Llb_Man_t * p, DdNode * bInit, int fBackward )
func (p *LlbMan) LlbManComputeImage(bInit *DdNode, fBackward int) *DdNode {
	panic("TODO, stubbed")
}

// Llb_ManCreateConstraints
// C: DdNode * Llb_ManCreateConstraints( Llb_Man_t * p, Vec_Int_t * vHints, int fUseNsVars )
func (p *LlbMan) LlbManCreateConstraints(vHints *IntVector, fUseNsVars int) *DdNode {
	panic("TODO, stubbed")
}

// Llb_ManReachDeriveCex
// C: Abc_Cex_t * Llb_ManReachDeriveCex( Llb_Man_t * p )
func (p *LlbMan) LlbManReachDeriveCex() *Counterexample {
	panic("TODO, stubbed")
}

// Llb_ManReachability
// C: int Llb_ManReachability( Llb_Man_t * p, Vec_Int_t * vHints, DdManager ** pddGlo )
func (p *LlbMan) LlbManReachability(vHints *IntVector, pddGlo **CuddFoundation) int {
	panic("TODO, stubbed")
}
