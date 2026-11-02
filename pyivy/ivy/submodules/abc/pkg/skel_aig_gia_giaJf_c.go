package abc

// Hand-checked skeleton of src/aig/gia/giaJf.c.

const (
	JF_LEAF_MAX = 8
	JF_WORD_MAX = 1 << (JF_LEAF_MAX - 6) // ((JF_LEAF_MAX > 6) ? 1 << (JF_LEAF_MAX-6) : 1)
	JF_CUT_MAX  = 16
	JF_EPSILON  = 0.005
)

// JfCut is Jf_Cut_t.
type JfCut struct {
	Sign  uint64               // word : signature
	Flow  float32              // flow
	Time  int                  // arrival time
	iFunc int                  // function
	Cost  int                  // cut cost
	pCut  [JF_LEAF_MAX + 2]int // cut
}

// JfMan is Jf_Man_t.
type JfMan struct {
	pGia     *GiaManager                  // user's manager
	pPars    *JfPar                       // users parameter
	pDsd     *ExtraDSDManager             // Sdm_Man_t * : extern DSD manager
	vCnfs    *IntVector                   // costs of elementary CNFs
	vTtMem   *WordMemory                  // truth table memory and hash table
	vCuts    IntVector                    // Vec_Int_t (by value) : cuts for each node
	vArr     IntVector                    // Vec_Int_t (by value) : arrival time
	vDep     IntVector                    // Vec_Int_t (by value) : departure time
	vFlow    FloatVector                  // Vec_Flt_t (by value) : area flow
	vRefs    FloatVector                  // Vec_Flt_t (by value) : ref counters
	pMem     PagedWordSet                 // Vec_Set_t (by value) : cut storage
	vTemp    *IntVector                   // temporary
	pCutCmp  func(*JfCut, *JfCut) float32 // procedure to compare cuts
	clkStart int64                        // abctime : starting time
	CutCount [4]uint64                    // word : statistics
	nCoarse  int                          // coarse nodes
}

// Jf_ObjIsUnit (static inline)
// C: int Jf_ObjIsUnit( Gia_Obj_t * p )
func jfObjIsUnit(p *GiaObject) int {
	panic("TODO, stubbed")
}

// Jf_ObjCleanUnit (static inline)
// C: void Jf_ObjCleanUnit( Gia_Obj_t * p )
func jfObjCleanUnit(p *GiaObject) {
	panic("TODO, stubbed")
}

// Jf_ObjSetUnit (static inline)
// C: void Jf_ObjSetUnit( Gia_Obj_t * p )
func jfObjSetUnit(p *GiaObject) {
	panic("TODO, stubbed")
}

// Jf_ObjCutH (static inline)
// C: int Jf_ObjCutH( Jf_Man_t * p, int i )
func (p *JfMan) jfObjCutH(i int) int {
	panic("TODO, stubbed")
}

// Jf_ObjCuts (static inline)
// C: int * Jf_ObjCuts( Jf_Man_t * p, int i )
func (p *JfMan) jfObjCuts(i int) []int {
	panic("TODO, stubbed")
}

// Jf_ObjCutBest (static inline)
// C: int * Jf_ObjCutBest( Jf_Man_t * p, int i )
func (p *JfMan) jfObjCutBest(i int) []int {
	panic("TODO, stubbed")
}

// Jf_ObjArr (static inline)
// C: int Jf_ObjArr( Jf_Man_t * p, int i )
func (p *JfMan) jfObjArr(i int) int {
	panic("TODO, stubbed")
}

// Jf_ObjDep (static inline)
// C: int Jf_ObjDep( Jf_Man_t * p, int i )
func (p *JfMan) jfObjDep(i int) int {
	panic("TODO, stubbed")
}

// Jf_ObjFlow (static inline)
// C: float Jf_ObjFlow( Jf_Man_t * p, int i )
func (p *JfMan) jfObjFlow(i int) float32 {
	panic("TODO, stubbed")
}

// Jf_ObjRefs (static inline)
// C: float Jf_ObjRefs( Jf_Man_t * p, int i )
func (p *JfMan) jfObjRefs(i int) float32 {
	panic("TODO, stubbed")
}

