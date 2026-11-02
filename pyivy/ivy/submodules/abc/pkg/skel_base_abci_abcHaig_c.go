package abc

// Hand-checked skeleton of src/base/abci/abcHaig.c.

// Abc_NtkHaigCollectMembers
// C: Vec_Ptr_t * Abc_NtkHaigCollectMembers( Hop_Man_t * p )
func (p *HopManager) AbcNtkHaigCollectMembers() *PointerVector[HopObject] {
	panic("TODO, stubbed")
}

// Abc_NtkHaigCreateClasses
// C: Vec_Ptr_t * Abc_NtkHaigCreateClasses( Vec_Ptr_t * vMembers )
func AbcNtkHaigCreateClasses(vMembers *PointerVector[HopObject]) *PointerVector[PointerVector[HopObject]] {
	panic("TODO, stubbed")
}

// Abc_NtkHaigCountFans
// C: int Abc_NtkHaigCountFans( Hop_Man_t * p )
func (p *HopManager) AbcNtkHaigCountFans() int {
	panic("TODO, stubbed")
}

// Hop_ObjReprHop (static)
// C: Hop_Obj_t * Hop_ObjReprHop( Hop_Obj_t * pObj )
func hopObjReprHop(pObj *HopObject) *HopObject {
	panic("TODO, stubbed")
}

// Hop_ObjChild0Hop (static)
// C: Hop_Obj_t * Hop_ObjChild0Hop( Hop_Obj_t * pObj )
func hopObjChild0Hop(pObj *HopObject) *HopObject {
	panic("TODO, stubbed")
}

// Hop_ObjChild1Hop (static)
// C: Hop_Obj_t * Hop_ObjChild1Hop( Hop_Obj_t * pObj )
func hopObjChild1Hop(pObj *HopObject) *HopObject {
	panic("TODO, stubbed")
}

// Abc_NtkHaigReconstruct
// C: Hop_Man_t * Abc_NtkHaigReconstruct( Hop_Man_t * p )
func (p *HopManager) AbcNtkHaigReconstruct() *HopManager {
	panic("TODO, stubbed")
}

// Abc_NtkHaigCheckTfi_rec
// C: int Abc_NtkHaigCheckTfi_rec( Abc_Obj_t * pNode, Abc_Obj_t * pOld )
func AbcNtkHaigCheckTfi_rec(pNode *ABCNetworkObject, pOld *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_NtkHaigCheckTfi
// C: int Abc_NtkHaigCheckTfi( Abc_Ntk_t * pNtk, Abc_Obj_t * pOld, Abc_Obj_t * pNew )
func (pNtk *ABCNetwork) AbcNtkHaigCheckTfi(pOld *ABCNetworkObject, pNew *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Hop_ObjChild0Next (static)
// C: Abc_Obj_t * Hop_ObjChild0Next( Hop_Obj_t * pObj )
func hopObjChild0Next(pObj *HopObject) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Hop_ObjChild1Next (static)
// C: Abc_Obj_t * Hop_ObjChild1Next( Hop_Obj_t * pObj )
func hopObjChild1Next(pObj *HopObject) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_NtkHaigRecreateAig
// C: Abc_Ntk_t * Abc_NtkHaigRecreateAig( Abc_Ntk_t * pNtk, Hop_Man_t * p )
func (pNtk *ABCNetwork) AbcNtkHaigRecreateAig(p *HopManager) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkHaigResetReprsOld
// C: void Abc_NtkHaigResetReprsOld( Hop_Man_t * pMan )
func (pMan *HopManager) AbcNtkHaigResetReprsOld() {
	panic("TODO, stubbed")
}

// Abc_NtkHaigResetReprs
// C: int Abc_NtkHaigResetReprs( Hop_Man_t * p )
func (p *HopManager) AbcNtkHaigResetReprs() int {
	panic("TODO, stubbed")
}

// Abc_NtkHopRemoveLoops
// C: Abc_Ntk_t * Abc_NtkHopRemoveLoops( Abc_Ntk_t * pNtk, Hop_Man_t * pMan )
func (pNtk *ABCNetwork) AbcNtkHopRemoveLoops(pMan *HopManager) *ABCNetwork {
	panic("TODO, stubbed")
}
