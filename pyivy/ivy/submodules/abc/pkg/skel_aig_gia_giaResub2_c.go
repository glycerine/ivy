package abc

// Hand-checked skeleton of src/aig/gia/giaResub2.c.

// GiaRsb2Man is Gia_Rsb2Man_t.
type GiaRsb2Man struct {
	nDivsMax       int
	nLevelIncrease int
	fUseXor        int
	fUseZeroCost   int
	fDebug         int
	fVerbose       int
	nObjs          int
	nPis           int
	nNodes         int
	nPos           int
	iFirstPo       int
	Level          int
	nMffc          int
	vObjs          IntVector             // Vec_Int_t (by value)
	vSims          WordVector            // Vec_Wrd_t (by value)
	vpDivs         PointerVector[uint64] // Vec_Ptr_t (by value) : pointers to the simulation words of the divisors
	vDivs          IntVector             // Vec_Int_t (by value)
	vLevels        IntVector             // Vec_Int_t (by value)
	vRefs          IntVector             // Vec_Int_t (by value)
	vCopies        IntVector             // Vec_Int_t (by value)
	vTried         IntVector             // Vec_Int_t (by value)
	Truth0         uint64                // word
	Truth1         uint64                // word
	CareSet        uint64                // word
}

// Gia_Rsb2ManAlloc
// C: Gia_Rsb2Man_t * Gia_Rsb2ManAlloc()
func GiaRsb2ManAlloc() *GiaRsb2Man {
	panic("TODO, stubbed")
}

// Gia_Rsb2ManFree
// C: void Gia_Rsb2ManFree( Gia_Rsb2Man_t * p )
func (p *GiaRsb2Man) GiaRsb2ManFree() {
	panic("TODO, stubbed")
}

// Gia_Rsb2ManStart
// C: void Gia_Rsb2ManStart( Gia_Rsb2Man_t * p, int * pObjs, int nObjs, int nDivsMax, int nLevelIncrease, int fUseXor, int fUseZeroCost, int fDebug, int fVerbose )
func (p *GiaRsb2Man) GiaRsb2ManStart(pObjs []int, nObjs int, nDivsMax int, nLevelIncrease int, fUseXor int, fUseZeroCost int, fDebug int, fVerbose int) {
	panic("TODO, stubbed")
}

// Gia_Rsb2ManPrint
// C: void Gia_Rsb2ManPrint( Gia_Rsb2Man_t * p )
func (p *GiaRsb2Man) GiaRsb2ManPrint() {
	panic("TODO, stubbed")
}

// Gia_Rsb2ManLevel
// C: int Gia_Rsb2ManLevel( Gia_Rsb2Man_t * p )
func (p *GiaRsb2Man) GiaRsb2ManLevel() int {
	panic("TODO, stubbed")
}

// Gia_Rsb2ManOdcs
// C: word Gia_Rsb2ManOdcs( Gia_Rsb2Man_t * p, int iNode )
func (p *GiaRsb2Man) GiaRsb2ManOdcs(iNode int) uint64 {
	panic("TODO, stubbed")
}

// Gia_Rsb2ManDeref_rec
// C: int Gia_Rsb2ManDeref_rec( Gia_Rsb2Man_t * p, int * pObjs, int * pRefs, int iNode )
func (p *GiaRsb2Man) GiaRsb2ManDeref_rec(pObjs []int, pRefs []int, iNode int) int {
	panic("TODO, stubbed")
}

// Gia_Rsb2ManMffc
// C: int Gia_Rsb2ManMffc( Gia_Rsb2Man_t * p, int iNode )
func (p *GiaRsb2Man) GiaRsb2ManMffc(iNode int) int {
	panic("TODO, stubbed")
}

// Gia_Rsb2ManDivs
// C: int Gia_Rsb2ManDivs( Gia_Rsb2Man_t * p, int iNode )
func (p *GiaRsb2Man) GiaRsb2ManDivs(iNode int) int {
	panic("TODO, stubbed")
}

// Gia_Rsb2AddNode
// C: int Gia_Rsb2AddNode( Vec_Int_t * vRes, int iLit0, int iLit1, int iRes0, int iRes1 )
func GiaRsb2AddNode(vRes *IntVector, iLit0 int, iLit1 int, iRes0 int, iRes1 int) int {
	panic("TODO, stubbed")
}

// Gia_Rsb2ManInsert_rec
// C: int Gia_Rsb2ManInsert_rec( Vec_Int_t * vRes, int nPis, Vec_Int_t * vObjs, int iNode, Vec_Int_t * vResub, Vec_Int_t * vDivs, Vec_Int_t * vCopies, int iObj )
func GiaRsb2ManInsert_rec(vRes *IntVector, nPis int, vObjs *IntVector, iNode int, vResub *IntVector, vDivs *IntVector, vCopies *IntVector, iObj int) int {
	panic("TODO, stubbed")
}

