package abc

// Hand-checked skeleton of src/aig/gia/giaBalLut.c.

const (
	BAL_LEAF_MAX = 6
	BAL_CUT_MAX  = 8
	BAL_SUPER    = 50
	BAL_NO_LEAF  = 31
	BAL_NO_FUNC  = 134217727 // (1<<27)-1
)

// BalCut is Bal_Cut_t.
type BalCut struct {
	Sign    uint64            // word : signature
	Delay   int               // delay
	iFunc   uint32            // unsigned : 27 : function (BAL_NO_FUNC)
	nLeaves uint32            // unsigned : 5 : leaf number (Bal_NO_LEAF)
	pLeaves [BAL_LEAF_MAX]int // leaves
}

// BalMan is Bal_Man_t: the operation manager.
type BalMan struct {
	pGia     *GiaManager            // user AIG
	nLutSize int                    // LUT size
	nCutNum  int                    // cut number
	fCutMin  int                    // cut minimization
	fVerbose int                    // verbose
	pNew     *GiaManager            // derived AIG
	vCosts   *IntVector             // cost of supergate nodes
	vCutSets *PointerVector[BalCut] // object cutsets
	clkStart int64                  // abctime : starting the clock
}

// Bal_GiaMan (static inline)
// C: Bal_Man_t * Bal_GiaMan( Gia_Man_t * p )
func (p *GiaManager) balGiaMan() *BalMan {
	panic("TODO, stubbed")
}

// Bal_ObjCost (static inline)
// C: int Bal_ObjCost( Bal_Man_t * p, int i )
func (p *BalMan) balObjCost(i int) int {
	panic("TODO, stubbed")
}

// Bal_LitCost (static inline)
// C: int Bal_LitCost( Bal_Man_t * p, int i )
func (p *BalMan) balLitCost(i int) int {
	panic("TODO, stubbed")
}

// Bal_ObjDelay (static inline)
// C: int Bal_ObjDelay( Bal_Man_t * p, int i )
func (p *BalMan) balObjDelay(i int) int {
	panic("TODO, stubbed")
}

// Bal_LitDelay (static inline)
// C: int Bal_LitDelay( Bal_Man_t * p, int i )
func (p *BalMan) balLitDelay(i int) int {
	panic("TODO, stubbed")
}

// Bal_ManAlloc
// C: Bal_Man_t * Bal_ManAlloc( Gia_Man_t * pGia, Gia_Man_t * pNew, int nLutSize, int nCutNum, int fVerbose )
func (pGia *GiaManager) BalManAlloc(pNew *GiaManager, nLutSize int, nCutNum int, fVerbose int) *BalMan {
	panic("TODO, stubbed")
}

// Bal_ManFree
// C: void Bal_ManFree( Bal_Man_t * p )
func (p *BalMan) BalManFree() {
	panic("TODO, stubbed")
}

// Bal_CutCountBits (static inline)
// C: int Bal_CutCountBits( word i )
func balCutCountBits(i uint64) int {
	panic("TODO, stubbed")
}

// Bal_CutGetSign (static inline)
// C: word Bal_CutGetSign( int * pLeaves, int nLeaves )
func balCutGetSign(pLeaves []int, nLeaves int) uint64 {
	panic("TODO, stubbed")
}

// Bal_CutCreateUnit (static inline)
// C: int Bal_CutCreateUnit( Bal_Cut_t * p, int i, int Delay )
func balCutCreateUnit(p *BalCut, i int, Delay int) int {
	panic("TODO, stubbed")
}

// Bal_ManPrepareSet (static inline)
// C: int Bal_ManPrepareSet( Bal_Man_t * p, int iObj, int Index, int fUnit, Bal_Cut_t ** ppCutSet )
func (p *BalMan) balManPrepareSet(iObj int, Index int, fUnit int, ppCutSet **BalCut) int {
	panic("TODO, stubbed")
}

// Bal_CutCheck (static inline)
// C: int Bal_CutCheck( Bal_Cut_t * pBase, Bal_Cut_t * pCut )
func balCutCheck(pBase *BalCut, pCut *BalCut) int {
	panic("TODO, stubbed")
}

