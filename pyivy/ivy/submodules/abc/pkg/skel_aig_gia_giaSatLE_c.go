package abc

// Hand-checked skeleton of src/aig/gia/giaSatLE.c.

// SleMan is Sle_Man_t.
type SleMan struct {
	pGia         *GiaManager     // user's manager (with mapping and edges)
	nLevels      int             // total number of levels
	fVerbose     int             // verbose flag
	nSatCalls    int             // the number of SAT calls
	nNodeVars    int             // node variables (Gia_ManAndNum(pGia))
	nCutVars     int             // cut variables (total number of non-trivial cuts)
	nEdgeVars    int             // edge variables (total number of internal edges)
	nDelayVars   int             // delay variables (nNodeVars * nLevelsMax)
	nVarsTotal   int             // total number of variables
	nCutClas     int             // cut clauses
	nEdgeClas    int             // edge clauses
	nEdgeClas2   int             // edge clauses exclusivity
	nDelayClas   int             // delay clauses
	pSat         *SatSolver      // SAT solver
	vMask        *BitVector      // internal node mask
	vCuts        *IntVector      // cuts for each node
	vCutFanins   *IntLevelVector // internal cut fanins of each node
	vFanoutEdges *IntLevelVector // internal cut fanins of each node
	vEdgeCuts    *IntLevelVector // cuts of each edge for one node
	vObjMap      *IntVector      // temporary object map
	vCutFirst    *IntVector      // first cut of each node
	vEdgeFirst   *IntVector      // first edge of each node
	vDelayFirst  *IntVector      // first edge of each node
	vPolars      *IntVector      // initial
	vLits        *IntVector      // literals
	timeStart    int64           // abctime
}

// Sle_CutSize (static inline)
// C: int Sle_CutSize( int * pCut )
func sleCutSize(pCut []int) int {
	panic("TODO, stubbed")
}

// Sle_CutSign (static inline)
// C: int Sle_CutSign( int * pCut )
func sleCutSign(pCut []int) int {
	panic("TODO, stubbed")
}

// Sle_CutSetSizeSign (static inline)
// C: int Sle_CutSetSizeSign( int s, int S )
func sleCutSetSizeSign(s int, S int) int {
	panic("TODO, stubbed")
}

// Sle_CutLeaves (static inline)
// C: int * Sle_CutLeaves( int * pCut )
func sleCutLeaves(pCut []int) []int {
	panic("TODO, stubbed")
}

// Sle_CutIsUsed (static inline)
// C: int Sle_CutIsUsed( int * pCut )
func sleCutIsUsed(pCut []int) int {
	panic("TODO, stubbed")
}

// Sle_CutSetUnused (static inline)
// C: void Sle_CutSetUnused( int * pCut )
func sleCutSetUnused(pCut []int) {
	panic("TODO, stubbed")
}

// Sle_ListCutNum (static inline)
// C: int Sle_ListCutNum( int * pList )
func sleListCutNum(pList []int) int {
	panic("TODO, stubbed")
}

// Sle_CutMergeOrder (static inline)
// C: int Sle_CutMergeOrder( int * pCut0, int * pCut1, int * pCut, int nLutSize )
func sleCutMergeOrder(pCut0 []int, pCut1 []int, pCut []int, nLutSize int) int {
	panic("TODO, stubbed")
}

// Sle_SetCutIsContainedOrder (static inline)
// C: int Sle_SetCutIsContainedOrder( int * pBase, int * pCut )
func sleSetCutIsContainedOrder(pBase []int, pCut []int) int {
	panic("TODO, stubbed")
}

// Sle_CutCountBits (static inline)
// C: int Sle_CutCountBits( unsigned i )
func sleCutCountBits(i uint32) int {
	panic("TODO, stubbed")
}

// Sle_SetLastCutIsContained (static inline)
// C: int Sle_SetLastCutIsContained( Vec_Int_t * vTemp, int * pBase )
func sleSetLastCutIsContained(vTemp *IntVector, pBase []int) int {
	panic("TODO, stubbed")
}

// Sle_SetAddCut (static inline)
// C: void Sle_SetAddCut( Vec_Int_t * vTemp, int * pCut )
func sleSetAddCut(vTemp *IntVector, pCut []int) {
	panic("TODO, stubbed")
}

// Sle_ManCutMerge
// C: int Sle_ManCutMerge( Gia_Man_t * p, int iObj, Vec_Int_t * vCuts, Vec_Int_t * vTemp, int nLutSize )
func (p *GiaManager) SleManCutMerge(iObj int, vCuts *IntVector, vTemp *IntVector, nLutSize int) int {
	panic("TODO, stubbed")
}

// Sle_ManComputeCuts
// C: Vec_Int_t * Sle_ManComputeCuts( Gia_Man_t * p, int nLutSize, int fVerbose )
func (p *GiaManager) SleManComputeCuts(nLutSize int, fVerbose int) *IntVector {
	panic("TODO, stubbed")
}

// Sle_ManComputeDelayCut
// C: int Sle_ManComputeDelayCut( Gia_Man_t * p, int * pCut, Vec_Int_t * vTime )
func (p *GiaManager) SleManComputeDelayCut(pCut []int, vTime *IntVector) int {
	panic("TODO, stubbed")
}

// Sle_ManComputeDelayOne
// C: int Sle_ManComputeDelayOne( Gia_Man_t * p, int iObj, Vec_Int_t * vCuts, Vec_Int_t * vTime )
func (p *GiaManager) SleManComputeDelayOne(iObj int, vCuts *IntVector, vTime *IntVector) int {
	panic("TODO, stubbed")
}