// Gia_Rsb2ManInsert
// C: Vec_Int_t * Gia_Rsb2ManInsert( int nPis, int nPos, Vec_Int_t * vObjs, int iNode, Vec_Int_t * vResub, Vec_Int_t * vDivs, Vec_Int_t * vCopies )
func GiaRsb2ManInsert(nPis int, nPos int, vObjs *IntVector, iNode int, vResub *IntVector, vDivs *IntVector, vCopies *IntVector) *IntVector {
	panic("TODO, stubbed")
}

// Abc_ResubPrintDivs (ppDivs is a void ** array of pointers to simulation words)
// C: void Abc_ResubPrintDivs( void ** ppDivs, int nDivs )
func AbcResubPrintDivs(ppDivs [][]uint64, nDivs int) {
	panic("TODO, stubbed")
}

// Abc_ResubNodeToTry
// C: int Abc_ResubNodeToTry( Vec_Int_t * vTried, int iFirst, int iLast )
func AbcResubNodeToTry(vTried *IntVector, iFirst int, iLast int) int {
	panic("TODO, stubbed")
}

// Abc_ResubComputeWindow
// C: int Abc_ResubComputeWindow( int * pObjs, int nObjs, int nDivsMax, int nLevelIncrease, int fUseXor, int fUseZeroCost, int fDebug, int fVerbose, int ** ppArray, int * pnResubs )
func AbcResubComputeWindow(pObjs []int, nObjs int, nDivsMax int, nLevelIncrease int, fUseXor int, fUseZeroCost int, fDebug int, fVerbose int, ppArray *[]int, pnResubs *int) int {
	panic("TODO, stubbed")
}

// Abc_ResubComputeWindow2
// C: int Abc_ResubComputeWindow2( int * pObjs, int nObjs, int nDivsMax, int nLevelIncrease, int fUseXor, int fUseZeroCost, int fDebug, int fVerbose, int ** ppArray, int * pnResubs )
func AbcResubComputeWindow2(pObjs []int, nObjs int, nDivsMax int, nLevelIncrease int, fUseXor int, fUseZeroCost int, fDebug int, fVerbose int, ppArray *[]int, pnResubs *int) int {
	panic("TODO, stubbed")
}

// Gia_ManToResub
// C: int * Gia_ManToResub( Gia_Man_t * p )
func (p *GiaManager) GiaManToResub() []int {
	panic("TODO, stubbed")
}