// Bal_SetCheckArray (static inline)
// C: int Bal_SetCheckArray( Bal_Cut_t ** ppCuts, int nCuts )
func balSetCheckArray(ppCuts []*BalCut, nCuts int) int {
	panic("TODO, stubbed")
}

// Bal_CutMergeOrder (static inline)
// C: int Bal_CutMergeOrder( Bal_Cut_t * pCut0, Bal_Cut_t * pCut1, Bal_Cut_t * pCut, int nLutSize )
func balCutMergeOrder(pCut0 *BalCut, pCut1 *BalCut, pCut *BalCut, nLutSize int) int {
	panic("TODO, stubbed")
}

// Bal_CutMergeOrderMux (static inline)
// C: int Bal_CutMergeOrderMux( Bal_Cut_t * pCut0, Bal_Cut_t * pCut1, Bal_Cut_t * pCut2, Bal_Cut_t * pCut, int nLutSize )
func balCutMergeOrderMux(pCut0 *BalCut, pCut1 *BalCut, pCut2 *BalCut, pCut *BalCut, nLutSize int) int {
	panic("TODO, stubbed")
}

// Bal_SetCutIsContainedOrder (static inline)
// C: int Bal_SetCutIsContainedOrder( Bal_Cut_t * pBase, Bal_Cut_t * pCut )
func balSetCutIsContainedOrder(pBase *BalCut, pCut *BalCut) int {
	panic("TODO, stubbed")
}

// Bal_SetLastCutIsContained (static inline)
// C: int Bal_SetLastCutIsContained( Bal_Cut_t ** pCuts, int nCuts )
func balSetLastCutIsContained(pCuts []*BalCut, nCuts int) int {
	panic("TODO, stubbed")
}

// Bal_SetLastCutContains (static inline)
// C: int Bal_SetLastCutContains( Bal_Cut_t ** pCuts, int nCuts )
func balSetLastCutContains(pCuts []*BalCut, nCuts int) int {
	panic("TODO, stubbed")
}

// Bal_CutCompareArea (static inline)
// C: int Bal_CutCompareArea( Bal_Cut_t * pCut0, Bal_Cut_t * pCut1 )
func balCutCompareArea(pCut0 *BalCut, pCut1 *BalCut) int {
	panic("TODO, stubbed")
}

// Bal_SetSortByDelay (static inline)
// C: void Bal_SetSortByDelay( Bal_Cut_t ** pCuts, int nCuts )
func balSetSortByDelay(pCuts []*BalCut, nCuts int) {
	panic("TODO, stubbed")
}

// Bal_SetAddCut (static inline)
// C: int Bal_SetAddCut( Bal_Cut_t ** pCuts, int nCuts, int nCutNum )
func balSetAddCut(pCuts []*BalCut, nCuts int, nCutNum int) int {
	panic("TODO, stubbed")
}

// Bal_ManDeriveCuts
// C: int Bal_ManDeriveCuts( Bal_Man_t * p, int iFan0, int iFan1, int iFan2, int fCompl0, int fCompl1, int fCompl2, int fUnit0, int fUnit1, int fUnit2, int fIsXor, int Target, int fSave )
func (p *BalMan) BalManDeriveCuts(iFan0 int, iFan1 int, iFan2 int, fCompl0 int, fCompl1 int, fCompl2 int, fUnit0 int, fUnit1 int, fUnit2 int, fIsXor int, Target int, fSave int) int {
	panic("TODO, stubbed")
}

// Bal_ManSetGateLevel
// C: int Bal_ManSetGateLevel( Bal_Man_t * p, Gia_Obj_t * pObjOld, int iLitNew )
func (p *BalMan) BalManSetGateLevel(pObjOld *GiaObject, iLitNew int) int {
	panic("TODO, stubbed")
}

// Bal_ManEvalTwo
// C: int Bal_ManEvalTwo( Bal_Man_t * p, int iLitNew0, int iLitNew1, int iLitNew2, int fIsXor )
func (p *BalMan) BalManEvalTwo(iLitNew0 int, iLitNew1 int, iLitNew2 int, fIsXor int) int {
	panic("TODO, stubbed")
}

// Vec_IntSelectSortCostLit (static inline)
// C: void Vec_IntSelectSortCostLit( Vec_Int_t * vSuper, Vec_Int_t * vCosts )
func vecIntSelectSortCostLit(vSuper *IntVector, vCosts *IntVector) {
	panic("TODO, stubbed")
}

