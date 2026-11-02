package abc

// Hand-checked skeleton of src/aig/gia/giaShrink6.c.

// ShrFan is Shr_Fan_t.
type ShrFan struct {
	iFan int // fanout ID
	Next int // next structure
}

// ShrMan is Shr_Man_t.
type ShrMan struct {
	pGia       *GiaManager // user's AIG
	pNew       *GiaManager // constructed AIG
	nDivMax    int         // max number of divisors
	nNewSize   int         // max growth size
	vFanMem    *WordVector // fanout memory
	vObj2Fan   *IntVector  // fanout
	pFanTemp   *ShrFan     // temporary fanout
	vDivs      *IntVector  // divisors
	vPrio      *IntVector  // priority queue
	vDivResub  *IntVector  // resubstitution
	vLeaves    *IntVector  // cut leaves
	vTruths    *WordVector // truth tables
	vDivTruths *WordVector // truth tables
	pManRsb    *RsbManager // Rsb_Man_t *
	pManDec    *BdcTable   // Bdc_Man_t *
	Pars       BdcPar      // Bdc_Par_t (by value)
}

// Shr_ManAlloc
// C: Shr_Man_t * Shr_ManAlloc( Gia_Man_t * pGia )
func (pGia *GiaManager) ShrManAlloc() *ShrMan {
	panic("TODO, stubbed")
}

// Shr_ManFree
// C: Gia_Man_t * Shr_ManFree( Shr_Man_t * p )
func (p *ShrMan) ShrManFree() *GiaManager {
	panic("TODO, stubbed")
}

// Shr_ManAddFanout (static inline)
// C: void Shr_ManAddFanout( Shr_Man_t * p, int iFanin, int iFanout )
func (p *ShrMan) shrManAddFanout(iFanin int, iFanout int) {
	panic("TODO, stubbed")
}

// Shr_ManFanIterStart (static inline)
// C: int Shr_ManFanIterStart( Shr_Man_t * p, int iNode )
func (p *ShrMan) shrManFanIterStart(iNode int) int {
	panic("TODO, stubbed")
}

// Shr_ManFanIterNext (static inline)
// C: int Shr_ManFanIterNext( Shr_Man_t * p )
func (p *ShrMan) shrManFanIterNext() int {
	panic("TODO, stubbed")
}

// Shr_ManDivPushOrderByLevel (static inline)
// C: int Shr_ManDivPushOrderByLevel( Shr_Man_t * p, int iDiv )
func (p *ShrMan) shrManDivPushOrderByLevel(iDiv int) int {
	panic("TODO, stubbed")
}

// Shr_ManCollectDivisors (static inline)
// C: int Shr_ManCollectDivisors( Shr_Man_t * p, Vec_Int_t * vLeaves, int Limit, int nFanoutMax )
func (p *ShrMan) shrManCollectDivisors(vLeaves *IntVector, Limit int, nFanoutMax int) int {
	panic("TODO, stubbed")
}

// Shr_ObjPerformBidec
// C: int Shr_ObjPerformBidec( Shr_Man_t * p, Bdc_Man_t * pManDec, Gia_Man_t * pNew, Vec_Int_t * vLeafLits, word uTruth1, word uTruthC )
func (p *ShrMan) ShrObjPerformBidec(pManDec *BdcTable, pNew *GiaManager, vLeafLits *IntVector, uTruth1 uint64, uTruthC uint64) int {
	panic("TODO, stubbed")
}

// Shr_ManComputeTruth6_rec
// C: word Shr_ManComputeTruth6_rec( Gia_Man_t * p, int iNode, Vec_Wrd_t * vTruths )
func (p *GiaManager) ShrManComputeTruth6_rec(iNode int, vTruths *WordVector) uint64 {
	panic("TODO, stubbed")
}

// Shr_ManComputeTruth6
// C: word Shr_ManComputeTruth6( Gia_Man_t * p, Gia_Obj_t * pObj, Vec_Int_t * vLeaves, Vec_Wrd_t * vTruths )
func (p *GiaManager) ShrManComputeTruth6(pObj *GiaObject, vLeaves *IntVector, vTruths *WordVector) uint64 {
	panic("TODO, stubbed")
}

// Shr_ManComputeTruths
// C: void Shr_ManComputeTruths( Gia_Man_t * p, int nVars, Vec_Int_t * vDivs, Vec_Wrd_t * vDivTruths, Vec_Wrd_t * vTruths )
func (p *GiaManager) ShrManComputeTruths(nVars int, vDivs *IntVector, vDivTruths *WordVector, vTruths *WordVector) {
	panic("TODO, stubbed")
}

// Gia_ManMapShrink6
// C: Gia_Man_t * Gia_ManMapShrink6( Gia_Man_t * p, int nFanoutMax, int fKeepLevel, int fVerbose )
func (p *GiaManager) GiaManMapShrink6(nFanoutMax int, fKeepLevel int, fVerbose int) *GiaManager {
	panic("TODO, stubbed")
}
