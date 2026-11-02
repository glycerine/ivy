package abc

// Hand-checked skeleton of src/opt/mfs/mfsResub.c.

// Abc_NtkMfsUpdateNetwork
// C: void Abc_NtkMfsUpdateNetwork( Mfs_Man_t * p, Abc_Obj_t * pObj, Vec_Ptr_t * vMfsFanins, Hop_Obj_t * pFunc )
func (p *MfsMan) AbcNtkMfsUpdateNetwork(pObj *ABCNetworkObject, vMfsFanins *PointerVector[ABCNetworkObject], pFunc *HopObject) {
	panic("TODO, stubbed")
}

// Abc_NtkMfsPrintResubStats
// C: void Abc_NtkMfsPrintResubStats( Mfs_Man_t * p )
func (p *MfsMan) AbcNtkMfsPrintResubStats() {
	panic("TODO, stubbed")
}

// Abc_NtkMfsTryResubOnce
// C: int Abc_NtkMfsTryResubOnce( Mfs_Man_t * p, int * pCands, int nCands )
func (p *MfsMan) AbcNtkMfsTryResubOnce(pCands []int, nCands int) int {
	panic("TODO, stubbed")
}

// Abc_NtkMfsSolveSatResub
// C: int Abc_NtkMfsSolveSatResub( Mfs_Man_t * p, Abc_Obj_t * pNode, int iFanin, int fOnlyRemove, int fSkipUpdate )
func (p *MfsMan) AbcNtkMfsSolveSatResub(pNode *ABCNetworkObject, iFanin int, fOnlyRemove int, fSkipUpdate int) int {
	panic("TODO, stubbed")
}

// Abc_NtkMfsSolveSatResub2
// C: int Abc_NtkMfsSolveSatResub2( Mfs_Man_t * p, Abc_Obj_t * pNode, int iFanin, int iFanin2 )
func (p *MfsMan) AbcNtkMfsSolveSatResub2(pNode *ABCNetworkObject, iFanin int, iFanin2 int) int {
	panic("TODO, stubbed")
}

// Abc_NtkMfsEdgeSwapEval
// C: int Abc_NtkMfsEdgeSwapEval( Mfs_Man_t * p, Abc_Obj_t * pNode )
func (p *MfsMan) AbcNtkMfsEdgeSwapEval(pNode *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_NtkMfsEdgePower
// C: int Abc_NtkMfsEdgePower( Mfs_Man_t * p, Abc_Obj_t * pNode )
func (p *MfsMan) AbcNtkMfsEdgePower(pNode *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_NtkMfsResubNode
// C: int Abc_NtkMfsResubNode( Mfs_Man_t * p, Abc_Obj_t * pNode )
func (p *MfsMan) AbcNtkMfsResubNode(pNode *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_NtkMfsResubNode2
// C: int Abc_NtkMfsResubNode2( Mfs_Man_t * p, Abc_Obj_t * pNode )
func (p *MfsMan) AbcNtkMfsResubNode2(pNode *ABCNetworkObject) int {
	panic("TODO, stubbed")
}
