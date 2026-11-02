package abc

// Hand-checked skeleton of src/opt/sbd/sbdCore.c.

// SBD_MAX_LUTSIZE is 6.

// SbdMan is Sbd_Man_t.
type SbdMan struct {
	pPars     *SbdPar         // user's parameters
	pGia      *GiaManager     // user's AIG manager (will be modified by adding nodes)
	vTfos     *IntLevelVector // TFO for each node (roots are marked) (windowing)
	vLutLevs  *IntVector      // LUT level for each node after resynthesis
	vLutCuts  *IntVector      // LUT cut for each nodes after resynthesis
	vLutCuts2 *IntVector      // LUT cut for each nodes after resynthesis
	vMirrors  *IntVector      // alternative node
	vSims     [4]*WordVector  // simulation information (main, backup, controlability)
	vCover    *IntVector      // temporary
	vLits     *IntVector      // temporary
	vLits2    *IntVector      // temporary
	nLuts     [6]int          // 0=const, 1=1lut, 2=2lut, 3=3lut
	nTried    int
	nUsed     int
	timeWin   int64 // abctime
	timeCut   int64
	timeCov   int64
	timeCnf   int64
	timeSat   int64
	timeQbf   int64
	timeNew   int64
	timeOther int64
	timeTotal int64
	pSto      *SbdSto
	pSrv      *SbdSrv
	// target node
	Pivot      int             // target node
	DivCutoff  int             // the place where D-2 divisors begin
	vTfo       *IntVector      // TFO (excludes node, includes roots) - precomputed
	vRoots     *IntVector      // TFO root nodes
	vWinObjs   *IntVector      // TFI + Pivot + sideTFI + TFO (including roots)
	vObj2Var   *IntVector      // SAT variables for the window (indexes of objects in vWinObjs)
	vDivSet    *IntVector      // divisor variables
	vDivVars   *IntVector      // divisor variables
	vDivValues *IntVector      // SAT variables values for the divisor variables
	vDivLevels *IntLevelVector // divisors collected by levels
	vCounts    [2]*IntVector   // counters of zeros and ones
	vMatrix    *WordVector     // covering matrix
	pSat       *SatSolver      // SAT solver
}

// Sbd_ObjCut (static; a pointer into vLutCuts)
// C: int * Sbd_ObjCut( Sbd_Man_t * p, int i )
func (p *SbdMan) sbdObjCut(i int) []int {
	panic("TODO, stubbed")
}

// Sbd_ObjCut2 (static; a pointer into vLutCuts2)
// C: int * Sbd_ObjCut2( Sbd_Man_t * p, int i )
func (p *SbdMan) sbdObjCut2(i int) []int {
	panic("TODO, stubbed")
}

// Sbd_ObjSim0 (static; a pointer into vSims[0])
// C: word * Sbd_ObjSim0( Sbd_Man_t * p, int i )
func (p *SbdMan) sbdObjSim0(i int) []uint64 {
	panic("TODO, stubbed")
}

// Sbd_ObjSim1 (static; a pointer into vSims[1])
// C: word * Sbd_ObjSim1( Sbd_Man_t * p, int i )
func (p *SbdMan) sbdObjSim1(i int) []uint64 {
	panic("TODO, stubbed")
}

// Sbd_ObjSim2 (static; a pointer into vSims[2])
// C: word * Sbd_ObjSim2( Sbd_Man_t * p, int i )
func (p *SbdMan) sbdObjSim2(i int) []uint64 {
	panic("TODO, stubbed")
}

// Sbd_ObjSim3 (static; a pointer into vSims[3])
// C: word * Sbd_ObjSim3( Sbd_Man_t * p, int i )
func (p *SbdMan) sbdObjSim3(i int) []uint64 {
	panic("TODO, stubbed")
}

// Sbd_ParSetDefault
// C: void Sbd_ParSetDefault( Sbd_Par_t * pPars )
func (pPars *SbdPar) SbdParSetDefault() {
	panic("TODO, stubbed")
}

// Sbd_ManWindowRoots
// C: Vec_Wec_t * Sbd_ManWindowRoots( Gia_Man_t * p, int nTfoLevels, int nTfoFanMax )
func (p *GiaManager) SbdManWindowRoots(nTfoLevels int, nTfoFanMax int) *IntLevelVector {
	panic("TODO, stubbed")
}

// Sbd_ManStart
// C: Sbd_Man_t * Sbd_ManStart( Gia_Man_t * pGia, Sbd_Par_t * pPars )
func (pGia *GiaManager) SbdManStart(pPars *SbdPar) *SbdMan {
	panic("TODO, stubbed")
}

// Sbd_ManStop
// C: void Sbd_ManStop( Sbd_Man_t * p )
func (p *SbdMan) SbdManStop() {
	panic("TODO, stubbed")
}