// Vec_IntPushOrderCost2 (static inline)
// C: void Vec_IntPushOrderCost2( Vec_Int_t * vSuper, Vec_Int_t * vCosts, int iLit )
func vecIntPushOrderCost2(vSuper *IntVector, vCosts *IntVector, iLit int) {
	panic("TODO, stubbed")
}

// Vec_IntFindFirstSameDelayAsLast (static inline)
// C: int Vec_IntFindFirstSameDelayAsLast( Bal_Man_t * p, Vec_Int_t * vSuper )
func (p *BalMan) vecIntFindFirstSameDelayAsLast(vSuper *IntVector) int {
	panic("TODO, stubbed")
}

// Bal_ManFindBestPair (static inline)
// C: int Bal_ManFindBestPair( Bal_Man_t * p, Vec_Int_t * vSuper, Gia_Obj_t * pObj )
func (p *BalMan) balManFindBestPair(vSuper *IntVector, pObj *GiaObject) int {
	panic("TODO, stubbed")
}

// Gia_ManSimplifyXor (static inline)
// C: void Gia_ManSimplifyXor( Vec_Int_t * vSuper )
func giaManSimplifyXor(vSuper *IntVector) {
	panic("TODO, stubbed")
}

// Gia_ManSimplifyAnd (static inline)
// C: void Gia_ManSimplifyAnd( Vec_Int_t * vSuper )
func giaManSimplifyAnd(vSuper *IntVector) {
	panic("TODO, stubbed")
}

// Gia_ManSuperCollectXor_rec (static inline)
// C: void Gia_ManSuperCollectXor_rec( Gia_Man_t * p, Gia_Obj_t * pObj )
func (p *GiaManager) giaManSuperCollectXor_rec(pObj *GiaObject) {
	panic("TODO, stubbed")
}

// Gia_ManSuperCollectAnd_rec (static inline)
// C: void Gia_ManSuperCollectAnd_rec( Gia_Man_t * p, Gia_Obj_t * pObj )
func (p *GiaManager) giaManSuperCollectAnd_rec(pObj *GiaObject) {
	panic("TODO, stubbed")
}

// Gia_ManSuperCollect (static inline)
// C: void Gia_ManSuperCollect( Gia_Man_t * p, Gia_Obj_t * pObj )
func (p *GiaManager) giaManSuperCollect(pObj *GiaObject) {
	panic("TODO, stubbed")
}

// Gia_ManCreateGate (static inline)
// C: void Gia_ManCreateGate( Gia_Man_t * pNew, Gia_Obj_t * pObj, Vec_Int_t * vSuper )
func (pNew *GiaManager) giaManCreateGate(pObj *GiaObject, vSuper *IntVector) {
	panic("TODO, stubbed")
}

// Gia_ManBalanceGate (static inline)
// C: int Gia_ManBalanceGate( Gia_Man_t * pNew, Gia_Obj_t * pObj, Vec_Int_t * vSuper, int * pLits, int nLits )
func (pNew *GiaManager) giaManBalanceGate(pObj *GiaObject, vSuper *IntVector, pLits []int, nLits int) int {
	panic("TODO, stubbed")
}

// Gia_ManBalance_rec (static inline)
// C: void Gia_ManBalance_rec( Gia_Man_t * pNew, Gia_Man_t * p, Gia_Obj_t * pObj )
func (pNew *GiaManager) giaManBalance_rec(p *GiaManager, pObj *GiaObject) {
	panic("TODO, stubbed")
}

// Gia_ManBalanceInt (static inline)
// C: Gia_Man_t * Gia_ManBalanceInt( Gia_Man_t * p, int nLutSize, int nCutNum, int fVerbose )
func (p *GiaManager) giaManBalanceInt(nLutSize int, nCutNum int, fVerbose int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManBalanceLut
// C: Gia_Man_t * Gia_ManBalanceLut( Gia_Man_t * p, int nLutSize, int nCutNum, int fVerbose )
func (p *GiaManager) GiaManBalanceLut(nLutSize int, nCutNum int, fVerbose int) *GiaManager {
	panic("TODO, stubbed")
}
