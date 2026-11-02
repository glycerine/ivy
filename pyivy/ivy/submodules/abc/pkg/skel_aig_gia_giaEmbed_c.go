package abc

// Hand-checked skeleton of src/aig/gia/giaEmbed.c.

const GIA_PLACE_SIZE = 0x7fff

// EmbDat is Emb_Dat_t.
type EmbDat = float32

// EmbObj is Emb_Obj_t.
type EmbObj struct {
	fCi      uint32 // unsigned : 1 : terminal node CI
	fCo      uint32 // unsigned : 1 : terminal node CO
	fMark0   uint32 // unsigned : 1 : first user-controlled mark
	fMark1   uint32 // unsigned : 1 : second user-controlled mark
	nFanins  uint32 // unsigned : 28 : the number of fanins
	nFanouts uint32 // the number of fanouts
	hHandle  int    // the handle of the node
	// union { unsigned TravId; unsigned iFanin; }
	TravId uint32 // user-specified value
	iFanin uint32
	// union { unsigned Value; unsigned iFanout; }
	Value   uint32 // user-specified value
	iFanout uint32
	Fanios  []int // int Fanios[0] : the array of fanins/fanouts (flexible array member)
}

// EmbMan is Emb_Man_t.
type EmbMan struct {
	pGia       *GiaManager // the original AIG manager
	vCis       *IntVector  // the vector of CIs (PIs + LOs)
	vCos       *IntVector  // the vector of COs (POs + LIs)
	nObjs      int         // the number of objects
	nRegs      int         // the number of registers
	nTravIds   int         // traversal ID of the network
	pObjData   []int       // int * : the array containing data for objects
	nObjData   int         // the size of array to store the logic network
	fVerbose   int         // verbose output flag
	pVecs      []EmbDat    // Emb_Dat_t * : array of vectors of size nObjs * nDims
	nReached   int         // the number of nodes reachable from the pivot
	nDistMax   int         // the maximum distance from the node
	pMatr      [][]float32 // float ** : covariance matrix nDims * nDims
	pEigen     [][]float32 // float ** : the first several eigen values of the matrix
	pSols      []float32   // float * : solutions to the problem nObjs * nSols;
	pPlacement []uint16    // unsigned short * : (x,y) coordinates for each cell
}

// Emb_ManRegNum (static inline)
// C: int Emb_ManRegNum( Emb_Man_t * p )
func (p *EmbMan) embManRegNum() int {
	panic("TODO, stubbed")
}

// Emb_ManCiNum (static inline)
// C: int Emb_ManCiNum( Emb_Man_t * p )
func (p *EmbMan) embManCiNum() int {
	panic("TODO, stubbed")
}

// Emb_ManCoNum (static inline)
// C: int Emb_ManCoNum( Emb_Man_t * p )
func (p *EmbMan) embManCoNum() int {
	panic("TODO, stubbed")
}

// Emb_ManPiNum (static inline)
// C: int Emb_ManPiNum( Emb_Man_t * p )
func (p *EmbMan) embManPiNum() int {
	panic("TODO, stubbed")
}

// Emb_ManPoNum (static inline)
// C: int Emb_ManPoNum( Emb_Man_t * p )
func (p *EmbMan) embManPoNum() int {
	panic("TODO, stubbed")
}

// Emb_ManObjNum (static inline)
// C: int Emb_ManObjNum( Emb_Man_t * p )
func (p *EmbMan) embManObjNum() int {
	panic("TODO, stubbed")
}

// Emb_ManNodeNum (static inline)
// C: int Emb_ManNodeNum( Emb_Man_t * p )
func (p *EmbMan) embManNodeNum() int {
	panic("TODO, stubbed")
}

// Emb_ManObj (static inline)
// C: Emb_Obj_t * Emb_ManObj( Emb_Man_t * p, unsigned hHandle )
func (p *EmbMan) embManObj(hHandle uint32) *EmbObj {
	panic("TODO, stubbed")
}

// Emb_ManCi (static inline)
// C: Emb_Obj_t * Emb_ManCi( Emb_Man_t * p, int i )
func (p *EmbMan) embManCi(i int) *EmbObj {
	panic("TODO, stubbed")
}

// Emb_ManCo (static inline)
// C: Emb_Obj_t * Emb_ManCo( Emb_Man_t * p, int i )
func (p *EmbMan) embManCo(i int) *EmbObj {
	panic("TODO, stubbed")
}

// Emb_ObjIsTerm (static inline)
// C: int Emb_ObjIsTerm( Emb_Obj_t * pObj )
func embObjIsTerm(pObj *EmbObj) int {
	panic("TODO, stubbed")
}