// Sbd_ManPropagateControlOne
// C: void Sbd_ManPropagateControlOne( Sbd_Man_t * p, int Node )
func (p *SbdMan) SbdManPropagateControlOne(Node int) {
	panic("TODO, stubbed")
}

// Sbd_ManPropagateControl
// C: void Sbd_ManPropagateControl( Sbd_Man_t * p, int Pivot )
func (p *SbdMan) SbdManPropagateControl(Pivot int) {
	panic("TODO, stubbed")
}

// Sbd_ManUpdateOrder
// C: void Sbd_ManUpdateOrder( Sbd_Man_t * p, int Pivot )
func (p *SbdMan) SbdManUpdateOrder(Pivot int) {
	panic("TODO, stubbed")
}

// Sbd_ManWindowSim_rec
// C: void Sbd_ManWindowSim_rec( Sbd_Man_t * p, int NodeInit )
func (p *SbdMan) SbdManWindowSim_rec(NodeInit int) {
	panic("TODO, stubbed")
}

// Sbd_ManWindow
// C: int Sbd_ManWindow( Sbd_Man_t * p, int Pivot )
func (p *SbdMan) SbdManWindow(Pivot int) int {
	panic("TODO, stubbed")
}

// Sbd_ManCheckConst
// C: int Sbd_ManCheckConst( Sbd_Man_t * p, int Pivot )
func (p *SbdMan) SbdManCheckConst(Pivot int) int {
	panic("TODO, stubbed")
}

// Sbd_TransposeMatrix64 (static; transposed in place)
// C: void Sbd_TransposeMatrix64( word A[64] )
func sbdTransposeMatrix64(A *[64]uint64) {
	panic("TODO, stubbed")
}

// Sbd_PrintMatrix64 (static)
// C: void Sbd_PrintMatrix64( word A[64] )
func sbdPrintMatrix64(A *[64]uint64) {
	panic("TODO, stubbed")
}

// Sbd_ManPrintObj
// C: void Sbd_ManPrintObj( Sbd_Man_t * p, int Pivot )
func (p *SbdMan) SbdManPrintObj(Pivot int) {
	panic("TODO, stubbed")
}

// Sbd_ManMatrPrint
// C: void Sbd_ManMatrPrint( Sbd_Man_t * p, word Cover[], int nCol, int nRows )
func (p *SbdMan) SbdManMatrPrint(Cover []uint64, nCol int, nRows int) {
	panic("TODO, stubbed")
}

// Sbd_ManCoverReverseOrder (static; reversed in place)
// C: void Sbd_ManCoverReverseOrder( word Cover[64] )
func sbdManCoverReverseOrder(Cover *[64]uint64) {
	panic("TODO, stubbed")
}

// Sbd_ManAddCube1 (static)
// C: int Sbd_ManAddCube1( int nRowLimit, word Cover[], int nRows, word Cube )
func sbdManAddCube1(nRowLimit int, Cover []uint64, nRows int, Cube uint64) int {
	panic("TODO, stubbed")
}

// Sbd_ManAddCube2 (static)
// C: int Sbd_ManAddCube2( word Cover[2][64], int nRows, word Cube[2] )
func sbdManAddCube2(Cover *[2][64]uint64, nRows int, Cube *[2]uint64) int {
	panic("TODO, stubbed")
}

// Sbd_ManFindCandsSimple (static)
// C: int Sbd_ManFindCandsSimple( Sbd_Man_t * p, word Cover[], int nDivs )
func (p *SbdMan) sbdManFindCandsSimple(Cover []uint64, nDivs int) int {
	panic("TODO, stubbed")
}

// Sbd_ManFindCands (static)
// C: int Sbd_ManFindCands( Sbd_Man_t * p, word Cover[64], int nDivs )
func (p *SbdMan) sbdManFindCands(Cover *[64]uint64, nDivs int) int {
	panic("TODO, stubbed")
}

// Sbd_ManExplore (pTruth is an out parameter)
// C: int Sbd_ManExplore( Sbd_Man_t * p, int Pivot, word * pTruth )
func (p *SbdMan) SbdManExplore(Pivot int, pTruth *uint64) int {
	panic("TODO, stubbed")
}

// Sbd_ManExplore2 (pTruth is an out parameter)
// C: int Sbd_ManExplore2( Sbd_Man_t * p, int Pivot, word * pTruth )
func (p *SbdMan) SbdManExplore2(Pivot int, pTruth *uint64) int {
	panic("TODO, stubbed")
}

// Sbd_ManExploreCut (pnStrs and pFreeVar are out parameters; Strs is filled in)
// C: int Sbd_ManExploreCut( Sbd_Man_t * p, int Pivot, int nLeaves, int * pLeaves, int * pnStrs, Sbd_Str_t * Strs, int * pFreeVar )
func (p *SbdMan) SbdManExploreCut(Pivot int, nLeaves int, pLeaves []int, pnStrs *int, Strs []SbdStr, pFreeVar *int) int {
	panic("TODO, stubbed")
}

