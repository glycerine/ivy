package abc

// Hand-checked skeleton of src/base/abci/abcBalance.c.

// Abc_NtkBalance
// C: Abc_Ntk_t * Abc_NtkBalance( Abc_Ntk_t * pNtk, int fDuplicate, int fSelective, int fUpdateLevel )
func (pNtk *ABCNetwork) AbcNtkBalance(fDuplicate int, fSelective int, fUpdateLevel int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkBalancePerform (static)
// C: void Abc_NtkBalancePerform( Abc_Ntk_t * pNtk, Abc_Ntk_t * pNtkAig, int fDuplicate, int fSelective, int fUpdateLevel )
func (pNtk *ABCNetwork) abcNtkBalancePerform(pNtkAig *ABCNetwork, fDuplicate int, fSelective int, fUpdateLevel int) {
	panic("TODO, stubbed")
}

// Abc_NodeBalanceFindLeft
// C: int Abc_NodeBalanceFindLeft( Vec_Ptr_t * vSuper )
func AbcNodeBalanceFindLeft(vSuper *PointerVector[ABCNetworkObject]) int {
	panic("TODO, stubbed")
}

// Abc_NodeBalancePermute
// C: void Abc_NodeBalancePermute( Abc_Ntk_t * pNtkNew, Vec_Ptr_t * vSuper, int LeftBound )
func (pNtkNew *ABCNetwork) AbcNodeBalancePermute(vSuper *PointerVector[ABCNetworkObject], LeftBound int) {
	panic("TODO, stubbed")
}

// Abc_NodeBalance_rec (static)
// C: Abc_Obj_t * Abc_NodeBalance_rec( Abc_Ntk_t * pNtkNew, Abc_Obj_t * pNodeOld, Vec_Vec_t * vStorage, int Level, int fDuplicate, int fSelective, int fUpdateLevel )
func (pNtkNew *ABCNetwork) abcNodeBalance_rec(pNodeOld *ABCNetworkObject, vStorage *NestedVector[*ABCNetworkObject], Level int, fDuplicate int, fSelective int, fUpdateLevel int) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_NodeBalanceCone (static)
// C: Vec_Ptr_t * Abc_NodeBalanceCone( Abc_Obj_t * pNode, Vec_Vec_t * vStorage, int Level, int fDuplicate, int fSelective )
func abcNodeBalanceCone(pNode *ABCNetworkObject, vStorage *NestedVector[*ABCNetworkObject], Level int, fDuplicate int, fSelective int) *PointerVector[ABCNetworkObject] {
	panic("TODO, stubbed")
}

// Abc_NodeBalanceCone_rec (static)
// C: int Abc_NodeBalanceCone_rec( Abc_Obj_t * pNode, Vec_Ptr_t * vSuper, int fFirst, int fDuplicate, int fSelective )
func abcNodeBalanceCone_rec(pNode *ABCNetworkObject, vSuper *PointerVector[ABCNetworkObject], fFirst int, fDuplicate int, fSelective int) int {
	panic("TODO, stubbed")
}

// Abc_NodeBalanceConeExor_rec
// C: int Abc_NodeBalanceConeExor_rec( Abc_Obj_t * pNode, Vec_Ptr_t * vSuper, int fFirst )
func AbcNodeBalanceConeExor_rec(pNode *ABCNetworkObject, vSuper *PointerVector[ABCNetworkObject], fFirst int) int {
	panic("TODO, stubbed")
}

// Abc_NodeBalanceConeExor (static)
// C: Vec_Ptr_t * Abc_NodeBalanceConeExor( Abc_Obj_t * pNode )
func abcNodeBalanceConeExor(pNode *ABCNetworkObject) *PointerVector[ABCNetworkObject] {
	panic("TODO, stubbed")
}

// Abc_NodeFindCone_rec
// C: Vec_Ptr_t * Abc_NodeFindCone_rec( Abc_Obj_t * pNode )
func AbcNodeFindCone_rec(pNode *ABCNetworkObject) *PointerVector[ABCNetworkObject] {
	panic("TODO, stubbed")
}

// Abc_NtkBalanceAttach
// C: void Abc_NtkBalanceAttach( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkBalanceAttach() {
	panic("TODO, stubbed")
}

// Abc_NtkBalanceDetach
// C: void Abc_NtkBalanceDetach( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkBalanceDetach() {
	panic("TODO, stubbed")
}

// Abc_NtkBalanceLevel_rec
// C: int Abc_NtkBalanceLevel_rec( Abc_Obj_t * pNode )
func AbcNtkBalanceLevel_rec(pNode *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_NtkBalanceLevel
// C: void Abc_NtkBalanceLevel( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkBalanceLevel() {
	panic("TODO, stubbed")
}

// Abc_NtkMarkCriticalNodes (static)
// C: void Abc_NtkMarkCriticalNodes( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) abcNtkMarkCriticalNodes() {
	panic("TODO, stubbed")
}
