package abc

// Hand-checked skeleton of src/aig/gia/giaResub.c.

// GiaResbMan is Gia_ResbMan_t.
type GiaResbMan struct {
	nWords        int
	nLimit        int
	nDivsMax      int
	iChoice       int
	fUseXor       int
	fDebug        int
	fVerbose      int
	fVeryVerbose  int
	vDivs         *PointerVector[uint64] // pointers to the simulation words of the divisors
	vGates        *IntVector
	vUnateLits    [2]*IntVector
	vNotUnateVars [2]*IntVector
	vUnatePairs   [2]*IntVector
	vBinateVars   *IntVector
	vUnateLitsW   [2]*IntVector
	vUnatePairsW  [2]*IntVector
	vSorter       *IntLevelVector
	pSets         [2][]uint64 // word * pSets[2]
	pDivA         []uint64    // word *
	pDivB         []uint64    // word *
	vSims         *WordVector
}

// Gia_ObjCheckMffc_rec
// C: int Gia_ObjCheckMffc_rec( Gia_Man_t * p,Gia_Obj_t * pObj, int Limit, Vec_Int_t * vNodes )
func (p *GiaManager) GiaObjCheckMffc_rec(pObj *GiaObject, Limit int, vNodes *IntVector) int {
	panic("TODO, stubbed")
}

// Gia_ObjCheckMffc (static inline)
// C: int Gia_ObjCheckMffc( Gia_Man_t * p, Gia_Obj_t * pRoot, int Limit, Vec_Int_t * vNodes, Vec_Int_t * vLeaves, Vec_Int_t * vInners )
func (p *GiaManager) giaObjCheckMffc(pRoot *GiaObject, Limit int, vNodes *IntVector, vLeaves *IntVector, vInners *IntVector) int {
	panic("TODO, stubbed")
}

// Gia_ManComputeMffcs
// C: Vec_Wec_t * Gia_ManComputeMffcs( Gia_Man_t * p, int LimitMin, int LimitMax, int SuppMax, int RatioBest )
func (p *GiaManager) GiaManComputeMffcs(LimitMin int, LimitMax int, SuppMax int, RatioBest int) *IntLevelVector {
	panic("TODO, stubbed")
}

// Gia_ManPrintDivStats
// C: void Gia_ManPrintDivStats( Gia_Man_t * p, Vec_Wec_t * vMffcs, Vec_Wec_t * vPivots )
func (p *GiaManager) GiaManPrintDivStats(vMffcs *IntLevelVector, vPivots *IntLevelVector) {
	panic("TODO, stubbed")
}

// Gia_ManAddDivisors
// C: void Gia_ManAddDivisors( Gia_Man_t * p, Vec_Wec_t * vMffcs )
func (p *GiaManager) GiaManAddDivisors(vMffcs *IntLevelVector) {
	panic("TODO, stubbed")
}

// Gia_ManResubTest
// C: void Gia_ManResubTest( Gia_Man_t * p )
func (p *GiaManager) GiaManResubTest() {
	panic("TODO, stubbed")
}

// Gia_ResbAlloc
// C: Gia_ResbMan_t * Gia_ResbAlloc( int nWords )
func GiaResbAlloc(nWords int) *GiaResbMan {
	panic("TODO, stubbed")
}

// Gia_ResbInit
// C: void Gia_ResbInit( Gia_ResbMan_t * p, Vec_Ptr_t * vDivs, int nWords, int nLimit, int nDivsMax, int iChoice, int fUseXor, int fDebug, int fVerbose, int fVeryVerbose )
func (p *GiaResbMan) GiaResbInit(vDivs *PointerVector[uint64], nWords int, nLimit int, nDivsMax int, iChoice int, fUseXor int, fDebug int, fVerbose int, fVeryVerbose int) {
	panic("TODO, stubbed")
}

// Gia_ResbFree
// C: void Gia_ResbFree( Gia_ResbMan_t * p )
func (p *GiaResbMan) GiaResbFree() {
	panic("TODO, stubbed")
}

// Gia_ManResubPrintNode
// C: void Gia_ManResubPrintNode( Vec_Int_t * vRes, int nVars, int Node, int fCompl )
func GiaManResubPrintNode(vRes *IntVector, nVars int, Node int, fCompl int) {
	panic("TODO, stubbed")
}

// Gia_ManResubPrintLit
// C: void Gia_ManResubPrintLit( Vec_Int_t * vRes, int nVars, int iLit )
func GiaManResubPrintLit(vRes *IntVector, nVars int, iLit int) {
	panic("TODO, stubbed")
}