// Emb_ObjIsCi (static inline)
// C: int Emb_ObjIsCi( Emb_Obj_t * pObj )
func embObjIsCi(pObj *EmbObj) int {
	panic("TODO, stubbed")
}

// Emb_ObjIsCo (static inline)
// C: int Emb_ObjIsCo( Emb_Obj_t * pObj )
func embObjIsCo(pObj *EmbObj) int {
	panic("TODO, stubbed")
}

// Emb_ObjIsNode (static inline)
// C: int Emb_ObjIsNode( Emb_Obj_t * pObj )
func embObjIsNode(pObj *EmbObj) int {
	panic("TODO, stubbed")
}

// Emb_ObjSize (static inline)
// C: int Emb_ObjSize( Emb_Obj_t * pObj )
func embObjSize(pObj *EmbObj) int {
	panic("TODO, stubbed")
}

// Emb_ObjFaninNum (static inline)
// C: int Emb_ObjFaninNum( Emb_Obj_t * pObj )
func embObjFaninNum(pObj *EmbObj) int {
	panic("TODO, stubbed")
}

// Emb_ObjFanoutNum (static inline)
// C: int Emb_ObjFanoutNum( Emb_Obj_t * pObj )
func embObjFanoutNum(pObj *EmbObj) int {
	panic("TODO, stubbed")
}

// Emb_ObjFanin (static inline)
// C: Emb_Obj_t * Emb_ObjFanin( Emb_Obj_t * pObj, int i )
func embObjFanin(pObj *EmbObj, i int) *EmbObj {
	panic("TODO, stubbed")
}

// Emb_ObjFanout (static inline)
// C: Emb_Obj_t * Emb_ObjFanout( Emb_Obj_t * pObj, int i )
func embObjFanout(pObj *EmbObj, i int) *EmbObj {
	panic("TODO, stubbed")
}

// Emb_ManResetTravId (static inline)
// C: void Emb_ManResetTravId( Emb_Man_t * p )
func (p *EmbMan) embManResetTravId() {
	panic("TODO, stubbed")
}

// Emb_ManIncrementTravId (static inline)
// C: void Emb_ManIncrementTravId( Emb_Man_t * p )
func (p *EmbMan) embManIncrementTravId() {
	panic("TODO, stubbed")
}

// Emb_ObjSetTravId (static inline)
// C: void Emb_ObjSetTravId( Emb_Obj_t * pObj, int TravId )
func embObjSetTravId(pObj *EmbObj, TravId int) {
	panic("TODO, stubbed")
}

// Emb_ObjSetTravIdCurrent (static inline)
// C: void Emb_ObjSetTravIdCurrent( Emb_Man_t * p, Emb_Obj_t * pObj )
func (p *EmbMan) embObjSetTravIdCurrent(pObj *EmbObj) {
	panic("TODO, stubbed")
}

// Emb_ObjSetTravIdPrevious (static inline)
// C: void Emb_ObjSetTravIdPrevious( Emb_Man_t * p, Emb_Obj_t * pObj )
func (p *EmbMan) embObjSetTravIdPrevious(pObj *EmbObj) {
	panic("TODO, stubbed")
}

// Emb_ObjIsTravIdCurrent (static inline)
// C: int Emb_ObjIsTravIdCurrent( Emb_Man_t * p, Emb_Obj_t * pObj )
func (p *EmbMan) embObjIsTravIdCurrent(pObj *EmbObj) int {
	panic("TODO, stubbed")
}

// Emb_ObjIsTravIdPrevious (static inline)
// C: int Emb_ObjIsTravIdPrevious( Emb_Man_t * p, Emb_Obj_t * pObj )
func (p *EmbMan) embObjIsTravIdPrevious(pObj *EmbObj) int {
	panic("TODO, stubbed")
}

// Emb_ManVec (static inline)
// C: Emb_Dat_t * Emb_ManVec( Emb_Man_t * p, int v )
func (p *EmbMan) embManVec(v int) []EmbDat {
	panic("TODO, stubbed")
}

// Emb_ManSol (static inline)
// C: float * Emb_ManSol( Emb_Man_t * p, int v )
func (p *EmbMan) embManSol(v int) []float32 {
	panic("TODO, stubbed")
}

// Emb_ObjAddFanin
// C: void Emb_ObjAddFanin( Emb_Obj_t * pObj, Emb_Obj_t * pFanin )
func EmbObjAddFanin(pObj *EmbObj, pFanin *EmbObj) {
	panic("TODO, stubbed")
}

// Emb_ManStartSimple
// C: Emb_Man_t * Emb_ManStartSimple( Gia_Man_t * pGia )
func (pGia *GiaManager) EmbManStartSimple() *EmbMan {
	panic("TODO, stubbed")
}

