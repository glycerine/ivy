package abc

// Hand-checked skeleton of src/aig/gia/giaOf.c.

const (
	OF_LEAF_MAX  = 6
	OF_CUT_MAX   = 32
	OF_NO_LEAF   = 31
	OF_NO_FUNC   = 0x7FFFFFF
	OF_CUT_EXTRA = 4 // size; delay1, delay2; area
	OF_NUM       = 10
	OF_NUMINV    = 0.1
)

// OfCut is Of_Cut_t.
type OfCut struct {
	Sign    uint64               // word : signature
	Delay   int                  // delay
	Flow    int                  // flow
	iFunc   uint32               // unsigned : 27 : function (OF_NO_FUNC)
	nLeaves uint32               // unsigned : 5 : leaf number (OF_NO_LEAF)
	pLeaves [OF_LEAF_MAX + 1]int // leaves
}

// OfObj is Of_Obj_t.
type OfObj struct {
	iCutH    int // best cut
	iCutH2   int // best cut
	Delay1   int // arrival time
	Delay2   int // arrival time
	Required int // required
	nRefs    int // references
	Flow     int // area flow
	Temp     int // unused
}

// OfMan is Of_Man_t.
type OfMan struct {
	pGia       *GiaManager        // derived manager
	pPars      *JfPar             // parameters
	vTtMem     *WordMemory        // truth tables
	vPages     PointerVector[int] // Vec_Ptr_t (by value) : cut memory
	vCutSets   IntVector          // Vec_Int_t (by value) : cut offsets
	vCutFlows  IntVector          // Vec_Int_t (by value) : temporary cut area
	vCutDelays IntVector          // Vec_Int_t (by value) : temporary cut delay
	vCutRefs   IntVector          // Vec_Int_t (by value) : temporary cut referebces
	iCur       int                // current position
	Iter       int                // mapping iterations
	pObjs      []OfObj            // Of_Obj_t *
	clkStart   int64              // abctime : starting time
	CutCount   [6]float64         // double : cut counts
}

// Of_Flt2Int (static inline)
// C: int Of_Flt2Int( float f )
func ofFlt2Int(f float32) int {
	panic("TODO, stubbed")
}

// Of_Int2Flt (static inline)
// C: float Of_Int2Flt( int i )
func ofInt2Flt(i int) float32 {
	panic("TODO, stubbed")
}

// Of_ManCutSet (static inline)
// C: int * Of_ManCutSet( Of_Man_t * p, int i )
func (p *OfMan) ofManCutSet(i int) []int {
	panic("TODO, stubbed")
}

// Of_ObjCutSetId (static inline)
// C: int Of_ObjCutSetId( Of_Man_t * p, int i )
func (p *OfMan) ofObjCutSetId(i int) int {
	panic("TODO, stubbed")
}

// Of_ObjCutSet (static inline)
// C: int * Of_ObjCutSet( Of_Man_t * p, int i )
func (p *OfMan) ofObjCutSet(i int) []int {
	panic("TODO, stubbed")
}

// Of_ObjHasCuts (static inline)
// C: int Of_ObjHasCuts( Of_Man_t * p, int i )
func (p *OfMan) ofObjHasCuts(i int) int {
	panic("TODO, stubbed")
}

// Of_ObjCutFlow (static inline)
// C: int Of_ObjCutFlow( Of_Man_t * p, int i )
func (p *OfMan) ofObjCutFlow(i int) int {
	panic("TODO, stubbed")
}

// Of_ObjCutDelay (static inline)
// C: int Of_ObjCutDelay( Of_Man_t * p, int i )
func (p *OfMan) ofObjCutDelay(i int) int {
	panic("TODO, stubbed")
}

// Of_ObjSetCutFlow (static inline)
// C: void Of_ObjSetCutFlow( Of_Man_t * p, int i, int a )
func (p *OfMan) ofObjSetCutFlow(i int, a int) {
	panic("TODO, stubbed")
}

