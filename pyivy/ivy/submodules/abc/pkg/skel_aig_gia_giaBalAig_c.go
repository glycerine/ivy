package abc

// Hand-checked skeleton of src/aig/gia/giaBalAig.c.

// DamMan is Dam_Man_t: the operation manager.
type DamMan struct {
	pGia      *GiaManager         // user's AIG
	vNod2Set  *IntVector          // node ID into fanin set
	vDiv2Nod  *IntVector          // div ID into root node set
	vSetStore *IntVector          // fanin set storage
	vNodStore *IntVector          // root node set storage
	vCounts   *FloatVector        // occur counts
	vNodLevR  *IntVector          // node reverse level
	vDivLevR  *IntVector          // divisor reverse level
	vVisit    *IntVector          // visited MUXes
	vQue      *FloatPriorityQueue // pairs by their weight
	vHash     *IntTupleHash       // Hash_IntMan_t * : pair hash table
	clkStart  int64               // abctime : starting the clock
	nLevelMax int                 // maximum level
	nDivs     int                 // extracted divisor count
	nAnds     int                 // total AND node count
	nGain     int                 // total gain in AND nodes
	nGainX    int                 // gain from XOR nodes
}

// Dam_ObjHand (static inline)
// C: int Dam_ObjHand( Dam_Man_t * p, int i )
func (p *DamMan) damObjHand(i int) int {
	panic("TODO, stubbed")
}

// Dam_ObjSet (static inline)
// C: int * Dam_ObjSet( Dam_Man_t * p, int i )
func (p *DamMan) damObjSet(i int) []int {
	panic("TODO, stubbed")
}

// Dam_DivHand (static inline)
// C: int Dam_DivHand( Dam_Man_t * p, int d )
func (p *DamMan) damDivHand(d int) int {
	panic("TODO, stubbed")
}

// Dam_DivSet (static inline)
// C: int * Dam_DivSet( Dam_Man_t * p, int d )
func (p *DamMan) damDivSet(d int) []int {
	panic("TODO, stubbed")
}

// Gia_ManSimplifyXor
// C: void Gia_ManSimplifyXor( Vec_Int_t * vSuper )
func GiaManSimplifyXor(vSuper *IntVector) {
	panic("TODO, stubbed")
}

// Gia_ManSimplifyAnd
// C: void Gia_ManSimplifyAnd( Vec_Int_t * vSuper )
func GiaManSimplifyAnd(vSuper *IntVector) {
	panic("TODO, stubbed")
}

// Gia_ManSuperCollectXor_rec
// C: void Gia_ManSuperCollectXor_rec( Gia_Man_t * p, Gia_Obj_t * pObj, int fStrict )
func (p *GiaManager) GiaManSuperCollectXor_rec(pObj *GiaObject, fStrict int) {
	panic("TODO, stubbed")
}

// Gia_ManSuperCollectAnd_rec
// C: void Gia_ManSuperCollectAnd_rec( Gia_Man_t * p, Gia_Obj_t * pObj, int fStrict )
func (p *GiaManager) GiaManSuperCollectAnd_rec(pObj *GiaObject, fStrict int) {
	panic("TODO, stubbed")
}

// Gia_ManSuperCollect
// C: void Gia_ManSuperCollect( Gia_Man_t * p, Gia_Obj_t * pObj, int fStrict )
func (p *GiaManager) GiaManSuperCollect(pObj *GiaObject, fStrict int) {
	panic("TODO, stubbed")
}

// Gia_ManFindSharedNode
// C: int Gia_ManFindSharedNode( Gia_Man_t * pNew, Vec_Int_t * vSuper, int iLit0 )
func (pNew *GiaManager) GiaManFindSharedNode(vSuper *IntVector, iLit0 int) int {
	panic("TODO, stubbed")
}

// Gia_ManPrepareLastTwo
// C: void Gia_ManPrepareLastTwo( Gia_Man_t * pNew, Vec_Int_t * vSuper )
func (pNew *GiaManager) GiaManPrepareLastTwo(vSuper *IntVector) {
	panic("TODO, stubbed")
}

// Gia_ManCreateGate
// C: void Gia_ManCreateGate( Gia_Man_t * pNew, Gia_Obj_t * pObj, Vec_Int_t * vSuper )
func (pNew *GiaManager) GiaManCreateGate(pObj *GiaObject, vSuper *IntVector) {
	panic("TODO, stubbed")
}

// Gia_ManBalanceGate
// C: int Gia_ManBalanceGate( Gia_Man_t * pNew, Gia_Obj_t * pObj, Vec_Int_t * vSuper, int * pLits, int nLits )
func (pNew *GiaManager) GiaManBalanceGate(pObj *GiaObject, vSuper *IntVector, pLits []int, nLits int) int {
	panic("TODO, stubbed")
}

