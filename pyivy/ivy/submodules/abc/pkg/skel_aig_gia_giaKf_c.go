package abc

// Hand-checked skeleton of src/aig/gia/giaKf.c.

const (
	KF_LEAF_MAX  = 16
	KF_CUT_MAX   = 32
	KF_PROC_MAX  = 32
	KF_WORD_MAX  = 1 << (KF_LEAF_MAX - 6) // ((KF_LEAF_MAX > 6) ? 1 << (KF_LEAF_MAX-6) : 1)
	KF_LOG_TABLE = 8
	KF_ADD_ON1   = 2 // offset in cut storage for each node (cut count; best cut)
	KF_ADD_ON2   = 4 // offset in cut storage for each cut (leaf count; function, cut delay; cut area)
)

// KfCut is Kf_Cut_t.
type KfCut struct {
	Sign    uint64  // word : signature
	Polar   int     // polarity
	Delay   int     // delay
	Area    float32 // area
	iFunc   int     // function
	iNext   int     // next cut
	nLeaves int     // number of leaves
	pLeaves [KF_LEAF_MAX]int
}

// KfSet is Kf_Set_t.
type KfSet struct {
	pMan      *KfMan // manager
	nLutSize  uint16 // lut size
	nCutNum   uint16 // cut count
	nCuts0    int    // fanin0 cut count
	nCuts1    int    // fanin1 cut count
	nCuts     int    // resulting cut count
	nTEntries int    // hash table entries
	TableMask int    // hash table mask
	pTable    [1 << KF_LOG_TABLE]int
	pValue    [1 << KF_LOG_TABLE]int
	pPlace    [KF_LEAF_MAX]int
	pList     [KF_LEAF_MAX + 1]int
	pCuts0    [KF_CUT_MAX]KfCut
	pCuts1    [KF_CUT_MAX]KfCut
	pCutsR    [KF_CUT_MAX * KF_CUT_MAX]KfCut
	ppCuts    [KF_CUT_MAX]*KfCut
	pCutBest  *KfCut
	CutCount  [4]uint64 // word : statistics
}

// KfMan is Kf_Man_t.
type KfMan struct {
	pGia     *GiaManager  // user's manager
	pPars    *JfPar       // user's parameters
	pMem     PagedWordSet // Vec_Set_t (by value) : cut storage
	vCuts    IntVector    // Vec_Int_t (by value) : node params
	vTime    IntVector    // Vec_Int_t (by value) : node params
	vArea    FloatVector  // Vec_Flt_t (by value) : node params
	vRefs    FloatVector  // Vec_Flt_t (by value) : node params
	vTemp    *IntVector   // temporary
	clkStart int64        // abctime : starting time
	pSett    [KF_PROC_MAX]KfSet
}

// KfThData is Kf_ThData_t.
type KfThData struct {
	pSett   *KfSet
	Id      int
	Status  int
	clkUsed int64 // abctime
}

// Kf_ManSetDefaultPars
// C: void Kf_ManSetDefaultPars( Jf_Par_t * pPars )
func KfManSetDefaultPars(pPars *JfPar) {
	panic("TODO, stubbed")
}

// Kf_ManPerformMapping
// C: Gia_Man_t * Kf_ManPerformMapping( Gia_Man_t * pGia, Jf_Par_t * pPars )
func (pGia *GiaManager) KfManPerformMapping(pPars *JfPar) *GiaManager {
	panic("TODO, stubbed")
}

// Kf_SetCutId (static inline)
// C: int Kf_SetCutId( Kf_Set_t * p, Kf_Cut_t * pCut )
func (p *KfSet) kfSetCutId(pCut *KfCut) int {
	panic("TODO, stubbed")
}

// Kf_SetCut (static inline)
// C: Kf_Cut_t * Kf_SetCut( Kf_Set_t * p, int i )
func (p *KfSet) kfSetCut(i int) *KfCut {
	panic("TODO, stubbed")
}

