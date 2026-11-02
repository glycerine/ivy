package abc

// Hand-checked skeleton of src/aig/gia/giaMf.c.

const (
	MF_LEAF_MAX = 10
	MF_CUT_MAX  = 16
	MF_LOG_PAGE = 12
	MF_NO_LEAF  = 31
	MF_TT_WORDS = 1 << (MF_LEAF_MAX - 6) // ((MF_LEAF_MAX > 6) ? 1 << (MF_LEAF_MAX-6) : 1)
	MF_NO_FUNC  = 134217727              // (1<<27)-1
	MF_EPSILON  = 0.005
)

// MfCut is Mf_Cut_t.
type MfCut struct {
	Sign    uint64               // word : signature
	Delay   int                  // delay
	Flow    float32              // flow
	iFunc   uint32               // unsigned : 27 : function (MF_NO_FUNC)
	nLeaves uint32               // unsigned : 5 : leaf number (MF_NO_LEAF)
	pLeaves [MF_LEAF_MAX + 1]int // leaves
}

// MfObj is Mf_Obj_t.
type MfObj struct {
	iCutSet   int     // cutset
	Flow      float32 // area
	nFlowRefs float32 // flow references
	Delay     uint32  // unsigned : 16 : delay
	nMapRefs  uint32  // unsigned : 16 : map references
}

// MfMan is Mf_Man_t.
type MfMan struct {
	pGia0      *GiaManager        // original manager
	pGia       *GiaManager        // derived manager
	pPars      *JfPar             // parameters
	pLfObjs    []MfObj            // Mf_Obj_t * : best cuts
	vPages     PointerVector[int] // Vec_Ptr_t (by value) : cut memory
	vTtMem     *WordMemory        // truth tables
	vCnfSizes  IntVector          // Vec_Int_t (by value) : handles to CNF
	vCnfMem    IntVector          // Vec_Int_t (by value) : memory for CNF
	vTemp      IntVector          // Vec_Int_t (by value) : temporary array
	iCur       int                // current position
	Iter       int                // mapping iterations
	fUseEla    int                // use exact area
	clkStart   int64              // abctime : starting time
	CutCount   [4]float64         // double : cut counts
	nCutCounts [MF_LEAF_MAX + 1]int
}

// Mf_ManObj (static inline)
// C: Mf_Obj_t * Mf_ManObj( Mf_Man_t * p, int i )
func (p *MfMan) mfManObj(i int) *MfObj {
	panic("TODO, stubbed")
}

// Mf_ManCutSet (static inline)
// C: int * Mf_ManCutSet( Mf_Man_t * p, int i )
func (p *MfMan) mfManCutSet(i int) []int {
	panic("TODO, stubbed")
}

// Mf_ObjCutSet (static inline)
// C: int * Mf_ObjCutSet( Mf_Man_t * p, int i )
func (p *MfMan) mfObjCutSet(i int) []int {
	panic("TODO, stubbed")
}

// Mf_ObjCutBest (static inline)
// C: int * Mf_ObjCutBest( Mf_Man_t * p, int i )
func (p *MfMan) mfObjCutBest(i int) []int {
	panic("TODO, stubbed")
}

// Mf_ObjMapRefNum (static inline)
// C: int Mf_ObjMapRefNum( Mf_Man_t * p, int i )
func (p *MfMan) mfObjMapRefNum(i int) int {
	panic("TODO, stubbed")
}

// Mf_ObjMapRefInc (static inline)
// C: int Mf_ObjMapRefInc( Mf_Man_t * p, int i )
func (p *MfMan) mfObjMapRefInc(i int) int {
	panic("TODO, stubbed")
}

// Mf_ObjMapRefDec (static inline)
// C: int Mf_ObjMapRefDec( Mf_Man_t * p, int i )
func (p *MfMan) mfObjMapRefDec(i int) int {
	panic("TODO, stubbed")
}

