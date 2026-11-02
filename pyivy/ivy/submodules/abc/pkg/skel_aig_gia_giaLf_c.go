package abc

// Hand-checked skeleton of src/aig/gia/giaLf.c.

const (
	LF_LEAF_MAX  = 13
	LF_CUT_MAX   = 32
	LF_LOG_PAGE  = 12
	LF_NO_LEAF   = 255
	LF_CUT_WORDS = 4 + LF_LEAF_MAX/2
	LF_TT_WORDS  = 1 << (LF_LEAF_MAX - 6) // ((LF_LEAF_MAX > 6) ? 1 << (LF_LEAF_MAX-6) : 1)
	LF_EPSILON   = 0.005
)

// LfCut is Lf_Cut_t.
type LfCut struct {
	Sign    uint64  // word : signature
	Delay   int     // delay
	Flow    float32 // flow
	iFunc   int     // functionality
	Cost    uint32  // unsigned : 22 : misc cut cost
	fLate   uint32  // unsigned : 1 : fails timing
	fMux7   uint32  // unsigned : 1 : specialized cut
	nLeaves uint32  // unsigned : 8 : the number of leaves
	pLeaves []int   // int pLeaves[0] : leaves (flexible array member)
}

// LfPlc is Lf_Plc_t.
type LfPlc struct {
	fUsed  uint32 // unsigned : 1 : the cut is used
	Handle uint32 // unsigned : 31 : the cut handle
}

// LfBst is Lf_Bst_t.
type LfBst struct {
	Delay [3]int     // delay
	Flow  [3]float32 // flow
	Cut   [2]LfPlc   // cut info
}

// LfMem is Lf_Mem_t.
type LfMem struct {
	LogPage   int                  // log size of memory page
	MaskPage  int                  // page mask
	nCutWords int                  // cut size in words
	iCur      int                  // writing position
	vPages    PointerVector[byte]  // Vec_Ptr_t (by value) : memory pages
	vFree     *PointerVector[byte] // free pages
}

// LfMan is Lf_Man_t.
type LfMan struct {
	pGia        *GiaManager          // manager
	pPars       *JfPar               // parameters
	nCutWords   int                  // cut size in words
	nSetWords   int                  // set size in words
	pObjBests   []LfBst              // Lf_Bst_t * : best cuts
	vMemSets    PointerVector[LfCut] // Vec_Ptr_t (by value) : memory for cutsets
	vFreeSets   IntVector            // Vec_Int_t (by value) : free cutsets
	vTtMem      *WordMemory          // truth tables
	vFreePages  PointerVector[byte]  // Vec_Ptr_t (by value) : free memory pages
	vStoreOld   LfMem                // previous cuts
	vStoreNew   LfMem                // current cuts
	vOffsets    IntVector            // Vec_Int_t (by value) : offsets
	vRequired   IntVector            // Vec_Int_t (by value) : required times
	vCutSets    IntVector            // Vec_Int_t (by value) : cutsets (pObj->Value stores cut refs)
	vFlowRefs   FloatVector          // Vec_Flt_t (by value) : flow refs
	vMapRefs    IntVector            // Vec_Int_t (by value) : mapping refs
	vSwitches   FloatVector          // Vec_Flt_t (by value) : switching activity
	vCiArrivals IntVector            // Vec_Int_t (by value) : arrival times of the CIs
	clkStart    int64                // abctime : starting time
	CutCount    [4]float64           // double : cut counts
	Switches    float64              // switching activity
	nFrontMax   int                  // frontier
	nCoDrivers  int                  // CO drivers
	nInverters  int                  // inverters
	nTimeFails  int                  // timing fails
	Iter        int                  // mapping iteration
	fUseEla     int                  // use exact local area
	nCutMux     int                  // non-trivial MUX cuts
	nCutEqual   int                  // equal two cuts
	nCutCounts  [LF_LEAF_MAX + 1]int
}

// Lf_CutCopy (static inline)
// C: void Lf_CutCopy( Lf_Cut_t * p, Lf_Cut_t * q, int n )
func lfCutCopy(p *LfCut, q *LfCut, n int) {
	panic("TODO, stubbed")
}

