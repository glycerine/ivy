package abc

// Hand-checked skeleton of src/aig/saig/saigSimMv.c.

const (
	SAIG_DIFF_VALUES = 8
	SAIG_UNDEF_VALUE = 0x1ffffffe // 536870910
)

// SaigMvObj is Saig_MvObj_t: old AIG.
type SaigMvObj struct {
	iFan0 int
	iFan1 int
	Type  uint32 // unsigned : 3
	Value uint32 // unsigned : 29
}

// SaigMvAnd is Saig_MvAnd_t: new AIG.
type SaigMvAnd struct {
	iFan0 int
	iFan1 int
	iNext int
}

// SaigMvMan is Saig_MvMan_t: simulation manager.
type SaigMvMan struct {
	// user data
	pAig *AigManager // original AIG
	// parameters
	nStatesMax int // maximum number of states
	nLevelsMax int // maximum number of levels
	nValuesMax int // maximum number of values
	nFlops     int // number of flops
	// compacted AIG
	pAigOld      []SaigMvObj               // Saig_MvObj_t * : AIG objects
	vFlops       *PointerVector[SaigMvObj] // collected flops
	vXFlops      *IntVector                // flops that had at least one X-value
	vTired       *PointerVector[any]       // collected flops (entries are never added in the source)
	pTStates     []uint32                  // unsigned * : hash table for states
	nTStatesSize int                       // hash table size
	pMemStates   *AigFixedArena            // Aig_MmFixed_t * : memory for states
	vStates      *PointerVector[uint32]    // reached states
	pRegsUndef   []int                     // int * : count the number of undef values
	pRegsValues  [][]int                   // int ** : write the first different values
	nRegsValues  []int                     // int * : count the number of different values
	nRUndefs     int                       // the number of undef registers
	nRValues     [SAIG_DIFF_VALUES + 1]int // the number of registers with given values
	// internal AIG
	pAigNew     []SaigMvAnd // Saig_MvAnd_t * : AIG nodes
	nObjsAlloc  int         // the number of objects allocated
	nObjs       int         // the number of objects
	nPis        int         // the number of primary inputs
	pTNodes     []int       // int * : hash table
	nTNodesSize int         // hash table size
	pLevels     []uint8     // unsigned char * : levels of AIG nodes
}

// Saig_MvObjFaninC0 (static inline)
// C: int Saig_MvObjFaninC0( Saig_MvObj_t * pObj )
func saigMvObjFaninC0(pObj *SaigMvObj) int {
	panic("TODO, stubbed")
}

// Saig_MvObjFaninC1 (static inline)
// C: int Saig_MvObjFaninC1( Saig_MvObj_t * pObj )
func saigMvObjFaninC1(pObj *SaigMvObj) int {
	panic("TODO, stubbed")
}

// Saig_MvObjFanin0 (static inline)
// C: int Saig_MvObjFanin0( Saig_MvObj_t * pObj )
func saigMvObjFanin0(pObj *SaigMvObj) int {
	panic("TODO, stubbed")
}

// Saig_MvObjFanin1 (static inline)
// C: int Saig_MvObjFanin1( Saig_MvObj_t * pObj )
func saigMvObjFanin1(pObj *SaigMvObj) int {
	panic("TODO, stubbed")
}

// Saig_MvConst0 (static inline)
// C: int Saig_MvConst0()
func saigMvConst0() int {
	panic("TODO, stubbed")
}

// Saig_MvConst1 (static inline)
// C: int Saig_MvConst1()
func saigMvConst1() int {
	panic("TODO, stubbed")
}

// Saig_MvConst (static inline)
// C: int Saig_MvConst( int c )
func saigMvConst(c int) int {
	panic("TODO, stubbed")
}

// Saig_MvUndef (static inline)
// C: int Saig_MvUndef()
func saigMvUndef() int {
	panic("TODO, stubbed")
}

// Saig_MvIsConst0 (static inline)
// C: int Saig_MvIsConst0( int iNode )
func saigMvIsConst0(iNode int) int {
	panic("TODO, stubbed")
}

