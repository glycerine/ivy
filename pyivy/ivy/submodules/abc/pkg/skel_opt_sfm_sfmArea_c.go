package abc

// Hand-checked skeleton of src/opt/sfm/sfmArea.c.

// Abc_NtkPrecomputeCellPairs
// C: Vec_Int_t * Abc_NtkPrecomputeCellPairs( Mio_Cell2_t * pCells, int nCells )
func AbcNtkPrecomputeCellPairs(pCells []MioCell2, nCells int) *IntVector {
	panic("TODO, stubbed")
}

// Abc_NtkPrecomputeFirsts
// C: Vec_Int_t * Abc_NtkPrecomputeFirsts( Mio_Cell2_t * pCells, int nCells )
func AbcNtkPrecomputeFirsts(pCells []MioCell2, nCells int) *IntVector {
	panic("TODO, stubbed")
}

// Abc_NtkPrecomputePrint
// C: int Abc_NtkPrecomputePrint( Mio_Cell2_t * pCells, int nCells, Vec_Int_t * vInfo )
func AbcNtkPrecomputePrint(pCells []MioCell2, nCells int, vInfo *IntVector) int {
	panic("TODO, stubbed")
}

// Abc_NtkPrecomputeCellPairsTest
// C: void Abc_NtkPrecomputeCellPairsTest()
func AbcNtkPrecomputeCellPairsTest() {
	panic("TODO, stubbed")
}

// Abc_NodeCheckFanoutHasFanin
// C: int Abc_NodeCheckFanoutHasFanin( Abc_Obj_t * pNode, Abc_Obj_t * pFanin )
func AbcNodeCheckFanoutHasFanin(pNode *ABCNetworkObject, pFanin *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_ObjHasDupFanins
// C: int Abc_ObjHasDupFanins( Abc_Obj_t * pObj )
func AbcObjHasDupFanins(pObj *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_ObjHasDupFanouts
// C: int Abc_ObjHasDupFanouts( Abc_Obj_t * pObj )
func AbcObjHasDupFanouts(pObj *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_ObjChangeEval (pfUseInv is an out parameter)
// C: int Abc_ObjChangeEval( Abc_Obj_t * pObj, Vec_Int_t * vInfo, Vec_Int_t * vFirst, int InvArea, int * pfUseInv )
func AbcObjChangeEval(pObj *ABCNetworkObject, vInfo *IntVector, vFirst *IntVector, InvArea int, pfUseInv *int) int {
	panic("TODO, stubbed")
}

// Abc_ObjChangeUpdate
// C: void Abc_ObjChangeUpdate( Abc_Obj_t * pObj, int iFanin, Mio_Cell2_t * pCells, int * pNodeInfo, Vec_Int_t * vTemp )
func AbcObjChangeUpdate(pObj *ABCNetworkObject, iFanin int, pCells []MioCell2, pNodeInfo []int, vTemp *IntVector) {
	panic("TODO, stubbed")
}

// Abc_ObjChangePerform
// C: void Abc_ObjChangePerform( Abc_Obj_t * pObj, Vec_Int_t * vInfo, Vec_Int_t * vFirst, int fUseInv, Vec_Int_t * vTemp, Vec_Ptr_t * vFanout, Vec_Ptr_t * vFanout2, Mio_Cell2_t * pCells )
func AbcObjChangePerform(pObj *ABCNetworkObject, vInfo *IntVector, vFirst *IntVector, fUseInv int, vTemp *IntVector, vFanout *PointerVector[ABCNetworkObject], vFanout2 *PointerVector[ABCNetworkObject], pCells []MioCell2) {
	panic("TODO, stubbed")
}

// Abc_NtkChangePerform
// C: void Abc_NtkChangePerform( Abc_Ntk_t * pNtk, int fVerbose )
func AbcNtkChangePerform(pNtk *ABCNetwork, fVerbose int) {
	panic("TODO, stubbed")
}
