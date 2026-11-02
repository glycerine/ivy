package abc

// Hand-checked skeleton of src/map/if/ifReduce.c.

// If_ManImproveMapping
// C: void If_ManImproveMapping( If_Man_t * p )
func (p *IfMan) IfManImproveMapping() {
	panic("TODO, stubbed")
}

// If_ManImproveExpand
// C: void If_ManImproveExpand( If_Man_t * p, int nLimit )
func (p *IfMan) ifManImproveExpand(nLimit int) {
	panic("TODO, stubbed")
}

// If_ManImproveCutCost
// C: int If_ManImproveCutCost( If_Man_t * p, Vec_Ptr_t * vFront )
func (p *IfMan) IfManImproveCutCost(vFront *PointerVector[IfObj]) int {
	panic("TODO, stubbed")
}

// If_ManImproveNodeExpand
// C: void If_ManImproveNodeExpand( If_Man_t * p, If_Obj_t * pObj, int nLimit, Vec_Ptr_t * vFront, Vec_Ptr_t * vFrontOld, Vec_Ptr_t * vVisited )
func (p *IfMan) ifManImproveNodeExpand(pObj *IfObj, nLimit int, vFront *PointerVector[IfObj], vFrontOld *PointerVector[IfObj], vVisited *PointerVector[IfObj]) {
	panic("TODO, stubbed")
}

// If_ManImproveMark_rec
// C: void If_ManImproveMark_rec( If_Man_t * p, If_Obj_t * pObj, Vec_Ptr_t * vVisited )
func (p *IfMan) IfManImproveMark_rec(pObj *IfObj, vVisited *PointerVector[IfObj]) {
	panic("TODO, stubbed")
}

// If_ManImproveNodePrepare
// C: void If_ManImproveNodePrepare( If_Man_t * p, If_Obj_t * pObj, int nLimit, Vec_Ptr_t * vFront, Vec_Ptr_t * vFrontOld, Vec_Ptr_t * vVisited )
func (p *IfMan) ifManImproveNodePrepare(pObj *IfObj, nLimit int, vFront *PointerVector[IfObj], vFrontOld *PointerVector[IfObj], vVisited *PointerVector[IfObj]) {
	panic("TODO, stubbed")
}

// If_ManImproveNodeUpdate
// C: void If_ManImproveNodeUpdate( If_Man_t * p, If_Obj_t * pObj, Vec_Ptr_t * vFront )
func (p *IfMan) ifManImproveNodeUpdate(pObj *IfObj, vFront *PointerVector[IfObj]) {
	panic("TODO, stubbed")
}

// If_ManImproveNodeWillGrow
// C: int If_ManImproveNodeWillGrow( If_Man_t * p, If_Obj_t * pObj )
func (p *IfMan) IfManImproveNodeWillGrow(pObj *IfObj) int {
	panic("TODO, stubbed")
}

// If_ManImproveNodeFaninCost
// C: int If_ManImproveNodeFaninCost( If_Man_t * p, If_Obj_t * pObj )
func (p *IfMan) IfManImproveNodeFaninCost(pObj *IfObj) int {
	panic("TODO, stubbed")
}

// If_ManImproveNodeFaninUpdate
// C: void If_ManImproveNodeFaninUpdate( If_Man_t * p, If_Obj_t * pObj, Vec_Ptr_t * vFront, Vec_Ptr_t * vVisited )
func (p *IfMan) IfManImproveNodeFaninUpdate(pObj *IfObj, vFront *PointerVector[IfObj], vVisited *PointerVector[IfObj]) {
	panic("TODO, stubbed")
}

// If_ManImproveNodeFaninCompact0
// C: int If_ManImproveNodeFaninCompact0( If_Man_t * p, If_Obj_t * pObj, int nLimit, Vec_Ptr_t * vFront, Vec_Ptr_t * vVisited )
func (p *IfMan) IfManImproveNodeFaninCompact0(pObj *IfObj, nLimit int, vFront *PointerVector[IfObj], vVisited *PointerVector[IfObj]) int {
	panic("TODO, stubbed")
}

// If_ManImproveNodeFaninCompact1
// C: int If_ManImproveNodeFaninCompact1( If_Man_t * p, If_Obj_t * pObj, int nLimit, Vec_Ptr_t * vFront, Vec_Ptr_t * vVisited )
func (p *IfMan) IfManImproveNodeFaninCompact1(pObj *IfObj, nLimit int, vFront *PointerVector[IfObj], vVisited *PointerVector[IfObj]) int {
	panic("TODO, stubbed")
}

// If_ManImproveNodeFaninCompact2
// C: int If_ManImproveNodeFaninCompact2( If_Man_t * p, If_Obj_t * pObj, int nLimit, Vec_Ptr_t * vFront, Vec_Ptr_t * vVisited )
func (p *IfMan) IfManImproveNodeFaninCompact2(pObj *IfObj, nLimit int, vFront *PointerVector[IfObj], vVisited *PointerVector[IfObj]) int {
	panic("TODO, stubbed")
}

// If_ManImproveNodeFaninCompact_int
// C: int If_ManImproveNodeFaninCompact_int( If_Man_t * p, If_Obj_t * pObj, int nLimit, Vec_Ptr_t * vFront, Vec_Ptr_t * vVisited )
func (p *IfMan) IfManImproveNodeFaninCompact_int(pObj *IfObj, nLimit int, vFront *PointerVector[IfObj], vVisited *PointerVector[IfObj]) int {
	panic("TODO, stubbed")
}

// If_ManImproveNodeFaninCompact
// C: void If_ManImproveNodeFaninCompact( If_Man_t * p, If_Obj_t * pObj, int nLimit, Vec_Ptr_t * vFront, Vec_Ptr_t * vVisited )
func (p *IfMan) ifManImproveNodeFaninCompact(pObj *IfObj, nLimit int, vFront *PointerVector[IfObj], vVisited *PointerVector[IfObj]) {
	panic("TODO, stubbed")
}
