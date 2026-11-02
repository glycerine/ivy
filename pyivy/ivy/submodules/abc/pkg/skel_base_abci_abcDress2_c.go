package abc

// Hand-checked skeleton of src/base/abci/abcDress2.c.

// Abc_ObjEquivId2ObjId
// C: int Abc_ObjEquivId2ObjId( int EquivId )
func AbcObjEquivId2ObjId(EquivId int) int {
	panic("TODO, stubbed")
}

// Abc_ObjEquivId2Polar
// C: int Abc_ObjEquivId2Polar( int EquivId )
func AbcObjEquivId2Polar(EquivId int) int {
	panic("TODO, stubbed")
}

// Abc_ObjEquivId2NtkId
// C: int Abc_ObjEquivId2NtkId( int EquivId )
func AbcObjEquivId2NtkId(EquivId int) int {
	panic("TODO, stubbed")
}

// Aig_ManCreateDualOutputMiter
// C: Aig_Man_t * Aig_ManCreateDualOutputMiter( Aig_Man_t * p1, Aig_Man_t * p2 )
func (p1 *AigManager) AigManCreateDualOutputMiter(p2 *AigManager) *AigManager {
	panic("TODO, stubbed")
}

// Abc_NtkDressMapSetPolarity
// C: void Abc_NtkDressMapSetPolarity( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkDressMapSetPolarity() {
	panic("TODO, stubbed")
}

// Abc_NtkDressMapClasses
// C: Vec_Int_t * Abc_NtkDressMapClasses( Aig_Man_t * pMiter, Abc_Ntk_t * pNtk )
func (pMiter *AigManager) AbcNtkDressMapClasses(pNtk *ABCNetwork) *IntVector {
	panic("TODO, stubbed")
}

// Abc_ObjDressClass
// C: Vec_Int_t * Abc_ObjDressClass( Vec_Ptr_t * vRes, Vec_Int_t * vClass2Num, int Class )
func AbcObjDressClass(vRes *PointerVector[IntVector], vClass2Num *IntVector, Class int) *IntVector {
	panic("TODO, stubbed")
}

// Abc_ObjDressMakeId
// C: int Abc_ObjDressMakeId( Abc_Ntk_t * pNtk, int ObjId, int iNtk )
func (pNtk *ABCNetwork) AbcObjDressMakeId(ObjId int, iNtk int) int {
	panic("TODO, stubbed")
}

// Abc_NtkDressMapIds
// C: Vec_Ptr_t * Abc_NtkDressMapIds( Aig_Man_t * pMiter, Abc_Ntk_t * pNtk1, Abc_Ntk_t * pNtk2 )
func (pMiter *AigManager) AbcNtkDressMapIds(pNtk1 *ABCNetwork, pNtk2 *ABCNetwork) *PointerVector[IntVector] {
	panic("TODO, stubbed")
}

// Dch_ComputeEquivalences2
// C: void Dch_ComputeEquivalences2( Aig_Man_t * pMiter, Dch_Pars_t * pPars )
func (pMiter *AigManager) DchComputeEquivalences2(pPars *DchOptions) {
	panic("TODO, stubbed")
}

// Abc_NtkDressComputeEquivs
// C: Vec_Ptr_t * Abc_NtkDressComputeEquivs( Abc_Ntk_t * pNtk1, Abc_Ntk_t * pNtk2, int nConflictLimit, int fVerbose )
func (pNtk1 *ABCNetwork) AbcNtkDressComputeEquivs(pNtk2 *ABCNetwork, nConflictLimit int, fVerbose int) *PointerVector[IntVector] {
	panic("TODO, stubbed")
}

// Abc_NtkDressPrintEquivs
// C: void Abc_NtkDressPrintEquivs( Vec_Ptr_t * vRes )
func AbcNtkDressPrintEquivs(vRes *PointerVector[IntVector]) {
	panic("TODO, stubbed")
}

// Abc_NtkDressPrintStats
// C: void Abc_NtkDressPrintStats( Vec_Ptr_t * vRes, int nNodes0, int nNodes1, abctime Time )
func AbcNtkDressPrintStats(vRes *PointerVector[IntVector], nNodes0 int, nNodes1 int, Time int64) {
	panic("TODO, stubbed")
}

// Abc_NtkDress2Transfer
// C: void Abc_NtkDress2Transfer( Abc_Ntk_t * pNtk0, Abc_Ntk_t * pNtk1, Vec_Ptr_t * vRes, int fVerbose )
func (pNtk0 *ABCNetwork) AbcNtkDress2Transfer(pNtk1 *ABCNetwork, vRes *PointerVector[IntVector], fVerbose int) {
	panic("TODO, stubbed")
}

// Abc_NtkDress2
// C: void Abc_NtkDress2( Abc_Ntk_t * pNtk1, Abc_Ntk_t * pNtk2, int nConflictLimit, int fVerbose )
func (pNtk1 *ABCNetwork) AbcNtkDress2(pNtk2 *ABCNetwork, nConflictLimit int, fVerbose int) {
	panic("TODO, stubbed")
}