// Of_ObjSetCutDelay (static inline)
// C: void Of_ObjSetCutDelay( Of_Man_t * p, int i, int d )
func (p *OfMan) ofObjSetCutDelay(i int, d int) {
	panic("TODO, stubbed")
}

// Of_CutSize (static inline)
// C: int Of_CutSize( int * pCut )
func ofCutSize(pCut []int) int {
	panic("TODO, stubbed")
}

// Of_CutFunc (static inline)
// C: int Of_CutFunc( int * pCut )
func ofCutFunc(pCut []int) int {
	panic("TODO, stubbed")
}

// Of_CutLeaves (static inline)
// C: int * Of_CutLeaves( int * pCut )
func ofCutLeaves(pCut []int) []int {
	panic("TODO, stubbed")
}

// Of_CutSetBoth (static inline)
// C: int Of_CutSetBoth( int n, int f )
func ofCutSetBoth(n int, f int) int {
	panic("TODO, stubbed")
}

// Of_CutHandle (static inline)
// C: int Of_CutHandle( int * pCutSet, int * pCut )
func ofCutHandle(pCutSet []int, pCut []int) int {
	panic("TODO, stubbed")
}

// Of_CutFromHandle (static inline)
// C: int * Of_CutFromHandle( int * pCutSet, int h )
func ofCutFromHandle(pCutSet []int, h int) []int {
	panic("TODO, stubbed")
}

// Of_CutDelay1 (static inline)
// C: int Of_CutDelay1( int * pCut )
func ofCutDelay1(pCut []int) int {
	panic("TODO, stubbed")
}

// Of_CutDelay2 (static inline)
// C: int Of_CutDelay2( int * pCut )
func ofCutDelay2(pCut []int) int {
	panic("TODO, stubbed")
}

// Of_CutAreaFlow (static inline)
// C: int Of_CutAreaFlow( int * pCut )
func ofCutAreaFlow(pCut []int) int {
	panic("TODO, stubbed")
}

// Of_CutSetDelay1 (static inline)
// C: void Of_CutSetDelay1( int * pCut, int d )
func ofCutSetDelay1(pCut []int, d int) {
	panic("TODO, stubbed")
}

// Of_CutSetDelay2 (static inline)
// C: void Of_CutSetDelay2( int * pCut, int d )
func ofCutSetDelay2(pCut []int, d int) {
	panic("TODO, stubbed")
}

// Of_CutSetAreaFlow (static inline)
// C: void Of_CutSetAreaFlow( int * pCut, int d )
func ofCutSetAreaFlow(pCut []int, d int) {
	panic("TODO, stubbed")
}

// Of_CutVar (static inline)
// C: int Of_CutVar( int * pCut, int v )
func ofCutVar(pCut []int, v int) int {
	panic("TODO, stubbed")
}

// Of_CutFlag (static inline)
// C: int Of_CutFlag( int * pCut, int v )
func ofCutFlag(pCut []int, v int) int {
	panic("TODO, stubbed")
}

// Of_CutCleanFlag (static inline)
// C: void Of_CutCleanFlag( int * pCut, int v )
func ofCutCleanFlag(pCut []int, v int) {
	panic("TODO, stubbed")
}

// Of_CutSetFlag (static inline)
// C: void Of_CutSetFlag( int * pCut, int v, int x )
func ofCutSetFlag(pCut []int, v int, x int) {
	panic("TODO, stubbed")
}

// Of_ObjData (static inline)
// C: Of_Obj_t * Of_ObjData( Of_Man_t * p, int i )
func (p *OfMan) ofObjData(i int) *OfObj {
	panic("TODO, stubbed")
}

// Of_ObjCutBest (static inline)
// C: int Of_ObjCutBest( Of_Man_t * p, int i )
func (p *OfMan) ofObjCutBest(i int) int {
	panic("TODO, stubbed")
}

