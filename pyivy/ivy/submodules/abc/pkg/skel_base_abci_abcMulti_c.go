package abc

// Hand-checked skeleton of src/base/abci/abcMulti.c.

// Abc_NtkMulti
// C: Abc_Ntk_t * Abc_NtkMulti( Abc_Ntk_t * pNtk, int nThresh, int nFaninMax, int fCnf, int fMulti, int fSimple, int fFactor )
func (pNtk *ABCNetwork) AbcNtkMulti(nThresh int, nFaninMax int, fCnf int, fMulti int, fSimple int, fFactor int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkMultiInt (static)
// C: void Abc_NtkMultiInt( Abc_Ntk_t * pNtk, Abc_Ntk_t * pNtkNew )
func (pNtk *ABCNetwork) abcNtkMultiInt(pNtkNew *ABCNetwork) {
	panic("TODO, stubbed")
}

// Abc_NtkMulti_rec (static)
// C: Abc_Obj_t * Abc_NtkMulti_rec( Abc_Ntk_t * pNtkNew, Abc_Obj_t * pNodeOld )
func (pNtkNew *ABCNetwork) abcNtkMulti_rec(pNodeOld *ABCNetworkObject) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_NtkMultiDeriveBdd (static)
// C: DdNode * Abc_NtkMultiDeriveBdd( DdManager * dd, Abc_Obj_t * pNodeOld, Vec_Ptr_t * vFaninsOld )
func abcNtkMultiDeriveBdd(dd *CuddFoundation, pNodeOld *ABCNetworkObject, vFaninsOld *PointerVector[ABCNetworkObject]) *DdNode {
	panic("TODO, stubbed")
}

// Abc_NtkMultiDeriveBdd_rec (static)
// C: DdNode * Abc_NtkMultiDeriveBdd_rec( DdManager * dd, Abc_Obj_t * pNode, Vec_Ptr_t * vFanins )
func abcNtkMultiDeriveBdd_rec(dd *CuddFoundation, pNode *ABCNetworkObject, vFanins *PointerVector[ABCNetworkObject]) *DdNode {
	panic("TODO, stubbed")
}

// Abc_NtkMultiLimit_rec
// C: int Abc_NtkMultiLimit_rec( Abc_Obj_t * pNode, Vec_Ptr_t * vCone, int nFaninMax, int fCanStop, int fFirst )
func AbcNtkMultiLimit_rec(pNode *ABCNetworkObject, vCone *PointerVector[ABCNetworkObject], nFaninMax int, fCanStop int, fFirst int) int {
	panic("TODO, stubbed")
}

// Abc_NtkMultiLimit
// C: int Abc_NtkMultiLimit( Abc_Obj_t * pNode, Vec_Ptr_t * vCone, int nFaninMax )
func AbcNtkMultiLimit(pNode *ABCNetworkObject, vCone *PointerVector[ABCNetworkObject], nFaninMax int) int {
	panic("TODO, stubbed")
}

// Abc_NtkMultiSetBounds (static)
// C: void Abc_NtkMultiSetBounds( Abc_Ntk_t * pNtk, int nThresh, int nFaninMax )
func (pNtk *ABCNetwork) abcNtkMultiSetBounds(nThresh int, nFaninMax int) {
	panic("TODO, stubbed")
}

// Abc_NtkMultiSetBoundsCnf (static)
// C: void Abc_NtkMultiSetBoundsCnf( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) abcNtkMultiSetBoundsCnf() {
	panic("TODO, stubbed")
}

// Abc_NtkMultiSetBoundsMulti (static)
// C: void Abc_NtkMultiSetBoundsMulti( Abc_Ntk_t * pNtk, int nThresh )
func (pNtk *ABCNetwork) abcNtkMultiSetBoundsMulti(nThresh int) {
	panic("TODO, stubbed")
}

// Abc_NtkMultiSetBoundsSimple (static)
// C: void Abc_NtkMultiSetBoundsSimple( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) abcNtkMultiSetBoundsSimple() {
	panic("TODO, stubbed")
}

// Abc_NtkMultiSetBoundsFactor (static)
// C: void Abc_NtkMultiSetBoundsFactor( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) abcNtkMultiSetBoundsFactor() {
	panic("TODO, stubbed")
}

// Abc_NtkMultiCone_rec
// C: void Abc_NtkMultiCone_rec( Abc_Obj_t * pNode, Vec_Ptr_t * vCone )
func AbcNtkMultiCone_rec(pNode *ABCNetworkObject, vCone *PointerVector[ABCNetworkObject]) {
	panic("TODO, stubbed")
}

// Abc_NtkMultiCone (static)
// C: void Abc_NtkMultiCone( Abc_Obj_t * pNode, Vec_Ptr_t * vCone )
func abcNtkMultiCone(pNode *ABCNetworkObject, vCone *PointerVector[ABCNetworkObject]) {
	panic("TODO, stubbed")
}