// Jf_ObjLit (static inline)
// C: int Jf_ObjLit( int i, int c )
func jfObjLit(i int, c int) int {
	panic("TODO, stubbed")
}

// Jf_CutSize (static inline)
// C: int Jf_CutSize( int * pCut )
func jfCutSize(pCut []int) int {
	panic("TODO, stubbed")
}

// Jf_CutCost (static inline)
// C: int Jf_CutCost( int * pCut )
func jfCutCost(pCut []int) int {
	panic("TODO, stubbed")
}

// Jf_CutFunc (static inline)
// C: int Jf_CutFunc( int * pCut )
func jfCutFunc(pCut []int) int {
	panic("TODO, stubbed")
}

// Jf_CutSetAll (static inline)
// C: int Jf_CutSetAll( int f, int c, int s )
func jfCutSetAll(f int, c int, s int) int {
	panic("TODO, stubbed")
}

// Jf_CutSetSize (static inline)
// C: void Jf_CutSetSize( int * pCut, int s )
func jfCutSetSize(pCut []int, s int) {
	panic("TODO, stubbed")
}

// Jf_CutSetCost (static inline)
// C: void Jf_CutSetCost( int * pCut, int c )
func jfCutSetCost(pCut []int, c int) {
	panic("TODO, stubbed")
}

// Jf_CutSetFunc (static inline)
// C: void Jf_CutSetFunc( int * pCut, int f )
func jfCutSetFunc(pCut []int, f int) {
	panic("TODO, stubbed")
}

// Jf_CutFuncClass (static inline)
// C: int Jf_CutFuncClass( int * pCut )
func jfCutFuncClass(pCut []int) int {
	panic("TODO, stubbed")
}

// Jf_CutFuncCompl (static inline)
// C: int Jf_CutFuncCompl( int * pCut )
func jfCutFuncCompl(pCut []int) int {
	panic("TODO, stubbed")
}

// Jf_CutLits (static inline)
// C: int * Jf_CutLits( int * pCut )
func jfCutLits(pCut []int) []int {
	panic("TODO, stubbed")
}

// Jf_CutLit (static inline)
// C: int Jf_CutLit( int * pCut, int i )
func jfCutLit(pCut []int, i int) int {
	panic("TODO, stubbed")
}

// Jf_CutVar (static inline)
// C: int Jf_CutVar( int * pCut, int i )
func jfCutVar(pCut []int, i int) int {
	panic("TODO, stubbed")
}

// Jf_CutIsTriv (static inline)
// C: int Jf_CutIsTriv( int * pCut, int i )
func jfCutIsTriv(pCut []int, i int) int {
	panic("TODO, stubbed")
}

// Jf_CutCnfSizeF (static inline)
// C: int Jf_CutCnfSizeF( Jf_Man_t * p, int f )
func (p *JfMan) jfCutCnfSizeF(f int) int {
	panic("TODO, stubbed")
}

// Jf_CutCnfSize (static inline)
// C: int Jf_CutCnfSize( Jf_Man_t * p, int * c )
func (p *JfMan) jfCutCnfSize(c []int) int {
	panic("TODO, stubbed")
}

// Jf_ObjFunc0 (static inline)
// C: int Jf_ObjFunc0( Gia_Obj_t * p, int * c )
func jfObjFunc0(p *GiaObject, c []int) int {
	panic("TODO, stubbed")
}

// Jf_ObjFunc1 (static inline)
// C: int Jf_ObjFunc1( Gia_Obj_t * p, int * c )
func jfObjFunc1(p *GiaObject, c []int) int {
	panic("TODO, stubbed")
}

// Jf_ManGenCnf
// C: void Jf_ManGenCnf( word uTruth, int iLitOut, Vec_Int_t * vLeaves, Vec_Int_t * vLits, Vec_Int_t * vClas, Vec_Int_t * vCover )
func JfManGenCnf(uTruth uint64, iLitOut int, vLeaves *IntVector, vLits *IntVector, vClas *IntVector, vCover *IntVector) {
	panic("TODO, stubbed")
}

// Jf_ManCreateCnfRemap
// C: Cnf_Dat_t * Jf_ManCreateCnfRemap( Gia_Man_t * p, Vec_Int_t * vLits, Vec_Int_t * vClas, int fAddOrCla )
func (p *GiaManager) JfManCreateCnfRemap(vLits *IntVector, vClas *IntVector, fAddOrCla int) *CnfDat {
	panic("TODO, stubbed")
}