// Of_ObjCutBest2 (static inline)
// C: int Of_ObjCutBest2( Of_Man_t * p, int i )
func (p *OfMan) ofObjCutBest2(i int) int {
	panic("TODO, stubbed")
}

// Of_ObjDelay1 (static inline)
// C: int Of_ObjDelay1( Of_Man_t * p, int i )
func (p *OfMan) ofObjDelay1(i int) int {
	panic("TODO, stubbed")
}

// Of_ObjDelay2 (static inline)
// C: int Of_ObjDelay2( Of_Man_t * p, int i )
func (p *OfMan) ofObjDelay2(i int) int {
	panic("TODO, stubbed")
}

// Of_ObjRequired (static inline)
// C: int Of_ObjRequired( Of_Man_t * p, int i )
func (p *OfMan) ofObjRequired(i int) int {
	panic("TODO, stubbed")
}

// Of_ObjRefNum (static inline)
// C: int Of_ObjRefNum( Of_Man_t * p, int i )
func (p *OfMan) ofObjRefNum(i int) int {
	panic("TODO, stubbed")
}

// Of_ObjFlow (static inline)
// C: int Of_ObjFlow( Of_Man_t * p, int i )
func (p *OfMan) ofObjFlow(i int) int {
	panic("TODO, stubbed")
}

// Of_ObjSetCutBest (static inline)
// C: void Of_ObjSetCutBest( Of_Man_t * p, int i, int x )
func (p *OfMan) ofObjSetCutBest(i int, x int) {
	panic("TODO, stubbed")
}

// Of_ObjSetCutBest2 (static inline)
// C: void Of_ObjSetCutBest2( Of_Man_t * p, int i, int x )
func (p *OfMan) ofObjSetCutBest2(i int, x int) {
	panic("TODO, stubbed")
}

// Of_ObjSetDelay1 (static inline)
// C: void Of_ObjSetDelay1( Of_Man_t * p, int i, int x )
func (p *OfMan) ofObjSetDelay1(i int, x int) {
	panic("TODO, stubbed")
}

// Of_ObjSetDelay2 (static inline)
// C: void Of_ObjSetDelay2( Of_Man_t * p, int i, int x )
func (p *OfMan) ofObjSetDelay2(i int, x int) {
	panic("TODO, stubbed")
}

// Of_ObjSetRequired (static inline)
// C: void Of_ObjSetRequired( Of_Man_t * p, int i, int x )
func (p *OfMan) ofObjSetRequired(i int, x int) {
	panic("TODO, stubbed")
}

// Of_ObjSetRefNum (static inline)
// C: void Of_ObjSetRefNum( Of_Man_t * p, int i, int x )
func (p *OfMan) ofObjSetRefNum(i int, x int) {
	panic("TODO, stubbed")
}

// Of_ObjSetFlow (static inline)
// C: void Of_ObjSetFlow( Of_Man_t * p, int i, int x )
func (p *OfMan) ofObjSetFlow(i int, x int) {
	panic("TODO, stubbed")
}

// Of_ObjUpdateRequired (static inline)
// C: void Of_ObjUpdateRequired( Of_Man_t * p,int i, int x )
func (p *OfMan) ofObjUpdateRequired(i int, x int) {
	panic("TODO, stubbed")
}

// Of_ObjRefInc (static inline)
// C: int Of_ObjRefInc( Of_Man_t * p, int i )
func (p *OfMan) ofObjRefInc(i int) int {
	panic("TODO, stubbed")
}

// Of_ObjRefDec (static inline)
// C: int Of_ObjRefDec( Of_Man_t * p, int i )
func (p *OfMan) ofObjRefDec(i int) int {
	panic("TODO, stubbed")
}

// Of_ObjCutBestP (static inline)
// C: int * Of_ObjCutBestP( Of_Man_t * p, int iObj )
func (p *OfMan) ofObjCutBestP(iObj int) []int {
	panic("TODO, stubbed")
}

