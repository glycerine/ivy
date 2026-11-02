package abc

// Hand-checked skeleton of src/opt/sbd/sbdCut.c.

// SbdCut is Sbd_Cut_t (defined identically in sbdCut.c and sbdCut2.c).
// SBD_MAX_CUTSIZE is 10 and SBD_MAX_CUTNUM is 501. Go has no bitfields; each
// is a uint32 whose C width is in the comment.
type SbdCut struct {
	Sign        uint64  // word : signature
	iFunc       int     // functionality
	Cost        int     // cut cost
	CostLev     int     // cut cost
	nTreeLeaves uint32  // unsigned : 9, tree leaves
	nSlowLeaves uint32  // unsigned : 9, slow leaves
	nTopLeaves  uint32  // unsigned : 10, top leaves
	nLeaves     uint32  // unsigned : 4, leaf count
	pLeaves     [10]int // leaves
}

// SbdSto is Sbd_Sto_t.
type SbdSto struct {
	nLutSize  int
	nCutSize  int
	nCutNum   int
	fCutMin   int
	fVerbose  int
	pGia      *GiaManager     // user's AIG manager (will be modified by adding nodes)
	vMirrors  *IntVector      // mirrors for each node
	vDelays   *IntVector      // delays for each node
	vLevels   *IntVector      // levels for each node
	vRefs     *IntVector      // refs for each node
	vCuts     *IntLevelVector // cuts for each node
	vTtMem    *WordMemory     // truth tables
	pCuts     [3][501]SbdCut  // temporary cuts
	ppCuts    [501]*SbdCut    // temporary cut pointers
	nCutsR    int             // the number of cuts
	Pivot     int             // current object
	iCutBest  int             // best-delay cut
	nCutsSpec int             // special cuts
	nCutsOver int             // overflow cuts
	DelayMin  int             // minimum delay
	CutCount  [4]float64      // cut counters
	clkStart  int64           // abctime : starting time
}

// Sbd_CutTruth (static; a pointer into the truth-table memory)
// C: word * Sbd_CutTruth( Sbd_Sto_t * p, Sbd_Cut_t * pCut )
func (p *SbdSto) sbdCutTruth(pCut *SbdCut) []uint64 {
	panic("TODO, stubbed")
}

// Sbd_CutGetSign (static)
// C: word Sbd_CutGetSign( Sbd_Cut_t * pCut )
func sbdCutGetSign(pCut *SbdCut) uint64 {
	panic("TODO, stubbed")
}

// Sbd_CutCheck (static)
// C: int Sbd_CutCheck( Sbd_Cut_t * pBase, Sbd_Cut_t * pCut )
func sbdCutCheck(pBase *SbdCut, pCut *SbdCut) int {
	panic("TODO, stubbed")
}

// Sbd_CutSetCheckArray (static)
// C: int Sbd_CutSetCheckArray( Sbd_Cut_t ** ppCuts, int nCuts )
func sbdCutSetCheckArray(ppCuts []*SbdCut, nCuts int) int {
	panic("TODO, stubbed")
}

// Sbd_CutMergeOrder (static)
// C: int Sbd_CutMergeOrder( Sbd_Cut_t * pCut0, Sbd_Cut_t * pCut1, Sbd_Cut_t * pCut, int nCutSize )
func sbdCutMergeOrder(pCut0 *SbdCut, pCut1 *SbdCut, pCut *SbdCut, nCutSize int) int {
	panic("TODO, stubbed")
}

// Sbd_CutMergeOrder2 (static)
// C: int Sbd_CutMergeOrder2( Sbd_Cut_t * pCut0, Sbd_Cut_t * pCut1, Sbd_Cut_t * pCut, int nCutSize )
func sbdCutMergeOrder2(pCut0 *SbdCut, pCut1 *SbdCut, pCut *SbdCut, nCutSize int) int {
	panic("TODO, stubbed")
}

// Sbd_CutSetCutIsContainedOrder (static)
// C: int Sbd_CutSetCutIsContainedOrder( Sbd_Cut_t * pBase, Sbd_Cut_t * pCut )
func sbdCutSetCutIsContainedOrder(pBase *SbdCut, pCut *SbdCut) int {
	panic("TODO, stubbed")
}

// Sbd_CutSetLastCutIsContained (static)
// C: int Sbd_CutSetLastCutIsContained( Sbd_Cut_t ** pCuts, int nCuts )
func sbdCutSetLastCutIsContained(pCuts []*SbdCut, nCuts int) int {
	panic("TODO, stubbed")
}

// Sbd_CutCompare (static)
// C: int Sbd_CutCompare( Sbd_Cut_t * pCut0, Sbd_Cut_t * pCut1 )
func sbdCutCompare(pCut0 *SbdCut, pCut1 *SbdCut) int {
	panic("TODO, stubbed")
}