// Lf_CutNext (static inline)
// C: Lf_Cut_t * Lf_CutNext( Lf_Cut_t * p, int n )
func lfCutNext(p *LfCut, n int) *LfCut {
	panic("TODO, stubbed")
}

// Lf_CutTruth (static inline)
// C: word * Lf_CutTruth( Lf_Man_t * p, Lf_Cut_t * pCut )
func (p *LfMan) lfCutTruth(pCut *LfCut) []uint64 {
	panic("TODO, stubbed")
}

// Lf_ObjOff (static inline)
// C: int Lf_ObjOff( Lf_Man_t * p, int i )
func (p *LfMan) lfObjOff(i int) int {
	panic("TODO, stubbed")
}

// Lf_ObjRequired (static inline)
// C: int Lf_ObjRequired( Lf_Man_t * p, int i )
func (p *LfMan) lfObjRequired(i int) int {
	panic("TODO, stubbed")
}

// Lf_ObjSetRequired (static inline)
// C: void Lf_ObjSetRequired( Lf_Man_t * p, int i, int t )
func (p *LfMan) lfObjSetRequired(i int, t int) {
	panic("TODO, stubbed")
}

// Lf_ObjReadBest (static inline)
// C: Lf_Bst_t * Lf_ObjReadBest( Lf_Man_t * p, int i )
func (p *LfMan) lfObjReadBest(i int) *LfBst {
	panic("TODO, stubbed")
}

// Lf_ObjFlowRefs (static inline)
// C: float Lf_ObjFlowRefs( Lf_Man_t * p, int i )
func (p *LfMan) lfObjFlowRefs(i int) float32 {
	panic("TODO, stubbed")
}

// Lf_ObjMapRefNum (static inline)
// C: int Lf_ObjMapRefNum( Lf_Man_t * p, int i )
func (p *LfMan) lfObjMapRefNum(i int) int {
	panic("TODO, stubbed")
}

// Lf_ObjMapRefInc (static inline)
// C: int Lf_ObjMapRefInc( Lf_Man_t * p, int i )
func (p *LfMan) lfObjMapRefInc(i int) int {
	panic("TODO, stubbed")
}

// Lf_ObjMapRefDec (static inline)
// C: int Lf_ObjMapRefDec( Lf_Man_t * p, int i )
func (p *LfMan) lfObjMapRefDec(i int) int {
	panic("TODO, stubbed")
}

// Lf_ObjSwitches (static inline)
// C: float Lf_ObjSwitches( Lf_Man_t * p, int i )
func (p *LfMan) lfObjSwitches(i int) float32 {
	panic("TODO, stubbed")
}

// Lf_BestDiffCuts (static inline)
// C: int Lf_BestDiffCuts( Lf_Bst_t * p )
func lfBestDiffCuts(p *LfBst) int {
	panic("TODO, stubbed")
}

// Lf_BestIsMapped (static inline)
// C: int Lf_BestIsMapped( Lf_Bst_t * p )
func lfBestIsMapped(p *LfBst) int {
	panic("TODO, stubbed")
}

// Lf_BestIndex (static inline)
// C: int Lf_BestIndex( Lf_Bst_t * p )
func lfBestIndex(p *LfBst) int {
	panic("TODO, stubbed")
}

// Lf_BestCutIndex (static inline)
// C: int Lf_BestCutIndex( Lf_Bst_t * p )
func lfBestCutIndex(p *LfBst) int {
	panic("TODO, stubbed")
}

// Lf_ObjSetCiArrival (static inline)
// C: void Lf_ObjSetCiArrival( Lf_Man_t * p, int iCi, int Time )
func (p *LfMan) lfObjSetCiArrival(iCi int, Time int) {
	panic("TODO, stubbed")
}

// Lf_ObjCiArrival (static inline)
// C: int Lf_ObjCiArrival( Lf_Man_t * p, int iCi )
func (p *LfMan) lfObjCiArrival(iCi int) int {
	panic("TODO, stubbed")
}

// Lf_ObjArrival_rec
// C: int Lf_ObjArrival_rec( Lf_Man_t * p, Gia_Obj_t * pDriver )
func (p *LfMan) LfObjArrival_rec(pDriver *GiaObject) int {
	panic("TODO, stubbed")
}