// Sle_ManComputeDelay
// C: int Sle_ManComputeDelay( Gia_Man_t * p, Vec_Int_t * vCuts )
func (p *GiaManager) SleManComputeDelay(vCuts *IntVector) int {
	panic("TODO, stubbed")
}

// Sle_ManPrintCut
// C: void Sle_ManPrintCut( int * pCut )
func SleManPrintCut(pCut []int) {
	panic("TODO, stubbed")
}

// Sle_ManPrintCuts
// C: void Sle_ManPrintCuts( Gia_Man_t * p, Vec_Int_t * vCuts, int iObj )
func (p *GiaManager) SleManPrintCuts(vCuts *IntVector, iObj int) {
	panic("TODO, stubbed")
}

// Sle_ManPrintCutsAll
// C: void Sle_ManPrintCutsAll( Gia_Man_t * p, Vec_Int_t * vCuts )
func (p *GiaManager) SleManPrintCutsAll(vCuts *IntVector) {
	panic("TODO, stubbed")
}

// Sle_ManComputeCutsTest
// C: void Sle_ManComputeCutsTest( Gia_Man_t * p )
func (p *GiaManager) SleManComputeCutsTest() {
	panic("TODO, stubbed")
}

// Sle_ManInternalNodeMask
// C: Vec_Bit_t * Sle_ManInternalNodeMask( Gia_Man_t * pGia )
func (pGia *GiaManager) SleManInternalNodeMask() *BitVector {
	panic("TODO, stubbed")
}

// Sle_ManCutHasPisOnly
// C: int Sle_ManCutHasPisOnly( int * pCut, Vec_Bit_t * vMask )
func SleManCutHasPisOnly(pCut []int, vMask *BitVector) int {
	panic("TODO, stubbed")
}

// Sle_ManCollectCutFaninsOne
// C: void Sle_ManCollectCutFaninsOne( Gia_Man_t * pGia, int iObj, Vec_Int_t * vCuts, Vec_Bit_t * vMask, Vec_Int_t * vCutFanins, Vec_Bit_t * vMap )
func (pGia *GiaManager) SleManCollectCutFaninsOne(iObj int, vCuts *IntVector, vMask *BitVector, vCutFanins *IntVector, vMap *BitVector) {
	panic("TODO, stubbed")
}

// Sle_ManCollectCutFanins
// C: Vec_Wec_t * Sle_ManCollectCutFanins( Gia_Man_t * pGia, Vec_Int_t * vCuts, Vec_Bit_t * vMask )
func (pGia *GiaManager) SleManCollectCutFanins(vCuts *IntVector, vMask *BitVector) *IntLevelVector {
	panic("TODO, stubbed")
}

// Sle_ManList (static inline)
// C: int * Sle_ManList( Sle_Man_t * p, int i )
func (p *SleMan) sleManList(i int) []int {
	panic("TODO, stubbed")
}

// Sle_ManAlloc
// C: Sle_Man_t * Sle_ManAlloc( Gia_Man_t * pGia, int nLevels, int fVerbose )
func (pGia *GiaManager) SleManAlloc(nLevels int, fVerbose int) *SleMan {
	panic("TODO, stubbed")
}

// Sle_ManStop
// C: void Sle_ManStop( Sle_Man_t * p )
func (p *SleMan) SleManStop() {
	panic("TODO, stubbed")
}

// Sle_ManMarkupVariables
// C: void Sle_ManMarkupVariables( Sle_Man_t * p )
func (p *SleMan) SleManMarkupVariables() {
	panic("TODO, stubbed")
}

// Sle_ManCheckContained (static inline)
// C: int Sle_ManCheckContained( int * pCutLeaves, int nCutLeaves, int * pLutFanins, int nLutFanins )
func sleManCheckContained(pCutLeaves []int, nCutLeaves int, pLutFanins []int, nLutFanins int) int {
	panic("TODO, stubbed")
}

// Sle_ManDeriveInit
// C: void Sle_ManDeriveInit( Sle_Man_t * p )
func (p *SleMan) SleManDeriveInit() {
	panic("TODO, stubbed")
}

// Sle_ManDeriveCnf
// C: void Sle_ManDeriveCnf( Sle_Man_t * p, int nBTLimit, int fDynamic )
func (p *SleMan) SleManDeriveCnf(nBTLimit int, fDynamic int) {
	panic("TODO, stubbed")
}

// Sle_ManAddEdgeConstraints
// C: int Sle_ManAddEdgeConstraints( Sle_Man_t * p, int nEdges )
func (p *SleMan) SleManAddEdgeConstraints(nEdges int) int {
	panic("TODO, stubbed")
}

// Sle_ManDeriveResult
// C: void Sle_ManDeriveResult( Sle_Man_t * p, Vec_Int_t * vEdge2, Vec_Int_t * vMapping )
func (p *SleMan) SleManDeriveResult(vEdge2 *IntVector, vMapping *IntVector) {
	panic("TODO, stubbed")
}

// Sle_ManExplore
// C: void Sle_ManExplore( Gia_Man_t * pGia, int nBTLimit, int DelayInit, int fDynamic, int fTwoEdges, int fVerbose )
func (pGia *GiaManager) SleManExplore(nBTLimit int, DelayInit int, fDynamic int, fTwoEdges int, fVerbose int) {
	panic("TODO, stubbed")
}