// Mf_CutSize (static inline)
// C: int Mf_CutSize( int * pCut )
func mfCutSize(pCut []int) int {
	panic("TODO, stubbed")
}

// Mf_CutFunc (static inline)
// C: int Mf_CutFunc( int * pCut )
func mfCutFunc(pCut []int) int {
	panic("TODO, stubbed")
}

// Mf_CutSetBoth (static inline)
// C: int Mf_CutSetBoth( int n, int f )
func mfCutSetBoth(n int, f int) int {
	panic("TODO, stubbed")
}

// Mf_CutIsTriv (static inline)
// C: int Mf_CutIsTriv( int * pCut, int i )
func mfCutIsTriv(pCut []int, i int) int {
	panic("TODO, stubbed")
}

// Mf_ManTruthCanonicize
// C: int Mf_ManTruthCanonicize( word * t, int nVars )
func MfManTruthCanonicize(t []uint64, nVars int) int {
	panic("TODO, stubbed")
}

// Mf_ManTruthQuit
// C: void Mf_ManTruthQuit()
func MfManTruthQuit() {
	panic("TODO, stubbed")
}

// Mf_ManTruthCollect
// C: Vec_Wrd_t * Mf_ManTruthCollect( int Limit )
func MfManTruthCollect(Limit int) *WordVector {
	panic("TODO, stubbed")
}

// Mf_ManTruthCount
// C: int Mf_ManTruthCount()
func MfManTruthCount() int {
	panic("TODO, stubbed")
}

// Mf_ManProfileTruths
// C: void Mf_ManProfileTruths( Mf_Man_t * p )
func (p *MfMan) MfManProfileTruths() {
	panic("TODO, stubbed")
}

// Mf_CutPrintOne (static inline)
// C: void Mf_CutPrintOne( int * pCut )
func mfCutPrintOne(pCut []int) {
	panic("TODO, stubbed")
}

// Mf_CubeLit (static inline)
// C: int Mf_CubeLit( int Cube, int iVar )
func mfCubeLit(Cube int, iVar int) int {
	panic("TODO, stubbed")
}

// Mf_ManCountLits (static inline)
// C: int Mf_ManCountLits( int * pCnf, int nCubes, int nVars )
func mfManCountLits(pCnf []int, nCubes int, nVars int) int {
	panic("TODO, stubbed")
}

// Mf_ManDeriveCnfs
// C: Vec_Int_t * Mf_ManDeriveCnfs( Mf_Man_t * p, int * pnVars, int * pnClas, int * pnLits )
func (p *MfMan) MfManDeriveCnfs(pnVars *int, pnClas *int, pnLits *int) *IntVector {
	panic("TODO, stubbed")
}

// Mf_ManDeriveCnf
// C: Cnf_Dat_t * Mf_ManDeriveCnf( Mf_Man_t * p, int fCnfObjIds, int fAddOrCla )
func (p *MfMan) MfManDeriveCnf(fCnfObjIds int, fAddOrCla int) *CnfDat {
	panic("TODO, stubbed")
}

// Mf_CutComputeTruth6 (static inline)
// C: int Mf_CutComputeTruth6( Mf_Man_t * p, Mf_Cut_t * pCut0, Mf_Cut_t * pCut1, int fCompl0, int fCompl1, Mf_Cut_t * pCutR, int fIsXor )
func (p *MfMan) mfCutComputeTruth6(pCut0 *MfCut, pCut1 *MfCut, fCompl0 int, fCompl1 int, pCutR *MfCut, fIsXor int) int {
	panic("TODO, stubbed")
}

// Mf_CutComputeTruth (static inline)
// C: int Mf_CutComputeTruth( Mf_Man_t * p, Mf_Cut_t * pCut0, Mf_Cut_t * pCut1, int fCompl0, int fCompl1, Mf_Cut_t * pCutR, int fIsXor )
func (p *MfMan) mfCutComputeTruth(pCut0 *MfCut, pCut1 *MfCut, fCompl0 int, fCompl1 int, pCutR *MfCut, fIsXor int) int {
	panic("TODO, stubbed")
}

