package abc

// Hand-checked skeleton of src/base/abci/abcRec3.c.

const (
	LMS_VAR_MAX  = 16 // LMS_VAR_MAX >= 6
	LMS_MAX_WORD = 1 << (LMS_VAR_MAX - 6)
)

// LmsMan is Lms_Man_t.
type LmsMan struct {
	nVars          int                       // the number of variables
	nWords         int                       // the number of TT words
	nCuts          int                       // the max number of cuts to use
	fFuncOnly      int                       // record only functions
	fLibConstr     int                       // this manager is used for library construction
	pGia           *GiaManager               // the record
	vTtMem         *WordMemory               // truth table memory and hash table
	vTruthIds      *IntVector                // truth table IDs of each PO
	vTruthPo       *IntVector                // first PO where this canonicized truth table was seen
	vDelays        *WordVector               // pin-to-pin delays of each PO
	vAreas         *ByteVector               // number of AND gates in each PO
	vFreqs         *IntVector                // subgraph usage frequencies
	vTruthFreqs    *IntVector                // truth table usage frequencies
	vNodes         *PointerVector[IfObj]     // the temporary nodes (If_Obj_t * entries, see Abc_RecToGia3)
	vLabelsP       *PointerVector[HopObject] // temporary storage for HOP node labels
	vLabels        *IntVector                // temporary storage for AIG node labels
	vSupps         *ByteVector               // used temporarily by TT dumping
	pTemp1         [LMS_MAX_WORD]uint64      // word : copy of the truth table
	pTemp2         [LMS_MAX_WORD]uint64      // word : copy of the truth table
	nTried         int
	nFilterSize    int
	nFilterRedund  int
	nFilterVolume  int
	nFilterTruth   int
	nFilterError   int
	nFilterSame    int
	nAdded         int
	nAddedFuncs    int
	nHoleInTheWall int
	timeTruth      int64 // abctime
	timeCanon      int64 // abctime
	timeBuild      int64 // abctime
	timeCheck      int64 // abctime
	timeInsert     int64 // abctime
	timeOther      int64 // abctime
	timeTotal      int64 // abctime
}

// Lms_DelayGet (static)
// C: int Lms_DelayGet( word D, int v )
func lmsDelayGet(D uint64, v int) int {
	panic("TODO, stubbed")
}

// Lms_DelaySet (static)
// C: void Lms_DelaySet( word * pD, int v, int d )
func lmsDelaySet(pD *uint64, v int, d int) {
	panic("TODO, stubbed")
}

// Lms_DelayInit (static)
// C: word Lms_DelayInit( int v )
func lmsDelayInit(v int) uint64 {
	panic("TODO, stubbed")
}

// Lms_DelayMax (static)
// C: word Lms_DelayMax( word D1, word D2, int nVars )
func lmsDelayMax(D1 uint64, D2 uint64, nVars int) uint64 {
	panic("TODO, stubbed")
}

// Lms_DelayDecrement (static)
// C: word Lms_DelayDecrement( word D1, int nVars )
func lmsDelayDecrement(D1 uint64, nVars int) uint64 {
	panic("TODO, stubbed")
}

// Lms_DelayEqual (static)
// C: int Lms_DelayEqual( word D1, word D2, int nVars )
func lmsDelayEqual(D1 uint64, D2 uint64, nVars int) int {
	panic("TODO, stubbed")
}

// Lms_DelayDom (static)
// C: int Lms_DelayDom( word D1, word D2, int nVars )
func lmsDelayDom(D1 uint64, D2 uint64, nVars int) int {
	panic("TODO, stubbed")
}

// Lms_DelayPrint (static)
// C: void Lms_DelayPrint( word D, int nVars )
func lmsDelayPrint(D uint64, nVars int) {
	panic("TODO, stubbed")
}

// Lms_GiaDelays
// C: Vec_Wrd_t * Lms_GiaDelays( Gia_Man_t * p )
func (p *GiaManager) LmsGiaDelays() *WordVector {
	panic("TODO, stubbed")
}

// Lms_ObjAreaMark_rec
// C: void Lms_ObjAreaMark_rec( Gia_Obj_t * pObj )
func LmsObjAreaMark_rec(pObj *GiaObject) {
	panic("TODO, stubbed")
}

