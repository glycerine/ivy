package abc

// Hand-checked skeleton of src/opt/mfs/mfsResub_.c.

// A stale alternate copy of mfsResub.c that defines the same function names;
// every name here carries a trailing underscore to stay distinct.

// Abc_NtkMfsUpdateNetwork (alternate copy; see mfsResub.c)
// C: void Abc_NtkMfsUpdateNetwork( Mfs_Man_t * p, Abc_Obj_t * pObj, Vec_Ptr_t * vFanins, Hop_Obj_t * pFunc )
func (p *MfsMan) AbcNtkMfsUpdateNetwork_(pObj *ABCNetworkObject, vFanins *PointerVector[ABCNetworkObject], pFunc *HopObject) {
	panic("TODO, stubbed")
}

// Abc_NtkMfsPrintResubStats (alternate copy; see mfsResub.c)
// C: void Abc_NtkMfsPrintResubStats( Mfs_Man_t * p )
func (p *MfsMan) AbcNtkMfsPrintResubStats_() {
	panic("TODO, stubbed")
}

// Abc_NtkMfsTryResubOnce (alternate copy; see mfsResub.c)
// C: int Abc_NtkMfsTryResubOnce( Mfs_Man_t * p, int * pCands, int nCands )
func (p *MfsMan) AbcNtkMfsTryResubOnce_(pCands []int, nCands int) int {
	panic("TODO, stubbed")
}

// Abc_NtkMfsSolveSatResub (alternate copy; see mfsResub.c)
// C: int Abc_NtkMfsSolveSatResub( Mfs_Man_t * p, Abc_Obj_t * pNode, int iFanin, int fOnlyRemove, int fSkipUpdate )
func (p *MfsMan) AbcNtkMfsSolveSatResub_(pNode *ABCNetworkObject, iFanin int, fOnlyRemove int, fSkipUpdate int) int {
	panic("TODO, stubbed")
}

// Abc_NtkMfsSolveSatResub2 (alternate copy; see mfsResub.c)
// C: int Abc_NtkMfsSolveSatResub2( Mfs_Man_t * p, Abc_Obj_t * pNode, int iFanin, int iFanin2 )
func (p *MfsMan) AbcNtkMfsSolveSatResub2_(pNode *ABCNetworkObject, iFanin int, iFanin2 int) int {
	panic("TODO, stubbed")
}

// Abc_NtkMfsEdgeSwapEval (alternate copy; see mfsResub.c)
// C: int Abc_NtkMfsEdgeSwapEval( Mfs_Man_t * p, Abc_Obj_t * pNode )
func (p *MfsMan) AbcNtkMfsEdgeSwapEval_(pNode *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_NtkMfsEdgePower (alternate copy; see mfsResub.c)
// C: int Abc_NtkMfsEdgePower( Mfs_Man_t * p, Abc_Obj_t * pNode )
func (p *MfsMan) AbcNtkMfsEdgePower_(pNode *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_NtkMfsResubNode (alternate copy; see mfsResub.c)
// C: int Abc_NtkMfsResubNode( Mfs_Man_t * p, Abc_Obj_t * pNode )
func (p *MfsMan) AbcNtkMfsResubNode_(pNode *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_NtkMfsResubNode2 (alternate copy; see mfsResub.c)
// C: int Abc_NtkMfsResubNode2( Mfs_Man_t * p, Abc_Obj_t * pNode )
func (p *MfsMan) AbcNtkMfsResubNode2_(pNode *ABCNetworkObject) int {
	panic("TODO, stubbed")
}
