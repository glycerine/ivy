package abc

// Hand-checked skeleton of src/base/abci/abcLutmin.c.

// Abc_ObjCheckAbsorb
// C: int Abc_ObjCheckAbsorb( Abc_Obj_t * pObj, Abc_Obj_t * pPivot, int nLutSize, Vec_Ptr_t * vFanins )
func AbcObjCheckAbsorb(pObj *ABCNetworkObject, pPivot *ABCNetworkObject, nLutSize int, vFanins *PointerVector[ABCNetworkObject]) int {
	panic("TODO, stubbed")
}

// Abc_NtkCheckAbsorb
// C: void Abc_NtkCheckAbsorb( Abc_Ntk_t * pNtk, int nLutSize )
func (pNtk *ABCNetwork) AbcNtkCheckAbsorb(nLutSize int) {
	panic("TODO, stubbed")
}

// Abc_NtkBddMux21
// C: Abc_Obj_t * Abc_NtkBddMux21( Abc_Ntk_t * pNtkNew, Abc_Obj_t * pFanins[] )
func (pNtkNew *ABCNetwork) AbcNtkBddMux21(pFanins []*ABCNetworkObject) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_NtkBddMux411
// C: Abc_Obj_t * Abc_NtkBddMux411( Abc_Ntk_t * pNtkNew, Abc_Obj_t * pFanins[] )
func (pNtkNew *ABCNetwork) AbcNtkBddMux411(pFanins []*ABCNetworkObject) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_NtkBddMux412
// C: Abc_Obj_t * Abc_NtkBddMux412( Abc_Ntk_t * pNtkNew, Abc_Obj_t * pFanins[] )
func (pNtkNew *ABCNetwork) AbcNtkBddMux412(pFanins []*ABCNetworkObject) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_NtkBddMux412a
// C: Abc_Obj_t * Abc_NtkBddMux412a( Abc_Ntk_t * pNtkNew, Abc_Obj_t * pFanins[] )
func (pNtkNew *ABCNetwork) AbcNtkBddMux412a(pFanins []*ABCNetworkObject) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_NtkBddMux413
// C: Abc_Obj_t * Abc_NtkBddMux413( Abc_Ntk_t * pNtkNew, Abc_Obj_t * pFanins[] )
func (pNtkNew *ABCNetwork) AbcNtkBddMux413(pFanins []*ABCNetworkObject) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_NtkBddCofactors_rec
// C: DdNode * Abc_NtkBddCofactors_rec( DdManager * dd, DdNode * bNode, int iCof, int iLevel, int nLevels )
func AbcNtkBddCofactors_rec(dd *CuddFoundation, bNode *DdNode, iCof int, iLevel int, nLevels int) *DdNode {
	panic("TODO, stubbed")
}

// Abc_NtkBddCofactors
// C: Vec_Ptr_t * Abc_NtkBddCofactors( DdManager * dd, DdNode * bNode, int Level )
func AbcNtkBddCofactors(dd *CuddFoundation, bNode *DdNode, Level int) *PointerVector[DdNode] {
	panic("TODO, stubbed")
}

// Vec_PtrSortCompare (static)
// C: int Vec_PtrSortCompare( void ** pp1, void ** pp2 )
func vecPtrSortCompare(pp1 **DdNode, pp2 **DdNode) int {
	panic("TODO, stubbed")
}

// Abc_NtkCreateCofLut
// C: Abc_Obj_t * Abc_NtkCreateCofLut( Abc_Ntk_t * pNtkNew, DdManager * dd, DdNode * bCof, Abc_Obj_t * pNode, int Level )
func (pNtkNew *ABCNetwork) AbcNtkCreateCofLut(dd *CuddFoundation, bCof *DdNode, pNode *ABCNetworkObject, Level int) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_NtkBddCurtis
// C: Abc_Obj_t * Abc_NtkBddCurtis( Abc_Ntk_t * pNtkNew, Abc_Obj_t * pNode, Vec_Ptr_t * vCofs, Vec_Ptr_t * vUniq )
func (pNtkNew *ABCNetwork) AbcNtkBddCurtis(pNode *ABCNetworkObject, vCofs *PointerVector[DdNode], vUniq *PointerVector[DdNode]) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_NtkBddFindCofactor
// C: Abc_Obj_t * Abc_NtkBddFindCofactor( Abc_Ntk_t * pNtkNew, Abc_Obj_t * pNode, int nLutSize )
func (pNtkNew *ABCNetwork) AbcNtkBddFindCofactor(pNode *ABCNetworkObject, nLutSize int) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_NtkBddDecompose
// C: Abc_Obj_t * Abc_NtkBddDecompose( Abc_Ntk_t * pNtkNew, Abc_Obj_t * pNode, int nLutSize, int fVerbose )
func (pNtkNew *ABCNetwork) AbcNtkBddDecompose(pNode *ABCNetworkObject, nLutSize int, fVerbose int) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_NtkLutminConstruct
// C: void Abc_NtkLutminConstruct( Abc_Ntk_t * pNtkClp, Abc_Ntk_t * pNtkDec, int nLutSize, int fVerbose )
func (pNtkClp *ABCNetwork) AbcNtkLutminConstruct(pNtkDec *ABCNetwork, nLutSize int, fVerbose int) {
	panic("TODO, stubbed")
}

// Abc_NtkLutminInt
// C: Abc_Ntk_t * Abc_NtkLutminInt( Abc_Ntk_t * pNtk, int nLutSize, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkLutminInt(nLutSize int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkLutmin
// C: Abc_Ntk_t * Abc_NtkLutmin( Abc_Ntk_t * pNtkInit, int nLutSize, int fVerbose )
func (pNtkInit *ABCNetwork) AbcNtkLutmin(nLutSize int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}
