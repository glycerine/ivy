package abc

// Hand-checked skeleton of src/base/abci/abcFraig.c.

// Abc_NtkFraig (pParams is a void * that holds a Fraig_Params_t *)
// C: Abc_Ntk_t * Abc_NtkFraig( Abc_Ntk_t * pNtk, void * pParams, int fAllNodes, int fExdc )
func (pNtk *ABCNetwork) AbcNtkFraig(pParams *FraigPar, fAllNodes int, fExdc int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkToFraig (returns a void * that holds a Fraig_Man_t *)
// C: void * Abc_NtkToFraig( Abc_Ntk_t * pNtk, void * pParams, int fAllNodes, int fExdc )
func (pNtk *ABCNetwork) AbcNtkToFraig(pParams *FraigPar, fAllNodes int, fExdc int) *FraigManager {
	panic("TODO, stubbed")
}

// Abc_NtkToFraigExdc (Fraig_Node_t * is a complement-tagged node pointer)
// C: Fraig_Node_t * Abc_NtkToFraigExdc( Fraig_Man_t * pMan, Abc_Ntk_t * pNtkMain, Abc_Ntk_t * pNtkExdc )
func AbcNtkToFraigExdc(pMan *FraigManager, pNtkMain *ABCNetwork, pNtkExdc *ABCNetwork) FraigEdge {
	panic("TODO, stubbed")
}

// Abc_NtkFraigRemapUsingExdc (static)
// C: void Abc_NtkFraigRemapUsingExdc( Fraig_Man_t * pMan, Abc_Ntk_t * pNtk )
func abcNtkFraigRemapUsingExdc(pMan *FraigManager, pNtk *ABCNetwork) {
	panic("TODO, stubbed")
}

// Abc_NtkFromFraig
// C: Abc_Ntk_t * Abc_NtkFromFraig( Fraig_Man_t * pMan, Abc_Ntk_t * pNtk )
func AbcNtkFromFraig(pMan *FraigManager, pNtk *ABCNetwork) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NodeFromFraig_rec (static)
// C: Abc_Obj_t * Abc_NodeFromFraig_rec( Abc_Ntk_t * pNtkNew, Fraig_Node_t * pNodeFraig )
func (pNtkNew *ABCNetwork) abcNodeFromFraig_rec(pNodeFraig FraigEdge) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_NtkFromFraig2 (static)
// C: Abc_Ntk_t * Abc_NtkFromFraig2( Fraig_Man_t * pMan, Abc_Ntk_t * pNtk )
func abcNtkFromFraig2(pMan *FraigManager, pNtk *ABCNetwork) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkFromFraig2_rec (static)
// C: void Abc_NtkFromFraig2_rec( Abc_Ntk_t * pNtkNew, Abc_Obj_t * pNode, Vec_Ptr_t * vNodeReprs )
func (pNtkNew *ABCNetwork) abcNtkFromFraig2_rec(pNode *ABCNetworkObject, vNodeReprs *PointerVector[ABCNetworkObject]) {
	panic("TODO, stubbed")
}

// Abc_NtkFraigTrust
// C: Abc_Ntk_t * Abc_NtkFraigTrust( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkFraigTrust() *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkFraigTrustCheck (static)
// C: int Abc_NtkFraigTrustCheck( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) abcNtkFraigTrustCheck() int {
	panic("TODO, stubbed")
}

// Abc_NtkFraigTrustOne (static)
// C: void Abc_NtkFraigTrustOne( Abc_Ntk_t * pNtk, Abc_Ntk_t * pNtkNew )
func (pNtk *ABCNetwork) abcNtkFraigTrustOne(pNtkNew *ABCNetwork) {
	panic("TODO, stubbed")
}

// Abc_NodeFraigTrust (static)
// C: Abc_Obj_t * Abc_NodeFraigTrust( Abc_Ntk_t * pNtkNew, Abc_Obj_t * pNode )
func (pNtkNew *ABCNetwork) abcNodeFraigTrust(pNode *ABCNetworkObject) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_NtkFraigStore
// C: int Abc_NtkFraigStore( Abc_Ntk_t * pNtkAdd )
func (pNtkAdd *ABCNetwork) AbcNtkFraigStore() int {
	panic("TODO, stubbed")
}

// Abc_NtkFraigRestore
// C: Abc_Ntk_t * Abc_NtkFraigRestore( int nPatsRand, int nPatsDyna, int nBTLimit )
func AbcNtkFraigRestore(nPatsRand int, nPatsDyna int, nBTLimit int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkFraigStoreClean
// C: void Abc_NtkFraigStoreClean()
func AbcNtkFraigStoreClean() {
	panic("TODO, stubbed")
}

// Abc_NtkFraigStoreCheck
// C: void Abc_NtkFraigStoreCheck( Abc_Ntk_t * pFraig )
func (pFraig *ABCNetwork) AbcNtkFraigStoreCheck() {
	panic("TODO, stubbed")
}