// Sbd_CutCompare2 (static)
// C: int Sbd_CutCompare2( Sbd_Cut_t * pCut0, Sbd_Cut_t * pCut1 )
func sbdCutCompare2(pCut0 *SbdCut, pCut1 *SbdCut) int {
	panic("TODO, stubbed")
}

// Sbd_CutSetLastCutContains (static)
// C: int Sbd_CutSetLastCutContains( Sbd_Cut_t ** pCuts, int nCuts )
func sbdCutSetLastCutContains(pCuts []*SbdCut, nCuts int) int {
	panic("TODO, stubbed")
}

// Sbd_CutSetSortByCost (static)
// C: void Sbd_CutSetSortByCost( Sbd_Cut_t ** pCuts, int nCuts )
func sbdCutSetSortByCost(pCuts []*SbdCut, nCuts int) {
	panic("TODO, stubbed")
}

// Sbd_CutSetAddCut (static)
// C: int Sbd_CutSetAddCut( Sbd_Cut_t ** pCuts, int nCuts, int nCutNum )
func sbdCutSetAddCut(pCuts []*SbdCut, nCuts int, nCutNum int) int {
	panic("TODO, stubbed")
}

// Sbd_CutComputeTruth6 (static)
// C: int Sbd_CutComputeTruth6( Sbd_Sto_t * p, Sbd_Cut_t * pCut0, Sbd_Cut_t * pCut1, int fCompl0, int fCompl1, Sbd_Cut_t * pCutR, int fIsXor )
func (p *SbdSto) sbdCutComputeTruth6(pCut0 *SbdCut, pCut1 *SbdCut, fCompl0 int, fCompl1 int, pCutR *SbdCut, fIsXor int) int {
	panic("TODO, stubbed")
}

// Sbd_CutComputeTruth (static)
// C: int Sbd_CutComputeTruth( Sbd_Sto_t * p, Sbd_Cut_t * pCut0, Sbd_Cut_t * pCut1, int fCompl0, int fCompl1, Sbd_Cut_t * pCutR, int fIsXor )
func (p *SbdSto) sbdCutComputeTruth(pCut0 *SbdCut, pCut1 *SbdCut, fCompl0 int, fCompl1 int, pCutR *SbdCut, fIsXor int) int {
	panic("TODO, stubbed")
}

// Sbd_CutCountBits (static)
// C: int Sbd_CutCountBits( word i )
func sbdCutCountBits(i uint64) int {
	panic("TODO, stubbed")
}

// Sbd_CutCost (static)
// C: int Sbd_CutCost( Sbd_Sto_t * p, Sbd_Cut_t * pCut )
func (p *SbdSto) sbdCutCost(pCut *SbdCut) int {
	panic("TODO, stubbed")
}

// Sbd_CutCostLev (static)
// C: int Sbd_CutCostLev( Sbd_Sto_t * p, Sbd_Cut_t * pCut )
func (p *SbdSto) sbdCutCostLev(pCut *SbdCut) int {
	panic("TODO, stubbed")
}

// Sbd_CutTreeLeaves (static)
// C: int Sbd_CutTreeLeaves( Sbd_Sto_t * p, Sbd_Cut_t * pCut )
func (p *SbdSto) sbdCutTreeLeaves(pCut *SbdCut) int {
	panic("TODO, stubbed")
}

// Sbd_CutSlowLeaves (static)
// C: int Sbd_CutSlowLeaves( Sbd_Sto_t * p, int iObj, Sbd_Cut_t * pCut )
func (p *SbdSto) sbdCutSlowLeaves(iObj int, pCut *SbdCut) int {
	panic("TODO, stubbed")
}

// Sbd_CutTopLeaves (static)
// C: int Sbd_CutTopLeaves( Sbd_Sto_t * p, int iObj, Sbd_Cut_t * pCut )
func (p *SbdSto) sbdCutTopLeaves(iObj int, pCut *SbdCut) int {
	panic("TODO, stubbed")
}

// Sbd_CutAddUnit (static)
// C: void Sbd_CutAddUnit( Sbd_Sto_t * p, int iObj )
func (p *SbdSto) sbdCutAddUnit(iObj int) {
	panic("TODO, stubbed")
}

// Sbd_CutAddZero (static)
// C: void Sbd_CutAddZero( Sbd_Sto_t * p, int iObj )
func (p *SbdSto) sbdCutAddZero(iObj int) {
	panic("TODO, stubbed")
}

// Sbd_StoPrepareSet (static)
// C: int Sbd_StoPrepareSet( Sbd_Sto_t * p, int iObj, int Index )
func (p *SbdSto) sbdStoPrepareSet(iObj int, Index int) int {
	panic("TODO, stubbed")
}

// Sbd_StoInitResult (static)
// C: void Sbd_StoInitResult( Sbd_Sto_t * p )
func (p *SbdSto) sbdStoInitResult() {
	panic("TODO, stubbed")
}