// Jf_ManCreateCnf
// C: Cnf_Dat_t * Jf_ManCreateCnf( Gia_Man_t * p, Vec_Int_t * vLits, Vec_Int_t * vClas )
func (p *GiaManager) JfManCreateCnf(vLits *IntVector, vClas *IntVector) *CnfDat {
	panic("TODO, stubbed")
}

// Jf_ManInitRefs
// C: float * Jf_ManInitRefs( Jf_Man_t * pMan )
func (pMan *JfMan) JfManInitRefs() []float32 {
	panic("TODO, stubbed")
}

// Jf_ManProfileClasses
// C: void Jf_ManProfileClasses( Jf_Man_t * p )
func (p *JfMan) JfManProfileClasses() {
	panic("TODO, stubbed")
}

// Jf_ManAlloc
// C: Jf_Man_t * Jf_ManAlloc( Gia_Man_t * pGia, Jf_Par_t * pPars )
func (pGia *GiaManager) JfManAlloc(pPars *JfPar) *JfMan {
	panic("TODO, stubbed")
}

// Jf_ManFree
// C: void Jf_ManFree( Jf_Man_t * p )
func (p *JfMan) JfManFree() {
	panic("TODO, stubbed")
}

// Jf_CutPrint (static inline)
// C: void Jf_CutPrint( int * pCut )
func jfCutPrint(pCut []int) {
	panic("TODO, stubbed")
}

// Jf_ObjCutPrint (static inline)
// C: void Jf_ObjCutPrint( int * pCuts )
func jfObjCutPrint(pCuts []int) {
	panic("TODO, stubbed")
}

// Jf_ObjBestCutConePrint (static inline)
// C: void Jf_ObjBestCutConePrint( Jf_Man_t * p, Gia_Obj_t * pObj )
func (p *JfMan) jfObjBestCutConePrint(pObj *GiaObject) {
	panic("TODO, stubbed")
}

// Jf_CutCheck (static inline)
// C: void Jf_CutCheck( int * pCut )
func jfCutCheck(pCut []int) {
	panic("TODO, stubbed")
}

// Jf_CountBitsSimple (static inline)
// C: int Jf_CountBitsSimple( unsigned n )
func jfCountBitsSimple(n uint32) int {
	panic("TODO, stubbed")
}

// Jf_CountBits32 (static inline)
// C: int Jf_CountBits32( unsigned i )
func jfCountBits32(i uint32) int {
	panic("TODO, stubbed")
}

// Jf_CountBits (static inline)
// C: int Jf_CountBits( word i )
func jfCountBits(i uint64) int {
	panic("TODO, stubbed")
}

// Jf_CutGetSign32 (static inline)
// C: unsigned Jf_CutGetSign32( int * pCut )
func jfCutGetSign32(pCut []int) uint32 {
	panic("TODO, stubbed")
}

// Jf_CutGetSign (static inline)
// C: word Jf_CutGetSign( int * pCut )
func jfCutGetSign(pCut []int) uint64 {
	panic("TODO, stubbed")
}

// Jf_CutArr (static inline)
// C: int Jf_CutArr( Jf_Man_t * p, int * pCut )
func (p *JfMan) jfCutArr(pCut []int) int {
	panic("TODO, stubbed")
}

// Jf_ObjSetBestCut (static inline)
// C: void Jf_ObjSetBestCut( int * pCuts, int * pCut, Vec_Int_t * vTemp )
func jfObjSetBestCut(pCuts []int, pCut []int, vTemp *IntVector) {
	panic("TODO, stubbed")
}

// Jf_CutRef (static inline)
// C: void Jf_CutRef( Jf_Man_t * p, int * pCut )
func (p *JfMan) jfCutRef(pCut []int) {
	panic("TODO, stubbed")
}

// Jf_CutDeref (static inline)
// C: void Jf_CutDeref( Jf_Man_t * p, int * pCut )
func (p *JfMan) jfCutDeref(pCut []int) {
	panic("TODO, stubbed")
}

