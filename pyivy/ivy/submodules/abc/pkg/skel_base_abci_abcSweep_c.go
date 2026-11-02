package abc

// Hand-checked skeleton of src/base/abci/abcSweep.c.

// Abc_NtkFraigSweep
// C: int Abc_NtkFraigSweep( Abc_Ntk_t * pNtk, int fUseInv, int fExdc, int fVerbose, int fVeryVerbose )
func (pNtk *ABCNetwork) AbcNtkFraigSweep(fUseInv int, fExdc int, fVerbose int, fVeryVerbose int) int {
	panic("TODO, stubbed")
}

// Abc_NtkFraigSweepUsingExdc (static)
// C: void Abc_NtkFraigSweepUsingExdc( Fraig_Man_t * pMan, Abc_Ntk_t * pNtk )
func abcNtkFraigSweepUsingExdc(pMan *FraigManager, pNtk *ABCNetwork) {
	panic("TODO, stubbed")
}

// Abc_NtkFraigEquiv (static; stmm_table * of node chains)
// C: stmm_table * Abc_NtkFraigEquiv( Abc_Ntk_t * pNtk, int fUseInv, int fVerbose, int fVeryVerbose )
func (pNtk *ABCNetwork) abcNtkFraigEquiv(fUseInv int, fVerbose int, fVeryVerbose int) *StableTable[*ABCNetworkObject, *ABCNetworkObject] {
	panic("TODO, stubbed")
}

// Abc_NtkFraigTransform (static)
// C: void Abc_NtkFraigTransform( Abc_Ntk_t * pNtk, stmm_table * tEquiv, int fUseInv, int fVerbose )
func (pNtk *ABCNetwork) abcNtkFraigTransform(tEquiv *StableTable[*ABCNetworkObject, *ABCNetworkObject], fUseInv int, fVerbose int) {
	panic("TODO, stubbed")
}

// Abc_NtkFraigMergeClassMapped (static)
// C: void Abc_NtkFraigMergeClassMapped( Abc_Ntk_t * pNtk, Abc_Obj_t * pChain, int fUseInv, int fVerbose )
func (pNtk *ABCNetwork) abcNtkFraigMergeClassMapped(pChain *ABCNetworkObject, fUseInv int, fVerbose int) {
	panic("TODO, stubbed")
}

// Abc_NtkFraigMergeClass (static)
// C: void Abc_NtkFraigMergeClass( Abc_Ntk_t * pNtk, Abc_Obj_t * pChain, int fUseInv, int fVerbose )
func (pNtk *ABCNetwork) abcNtkFraigMergeClass(pChain *ABCNetworkObject, fUseInv int, fVerbose int) {
	panic("TODO, stubbed")
}

// Abc_NodeDroppingCost (static)
// C: int Abc_NodeDroppingCost( Abc_Obj_t * pNode )
func abcNodeDroppingCost(pNode *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_NtkCleanup
// C: int Abc_NtkCleanup( Abc_Ntk_t * pNtk, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkCleanup(fVerbose int) int {
	panic("TODO, stubbed")
}

// Abc_NtkCleanupNodes
// C: int Abc_NtkCleanupNodes( Abc_Ntk_t * pNtk, Vec_Ptr_t * vRoots, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkCleanupNodes(vRoots *PointerVector[ABCNetworkObject], fVerbose int) int {
	panic("TODO, stubbed")
}

// Abc_NtkReduceNodes (static)
// C: int Abc_NtkReduceNodes( Abc_Ntk_t * pNtk, Vec_Ptr_t * vNodes )
func (pNtk *ABCNetwork) abcNtkReduceNodes(vNodes *PointerVector[ABCNetworkObject]) int {
	panic("TODO, stubbed")
}

// Abc_NodeConstantInput
// C: void Abc_NodeConstantInput( Abc_Obj_t * pNode, Abc_Obj_t * pFanin, int fConst0 )
func AbcNodeConstantInput(pNode *ABCNetworkObject, pFanin *ABCNetworkObject, fConst0 int) {
	panic("TODO, stubbed")
}

// Abc_NtkSweep
// C: int Abc_NtkSweep( Abc_Ntk_t * pNtk, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkSweep(fVerbose int) int {
	panic("TODO, stubbed")
}

// Abc_NodeRemoveNonCurrentObjects
// C: int Abc_NodeRemoveNonCurrentObjects( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNodeRemoveNonCurrentObjects() int {
	panic("TODO, stubbed")
}

// Abc_NtkSetTravId_rec
// C: void Abc_NtkSetTravId_rec( Abc_Obj_t * pObj )
func AbcNtkSetTravId_rec(pObj *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Abc_NtkCheckConstant_rec
// C: int Abc_NtkCheckConstant_rec( Abc_Obj_t * pObj )
func AbcNtkCheckConstant_rec(pObj *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_NtkLatchSweep
// C: int Abc_NtkLatchSweep( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkLatchSweep() int {
	panic("TODO, stubbed")
}

// Abc_NtkReplaceAutonomousLogic
// C: int Abc_NtkReplaceAutonomousLogic( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkReplaceAutonomousLogic() int {
	panic("TODO, stubbed")
}

// Abc_NtkCleanupSeq
// C: int Abc_NtkCleanupSeq( Abc_Ntk_t * pNtk, int fLatchSweep, int fAutoSweep, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkCleanupSeq(fLatchSweep int, fAutoSweep int, fVerbose int) int {
	panic("TODO, stubbed")
}

// Abc_NtkSweepBufsInvs
// C: int Abc_NtkSweepBufsInvs( Abc_Ntk_t * pNtk, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkSweepBufsInvs(fVerbose int) int {
	panic("TODO, stubbed")
}