// Kf_ObjTime (static inline)
// C: int Kf_ObjTime( Kf_Man_t * p, int i )
func (p *KfMan) kfObjTime(i int) int {
	panic("TODO, stubbed")
}

// Kf_ObjArea (static inline)
// C: float Kf_ObjArea( Kf_Man_t * p, int i )
func (p *KfMan) kfObjArea(i int) float32 {
	panic("TODO, stubbed")
}

// Kf_ObjRefs (static inline)
// C: float Kf_ObjRefs( Kf_Man_t * p, int i )
func (p *KfMan) kfObjRefs(i int) float32 {
	panic("TODO, stubbed")
}

// Kf_ObjSetCuts (static inline)
// C: void Kf_ObjSetCuts( Kf_Man_t * p, int i, Vec_Int_t * vVec )
func (p *KfMan) kfObjSetCuts(i int, vVec *IntVector) {
	panic("TODO, stubbed")
}

// Kf_ObjCuts (static inline)
// C: int * Kf_ObjCuts( Kf_Man_t * p, int i )
func (p *KfMan) kfObjCuts(i int) []int {
	panic("TODO, stubbed")
}

// Kf_ObjCuts0 (static inline)
// C: int * Kf_ObjCuts0( Kf_Man_t * p, int i )
func (p *KfMan) kfObjCuts0(i int) []int {
	panic("TODO, stubbed")
}

// Kf_ObjCuts1 (static inline)
// C: int * Kf_ObjCuts1( Kf_Man_t * p, int i )
func (p *KfMan) kfObjCuts1(i int) []int {
	panic("TODO, stubbed")
}

// Kf_ObjCutBest (static inline)
// C: int * Kf_ObjCutBest( Kf_Man_t * p, int i )
func (p *KfMan) kfObjCutBest(i int) []int {
	panic("TODO, stubbed")
}

// Kf_SetLoadCuts (static inline)
// C: int Kf_SetLoadCuts( Kf_Cut_t * pCuts, int * pIntCuts )
func kfSetLoadCuts(pCuts []KfCut, pIntCuts []int) int {
	panic("TODO, stubbed")
}

// Kf_SetPrepare (static inline)
// C: void Kf_SetPrepare( Kf_Set_t * p, int * pCuts0, int * pCuts1 )
func (p *KfSet) kfSetPrepare(pCuts0 []int, pCuts1 []int) {
	panic("TODO, stubbed")
}

// Kf_ManStoreStart (static inline)
// C: void Kf_ManStoreStart( Vec_Int_t * vTemp, int nCuts )
func kfManStoreStart(vTemp *IntVector, nCuts int) {
	panic("TODO, stubbed")
}

// Kf_ManStoreAddUnit (static inline)
// C: void Kf_ManStoreAddUnit( Vec_Int_t * vTemp, int iObj, int Time, float Area )
func kfManStoreAddUnit(vTemp *IntVector, iObj int, Time int, Area float32) {
	panic("TODO, stubbed")
}

// Kf_ManSaveResults (static inline)
// C: void Kf_ManSaveResults( Kf_Cut_t ** ppCuts, int nCuts, Kf_Cut_t * pCutBest, Vec_Int_t * vTemp )
func kfManSaveResults(ppCuts []*KfCut, nCuts int, pCutBest *KfCut, vTemp *IntVector) {
	panic("TODO, stubbed")
}

// Kf_SetCompareCuts (static inline)
// C: int Kf_SetCompareCuts( Kf_Cut_t * p1, Kf_Cut_t * p2 )
func kfSetCompareCuts(p1 *KfCut, p2 *KfCut) int {
	panic("TODO, stubbed")
}

// Kf_SetAddToList (static inline)
// C: void Kf_SetAddToList( Kf_Set_t * p, Kf_Cut_t * pCut, int fSort )
func (p *KfSet) kfSetAddToList(pCut *KfCut, fSort int) {
	panic("TODO, stubbed")
}

