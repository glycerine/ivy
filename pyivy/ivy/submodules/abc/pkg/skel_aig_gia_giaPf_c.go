package abc

// Hand-checked skeleton of src/aig/gia/giaPf.c.

const (
	PF_LEAF_MAX = 6
	PF_CUT_MAX  = 32
	PF_NO_LEAF  = 31
	PF_NO_FUNC  = 0x3FFFFFF
	PF_INFINITY = 3.40282346638528859811704183484516925440e+38 // FLT_MAX
)

// PfCut is Pf_Cut_t.
type PfCut struct {
	Sign    uint64               // word : signature
	Delay   int                  // delay
	Flow    float32              // flow
	iFunc   uint32               // unsigned : 26 : function (PF_NO_FUNC)
	Useless uint32               // unsigned : 1 : function
	nLeaves uint32               // unsigned : 5 : leaf number (PF_NO_LEAF)
	pLeaves [PF_LEAF_MAX + 1]int // leaves
}

// PfMat is Pf_Mat_t.
type PfMat struct {
	fCompl uint32 // unsigned : 8 : complemented
	Phase  uint32 // unsigned : 6 : match phase
	Perm   uint32 // unsigned : 18 : match permutation
}

// PfObj is Pf_Obj_t.
type PfObj struct {
	Area    float32
	Gate    uint32 // unsigned : 7 : gate
	nLeaves uint32 // unsigned : 3 : fanin count
	nRefs   uint32 // unsigned : 22 : ref count
	pLeaves [6]int // leaf literals
}

// PfMan is Pf_Man_t.
type PfMan struct {
	pGia       *GiaManager        // derived manager
	pPars      *JfPar             // parameters
	vTtMem     *WordMemory        // truth tables
	vTt2Match  *IntLevelVector    // matches for truth tables
	pCells     []MioCell          // Mio_Cell_t * : library gates
	nCells     int                // library gate count
	pPfObjs    []PfObj            // Pf_Obj_t * : best cuts
	vPages     PointerVector[int] // Vec_Ptr_t (by value) : cut memory
	vCutSets   IntVector          // Vec_Int_t (by value) : cut offsets
	vCutFlows  FloatVector        // Vec_Flt_t (by value) : temporary cut area
	vCutDelays IntVector          // Vec_Int_t (by value) : temporary cut delay
	iCur       int                // current position
	Iter       int                // mapping iterations
	fUseEla    int                // use exact area
	nInvs      int                // the inverter count
	InvDelay   float32            // inverter delay
	InvArea    float32            // inverter area
	clkStart   int64              // abctime : starting time
	CutCount   [6]float64         // double : cut counts
	nCutUseAll int                // objects with useful cuts
}

// Pf_Mat2Int (static inline)
// C: int Pf_Mat2Int( Pf_Mat_t Mat )
func pfMat2Int(Mat PfMat) int {
	panic("TODO, stubbed")
}

// Pf_Int2Mat (static inline)
// C: Pf_Mat_t Pf_Int2Mat( int Int )
func pfInt2Mat(Int int) PfMat {
	panic("TODO, stubbed")
}

// Pf_ManObj (static inline)
// C: Pf_Obj_t * Pf_ManObj( Pf_Man_t * p, int i )
func (p *PfMan) pfManObj(i int) *PfObj {
	panic("TODO, stubbed")
}

// Pf_ManCell (static inline)
// C: Mio_Cell_t* Pf_ManCell( Pf_Man_t * p, int i )
func (p *PfMan) pfManCell(i int) *MioCell {
	panic("TODO, stubbed")
}

// Pf_ManCutSet (static inline)
// C: int * Pf_ManCutSet( Pf_Man_t * p, int i )
func (p *PfMan) pfManCutSet(i int) []int {
	panic("TODO, stubbed")
}

// Pf_ObjCutSetId (static inline)
// C: int Pf_ObjCutSetId( Pf_Man_t * p, int i )
func (p *PfMan) pfObjCutSetId(i int) int {
	panic("TODO, stubbed")
}

// Pf_ObjCutSet (static inline)
// C: int * Pf_ObjCutSet( Pf_Man_t * p, int i )
func (p *PfMan) pfObjCutSet(i int) []int {
	panic("TODO, stubbed")
}

// Pf_ObjHasCuts (static inline)
// C: int Pf_ObjHasCuts( Pf_Man_t * p, int i )
func (p *PfMan) pfObjHasCuts(i int) int {
	panic("TODO, stubbed")
}

// Pf_ObjCutUseless (static inline)
// C: int Pf_ObjCutUseless( Pf_Man_t * p, int TruthId )
func (p *PfMan) pfObjCutUseless(TruthId int) int {
	panic("TODO, stubbed")
}