// Saig_MvIsConst1 (static inline)
// C: int Saig_MvIsConst1( int iNode )
func saigMvIsConst1(iNode int) int {
	panic("TODO, stubbed")
}

// Saig_MvIsConst (static inline)
// C: int Saig_MvIsConst( int iNode )
func saigMvIsConst(iNode int) int {
	panic("TODO, stubbed")
}

// Saig_MvIsUndef (static inline)
// C: int Saig_MvIsUndef( int iNode )
func saigMvIsUndef(iNode int) int {
	panic("TODO, stubbed")
}

// Saig_MvRegular (static inline)
// C: int Saig_MvRegular( int iNode )
func saigMvRegular(iNode int) int {
	panic("TODO, stubbed")
}

// Saig_MvNot (static inline)
// C: int Saig_MvNot( int iNode )
func saigMvNot(iNode int) int {
	panic("TODO, stubbed")
}

// Saig_MvNotCond (static inline)
// C: int Saig_MvNotCond( int iNode, int c )
func saigMvNotCond(iNode int, c int) int {
	panic("TODO, stubbed")
}

// Saig_MvIsComplement (static inline)
// C: int Saig_MvIsComplement( int iNode )
func saigMvIsComplement(iNode int) int {
	panic("TODO, stubbed")
}

// Saig_MvLit2Var (static inline)
// C: int Saig_MvLit2Var( int iNode )
func saigMvLit2Var(iNode int) int {
	panic("TODO, stubbed")
}

// Saig_MvVar2Lit (static inline)
// C: int Saig_MvVar2Lit( int iVar )
func saigMvVar2Lit(iVar int) int {
	panic("TODO, stubbed")
}

// Saig_MvLev (static inline)
// C: int Saig_MvLev( Saig_MvMan_t * p, int iNode )
func (p *SaigMvMan) saigMvLev(iNode int) int {
	panic("TODO, stubbed")
}

// Saig_ManCreateReducedAig
// C: Saig_MvObj_t * Saig_ManCreateReducedAig( Aig_Man_t * p, Vec_Ptr_t ** pvFlops )
func (p *AigManager) SaigManCreateReducedAig(pvFlops **PointerVector[SaigMvObj]) []SaigMvObj {
	panic("TODO, stubbed")
}

// Saig_MvCreateObj (static inline)
// C: int Saig_MvCreateObj( Saig_MvMan_t * p, int iFan0, int iFan1 )
func (p *SaigMvMan) saigMvCreateObj(iFan0 int, iFan1 int) int {
	panic("TODO, stubbed")
}

// Saig_MvManStart
// C: Saig_MvMan_t * Saig_MvManStart( Aig_Man_t * pAig, int nFramesSatur )
func (pAig *AigManager) SaigMvManStart(nFramesSatur int) *SaigMvMan {
	panic("TODO, stubbed")
}

// Saig_MvManStop
// C: void Saig_MvManStop( Saig_MvMan_t * p )
func (p *SaigMvMan) SaigMvManStop() {
	panic("TODO, stubbed")
}

// Saig_MvHash (static inline)
// C: int Saig_MvHash( int iFan0, int iFan1, int TableSize )
func saigMvHash(iFan0 int, iFan1 int, TableSize int) int {
	panic("TODO, stubbed")
}

// Saig_MvTableFind (static inline)
// C: int * Saig_MvTableFind( Saig_MvMan_t * p, int iFan0, int iFan1 )
func (p *SaigMvMan) saigMvTableFind(iFan0 int, iFan1 int) *int {
	panic("TODO, stubbed")
}

// Saig_MvAnd (static inline)
// C: int Saig_MvAnd( Saig_MvMan_t * p, int iFan0, int iFan1, int fFirst )
func (p *SaigMvMan) saigMvAnd(iFan0 int, iFan1 int, fFirst int) int {
	panic("TODO, stubbed")
}

// Saig_MvSimulateValue0 (static inline)
// C: int Saig_MvSimulateValue0( Saig_MvObj_t * pAig, Saig_MvObj_t * pObj )
func saigMvSimulateValue0(pAig *SaigMvObj, pObj *SaigMvObj) int {
	panic("TODO, stubbed")
}