// Kf_CutCompare (static inline)
// C: int Kf_CutCompare( Kf_Cut_t * pCut0, Kf_Cut_t * pCut1, int fArea )
func kfCutCompare(pCut0 *KfCut, pCut1 *KfCut, fArea int) int {
	panic("TODO, stubbed")
}

// Kf_SetStoreAddOne (static inline)
// C: int Kf_SetStoreAddOne( Kf_Set_t * p, int nCuts, int nCutNum, Kf_Cut_t * pCut, int fArea )
func (p *KfSet) kfSetStoreAddOne(nCuts int, nCutNum int, pCut *KfCut, fArea int) int {
	panic("TODO, stubbed")
}

// Kf_SetSelectBest (static inline)
// C: void Kf_SetSelectBest( Kf_Set_t * p, int fArea, int fSort )
func (p *KfSet) kfSetSelectBest(fArea int, fSort int) {
	panic("TODO, stubbed")
}

// Kf_CheckCut (static inline)
// C: int Kf_CheckCut( Kf_Cut_t * pBase, Kf_Cut_t * pCut )
func kfCheckCut(pBase *KfCut, pCut *KfCut) int {
	panic("TODO, stubbed")
}

// Kf_CheckCuts (static inline)
// C: int Kf_CheckCuts( Kf_Set_t * p )
func (p *KfSet) kfCheckCuts() int {
	panic("TODO, stubbed")
}

// Kf_HashLookup (static inline)
// C: int Kf_HashLookup( Kf_Set_t * p, int i )
func (p *KfSet) kfHashLookup(i int) int {
	panic("TODO, stubbed")
}

// Kf_HashFindOrAdd (static inline)
// C: int Kf_HashFindOrAdd( Kf_Set_t * p, int i )
func (p *KfSet) kfHashFindOrAdd(i int) int {
	panic("TODO, stubbed")
}

// Kf_HashPopulate (static inline)
// C: void Kf_HashPopulate( Kf_Set_t * p, Kf_Cut_t * pCut )
func (p *KfSet) kfHashPopulate(pCut *KfCut) {
	panic("TODO, stubbed")
}

// Kf_HashCleanup (static inline)
// C: void Kf_HashCleanup( Kf_Set_t * p, int iStart )
func (p *KfSet) kfHashCleanup(iStart int) {
	panic("TODO, stubbed")
}

// Kf_SetCountBits (static inline)
// C: int Kf_SetCountBits( word i )
func kfSetCountBits(i uint64) int {
	panic("TODO, stubbed")
}

// Kf_SetCutGetSign (static inline)
// C: word Kf_SetCutGetSign( Kf_Cut_t * p )
func kfSetCutGetSign(p *KfCut) uint64 {
	panic("TODO, stubbed")
}

// Kf_SetCutDominatedByThis (static inline)
// C: int Kf_SetCutDominatedByThis( Kf_Set_t * p, Kf_Cut_t * pCut )
func (p *KfSet) kfSetCutDominatedByThis(pCut *KfCut) int {
	panic("TODO, stubbed")
}

// Kf_SetRemoveDuplicates (static inline)
// C: int Kf_SetRemoveDuplicates( Kf_Set_t * p, int nLeaves, word Sign )
func (p *KfSet) kfSetRemoveDuplicates(nLeaves int, Sign uint64) int {
	panic("TODO, stubbed")
}

// Kf_SetFilter (static inline)
// C: void Kf_SetFilter( Kf_Set_t * p )
func (p *KfSet) kfSetFilter() {
	panic("TODO, stubbed")
}

// Kf_SetMergePairs (static inline)
// C: void Kf_SetMergePairs( Kf_Set_t * p, Kf_Cut_t * pCut0, Kf_Cut_t * pCuts, int nCuts, int fArea )
func (p *KfSet) kfSetMergePairs(pCut0 *KfCut, pCuts []KfCut, nCuts int, fArea int) {
	panic("TODO, stubbed")
}