// Of_ObjSetCutBestP (static inline)
// C: void Of_ObjSetCutBestP( Of_Man_t * p, int * pCutSet, int iObj, int * pCut )
func (p *OfMan) ofObjSetCutBestP(pCutSet []int, iObj int, pCut []int) {
	panic("TODO, stubbed")
}

// Of_ObjCutBestP2 (static inline)
// C: int * Of_ObjCutBestP2( Of_Man_t * p, int iObj )
func (p *OfMan) ofObjCutBestP2(iObj int) []int {
	panic("TODO, stubbed")
}

// Of_ObjSetCutBestP2 (static inline)
// C: void Of_ObjSetCutBestP2( Of_Man_t * p, int * pCutSet, int iObj, int * pCut )
func (p *OfMan) ofObjSetCutBestP2(pCutSet []int, iObj int, pCut []int) {
	panic("TODO, stubbed")
}

// Of_ManAreaFlow
// C: void Of_ManAreaFlow( Of_Man_t * p )
func (p *OfMan) OfManAreaFlow() {
	panic("TODO, stubbed")
}

// Of_StoCreate
// C: Of_Man_t * Of_StoCreate( Gia_Man_t * pGia, Jf_Par_t * pPars )
func (pGia *GiaManager) OfStoCreate(pPars *JfPar) *OfMan {
	panic("TODO, stubbed")
}

// Of_StoDelete
// C: void Of_StoDelete( Of_Man_t * p )
func (p *OfMan) OfStoDelete() {
	panic("TODO, stubbed")
}

// Of_CutComputeTruth6 (static inline)
// C: int Of_CutComputeTruth6( Of_Man_t * p, Of_Cut_t * pCut0, Of_Cut_t * pCut1, int fCompl0, int fCompl1, Of_Cut_t * pCutR, int fIsXor )
func (p *OfMan) ofCutComputeTruth6(pCut0 *OfCut, pCut1 *OfCut, fCompl0 int, fCompl1 int, pCutR *OfCut, fIsXor int) int {
	panic("TODO, stubbed")
}

// Of_CutComputeTruthMux6 (static inline)
// C: int Of_CutComputeTruthMux6( Of_Man_t * p, Of_Cut_t * pCut0, Of_Cut_t * pCut1, Of_Cut_t * pCutC, int fCompl0, int fCompl1, int fComplC, Of_Cut_t * pCutR )
func (p *OfMan) ofCutComputeTruthMux6(pCut0 *OfCut, pCut1 *OfCut, pCutC *OfCut, fCompl0 int, fCompl1 int, fComplC int, pCutR *OfCut) int {
	panic("TODO, stubbed")
}

// Of_CutCountBits (static inline)
// C: int Of_CutCountBits( word i )
func ofCutCountBits(i uint64) int {
	panic("TODO, stubbed")
}

// Of_CutGetSign (static inline)
// C: word Of_CutGetSign( int * pLeaves, int nLeaves )
func ofCutGetSign(pLeaves []int, nLeaves int) uint64 {
	panic("TODO, stubbed")
}

// Of_CutCreateUnit (static inline)
// C: int Of_CutCreateUnit( Of_Cut_t * p, int i )
func ofCutCreateUnit(p *OfCut, i int) int {
	panic("TODO, stubbed")
}

// Of_Cutprintf (static inline)
// C: void Of_Cutprintf( Of_Man_t * p, Of_Cut_t * pCut )
func (p *OfMan) ofCutprintf(pCut *OfCut) {
	panic("TODO, stubbed")
}

// Of_ManPrepareCuts (static inline)
// C: int Of_ManPrepareCuts( Of_Cut_t * pCuts, Of_Man_t * p, int iObj, int fAddUnit )
func ofManPrepareCuts(pCuts []OfCut, p *OfMan, iObj int, fAddUnit int) int {
	panic("TODO, stubbed")
}

