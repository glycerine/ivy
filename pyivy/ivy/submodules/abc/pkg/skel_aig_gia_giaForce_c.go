package abc

// Hand-checked skeleton of src/aig/gia/giaForce.c.

// FrcObj is Frc_Obj_t.
type FrcObj struct {
	fCi      uint32 // unsigned : 1 : terminal node CI
	fCo      uint32 // unsigned : 1 : terminal node CO
	fMark0   uint32 // unsigned : 1 : first user-controlled mark
	fMark1   uint32 // unsigned : 1 : second user-controlled mark
	nFanins  uint32 // unsigned : 28 : the number of fanins
	nFanouts uint32 // the number of fanouts
	iFanout  uint32 // the current number of fanouts
	hHandle  int    // the handle of the node
	pPlace   int    // the placement of each node
	// union { float fEdgeCenter; unsigned iFanin; }
	fEdgeCenter float32 // center-of-gravity of the edge
	iFanin      uint32
	Fanios      []int // int Fanios[0] : the array of fanins/fanouts (flexible array member)
}

// FrcMan is Frc_Man_t.
type FrcMan struct {
	pGia     *GiaManager // the original AIG manager
	vCis     *IntVector  // the vector of CIs (PIs + LOs)
	vCos     *IntVector  // the vector of COs (POs + LIs)
	nObjs    int         // the number of objects
	nRegs    int         // the number of registers
	pObjData []int       // int * : the array containing data for objects
	nObjData int         // the size of array to store the logic network
	fVerbose int         // verbose output flag
	nCutCur  int         // current cut
	nCutMax  int         // max cut seen
}

// Frc_ManRegNum (static inline)
// C: int Frc_ManRegNum( Frc_Man_t * p )
func (p *FrcMan) frcManRegNum() int {
	panic("TODO, stubbed")
}

// Frc_ManCiNum (static inline)
// C: int Frc_ManCiNum( Frc_Man_t * p )
func (p *FrcMan) frcManCiNum() int {
	panic("TODO, stubbed")
}

// Frc_ManCoNum (static inline)
// C: int Frc_ManCoNum( Frc_Man_t * p )
func (p *FrcMan) frcManCoNum() int {
	panic("TODO, stubbed")
}

// Frc_ManPiNum (static inline)
// C: int Frc_ManPiNum( Frc_Man_t * p )
func (p *FrcMan) frcManPiNum() int {
	panic("TODO, stubbed")
}

// Frc_ManPoNum (static inline)
// C: int Frc_ManPoNum( Frc_Man_t * p )
func (p *FrcMan) frcManPoNum() int {
	panic("TODO, stubbed")
}

// Frc_ManObjNum (static inline)
// C: int Frc_ManObjNum( Frc_Man_t * p )
func (p *FrcMan) frcManObjNum() int {
	panic("TODO, stubbed")
}

// Frc_ManNodeNum (static inline)
// C: int Frc_ManNodeNum( Frc_Man_t * p )
func (p *FrcMan) frcManNodeNum() int {
	panic("TODO, stubbed")
}

// Frc_ManObj (static inline)
// C: Frc_Obj_t * Frc_ManObj( Frc_Man_t * p, int hHandle )
func (p *FrcMan) frcManObj(hHandle int) *FrcObj {
	panic("TODO, stubbed")
}

// Frc_ManCi (static inline)
// C: Frc_Obj_t * Frc_ManCi( Frc_Man_t * p, int i )
func (p *FrcMan) frcManCi(i int) *FrcObj {
	panic("TODO, stubbed")
}

// Frc_ManCo (static inline)
// C: Frc_Obj_t * Frc_ManCo( Frc_Man_t * p, int i )
func (p *FrcMan) frcManCo(i int) *FrcObj {
	panic("TODO, stubbed")
}

// Frc_ObjIsTerm (static inline)
// C: int Frc_ObjIsTerm( Frc_Obj_t * pObj )
func frcObjIsTerm(pObj *FrcObj) int {
	panic("TODO, stubbed")
}

// Frc_ObjIsCi (static inline)
// C: int Frc_ObjIsCi( Frc_Obj_t * pObj )
func frcObjIsCi(pObj *FrcObj) int {
	panic("TODO, stubbed")
}