// Kf_SetMerge (static inline)
// C: void Kf_SetMerge( Kf_Set_t * p, int * pCuts0, int * pCuts1, int fArea, int fCutMin )
func (p *KfSet) kfSetMerge(pCuts0 []int, pCuts1 []int, fArea int, fCutMin int) {
	panic("TODO, stubbed")
}

// Kf_SetCutIsContainedSimple (static inline)
// C: int Kf_SetCutIsContainedSimple( Kf_Cut_t * pBase, Kf_Cut_t * pCut )
func kfSetCutIsContainedSimple(pBase *KfCut, pCut *KfCut) int {
	panic("TODO, stubbed")
}

// Kf_SetMergeSimpleOne (static inline)
// C: int Kf_SetMergeSimpleOne( Kf_Cut_t * pCut0, Kf_Cut_t * pCut1, Kf_Cut_t * pCut, int nLutSize )
func kfSetMergeSimpleOne(pCut0 *KfCut, pCut1 *KfCut, pCut *KfCut, nLutSize int) int {
	panic("TODO, stubbed")
}

// Kf_SetRemoveDuplicatesSimple (static inline)
// C: int Kf_SetRemoveDuplicatesSimple( Kf_Set_t * p, Kf_Cut_t * pCutNew )
func (p *KfSet) kfSetRemoveDuplicatesSimple(pCutNew *KfCut) int {
	panic("TODO, stubbed")
}

// Kf_SetFilterSimple (static inline)
// C: void Kf_SetFilterSimple( Kf_Set_t * p )
func (p *KfSet) kfSetFilterSimple() {
	panic("TODO, stubbed")
}

// Kf_SetMergeSimple (static inline)
// C: void Kf_SetMergeSimple( Kf_Set_t * p, int * pCuts0, int * pCuts1, int fArea, int fCutMin )
func (p *KfSet) kfSetMergeSimple(pCuts0 []int, pCuts1 []int, fArea int, fCutMin int) {
	panic("TODO, stubbed")
}

// Kf_SetCutIsContainedOrder (static inline)
// C: int Kf_SetCutIsContainedOrder( Kf_Cut_t * pBase, Kf_Cut_t * pCut )
func kfSetCutIsContainedOrder(pBase *KfCut, pCut *KfCut) int {
	panic("TODO, stubbed")
}

// Kf_SetMergeOrderOne (static inline)
// C: int Kf_SetMergeOrderOne( Kf_Cut_t * pCut0, Kf_Cut_t * pCut1, Kf_Cut_t * pCut, int nLutSize )
func kfSetMergeOrderOne(pCut0 *KfCut, pCut1 *KfCut, pCut *KfCut, nLutSize int) int {
	panic("TODO, stubbed")
}

// Kf_SetRemoveDuplicatesOrder (static inline)
// C: int Kf_SetRemoveDuplicatesOrder( Kf_Set_t * p, Kf_Cut_t * pCutNew )
func (p *KfSet) kfSetRemoveDuplicatesOrder(pCutNew *KfCut) int {
	panic("TODO, stubbed")
}

// Kf_SetFilterOrder (static inline)
// C: void Kf_SetFilterOrder( Kf_Set_t * p )
func (p *KfSet) kfSetFilterOrder() {
	panic("TODO, stubbed")
}

// Kf_SetMergeOrder (static inline)
// C: void Kf_SetMergeOrder( Kf_Set_t * p, int * pCuts0, int * pCuts1, int fArea, int fCutMin )
func (p *KfSet) kfSetMergeOrder(pCuts0 []int, pCuts1 []int, fArea int, fCutMin int) {
	panic("TODO, stubbed")
}

// Kf_CutSize (static inline)
// C: int Kf_CutSize( int * pCut )
func kfCutSize(pCut []int) int {
	panic("TODO, stubbed")
}