// Emb_ManCollectSuper_rec
// C: void Emb_ManCollectSuper_rec( Gia_Man_t * p, Gia_Obj_t * pObj, Vec_Int_t * vSuper, Vec_Int_t * vVisit )
func (p *GiaManager) EmbManCollectSuper_rec(pObj *GiaObject, vSuper *IntVector, vVisit *IntVector) {
	panic("TODO, stubbed")
}

// Emb_ManCollectSuper
// C: void Emb_ManCollectSuper( Gia_Man_t * p, Gia_Obj_t * pObj, Vec_Int_t * vSuper, Vec_Int_t * vVisit )
func (p *GiaManager) EmbManCollectSuper(pObj *GiaObject, vSuper *IntVector, vVisit *IntVector) {
	panic("TODO, stubbed")
}

// Emb_ManCreateRefsSpecial
// C: void Emb_ManCreateRefsSpecial( Gia_Man_t * p )
func (p *GiaManager) EmbManCreateRefsSpecial() {
	panic("TODO, stubbed")
}

// Emb_ManTransformRefs
// C: void Emb_ManTransformRefs( Gia_Man_t * p, int * pnObjs, int * pnFanios )
func (p *GiaManager) EmbManTransformRefs(pnObjs *int, pnFanios *int) {
	panic("TODO, stubbed")
}

// Emb_ManCleanTravId
// C: void Emb_ManCleanTravId( Emb_Man_t * p )
func (p *EmbMan) EmbManCleanTravId() {
	panic("TODO, stubbed")
}

// Emb_ManSetValue
// C: void Emb_ManSetValue( Emb_Man_t * p )
func (p *EmbMan) EmbManSetValue() {
	panic("TODO, stubbed")
}

// Emb_ManStart
// C: Emb_Man_t * Emb_ManStart( Gia_Man_t * pGia )
func (pGia *GiaManager) EmbManStart() *EmbMan {
	panic("TODO, stubbed")
}

// Emb_ManPrintStats
// C: void Emb_ManPrintStats( Emb_Man_t * p )
func (p *EmbMan) EmbManPrintStats() {
	panic("TODO, stubbed")
}

// Emb_ManStop
// C: void Emb_ManStop( Emb_Man_t * p )
func (p *EmbMan) EmbManStop() {
	panic("TODO, stubbed")
}

// Emb_ManPrintFanio
// C: void Emb_ManPrintFanio( Emb_Man_t * p )
func (p *EmbMan) EmbManPrintFanio() {
	panic("TODO, stubbed")
}

// Emb_ManComputeDistance_old
// C: int Emb_ManComputeDistance_old( Emb_Man_t * p, Emb_Obj_t * pPivot )
func (p *EmbMan) EmbManComputeDistance_old(pPivot *EmbObj) int {
	panic("TODO, stubbed")
}

// Gia_ManTestDistanceInternal
// C: void Gia_ManTestDistanceInternal( Emb_Man_t * p )
func (p *EmbMan) GiaManTestDistanceInternal() {
	panic("TODO, stubbed")
}

// Gia_ManTestDistance
// C: void Gia_ManTestDistance( Gia_Man_t * pGia )
func (pGia *GiaManager) GiaManTestDistance() {
	panic("TODO, stubbed")
}

// Emb_ManPerformBfs
// C: Emb_Obj_t * Emb_ManPerformBfs( Emb_Man_t * p, Vec_Int_t * vThis, Vec_Int_t * vNext, Emb_Dat_t * pDist )
func (p *EmbMan) EmbManPerformBfs(vThis *IntVector, vNext *IntVector, pDist []EmbDat) *EmbObj {
	panic("TODO, stubbed")
}

// Emb_ManConnectedComponents
// C: Vec_Int_t * Emb_ManConnectedComponents( Emb_Man_t * p )
func (p *EmbMan) EmbManConnectedComponents() *IntVector {
	panic("TODO, stubbed")
}

// Emb_ManFindDistances
// C: Emb_Obj_t * Emb_ManFindDistances( Emb_Man_t * p, Vec_Int_t * vStart, Emb_Dat_t * pDist )
func (p *EmbMan) EmbManFindDistances(vStart *IntVector, pDist []EmbDat) *EmbObj {
	panic("TODO, stubbed")
}

// Emb_ManRandomVertex
// C: Emb_Obj_t * Emb_ManRandomVertex( Emb_Man_t * p )
func (p *EmbMan) EmbManRandomVertex() *EmbObj {
	panic("TODO, stubbed")
}

// Emb_DumpGraphIntoFile
// C: void Emb_DumpGraphIntoFile( Emb_Man_t * p )
func (p *EmbMan) EmbDumpGraphIntoFile() {
	panic("TODO, stubbed")
}