// Pf_ObjCutFlow (static inline)
// C: float Pf_ObjCutFlow( Pf_Man_t * p, int i )
func (p *PfMan) pfObjCutFlow(i int) float32 {
	panic("TODO, stubbed")
}

// Pf_ObjCutDelay (static inline)
// C: int Pf_ObjCutDelay( Pf_Man_t * p, int i )
func (p *PfMan) pfObjCutDelay(i int) int {
	panic("TODO, stubbed")
}

// Pf_ObjSetCutFlow (static inline)
// C: void Pf_ObjSetCutFlow( Pf_Man_t * p, int i, float a )
func (p *PfMan) pfObjSetCutFlow(i int, a float32) {
	panic("TODO, stubbed")
}

// Pf_ObjSetCutDelay (static inline)
// C: void Pf_ObjSetCutDelay( Pf_Man_t * p, int i, int d )
func (p *PfMan) pfObjSetCutDelay(i int, d int) {
	panic("TODO, stubbed")
}

// Pf_CutSize (static inline)
// C: int Pf_CutSize( int * pCut )
func pfCutSize(pCut []int) int {
	panic("TODO, stubbed")
}

// Pf_CutFunc (static inline)
// C: int Pf_CutFunc( int * pCut )
func pfCutFunc(pCut []int) int {
	panic("TODO, stubbed")
}

// Pf_CutLeaves (static inline)
// C: int * Pf_CutLeaves( int * pCut )
func pfCutLeaves(pCut []int) []int {
	panic("TODO, stubbed")
}

// Pf_CutSetBoth (static inline)
// C: int Pf_CutSetBoth( int n, int f )
func pfCutSetBoth(n int, f int) int {
	panic("TODO, stubbed")
}

// Pf_CutIsTriv (static inline)
// C: int Pf_CutIsTriv( int * pCut, int i )
func pfCutIsTriv(pCut []int, i int) int {
	panic("TODO, stubbed")
}

// Pf_CutHandle (static inline)
// C: int Pf_CutHandle( int * pCutSet, int * pCut )
func pfCutHandle(pCutSet []int, pCut []int) int {
	panic("TODO, stubbed")
}

// Pf_CutFromHandle (static inline)
// C: int * Pf_CutFromHandle( int * pCutSet, int h )
func pfCutFromHandle(pCutSet []int, h int) []int {
	panic("TODO, stubbed")
}

// Pf_CutConfLit (static inline)
// C: int Pf_CutConfLit( int Conf, int i )
func pfCutConfLit(Conf int, i int) int {
	panic("TODO, stubbed")
}

// Pf_CutConfVar (static inline)
// C: int Pf_CutConfVar( int Conf, int i )
func pfCutConfVar(Conf int, i int) int {
	panic("TODO, stubbed")
}

// Pf_CutConfC (static inline)
// C: int Pf_CutConfC( int Conf, int i )
func pfCutConfC(Conf int, i int) int {
	panic("TODO, stubbed")
}

// Pf_StoCreateGateAdd
// C: void Pf_StoCreateGateAdd( Pf_Man_t * pMan, word uTruth, int * pFans, int nFans, int CellId )
func (pMan *PfMan) PfStoCreateGateAdd(uTruth uint64, pFans []int, nFans int, CellId int) {
	panic("TODO, stubbed")
}

// Pf_StoCreateGate
// C: void Pf_StoCreateGate( Pf_Man_t * pMan, Mio_Cell_t * pCell, int ** pComp, int ** pPerm, int * pnPerms )
func (pMan *PfMan) PfStoCreateGate(pCell *MioCell, pComp [][]int, pPerm [][]int, pnPerms []int) {
	panic("TODO, stubbed")
}

// Pf_StoDeriveMatches
// C: void Pf_StoDeriveMatches( Pf_Man_t * p, int fVerbose )
func (p *PfMan) PfStoDeriveMatches(fVerbose int) {
	panic("TODO, stubbed")
}

// Pf_StoPrintOne
// C: void Pf_StoPrintOne( Pf_Man_t * p, int Count, int t, int i, int GateId, Pf_Mat_t Mat )
func (p *PfMan) PfStoPrintOne(Count int, t int, i int, GateId int, Mat PfMat) {
	panic("TODO, stubbed")
}

// Pf_StoPrint
// C: void Pf_StoPrint( Pf_Man_t * p, int fVerbose )
func (p *PfMan) PfStoPrint(fVerbose int) {
	panic("TODO, stubbed")
}