// Gia_ManFromResub
// C: Gia_Man_t * Gia_ManFromResub( int * pObjs, int nObjs, int nIns )
func GiaManFromResub(pObjs []int, nObjs int, nIns int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManResub2Test
// C: Gia_Man_t * Gia_ManResub2Test( Gia_Man_t * p )
func (p *GiaManager) GiaManResub2Test() *GiaManager {
	panic("TODO, stubbed")
}

// Gia_WinTryAddingNode
// C: int Gia_WinTryAddingNode( Gia_Man_t * p, int iPivot, int iPivot2, Vec_Wec_t * vLevels, Vec_Int_t * vNodes )
func (p *GiaManager) GiaWinTryAddingNode(iPivot int, iPivot2 int, vLevels *IntLevelVector, vNodes *IntVector) int {
	panic("TODO, stubbed")
}

// Gia_WinAddCiWithMaxFanouts
// C: int Gia_WinAddCiWithMaxFanouts( Gia_Man_t * p )
func (p *GiaManager) GiaWinAddCiWithMaxFanouts() int {
	panic("TODO, stubbed")
}

// Gia_WinAddCiWithMaxDivisors
// C: int Gia_WinAddCiWithMaxDivisors( Gia_Man_t * p, Vec_Wec_t * vLevels )
func (p *GiaManager) GiaWinAddCiWithMaxDivisors(vLevels *IntLevelVector) int {
	panic("TODO, stubbed")
}

// Gia_WinNodeHasUnmarkedFanouts
// C: int Gia_WinNodeHasUnmarkedFanouts( Gia_Man_t * p, int iPivot )
func (p *GiaManager) GiaWinNodeHasUnmarkedFanouts(iPivot int) int {
	panic("TODO, stubbed")
}

// Gia_RsbCiTranslate
// C: Vec_Int_t * Gia_RsbCiTranslate( Gia_Man_t * p, Vec_Int_t * vObjs, Vec_Int_t * vMap )
func (p *GiaManager) GiaRsbCiTranslate(vObjs *IntVector, vMap *IntVector) *IntVector {
	panic("TODO, stubbed")
}

// Gia_RsbCiWindow
// C: Vec_Int_t * Gia_RsbCiWindow( Gia_Man_t * p, int nPis )
func (p *GiaManager) GiaRsbCiWindow(nPis int) *IntVector {
	panic("TODO, stubbed")
}

// Gia_RsbCiWindowTest
// C: void Gia_RsbCiWindowTest( Gia_Man_t * p )
func (p *GiaManager) GiaRsbCiWindowTest() {
	panic("TODO, stubbed")
}

// Gia_ObjFaninId (static inline)
// C: int Gia_ObjFaninId( Gia_Obj_t * pObj, int iObj, int n )
func giaObjFaninId(pObj *GiaObject, iObj int, n int) int {
	panic("TODO, stubbed")
}

// Gia_ObjTravIsTopTwo (static inline)
// C: int Gia_ObjTravIsTopTwo( Gia_Man_t * p, int iNodeA )
func (p *GiaManager) giaObjTravIsTopTwo(iNodeA int) int {
	panic("TODO, stubbed")
}

// Gia_ObjTravIsSame (static inline)
// C: int Gia_ObjTravIsSame( Gia_Man_t * p, int iNodeA, int iNodeB )
func (p *GiaManager) giaObjTravIsSame(iNodeA int, iNodeB int) int {
	panic("TODO, stubbed")
}

// Gia_ObjTravSetSame (static inline)
// C: void Gia_ObjTravSetSame( Gia_Man_t * p, int iNodeA, int iNodeB )
func (p *GiaManager) giaObjTravSetSame(iNodeA int, iNodeB int) {
	panic("TODO, stubbed")
}

// Gia_RsbWindowGather
// C: void Gia_RsbWindowGather( Gia_Man_t * p, Vec_Int_t * vPaths, int iNode, Vec_Int_t * vVisited )
func (p *GiaManager) GiaRsbWindowGather(vPaths *IntVector, iNode int, vVisited *IntVector) {
	panic("TODO, stubbed")
}

// Gia_RsbWindowExplore
// C: int Gia_RsbWindowExplore( Gia_Man_t * p, Vec_Int_t * vVisited, int iStart, Vec_Int_t * vPaths, int * piMeet, int * piNode )
func (p *GiaManager) GiaRsbWindowExplore(vVisited *IntVector, iStart int, vPaths *IntVector, piMeet *int, piNode *int) int {
	panic("TODO, stubbed")
}

// Gia_RsbWindowInit
// C: Vec_Int_t * Gia_RsbWindowInit( Gia_Man_t * p, Vec_Int_t * vPaths, int iPivot, int nIter )
func (p *GiaManager) GiaRsbWindowInit(vPaths *IntVector, iPivot int, nIter int) *IntVector {
	panic("TODO, stubbed")
}

// Gia_RsbCreateWindowInputs
// C: Vec_Int_t * Gia_RsbCreateWindowInputs( Gia_Man_t * p, Vec_Int_t * vWin )
func (p *GiaManager) GiaRsbCreateWindowInputs(vWin *IntVector) *IntVector {
	panic("TODO, stubbed")
}

// Gia_RsbAddSideInputs
// C: void Gia_RsbAddSideInputs( Gia_Man_t * p, Vec_Wec_t * vLevels, Vec_Int_t * vWin )
func (p *GiaManager) GiaRsbAddSideInputs(vLevels *IntLevelVector, vWin *IntVector) {
	panic("TODO, stubbed")
}

// Gia_RsbExpandInputs
// C: void Gia_RsbExpandInputs( Gia_Man_t * p, Vec_Wec_t * vLevels, Vec_Int_t * vWin, Vec_Int_t * vInputs )
func (p *GiaManager) GiaRsbExpandInputs(vLevels *IntLevelVector, vWin *IntVector, vInputs *IntVector) {
	panic("TODO, stubbed")
}

// Gia_RsbSelectOneInput
// C: int Gia_RsbSelectOneInput( Gia_Man_t * p, Vec_Wec_t * vLevels, Vec_Int_t * vIns )
func (p *GiaManager) GiaRsbSelectOneInput(vLevels *IntLevelVector, vIns *IntVector) int {
	panic("TODO, stubbed")
}

// Gia_RsbWindowGrow
// C: void Gia_RsbWindowGrow( Gia_Man_t * p, Vec_Wec_t * vLevels, Vec_Int_t * vWin, Vec_Int_t * vIns, int nInputsMax )
func (p *GiaManager) GiaRsbWindowGrow(vLevels *IntLevelVector, vWin *IntVector, vIns *IntVector, nInputsMax int) {
	panic("TODO, stubbed")
}

// Gia_WinCreateFromCut_rec
// C: void Gia_WinCreateFromCut_rec( Gia_Man_t * p, int iObj, Vec_Int_t * vWin )
func (p *GiaManager) GiaWinCreateFromCut_rec(iObj int, vWin *IntVector) {
	panic("TODO, stubbed")
}

// Gia_WinCreateFromCut
// C: void Gia_WinCreateFromCut( Gia_Man_t * p, int iPivot, Vec_Int_t * vIn, Vec_Wec_t * vLevels, Vec_Int_t * vWin )
func (p *GiaManager) GiaWinCreateFromCut(iPivot int, vIn *IntVector, vLevels *IntLevelVector, vWin *IntVector) {
	panic("TODO, stubbed")
}

// Gia_RsbExpandCut
// C: int Gia_RsbExpandCut( Gia_Man_t * p, Vec_Int_t * vIns )
func (p *GiaManager) GiaRsbExpandCut(vIns *IntVector) int {
	panic("TODO, stubbed")
}

// Gia_RsbFindFaninAdd
// C: int Gia_RsbFindFaninAdd( int iFan, int pFanins[32], int pFaninCounts[32], int nFanins )
func GiaRsbFindFaninAdd(iFan int, pFanins []int, pFaninCounts []int, nFanins int) int {
	panic("TODO, stubbed")
}

// Gia_RsbFindFaninToAddToCut
// C: int Gia_RsbFindFaninToAddToCut( Gia_Man_t * p, Vec_Int_t * vIns )
func (p *GiaManager) GiaRsbFindFaninToAddToCut(vIns *IntVector) int {
	panic("TODO, stubbed")
}

// Gia_RsbWindowGrow2
// C: void Gia_RsbWindowGrow2( Gia_Man_t * p, int iObj, Vec_Wec_t * vLevels, Vec_Int_t * vWin, Vec_Int_t * vIns, int nInputsMax )
func (p *GiaManager) GiaRsbWindowGrow2(iObj int, vLevels *IntLevelVector, vWin *IntVector, vIns *IntVector, nInputsMax int) {
	panic("TODO, stubbed")
}

// Gia_RsbWindowCompute
// C: int Gia_RsbWindowCompute( Gia_Man_t * p, int iObj, int nInputsMax, int nLevelsMax, Vec_Wec_t * vLevels, Vec_Int_t * vPaths, Vec_Int_t ** pvWin, Vec_Int_t ** pvIns )
func (p *GiaManager) GiaRsbWindowCompute(iObj int, nInputsMax int, nLevelsMax int, vLevels *IntLevelVector, vPaths *IntVector, pvWin **IntVector, pvIns **IntVector) int {
	panic("TODO, stubbed")
}

// Gia_RsbFindOutputs
// C: Vec_Int_t * Gia_RsbFindOutputs( Gia_Man_t * p, Vec_Int_t * vWin, Vec_Int_t * vIns, Vec_Int_t * vRefs )
func (p *GiaManager) GiaRsbFindOutputs(vWin *IntVector, vIns *IntVector, vRefs *IntVector) *IntVector {
	panic("TODO, stubbed")
}

// Gia_RsbDeriveGiaFromWindows
// C: Gia_Man_t * Gia_RsbDeriveGiaFromWindows( Gia_Man_t * p, Vec_Int_t * vWin, Vec_Int_t * vIns, Vec_Int_t * vOuts )
func (p *GiaManager) GiaRsbDeriveGiaFromWindows(vWin *IntVector, vIns *IntVector, vOuts *IntVector) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_LutComputeTruth66_rec
// C: word Gia_LutComputeTruth66_rec( Gia_Man_t * p, Gia_Obj_t * pObj )
func (p *GiaManager) GiaLutComputeTruth66_rec(pObj *GiaObject) uint64 {
	panic("TODO, stubbed")
}

// Gia_ManVerifyTwoTruths
// C: int Gia_ManVerifyTwoTruths( Gia_Man_t * p1, Gia_Man_t * p2 )
func (p1 *GiaManager) GiaManVerifyTwoTruths(p2 *GiaManager) int {
	panic("TODO, stubbed")
}

// Gia_RsbEnumerateWindows
// C: void Gia_RsbEnumerateWindows( Gia_Man_t * p, int nInputsMax, int nLevelsMax )
func (p *GiaManager) GiaRsbEnumerateWindows(nInputsMax int, nLevelsMax int) {
	panic("TODO, stubbed")
}

// Gia_RsbTryOneWindow
// C: Gia_Man_t * Gia_RsbTryOneWindow( Gia_Man_t * p )
func (p *GiaManager) GiaRsbTryOneWindow() *GiaManager {
	panic("TODO, stubbed")
}

// Gia_RsbTestArray
// C: void Gia_RsbTestArray()
func GiaRsbTestArray() {
	panic("TODO, stubbed")
}