// Gia_ManResubPrint
// C: int Gia_ManResubPrint( Vec_Int_t * vRes, int nVars )
func GiaManResubPrint(vRes *IntVector, nVars int) int {
	panic("TODO, stubbed")
}

// Gia_ManResubVerify
// C: int Gia_ManResubVerify( Gia_ResbMan_t * p, word * pFunc )
func (p *GiaResbMan) GiaManResubVerify(pFunc []uint64) int {
	panic("TODO, stubbed")
}

// Gia_ManConstructFromMap
// C: int Gia_ManConstructFromMap( Gia_Man_t * pNew, Vec_Int_t * vGates, int nVars, Vec_Int_t * vUsed, Vec_Int_t * vCopy, int fHash )
func (pNew *GiaManager) GiaManConstructFromMap(vGates *IntVector, nVars int, vUsed *IntVector, vCopy *IntVector, fHash int) int {
	panic("TODO, stubbed")
}

// Gia_ManConstructFromGates
// C: Gia_Man_t * Gia_ManConstructFromGates( Vec_Wec_t * vFuncs, int nVars )
func GiaManConstructFromGates(vFuncs *IntLevelVector, nVars int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManInsertOrder_rec
// C: void Gia_ManInsertOrder_rec( Gia_Man_t * p, int iObj, Vec_Int_t * vObjs, Vec_Wec_t * vFuncs, Vec_Int_t * vNodes )
func (p *GiaManager) GiaManInsertOrder_rec(iObj int, vObjs *IntVector, vFuncs *IntLevelVector, vNodes *IntVector) {
	panic("TODO, stubbed")
}

// Gia_ManInsertOrder
// C: Vec_Int_t * Gia_ManInsertOrder( Gia_Man_t * p, Vec_Int_t * vObjs, Vec_Wec_t * vFuncs )
func (p *GiaManager) GiaManInsertOrder(vObjs *IntVector, vFuncs *IntLevelVector) *IntVector {
	panic("TODO, stubbed")
}

// Gia_ManInsertFromGates
// C: Gia_Man_t * Gia_ManInsertFromGates( Gia_Man_t * p, Vec_Int_t * vObjs, Vec_Wec_t * vFuncs )
func (p *GiaManager) GiaManInsertFromGates(vObjs *IntVector, vFuncs *IntLevelVector) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManFindFirstCommonLit (static inline)
// C: int Gia_ManFindFirstCommonLit( Vec_Int_t * vArr1, Vec_Int_t * vArr2, int fVerbose )
func giaManFindFirstCommonLit(vArr1 *IntVector, vArr2 *IntVector, fVerbose int) int {
	panic("TODO, stubbed")
}

// Gia_ManFindOneUnateInt
// C: void Gia_ManFindOneUnateInt( word * pOff, word * pOn, Vec_Ptr_t * vDivs, int nWords, Vec_Int_t * vUnateLits, Vec_Int_t * vNotUnateVars )
func GiaManFindOneUnateInt(pOff []uint64, pOn []uint64, vDivs *PointerVector[uint64], nWords int, vUnateLits *IntVector, vNotUnateVars *IntVector) {
	panic("TODO, stubbed")
}

// Gia_ManFindOneUnate
// C: int Gia_ManFindOneUnate( word * pSets[2], Vec_Ptr_t * vDivs, int nWords, Vec_Int_t * vUnateLits[2], Vec_Int_t * vNotUnateVars[2], int fVerbose )
func GiaManFindOneUnate(pSets [][]uint64, vDivs *PointerVector[uint64], nWords int, vUnateLits []*IntVector, vNotUnateVars []*IntVector, fVerbose int) int {
	panic("TODO, stubbed")
}

// Gia_ManDivCover (static inline)
// C: int Gia_ManDivCover( word * pOff, word * pOn, word * pDivA, int ComplA, word * pDivB, int ComplB, int nWords )
func giaManDivCover(pOff []uint64, pOn []uint64, pDivA []uint64, ComplA int, pDivB []uint64, ComplB int, nWords int) int {
	panic("TODO, stubbed")
}

// Gia_ManFindTwoUnateInt
// C: int Gia_ManFindTwoUnateInt( word * pOff, word * pOn, Vec_Ptr_t * vDivs, int nWords, Vec_Int_t * vUnateLits, Vec_Int_t * vUnateLitsW, int * pnPairs )
func GiaManFindTwoUnateInt(pOff []uint64, pOn []uint64, vDivs *PointerVector[uint64], nWords int, vUnateLits *IntVector, vUnateLitsW *IntVector, pnPairs *int) int {
	panic("TODO, stubbed")
}

// Gia_ManFindTwoUnate
// C: int Gia_ManFindTwoUnate( word * pSets[2], Vec_Ptr_t * vDivs, int nWords, Vec_Int_t * vUnateLits[2], Vec_Int_t * vUnateLitsW[2], int fVerbose )
func GiaManFindTwoUnate(pSets [][]uint64, vDivs *PointerVector[uint64], nWords int, vUnateLits []*IntVector, vUnateLitsW []*IntVector, fVerbose int) int {
	panic("TODO, stubbed")
}

// Gia_ManFindXorInt
// C: void Gia_ManFindXorInt( word * pOff, word * pOn, Vec_Int_t * vBinate, Vec_Ptr_t * vDivs, int nWords, Vec_Int_t * vUnatePairs )
func GiaManFindXorInt(pOff []uint64, pOn []uint64, vBinate *IntVector, vDivs *PointerVector[uint64], nWords int, vUnatePairs *IntVector) {
	panic("TODO, stubbed")
}

// Gia_ManFindXor
// C: int Gia_ManFindXor( word * pSets[2], Vec_Ptr_t * vDivs, int nWords, Vec_Int_t * vBinateVars, Vec_Int_t * vUnatePairs[2], int fVerbose )
func GiaManFindXor(pSets [][]uint64, vDivs *PointerVector[uint64], nWords int, vBinateVars *IntVector, vUnatePairs []*IntVector, fVerbose int) int {
	panic("TODO, stubbed")
}

// Gia_ManFindUnatePairsInt
// C: void Gia_ManFindUnatePairsInt( word * pOff, word * pOn, Vec_Int_t * vBinate, Vec_Ptr_t * vDivs, int nWords, Vec_Int_t * vUnatePairs )
func GiaManFindUnatePairsInt(pOff []uint64, pOn []uint64, vBinate *IntVector, vDivs *PointerVector[uint64], nWords int, vUnatePairs *IntVector) {
	panic("TODO, stubbed")
}

// Gia_ManFindUnatePairs
// C: void Gia_ManFindUnatePairs( word * pSets[2], Vec_Ptr_t * vDivs, int nWords, Vec_Int_t * vBinateVars, Vec_Int_t * vUnatePairs[2], int fVerbose )
func GiaManFindUnatePairs(pSets [][]uint64, vDivs *PointerVector[uint64], nWords int, vBinateVars *IntVector, vUnatePairs []*IntVector, fVerbose int) {
	panic("TODO, stubbed")
}

// Gia_ManDeriveDivPair
// C: void Gia_ManDeriveDivPair( int iDiv, Vec_Ptr_t * vDivs, int nWords, word * pRes )
func GiaManDeriveDivPair(iDiv int, vDivs *PointerVector[uint64], nWords int, pRes []uint64) {
	panic("TODO, stubbed")
}

// Gia_ManFindDivGateInt
// C: int Gia_ManFindDivGateInt( word * pOff, word * pOn, Vec_Ptr_t * vDivs, int nWords, Vec_Int_t * vUnateLits, Vec_Int_t * vUnatePairs, Vec_Int_t * vUnateLitsW, Vec_Int_t * vUnatePairsW, word * pDivTemp )
func GiaManFindDivGateInt(pOff []uint64, pOn []uint64, vDivs *PointerVector[uint64], nWords int, vUnateLits *IntVector, vUnatePairs *IntVector, vUnateLitsW *IntVector, vUnatePairsW *IntVector, pDivTemp []uint64) int {
	panic("TODO, stubbed")
}

// Gia_ManFindDivGate
// C: int Gia_ManFindDivGate( word * pSets[2], Vec_Ptr_t * vDivs, int nWords, Vec_Int_t * vUnateLits[2], Vec_Int_t * vUnatePairs[2], Vec_Int_t * vUnateLitsW[2], Vec_Int_t * vUnatePairsW[2], word * pDivTemp )
func GiaManFindDivGate(pSets [][]uint64, vDivs *PointerVector[uint64], nWords int, vUnateLits []*IntVector, vUnatePairs []*IntVector, vUnateLitsW []*IntVector, vUnatePairsW []*IntVector, pDivTemp []uint64) int {
	panic("TODO, stubbed")
}

// Gia_ManFindGateGateInt
// C: int Gia_ManFindGateGateInt( word * pOff, word * pOn, Vec_Ptr_t * vDivs, int nWords, Vec_Int_t * vUnatePairs, Vec_Int_t * vUnatePairsW, word * pDivTempA, word * pDivTempB )
func GiaManFindGateGateInt(pOff []uint64, pOn []uint64, vDivs *PointerVector[uint64], nWords int, vUnatePairs *IntVector, vUnatePairsW *IntVector, pDivTempA []uint64, pDivTempB []uint64) int {
	panic("TODO, stubbed")
}

// Gia_ManFindGateGate
// C: int Gia_ManFindGateGate( word * pSets[2], Vec_Ptr_t * vDivs, int nWords, Vec_Int_t * vUnatePairs[2], Vec_Int_t * vUnatePairsW[2], word * pDivTempA, word * pDivTempB )
func GiaManFindGateGate(pSets [][]uint64, vDivs *PointerVector[uint64], nWords int, vUnatePairs []*IntVector, vUnatePairsW []*IntVector, pDivTempA []uint64, pDivTempB []uint64) int {
	panic("TODO, stubbed")
}

// Gia_ManSortUnatesInt
// C: void Gia_ManSortUnatesInt( word * pOff, word * pOn, Vec_Ptr_t * vDivs, int nWords, Vec_Int_t * vUnateLits, Vec_Int_t * vUnateLitsW, Vec_Wec_t * vSorter )
func GiaManSortUnatesInt(pOff []uint64, pOn []uint64, vDivs *PointerVector[uint64], nWords int, vUnateLits *IntVector, vUnateLitsW *IntVector, vSorter *IntLevelVector) {
	panic("TODO, stubbed")
}

// Gia_ManSortUnates
// C: void Gia_ManSortUnates( word * pSets[2], Vec_Ptr_t * vDivs, int nWords, Vec_Int_t * vUnateLits[2], Vec_Int_t * vUnateLitsW[2], Vec_Wec_t * vSorter )
func GiaManSortUnates(pSets [][]uint64, vDivs *PointerVector[uint64], nWords int, vUnateLits []*IntVector, vUnateLitsW []*IntVector, vSorter *IntLevelVector) {
	panic("TODO, stubbed")
}

// Gia_ManSortPairsInt
// C: void Gia_ManSortPairsInt( word * pOff, word * pOn, Vec_Ptr_t * vDivs, int nWords, Vec_Int_t * vUnatePairs, Vec_Int_t * vUnatePairsW, Vec_Wec_t * vSorter )
func GiaManSortPairsInt(pOff []uint64, pOn []uint64, vDivs *PointerVector[uint64], nWords int, vUnatePairs *IntVector, vUnatePairsW *IntVector, vSorter *IntLevelVector) {
	panic("TODO, stubbed")
}

// Gia_ManSortPairs
// C: void Gia_ManSortPairs( word * pSets[2], Vec_Ptr_t * vDivs, int nWords, Vec_Int_t * vUnateLits[2], Vec_Int_t * vUnateLitsW[2], Vec_Wec_t * vSorter )
func GiaManSortPairs(pSets [][]uint64, vDivs *PointerVector[uint64], nWords int, vUnateLits []*IntVector, vUnateLitsW []*IntVector, vSorter *IntLevelVector) {
	panic("TODO, stubbed")
}

// Gia_ManSortBinate
// C: void Gia_ManSortBinate( word * pSets[2], Vec_Ptr_t * vDivs, int nWords, Vec_Int_t * vBinateVars, Vec_Wec_t * vSorter )
func GiaManSortBinate(pSets [][]uint64, vDivs *PointerVector[uint64], nWords int, vBinateVars *IntVector, vSorter *IntLevelVector) {
	panic("TODO, stubbed")
}

// Gia_ManResubFindBestBinate
// C: int Gia_ManResubFindBestBinate( Gia_ResbMan_t * p )
func (p *GiaResbMan) GiaManResubFindBestBinate() int {
	panic("TODO, stubbed")
}

// Gia_ManResubAddNode
// C: int Gia_ManResubAddNode( Gia_ResbMan_t * p, int iLit0, int iLit1, int Type )
func (p *GiaResbMan) GiaManResubAddNode(iLit0 int, iLit1 int, Type int) int {
	panic("TODO, stubbed")
}

// Gia_ManResubPerformMux_rec
// C: int Gia_ManResubPerformMux_rec( Gia_ResbMan_t * p, int nLimit )
func (p *GiaResbMan) GiaManResubPerformMux_rec(nLimit int) int {
	panic("TODO, stubbed")
}

// Gia_ManResubPerform_rec
// C: int Gia_ManResubPerform_rec( Gia_ResbMan_t * p, int nLimit )
func (p *GiaResbMan) GiaManResubPerform_rec(nLimit int) int {
	panic("TODO, stubbed")
}

// Gia_ManResubPerform
// C: void Gia_ManResubPerform( Gia_ResbMan_t * p, Vec_Ptr_t * vDivs, int nWords, int nLimit, int nDivsMax, int iChoice, int fUseXor, int fDebug, int fVerbose )
func (p *GiaResbMan) GiaManResubPerform(vDivs *PointerVector[uint64], nWords int, nLimit int, nDivsMax int, iChoice int, fUseXor int, fDebug int, fVerbose int) {
	panic("TODO, stubbed")
}

// Gia_ManResubOne
// C: Vec_Int_t * Gia_ManResubOne( Vec_Ptr_t * vDivs, int nWords, int nLimit, int nDivsMax, int iChoice, int fUseXor, int fDebug, int fVerbose, word * pFunc )
func GiaManResubOne(vDivs *PointerVector[uint64], nWords int, nLimit int, nDivsMax int, iChoice int, fUseXor int, fDebug int, fVerbose int, pFunc []uint64) *IntVector {
	panic("TODO, stubbed")
}

// Abc_ResubPrepareManager
// C: void Abc_ResubPrepareManager( int nWords )
func AbcResubPrepareManager(nWords int) {
	panic("TODO, stubbed")
}

// Abc_ResubComputeFunction (ppDivs is a void ** array of pointers to simulation words)
// C: int Abc_ResubComputeFunction( void ** ppDivs, int nDivs, int nWords, int nLimit, int nDivsMax, int iChoice, int fUseXor, int fDebug, int fVerbose, int ** ppArray )
func AbcResubComputeFunction(ppDivs [][]uint64, nDivs int, nWords int, nLimit int, nDivsMax int, iChoice int, fUseXor int, fDebug int, fVerbose int, ppArray *[]int) int {
	panic("TODO, stubbed")
}

// Abc_ResubDumpProblem (ppDivs is a void ** array of pointers to simulation words)
// C: void Abc_ResubDumpProblem( char * pFileName, void ** ppDivs, int nDivs, int nWords )
func AbcResubDumpProblem(pFileName string, ppDivs [][]uint64, nDivs int, nWords int) {
	panic("TODO, stubbed")
}

// Gia_ManResubTest3
// C: void Gia_ManResubTest3()
func GiaManResubTest3() {
	panic("TODO, stubbed")
}

// Gia_ManResubTest3_
// C: void Gia_ManResubTest3_()
func GiaManResubTest3_() {
	panic("TODO, stubbed")
}

// Gia_ManCheckResub
// C: void Gia_ManCheckResub( Vec_Ptr_t * vDivs, int nWords )
func GiaManCheckResub(vDivs *PointerVector[uint64], nWords int) {
	panic("TODO, stubbed")
}

// Gia_ManDeriveDivs
// C: Vec_Ptr_t * Gia_ManDeriveDivs( Vec_Wrd_t * vSims, int nWords )
func GiaManDeriveDivs(vSims *WordVector, nWords int) *PointerVector[uint64] {
	panic("TODO, stubbed")
}

// Gia_ManResub2
// C: Gia_Man_t * Gia_ManResub2( Gia_Man_t * pGia, int nNodes, int nSupp, int nDivs, int iChoice, int fUseXor, int fVerbose, int fVeryVerbose )
func (pGia *GiaManager) GiaManResub2(nNodes int, nSupp int, nDivs int, iChoice int, fUseXor int, fVerbose int, fVeryVerbose int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManResub1
// C: Gia_Man_t * Gia_ManResub1( char * pFileName, int nNodes, int nSupp, int nDivs, int iChoice, int fUseXor, int fVerbose, int fVeryVerbose )
func GiaManResub1(pFileName string, nNodes int, nSupp int, nDivs int, iChoice int, fUseXor int, fVerbose int, fVeryVerbose int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManUnivTfo_rec
// C: int Gia_ManUnivTfo_rec( Gia_Man_t * p, int iObj, Vec_Int_t * vNodes, Vec_Int_t * vPos )
func (p *GiaManager) GiaManUnivTfo_rec(iObj int, vNodes *IntVector, vPos *IntVector) int {
	panic("TODO, stubbed")
}

// Gia_ManUnivTfo
// C: int Gia_ManUnivTfo( Gia_Man_t * p, int * pObjs, int nObjs, Vec_Int_t ** pvNodes, Vec_Int_t ** pvPos )
func (p *GiaManager) GiaManUnivTfo(pObjs []int, nObjs int, pvNodes **IntVector, pvPos **IntVector) int {
	panic("TODO, stubbed")
}

// Gia_ManTryResub
// C: void Gia_ManTryResub( Gia_Man_t * p )
func (p *GiaManager) GiaManTryResub() {
	panic("TODO, stubbed")
}
