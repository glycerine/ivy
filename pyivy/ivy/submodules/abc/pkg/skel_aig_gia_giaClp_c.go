package abc

// Hand-checked skeleton of src/aig/gia/giaClp.c.

// Gia_GetFakeNames
// C: Vec_Ptr_t * Gia_GetFakeNames( int nNames )
func GiaGetFakeNames(nNames int) *PointerVector[string] {
	panic("TODO, stubbed")
}

// Gia_ManRebuildIsop
// C: int Gia_ManRebuildIsop( DdManager * dd, DdNode * bLocal, Gia_Man_t * pNew, Vec_Int_t * vFanins, Vec_Str_t * vSop, Vec_Str_t * vCube )
func GiaManRebuildIsop(dd *CuddFoundation, bLocal *DdNode, pNew *GiaManager, vFanins *IntVector, vSop *ByteVector, vCube *ByteVector) int {
	panic("TODO, stubbed")
}

// Gia_ManRebuildNode
// C: int Gia_ManRebuildNode( Dsd_Manager_t * pManDsd, Dsd_Node_t * pNodeDsd, Gia_Man_t * pNew, DdManager * ddNew, Vec_Int_t * vFanins, Vec_Str_t * vSop, Vec_Str_t * vCube )
func GiaManRebuildNode(pManDsd *DSDManager, pNodeDsd *DSDNode, pNew *GiaManager, ddNew *CuddFoundation, vFanins *IntVector, vSop *ByteVector, vCube *ByteVector) int {
	panic("TODO, stubbed")
}

// Gia_ManRebuild
// C: Gia_Man_t * Gia_ManRebuild( Gia_Man_t * p, Dsd_Manager_t * pManDsd, DdManager * ddNew )
func (p *GiaManager) GiaManRebuild(pManDsd *DSDManager, ddNew *CuddFoundation) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManCollapseDeref
// C: void Gia_ManCollapseDeref( DdManager * dd, Vec_Ptr_t * vFuncs )
func GiaManCollapseDeref(dd *CuddFoundation, vFuncs *PointerVector[DdNode]) {
	panic("TODO, stubbed")
}

// Gia_ObjCollapseDeref
// C: void Gia_ObjCollapseDeref( Gia_Man_t * p, DdManager * dd, Vec_Ptr_t * vFuncs, int Id )
func (p *GiaManager) GiaObjCollapseDeref(dd *CuddFoundation, vFuncs *PointerVector[DdNode], Id int) {
	panic("TODO, stubbed")
}

// Gia_ManCollapse
// C: Vec_Ptr_t * Gia_ManCollapse( Gia_Man_t * p, DdManager * dd, int nBddLimit, int fVerbose )
func (p *GiaManager) GiaManCollapse(dd *CuddFoundation, nBddLimit int, fVerbose int) *PointerVector[DdNode] {
	panic("TODO, stubbed")
}

// Gia_ManCollapseTest
// C: Gia_Man_t * Gia_ManCollapseTest( Gia_Man_t * p, int fVerbose )
func (p *GiaManager) GiaManCollapseTest(fVerbose int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManCollapseTestTest
// C: void Gia_ManCollapseTestTest( Gia_Man_t * p )
func (p *GiaManager) GiaManCollapseTestTest() {
	panic("TODO, stubbed")
}
