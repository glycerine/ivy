package abc

// Hand-checked skeleton of src/bdd/llb/llb1Constr.c.

// Llb_ManCountEntries
// C: int Llb_ManCountEntries( Vec_Int_t * vCands )
func LlbManCountEntries(vCands *IntVector) int {
	panic("TODO, stubbed")
}

// Llb_ManPrintEntries
// C: void Llb_ManPrintEntries( Aig_Man_t * p, Vec_Int_t * vCands )
func (p *AigManager) LlbManPrintEntries(vCands *IntVector) {
	panic("TODO, stubbed")
}

// Llb_ManDerefenceBdds
// C: void Llb_ManDerefenceBdds( Aig_Man_t * p, DdManager * dd )
func (p *AigManager) LlbManDerefenceBdds(dd *CuddFoundation) {
	panic("TODO, stubbed")
}

// Llb_ManComputeIndCase_rec
// C: DdNode * Llb_ManComputeIndCase_rec( Aig_Man_t * p, Aig_Obj_t * pObj, DdManager * dd, Vec_Ptr_t * vBdds )
func (p *AigManager) LlbManComputeIndCase_rec(pObj *AigObject, dd *CuddFoundation, vBdds *PointerVector[DdNode]) *DdNode {
	panic("TODO, stubbed")
}

// Llb_ManComputeIndCase
// C: void Llb_ManComputeIndCase( Aig_Man_t * p, DdManager * dd, Vec_Int_t * vNodes )
func (p *AigManager) LlbManComputeIndCase(dd *CuddFoundation, vNodes *IntVector) {
	panic("TODO, stubbed")
}

// Llb_ManComputeBaseCase
// C: Vec_Int_t * Llb_ManComputeBaseCase( Aig_Man_t * p, DdManager * dd )
func (p *AigManager) LlbManComputeBaseCase(dd *CuddFoundation) *IntVector {
	panic("TODO, stubbed")
}

// Llb_ManConstructGlobalBdds
// C: DdManager * Llb_ManConstructGlobalBdds( Aig_Man_t * p )
func (p *AigManager) LlbManConstructGlobalBdds() *CuddFoundation {
	panic("TODO, stubbed")
}

// Llb_ManDeriveConstraints
// C: Vec_Int_t * Llb_ManDeriveConstraints( Aig_Man_t * p )
func (p *AigManager) LlbManDeriveConstraints() *IntVector {
	panic("TODO, stubbed")
}

// Llb_ManConstrTest
// C: void Llb_ManConstrTest( Aig_Man_t * p )
func (p *AigManager) LlbManConstrTest() {
	panic("TODO, stubbed")
}