// Mf_CutComputeTruthMux6 (static inline)
// C: int Mf_CutComputeTruthMux6( Mf_Man_t * p, Mf_Cut_t * pCut0, Mf_Cut_t * pCut1, Mf_Cut_t * pCutC, int fCompl0, int fCompl1, int fComplC, Mf_Cut_t * pCutR )
func (p *MfMan) mfCutComputeTruthMux6(pCut0 *MfCut, pCut1 *MfCut, pCutC *MfCut, fCompl0 int, fCompl1 int, fComplC int, pCutR *MfCut) int {
	panic("TODO, stubbed")
}

// Mf_CutComputeTruthMux (static inline)
// C: int Mf_CutComputeTruthMux( Mf_Man_t * p, Mf_Cut_t * pCut0, Mf_Cut_t * pCut1, Mf_Cut_t * pCutC, int fCompl0, int fCompl1, int fComplC, Mf_Cut_t * pCutR )
func (p *MfMan) mfCutComputeTruthMux(pCut0 *MfCut, pCut1 *MfCut, pCutC *MfCut, fCompl0 int, fCompl1 int, fComplC int, pCutR *MfCut) int {
	panic("TODO, stubbed")
}

// Mf_CutCountBits (static inline)
// C: int Mf_CutCountBits( word i )
func mfCutCountBits(i uint64) int {
	panic("TODO, stubbed")
}

// Mf_CutGetSign (static inline)
// C: word Mf_CutGetSign( int * pLeaves, int nLeaves )
func mfCutGetSign(pLeaves []int, nLeaves int) uint64 {
	panic("TODO, stubbed")
}

// Mf_CutCreateUnit (static inline)
// C: int Mf_CutCreateUnit( Mf_Cut_t * p, int i )
func mfCutCreateUnit(p *MfCut, i int) int {
	panic("TODO, stubbed")
}

// Mf_CutPrint (static inline)
// C: void Mf_CutPrint( Mf_Man_t * p, Mf_Cut_t * pCut )
func (p *MfMan) mfCutPrint(pCut *MfCut) {
	panic("TODO, stubbed")
}

// Mf_ManPrepareCuts (static inline)
// C: int Mf_ManPrepareCuts( Mf_Cut_t * pCuts, Mf_Man_t * p, int iObj, int fAddUnit )
func mfManPrepareCuts(pCuts []MfCut, p *MfMan, iObj int, fAddUnit int) int {
	panic("TODO, stubbed")
}

// Mf_ManSaveCuts (static inline)
// C: int Mf_ManSaveCuts( Mf_Man_t * p, Mf_Cut_t ** pCuts, int nCuts )
func (p *MfMan) mfManSaveCuts(pCuts []*MfCut, nCuts int) int {
	panic("TODO, stubbed")
}

// Mf_ObjSetBestCut (static inline)
// C: void Mf_ObjSetBestCut( int * pCuts, int * pCut )
func mfObjSetBestCut(pCuts []int, pCut []int) {
	panic("TODO, stubbed")
}

// Mf_CutCheck (static inline)
// C: int Mf_CutCheck( Mf_Cut_t * pBase, Mf_Cut_t * pCut )
func mfCutCheck(pBase *MfCut, pCut *MfCut) int {
	panic("TODO, stubbed")
}

// Mf_SetCheckArray (static inline)
// C: int Mf_SetCheckArray( Mf_Cut_t ** ppCuts, int nCuts )
func mfSetCheckArray(ppCuts []*MfCut, nCuts int) int {
	panic("TODO, stubbed")
}

// Mf_CutMergeOrder (static inline)
// C: int Mf_CutMergeOrder( Mf_Cut_t * pCut0, Mf_Cut_t * pCut1, Mf_Cut_t * pCut, int nLutSize )
func mfCutMergeOrder(pCut0 *MfCut, pCut1 *MfCut, pCut *MfCut, nLutSize int) int {
	panic("TODO, stubbed")
}

