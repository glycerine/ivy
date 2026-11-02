package abc

// Hand-checked skeleton of src/aig/gia/giaIso.c.

// GiaIsoMan is Gia_IsoMan_t.
type GiaIsoMan struct {
	pGia       *GiaManager
	nObjs      int
	nUniques   int
	nSingles   int
	nEntries   int
	pLevels    []int    // int *
	pUniques   []int    // int *
	pStoreW    []uint64 // word *
	pStoreU    []uint32 // unsigned *
	vLevCounts *IntVector
	vClasses   *IntVector
	vClasses2  *IntVector
	timeStart  int64 // abctime
	timeSim    int64 // abctime
	timeRefine int64 // abctime
	timeSort   int64 // abctime
	timeOther  int64 // abctime
	timeTotal  int64 // abctime
}

// Gia_IsoGetValue (static inline)
// C: unsigned Gia_IsoGetValue( Gia_IsoMan_t * p, int i )
func (p *GiaIsoMan) giaIsoGetValue(i int) uint32 {
	panic("TODO, stubbed")
}

// Gia_IsoGetItem (static inline)
// C: unsigned Gia_IsoGetItem( Gia_IsoMan_t * p, int i )
func (p *GiaIsoMan) giaIsoGetItem(i int) uint32 {
	panic("TODO, stubbed")
}

// Gia_IsoSetValue (static inline)
// C: void Gia_IsoSetValue( Gia_IsoMan_t * p, int i, unsigned v )
func (p *GiaIsoMan) giaIsoSetValue(i int, v uint32) {
	panic("TODO, stubbed")
}

// Gia_IsoSetItem (static inline)
// C: void Gia_IsoSetItem( Gia_IsoMan_t * p, int i, unsigned v )
func (p *GiaIsoMan) giaIsoSetItem(i int, v uint32) {
	panic("TODO, stubbed")
}

// Gia_IsoManStart
// C: Gia_IsoMan_t * Gia_IsoManStart( Gia_Man_t * pGia )
func (pGia *GiaManager) GiaIsoManStart() *GiaIsoMan {
	panic("TODO, stubbed")
}

// Gia_IsoManStop
// C: void Gia_IsoManStop( Gia_IsoMan_t * p )
func (p *GiaIsoMan) GiaIsoManStop() {
	panic("TODO, stubbed")
}

// Gia_IsoManTransferUnique
// C: void Gia_IsoManTransferUnique( Gia_IsoMan_t * p )
func (p *GiaIsoMan) GiaIsoManTransferUnique() {
	panic("TODO, stubbed")
}

// Gia_IsoPrintClasses
// C: void Gia_IsoPrintClasses( Gia_IsoMan_t * p )
func (p *GiaIsoMan) GiaIsoPrintClasses() {
	panic("TODO, stubbed")
}

// Gia_IsoPrint
// C: void Gia_IsoPrint( Gia_IsoMan_t * p, int Iter, abctime Time )
func (p *GiaIsoMan) GiaIsoPrint(Iter int, Time int64) {
	panic("TODO, stubbed")
}

// Gia_IsoPrepare
// C: void Gia_IsoPrepare( Gia_IsoMan_t * p )
func (p *GiaIsoMan) GiaIsoPrepare() {
	panic("TODO, stubbed")
}

// Gia_IsoAssignUnique
// C: void Gia_IsoAssignUnique( Gia_IsoMan_t * p )
func (p *GiaIsoMan) GiaIsoAssignUnique() {
	panic("TODO, stubbed")
}

// Gia_IsoSort
// C: int Gia_IsoSort( Gia_IsoMan_t * p )
func (p *GiaIsoMan) GiaIsoSort() int {
	panic("TODO, stubbed")
}

// Gia_IsoCollectCosClasses
// C: Vec_Ptr_t * Gia_IsoCollectCosClasses( Gia_IsoMan_t * p, int fVerbose )
func (p *GiaIsoMan) GiaIsoCollectCosClasses(fVerbose int) *PointerVector[IntVector] {
	panic("TODO, stubbed")
}

// Gia_IsoUpdateValue (static inline)
// C: unsigned Gia_IsoUpdateValue( int Value, int fCompl )
func giaIsoUpdateValue(Value int, fCompl int) uint32 {
	panic("TODO, stubbed")
}

// Gia_IsoUpdate (static inline)
// C: unsigned Gia_IsoUpdate( Gia_IsoMan_t * p, int Iter, int iObj, int fCompl )
func (p *GiaIsoMan) giaIsoUpdate(Iter int, iObj int, fCompl int) uint32 {
	panic("TODO, stubbed")
}

// Gia_IsoSimulate
// C: void Gia_IsoSimulate( Gia_IsoMan_t * p, int Iter )
func (p *GiaIsoMan) GiaIsoSimulate(Iter int) {
	panic("TODO, stubbed")
}

// Gia_IsoSimulateBack
// C: void Gia_IsoSimulateBack( Gia_IsoMan_t * p, int Iter )
func (p *GiaIsoMan) GiaIsoSimulateBack(Iter int) {
	panic("TODO, stubbed")
}