// Lf_ObjCoArrival (static inline)
// C: int Lf_ObjCoArrival( Lf_Man_t * p, int iCo )
func (p *LfMan) lfObjCoArrival(iCo int) int {
	panic("TODO, stubbed")
}

// Lf_ObjCoArrival2_rec
// C: int Lf_ObjCoArrival2_rec( Lf_Man_t * p, Gia_Obj_t * pDriver )
func (p *LfMan) LfObjCoArrival2_rec(pDriver *GiaObject) int {
	panic("TODO, stubbed")
}

// Lf_ObjCoArrival2 (static inline)
// C: int Lf_ObjCoArrival2( Lf_Man_t * p, int iCo )
func (p *LfMan) lfObjCoArrival2(iCo int) int {
	panic("TODO, stubbed")
}

// Lf_ManComputeCrossCut
// C: int Lf_ManComputeCrossCut( Gia_Man_t * p )
func (p *GiaManager) LfManComputeCrossCut() int {
	panic("TODO, stubbed")
}

// Lf_ManTtIsMux
// C: int Lf_ManTtIsMux( word t )
func LfManTtIsMux(t uint64) int {
	panic("TODO, stubbed")
}

// Lf_ManAnalyzeCoDrivers
// C: void Lf_ManAnalyzeCoDrivers( Gia_Man_t * p, int * pnDrivers, int * pnInverts )
func (p *GiaManager) LfManAnalyzeCoDrivers(pnDrivers *int, pnInverts *int) {
	panic("TODO, stubbed")
}

// Lf_ManComputeSwitching
// C: void Lf_ManComputeSwitching( Gia_Man_t * p, Vec_Flt_t * vSwitches )
func (p *GiaManager) LfManComputeSwitching(vSwitches *FloatVector) {
	panic("TODO, stubbed")
}

// Lf_CutCreateUnit (static inline)
// C: int Lf_CutCreateUnit( Lf_Cut_t * p, int i )
func lfCutCreateUnit(p *LfCut, i int) int {
	panic("TODO, stubbed")
}

// Lf_ManFetchSet (static inline)
// C: Lf_Cut_t * Lf_ManFetchSet( Lf_Man_t * p, int i )
func (p *LfMan) lfManFetchSet(i int) *LfCut {
	panic("TODO, stubbed")
}

// Lf_ManPrepareSet (static inline)
// C: int Lf_ManPrepareSet( Lf_Man_t * p, int iObj, int Index, Lf_Cut_t ** ppCutSet )
func (p *LfMan) lfManPrepareSet(iObj int, Index int, ppCutSet **LfCut) int {
	panic("TODO, stubbed")
}

// Lf_CutGetSign (static inline)
// C: word Lf_CutGetSign( Lf_Cut_t * pCut )
func lfCutGetSign(pCut *LfCut) uint64 {
	panic("TODO, stubbed")
}

// Lf_CutCountBits (static inline)
// C: int Lf_CutCountBits( word i )
func lfCutCountBits(i uint64) int {
	panic("TODO, stubbed")
}

// Lf_CutEqual (static inline)
// C: int Lf_CutEqual( Lf_Cut_t * pCut0, Lf_Cut_t * pCut1 )
func lfCutEqual(pCut0 *LfCut, pCut1 *LfCut) int {
	panic("TODO, stubbed")
}

// Lf_CutSwitches (static inline)
// C: float Lf_CutSwitches( Lf_Man_t * p, Lf_Cut_t * pCut )
func (p *LfMan) lfCutSwitches(pCut *LfCut) float32 {
	panic("TODO, stubbed")
}

// Lf_CutPrint (static inline)
// C: void Lf_CutPrint( Lf_Man_t * p, Lf_Cut_t * pCut )
func (p *LfMan) lfCutPrint(pCut *LfCut) {
	panic("TODO, stubbed")
}

// Lf_CutArea (static inline)
// C: float Lf_CutArea( Lf_Man_t * p, Lf_Cut_t * pCut )
func (p *LfMan) lfCutArea(pCut *LfCut) float32 {
	panic("TODO, stubbed")
}

// Lf_CutIsMux (static inline)
// C: int Lf_CutIsMux( Lf_Man_t * p, Lf_Cut_t * pCut, Gia_Obj_t * pMux )
func (p *LfMan) lfCutIsMux(pCut *LfCut, pMux *GiaObject) int {
	panic("TODO, stubbed")
}

