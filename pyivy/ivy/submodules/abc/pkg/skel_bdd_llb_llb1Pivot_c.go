package abc

// Hand-checked skeleton of src/bdd/llb/llb1Pivot.c.

// Llb_ManTracePaths_rec
// C: int Llb_ManTracePaths_rec( Aig_Man_t * p, Aig_Obj_t * pObj, Aig_Obj_t * pPivot )
func (p *AigManager) LlbManTracePaths_rec(pObj *AigObject, pPivot *AigObject) int {
	panic("TODO, stubbed")
}

// Llb_ManTracePaths
// C: int Llb_ManTracePaths( Aig_Man_t * p, Aig_Obj_t * pPivot )
func (p *AigManager) LlbManTracePaths(pPivot *AigObject) int {
	panic("TODO, stubbed")
}

// Llb_ManTestCuts
// C: void Llb_ManTestCuts( Aig_Man_t * p )
func (p *AigManager) LlbManTestCuts() {
	panic("TODO, stubbed")
}

// Llb_ManLabelLiCones_rec
// C: void Llb_ManLabelLiCones_rec( Aig_Man_t * p, Aig_Obj_t * pObj )
func (p *AigManager) LlbManLabelLiCones_rec(pObj *AigObject) {
	panic("TODO, stubbed")
}

// Llb_ManLabelLiCones
// C: void Llb_ManLabelLiCones( Aig_Man_t * p )
func (p *AigManager) LlbManLabelLiCones() {
	panic("TODO, stubbed")
}

// Llb_ManMarkInternalPivots
// C: void Llb_ManMarkInternalPivots( Aig_Man_t * p )
func (p *AigManager) LlbManMarkInternalPivots() {
	panic("TODO, stubbed")
}

// Llb_ManMarkPivotNodes
// C: Vec_Int_t * Llb_ManMarkPivotNodes( Aig_Man_t * p, int fUseInternal )
func (p *AigManager) LlbManMarkPivotNodes(fUseInternal int) *IntVector {
	panic("TODO, stubbed")
}