// Of_ManSaveCuts (static inline)
// C: int Of_ManSaveCuts( Of_Man_t * p, Of_Cut_t ** pCuts, int nCuts )
func (p *OfMan) ofManSaveCuts(pCuts []*OfCut, nCuts int) int {
	panic("TODO, stubbed")
}

// Of_ManLiftCuts (static inline)
// C: void Of_ManLiftCuts( Of_Man_t * p, int iObj )
func (p *OfMan) ofManLiftCuts(iObj int) {
	panic("TODO, stubbed")
}

// Of_CutPrint (static inline)
// C: void Of_CutPrint( int * pCut )
func ofCutPrint(pCut []int) {
	panic("TODO, stubbed")
}

// Of_CutCheck (static inline)
// C: int Of_CutCheck( Of_Cut_t * pBase, Of_Cut_t * pCut )
func ofCutCheck(pBase *OfCut, pCut *OfCut) int {
	panic("TODO, stubbed")
}

// Of_SetCheckArray (static inline)
// C: int Of_SetCheckArray( Of_Cut_t ** ppCuts, int nCuts )
func ofSetCheckArray(ppCuts []*OfCut, nCuts int) int {
	panic("TODO, stubbed")
}

// Of_CutMergeOrder (static inline)
// C: int Of_CutMergeOrder( Of_Cut_t * pCut0, Of_Cut_t * pCut1, Of_Cut_t * pCut, int nLutSize )
func ofCutMergeOrder(pCut0 *OfCut, pCut1 *OfCut, pCut *OfCut, nLutSize int) int {
	panic("TODO, stubbed")
}

// Of_CutMergeOrderMux (static inline)
// C: int Of_CutMergeOrderMux( Of_Cut_t * pCut0, Of_Cut_t * pCut1, Of_Cut_t * pCut2, Of_Cut_t * pCut, int nLutSize )
func ofCutMergeOrderMux(pCut0 *OfCut, pCut1 *OfCut, pCut2 *OfCut, pCut *OfCut, nLutSize int) int {
	panic("TODO, stubbed")
}

// Of_SetCutIsContainedOrder (static inline)
// C: int Of_SetCutIsContainedOrder( Of_Cut_t * pBase, Of_Cut_t * pCut )
func ofSetCutIsContainedOrder(pBase *OfCut, pCut *OfCut) int {
	panic("TODO, stubbed")
}

// Of_SetLastCutIsContained (static inline)
// C: int Of_SetLastCutIsContained( Of_Cut_t ** pCuts, int nCuts )
func ofSetLastCutIsContained(pCuts []*OfCut, nCuts int) int {
	panic("TODO, stubbed")
}

// Of_SetLastCutContainsArea (static inline)
// C: int Of_SetLastCutContainsArea( Of_Cut_t ** pCuts, int nCuts )
func ofSetLastCutContainsArea(pCuts []*OfCut, nCuts int) int {
	panic("TODO, stubbed")
}

// Of_CutCompareArea (static inline)
// C: int Of_CutCompareArea( Of_Cut_t * pCut0, Of_Cut_t * pCut1 )
func ofCutCompareArea(pCut0 *OfCut, pCut1 *OfCut) int {
	panic("TODO, stubbed")
}

// Of_SetSortByArea (static inline)
// C: void Of_SetSortByArea( Of_Cut_t ** pCuts, int nCuts )
func ofSetSortByArea(pCuts []*OfCut, nCuts int) {
	panic("TODO, stubbed")
}

// Of_SetAddCut (static inline)
// C: int Of_SetAddCut( Of_Cut_t ** pCuts, int nCuts, int nCutNum )
func ofSetAddCut(pCuts []*OfCut, nCuts int, nCutNum int) int {
	panic("TODO, stubbed")
}

