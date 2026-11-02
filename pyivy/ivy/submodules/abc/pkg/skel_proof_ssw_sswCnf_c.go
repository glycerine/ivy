package abc

// Hand-checked skeleton of src/proof/ssw/sswCnf.c.

// Ssw_SatStart
// C: Ssw_Sat_t * Ssw_SatStart( int fPolarFlip )
func SswSatStart(fPolarFlip int) *SswSat {
	panic("TODO, stubbed")
}

// Ssw_SatStop
// C: void Ssw_SatStop( Ssw_Sat_t * p )
func (p *SswSat) SswSatStop() {
	panic("TODO, stubbed")
}

// Ssw_AddClausesMux
// C: void Ssw_AddClausesMux( Ssw_Sat_t * p, Aig_Obj_t * pNode )
func (p *SswSat) SswAddClausesMux(pNode *AigObject) {
	panic("TODO, stubbed")
}

// Ssw_AddClausesSuper
// C: void Ssw_AddClausesSuper( Ssw_Sat_t * p, Aig_Obj_t * pNode, Vec_Ptr_t * vSuper )
func (p *SswSat) SswAddClausesSuper(pNode *AigObject, vSuper *PointerVector[AigObject]) {
	panic("TODO, stubbed")
}

// Ssw_CollectSuper_rec
// C: void Ssw_CollectSuper_rec( Aig_Obj_t * pObj, Vec_Ptr_t * vSuper, int fFirst, int fUseMuxes )
func SswCollectSuper_rec(pObj *AigObject, vSuper *PointerVector[AigObject], fFirst int, fUseMuxes int) {
	panic("TODO, stubbed")
}

// Ssw_CollectSuper
// C: void Ssw_CollectSuper( Aig_Obj_t * pObj, int fUseMuxes, Vec_Ptr_t * vSuper )
func SswCollectSuper(pObj *AigObject, fUseMuxes int, vSuper *PointerVector[AigObject]) {
	panic("TODO, stubbed")
}

// Ssw_ObjAddToFrontier
// C: void Ssw_ObjAddToFrontier( Ssw_Sat_t * p, Aig_Obj_t * pObj, Vec_Ptr_t * vFrontier )
func (p *SswSat) SswObjAddToFrontier(pObj *AigObject, vFrontier *PointerVector[AigObject]) {
	panic("TODO, stubbed")
}

// Ssw_CnfNodeAddToSolver
// C: void Ssw_CnfNodeAddToSolver( Ssw_Sat_t * p, Aig_Obj_t * pObj )
func (p *SswSat) SswCnfNodeAddToSolver(pObj *AigObject) {
	panic("TODO, stubbed")
}

// Ssw_CnfGetNodeValue
// C: int Ssw_CnfGetNodeValue( Ssw_Sat_t * p, Aig_Obj_t * pObj )
func (p *SswSat) SswCnfGetNodeValue(pObj *AigObject) int {
	panic("TODO, stubbed")
}