// Sbd_ManExplore3 (pnStrs is an out parameter; Strs is filled in)
// C: int Sbd_ManExplore3( Sbd_Man_t * p, int Pivot, int * pnStrs, Sbd_Str_t * Strs )
func (p *SbdMan) SbdManExplore3(Pivot int, pnStrs *int, Strs []SbdStr) int {
	panic("TODO, stubbed")
}

// Sbd_CutMergeSimple
// C: int Sbd_CutMergeSimple( Sbd_Man_t * p, int * pCut1, int * pCut2, int * pCut )
func (p *SbdMan) SbdCutMergeSimple(pCut1 []int, pCut2 []int, pCut []int) int {
	panic("TODO, stubbed")
}

// Sbd_ManMergeCuts
// C: int Sbd_ManMergeCuts( Sbd_Man_t * p, int Node )
func (p *SbdMan) SbdManMergeCuts(Node int) int {
	panic("TODO, stubbed")
}

// Sbd_ManDelay
// C: int Sbd_ManDelay( Sbd_Man_t * p )
func (p *SbdMan) SbdManDelay() int {
	panic("TODO, stubbed")
}

// Sbd_ManMergeTest
// C: void Sbd_ManMergeTest( Sbd_Man_t * p )
func (p *SbdMan) SbdManMergeTest() {
	panic("TODO, stubbed")
}

// Sbd_ManFindCut_rec
// C: void Sbd_ManFindCut_rec( Gia_Man_t * p, Gia_Obj_t * pObj )
func (p *GiaManager) SbdManFindCut_rec(pObj *GiaObject) {
	panic("TODO, stubbed")
}

// Sbd_ManFindCutUnmark_rec
// C: void Sbd_ManFindCutUnmark_rec( Gia_Man_t * p, Gia_Obj_t * pObj )
func (p *GiaManager) SbdManFindCutUnmark_rec(pObj *GiaObject) {
	panic("TODO, stubbed")
}

// Sbd_ManFindCut
// C: void Sbd_ManFindCut( Sbd_Man_t * p, int Node, Vec_Int_t * vCutLits )
func (p *SbdMan) SbdManFindCut(Node int, vCutLits *IntVector) {
	panic("TODO, stubbed")
}

// Sbd_ManImplement
// C: int Sbd_ManImplement( Sbd_Man_t * p, int Pivot, word Truth )
func (p *SbdMan) SbdManImplement(Pivot int, Truth uint64) int {
	panic("TODO, stubbed")
}

// Sbd_ManImplement2
// C: int Sbd_ManImplement2( Sbd_Man_t * p, int Pivot, int nStrs, Sbd_Str_t * pStrs )
func (p *SbdMan) SbdManImplement2(Pivot int, nStrs int, pStrs []SbdStr) int {
	panic("TODO, stubbed")
}

// Sbd_ManDeriveMapping_rec
// C: void Sbd_ManDeriveMapping_rec( Sbd_Man_t * p, Gia_Man_t * pNew, int iObj )
func (p *SbdMan) SbdManDeriveMapping_rec(pNew *GiaManager, iObj int) {
	panic("TODO, stubbed")
}

// Sbd_ManDeriveMapping
// C: void Sbd_ManDeriveMapping( Sbd_Man_t * p, Gia_Man_t * pNew )
func (p *SbdMan) SbdManDeriveMapping(pNew *GiaManager) {
	panic("TODO, stubbed")
}

// Sbd_ManDerive_rec
// C: void Sbd_ManDerive_rec( Gia_Man_t * pNew, Gia_Man_t * p, int Node, Vec_Int_t * vMirrors )
func (pNew *GiaManager) SbdManDerive_rec(p *GiaManager, Node int, vMirrors *IntVector) {
	panic("TODO, stubbed")
}

// Sbd_ManDerive
// C: Gia_Man_t * Sbd_ManDerive( Sbd_Man_t * pMan, Gia_Man_t * p, Vec_Int_t * vMirrors )
func (pMan *SbdMan) SbdManDerive(p *GiaManager, vMirrors *IntVector) *GiaManager {
	panic("TODO, stubbed")
}

// Sbd_NtkPerformOne
// C: void Sbd_NtkPerformOne( Sbd_Man_t * p, int Pivot )
func (p *SbdMan) SbdNtkPerformOne(Pivot int) {
	panic("TODO, stubbed")
}

// Sbd_NtkPerform
// C: Gia_Man_t * Sbd_NtkPerform( Gia_Man_t * pGia, Sbd_Par_t * pPars )
func (pGia *GiaManager) SbdNtkPerform(pPars *SbdPar) *GiaManager {
	panic("TODO, stubbed")
}