// Frc_ObjIsCo (static inline)
// C: int Frc_ObjIsCo( Frc_Obj_t * pObj )
func frcObjIsCo(pObj *FrcObj) int {
	panic("TODO, stubbed")
}

// Frc_ObjIsPi (static inline)
// C: int Frc_ObjIsPi( Frc_Obj_t * pObj )
func frcObjIsPi(pObj *FrcObj) int {
	panic("TODO, stubbed")
}

// Frc_ObjIsPo (static inline)
// C: int Frc_ObjIsPo( Frc_Obj_t * pObj )
func frcObjIsPo(pObj *FrcObj) int {
	panic("TODO, stubbed")
}

// Frc_ObjIsNode (static inline)
// C: int Frc_ObjIsNode( Frc_Obj_t * pObj )
func frcObjIsNode(pObj *FrcObj) int {
	panic("TODO, stubbed")
}

// Frc_ObjIsConst0 (static inline)
// C: int Frc_ObjIsConst0( Frc_Obj_t * pObj )
func frcObjIsConst0(pObj *FrcObj) int {
	panic("TODO, stubbed")
}

// Frc_ObjSize (static inline)
// C: int Frc_ObjSize( Frc_Obj_t * pObj )
func frcObjSize(pObj *FrcObj) int {
	panic("TODO, stubbed")
}

// Frc_ObjFaninNum (static inline)
// C: int Frc_ObjFaninNum( Frc_Obj_t * pObj )
func frcObjFaninNum(pObj *FrcObj) int {
	panic("TODO, stubbed")
}

// Frc_ObjFanoutNum (static inline)
// C: int Frc_ObjFanoutNum( Frc_Obj_t * pObj )
func frcObjFanoutNum(pObj *FrcObj) int {
	panic("TODO, stubbed")
}

// Frc_ObjFanin (static inline)
// C: Frc_Obj_t * Frc_ObjFanin( Frc_Obj_t * pObj, int i )
func frcObjFanin(pObj *FrcObj, i int) *FrcObj {
	panic("TODO, stubbed")
}

// Frc_ObjFanout (static inline)
// C: Frc_Obj_t * Frc_ObjFanout( Frc_Obj_t * pObj, int i )
func frcObjFanout(pObj *FrcObj, i int) *FrcObj {
	panic("TODO, stubbed")
}

// Frc_ObjAddFanin
// C: void Frc_ObjAddFanin( Frc_Obj_t * pObj, Frc_Obj_t * pFanin )
func FrcObjAddFanin(pObj *FrcObj, pFanin *FrcObj) {
	panic("TODO, stubbed")
}

// Frc_ManStartSimple
// C: Frc_Man_t * Frc_ManStartSimple( Gia_Man_t * pGia )
func (pGia *GiaManager) FrcManStartSimple() *FrcMan {
	panic("TODO, stubbed")
}

// Frc_ManCollectSuper_rec
// C: void Frc_ManCollectSuper_rec( Gia_Man_t * p, Gia_Obj_t * pObj, Vec_Int_t * vSuper, Vec_Int_t * vVisit )
func (p *GiaManager) FrcManCollectSuper_rec(pObj *GiaObject, vSuper *IntVector, vVisit *IntVector) {
	panic("TODO, stubbed")
}

// Frc_ManCollectSuper
// C: void Frc_ManCollectSuper( Gia_Man_t * p, Gia_Obj_t * pObj, Vec_Int_t * vSuper, Vec_Int_t * vVisit )
func (p *GiaManager) FrcManCollectSuper(pObj *GiaObject, vSuper *IntVector, vVisit *IntVector) {
	panic("TODO, stubbed")
}

// Frc_ManCreateRefsSpecial
// C: void Frc_ManCreateRefsSpecial( Gia_Man_t * p )
func (p *GiaManager) FrcManCreateRefsSpecial() {
	panic("TODO, stubbed")
}

// Frc_ManTransformRefs
// C: void Frc_ManTransformRefs( Gia_Man_t * p, int * pnObjs, int * pnFanios )
func (p *GiaManager) FrcManTransformRefs(pnObjs *int, pnFanios *int) {
	panic("TODO, stubbed")
}

// Frc_ManStart
// C: Frc_Man_t * Frc_ManStart( Gia_Man_t * pGia )
func (pGia *GiaManager) FrcManStart() *FrcMan {
	panic("TODO, stubbed")
}