// Pf_StoCreate
// C: Pf_Man_t * Pf_StoCreate( Gia_Man_t * pGia, Jf_Par_t * pPars )
func (pGia *GiaManager) PfStoCreate(pPars *JfPar) *PfMan {
	panic("TODO, stubbed")
}

// Pf_StoDelete
// C: void Pf_StoDelete( Pf_Man_t * p )
func (p *PfMan) PfStoDelete() {
	panic("TODO, stubbed")
}

// Pf_CutComputeTruth6 (static inline)
// C: int Pf_CutComputeTruth6( Pf_Man_t * p, Pf_Cut_t * pCut0, Pf_Cut_t * pCut1, int fCompl0, int fCompl1, Pf_Cut_t * pCutR, int fIsXor )
func (p *PfMan) pfCutComputeTruth6(pCut0 *PfCut, pCut1 *PfCut, fCompl0 int, fCompl1 int, pCutR *PfCut, fIsXor int) int {
	panic("TODO, stubbed")
}

// Pf_CutComputeTruthMux6 (static inline)
// C: int Pf_CutComputeTruthMux6( Pf_Man_t * p, Pf_Cut_t * pCut0, Pf_Cut_t * pCut1, Pf_Cut_t * pCutC, int fCompl0, int fCompl1, int fComplC, Pf_Cut_t * pCutR )
func (p *PfMan) pfCutComputeTruthMux6(pCut0 *PfCut, pCut1 *PfCut, pCutC *PfCut, fCompl0 int, fCompl1 int, fComplC int, pCutR *PfCut) int {
	panic("TODO, stubbed")
}

// Pf_CutCountBits (static inline)
// C: int Pf_CutCountBits( word i )
func pfCutCountBits(i uint64) int {
	panic("TODO, stubbed")
}

// Pf_CutGetSign (static inline)
// C: word Pf_CutGetSign( int * pLeaves, int nLeaves )
func pfCutGetSign(pLeaves []int, nLeaves int) uint64 {
	panic("TODO, stubbed")
}

// Pf_CutCreateUnit (static inline)
// C: int Pf_CutCreateUnit( Pf_Cut_t * p, int i )
func pfCutCreateUnit(p *PfCut, i int) int {
	panic("TODO, stubbed")
}

// Pf_Cutprintf (static inline)
// C: void Pf_Cutprintf( Pf_Man_t * p, Pf_Cut_t * pCut )
func (p *PfMan) pfCutprintf(pCut *PfCut) {
	panic("TODO, stubbed")
}

// Pf_ManPrepareCuts (static inline)
// C: int Pf_ManPrepareCuts( Pf_Cut_t * pCuts, Pf_Man_t * p, int iObj, int fAddUnit )
func pfManPrepareCuts(pCuts []PfCut, p *PfMan, iObj int, fAddUnit int) int {
	panic("TODO, stubbed")
}

// Pf_ManSaveCuts (static inline)
// C: int Pf_ManSaveCuts( Pf_Man_t * p, Pf_Cut_t ** pCuts, int nCuts, int fUseful )
func (p *PfMan) pfManSaveCuts(pCuts []*PfCut, nCuts int, fUseful int) int {
	panic("TODO, stubbed")
}

// Pf_ManCountUseful (static inline)
// C: int Pf_ManCountUseful( Pf_Cut_t ** pCuts, int nCuts )
func pfManCountUseful(pCuts []*PfCut, nCuts int) int {
	panic("TODO, stubbed")
}

// Pf_ManCountMatches (static inline)
// C: int Pf_ManCountMatches( Pf_Man_t * p, Pf_Cut_t ** pCuts, int nCuts )
func (p *PfMan) pfManCountMatches(pCuts []*PfCut, nCuts int) int {
	panic("TODO, stubbed")
}

// Pf_CutCheck (static inline)
// C: int Pf_CutCheck( Pf_Cut_t * pBase, Pf_Cut_t * pCut )
func pfCutCheck(pBase *PfCut, pCut *PfCut) int {
	panic("TODO, stubbed")
}

// Pf_SetCheckArray (static inline)
// C: int Pf_SetCheckArray( Pf_Cut_t ** ppCuts, int nCuts )
func pfSetCheckArray(ppCuts []*PfCut, nCuts int) int {
	panic("TODO, stubbed")
}

// Pf_CutMergeOrder (static inline)
// C: int Pf_CutMergeOrder( Pf_Cut_t * pCut0, Pf_Cut_t * pCut1, Pf_Cut_t * pCut, int nLutSize )
func pfCutMergeOrder(pCut0 *PfCut, pCut1 *PfCut, pCut *PfCut, nLutSize int) int {
	panic("TODO, stubbed")
}