// Mf_CutMergeOrderMux (static inline)
// C: int Mf_CutMergeOrderMux( Mf_Cut_t * pCut0, Mf_Cut_t * pCut1, Mf_Cut_t * pCut2, Mf_Cut_t * pCut, int nLutSize )
func mfCutMergeOrderMux(pCut0 *MfCut, pCut1 *MfCut, pCut2 *MfCut, pCut *MfCut, nLutSize int) int {
	panic("TODO, stubbed")
}

// Mf_SetCutIsContainedOrder (static inline)
// C: int Mf_SetCutIsContainedOrder( Mf_Cut_t * pBase, Mf_Cut_t * pCut )
func mfSetCutIsContainedOrder(pBase *MfCut, pCut *MfCut) int {
	panic("TODO, stubbed")
}

// Mf_SetLastCutIsContained (static inline)
// C: int Mf_SetLastCutIsContained( Mf_Cut_t ** pCuts, int nCuts )
func mfSetLastCutIsContained(pCuts []*MfCut, nCuts int) int {
	panic("TODO, stubbed")
}

// Mf_SetLastCutContainsArea (static inline)
// C: int Mf_SetLastCutContainsArea( Mf_Cut_t ** pCuts, int nCuts )
func mfSetLastCutContainsArea(pCuts []*MfCut, nCuts int) int {
	panic("TODO, stubbed")
}

// Mf_CutCompareArea (static inline)
// C: int Mf_CutCompareArea( Mf_Cut_t * pCut0, Mf_Cut_t * pCut1 )
func mfCutCompareArea(pCut0 *MfCut, pCut1 *MfCut) int {
	panic("TODO, stubbed")
}

// Mf_SetSortByArea (static inline)
// C: void Mf_SetSortByArea( Mf_Cut_t ** pCuts, int nCuts )
func mfSetSortByArea(pCuts []*MfCut, nCuts int) {
	panic("TODO, stubbed")
}

// Mf_SetAddCut (static inline)
// C: int Mf_SetAddCut( Mf_Cut_t ** pCuts, int nCuts, int nCutNum )
func mfSetAddCut(pCuts []*MfCut, nCuts int, nCutNum int) int {
	panic("TODO, stubbed")
}

// Mf_CutArea (static inline)
// C: int Mf_CutArea( Mf_Man_t * p, int nLeaves, int iFunc )
func (p *MfMan) mfCutArea(nLeaves int, iFunc int) int {
	panic("TODO, stubbed")
}

// Mf_CutParams (static inline)
// C: void Mf_CutParams( Mf_Man_t * p, Mf_Cut_t * pCut, float FlowRefs )
func (p *MfMan) mfCutParams(pCut *MfCut, FlowRefs float32) {
	panic("TODO, stubbed")
}

// Mf_ObjMergeOrder
// C: void Mf_ObjMergeOrder( Mf_Man_t * p, int iObj )
func (p *MfMan) MfObjMergeOrder(iObj int) {
	panic("TODO, stubbed")
}

// Mf_ManSetFlowRefs
// C: void Mf_ManSetFlowRefs( Gia_Man_t * p, Vec_Int_t * vRefs )
func (p *GiaManager) MfManSetFlowRefs(vRefs *IntVector) {
	panic("TODO, stubbed")
}

// Mf_ManSetMapRefs
// C: int Mf_ManSetMapRefs( Mf_Man_t * p )
func (p *MfMan) MfManSetMapRefs() int {
	panic("TODO, stubbed")
}

// Mf_ManDeriveMapping
// C: Gia_Man_t * Mf_ManDeriveMapping( Mf_Man_t * p )
func (p *MfMan) MfManDeriveMapping() *GiaManager {
	panic("TODO, stubbed")
}