// Frc_ManPrintStats
// C: void Frc_ManPrintStats( Frc_Man_t * p )
func (p *FrcMan) FrcManPrintStats() {
	panic("TODO, stubbed")
}

// Frc_ManStop
// C: void Frc_ManStop( Frc_Man_t * p )
func (p *FrcMan) FrcManStop() {
	panic("TODO, stubbed")
}

// Frc_ManCrossCut_rec
// C: int Frc_ManCrossCut_rec( Frc_Man_t * p, Frc_Obj_t * pObj )
func (p *FrcMan) FrcManCrossCut_rec(pObj *FrcObj) int {
	panic("TODO, stubbed")
}

// Frc_ManCrossCut2_rec
// C: int Frc_ManCrossCut2_rec( Frc_Man_t * p, Frc_Obj_t * pObj )
func (p *FrcMan) FrcManCrossCut2_rec(pObj *FrcObj) int {
	panic("TODO, stubbed")
}

// Frc_ManCrossCut
// C: int Frc_ManCrossCut( Frc_Man_t * p, Vec_Int_t * vOrder, int fReverse )
func (p *FrcMan) FrcManCrossCut(vOrder *IntVector, fReverse int) int {
	panic("TODO, stubbed")
}

// Frc_ManCollectCos
// C: Vec_Int_t * Frc_ManCollectCos( Frc_Man_t * p )
func (p *FrcMan) FrcManCollectCos() *IntVector {
	panic("TODO, stubbed")
}

// Frc_ManCrossCutTest
// C: void Frc_ManCrossCutTest( Frc_Man_t * p, Vec_Int_t * vOrderInit )
func (p *FrcMan) FrcManCrossCutTest(vOrderInit *IntVector) {
	panic("TODO, stubbed")
}

// Frc_ManPlaceRandom
// C: void Frc_ManPlaceRandom( Frc_Man_t * p )
func (p *FrcMan) FrcManPlaceRandom() {
	panic("TODO, stubbed")
}

// Frc_ManArrayShuffle
// C: void Frc_ManArrayShuffle( Vec_Int_t * vArray )
func FrcManArrayShuffle(vArray *IntVector) {
	panic("TODO, stubbed")
}

// Frc_ManPlaceDfs_rec
// C: void Frc_ManPlaceDfs_rec( Frc_Man_t * p, Frc_Obj_t * pObj, int * piPlace )
func (p *FrcMan) FrcManPlaceDfs_rec(pObj *FrcObj, piPlace *int) {
	panic("TODO, stubbed")
}

// Frc_ManPlaceDfs
// C: void Frc_ManPlaceDfs( Frc_Man_t * p, Vec_Int_t * vCoOrder )
func (p *FrcMan) FrcManPlaceDfs(vCoOrder *IntVector) {
	panic("TODO, stubbed")
}

// Frc_ManPlaceDfsBoth
// C: int Frc_ManPlaceDfsBoth( Frc_Man_t * p, Vec_Int_t * vCoOrder, int * piCutSize2 )
func (p *FrcMan) FrcManPlaceDfsBoth(vCoOrder *IntVector, piCutSize2 *int) int {
	panic("TODO, stubbed")
}

// Frc_ManPlacementRefine
// C: void Frc_ManPlacementRefine( Frc_Man_t * p, int nIters, int fVerbose )
func (p *FrcMan) FrcManPlacementRefine(nIters int, fVerbose int) {
	panic("TODO, stubbed")
}

// Frc_ObjFanoutsAreCos
// C: int Frc_ObjFanoutsAreCos( Frc_Obj_t * pThis )
func FrcObjFanoutsAreCos(pThis *FrcObj) int {
	panic("TODO, stubbed")
}

// Frc_DumpGraphIntoFile
// C: void Frc_DumpGraphIntoFile( Frc_Man_t * p )
func (p *FrcMan) FrcDumpGraphIntoFile() {
	panic("TODO, stubbed")
}

// For_ManExperiment
// C: void For_ManExperiment( Gia_Man_t * pGia, int nIters, int fClustered, int fVerbose )
func (pGia *GiaManager) ForManExperiment(nIters int, fClustered int, fVerbose int) {
	panic("TODO, stubbed")
}

// For_ManFileExperiment
// C: void For_ManFileExperiment()
func ForManFileExperiment() {
	panic("TODO, stubbed")
}