// Gia_ManBalance_rec
// C: void Gia_ManBalance_rec( Gia_Man_t * pNew, Gia_Man_t * p, Gia_Obj_t * pObj, int fStrict )
func (pNew *GiaManager) GiaManBalance_rec(p *GiaManager, pObj *GiaObject, fStrict int) {
	panic("TODO, stubbed")
}

// Gia_ManBalanceInt
// C: Gia_Man_t * Gia_ManBalanceInt( Gia_Man_t * p, int fStrict )
func (p *GiaManager) GiaManBalanceInt(fStrict int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManBalance
// C: Gia_Man_t * Gia_ManBalance( Gia_Man_t * p, int fSimpleAnd, int fStrict, int fVerbose )
func (p *GiaManager) GiaManBalance(fSimpleAnd int, fStrict int, fVerbose int) *GiaManager {
	panic("TODO, stubbed")
}

// Dam_ManAlloc
// C: Dam_Man_t * Dam_ManAlloc( Gia_Man_t * pGia )
func (pGia *GiaManager) DamManAlloc() *DamMan {
	panic("TODO, stubbed")
}

// Dam_ManFree
// C: void Dam_ManFree( Dam_Man_t * p )
func (p *DamMan) DamManFree() {
	panic("TODO, stubbed")
}

// Dam_ManCollectSets_rec
// C: void Dam_ManCollectSets_rec( Dam_Man_t * p, int Id )
func (p *DamMan) DamManCollectSets_rec(Id int) {
	panic("TODO, stubbed")
}

// Dam_ManCollectSets
// C: void Dam_ManCollectSets( Dam_Man_t * p )
func (p *DamMan) DamManCollectSets() {
	panic("TODO, stubbed")
}

// Dam_ManDivSlack
// C: int Dam_ManDivSlack( Dam_Man_t * p, int iLit0, int iLit1, int LevR )
func (p *DamMan) DamManDivSlack(iLit0 int, iLit1 int, LevR int) int {
	panic("TODO, stubbed")
}

// Dam_ManCreateMultiRefs
// C: void Dam_ManCreateMultiRefs( Dam_Man_t * p, Vec_Int_t ** pvRefsAnd, Vec_Int_t ** pvRefsXor )
func (p *DamMan) DamManCreateMultiRefs(pvRefsAnd **IntVector, pvRefsXor **IntVector) {
	panic("TODO, stubbed")
}

// Dam_ManCreatePairs
// C: void Dam_ManCreatePairs( Dam_Man_t * p, int fVerbose )
func (p *DamMan) DamManCreatePairs(fVerbose int) {
	panic("TODO, stubbed")
}

// Dam_ManMultiAig_rec
// C: void Dam_ManMultiAig_rec( Dam_Man_t * pMan, Gia_Man_t * pNew, Gia_Man_t * p, Gia_Obj_t * pObj )
func (pMan *DamMan) DamManMultiAig_rec(pNew *GiaManager, p *GiaManager, pObj *GiaObject) {
	panic("TODO, stubbed")
}

// Dam_ManMultiAig
// C: Gia_Man_t * Dam_ManMultiAig( Dam_Man_t * pMan )
func (pMan *DamMan) DamManMultiAig() *GiaManager {
	panic("TODO, stubbed")
}

// Dam_PrintDiv
// C: void Dam_PrintDiv( Dam_Man_t * p, int iDiv )
func (p *DamMan) DamPrintDiv(iDiv int) {
	panic("TODO, stubbed")
}

// Dam_PrintQue
// C: void Dam_PrintQue( Dam_Man_t * p )
func (p *DamMan) DamPrintQue() {
	panic("TODO, stubbed")
}

// Dam_ManUpdateNode
// C: int Dam_ManUpdateNode( Dam_Man_t * p, int iObj, int iLit0, int iLit1, int iLitNew, Vec_Int_t * vDivs )
func (p *DamMan) DamManUpdateNode(iObj int, iLit0 int, iLit1 int, iLitNew int, vDivs *IntVector) int {
	panic("TODO, stubbed")
}

// Dam_ManUpdate
// C: void Dam_ManUpdate( Dam_Man_t * p, int iDiv )
func (p *DamMan) DamManUpdate(iDiv int) {
	panic("TODO, stubbed")
}

// Dam_ManAreaBalanceInt
// C: Gia_Man_t * Dam_ManAreaBalanceInt( Gia_Man_t * pGia, Vec_Int_t * vCiLevels, int nNewNodesMax, int fVerbose, int fVeryVerbose )
func (pGia *GiaManager) DamManAreaBalanceInt(vCiLevels *IntVector, nNewNodesMax int, fVerbose int, fVeryVerbose int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManAreaBalance
// C: Gia_Man_t * Gia_ManAreaBalance( Gia_Man_t * p, int fSimpleAnd, int nNewNodesMax, int fVerbose, int fVeryVerbose )
func (p *GiaManager) GiaManAreaBalance(fSimpleAnd int, nNewNodesMax int, fVerbose int, fVeryVerbose int) *GiaManager {
	panic("TODO, stubbed")
}