// Lms_ObjAreaUnmark_rec
// C: int Lms_ObjAreaUnmark_rec( Gia_Obj_t * pObj )
func LmsObjAreaUnmark_rec(pObj *GiaObject) int {
	panic("TODO, stubbed")
}

// Lms_ObjArea
// C: int Lms_ObjArea( Gia_Obj_t * pObj )
func LmsObjArea(pObj *GiaObject) int {
	panic("TODO, stubbed")
}

// Lms_GiaAreas
// C: Vec_Str_t * Lms_GiaAreas( Gia_Man_t * p )
func (p *GiaManager) LmsGiaAreas() *ByteVector {
	panic("TODO, stubbed")
}

// Lms_GiaSuppSizes
// C: Vec_Str_t * Lms_GiaSuppSizes( Gia_Man_t * p )
func (p *GiaManager) LmsGiaSuppSizes() *ByteVector {
	panic("TODO, stubbed")
}

// Lms_GiaProfilesPrint
// C: void Lms_GiaProfilesPrint( Gia_Man_t * p )
func (p *GiaManager) LmsGiaProfilesPrint() {
	panic("TODO, stubbed")
}

// Lms_GiaPrintSubgraph_rec
// C: void Lms_GiaPrintSubgraph_rec( Gia_Man_t * p, Gia_Obj_t * pObj )
func (p *GiaManager) LmsGiaPrintSubgraph_rec(pObj *GiaObject) {
	panic("TODO, stubbed")
}

// Lms_GiaPrintSubgraph
// C: void Lms_GiaPrintSubgraph( Gia_Man_t * p, Gia_Obj_t * pObj )
func (p *GiaManager) LmsGiaPrintSubgraph(pObj *GiaObject) {
	panic("TODO, stubbed")
}

// Lms_ManStart
// C: Lms_Man_t * Lms_ManStart( Gia_Man_t * pGia, int nVars, int nCuts, int fFuncOnly, int fVerbose )
func (pGia *GiaManager) LmsManStart(nVars int, nCuts int, fFuncOnly int, fVerbose int) *LmsMan {
	panic("TODO, stubbed")
}

// Lms_ManStop
// C: void Lms_ManStop( Lms_Man_t * p )
func (p *LmsMan) LmsManStop() {
	panic("TODO, stubbed")
}

// Lms_ManPrepare
// C: void Lms_ManPrepare( Lms_Man_t * p )
func (p *LmsMan) LmsManPrepare() {
	panic("TODO, stubbed")
}

// Lms_ManPrintFuncStats
// C: void Lms_ManPrintFuncStats( Lms_Man_t * p )
func (p *LmsMan) LmsManPrintFuncStats() {
	panic("TODO, stubbed")
}

// Lms_ManPrintFreqStats
// C: void Lms_ManPrintFreqStats( Lms_Man_t * p )
func (p *LmsMan) LmsManPrintFreqStats() {
	panic("TODO, stubbed")
}

// Lms_ManPrint
// C: void Lms_ManPrint( Lms_Man_t * p )
func (p *LmsMan) LmsManPrint() {
	panic("TODO, stubbed")
}

// Abc_NtkRecLibMerge3
// C: void Abc_NtkRecLibMerge3( Gia_Man_t * pLib )
func (pLib *GiaManager) AbcNtkRecLibMerge3() {
	panic("TODO, stubbed")
}

// Abc_NtkRecAddCut3
// C: int Abc_NtkRecAddCut3( If_Man_t * pIfMan, If_Obj_t * pRoot, If_Cut_t * pCut )
func AbcNtkRecAddCut3(pIfMan *IfMan, pRoot *IfObj, pCut *IfCut) int {
	panic("TODO, stubbed")
}

// Abc_NtkRecAdd3
// C: void Abc_NtkRecAdd3( Abc_Ntk_t * pNtk, int fUseSOPB )
func (pNtk *ABCNetwork) AbcNtkRecAdd3(fUseSOPB int) {
	panic("TODO, stubbed")
}

// If_CutComputeDelay (static)
// C: int If_CutComputeDelay( If_Man_t * p, If_Cut_t * pCut, char * pCanonPerm, word DelayProfile )
func (p *IfMan) ifCutComputeDelay(pCut *IfCut, pCanonPerm []int8, DelayProfile uint64) int {
	panic("TODO, stubbed")
}