// Saig_MvSimulateValue1 (static inline)
// C: int Saig_MvSimulateValue1( Saig_MvObj_t * pAig, Saig_MvObj_t * pObj )
func saigMvSimulateValue1(pAig *SaigMvObj, pObj *SaigMvObj) int {
	panic("TODO, stubbed")
}

// Saig_MvPrintState
// C: void Saig_MvPrintState( int Iter, Saig_MvMan_t * p )
func SaigMvPrintState(Iter int, p *SaigMvMan) {
	panic("TODO, stubbed")
}

// Saig_MvSimulateFrame
// C: void Saig_MvSimulateFrame( Saig_MvMan_t * p, int fFirst, int fVerbose )
func (p *SaigMvMan) SaigMvSimulateFrame(fFirst int, fVerbose int) {
	panic("TODO, stubbed")
}

// Saig_MvSimHash
// C: int Saig_MvSimHash( unsigned * pState, int nFlops, int TableSize )
func SaigMvSimHash(pState []uint32, nFlops int, TableSize int) int {
	panic("TODO, stubbed")
}

// Saig_MvSimTableFind (static inline; the result points into the hash table or to a state)
// C: unsigned * Saig_MvSimTableFind( Saig_MvMan_t * p, unsigned * pState )
func (p *SaigMvMan) saigMvSimTableFind(pState []uint32) *uint32 {
	panic("TODO, stubbed")
}

// Saig_MvSaveState
// C: int Saig_MvSaveState( Saig_MvMan_t * p )
func (p *SaigMvMan) SaigMvSaveState() int {
	panic("TODO, stubbed")
}

// Saig_MvManPostProcess
// C: void Saig_MvManPostProcess( Saig_MvMan_t * p, int iState )
func (p *SaigMvMan) SaigMvManPostProcess(iState int) {
	panic("TODO, stubbed")
}

// Saig_MvManFindXFlops
// C: Vec_Int_t * Saig_MvManFindXFlops( Saig_MvMan_t * p )
func (p *SaigMvMan) SaigMvManFindXFlops() *IntVector {
	panic("TODO, stubbed")
}

// Saig_MvManCheckOscilator
// C: int Saig_MvManCheckOscilator( Saig_MvMan_t * p, int iFlop )
func (p *SaigMvMan) SaigMvManCheckOscilator(iFlop int) int {
	panic("TODO, stubbed")
}

// Saig_MvManFindConstBinaryFlops
// C: Vec_Int_t * Saig_MvManFindConstBinaryFlops( Saig_MvMan_t * p, Vec_Int_t ** pvBinary )
func (p *SaigMvMan) SaigMvManFindConstBinaryFlops(pvBinary **IntVector) *IntVector {
	panic("TODO, stubbed")
}

// Saig_MvManFindOscilators
// C: Vec_Int_t * Saig_MvManFindOscilators( Saig_MvMan_t * p, Vec_Int_t ** pvConst0 )
func (p *SaigMvMan) SaigMvManFindOscilators(pvConst0 **IntVector) *IntVector {
	panic("TODO, stubbed")
}

// Saig_MvManCreateNextSkip
// C: Vec_Int_t * Saig_MvManCreateNextSkip( Saig_MvMan_t * p )
func (p *SaigMvMan) SaigMvManCreateNextSkip() *IntVector {
	panic("TODO, stubbed")
}

// Saig_MvManDeriveMap
// C: Vec_Ptr_t * Saig_MvManDeriveMap( Saig_MvMan_t * p, int fVerbose )
func (p *SaigMvMan) SaigMvManDeriveMap(fVerbose int) *PointerVector[AigObject] {
	panic("TODO, stubbed")
}

// Saig_MvManSimulate
// C: Vec_Ptr_t * Saig_MvManSimulate( Aig_Man_t * pAig, int nFramesSymb, int nFramesSatur, int fVerbose, int fVeryVerbose )
func (pAig *AigManager) SaigMvManSimulate(nFramesSymb int, nFramesSatur int, fVerbose int, fVeryVerbose int) *PointerVector[AigObject] {
	panic("TODO, stubbed")
}
