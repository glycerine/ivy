package abc

// Hand-checked skeleton of src/base/abci/abcOdc.c.

const ABC_DC_MAX_NODES = 1 << 15

// OdcLit is Odc_Lit_t: typedef unsigned short Odc_Lit_t.
type OdcLit = uint16

// OdcObj is Odc_Obj_t.
type OdcObj struct {
	iFan0  OdcLit // first fanin
	iFan1  OdcLit // second fanin
	iNext  OdcLit // next node in the hash table
	TravId uint16 // the traversal ID
	uData  uint32 // the computed data
	uMask  uint32 // the variable mask
}

// OdcMan is Odc_Man_t.
type OdcMan struct {
	nVarsMax     int                              // the max number of cut variables
	nLevels      int                              // the number of ODC levels
	fVerbose     int                              // the verbosiness flag
	fVeryVerbose int                              // the verbosiness flag to print per-node stats
	nPercCutoff  int                              // cutoff percentage
	pNode        *ABCNetworkObject                // the node for windowing
	vLeaves      *PointerVector[ABCNetworkObject] // the number of the cut
	vRoots       *PointerVector[ABCNetworkObject] // the roots of the cut
	vBranches    *PointerVector[ABCNetworkObject] // additional inputs
	nPis         int                              // number of PIs (nVarsMax + 32)
	nObjs        int                              // number of objects (Const1, PIs, ANDs)
	nObjsAlloc   int                              // number of objects allocated
	pObjs        []OdcObj                         // Odc_Obj_t * : objects
	iRoot        OdcLit                           // the root object
	nTravIds     uint16                           // the number of travIDs
	pTable       []OdcLit                         // Odc_Lit_t * : hash table
	nTableSize   int                              // hash table size
	vUsedSpots   *IntVector                       // the used spots
	nBits        int                              // the number of bits
	nWords       int                              // the number of words
	vTruths      *PointerVector[uint32]           // truth tables for each node
	vTruthsElem  *PointerVector[uint32]           // elementary truth tables for the PIs
	puTruth      []uint32                         // unsigned * : the place where the resulting truth table goes
	nWins        int                              // the number of windows processed
	nWinsEmpty   int                              // the number of empty windows
	nSimsEmpty   int                              // the number of empty simulation infos
	nQuantsOver  int                              // the number of quantification overflows
	nWinsFinish  int                              // the number of windows that finished
	nTotalDcs    int                              // total percentage of DCs
	timeClean    int64                            // abctime : windowing
	timeWin      int64                            // abctime : windowing
	timeMiter    int64                            // abctime : computing the miter
	timeSim      int64                            // abctime : simulation
	timeQuant    int64                            // abctime : quantification
	timeTruth    int64                            // abctime : truth table
	timeTotal    int64                            // abctime : useful runtime
	timeAbort    int64                            // abctime : aborted runtime
}

// Odc_PiNum (static)
// C: int Odc_PiNum( Odc_Man_t * p )
func (p *OdcMan) odcPiNum() int {
	panic("TODO, stubbed")
}

// Odc_NodeNum (static)
// C: int Odc_NodeNum( Odc_Man_t * p )
func (p *OdcMan) odcNodeNum() int {
	panic("TODO, stubbed")
}

// Odc_ObjNum (static)
// C: int Odc_ObjNum( Odc_Man_t * p )
func (p *OdcMan) odcObjNum() int {
	panic("TODO, stubbed")
}

// Odc_IsComplement (static)
// C: int Odc_IsComplement( Odc_Lit_t Lit )
func odcIsComplement(Lit OdcLit) int {
	panic("TODO, stubbed")
}

// Odc_Regular (static)
// C: Odc_Lit_t Odc_Regular( Odc_Lit_t Lit )
func odcRegular(Lit OdcLit) OdcLit {
	panic("TODO, stubbed")
}

// Odc_Not (static)
// C: Odc_Lit_t Odc_Not( Odc_Lit_t Lit )
func odcNot(Lit OdcLit) OdcLit {
	panic("TODO, stubbed")
}

// Odc_NotCond (static)
// C: Odc_Lit_t Odc_NotCond( Odc_Lit_t Lit, int c )
func odcNotCond(Lit OdcLit, c int) OdcLit {
	panic("TODO, stubbed")
}