// Mf_ManDeriveMappingCoarse
// C: Gia_Man_t * Mf_ManDeriveMappingCoarse( Mf_Man_t * p )
func (p *MfMan) MfManDeriveMappingCoarse() *GiaManager {
	panic("TODO, stubbed")
}

// Mf_ManDeriveMappingGia
// C: Gia_Man_t * Mf_ManDeriveMappingGia( Mf_Man_t * p )
func (p *MfMan) MfManDeriveMappingGia() *GiaManager {
	panic("TODO, stubbed")
}

// Mf_ManAlloc
// C: Mf_Man_t * Mf_ManAlloc( Gia_Man_t * pGia, Jf_Par_t * pPars )
func (pGia *GiaManager) MfManAlloc(pPars *JfPar) *MfMan {
	panic("TODO, stubbed")
}

// Mf_ManFree
// C: void Mf_ManFree( Mf_Man_t * p )
func (p *MfMan) MfManFree() {
	panic("TODO, stubbed")
}

// Mf_ManSetDefaultPars
// C: void Mf_ManSetDefaultPars( Jf_Par_t * pPars )
func MfManSetDefaultPars(pPars *JfPar) {
	panic("TODO, stubbed")
}

// Mf_ManPrintStats
// C: void Mf_ManPrintStats( Mf_Man_t * p, char * pTitle )
func (p *MfMan) MfManPrintStats(pTitle string) {
	panic("TODO, stubbed")
}

// Mf_ManPrintInit
// C: void Mf_ManPrintInit( Mf_Man_t * p )
func (p *MfMan) MfManPrintInit() {
	panic("TODO, stubbed")
}

// Mf_ManPrintQuit
// C: void Mf_ManPrintQuit( Mf_Man_t * p, Gia_Man_t * pNew )
func (p *MfMan) MfManPrintQuit(pNew *GiaManager) {
	panic("TODO, stubbed")
}

// Mf_ManComputeCuts
// C: void Mf_ManComputeCuts( Mf_Man_t * p )
func (p *MfMan) MfManComputeCuts() {
	panic("TODO, stubbed")
}

// Mf_CutRef_rec
// C: int Mf_CutRef_rec( Mf_Man_t * p, int * pCut )
func (p *MfMan) MfCutRef_rec(pCut []int) int {
	panic("TODO, stubbed")
}

// Mf_CutDeref_rec
// C: int Mf_CutDeref_rec( Mf_Man_t * p, int * pCut )
func (p *MfMan) MfCutDeref_rec(pCut []int) int {
	panic("TODO, stubbed")
}

// Mf_CutAreaRefed (static inline)
// C: int Mf_CutAreaRefed( Mf_Man_t * p, int * pCut )
func (p *MfMan) mfCutAreaRefed(pCut []int) int {
	panic("TODO, stubbed")
}

// Mf_CutAreaDerefed (static inline)
// C: int Mf_CutAreaDerefed( Mf_Man_t * p, int * pCut )
func (p *MfMan) mfCutAreaDerefed(pCut []int) int {
	panic("TODO, stubbed")
}

// Mf_CutAreaMffc (static inline)
// C: int Mf_CutAreaMffc( Mf_Man_t * p, int iObj )
func (p *MfMan) mfCutAreaMffc(iObj int) int {
	panic("TODO, stubbed")
}

// Mf_CutRef2_rec
// C: int Mf_CutRef2_rec( Mf_Man_t * p, int * pCut, Vec_Int_t * vTemp, int Limit )
func (p *MfMan) MfCutRef2_rec(pCut []int, vTemp *IntVector, Limit int) int {
	panic("TODO, stubbed")
}

// Mf_CutDeref2_rec
// C: int Mf_CutDeref2_rec( Mf_Man_t * p, int * pCut, Vec_Int_t * vTemp, int Limit )
func (p *MfMan) MfCutDeref2_rec(pCut []int, vTemp *IntVector, Limit int) int {
	panic("TODO, stubbed")
}