// Jf_CutFlow (static inline)
// C: float Jf_CutFlow( Jf_Man_t * p, int * pCut )
func (p *JfMan) jfCutFlow(pCut []int) float32 {
	panic("TODO, stubbed")
}

// Jf_CutIsContainedOrder (static inline)
// C: int Jf_CutIsContainedOrder( int * pBase, int * pCut )
func jfCutIsContainedOrder(pBase []int, pCut []int) int {
	panic("TODO, stubbed")
}

// Jf_CutMergeOrder (static inline)
// C: int Jf_CutMergeOrder( int * pCut0, int * pCut1, int * pCut, int LutSize )
func jfCutMergeOrder(pCut0 []int, pCut1 []int, pCut []int, LutSize int) int {
	panic("TODO, stubbed")
}

// Jf_CutFindLeaf0 (static inline)
// C: int Jf_CutFindLeaf0( int * pCut, int iObj )
func jfCutFindLeaf0(pCut []int, iObj int) int {
	panic("TODO, stubbed")
}

// Jf_CutIsContained0 (static inline)
// C: int Jf_CutIsContained0( int * pBase, int * pCut )
func jfCutIsContained0(pBase []int, pCut []int) int {
	panic("TODO, stubbed")
}

// Jf_CutMerge0 (static inline)
// C: int Jf_CutMerge0( int * pCut0, int * pCut1, int * pCut, int LutSize )
func jfCutMerge0(pCut0 []int, pCut1 []int, pCut []int, LutSize int) int {
	panic("TODO, stubbed")
}

// Jf_CutFindLeaf1 (static inline)
// C: int Jf_CutFindLeaf1( int * pCut, int iLit )
func jfCutFindLeaf1(pCut []int, iLit int) int {
	panic("TODO, stubbed")
}

// Jf_CutIsContained1 (static inline)
// C: int Jf_CutIsContained1( int * pBase, int * pCut )
func jfCutIsContained1(pBase []int, pCut []int) int {
	panic("TODO, stubbed")
}

// Jf_CutMerge1 (static inline)
// C: int Jf_CutMerge1( int * pCut0, int * pCut1, int * pCut, int LutSize )
func jfCutMerge1(pCut0 []int, pCut1 []int, pCut []int, LutSize int) int {
	panic("TODO, stubbed")
}

// Jf_CutMerge2 (static inline)
// C: int Jf_CutMerge2( int * pCut0, int * pCut1, int * pCut, int LutSize )
func jfCutMerge2(pCut0 []int, pCut1 []int, pCut []int, LutSize int) int {
	panic("TODO, stubbed")
}

// Jf_ObjCutFilterBoth
// C: int Jf_ObjCutFilterBoth( Jf_Man_t * p, Jf_Cut_t ** pSto, int c )
func (p *JfMan) JfObjCutFilterBoth(pSto []*JfCut, c int) int {
	panic("TODO, stubbed")
}

// Jf_ObjCutFilter
// C: int Jf_ObjCutFilter( Jf_Man_t * p, Jf_Cut_t ** pSto, int c )
func (p *JfMan) JfObjCutFilter(pSto []*JfCut, c int) int {
	panic("TODO, stubbed")
}

// Jf_ObjSortCuts (static inline)
// C: void Jf_ObjSortCuts( Jf_Cut_t ** pSto, int nSize )
func jfObjSortCuts(pSto []*JfCut, nSize int) {
	panic("TODO, stubbed")
}

// Jf_CutRef_rec
// C: int Jf_CutRef_rec( Jf_Man_t * p, int * pCut )
func (p *JfMan) JfCutRef_rec(pCut []int) int {
	panic("TODO, stubbed")
}

// Jf_CutDeref_rec
// C: int Jf_CutDeref_rec( Jf_Man_t * p, int * pCut )
func (p *JfMan) JfCutDeref_rec(pCut []int) int {
	panic("TODO, stubbed")
}

// Jf_CutAreaOld (static inline)
// C: int Jf_CutAreaOld( Jf_Man_t * p, int * pCut )
func (p *JfMan) jfCutAreaOld(pCut []int) int {
	panic("TODO, stubbed")
}

// Jf_CutAreaRef_rec
// C: int Jf_CutAreaRef_rec( Jf_Man_t * p, int * pCut )
func (p *JfMan) JfCutAreaRef_rec(pCut []int) int {
	panic("TODO, stubbed")
}

