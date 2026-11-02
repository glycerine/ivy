package abc

// Hand-checked skeleton of src/base/abci/abcTim.c.

// Abc_NodeIsWhiteBox (static)
// C: int Abc_NodeIsWhiteBox( Abc_Obj_t * pObj )
func abcNodeIsWhiteBox(pObj *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_NtkTestTimNodeStrash_rec
// C: void Abc_NtkTestTimNodeStrash_rec( Gia_Man_t * pGia, Hop_Obj_t * pObj )
func (pGia *GiaManager) AbcNtkTestTimNodeStrash_rec(pObj *HopObject) {
	panic("TODO, stubbed")
}

// Abc_NtkTestTimNodeStrash
// C: int Abc_NtkTestTimNodeStrash( Gia_Man_t * pGia, Abc_Obj_t * pNode )
func (pGia *GiaManager) AbcNtkTestTimNodeStrash(pNode *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_NtkTestTimCollectCone_rec
// C: void Abc_NtkTestTimCollectCone_rec( Abc_Obj_t * pObj, Vec_Ptr_t * vNodes )
func AbcNtkTestTimCollectCone_rec(pObj *ABCNetworkObject, vNodes *PointerVector[ABCNetworkObject]) {
	panic("TODO, stubbed")
}

// Abc_NtkTestTimCollectCone
// C: Vec_Ptr_t * Abc_NtkTestTimCollectCone( Abc_Ntk_t * pNtk, Abc_Obj_t * pObj )
func (pNtk *ABCNetwork) AbcNtkTestTimCollectCone(pObj *ABCNetworkObject) *PointerVector[ABCNetworkObject] {
	panic("TODO, stubbed")
}

// Abc_NtkTestCreateArrivals
// C: Vec_Flt_t * Abc_NtkTestCreateArrivals( int nInputs )
func AbcNtkTestCreateArrivals(nInputs int) *FloatVector {
	panic("TODO, stubbed")
}

// Abc_NtkTestCreateRequired
// C: Vec_Flt_t * Abc_NtkTestCreateRequired( int nOutputs )
func AbcNtkTestCreateRequired(nOutputs int) *FloatVector {
	panic("TODO, stubbed")
}

// Abc_NtkTestTimDeriveGia
// C: Gia_Man_t * Abc_NtkTestTimDeriveGia( Abc_Ntk_t * pNtk, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkTestTimDeriveGia(fVerbose int) *GiaManager {
	panic("TODO, stubbed")
}

// Abc_NtkTestTimPerformSynthesis
// C: Gia_Man_t * Abc_NtkTestTimPerformSynthesis( Gia_Man_t * p, int fChoices )
func (p *GiaManager) AbcNtkTestTimPerformSynthesis(fChoices int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManVerifyChoices
// C: void Gia_ManVerifyChoices( Gia_Man_t * p )
func (p *GiaManager) GiaManVerifyChoices() {
	panic("TODO, stubbed")
}

// Gia_ManReverseClasses
// C: void Gia_ManReverseClasses( Gia_Man_t * p, int fNowIncreasing )
func (p *GiaManager) GiaManReverseClasses(fNowIncreasing int) {
	panic("TODO, stubbed")
}

// Abc_NtkTestTimByWritingFile
// C: void Abc_NtkTestTimByWritingFile( Gia_Man_t * pGia, char * pFileName )
func (pGia *GiaManager) AbcNtkTestTimByWritingFile(pFileName string) {
	panic("TODO, stubbed")
}

// Abc_NtkTestTim
// C: void Abc_NtkTestTim( Abc_Ntk_t * pNtk, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkTestTim(fVerbose int) {
	panic("TODO, stubbed")
}
