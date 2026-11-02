package abc

// Hand-checked skeleton of src/base/abci/abcStrash.c.

// Abc_NtkRestrash
// C: Abc_Ntk_t * Abc_NtkRestrash( Abc_Ntk_t * pNtk, int fCleanup )
func (pNtk *ABCNetwork) AbcNtkRestrash(fCleanup int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkRestrashRandom_rec
// C: void Abc_NtkRestrashRandom_rec( Abc_Ntk_t * pNtk, Abc_Obj_t * pObj )
func (pNtk *ABCNetwork) AbcNtkRestrashRandom_rec(pObj *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Abc_NtkRestrashRandom
// C: Abc_Ntk_t * Abc_NtkRestrashRandom( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkRestrashRandom() *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkRestrashZero
// C: Abc_Ntk_t * Abc_NtkRestrashZero( Abc_Ntk_t * pNtk, int fCleanup )
func (pNtk *ABCNetwork) AbcNtkRestrashZero(fCleanup int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkStrash
// C: Abc_Ntk_t * Abc_NtkStrash( Abc_Ntk_t * pNtk, int fAllNodes, int fCleanup, int fRecord )
func (pNtk *ABCNetwork) AbcNtkStrash(fAllNodes int, fCleanup int, fRecord int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkAppend
// C: int Abc_NtkAppend( Abc_Ntk_t * pNtk1, Abc_Ntk_t * pNtk2, int fAddPos )
func (pNtk1 *ABCNetwork) AbcNtkAppend(pNtk2 *ABCNetwork, fAddPos int) int {
	panic("TODO, stubbed")
}

// Abc_NtkStrashPerform (static)
// C: void Abc_NtkStrashPerform( Abc_Ntk_t * pNtkOld, Abc_Ntk_t * pNtkNew, int fAllNodes, int fRecord )
func (pNtkOld *ABCNetwork) abcNtkStrashPerform(pNtkNew *ABCNetwork, fAllNodes int, fRecord int) {
	panic("TODO, stubbed")
}

// Abc_NodeStrash_rec (pMan is an Abc_Aig_t * (no Go struct is declared for it))
// C: void Abc_NodeStrash_rec( Abc_Aig_t * pMan, Hop_Obj_t * pObj )
func AbcNodeStrash_rec(pMan *AbcAig, pObj *HopObject) {
	panic("TODO, stubbed")
}

// Abc_NodeStrash
// C: Abc_Obj_t * Abc_NodeStrash( Abc_Ntk_t * pNtkNew, Abc_Obj_t * pNodeOld, int fRecord )
func (pNtkNew *ABCNetwork) AbcNodeStrash(pNodeOld *ABCNetworkObject, fRecord int) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_NtkTopmost_rec
// C: Abc_Obj_t * Abc_NtkTopmost_rec( Abc_Ntk_t * pNtkNew, Abc_Obj_t * pNode, int LevelCut )
func (pNtkNew *ABCNetwork) AbcNtkTopmost_rec(pNode *ABCNetworkObject, LevelCut int) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_NtkTopmost
// C: Abc_Ntk_t * Abc_NtkTopmost( Abc_Ntk_t * pNtk, int nLevels )
func (pNtk *ABCNetwork) AbcNtkTopmost(nLevels int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkBottommost_rec
// C: Abc_Obj_t * Abc_NtkBottommost_rec( Abc_Ntk_t * pNtkNew, Abc_Obj_t * pNode, int LevelCut )
func (pNtkNew *ABCNetwork) AbcNtkBottommost_rec(pNode *ABCNetworkObject, LevelCut int) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_NtkBottommost
// C: Abc_Ntk_t * Abc_NtkBottommost( Abc_Ntk_t * pNtk, int nLevels )
func (pNtk *ABCNetwork) AbcNtkBottommost(nLevels int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Vec_CompareNodeIds (static)
// C: int Vec_CompareNodeIds( Abc_Obj_t ** pp1, Abc_Obj_t ** pp2 )
func vecCompareNodeIds(pp1 **ABCNetworkObject, pp2 **ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_NodeGetSuper
// C: Vec_Ptr_t * Abc_NodeGetSuper( Abc_Obj_t * pNode )
func AbcNodeGetSuper(pNode *ABCNetworkObject) *PointerVector[ABCNetworkObject] {
	panic("TODO, stubbed")
}

// Abc_NtkTopAnd
// C: Abc_Ntk_t * Abc_NtkTopAnd( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkTopAnd() *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkWriteAig
// C: void Abc_NtkWriteAig( Abc_Ntk_t * pNtk, char * pFileName )
func (pNtk *ABCNetwork) AbcNtkWriteAig(pFileName string) {
	panic("TODO, stubbed")
}

// Abc_NtkPutOnTop
// C: Abc_Ntk_t * Abc_NtkPutOnTop( Abc_Ntk_t * pNtk, Abc_Ntk_t * pNtk2 )
func (pNtk *ABCNetwork) AbcNtkPutOnTop(pNtk2 *ABCNetwork) *ABCNetwork {
	panic("TODO, stubbed")
}