// If_CutFindBestStruct (static)
// C: int If_CutFindBestStruct( If_Man_t * pIfMan, If_Cut_t * pCut, char * pCanonPerm, unsigned * puCanonPhase, int * pBestPo )
func (pIfMan *IfMan) ifCutFindBestStruct(pCut *IfCut, pCanonPerm []int8, puCanonPhase *uint32, pBestPo *int) int {
	panic("TODO, stubbed")
}

// If_CutDelayRecCost3
// C: int If_CutDelayRecCost3( If_Man_t * pIfMan, If_Cut_t * pCut, If_Obj_t * pObj )
func (pIfMan *IfMan) IfCutDelayRecCost3(pCut *IfCut, pObj *IfObj) int {
	panic("TODO, stubbed")
}

// Abc_RecToHop3
// C: Hop_Obj_t * Abc_RecToHop3( Hop_Man_t * pMan, If_Man_t * pIfMan, If_Cut_t * pCut, If_Obj_t * pIfObj )
func (pMan *HopManager) AbcRecToHop3(pIfMan *IfMan, pCut *IfCut, pIfObj *IfObj) *HopObject {
	panic("TODO, stubbed")
}

// Abc_RecToGia3
// C: int Abc_RecToGia3( Gia_Man_t * pMan, If_Man_t * pIfMan, If_Cut_t * pCut, Vec_Int_t * vLeaves, int fHash )
func (pMan *GiaManager) AbcRecToGia3(pIfMan *IfMan, pCut *IfCut, vLeaves *IntVector, fHash int) int {
	panic("TODO, stubbed")
}

// Lms_GiaCountTruths
// C: Vec_Int_t * Lms_GiaCountTruths( Lms_Man_t * p )
func (p *LmsMan) LmsGiaCountTruths() *IntVector {
	panic("TODO, stubbed")
}

// Lms_GiaCollectUsefulCos
// C: Vec_Int_t * Lms_GiaCollectUsefulCos( Lms_Man_t * p )
func (p *LmsMan) LmsGiaCollectUsefulCos() *IntVector {
	panic("TODO, stubbed")
}

// Lms_GiaFindNonRedundantCos
// C: Vec_Int_t * Lms_GiaFindNonRedundantCos( Lms_Man_t * p )
func (p *LmsMan) LmsGiaFindNonRedundantCos() *IntVector {
	panic("TODO, stubbed")
}

// Lms_GiaNormalize
// C: void Lms_GiaNormalize( Lms_Man_t * p )
func (p *LmsMan) LmsGiaNormalize() {
	panic("TODO, stubbed")
}

// Abc_NtkRecTruthCompare
// C: int Abc_NtkRecTruthCompare( int * p1, int * p2 )
func AbcNtkRecTruthCompare(p1 *int, p2 *int) int {
	panic("TODO, stubbed")
}

// Abc_NtkRecDumpTt3
// C: void Abc_NtkRecDumpTt3( char * pFileName, int fBinary )
func AbcNtkRecDumpTt3(pFileName string, fBinary int) {
	panic("TODO, stubbed")
}

// Abc_NtkRecInputNum3
// C: int Abc_NtkRecInputNum3()
func AbcNtkRecInputNum3() int {
	panic("TODO, stubbed")
}

// Abc_NtkRecIsRunning3
// C: int Abc_NtkRecIsRunning3()
func AbcNtkRecIsRunning3() int {
	panic("TODO, stubbed")
}

// Abc_NtkRecGetGia3
// C: Gia_Man_t * Abc_NtkRecGetGia3()
func AbcNtkRecGetGia3() *GiaManager {
	panic("TODO, stubbed")
}

// Abc_NtkRecPs3
// C: void Abc_NtkRecPs3(int fPrintLib)
func AbcNtkRecPs3(fPrintLib int) {
	panic("TODO, stubbed")
}

// Abc_NtkRecStart3
// C: void Abc_NtkRecStart3( Gia_Man_t * p, int nVars, int nCuts, int fFuncOnly, int fVerbose )
func (p *GiaManager) AbcNtkRecStart3(nVars int, nCuts int, fFuncOnly int, fVerbose int) {
	panic("TODO, stubbed")
}

// Abc_NtkRecStop3
// C: void Abc_NtkRecStop3()
func AbcNtkRecStop3() {
	panic("TODO, stubbed")
}