// Pf_CutMergeOrderMux (static inline)
// C: int Pf_CutMergeOrderMux( Pf_Cut_t * pCut0, Pf_Cut_t * pCut1, Pf_Cut_t * pCut2, Pf_Cut_t * pCut, int nLutSize )
func pfCutMergeOrderMux(pCut0 *PfCut, pCut1 *PfCut, pCut2 *PfCut, pCut *PfCut, nLutSize int) int {
	panic("TODO, stubbed")
}

// Pf_SetCutIsContainedOrder (static inline)
// C: int Pf_SetCutIsContainedOrder( Pf_Cut_t * pBase, Pf_Cut_t * pCut )
func pfSetCutIsContainedOrder(pBase *PfCut, pCut *PfCut) int {
	panic("TODO, stubbed")
}

// Pf_SetLastCutIsContained (static inline)
// C: int Pf_SetLastCutIsContained( Pf_Cut_t ** pCuts, int nCuts )
func pfSetLastCutIsContained(pCuts []*PfCut, nCuts int) int {
	panic("TODO, stubbed")
}

// Pf_SetLastCutContainsArea (static inline)
// C: int Pf_SetLastCutContainsArea( Pf_Cut_t ** pCuts, int nCuts )
func pfSetLastCutContainsArea(pCuts []*PfCut, nCuts int) int {
	panic("TODO, stubbed")
}

// Pf_CutCompareArea (static inline)
// C: int Pf_CutCompareArea( Pf_Cut_t * pCut0, Pf_Cut_t * pCut1 )
func pfCutCompareArea(pCut0 *PfCut, pCut1 *PfCut) int {
	panic("TODO, stubbed")
}

// Pf_SetSortByArea (static inline)
// C: void Pf_SetSortByArea( Pf_Cut_t ** pCuts, int nCuts )
func pfSetSortByArea(pCuts []*PfCut, nCuts int) {
	panic("TODO, stubbed")
}

// Pf_SetAddCut (static inline)
// C: int Pf_SetAddCut( Pf_Cut_t ** pCuts, int nCuts, int nCutNum )
func pfSetAddCut(pCuts []*PfCut, nCuts int, nCutNum int) int {
	panic("TODO, stubbed")
}

// Pf_CutArea (static inline)
// C: int Pf_CutArea( Pf_Man_t * p, int nLeaves )
func (p *PfMan) pfCutArea(nLeaves int) int {
	panic("TODO, stubbed")
}

// Pf_CutParams (static inline)
// C: void Pf_CutParams( Pf_Man_t * p, Pf_Cut_t * pCut, int nGiaRefs )
func (p *PfMan) pfCutParams(pCut *PfCut, nGiaRefs int) {
	panic("TODO, stubbed")
}

// Pf_ObjMergeOrder
// C: void Pf_ObjMergeOrder( Pf_Man_t * p, int iObj )
func (p *PfMan) PfObjMergeOrder(iObj int) {
	panic("TODO, stubbed")
}

// Pf_ManComputeCuts
// C: void Pf_ManComputeCuts( Pf_Man_t * p )
func (p *PfMan) PfManComputeCuts() {
	panic("TODO, stubbed")
}

// Pf_ManPrintStats
// C: void Pf_ManPrintStats( Pf_Man_t * p, char * pTitle )
func (p *PfMan) PfManPrintStats(pTitle string) {
	panic("TODO, stubbed")
}

// Pf_ManPrintInit
// C: void Pf_ManPrintInit( Pf_Man_t * p )
func (p *PfMan) PfManPrintInit() {
	panic("TODO, stubbed")
}

// Pf_ManPrintQuit
// C: void Pf_ManPrintQuit( Pf_Man_t * p )
func (p *PfMan) PfManPrintQuit() {
	panic("TODO, stubbed")
}

// Pf_ManComputeMapping
// C: void Pf_ManComputeMapping( Pf_Man_t * p )
func (p *PfMan) PfManComputeMapping() {
	panic("TODO, stubbed")
}

// Pf_ManSetDefaultPars
// C: void Pf_ManSetDefaultPars( Jf_Par_t * pPars )
func PfManSetDefaultPars(pPars *JfPar) {
	panic("TODO, stubbed")
}

// Pf_ManPerformMapping
// C: Gia_Man_t * Pf_ManPerformMapping( Gia_Man_t * pGia, Jf_Par_t * pPars )
func (pGia *GiaManager) PfManPerformMapping(pPars *JfPar) *GiaManager {
	panic("TODO, stubbed")
}