// Emb_ManComputeDimensions
// C: void Emb_ManComputeDimensions( Emb_Man_t * p, int nDims )
func (p *EmbMan) EmbManComputeDimensions(nDims int) {
	panic("TODO, stubbed")
}

// Emb_ManMatrAlloc
// C: float ** Emb_ManMatrAlloc( int nDims )
func EmbManMatrAlloc(nDims int) [][]float32 {
	panic("TODO, stubbed")
}

// Emb_ManComputeCovariance
// C: void Emb_ManComputeCovariance( Emb_Man_t * p, int nDims )
func (p *EmbMan) EmbManComputeCovariance(nDims int) {
	panic("TODO, stubbed")
}

// Emb_ManVecRandom
// C: void Emb_ManVecRandom( float * pVec, int nDims )
func EmbManVecRandom(pVec []float32, nDims int) {
	panic("TODO, stubbed")
}

// Emb_ManVecNormal
// C: void Emb_ManVecNormal( float * pVec, int nDims )
func EmbManVecNormal(pVec []float32, nDims int) {
	panic("TODO, stubbed")
}

// Emb_ManVecMultiplyOne
// C: float Emb_ManVecMultiplyOne( float * pVec0, float * pVec1, int nDims )
func EmbManVecMultiplyOne(pVec0 []float32, pVec1 []float32, nDims int) float32 {
	panic("TODO, stubbed")
}

// Emb_ManVecCopyOne
// C: void Emb_ManVecCopyOne( float * pVecDest, float * pVecSour, int nDims )
func EmbManVecCopyOne(pVecDest []float32, pVecSour []float32, nDims int) {
	panic("TODO, stubbed")
}

// Emb_ManVecMultiply
// C: void Emb_ManVecMultiply( float ** pMatr, float * pVec, int nDims, float * pRes )
func EmbManVecMultiply(pMatr [][]float32, pVec []float32, nDims int, pRes []float32) {
	panic("TODO, stubbed")
}

// Emb_ManVecOrthogonolizeOne
// C: void Emb_ManVecOrthogonolizeOne( float * pEigen, float * pVecI, int nDims, float * pVecRes )
func EmbManVecOrthogonolizeOne(pEigen []float32, pVecI []float32, nDims int, pVecRes []float32) {
	panic("TODO, stubbed")
}

// Emb_ManComputeEigenvectors
// C: void Emb_ManComputeEigenvectors( Emb_Man_t * p, int nDims, int nSols )
func (p *EmbMan) EmbManComputeEigenvectors(nDims int, nSols int) {
	panic("TODO, stubbed")
}

// Emb_ManComputeSolutions
// C: void Emb_ManComputeSolutions( Emb_Man_t * p, int nDims, int nSols )
func (p *EmbMan) EmbManComputeSolutions(nDims int, nSols int) {
	panic("TODO, stubbed")
}

// Emb_ManDerivePlacement
// C: void Emb_ManDerivePlacement( Emb_Man_t * p, int nSols )
func (p *EmbMan) EmbManDerivePlacement(nSols int) {
	panic("TODO, stubbed")
}

// Emb_ManComputeHPWL
// C: double Emb_ManComputeHPWL( Emb_Man_t * p )
func (p *EmbMan) EmbManComputeHPWL() float64 {
	panic("TODO, stubbed")
}

// Emb_ManPlacementRefine
// C: void Emb_ManPlacementRefine( Emb_Man_t * p, int nIters, int fVerbose )
func (p *EmbMan) EmbManPlacementRefine(nIters int, fVerbose int) {
	panic("TODO, stubbed")
}

// Emb_ManPrintSolutions
// C: void Emb_ManPrintSolutions( Emb_Man_t * p, int nSols )
func (p *EmbMan) EmbManPrintSolutions(nSols int) {
	panic("TODO, stubbed")
}

// Emb_ManDumpGnuplotPrepare
// C: Vec_Int_t * Emb_ManDumpGnuplotPrepare( Emb_Man_t * p )
func (p *EmbMan) EmbManDumpGnuplotPrepare() *IntVector {
	panic("TODO, stubbed")
}

// Emb_ManDumpGnuplot
// C: void Emb_ManDumpGnuplot( Emb_Man_t * p, char * pName, int fDumpLarge, int fShowImage )
func (p *EmbMan) EmbManDumpGnuplot(pName string, fDumpLarge int, fShowImage int) {
	panic("TODO, stubbed")
}

// Gia_ManSolveProblem
// C: void Gia_ManSolveProblem( Gia_Man_t * pGia, Emb_Par_t * pPars )
func (pGia *GiaManager) GiaManSolveProblem(pPars *EmbPar) {
	panic("TODO, stubbed")
}
