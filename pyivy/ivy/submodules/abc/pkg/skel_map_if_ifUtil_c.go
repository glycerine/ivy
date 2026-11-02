package abc

// Hand-checked skeleton of src/map/if/ifUtil.c.

// If_ManCleanNodeCopy
// C: void If_ManCleanNodeCopy( If_Man_t * p )
func (p *IfMan) IfManCleanNodeCopy() {
	panic("TODO, stubbed")
}

// If_ManCleanCutData
// C: void If_ManCleanCutData( If_Man_t * p )
func (p *IfMan) IfManCleanCutData() {
	panic("TODO, stubbed")
}

// If_ManCleanMarkV
// C: void If_ManCleanMarkV( If_Man_t * p )
func (p *IfMan) IfManCleanMarkV() {
	panic("TODO, stubbed")
}

// If_ManResetOriginalRefs
// C: void If_ManResetOriginalRefs( If_Man_t * p )
func (p *IfMan) IfManResetOriginalRefs() {
	panic("TODO, stubbed")
}

// If_ManCrossCut
// C: int If_ManCrossCut( If_Man_t * p )
func (p *IfMan) IfManCrossCut() int {
	panic("TODO, stubbed")
}

// If_ManReverseOrder
// C: Vec_Ptr_t * If_ManReverseOrder( If_Man_t * p )
func (p *IfMan) IfManReverseOrder() *PointerVector[IfObj] {
	panic("TODO, stubbed")
}

// If_ManMarkMapping_rec
// C: float If_ManMarkMapping_rec( If_Man_t * p, If_Obj_t * pObj )
func (p *IfMan) IfManMarkMapping_rec(pObj *IfObj) float32 {
	panic("TODO, stubbed")
}

// If_ManMarkMapping
// C: void If_ManMarkMapping( If_Man_t * p )
func (p *IfMan) IfManMarkMapping() {
	panic("TODO, stubbed")
}

// If_ManCollectMappingDirect
// C: Vec_Ptr_t * If_ManCollectMappingDirect( If_Man_t * p )
func (p *IfMan) IfManCollectMappingDirect() *PointerVector[IfObj] {
	panic("TODO, stubbed")
}

// If_ManCollectMappingInt
// C: Vec_Int_t * If_ManCollectMappingInt( If_Man_t * p )
func (p *IfMan) IfManCollectMappingInt() *IntVector {
	panic("TODO, stubbed")
}

// If_ManCountSpecialPos
// C: int If_ManCountSpecialPos( If_Man_t * p )
func (p *IfMan) IfManCountSpecialPos() int {
	panic("TODO, stubbed")
}

// If_CutTraverse_rec
// C: void If_CutTraverse_rec( If_Obj_t * pNode, Vec_Ptr_t * vNodes )
func ifCutTraverse_rec(pNode *IfObj, vNodes *PointerVector[IfObj]) {
	panic("TODO, stubbed")
}

// If_CutTraverse
// C: void If_CutTraverse( If_Man_t * p, If_Obj_t * pRoot, If_Cut_t * pCut, Vec_Ptr_t * vNodes )
func (p *IfMan) IfCutTraverse(pRoot *IfObj, pCut *IfCut, vNodes *PointerVector[IfObj]) {
	panic("TODO, stubbed")
}

// If_CutTraverseTest
// C: void If_CutTraverseTest( If_Man_t * p, If_Obj_t * pRoot, If_Cut_t * pCut )
func (p *IfMan) IfCutTraverseTest(pRoot *IfObj, pCut *IfCut) {
	panic("TODO, stubbed")
}

// If_ObjPrint
// C: void If_ObjPrint( If_Obj_t * pObj )
func IfObjPrint(pObj *IfObj) {
	panic("TODO, stubbed")
}