// Mf_CutAreaRefed2 (static inline)
// C: int Mf_CutAreaRefed2( Mf_Man_t * p, int * pCut )
func (p *MfMan) mfCutAreaRefed2(pCut []int) int {
	panic("TODO, stubbed")
}

// Mf_CutAreaDerefed2 (static inline)
// C: int Mf_CutAreaDerefed2( Mf_Man_t * p, int * pCut )
func (p *MfMan) mfCutAreaDerefed2(pCut []int) int {
	panic("TODO, stubbed")
}

// Mf_CutAreaRefed2Multi (static inline)
// C: int Mf_CutAreaRefed2Multi( Mf_Man_t * p, int iObj, int ** ppCuts, int nCuts )
func (p *MfMan) mfCutAreaRefed2Multi(iObj int, ppCuts [][]int, nCuts int) int {
	panic("TODO, stubbed")
}

// Mf_CutFlow (static inline)
// C: float Mf_CutFlow( Mf_Man_t * p, int * pCut, int * pTime )
func (p *MfMan) mfCutFlow(pCut []int, pTime *int) float32 {
	panic("TODO, stubbed")
}

// Mf_ObjComputeBestCut (static inline)
// C: void Mf_ObjComputeBestCut( Mf_Man_t * p, int iObj )
func (p *MfMan) mfObjComputeBestCut(iObj int) {
	panic("TODO, stubbed")
}

// Mf_ManMappingFromMapping
// C: int Mf_ManMappingFromMapping( Mf_Man_t * p )
func (p *MfMan) MfManMappingFromMapping() int {
	panic("TODO, stubbed")
}

// Mf_ManPrintFanoutProfile
// C: void Mf_ManPrintFanoutProfile( Mf_Man_t * p, Vec_Int_t * vFanCounts )
func (p *MfMan) MfManPrintFanoutProfile(vFanCounts *IntVector) {
	panic("TODO, stubbed")
}

// Mf_ManPrintMfccStats
// C: int Mf_ManPrintMfccStats( Mf_Man_t * p, int iObj )
func (p *MfMan) MfManPrintMfccStats(iObj int) int {
	panic("TODO, stubbed")
}

// Mf_ManOptimizationOne
// C: void Mf_ManOptimizationOne( Mf_Man_t * p, int iObj )
func (p *MfMan) MfManOptimizationOne(iObj int) {
	panic("TODO, stubbed")
}

// Mf_ManOptimization
// C: void Mf_ManOptimization( Mf_Man_t * p )
func (p *MfMan) MfManOptimization() {
	panic("TODO, stubbed")
}

// Mf_ManComputeMapping
// C: void Mf_ManComputeMapping( Mf_Man_t * p )
func (p *MfMan) MfManComputeMapping() {
	panic("TODO, stubbed")
}

// Mf_ManPerformMapping
// C: Gia_Man_t * Mf_ManPerformMapping( Gia_Man_t * pGia, Jf_Par_t * pPars )
func (pGia *GiaManager) MfManPerformMapping(pPars *JfPar) *GiaManager {
	panic("TODO, stubbed")
}

// Mf_ManGenerateCnf (returns a void *)
// C: void * Mf_ManGenerateCnf( Gia_Man_t * pGia, int nLutSize, int fCnfObjIds, int fAddOrCla, int fMapping, int fVerbose )
func (pGia *GiaManager) MfManGenerateCnf(nLutSize int, fCnfObjIds int, fAddOrCla int, fMapping int, fVerbose int) *CnfDat {
	panic("TODO, stubbed")
}

// Mf_ManDumpCnf
// C: void Mf_ManDumpCnf( Gia_Man_t * p, char * pFileName, int nLutSize, int fCnfObjIds, int fAddOrCla, int fVerbose )
func (p *GiaManager) MfManDumpCnf(pFileName string, nLutSize int, fCnfObjIds int, fAddOrCla int, fVerbose int) {
	panic("TODO, stubbed")
}
