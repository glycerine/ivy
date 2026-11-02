package abc

// Hand-checked skeleton of src/base/abci/abcDsd.c.

// Abc_NtkDsdGlobal
// C: Abc_Ntk_t * Abc_NtkDsdGlobal( Abc_Ntk_t * pNtk, int fVerbose, int fPrint, int fShort )
func (pNtk *ABCNetwork) AbcNtkDsdGlobal(fVerbose int, fPrint int, fShort int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkDsdInternal (static)
// C: Abc_Ntk_t * Abc_NtkDsdInternal( Abc_Ntk_t * pNtk, int fVerbose, int fPrint, int fShort )
func (pNtk *ABCNetwork) abcNtkDsdInternal(fVerbose int, fPrint int, fShort int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkDsdConstruct (static)
// C: void Abc_NtkDsdConstruct( Dsd_Manager_t * pManDsd, Abc_Ntk_t * pNtk, Abc_Ntk_t * pNtkNew )
func abcNtkDsdConstruct(pManDsd *DSDManager, pNtk *ABCNetwork, pNtkNew *ABCNetwork) {
	panic("TODO, stubbed")
}

// Abc_NtkDsdConstructNode (static)
// C: Abc_Obj_t * Abc_NtkDsdConstructNode( Dsd_Manager_t * pManDsd, Dsd_Node_t * pNodeDsd, Abc_Ntk_t * pNtkNew, int * pCounters )
func abcNtkDsdConstructNode(pManDsd *DSDManager, pNodeDsd *DSDNode, pNtkNew *ABCNetwork, pCounters []int) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_NtkDsdLocal
// C: int Abc_NtkDsdLocal( Abc_Ntk_t * pNtk, int fVerbose, int fRecursive )
func (pNtk *ABCNetwork) AbcNtkDsdLocal(fVerbose int, fRecursive int) int {
	panic("TODO, stubbed")
}

// Abc_NtkCollectNodesForDsd (static)
// C: Vec_Ptr_t * Abc_NtkCollectNodesForDsd( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) abcNtkCollectNodesForDsd() *PointerVector[ABCNetworkObject] {
	panic("TODO, stubbed")
}

// Abc_NodeDecompDsdAndMux (static)
// C: void Abc_NodeDecompDsdAndMux( Abc_Obj_t * pNode, Vec_Ptr_t * vNodes, Dsd_Manager_t * pManDsd, int fRecursive, int * pCounters )
func abcNodeDecompDsdAndMux(pNode *ABCNetworkObject, vNodes *PointerVector[ABCNetworkObject], pManDsd *DSDManager, fRecursive int, pCounters []int) {
	panic("TODO, stubbed")
}

// Abc_NodeIsForDsd (static)
// C: int Abc_NodeIsForDsd( Abc_Obj_t * pNode )
func abcNodeIsForDsd(pNode *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_NodeFindMuxVar (static)
// C: int Abc_NodeFindMuxVar( DdManager * dd, DdNode * bFunc, int nVars )
func abcNodeFindMuxVar(dd *CuddFoundation, bFunc *DdNode, nVars int) int {
	panic("TODO, stubbed")
}

// Extra_bddComputeSum
// C: DdNode * Extra_bddComputeSum( DdManager * dd, DdNode ** pbCubes, int nCubes )
func ExtraBddComputeSum(dd *CuddFoundation, pbCubes []*DdNode, nCubes int) *DdNode {
	panic("TODO, stubbed")
}

// Abc_NtkSparsifyInternalOne
// C: DdNode * Abc_NtkSparsifyInternalOne( DdManager * ddNew, DdNode * bFunc, int nFanins, int nPerc )
func AbcNtkSparsifyInternalOne(ddNew *CuddFoundation, bFunc *DdNode, nFanins int, nPerc int) *DdNode {
	panic("TODO, stubbed")
}

// Abc_NtkSparsifyInternal
// C: Abc_Ntk_t * Abc_NtkSparsifyInternal( Abc_Ntk_t * pNtk, int nPerc, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkSparsifyInternal(nPerc int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkSparsify
// C: Abc_Ntk_t * Abc_NtkSparsify( Abc_Ntk_t * pNtk, int nPerc, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkSparsify(nPerc int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}
