package abc

// Hand-checked skeleton of src/aig/saig/saigConstr.c.

// Saig_DetectConstrCollectSuper_rec
// C: void Saig_DetectConstrCollectSuper_rec( Aig_Obj_t * pObj, Vec_Ptr_t * vSuper )
func SaigDetectConstrCollectSuper_rec(pObj *AigObject, vSuper *PointerVector[AigObject]) {
	panic("TODO, stubbed")
}

// Saig_DetectConstrCollectSuper
// C: Vec_Ptr_t * Saig_DetectConstrCollectSuper( Aig_Obj_t * pObj )
func SaigDetectConstrCollectSuper(pObj *AigObject) *PointerVector[AigObject] {
	panic("TODO, stubbed")
}

// Saig_ManDetectConstrCheckCont
// C: Vec_Ptr_t * Saig_ManDetectConstrCheckCont( Vec_Ptr_t * vSuper, Vec_Ptr_t * vSuper2 )
func SaigManDetectConstrCheckCont(vSuper *PointerVector[AigObject], vSuper2 *PointerVector[AigObject]) *PointerVector[AigObject] {
	panic("TODO, stubbed")
}

// Saig_ManDetectConstr
// C: int Saig_ManDetectConstr( Aig_Man_t * p, int iOut, Vec_Ptr_t ** pvOuts, Vec_Ptr_t ** pvCons )
func (p *AigManager) SaigManDetectConstr(iOut int, pvOuts **PointerVector[PointerVector[AigObject]], pvCons **PointerVector[PointerVector[AigObject]]) int {
	panic("TODO, stubbed")
}

// Saig_ManDupCompare
// C: int Saig_ManDupCompare( Aig_Obj_t ** pp1, Aig_Obj_t ** pp2 )
func SaigManDupCompare(pp1 **AigObject, pp2 **AigObject) int {
	panic("TODO, stubbed")
}

// Saig_ManDupUnfoldConstrs
// C: Aig_Man_t * Saig_ManDupUnfoldConstrs( Aig_Man_t * pAig )
func (pAig *AigManager) SaigManDupUnfoldConstrs() *AigManager {
	panic("TODO, stubbed")
}

// Saig_ManDupFoldConstrs
// C: Aig_Man_t * Saig_ManDupFoldConstrs( Aig_Man_t * pAig, Vec_Int_t * vConstrs )
func (pAig *AigManager) SaigManDupFoldConstrs(vConstrs *IntVector) *AigManager {
	panic("TODO, stubbed")
}

// Saig_ManFoldConstrTest
// C: void Saig_ManFoldConstrTest( Aig_Man_t * pAig )
func (pAig *AigManager) SaigManFoldConstrTest() {
	panic("TODO, stubbed")
}

// Saig_ManDetectConstrTest
// C: int Saig_ManDetectConstrTest( Aig_Man_t * p )
func (p *AigManager) SaigManDetectConstrTest() int {
	panic("TODO, stubbed")
}
