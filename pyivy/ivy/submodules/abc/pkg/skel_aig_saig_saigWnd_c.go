package abc

// Hand-checked skeleton of src/aig/saig/saigWnd.c.

// Saig_ManWindowOutline_rec
// C: void Saig_ManWindowOutline_rec( Aig_Man_t * p, Aig_Obj_t * pObj, int nDist, Vec_Ptr_t * vNodes, int * pDists )
func (p *AigManager) SaigManWindowOutline_rec(pObj *AigObject, nDist int, vNodes *PointerVector[AigObject], pDists []int) {
	panic("TODO, stubbed")
}

// Saig_ManWindowOutline
// C: Vec_Ptr_t * Saig_ManWindowOutline( Aig_Man_t * p, Aig_Obj_t * pObj, int nDist )
func (p *AigManager) SaigManWindowOutline(pObj *AigObject, nDist int) *PointerVector[AigObject] {
	panic("TODO, stubbed")
}

// Saig_ObjHasUnlabeledFanout
// C: Aig_Obj_t * Saig_ObjHasUnlabeledFanout( Aig_Man_t * p, Aig_Obj_t * pObj )
func (p *AigManager) SaigObjHasUnlabeledFanout(pObj *AigObject) *AigObject {
	panic("TODO, stubbed")
}

// Saig_ManWindowCollectPis
// C: Vec_Ptr_t * Saig_ManWindowCollectPis( Aig_Man_t * p, Vec_Ptr_t * vNodes )
func (p *AigManager) SaigManWindowCollectPis(vNodes *PointerVector[AigObject]) *PointerVector[AigObject] {
	panic("TODO, stubbed")
}

// Saig_ManWindowCollectPos
// C: Vec_Ptr_t * Saig_ManWindowCollectPos( Aig_Man_t * p, Vec_Ptr_t * vNodes, Vec_Ptr_t ** pvPointers )
func (p *AigManager) SaigManWindowCollectPos(vNodes *PointerVector[AigObject], pvPointers **PointerVector[AigObject]) *PointerVector[AigObject] {
	panic("TODO, stubbed")
}

// Saig_ManWindowExtractNodes
// C: Aig_Man_t * Saig_ManWindowExtractNodes( Aig_Man_t * p, Vec_Ptr_t * vNodes )
func (p *AigManager) SaigManWindowExtractNodes(vNodes *PointerVector[AigObject]) *AigManager {
	panic("TODO, stubbed")
}

// Saig_ManWindowInsertBig_rec
// C: void Saig_ManWindowInsertBig_rec( Aig_Man_t * pNew, Aig_Obj_t * pObjBig, Vec_Ptr_t * vBigNode2SmallPo, Vec_Ptr_t * vSmallPi2BigNode )
func (pNew *AigManager) SaigManWindowInsertBig_rec(pObjBig *AigObject, vBigNode2SmallPo *PointerVector[AigObject], vSmallPi2BigNode *PointerVector[AigObject]) {
	panic("TODO, stubbed")
}

// Saig_ManWindowInsertSmall_rec
// C: void Saig_ManWindowInsertSmall_rec( Aig_Man_t * pNew, Aig_Obj_t * pObjSmall, Vec_Ptr_t * vBigNode2SmallPo, Vec_Ptr_t * vSmallPi2BigNode )
func (pNew *AigManager) saigManWindowInsertSmall_rec(pObjSmall *AigObject, vBigNode2SmallPo *PointerVector[AigObject], vSmallPi2BigNode *PointerVector[AigObject]) {
	panic("TODO, stubbed")
}

// Saig_ManWindowInsertNodes
// C: Aig_Man_t * Saig_ManWindowInsertNodes( Aig_Man_t * p, Vec_Ptr_t * vNodes, Aig_Man_t * pWnd )
func (p *AigManager) SaigManWindowInsertNodes(vNodes *PointerVector[AigObject], pWnd *AigManager) *AigManager {
	panic("TODO, stubbed")
}

// Saig_ManFindPivot
// C: Aig_Obj_t * Saig_ManFindPivot( Aig_Man_t * p )
func (p *AigManager) SaigManFindPivot() *AigObject {
	panic("TODO, stubbed")
}

// Saig_ManWindowExtract
// C: Aig_Man_t * Saig_ManWindowExtract( Aig_Man_t * p, Aig_Obj_t * pObj, int nDist )
func (p *AigManager) SaigManWindowExtract(pObj *AigObject, nDist int) *AigManager {
	panic("TODO, stubbed")
}

// Saig_ManWindowInsert
// C: Aig_Man_t * Saig_ManWindowInsert( Aig_Man_t * p, Aig_Obj_t * pObj, int nDist, Aig_Man_t * pWnd )
func (p *AigManager) SaigManWindowInsert(pObj *AigObject, nDist int, pWnd *AigManager) *AigManager {
	panic("TODO, stubbed")
}

// Saig_ManWindowTest
// C: Aig_Man_t * Saig_ManWindowTest( Aig_Man_t * p )
func (p *AigManager) SaigManWindowTest() *AigManager {
	panic("TODO, stubbed")
}

// Saig_ManCollectedDiffNodes
// C: Vec_Ptr_t * Saig_ManCollectedDiffNodes( Aig_Man_t * p0, Aig_Man_t * p1 )
func (p0 *AigManager) SaigManCollectedDiffNodes(p1 *AigManager) *PointerVector[AigObject] {
	panic("TODO, stubbed")
}

// Saig_ManWindowCreatePis
// C: void Saig_ManWindowCreatePis( Aig_Man_t * pNew, Aig_Man_t * p0, Aig_Man_t * p1, Vec_Ptr_t * vNodes0 )
func (pNew *AigManager) SaigManWindowCreatePis(p0 *AigManager, p1 *AigManager, vNodes0 *PointerVector[AigObject]) {
	panic("TODO, stubbed")
}

// Saig_ManWindowCreatePos
// C: void Saig_ManWindowCreatePos( Aig_Man_t * pNew, Aig_Man_t * p0, Aig_Man_t * p1 )
func (pNew *AigManager) SaigManWindowCreatePos(p0 *AigManager, p1 *AigManager) {
	panic("TODO, stubbed")
}

// Saig_ManWindowExtractMiter
// C: Aig_Man_t * Saig_ManWindowExtractMiter( Aig_Man_t * p0, Aig_Man_t * p1 )
func (p0 *AigManager) SaigManWindowExtractMiter(p1 *AigManager) *AigManager {
	panic("TODO, stubbed")
}