// Of_CutArea (static inline)
// C: int Of_CutArea( Of_Man_t * p, int nLeaves )
func (p *OfMan) ofCutArea(nLeaves int) int {
	panic("TODO, stubbed")
}

// Of_CutParams (static inline)
// C: void Of_CutParams( Of_Man_t * p, Of_Cut_t * pCut, int nGiaRefs )
func (p *OfMan) ofCutParams(pCut *OfCut, nGiaRefs int) {
	panic("TODO, stubbed")
}

// Of_ObjMergeOrder
// C: void Of_ObjMergeOrder( Of_Man_t * p, int iObj )
func (p *OfMan) OfObjMergeOrder(iObj int) {
	panic("TODO, stubbed")
}

// Of_ManComputeCuts
// C: void Of_ManComputeCuts( Of_Man_t * p )
func (p *OfMan) OfManComputeCuts() {
	panic("TODO, stubbed")
}

// Of_ManPrintStats
// C: void Of_ManPrintStats( Of_Man_t * p, char * pTitle )
func (p *OfMan) OfManPrintStats(pTitle string) {
	panic("TODO, stubbed")
}

// Of_ManPrintInit
// C: void Of_ManPrintInit( Of_Man_t * p )
func (p *OfMan) OfManPrintInit() {
	panic("TODO, stubbed")
}

// Of_ManPrintQuit
// C: void Of_ManPrintQuit( Of_Man_t * p )
func (p *OfMan) OfManPrintQuit() {
	panic("TODO, stubbed")
}

// Of_ManComputeForwardCut (static inline)
// C: int Of_ManComputeForwardCut( Of_Man_t * p, int iObj, int * pCut )
func (p *OfMan) ofManComputeForwardCut(iObj int, pCut []int) int {
	panic("TODO, stubbed")
}

// Of_ManComputeForwardCutArea (static inline)
// C: int Of_ManComputeForwardCutArea( Of_Man_t * p, int iObj, int * pCut )
func (p *OfMan) ofManComputeForwardCutArea(iObj int, pCut []int) int {
	panic("TODO, stubbed")
}

// Of_ManComputeForwardObj (static inline)
// C: void Of_ManComputeForwardObj( Of_Man_t * p, int iObj )
func (p *OfMan) ofManComputeForwardObj(iObj int) {
	panic("TODO, stubbed")
}

// Of_ManComputeForward1
// C: void Of_ManComputeForward1( Of_Man_t * p )
func (p *OfMan) OfManComputeForward1() {
	panic("TODO, stubbed")
}

// Of_CutRef_rec
// C: int Of_CutRef_rec( Of_Man_t * p, int * pCut )
func (p *OfMan) OfCutRef_rec(pCut []int) int {
	panic("TODO, stubbed")
}

// Of_CutDeref_rec
// C: int Of_CutDeref_rec( Of_Man_t * p, int * pCut )
func (p *OfMan) OfCutDeref_rec(pCut []int) int {
	panic("TODO, stubbed")
}

// Of_CutAreaDerefed (static inline)
// C: int Of_CutAreaDerefed( Of_Man_t * p, int * pCut )
func (p *OfMan) ofCutAreaDerefed(pCut []int) int {
	panic("TODO, stubbed")
}

// Of_CutRef2_rec
// C: int Of_CutRef2_rec( Of_Man_t * p, int * pCut )
func (p *OfMan) OfCutRef2_rec(pCut []int) int {
	panic("TODO, stubbed")
}

// Of_CutAreaDerefed2 (static inline)
// C: int Of_CutAreaDerefed2( Of_Man_t * p, int * pCut )
func (p *OfMan) ofCutAreaDerefed2(pCut []int) int {
	panic("TODO, stubbed")
}

// Of_ManComputeForwardObj2 (static inline)
// C: void Of_ManComputeForwardObj2( Of_Man_t * p, int iObj )
func (p *OfMan) ofManComputeForwardObj2(iObj int) {
	panic("TODO, stubbed")
}