// Lf_MemAlloc (static inline)
// C: void Lf_MemAlloc( Lf_Mem_t * p, int LogPage, Vec_Ptr_t * vFree, int nCutWords )
func (p *LfMem) lfMemAlloc(LogPage int, vFree *PointerVector[byte], nCutWords int) {
	panic("TODO, stubbed")
}

// Lf_MemSaveCut (static inline)
// C: int Lf_MemSaveCut( Lf_Mem_t * p, Lf_Cut_t * pCut, int iObj )
func (p *LfMem) lfMemSaveCut(pCut *LfCut, iObj int) int {
	panic("TODO, stubbed")
}

// Lf_MemLoadCut (static inline)
// C: Lf_Cut_t * Lf_MemLoadCut( Lf_Mem_t * p, int iCur, int iObj, Lf_Cut_t * pCut, int fTruth, int fRecycle )
func (p *LfMem) lfMemLoadCut(iCur int, iObj int, pCut *LfCut, fTruth int, fRecycle int) *LfCut {
	panic("TODO, stubbed")
}

// Lf_MemRecycle (static inline)
// C: void Lf_MemRecycle( Lf_Mem_t * p )
func (p *LfMem) lfMemRecycle() {
	panic("TODO, stubbed")
}

// Lf_MemLoadMuxCut (static inline)
// C: Lf_Cut_t * Lf_MemLoadMuxCut( Lf_Man_t * p, int iObj, Lf_Cut_t * pCut )
func (p *LfMan) lfMemLoadMuxCut(iObj int, pCut *LfCut) *LfCut {
	panic("TODO, stubbed")
}

// Lf_ObjCutMux (static inline)
// C: Lf_Cut_t * Lf_ObjCutMux( Lf_Man_t * p, int i )
func (p *LfMan) lfObjCutMux(i int) *LfCut {
	panic("TODO, stubbed")
}

// Lf_ObjCutBest (static inline)
// C: Lf_Cut_t * Lf_ObjCutBest( Lf_Man_t * p, int i )
func (p *LfMan) lfObjCutBest(i int) *LfCut {
	panic("TODO, stubbed")
}

// Lf_ObjCutBestNew (static inline)
// C: Lf_Cut_t * Lf_ObjCutBestNew( Lf_Man_t * p, int i, Lf_Cut_t * pCut )
func (p *LfMan) lfObjCutBestNew(i int, pCut *LfCut) *LfCut {
	panic("TODO, stubbed")
}

// Lf_CutCheck (static inline)
// C: int Lf_CutCheck( Lf_Cut_t * pBase, Lf_Cut_t * pCut )
func lfCutCheck(pBase *LfCut, pCut *LfCut) int {
	panic("TODO, stubbed")
}

// Lf_SetCheckArray (static inline)
// C: int Lf_SetCheckArray( Lf_Cut_t ** ppCuts, int nCuts )
func lfSetCheckArray(ppCuts []*LfCut, nCuts int) int {
	panic("TODO, stubbed")
}

// Lf_CutMergeOrder (static inline)
// C: int Lf_CutMergeOrder( Lf_Cut_t * pCut0, Lf_Cut_t * pCut1, Lf_Cut_t * pCut, int nLutSize )
func lfCutMergeOrder(pCut0 *LfCut, pCut1 *LfCut, pCut *LfCut, nLutSize int) int {
	panic("TODO, stubbed")
}

// Lf_CutMergeOrder2 (static inline)
// C: int Lf_CutMergeOrder2( Lf_Cut_t * pCut0, Lf_Cut_t * pCut1, Lf_Cut_t * pCut, int nLutSize )
func lfCutMergeOrder2(pCut0 *LfCut, pCut1 *LfCut, pCut *LfCut, nLutSize int) int {
	panic("TODO, stubbed")
}

// Lf_CutMergeOrderMux (static inline)
// C: int Lf_CutMergeOrderMux( Lf_Cut_t * pCut0, Lf_Cut_t * pCut1, Lf_Cut_t * pCut2, Lf_Cut_t * pCut, int nLutSize )
func lfCutMergeOrderMux(pCut0 *LfCut, pCut1 *LfCut, pCut2 *LfCut, pCut *LfCut, nLutSize int) int {
	panic("TODO, stubbed")
}