// Jf_CutAreaRefEdge_rec
// C: int Jf_CutAreaRefEdge_rec( Jf_Man_t * p, int * pCut )
func (p *JfMan) JfCutAreaRefEdge_rec(pCut []int) int {
	panic("TODO, stubbed")
}

// Jf_CutArea (static inline)
// C: int Jf_CutArea( Jf_Man_t * p, int * pCut, int fEdge )
func (p *JfMan) jfCutArea(pCut []int, fEdge int) int {
	panic("TODO, stubbed")
}

// Jf_CutCheckMffc_rec
// C: int Jf_CutCheckMffc_rec( Jf_Man_t * p, int * pCut, int Limit )
func (p *JfMan) JfCutCheckMffc_rec(pCut []int, Limit int) int {
	panic("TODO, stubbed")
}

// Jf_CutCheckMffc (static inline)
// C: int Jf_CutCheckMffc( Jf_Man_t * p, int * pCut, int Limit )
func (p *JfMan) jfCutCheckMffc(pCut []int, Limit int) int {
	panic("TODO, stubbed")
}

// Jf_CutCompareDelay
// C: float Jf_CutCompareDelay( Jf_Cut_t * pOld, Jf_Cut_t * pNew )
func JfCutCompareDelay(pOld *JfCut, pNew *JfCut) float32 {
	panic("TODO, stubbed")
}

// Jf_CutCompareArea
// C: float Jf_CutCompareArea( Jf_Cut_t * pOld, Jf_Cut_t * pNew )
func JfCutCompareArea(pOld *JfCut, pNew *JfCut) float32 {
	panic("TODO, stubbed")
}

// Jf_ObjAddCutToStore (static inline)
// C: int Jf_ObjAddCutToStore( Jf_Man_t * p, Jf_Cut_t ** pSto, int c, int cMax )
func (p *JfMan) jfObjAddCutToStore(pSto []*JfCut, c int, cMax int) int {
	panic("TODO, stubbed")
}

// Jf_ObjPrintStore (static inline)
// C: void Jf_ObjPrintStore( Jf_Man_t * p, Jf_Cut_t ** pSto, int c )
func (p *JfMan) jfObjPrintStore(pSto []*JfCut, c int) {
	panic("TODO, stubbed")
}

// Jf_ObjCheckPtrs (static inline)
// C: void Jf_ObjCheckPtrs( Jf_Cut_t ** pSto, int c )
func jfObjCheckPtrs(pSto []*JfCut, c int) {
	panic("TODO, stubbed")
}

// Jf_ObjCheckStore (static inline)
// C: void Jf_ObjCheckStore( Jf_Man_t * p, Jf_Cut_t ** pSto, int c, int iObj )
func (p *JfMan) jfObjCheckStore(pSto []*JfCut, c int, iObj int) {
	panic("TODO, stubbed")
}

// Jf_TtComputeForCut
// C: int Jf_TtComputeForCut( Jf_Man_t * p, int iFuncLit0, int iFuncLit1, int * pCut0, int * pCut1, int * pCutOut )
func (p *JfMan) JfTtComputeForCut(iFuncLit0 int, iFuncLit1 int, pCut0 []int, pCut1 []int, pCutOut []int) int {
	panic("TODO, stubbed")
}

// Jf_ObjAssignCut (static inline)
// C: void Jf_ObjAssignCut( Jf_Man_t * p, Gia_Obj_t * pObj )
func (p *JfMan) jfObjAssignCut(pObj *GiaObject) {
	panic("TODO, stubbed")
}

// Jf_ObjPropagateBuf (static inline)
// C: void Jf_ObjPropagateBuf( Jf_Man_t * p, Gia_Obj_t * pObj, int fReverse )
func (p *JfMan) jfObjPropagateBuf(pObj *GiaObject, fReverse int) {
	panic("TODO, stubbed")
}

// Jf_ObjHasCutWithSize (static inline)
// C: int Jf_ObjHasCutWithSize( Jf_Cut_t ** pSto, int c, int nSize )
func jfObjHasCutWithSize(pSto []*JfCut, c int, nSize int) int {
	panic("TODO, stubbed")
}