// Odc_Const0 (static)
// C: Odc_Lit_t Odc_Const0()
func odcConst0() OdcLit {
	panic("TODO, stubbed")
}

// Odc_Const1 (static)
// C: Odc_Lit_t Odc_Const1()
func odcConst1() OdcLit {
	panic("TODO, stubbed")
}

// Odc_Var (static)
// C: Odc_Lit_t Odc_Var( Odc_Man_t * p, int i )
func (p *OdcMan) odcVar(i int) OdcLit {
	panic("TODO, stubbed")
}

// Odc_IsConst (static)
// C: int Odc_IsConst( Odc_Lit_t Lit )
func odcIsConst(Lit OdcLit) int {
	panic("TODO, stubbed")
}

// Odc_IsTerm (static)
// C: int Odc_IsTerm( Odc_Man_t * p, Odc_Lit_t Lit )
func (p *OdcMan) odcIsTerm(Lit OdcLit) int {
	panic("TODO, stubbed")
}

// Odc_ObjNew (static)
// C: Odc_Obj_t * Odc_ObjNew( Odc_Man_t * p )
func (p *OdcMan) odcObjNew() *OdcObj {
	panic("TODO, stubbed")
}

// Odc_Obj2Lit (static)
// C: Odc_Lit_t Odc_Obj2Lit( Odc_Man_t * p, Odc_Obj_t * pObj )
func (p *OdcMan) odcObj2Lit(pObj *OdcObj) OdcLit {
	panic("TODO, stubbed")
}

// Odc_Lit2Obj (static)
// C: Odc_Obj_t * Odc_Lit2Obj( Odc_Man_t * p, Odc_Lit_t Lit )
func (p *OdcMan) odcLit2Obj(Lit OdcLit) *OdcObj {
	panic("TODO, stubbed")
}

// Odc_ObjChild0 (static)
// C: Odc_Lit_t Odc_ObjChild0( Odc_Obj_t * pObj )
func odcObjChild0(pObj *OdcObj) OdcLit {
	panic("TODO, stubbed")
}

// Odc_ObjChild1 (static)
// C: Odc_Lit_t Odc_ObjChild1( Odc_Obj_t * pObj )
func odcObjChild1(pObj *OdcObj) OdcLit {
	panic("TODO, stubbed")
}

// Odc_ObjFanin0 (static)
// C: Odc_Lit_t Odc_ObjFanin0( Odc_Obj_t * pObj )
func odcObjFanin0(pObj *OdcObj) OdcLit {
	panic("TODO, stubbed")
}

// Odc_ObjFanin1 (static)
// C: Odc_Lit_t Odc_ObjFanin1( Odc_Obj_t * pObj )
func odcObjFanin1(pObj *OdcObj) OdcLit {
	panic("TODO, stubbed")
}

// Odc_ObjFaninC0 (static)
// C: int Odc_ObjFaninC0( Odc_Obj_t * pObj )
func odcObjFaninC0(pObj *OdcObj) int {
	panic("TODO, stubbed")
}

// Odc_ObjFaninC1 (static)
// C: int Odc_ObjFaninC1( Odc_Obj_t * pObj )
func odcObjFaninC1(pObj *OdcObj) int {
	panic("TODO, stubbed")
}

// Odc_ManIncrementTravId (static)
// C: void Odc_ManIncrementTravId( Odc_Man_t * p )
func (p *OdcMan) odcManIncrementTravId() {
	panic("TODO, stubbed")
}

// Odc_ObjSetTravIdCurrent (static)
// C: void Odc_ObjSetTravIdCurrent( Odc_Man_t * p, Odc_Obj_t * pObj )
func (p *OdcMan) odcObjSetTravIdCurrent(pObj *OdcObj) {
	panic("TODO, stubbed")
}

// Odc_ObjIsTravIdCurrent (static)
// C: int Odc_ObjIsTravIdCurrent( Odc_Man_t * p, Odc_Obj_t * pObj )
func (p *OdcMan) odcObjIsTravIdCurrent(pObj *OdcObj) int {
	panic("TODO, stubbed")
}

// Odc_ObjTruth (static)
// C: unsigned * Odc_ObjTruth( Odc_Man_t * p, Odc_Lit_t Lit )
func (p *OdcMan) odcObjTruth(Lit OdcLit) []uint32 {
	panic("TODO, stubbed")
}