// Lf_SetCutIsContainedOrder (static inline)
// C: int Lf_SetCutIsContainedOrder( Lf_Cut_t * pBase, Lf_Cut_t * pCut )
func lfSetCutIsContainedOrder(pBase *LfCut, pCut *LfCut) int {
	panic("TODO, stubbed")
}

// Lf_SetLastCutIsContained (static inline)
// C: int Lf_SetLastCutIsContained( Lf_Cut_t ** pCuts, int nCuts )
func lfSetLastCutIsContained(pCuts []*LfCut, nCuts int) int {
	panic("TODO, stubbed")
}

// Lf_CutCompareDelay (static inline)
// C: int Lf_CutCompareDelay( Lf_Cut_t * pCut0, Lf_Cut_t * pCut1 )
func lfCutCompareDelay(pCut0 *LfCut, pCut1 *LfCut) int {
	panic("TODO, stubbed")
}

// Lf_CutCompareArea (static inline)
// C: int Lf_CutCompareArea( Lf_Cut_t * pCut0, Lf_Cut_t * pCut1 )
func lfCutCompareArea(pCut0 *LfCut, pCut1 *LfCut) int {
	panic("TODO, stubbed")
}

// Lf_SetLastCutContainsArea (static inline)
// C: int Lf_SetLastCutContainsArea( Lf_Cut_t ** pCuts, int nCuts )
func lfSetLastCutContainsArea(pCuts []*LfCut, nCuts int) int {
	panic("TODO, stubbed")
}

// Lf_SetSortByArea (static inline)
// C: void Lf_SetSortByArea( Lf_Cut_t ** pCuts, int nCuts )
func lfSetSortByArea(pCuts []*LfCut, nCuts int) {
	panic("TODO, stubbed")
}

// Lf_SetAddCut (static inline)
// C: int Lf_SetAddCut( Lf_Cut_t ** pCuts, int nCuts, int nCutNum )
func lfSetAddCut(pCuts []*LfCut, nCuts int, nCutNum int) int {
	panic("TODO, stubbed")
}

// Lf_SetSortBySize (static inline)
// C: void Lf_SetSortBySize( Lf_Cut_t ** pCutsR, int nCutsR )
func lfSetSortBySize(pCutsR []*LfCut, nCutsR int) {
	panic("TODO, stubbed")
}

// Lf_ManFindCofVar (static inline)
// C: int Lf_ManFindCofVar( word * pTruth, int nWords, int nVars )
func lfManFindCofVar(pTruth []uint64, nWords int, nVars int) int {
	panic("TODO, stubbed")
}

// Lf_CutComputeTruth6 (static inline)
// C: int Lf_CutComputeTruth6( Lf_Man_t * p, Lf_Cut_t * pCut0, Lf_Cut_t * pCut1, int fCompl0, int fCompl1, Lf_Cut_t * pCutR, int fIsXor )
func (p *LfMan) lfCutComputeTruth6(pCut0 *LfCut, pCut1 *LfCut, fCompl0 int, fCompl1 int, pCutR *LfCut, fIsXor int) int {
	panic("TODO, stubbed")
}

// Lf_CutComputeTruth (static inline)
// C: int Lf_CutComputeTruth( Lf_Man_t * p, Lf_Cut_t * pCut0, Lf_Cut_t * pCut1, int fCompl0, int fCompl1, Lf_Cut_t * pCutR, int fIsXor )
func (p *LfMan) lfCutComputeTruth(pCut0 *LfCut, pCut1 *LfCut, fCompl0 int, fCompl1 int, pCutR *LfCut, fIsXor int) int {
	panic("TODO, stubbed")
}

// Lf_CutComputeTruthMux6 (static inline)
// C: int Lf_CutComputeTruthMux6( Lf_Man_t * p, Lf_Cut_t * pCut0, Lf_Cut_t * pCut1, Lf_Cut_t * pCutC, int fCompl0, int fCompl1, int fComplC, Lf_Cut_t * pCutR )
func (p *LfMan) lfCutComputeTruthMux6(pCut0 *LfCut, pCut1 *LfCut, pCutC *LfCut, fCompl0 int, fCompl1 int, fComplC int, pCutR *LfCut) int {
	panic("TODO, stubbed")
}