// Gia_IsoAssignOneClass2
// C: void Gia_IsoAssignOneClass2( Gia_IsoMan_t * p )
func (p *GiaIsoMan) GiaIsoAssignOneClass2() {
	panic("TODO, stubbed")
}

// Gia_IsoAssignOneClass3
// C: void Gia_IsoAssignOneClass3( Gia_IsoMan_t * p )
func (p *GiaIsoMan) GiaIsoAssignOneClass3() {
	panic("TODO, stubbed")
}

// Gia_IsoAssignOneClass
// C: void Gia_IsoAssignOneClass( Gia_IsoMan_t * p, int fVerbose )
func (p *GiaIsoMan) GiaIsoAssignOneClass(fVerbose int) {
	panic("TODO, stubbed")
}

// Gia_IsoReportTopmost
// C: void Gia_IsoReportTopmost( Gia_IsoMan_t * p )
func (p *GiaIsoMan) GiaIsoReportTopmost() {
	panic("TODO, stubbed")
}

// Gia_IsoRecognizeMuxes
// C: void Gia_IsoRecognizeMuxes( Gia_Man_t * pGia )
func (pGia *GiaManager) GiaIsoRecognizeMuxes() {
	panic("TODO, stubbed")
}

// Gia_IsoDeriveEquivPos
// C: Vec_Ptr_t * Gia_IsoDeriveEquivPos( Gia_Man_t * pGia, int fForward, int fVerbose )
func (pGia *GiaManager) GiaIsoDeriveEquivPos(fForward int, fVerbose int) *PointerVector[IntVector] {
	panic("TODO, stubbed")
}

// Gia_ObjCompareByValue
// C: int Gia_ObjCompareByValue( Gia_Obj_t ** pp1, Gia_Obj_t ** pp2 )
func GiaObjCompareByValue(pp1 **GiaObject, pp2 **GiaObject) int {
	panic("TODO, stubbed")
}

// Gia_ManFindCaninicalOrder_rec
// C: void Gia_ManFindCaninicalOrder_rec( Gia_Man_t * p, Gia_Obj_t * pObj, Vec_Int_t * vAnds )
func (p *GiaManager) GiaManFindCaninicalOrder_rec(pObj *GiaObject, vAnds *IntVector) {
	panic("TODO, stubbed")
}

// Gia_ManFindCaninicalOrder
// C: void Gia_ManFindCaninicalOrder( Gia_Man_t * p, Vec_Int_t * vCis, Vec_Int_t * vAnds, Vec_Int_t * vCos, Vec_Int_t ** pvPiPerm )
func (p *GiaManager) GiaManFindCaninicalOrder(vCis *IntVector, vAnds *IntVector, vCos *IntVector, pvPiPerm **IntVector) {
	panic("TODO, stubbed")
}

// Gia_ManIsoCanonicize
// C: Gia_Man_t * Gia_ManIsoCanonicize( Gia_Man_t * p, int fVerbose )
func (p *GiaManager) GiaManIsoCanonicize(fVerbose int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManIsoFindString
// C: Vec_Str_t * Gia_ManIsoFindString( Gia_Man_t * p, int iPo, int fVerbose, Vec_Int_t ** pvPiPerm )
func (p *GiaManager) GiaManIsoFindString(iPo int, fVerbose int, pvPiPerm **IntVector) *ByteVector {
	panic("TODO, stubbed")
}

// Vec_IntCountNonTrivial
// C: int Vec_IntCountNonTrivial( Vec_Ptr_t * vEquivs, int * pnUsed )
func VecIntCountNonTrivial(vEquivs *PointerVector[IntVector], pnUsed *int) int {
	panic("TODO, stubbed")
}

// Gia_ManIsoReduce
// C: Gia_Man_t * Gia_ManIsoReduce( Gia_Man_t * pInit, Vec_Ptr_t ** pvPosEquivs, Vec_Ptr_t ** pvPiPerms, int fEstimate, int fDualOut, int fVerbose, int fVeryVerbose )
func (pInit *GiaManager) GiaManIsoReduce(pvPosEquivs **PointerVector[IntVector], pvPiPerms **PointerVector[IntVector], fEstimate int, fDualOut int, fVerbose int, fVeryVerbose int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_IsoTestOld
// C: void Gia_IsoTestOld( Gia_Man_t * p, int fVerbose )
func (p *GiaManager) GiaIsoTestOld(fVerbose int) {
	panic("TODO, stubbed")
}

// Gia_IsoTestGenPerm
// C: Vec_Int_t * Gia_IsoTestGenPerm( int nPis )
func GiaIsoTestGenPerm(nPis int) *IntVector {
	panic("TODO, stubbed")
}

// Gia_IsoTest
// C: void Gia_IsoTest( Gia_Man_t * p, Abc_Cex_t * pCex, int fVerbose )
func (p *GiaManager) GiaIsoTest(pCex *Counterexample, fVerbose int) {
	panic("TODO, stubbed")
}