// Abc_NtkDontCareAlloc
// C: Odc_Man_t * Abc_NtkDontCareAlloc( int nVarsMax, int nLevels, int fVerbose, int fVeryVerbose )
func AbcNtkDontCareAlloc(nVarsMax int, nLevels int, fVerbose int, fVeryVerbose int) *OdcMan {
	panic("TODO, stubbed")
}

// Abc_NtkDontCareClear
// C: void Abc_NtkDontCareClear( Odc_Man_t * p )
func (p *OdcMan) AbcNtkDontCareClear() {
	panic("TODO, stubbed")
}

// Abc_NtkDontCareFree
// C: void Abc_NtkDontCareFree( Odc_Man_t * p )
func (p *OdcMan) AbcNtkDontCareFree() {
	panic("TODO, stubbed")
}

// Abc_NtkDontCareWinSweepLeafTfo_rec
// C: void Abc_NtkDontCareWinSweepLeafTfo_rec( Abc_Obj_t * pObj, int nLevelLimit, Abc_Obj_t * pNode )
func AbcNtkDontCareWinSweepLeafTfo_rec(pObj *ABCNetworkObject, nLevelLimit int, pNode *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Abc_NtkDontCareWinSweepLeafTfo
// C: void Abc_NtkDontCareWinSweepLeafTfo( Odc_Man_t * p )
func (p *OdcMan) AbcNtkDontCareWinSweepLeafTfo() {
	panic("TODO, stubbed")
}

// Abc_NtkDontCareWinCollectRoots_rec
// C: void Abc_NtkDontCareWinCollectRoots_rec( Abc_Obj_t * pObj, Vec_Ptr_t * vRoots )
func AbcNtkDontCareWinCollectRoots_rec(pObj *ABCNetworkObject, vRoots *PointerVector[ABCNetworkObject]) {
	panic("TODO, stubbed")
}

// Abc_NtkDontCareWinCollectRoots
// C: void Abc_NtkDontCareWinCollectRoots( Odc_Man_t * p )
func (p *OdcMan) AbcNtkDontCareWinCollectRoots() {
	panic("TODO, stubbed")
}

// Abc_NtkDontCareWinAddMissing_rec
// C: int Abc_NtkDontCareWinAddMissing_rec( Odc_Man_t * p, Abc_Obj_t * pObj )
func (p *OdcMan) AbcNtkDontCareWinAddMissing_rec(pObj *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_NtkDontCareWinAddMissing
// C: int Abc_NtkDontCareWinAddMissing( Odc_Man_t * p )
func (p *OdcMan) AbcNtkDontCareWinAddMissing() int {
	panic("TODO, stubbed")
}

// Abc_NtkDontCareWindow
// C: int Abc_NtkDontCareWindow( Odc_Man_t * p )
func (p *OdcMan) AbcNtkDontCareWindow() int {
	panic("TODO, stubbed")
}

// Odc_HashKey (static)
// C: unsigned Odc_HashKey( Odc_Lit_t iFan0, Odc_Lit_t iFan1, int TableSize )
func odcHashKey(iFan0 OdcLit, iFan1 OdcLit, TableSize int) uint32 {
	panic("TODO, stubbed")
}

// Odc_HashLookup (static)
// C: Odc_Lit_t * Odc_HashLookup( Odc_Man_t * p, Odc_Lit_t iFan0, Odc_Lit_t iFan1 )
func (p *OdcMan) odcHashLookup(iFan0 OdcLit, iFan1 OdcLit) *OdcLit {
	panic("TODO, stubbed")
}

// Odc_And (static)
// C: Odc_Lit_t Odc_And( Odc_Man_t * p, Odc_Lit_t iFan0, Odc_Lit_t iFan1 )
func (p *OdcMan) odcAnd(iFan0 OdcLit, iFan1 OdcLit) OdcLit {
	panic("TODO, stubbed")
}

// Odc_Or (static)
// C: Odc_Lit_t Odc_Or( Odc_Man_t * p, Odc_Lit_t iFan0, Odc_Lit_t iFan1 )
func (p *OdcMan) odcOr(iFan0 OdcLit, iFan1 OdcLit) OdcLit {
	panic("TODO, stubbed")
}

// Odc_Xor (static)
// C: Odc_Lit_t Odc_Xor( Odc_Man_t * p, Odc_Lit_t iFan0, Odc_Lit_t iFan1 )
func (p *OdcMan) odcXor(iFan0 OdcLit, iFan1 OdcLit) OdcLit {
	panic("TODO, stubbed")
}

// Abc_NtkDontCareTransfer_rec (returns a void * that holds an Odc_Lit_t)
// C: void * Abc_NtkDontCareTransfer_rec( Odc_Man_t * p, Abc_Obj_t * pNode, Abc_Obj_t * pPivot )
func (p *OdcMan) AbcNtkDontCareTransfer_rec(pNode *ABCNetworkObject, pPivot *ABCNetworkObject) OdcLit {
	panic("TODO, stubbed")
}

// Abc_NtkDontCareTransfer
// C: int Abc_NtkDontCareTransfer( Odc_Man_t * p )
func (p *OdcMan) AbcNtkDontCareTransfer() int {
	panic("TODO, stubbed")
}

// Abc_NtkDontCareCofactors_rec
// C: unsigned Abc_NtkDontCareCofactors_rec( Odc_Man_t * p, Odc_Lit_t Lit, unsigned uMask )
func (p *OdcMan) AbcNtkDontCareCofactors_rec(Lit OdcLit, uMask uint32) uint32 {
	panic("TODO, stubbed")
}

// Abc_NtkDontCareQuantify
// C: int Abc_NtkDontCareQuantify( Odc_Man_t * p )
func (p *OdcMan) AbcNtkDontCareQuantify() int {
	panic("TODO, stubbed")
}

// Abc_NtkDontCareSimulateSetElem2
// C: void Abc_NtkDontCareSimulateSetElem2( Odc_Man_t * p )
func (p *OdcMan) AbcNtkDontCareSimulateSetElem2() {
	panic("TODO, stubbed")
}

// Abc_NtkDontCareSimulateSetElem
// C: void Abc_NtkDontCareSimulateSetElem( Odc_Man_t * p )
func (p *OdcMan) AbcNtkDontCareSimulateSetElem() {
	panic("TODO, stubbed")
}

// Abc_NtkDontCareSimulateSetRand
// C: void Abc_NtkDontCareSimulateSetRand( Odc_Man_t * p )
func (p *OdcMan) AbcNtkDontCareSimulateSetRand() {
	panic("TODO, stubbed")
}

// Abc_NtkDontCareCountMintsWord
// C: int Abc_NtkDontCareCountMintsWord( Odc_Man_t * p, unsigned * puTruth )
func (p *OdcMan) AbcNtkDontCareCountMintsWord(puTruth []uint32) int {
	panic("TODO, stubbed")
}

// Abc_NtkDontCareTruthOne
// C: void Abc_NtkDontCareTruthOne( Odc_Man_t * p, Odc_Lit_t Lit )
func (p *OdcMan) AbcNtkDontCareTruthOne(Lit OdcLit) {
	panic("TODO, stubbed")
}

// Abc_NtkDontCareSimulate_rec
// C: void Abc_NtkDontCareSimulate_rec( Odc_Man_t * p, Odc_Lit_t Lit )
func (p *OdcMan) AbcNtkDontCareSimulate_rec(Lit OdcLit) {
	panic("TODO, stubbed")
}

// Abc_NtkDontCareSimulate
// C: int Abc_NtkDontCareSimulate( Odc_Man_t * p, unsigned * puTruth )
func (p *OdcMan) AbcNtkDontCareSimulate(puTruth []uint32) int {
	panic("TODO, stubbed")
}

// Abc_NtkDontCareSimulateBefore
// C: int Abc_NtkDontCareSimulateBefore( Odc_Man_t * p, unsigned * puTruth )
func (p *OdcMan) AbcNtkDontCareSimulateBefore(puTruth []uint32) int {
	panic("TODO, stubbed")
}

// Abc_NtkDontCareCompute
// C: int Abc_NtkDontCareCompute( Odc_Man_t * p, Abc_Obj_t * pNode, Vec_Ptr_t * vLeaves, unsigned * puTruth )
func (p *OdcMan) AbcNtkDontCareCompute(pNode *ABCNetworkObject, vLeaves *PointerVector[ABCNetworkObject], puTruth []uint32) int {
	panic("TODO, stubbed")
}