// Lf_CutComputeTruthMux (static inline)
// C: int Lf_CutComputeTruthMux( Lf_Man_t * p, Lf_Cut_t * pCut0, Lf_Cut_t * pCut1, Lf_Cut_t * pCutC, int fCompl0, int fCompl1, int fComplC, Lf_Cut_t * pCutR )
func (p *LfMan) lfCutComputeTruthMux(pCut0 *LfCut, pCut1 *LfCut, pCutC *LfCut, fCompl0 int, fCompl1 int, fComplC int, pCutR *LfCut) int {
	panic("TODO, stubbed")
}

// Lf_CutRef_rec
// C: float Lf_CutRef_rec( Lf_Man_t * p, Lf_Cut_t * pCut )
func (p *LfMan) LfCutRef_rec(pCut *LfCut) float32 {
	panic("TODO, stubbed")
}

// Lf_CutDeref_rec
// C: float Lf_CutDeref_rec( Lf_Man_t * p, Lf_Cut_t * pCut )
func (p *LfMan) LfCutDeref_rec(pCut *LfCut) float32 {
	panic("TODO, stubbed")
}

// Lf_CutAreaDerefed (static inline)
// C: float Lf_CutAreaDerefed( Lf_Man_t * p, Lf_Cut_t * pCut )
func (p *LfMan) lfCutAreaDerefed(pCut *LfCut) float32 {
	panic("TODO, stubbed")
}

// Lf_CutRequired (static inline)
// C: int Lf_CutRequired( Lf_Man_t * p, Lf_Cut_t * pCut )
func (p *LfMan) lfCutRequired(pCut *LfCut) int {
	panic("TODO, stubbed")
}

// Lf_CutParams (static inline)
// C: void Lf_CutParams( Lf_Man_t * p, Lf_Cut_t * pCut, int Required, float FlowRefs, Gia_Obj_t * pMux )
func (p *LfMan) lfCutParams(pCut *LfCut, Required int, FlowRefs float32, pMux *GiaObject) {
	panic("TODO, stubbed")
}

// Lf_ObjMergeOrder
// C: void Lf_ObjMergeOrder( Lf_Man_t * p, int iObj )
func (p *LfMan) LfObjMergeOrder(iObj int) {
	panic("TODO, stubbed")
}

// Lf_ManSetFlowRefInc (static inline)
// C: void Lf_ManSetFlowRefInc( Gia_Man_t * p, Vec_Flt_t * vRefs, Vec_Int_t * vOffsets, int i )
func (p *GiaManager) lfManSetFlowRefInc(vRefs *FloatVector, vOffsets *IntVector, i int) {
	panic("TODO, stubbed")
}

// Lf_ManSetFlowRefs
// C: void Lf_ManSetFlowRefs( Gia_Man_t * p, Vec_Flt_t * vRefs, Vec_Int_t * vOffsets )
func (p *GiaManager) LfManSetFlowRefs(vRefs *FloatVector, vOffsets *IntVector) {
	panic("TODO, stubbed")
}

// Lf_ManSetCutRefs
// C: void Lf_ManSetCutRefs( Lf_Man_t * p )
func (p *LfMan) LfManSetCutRefs() {
	panic("TODO, stubbed")
}

// Lf_ManSetMuxCut (static inline)
// C: int Lf_ManSetMuxCut( Lf_Man_t * p, Lf_Bst_t * pBest, int iObj, int Required )
func (p *LfMan) lfManSetMuxCut(pBest *LfBst, iObj int, Required int) int {
	panic("TODO, stubbed")
}

// Lf_ManSetMapRefsOne
// C: void Lf_ManSetMapRefsOne( Lf_Man_t * p, int iObj )
func (p *LfMan) LfManSetMapRefsOne(iObj int) {
	panic("TODO, stubbed")
}

// Lf_ManSetMapRefs
// C: int Lf_ManSetMapRefs( Lf_Man_t * p )
func (p *LfMan) LfManSetMapRefs() int {
	panic("TODO, stubbed")
}

// Lf_ManCountMapRefsOne
// C: void Lf_ManCountMapRefsOne( Lf_Man_t * p, int iObj )
func (p *LfMan) LfManCountMapRefsOne(iObj int) {
	panic("TODO, stubbed")
}