// Of_ManComputeForward2
// C: void Of_ManComputeForward2( Of_Man_t * p )
func (p *OfMan) OfManComputeForward2() {
	panic("TODO, stubbed")
}

// Of_ManComputeOutputRequired (static inline)
// C: int Of_ManComputeOutputRequired( Of_Man_t * p, int fCleanRefs )
func (p *OfMan) ofManComputeOutputRequired(fCleanRefs int) int {
	panic("TODO, stubbed")
}

// Of_ManComputeBackwardCut (static inline)
// C: int Of_ManComputeBackwardCut( Of_Man_t * p, int * pCut )
func (p *OfMan) ofManComputeBackwardCut(pCut []int) int {
	panic("TODO, stubbed")
}

// Of_ManComputeBackward1
// C: void Of_ManComputeBackward1( Of_Man_t * p )
func (p *OfMan) OfManComputeBackward1() {
	panic("TODO, stubbed")
}

// Of_ManComputeBackward2
// C: void Of_ManComputeBackward2( Of_Man_t * p )
func (p *OfMan) OfManComputeBackward2() {
	panic("TODO, stubbed")
}

// Of_ManComputeBackward3
// C: void Of_ManComputeBackward3( Of_Man_t * p )
func (p *OfMan) OfManComputeBackward3() {
	panic("TODO, stubbed")
}

// Of_ManComputeForwardDirconCut
// C: void Of_ManComputeForwardDirconCut( Of_Man_t * p, int iObj, int * pCut, int * pDelay1, int * pDelay2 )
func (p *OfMan) OfManComputeForwardDirconCut(iObj int, pCut []int, pDelay1 *int, pDelay2 *int) {
	panic("TODO, stubbed")
}

// Of_ManComputeForwardDirconObj
// C: int Of_ManComputeForwardDirconObj( Of_Man_t * p, int iObj )
func (p *OfMan) OfManComputeForwardDirconObj(iObj int) int {
	panic("TODO, stubbed")
}

// Of_ManComputeForwardDircon1
// C: void Of_ManComputeForwardDircon1( Of_Man_t * p )
func (p *OfMan) OfManComputeForwardDircon1() {
	panic("TODO, stubbed")
}

// Of_ManComputeBackwardDircon1
// C: void Of_ManComputeBackwardDircon1( Of_Man_t * p )
func (p *OfMan) OfManComputeBackwardDircon1() {
	panic("TODO, stubbed")
}

// Of_ManCreateSat
// C: void Of_ManCreateSat( Of_Man_t * p, int nCutsAll, Vec_Int_t * vFirst, Vec_Int_t * vCutNum, Vec_Int_t * vBestNode, Vec_Int_t * vBestCut )
func (p *OfMan) OfManCreateSat(nCutsAll int, vFirst *IntVector, vCutNum *IntVector, vBestNode *IntVector, vBestCut *IntVector) {
	panic("TODO, stubbed")
}

// Of_ManPrintCuts
// C: void Of_ManPrintCuts( Of_Man_t * p )
func (p *OfMan) OfManPrintCuts() {
	panic("TODO, stubbed")
}

// Of_ManSetDefaultPars
// C: void Of_ManSetDefaultPars( Jf_Par_t * pPars )
func OfManSetDefaultPars(pPars *JfPar) {
	panic("TODO, stubbed")
}

// Of_ManDeriveMapping
// C: Gia_Man_t * Of_ManDeriveMapping( Of_Man_t * p )
func (p *OfMan) OfManDeriveMapping() *GiaManager {
	panic("TODO, stubbed")
}

// Of_ManPerformMapping
// C: Gia_Man_t * Of_ManPerformMapping( Gia_Man_t * pGia, Jf_Par_t * pPars )
func (pGia *GiaManager) OfManPerformMapping(pPars *JfPar) *GiaManager {
	panic("TODO, stubbed")
}
