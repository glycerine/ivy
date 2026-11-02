package abc

// Hand-checked skeleton of src/aig/gia/giaShrink7.c.

// UnmMan is Unm_Man_t.
type UnmMan struct {
	pGia     *GiaManager   // user's AIG
	pNew     *GiaManager   // constructed AIG
	pHash    *IntTupleHash // Hash_IntMan_t * : hash table
	nNewSize int           // expected size of new manager
	vUsed    *IntVector    // used nodes
	vId2Used *IntVector    // mapping of obj IDs into used node IDs
	vTruths  *WordVector   // truth tables
	vLeaves  *IntVector    // temporary storage for leaves
	clkStart int64         // abctime : starting the clock
}

// Unm_ManAlloc
// C: Unm_Man_t * Unm_ManAlloc( Gia_Man_t * pGia )
func (pGia *GiaManager) UnmManAlloc() *UnmMan {
	panic("TODO, stubbed")
}

// Unm_ManFree
// C: Gia_Man_t * Unm_ManFree( Unm_Man_t * p )
func (p *UnmMan) UnmManFree() *GiaManager {
	panic("TODO, stubbed")
}

// Unm_ManPrintPairStats
// C: int Unm_ManPrintPairStats( Hash_IntMan_t * pHash, int nTotal0, int nPairs0, int nPairs1, int fUseLit )
func UnmManPrintPairStats(pHash *IntTupleHash, nTotal0 int, nPairs0 int, nPairs1 int, fUseLit int) int {
	panic("TODO, stubbed")
}

// Unm_ManComputePairs
// C: Vec_Int_t * Unm_ManComputePairs( Unm_Man_t * p, int fVerbose )
func (p *UnmMan) UnmManComputePairs(fVerbose int) *IntVector {
	panic("TODO, stubbed")
}

// Unm_ManFindUsedNodes
// C: Vec_Int_t * Unm_ManFindUsedNodes( Vec_Int_t * vPairs, int nObjs )
func UnmManFindUsedNodes(vPairs *IntVector, nObjs int) *IntVector {
	panic("TODO, stubbed")
}

// Unm_ManComputeTruths
// C: Vec_Wrd_t * Unm_ManComputeTruths( Unm_Man_t * p )
func (p *UnmMan) UnmManComputeTruths() *WordVector {
	panic("TODO, stubbed")
}

// Unm_ManCollectDecomp
// C: Vec_Int_t * Unm_ManCollectDecomp( Unm_Man_t * p, Vec_Int_t * vPairs, int fVerbose )
func (p *UnmMan) UnmManCollectDecomp(vPairs *IntVector, fVerbose int) *IntVector {
	panic("TODO, stubbed")
}

// Unm_ManWork
// C: void Unm_ManWork( Unm_Man_t * p )
func (p *UnmMan) UnmManWork() {
	panic("TODO, stubbed")
}

// Unm_ManTest
// C: Gia_Man_t * Unm_ManTest( Gia_Man_t * pGia )
func (pGia *GiaManager) UnmManTest() *GiaManager {
	panic("TODO, stubbed")
}