// Sbd_StoStoreResult (static)
// C: void Sbd_StoStoreResult( Sbd_Sto_t * p, int iObj, Sbd_Cut_t ** pCuts, int nCuts )
func (p *SbdSto) sbdStoStoreResult(iObj int, pCuts []*SbdCut, nCuts int) {
	panic("TODO, stubbed")
}

// Sbd_StoComputeDelay (static)
// C: void Sbd_StoComputeDelay( Sbd_Sto_t * p, int iObj, Sbd_Cut_t ** pCuts, int nCuts )
func (p *SbdSto) sbdStoComputeDelay(iObj int, pCuts []*SbdCut, nCuts int) {
	panic("TODO, stubbed")
}

// Sbd_StoComputeSpec (static)
// C: void Sbd_StoComputeSpec( Sbd_Sto_t * p, int iObj, Sbd_Cut_t ** pCuts, int nCuts )
func (p *SbdSto) sbdStoComputeSpec(iObj int, pCuts []*SbdCut, nCuts int) {
	panic("TODO, stubbed")
}

// Sbd_CutPrint (static)
// C: void Sbd_CutPrint( Sbd_Sto_t * p, int iObj, Sbd_Cut_t * pCut )
func (p *SbdSto) sbdCutPrint(iObj int, pCut *SbdCut) {
	panic("TODO, stubbed")
}

// Sbd_StoMergeCuts
// C: void Sbd_StoMergeCuts( Sbd_Sto_t * p, int iObj )
func (p *SbdSto) SbdStoMergeCuts(iObj int) {
	panic("TODO, stubbed")
}

// Sbd_StoAlloc
// C: Sbd_Sto_t * Sbd_StoAlloc( Gia_Man_t * pGia, Vec_Int_t * vMirrors, int nLutSize, int nCutSize, int nCutNum, int fCutMin, int fVerbose )
func (pGia *GiaManager) SbdStoAlloc(vMirrors *IntVector, nLutSize int, nCutSize int, nCutNum int, fCutMin int, fVerbose int) *SbdSto {
	panic("TODO, stubbed")
}

// Sbd_StoFree
// C: void Sbd_StoFree( Sbd_Sto_t * p )
func (p *SbdSto) SbdStoFree() {
	panic("TODO, stubbed")
}

// Sbd_StoComputeCutsObj
// C: void Sbd_StoComputeCutsObj( Sbd_Sto_t * p, int iObj, int Delay, int Level )
func (p *SbdSto) SbdStoComputeCutsObj(iObj int, Delay int, Level int) {
	panic("TODO, stubbed")
}

// Sbd_StoComputeCutsConst0
// C: void Sbd_StoComputeCutsConst0( Sbd_Sto_t * p, int iObj )
func (p *SbdSto) SbdStoComputeCutsConst0(iObj int) {
	panic("TODO, stubbed")
}

// Sbd_StoComputeCutsCi
// C: void Sbd_StoComputeCutsCi( Sbd_Sto_t * p, int iObj, int Delay, int Level )
func (p *SbdSto) SbdStoComputeCutsCi(iObj int, Delay int, Level int) {
	panic("TODO, stubbed")
}

// Sbd_StoComputeCutsNode
// C: int Sbd_StoComputeCutsNode( Sbd_Sto_t * p, int iObj )
func (p *SbdSto) SbdStoComputeCutsNode(iObj int) int {
	panic("TODO, stubbed")
}

// Sbd_StoSaveBestDelayCut
// C: void Sbd_StoSaveBestDelayCut( Sbd_Sto_t * p, int iObj, int * pCut )
func (p *SbdSto) SbdStoSaveBestDelayCut(iObj int, pCut []int) {
	panic("TODO, stubbed")
}

// Sbd_StoObjRefs
// C: int Sbd_StoObjRefs( Sbd_Sto_t * p, int iObj )
func (p *SbdSto) SbdStoObjRefs(iObj int) int {
	panic("TODO, stubbed")
}

// Sbd_StoRefObj
// C: void Sbd_StoRefObj( Sbd_Sto_t * p, int iObj, int iMirror )
func (p *SbdSto) SbdStoRefObj(iObj int, iMirror int) {
	panic("TODO, stubbed")
}

// Sbd_StoDerefObj
// C: void Sbd_StoDerefObj( Sbd_Sto_t * p, int iObj )
func (p *SbdSto) SbdStoDerefObj(iObj int) {
	panic("TODO, stubbed")
}

// Sbd_StoObjBestCut
// C: int Sbd_StoObjBestCut( Sbd_Sto_t * p, int iObj, int nSize, int * pLeaves )
func (p *SbdSto) SbdStoObjBestCut(iObj int, nSize int, pLeaves []int) int {
	panic("TODO, stubbed")
}

// Sbd_StoComputeCutsTest
// C: void Sbd_StoComputeCutsTest( Gia_Man_t * pGia )
func (pGia *GiaManager) SbdStoComputeCutsTest() {
	panic("TODO, stubbed")
}
