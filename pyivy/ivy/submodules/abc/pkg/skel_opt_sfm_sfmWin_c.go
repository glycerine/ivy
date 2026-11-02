package abc

// Hand-checked skeleton of src/opt/sfm/sfmWin.c.

// Sfm_ObjRef_rec
// C: int Sfm_ObjRef_rec( Sfm_Ntk_t * p, int iObj )
func (p *SfmNtk) SfmObjRef_rec(iObj int) int {
	panic("TODO, stubbed")
}

// Sfm_ObjRef
// C: int Sfm_ObjRef( Sfm_Ntk_t * p, int iObj )
func (p *SfmNtk) SfmObjRef(iObj int) int {
	panic("TODO, stubbed")
}

// Sfm_ObjDeref_rec
// C: int Sfm_ObjDeref_rec( Sfm_Ntk_t * p, int iObj )
func (p *SfmNtk) SfmObjDeref_rec(iObj int) int {
	panic("TODO, stubbed")
}

// Sfm_ObjDeref
// C: int Sfm_ObjDeref( Sfm_Ntk_t * p, int iObj )
func (p *SfmNtk) SfmObjDeref(iObj int) int {
	panic("TODO, stubbed")
}

// Sfm_ObjMffcSize
// C: int Sfm_ObjMffcSize( Sfm_Ntk_t * p, int iObj )
func (p *SfmNtk) SfmObjMffcSize(iObj int) int {
	panic("TODO, stubbed")
}

// Sfm_NtkIncrementTravId (static inline)
// C: void Sfm_NtkIncrementTravId( Sfm_Ntk_t * p )
func (p *SfmNtk) sfmNtkIncrementTravId() {
	panic("TODO, stubbed")
}

// Sfm_ObjSetTravIdCurrent (static inline)
// C: void Sfm_ObjSetTravIdCurrent( Sfm_Ntk_t * p, int Id )
func (p *SfmNtk) sfmObjSetTravIdCurrent(Id int) {
	panic("TODO, stubbed")
}

// Sfm_ObjIsTravIdCurrent (static inline)
// C: int Sfm_ObjIsTravIdCurrent( Sfm_Ntk_t * p, int Id )
func (p *SfmNtk) sfmObjIsTravIdCurrent(Id int) int {
	panic("TODO, stubbed")
}

// Sfm_ObjIsTravIdPrevious (static inline)
// C: int Sfm_ObjIsTravIdPrevious( Sfm_Ntk_t * p, int Id )
func (p *SfmNtk) sfmObjIsTravIdPrevious(Id int) int {
	panic("TODO, stubbed")
}

// Sfm_NtkIncrementTravId2 (static inline)
// C: void Sfm_NtkIncrementTravId2( Sfm_Ntk_t * p )
func (p *SfmNtk) sfmNtkIncrementTravId2() {
	panic("TODO, stubbed")
}

// Sfm_ObjSetTravIdCurrent2 (static inline)
// C: void Sfm_ObjSetTravIdCurrent2( Sfm_Ntk_t * p, int Id )
func (p *SfmNtk) sfmObjSetTravIdCurrent2(Id int) {
	panic("TODO, stubbed")
}

// Sfm_ObjIsTravIdCurrent2 (static inline)
// C: int Sfm_ObjIsTravIdCurrent2( Sfm_Ntk_t * p, int Id )
func (p *SfmNtk) sfmObjIsTravIdCurrent2(Id int) int {
	panic("TODO, stubbed")
}

// Sfm_NtkDfs_rec
// C: void Sfm_NtkDfs_rec( Sfm_Ntk_t * p, int iNode, Vec_Int_t * vNodes, Vec_Wec_t * vGroups, Vec_Int_t * vGroupMap, Vec_Int_t * vBoxesLeft )
func (p *SfmNtk) SfmNtkDfs_rec(iNode int, vNodes *IntVector, vGroups *IntLevelVector, vGroupMap *IntVector, vBoxesLeft *IntVector) {
	panic("TODO, stubbed")
}

// Sfm_NtkDfs
// C: Vec_Int_t * Sfm_NtkDfs( Sfm_Ntk_t * p, Vec_Wec_t * vGroups, Vec_Int_t * vGroupMap, Vec_Int_t * vBoxesLeft, int fAllBoxes )
func (p *SfmNtk) SfmNtkDfs(vGroups *IntLevelVector, vGroupMap *IntVector, vBoxesLeft *IntVector, fAllBoxes int) *IntVector {
	panic("TODO, stubbed")
}

// Sfm_NtkCheckOverlap_rec
// C: int Sfm_NtkCheckOverlap_rec( Sfm_Ntk_t * p, int iThis, int iNode )
func (p *SfmNtk) SfmNtkCheckOverlap_rec(iThis int, iNode int) int {
	panic("TODO, stubbed")
}

// Sfm_NtkCheckOverlap
// C: int Sfm_NtkCheckOverlap( Sfm_Ntk_t * p, int iFan, int iNode )
func (p *SfmNtk) SfmNtkCheckOverlap(iFan int, iNode int) int {
	panic("TODO, stubbed")
}

// Sfm_NtkCheckRoot (static)
// C: int Sfm_NtkCheckRoot( Sfm_Ntk_t * p, int iNode, int nLevelMax )
func (p *SfmNtk) sfmNtkCheckRoot(iNode int, nLevelMax int) int {
	panic("TODO, stubbed")
}

// Sfm_NtkComputeRoots_rec
// C: void Sfm_NtkComputeRoots_rec( Sfm_Ntk_t * p, int iNode, int nLevelMax, Vec_Int_t * vRoots, Vec_Int_t * vTfo )
func (p *SfmNtk) SfmNtkComputeRoots_rec(iNode int, nLevelMax int, vRoots *IntVector, vTfo *IntVector) {
	panic("TODO, stubbed")
}

// Sfm_NtkAddDivisors
// C: void Sfm_NtkAddDivisors( Sfm_Ntk_t * p, int iNode, int nLevelMax )
func (p *SfmNtk) SfmNtkAddDivisors(iNode int, nLevelMax int) {
	panic("TODO, stubbed")
}

// Sfm_ObjIsUseful (static)
// C: int Sfm_ObjIsUseful( Sfm_Ntk_t * p, int iNode )
func (p *SfmNtk) sfmObjIsUseful(iNode int) int {
	panic("TODO, stubbed")
}

// Sfm_NtkCollectTfi_rec
// C: int Sfm_NtkCollectTfi_rec( Sfm_Ntk_t * p, int iNode, Vec_Int_t * vNodes )
func (p *SfmNtk) SfmNtkCollectTfi_rec(iNode int, vNodes *IntVector) int {
	panic("TODO, stubbed")
}

// Sfm_NtkCreateWindow
// C: int Sfm_NtkCreateWindow( Sfm_Ntk_t * p, int iNode, int fVerbose )
func (p *SfmNtk) SfmNtkCreateWindow(iNode int, fVerbose int) int {
	panic("TODO, stubbed")
}

// Sfm_NtkWindowTest
// C: void Sfm_NtkWindowTest( Sfm_Ntk_t * p, int iNode )
func (p *SfmNtk) SfmNtkWindowTest(iNode int) {
	panic("TODO, stubbed")
}