// Lf_ManCountMapRefs
// C: void Lf_ManCountMapRefs( Lf_Man_t * p )
func (p *LfMan) LfManCountMapRefs() {
	panic("TODO, stubbed")
}

// Lf_ManDeriveMapping
// C: Gia_Man_t * Lf_ManDeriveMapping( Lf_Man_t * p )
func (p *LfMan) LfManDeriveMapping() *GiaManager {
	panic("TODO, stubbed")
}

// Lf_ManDeriveMappingCoarse
// C: Gia_Man_t * Lf_ManDeriveMappingCoarse( Lf_Man_t * p )
func (p *LfMan) LfManDeriveMappingCoarse() *GiaManager {
	panic("TODO, stubbed")
}

// Lf_ManDerivePart (static inline)
// C: int Lf_ManDerivePart( Lf_Man_t * p, Gia_Man_t * pNew, Vec_Int_t * vMapping, Vec_Int_t * vMapping2, Vec_Int_t * vCopies, Lf_Cut_t * pCut, Vec_Int_t * vLeaves, Vec_Int_t * vCover, Gia_Obj_t * pObj )
func (p *LfMan) lfManDerivePart(pNew *GiaManager, vMapping *IntVector, vMapping2 *IntVector, vCopies *IntVector, pCut *LfCut, vLeaves *IntVector, vCover *IntVector, pObj *GiaObject) int {
	panic("TODO, stubbed")
}

// Lf_ManDeriveMappingGia
// C: Gia_Man_t * Lf_ManDeriveMappingGia( Lf_Man_t * p )
func (p *LfMan) LfManDeriveMappingGia() *GiaManager {
	panic("TODO, stubbed")
}

// Lf_ManAlloc
// C: Lf_Man_t * Lf_ManAlloc( Gia_Man_t * pGia, Jf_Par_t * pPars )
func (pGia *GiaManager) LfManAlloc(pPars *JfPar) *LfMan {
	panic("TODO, stubbed")
}

// Lf_ManFree
// C: void Lf_ManFree( Lf_Man_t * p )
func (p *LfMan) LfManFree() {
	panic("TODO, stubbed")
}

// Lf_ManSetDefaultPars
// C: void Lf_ManSetDefaultPars( Jf_Par_t * pPars )
func LfManSetDefaultPars(pPars *JfPar) {
	panic("TODO, stubbed")
}

// Lf_ManPrintStats
// C: void Lf_ManPrintStats( Lf_Man_t * p, char * pTitle )
func (p *LfMan) LfManPrintStats(pTitle string) {
	panic("TODO, stubbed")
}

// Lf_ManPrintInit
// C: void Lf_ManPrintInit( Lf_Man_t * p )
func (p *LfMan) LfManPrintInit() {
	panic("TODO, stubbed")
}

// Lf_ManPrintQuit
// C: void Lf_ManPrintQuit( Lf_Man_t * p, Gia_Man_t * pNew )
func (p *LfMan) LfManPrintQuit(pNew *GiaManager) {
	panic("TODO, stubbed")
}

// Lf_ManComputeMapping
// C: void Lf_ManComputeMapping( Lf_Man_t * p )
func (p *LfMan) LfManComputeMapping() {
	panic("TODO, stubbed")
}

// Lf_ManPerformMappingInt
// C: Gia_Man_t * Lf_ManPerformMappingInt( Gia_Man_t * pGia, Jf_Par_t * pPars )
func (pGia *GiaManager) LfManPerformMappingInt(pPars *JfPar) *GiaManager {
	panic("TODO, stubbed")
}

// Lf_ManPerformMapping
// C: Gia_Man_t * Lf_ManPerformMapping( Gia_Man_t * p, Jf_Par_t * pPars )
func (p *GiaManager) LfManPerformMapping(pPars *JfPar) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManPerformLfMapping
// C: Gia_Man_t * Gia_ManPerformLfMapping( Gia_Man_t * p, Jf_Par_t * pPars, int fNormalized )
func (p *GiaManager) GiaManPerformLfMapping(pPars *JfPar, fNormalized int) *GiaManager {
	panic("TODO, stubbed")
}
