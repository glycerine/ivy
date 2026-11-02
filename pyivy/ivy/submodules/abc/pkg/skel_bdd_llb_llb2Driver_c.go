package abc

// Hand-checked skeleton of src/bdd/llb/llb2Driver.c.

// Llb_DriverCountRefs
// C: Vec_Int_t * Llb_DriverCountRefs( Aig_Man_t * p )
func (p *AigManager) LlbDriverCountRefs() *IntVector {
	panic("TODO, stubbed")
}

// Llb_DriverCollectNs
// C: Vec_Int_t * Llb_DriverCollectNs( Aig_Man_t * pAig, Vec_Int_t * vDriRefs )
func (pAig *AigManager) LlbDriverCollectNs(vDriRefs *IntVector) *IntVector {
	panic("TODO, stubbed")
}

// Llb_DriverCollectCs
// C: Vec_Int_t * Llb_DriverCollectCs( Aig_Man_t * pAig )
func (pAig *AigManager) LlbDriverCollectCs() *IntVector {
	panic("TODO, stubbed")
}

// Llb_DriverPhaseCube
// C: DdNode * Llb_DriverPhaseCube( Aig_Man_t * pAig, Vec_Int_t * vDriRefs, DdManager * dd )
func (pAig *AigManager) LlbDriverPhaseCube(vDriRefs *IntVector, dd *CuddFoundation) *DdNode {
	panic("TODO, stubbed")
}

// Llb_DriverLastPartition
// C: DdManager * Llb_DriverLastPartition( Aig_Man_t * p, Vec_Int_t * vVarsNs, abctime TimeTarget )
func (p *AigManager) LlbDriverLastPartition(vVarsNs *IntVector, TimeTarget int64) *CuddFoundation {
	panic("TODO, stubbed")
}