// Kf_CutFunc (static inline)
// C: int Kf_CutFunc( int * pCut )
func kfCutFunc(pCut []int) int {
	panic("TODO, stubbed")
}

// Kf_CutLeaf (static inline)
// C: int Kf_CutLeaf( int * pCut, int i )
func kfCutLeaf(pCut []int, i int) int {
	panic("TODO, stubbed")
}

// Kf_CutTime (static inline)
// C: int Kf_CutTime( Kf_Man_t * p, int * pCut )
func (p *KfMan) kfCutTime(pCut []int) int {
	panic("TODO, stubbed")
}

// Kf_CutRef (static inline)
// C: void Kf_CutRef( Kf_Man_t * p, int * pCut )
func (p *KfMan) kfCutRef(pCut []int) {
	panic("TODO, stubbed")
}

// Kf_CutDeref (static inline)
// C: void Kf_CutDeref( Kf_Man_t * p, int * pCut )
func (p *KfMan) kfCutDeref(pCut []int) {
	panic("TODO, stubbed")
}

// Kf_CutPrint (static inline)
// C: void Kf_CutPrint( int * pCut )
func kfCutPrint(pCut []int) {
	panic("TODO, stubbed")
}

// Gia_CutSetPrint (static inline)
// C: void Gia_CutSetPrint( int * pCuts )
func giaCutSetPrint(pCuts []int) {
	panic("TODO, stubbed")
}

// Kf_ManComputeDelay
// C: int Kf_ManComputeDelay( Kf_Man_t * p, int fEval )
func (p *KfMan) KfManComputeDelay(fEval int) int {
	panic("TODO, stubbed")
}

// Kf_ManComputeRefs
// C: int Kf_ManComputeRefs( Kf_Man_t * p )
func (p *KfMan) KfManComputeRefs() int {
	panic("TODO, stubbed")
}

// Kf_WorkerThread (pArg is a void * that holds a Kf_ThData_t *)
// C: void * Kf_WorkerThread( void * pArg )
func KfWorkerThread(pArg *KfThData) any {
	panic("TODO, stubbed")
}

// Kf_ManCreateFaninCounts
// C: Vec_Int_t * Kf_ManCreateFaninCounts( Gia_Man_t * p )
func (p *GiaManager) KfManCreateFaninCounts() *IntVector {
	panic("TODO, stubbed")
}

// Kf_ManComputeCuts
// C: void Kf_ManComputeCuts( Kf_Man_t * p )
func (p *KfMan) KfManComputeCuts() {
	panic("TODO, stubbed")
}

// Kf_ManPrintStats
// C: void Kf_ManPrintStats( Kf_Man_t * p, char * pTitle )
func (p *KfMan) KfManPrintStats(pTitle string) {
	panic("TODO, stubbed")
}

// Kf_ManComputeMapping
// C: void Kf_ManComputeMapping( Kf_Man_t * p )
func (p *KfMan) KfManComputeMapping() {
	panic("TODO, stubbed")
}

// Kf_ManSetInitRefs
// C: void Kf_ManSetInitRefs( Gia_Man_t * p, Vec_Flt_t * vRefs )
func (p *GiaManager) KfManSetInitRefs(vRefs *FloatVector) {
	panic("TODO, stubbed")
}

// Kf_ManAlloc
// C: Kf_Man_t * Kf_ManAlloc( Gia_Man_t * pGia, Jf_Par_t * pPars )
func (pGia *GiaManager) KfManAlloc(pPars *JfPar) *KfMan {
	panic("TODO, stubbed")
}

// Kf_ManFree
// C: void Kf_ManFree( Kf_Man_t * p )
func (p *KfMan) KfManFree() {
	panic("TODO, stubbed")
}

// Kf_ManDerive
// C: Gia_Man_t * Kf_ManDerive( Kf_Man_t * p )
func (p *KfMan) KfManDerive() *GiaManager {
	panic("TODO, stubbed")
}
