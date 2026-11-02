package abc

// Hand-checked skeleton of src/bdd/llb/llb1Group.c.

// Llb_ManGroupAlloc
// C: Llb_Grp_t * Llb_ManGroupAlloc( Llb_Man_t * pMan )
func (pMan *LlbMan) LlbManGroupAlloc() *LlbGrp {
	panic("TODO, stubbed")
}

// Llb_ManGroupStop
// C: void Llb_ManGroupStop( Llb_Grp_t * p )
func LlbManGroupStop(p *LlbGrp) {
	panic("TODO, stubbed")
}

// Llb_ManGroupCollect_rec
// C: void Llb_ManGroupCollect_rec( Aig_Man_t * pAig, Aig_Obj_t * pObj, Vec_Ptr_t * vNodes )
func (pAig *AigManager) LlbManGroupCollect_rec(pObj *AigObject, vNodes *PointerVector[AigObject]) {
	panic("TODO, stubbed")
}

// Llb_ManGroupCollect
// C: Vec_Ptr_t * Llb_ManGroupCollect( Llb_Grp_t * pGroup )
func LlbManGroupCollect(pGroup *LlbGrp) *PointerVector[AigObject] {
	panic("TODO, stubbed")
}

// Llb_ManGroupCreate_rec
// C: void Llb_ManGroupCreate_rec( Aig_Man_t * pAig, Aig_Obj_t * pObj, Vec_Ptr_t * vSupp )
func (pAig *AigManager) LlbManGroupCreate_rec(pObj *AigObject, vSupp *PointerVector[AigObject]) {
	panic("TODO, stubbed")
}

// Llb_ManGroupCreate
// C: Llb_Grp_t * Llb_ManGroupCreate( Llb_Man_t * pMan, Aig_Obj_t * pObj )
func (pMan *LlbMan) LlbManGroupCreate(pObj *AigObject) *LlbGrp {
	panic("TODO, stubbed")
}

// Llb_ManGroupCreateFirst
// C: Llb_Grp_t * Llb_ManGroupCreateFirst( Llb_Man_t * pMan )
func (pMan *LlbMan) LlbManGroupCreateFirst() *LlbGrp {
	panic("TODO, stubbed")
}

// Llb_ManGroupCreateLast
// C: Llb_Grp_t * Llb_ManGroupCreateLast( Llb_Man_t * pMan )
func (pMan *LlbMan) LlbManGroupCreateLast() *LlbGrp {
	panic("TODO, stubbed")
}

// Llb_ManGroupsCombine
// C: Llb_Grp_t * Llb_ManGroupsCombine( Llb_Grp_t * p1, Llb_Grp_t * p2 )
func LlbManGroupsCombine(p1 *LlbGrp, p2 *LlbGrp) *LlbGrp {
	panic("TODO, stubbed")
}

// Llb_ManGroupMarkNodes_rec
// C: void Llb_ManGroupMarkNodes_rec( Aig_Man_t * p, Aig_Obj_t * pObj )
func (p *AigManager) LlbManGroupMarkNodes_rec(pObj *AigObject) {
	panic("TODO, stubbed")
}

// Llb_ManGroupCreateFromCuts
// C: Llb_Grp_t * Llb_ManGroupCreateFromCuts( Llb_Man_t * pMan, Vec_Int_t * vCut1, Vec_Int_t * vCut2 )
func (pMan *LlbMan) LlbManGroupCreateFromCuts(vCut1 *IntVector, vCut2 *IntVector) *LlbGrp {
	panic("TODO, stubbed")
}

// Llb_ManPrepareGroups
// C: void Llb_ManPrepareGroups( Llb_Man_t * pMan )
func (pMan *LlbMan) LlbManPrepareGroups() {
	panic("TODO, stubbed")
}

// Llb_ManPrintSpan
// C: void Llb_ManPrintSpan( Llb_Man_t * p )
func (p *LlbMan) LlbManPrintSpan() {
	panic("TODO, stubbed")
}

// Llb_ManGroupHasVar
// C: int Llb_ManGroupHasVar( Llb_Man_t * p, int iGroup, int iVar )
func (p *LlbMan) LlbManGroupHasVar(iGroup int, iVar int) int {
	panic("TODO, stubbed")
}

// Llb_ManPrintHisto
// C: void Llb_ManPrintHisto( Llb_Man_t * p )
func (p *LlbMan) LlbManPrintHisto() {
	panic("TODO, stubbed")
}