// Jf_ObjComputeCuts
// C: void Jf_ObjComputeCuts( Jf_Man_t * p, Gia_Obj_t * pObj, int fEdge )
func (p *JfMan) JfObjComputeCuts(pObj *GiaObject, fEdge int) {
	panic("TODO, stubbed")
}

// Jf_ManComputeCuts
// C: void Jf_ManComputeCuts( Jf_Man_t * p, int fEdge )
func (p *JfMan) JfManComputeCuts(fEdge int) {
	panic("TODO, stubbed")
}

// Jf_ManComputeDelay
// C: int Jf_ManComputeDelay( Jf_Man_t * p, int fEval )
func (p *JfMan) JfManComputeDelay(fEval int) int {
	panic("TODO, stubbed")
}

// Jf_ManComputeRefs
// C: int Jf_ManComputeRefs( Jf_Man_t * p )
func (p *JfMan) JfManComputeRefs() int {
	panic("TODO, stubbed")
}

// Jf_ObjComputeBestCut
// C: void Jf_ObjComputeBestCut( Jf_Man_t * p, Gia_Obj_t * pObj, int fEdge, int fEla )
func (p *JfMan) JfObjComputeBestCut(pObj *GiaObject, fEdge int, fEla int) {
	panic("TODO, stubbed")
}

// Jf_ManPropagateFlow
// C: void Jf_ManPropagateFlow( Jf_Man_t * p, int fEdge )
func (p *JfMan) JfManPropagateFlow(fEdge int) {
	panic("TODO, stubbed")
}

// Jf_ManPropagateEla
// C: void Jf_ManPropagateEla( Jf_Man_t * p, int fEdge )
func (p *JfMan) JfManPropagateEla(fEdge int) {
	panic("TODO, stubbed")
}

// Jf_ManDeriveMappingGia
// C: Gia_Man_t * Jf_ManDeriveMappingGia( Jf_Man_t * p )
func (p *JfMan) JfManDeriveMappingGia() *GiaManager {
	panic("TODO, stubbed")
}

// Jf_ManDeriveMapping
// C: void Jf_ManDeriveMapping( Jf_Man_t * p )
func (p *JfMan) JfManDeriveMapping() {
	panic("TODO, stubbed")
}

// Jf_ManDeriveGia
// C: Gia_Man_t * Jf_ManDeriveGia( Jf_Man_t * p )
func (p *JfMan) JfManDeriveGia() *GiaManager {
	panic("TODO, stubbed")
}

// Jf_ManSetDefaultPars
// C: void Jf_ManSetDefaultPars( Jf_Par_t * pPars )
func JfManSetDefaultPars(pPars *JfPar) {
	panic("TODO, stubbed")
}

// Jf_ManPrintStats
// C: void Jf_ManPrintStats( Jf_Man_t * p, char * pTitle )
func (p *JfMan) JfManPrintStats(pTitle string) {
	panic("TODO, stubbed")
}

// Jf_ManPerformMapping
// C: Gia_Man_t * Jf_ManPerformMapping( Gia_Man_t * pGia, Jf_Par_t * pPars )
func (pGia *GiaManager) JfManPerformMapping(pPars *JfPar) *GiaManager {
	panic("TODO, stubbed")
}

// Jf_ManDeriveCnf
// C: Gia_Man_t * Jf_ManDeriveCnf( Gia_Man_t * p, int fCnfObjIds )
func (p *GiaManager) JfManDeriveCnf(fCnfObjIds int) *GiaManager {
	panic("TODO, stubbed")
}

// Jf_ManDeriveCnfMiter
// C: Gia_Man_t * Jf_ManDeriveCnfMiter( Gia_Man_t * p, int fVerbose )
func (p *GiaManager) JfManDeriveCnfMiter(fVerbose int) *GiaManager {
	panic("TODO, stubbed")
}

// Jf_ManDumpCnf
// C: void Jf_ManDumpCnf( Gia_Man_t * p, char * pFileName, int fVerbose )
func (p *GiaManager) JfManDumpCnf(pFileName string, fVerbose int) {
	panic("TODO, stubbed")
}

// Jf_ManTestCnf
// C: void Jf_ManTestCnf( Gia_Man_t * p )
func (p *GiaManager) JfManTestCnf() {
	panic("TODO, stubbed")
}
