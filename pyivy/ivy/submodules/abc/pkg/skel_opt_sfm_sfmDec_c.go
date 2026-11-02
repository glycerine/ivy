package abc

// Hand-checked skeleton of src/opt/sfm/sfmDec.c.

// SFM_MASK_PI is 1, SFM_MASK_INPUT is 2, SFM_MASK_FANIN is 4, SFM_MASK_MFFC is 8
// and SFM_MASK_PIVOT is 16. SFM_SUPP_MAX is 8, SFM_WORD_MAX is 4 and
// SFM_SIM_WORDS is 8. The by-value Vec_* members are written as values.

// SfmDec is Sfm_Dec_t.
type SfmDec struct {
	// external
	pPars *SfmPar     // parameters
	pLib  *SfmLib     // library
	pTim  *SfmTim     // timing
	pMit  *SfmMit     // timing
	pNtk  *ABCNetwork // network
	// library
	vGateSizes IntVector              // Vec_Int_t (by value) : fanin counts
	vGateFuncs WordVector             // Vec_Wrd_t (by value) : gate truth tables
	vGateCnfs  IntLevelVector         // Vec_Wec_t (by value) : gate CNFs
	vGateHands PointerVector[MioGate] // Vec_Ptr_t (by value) : gate handles
	GateConst0 int                    // special gates
	GateConst1 int                    // special gates
	GateBuffer int                    // special gates
	GateInvert int                    // special gates
	GateAnd    [4]int                 // special gates
	GateOr     [4]int                 // special gates
	// objects
	nDivs       int                    // the number of divisors
	nMffc       int                    // the number of divisors
	AreaMffc    int                    // the area of gates in MFFC
	DelayMin    int                    // temporary min delay
	iTarget     int                    // target node
	iUseThis    int                    // next cofactoring var to try
	DeltaCrit   int                    // critical delta
	AreaInv     int                    // inverter area
	DelayInv    int                    // inverter delay
	pGateInv    *MioGate               // inverter
	uCareSet    uint64                 // word : computed careset
	vObjRoots   IntVector              // Vec_Int_t (by value) : roots of the window
	vObjGates   IntVector              // Vec_Int_t (by value) : functionality
	vObjFanins  IntLevelVector         // Vec_Wec_t (by value) : fanin IDs
	vObjMap     IntVector              // Vec_Int_t (by value) : object map
	vObjDec     IntVector              // Vec_Int_t (by value) : decomposition
	vObjMffc    IntVector              // Vec_Int_t (by value) : MFFC nodes
	vObjInMffc  IntVector              // Vec_Int_t (by value) : inputs of MFFC nodes
	vObjSims    WordVector             // Vec_Wrd_t (by value) : simulation patterns
	vObjSims2   WordVector             // Vec_Wrd_t (by value) : simulation patterns
	vMatchGates PointerVector[MioGate] // Vec_Ptr_t (by value) : matched gates
	vMatchFans  PointerVector[int8]    // Vec_Ptr_t (by value) : matched fanins (char * fanin arrays)
	// solver
	pSat           *SatSolver     // reusable solver
	vClauses       IntLevelVector // Vec_Wec_t (by value) : CNF clauses for the node
	vImpls         [2]IntVector   // Vec_Int_t vImpls[2] : onset/offset implications
	vSets          [2]WordVector  // Vec_Wrd_t vSets[2] : onset/offset patterns
	nPats          [2]int         // CEX count
	nPatWords      [2]int         // CEX words
	nDivWords      int            // div words
	nDivWordsAlloc int            // div words
	TtElems        [8][4]uint64   // word TtElems[SFM_SUPP_MAX][SFM_WORD_MAX]
	pTtElems       [8][]uint64    // word * pTtElems[SFM_SUPP_MAX]
	pDivWords      [8][]uint64    // word * pDivWords[SFM_SUPP_MAX]
	// temporary
	vNewNodes IntVector // Vec_Int_t (by value)
	vGateTfi  IntVector // Vec_Int_t (by value)
	vGateTfo  IntVector // Vec_Int_t (by value)
	vGateCut  IntVector // Vec_Int_t (by value)
	vGateTemp IntVector // Vec_Int_t (by value)
	vGateMffc IntVector // Vec_Int_t (by value)
	vCands    IntVector // Vec_Int_t (by value)
	Copy      [4]uint64 // word Copy[4]
	nSuppVars int
	// statistics
	timeLib        int64 // abctime
	timeWin        int64
	timeCnf        int64
	timeSat        int64
	timeSatSat     int64
	timeSatUnsat   int64
	timeEval       int64
	timeTime       int64
	timeOther      int64
	timeStart      int64
	timeTotal      int64
	nTotalNodesBeg int
	nTotalEdgesBeg int
	nTotalNodesEnd int
	nTotalEdgesEnd int
	nNodesTried    int
	nNodesChanged  int
	nNodesConst0   int
	nNodesConst1   int
	nNodesBuf      int
	nNodesInv      int
	nNodesAndOr    int
	nNodesResyn    int
	nSatCalls      int
	nSatCallsSat   int
	nSatCallsUnsat int
	nSatCallsOver  int
	nTimeOuts      int
	nNoDecs        int
	nEfforts       int
	nMaxDivs       int
	nMaxWin        int
	nAllDivs       uint64 // word
	nAllWin        uint64 // word
	nLuckySizes    [8 + 1]int
	nLuckyGates    [8 + 1]int
}

// Sfm_DecMan (static)
// C: Sfm_Dec_t * Sfm_DecMan( Abc_Obj_t * p )
func sfmDecMan(p *ABCNetworkObject) *SfmDec {
	panic("TODO, stubbed")
}

// Sfm_DecObjSim (static)
// C: word Sfm_DecObjSim( Sfm_Dec_t * p, Abc_Obj_t * pObj )
func (p *SfmDec) sfmDecObjSim(pObj *ABCNetworkObject) uint64 {
	panic("TODO, stubbed")
}

// Sfm_DecObjSim2 (static)
// C: word Sfm_DecObjSim2( Sfm_Dec_t * p, Abc_Obj_t * pObj )
func (p *SfmDec) sfmDecObjSim2(pObj *ABCNetworkObject) uint64 {
	panic("TODO, stubbed")
}

// Sfm_DecDivPats (static; a pointer into the pattern storage)
// C: word * Sfm_DecDivPats( Sfm_Dec_t * p, int d, int c )
func (p *SfmDec) sfmDecDivPats(d int, c int) []uint64 {
	panic("TODO, stubbed")
}

// Sfm_ManReadObjDelay (static)
// C: int Sfm_ManReadObjDelay( Sfm_Dec_t * p, int Id )
func (p *SfmDec) sfmManReadObjDelay(Id int) int {
	panic("TODO, stubbed")
}

// Sfm_ManReadNtkDelay (static)
// C: int Sfm_ManReadNtkDelay( Sfm_Dec_t * p )
func (p *SfmDec) sfmManReadNtkDelay() int {
	panic("TODO, stubbed")
}

// Sfm_ManReadNtkMinSlack (static)
// C: int Sfm_ManReadNtkMinSlack( Sfm_Dec_t * p )
func (p *SfmDec) sfmManReadNtkMinSlack() int {
	panic("TODO, stubbed")
}

// Sfm_ParSetDefault3
// C: void Sfm_ParSetDefault3( Sfm_Par_t * pPars )
func (pPars *SfmPar) SfmParSetDefault3() {
	panic("TODO, stubbed")
}

// Sfm_DecStart
// C: Sfm_Dec_t * Sfm_DecStart( Sfm_Par_t * pPars, Mio_Library_t * pLib, Abc_Ntk_t * pNtk )
func (pPars *SfmPar) SfmDecStart(pLib *MioLibrary, pNtk *ABCNetwork) *SfmDec {
	panic("TODO, stubbed")
}

// Sfm_DecStop
// C: void Sfm_DecStop( Sfm_Dec_t * p )
func (p *SfmDec) SfmDecStop() {
	panic("TODO, stubbed")
}

// Sfm_ObjSimulate (static)
// C: word Sfm_ObjSimulate( Abc_Obj_t * pObj )
func sfmObjSimulate(pObj *ABCNetworkObject) uint64 {
	panic("TODO, stubbed")
}

// Sfm_ObjSimulate2 (static)
// C: word Sfm_ObjSimulate2( Abc_Obj_t * pObj )
func sfmObjSimulate2(pObj *ABCNetworkObject) uint64 {
	panic("TODO, stubbed")
}

// Sfm_NtkSimulate (static)
// C: void Sfm_NtkSimulate( Abc_Ntk_t * pNtk )
func sfmNtkSimulate(pNtk *ABCNetwork) {
	panic("TODO, stubbed")
}

// Sfm_ObjSimulateNode (static)
// C: void Sfm_ObjSimulateNode( Abc_Obj_t * pObj )
func sfmObjSimulateNode(pObj *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Sfm_ObjFlipNode (static)
// C: void Sfm_ObjFlipNode( Abc_Obj_t * pObj )
func sfmObjFlipNode(pObj *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Sfm_ObjFindCareSet (static)
// C: word Sfm_ObjFindCareSet( Abc_Ntk_t * pNtk, Vec_Int_t * vRoots )
func sfmObjFindCareSet(pNtk *ABCNetwork, vRoots *IntVector) uint64 {
	panic("TODO, stubbed")
}

// Sfm_ObjSetupSimInfo (static)
// C: void Sfm_ObjSetupSimInfo( Abc_Obj_t * pObj )
func sfmObjSetupSimInfo(pObj *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Sfm_ObjSetdownSimInfo (static)
// C: void Sfm_ObjSetdownSimInfo( Abc_Obj_t * pObj )
func sfmObjSetdownSimInfo(pObj *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Sfm_DecPrepareSolver
// C: int Sfm_DecPrepareSolver( Sfm_Dec_t * p )
func (p *SfmDec) SfmDecPrepareSolver() int {
	panic("TODO, stubbed")
}

// Sfm_DecFindCost
// C: int Sfm_DecFindCost( Sfm_Dec_t * p, int c, int iLit, word * pMask )
func (p *SfmDec) SfmDecFindCost(c int, iLit int, pMask []uint64) int {
	panic("TODO, stubbed")
}

// Sfm_DecPrint
// C: void Sfm_DecPrint( Sfm_Dec_t * p, word Masks[2][SFM_SIM_WORDS] )
func (p *SfmDec) SfmDecPrint(Masks *[2][8]uint64) {
	panic("TODO, stubbed")
}

// Sfm_DecVarCost (Counts is an out parameter)
// C: void Sfm_DecVarCost( Sfm_Dec_t * p, word Masks[2][SFM_SIM_WORDS], int d, int Counts[2][2] )
func (p *SfmDec) SfmDecVarCost(Masks *[2][8]uint64, d int, Counts *[2][2]int) {
	panic("TODO, stubbed")
}

// Sfm_DecFindBestVar2
// C: int Sfm_DecFindBestVar2( Sfm_Dec_t * p, word Masks[2][SFM_SIM_WORDS] )
func (p *SfmDec) SfmDecFindBestVar2(Masks *[2][8]uint64) int {
	panic("TODO, stubbed")
}

// Sfm_DecFindBestVar
// C: int Sfm_DecFindBestVar( Sfm_Dec_t * p, word Masks[2][SFM_SIM_WORDS] )
func (p *SfmDec) SfmDecFindBestVar(Masks *[2][8]uint64) int {
	panic("TODO, stubbed")
}

// Sfm_DecMffcArea
// C: int Sfm_DecMffcArea( Abc_Ntk_t * pNtk, Vec_Int_t * vMffc )
func SfmDecMffcArea(pNtk *ABCNetwork, vMffc *IntVector) int {
	panic("TODO, stubbed")
}

// Sfm_MffcDeref_rec
// C: int Sfm_MffcDeref_rec( Abc_Obj_t * pObj )
func SfmMffcDeref_rec(pObj *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Sfm_MffcRef_rec
// C: int Sfm_MffcRef_rec( Abc_Obj_t * pObj, Vec_Int_t * vMffc )
func SfmMffcRef_rec(pObj *ABCNetworkObject, vMffc *IntVector) int {
	panic("TODO, stubbed")
}

// Sfm_DecMffcAreaReal
// C: int Sfm_DecMffcAreaReal( Abc_Obj_t * pPivot, Vec_Int_t * vCut, Vec_Int_t * vMffc )
func SfmDecMffcAreaReal(pPivot *ABCNetworkObject, vCut *IntVector, vMffc *IntVector) int {
	panic("TODO, stubbed")
}

// Sfm_DecPrepareVec
// C: void Sfm_DecPrepareVec( Vec_Int_t * vMap, int * pNodes, int nNodes, Vec_Int_t * vCut )
func SfmDecPrepareVec(vMap *IntVector, pNodes []int, nNodes int, vCut *IntVector) {
	panic("TODO, stubbed")
}

// Sfm_DecComputeFlipInvGain (pfNeedInv is an out parameter)
// C: int Sfm_DecComputeFlipInvGain( Sfm_Dec_t * p, Abc_Obj_t * pPivot, int * pfNeedInv )
func (p *SfmDec) SfmDecComputeFlipInvGain(pPivot *ABCNetworkObject, pfNeedInv *int) int {
	panic("TODO, stubbed")
}

// Sfm_DecCombineDec
// C: int Sfm_DecCombineDec( Sfm_Dec_t * p, word * pTruth0, word * pTruth1, int * pSupp0, int * pSupp1, int nSupp0, int nSupp1, word * pTruth, int * pSupp, int Var )
func (p *SfmDec) SfmDecCombineDec(pTruth0 []uint64, pTruth1 []uint64, pSupp0 []int, pSupp1 []int, nSupp0 int, nSupp1 int, pTruth []uint64, pSupp []int, Var int) int {
	panic("TODO, stubbed")
}

// Sfm_DecPeformDec_rec
// C: int Sfm_DecPeformDec_rec( Sfm_Dec_t * p, word * pTruth, int * pSupp, int * pAssump, int nAssump, word Masks[2][SFM_SIM_WORDS], int fCofactor, int nSuppAdd )
func (p *SfmDec) SfmDecPeformDec_rec(pTruth []uint64, pSupp []int, pAssump []int, nAssump int, Masks *[2][8]uint64, fCofactor int, nSuppAdd int) int {
	panic("TODO, stubbed")
}

// Sfm_DecPeformDec2
// C: int Sfm_DecPeformDec2( Sfm_Dec_t * p, Abc_Obj_t * pObj )
func (p *SfmDec) SfmDecPeformDec2(pObj *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Sfm_DecPeformDec3
// C: int Sfm_DecPeformDec3( Sfm_Dec_t * p, Abc_Obj_t * pObj )
func (p *SfmDec) SfmDecPeformDec3(pObj *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_NtkUpdateIncLevel_rec
// C: void Abc_NtkUpdateIncLevel_rec( Abc_Obj_t * pObj )
func AbcNtkUpdateIncLevel_rec(pObj *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Abc_NtkDfsCheck_rec
// C: int Abc_NtkDfsCheck_rec( Abc_Obj_t * pObj, Abc_Obj_t * pPivot )
func AbcNtkDfsCheck_rec(pObj *ABCNetworkObject, pPivot *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_NtkDfsReverseOne_rec
// C: void Abc_NtkDfsReverseOne_rec( Abc_Obj_t * pObj, Vec_Int_t * vTfo, int nLevelMax, int nFanoutMax )
func AbcNtkDfsReverseOne_rec(pObj *ABCNetworkObject, vTfo *IntVector, nLevelMax int, nFanoutMax int) {
	panic("TODO, stubbed")
}

// Abc_NtkDfsOne_rec
// C: int Abc_NtkDfsOne_rec( Abc_Obj_t * pObj, Vec_Int_t * vTfi, int nLevelMin, int CiLabel )
func AbcNtkDfsOne_rec(pObj *ABCNetworkObject, vTfi *IntVector, nLevelMin int, CiLabel int) int {
	panic("TODO, stubbed")
}

// Sfm_DecAddNode
// C: void Sfm_DecAddNode( Abc_Obj_t * pObj, Vec_Int_t * vMap, Vec_Int_t * vGates, int fSkip, int fVeryVerbose )
func SfmDecAddNode(pObj *ABCNetworkObject, vMap *IntVector, vGates *IntVector, fSkip int, fVeryVerbose int) {
	panic("TODO, stubbed")
}

// Sfm_DecNodeIsMffc (static)
// C: int Sfm_DecNodeIsMffc( Abc_Obj_t * p, int nLevelMin )
func sfmDecNodeIsMffc(p *ABCNetworkObject, nLevelMin int) int {
	panic("TODO, stubbed")
}

// Sfm_DecNodeIsMffcInput (static)
// C: int Sfm_DecNodeIsMffcInput( Abc_Obj_t * p, int nLevelMin, Sfm_Tim_t * pTim, Abc_Obj_t * pPivot )
func sfmDecNodeIsMffcInput(p *ABCNetworkObject, nLevelMin int, pTim *SfmTim, pPivot *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Sfm_DecNodeIsMffcInput2 (static)
// C: int Sfm_DecNodeIsMffcInput2( Abc_Obj_t * p, int nLevelMin, Sfm_Mit_t * pMit, Abc_Obj_t * pPivot )
func sfmDecNodeIsMffcInput2(p *ABCNetworkObject, nLevelMin int, pMit *SfmMit, pPivot *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Sfm_DecMarkMffc
// C: void Sfm_DecMarkMffc( Abc_Obj_t * pPivot, int nLevelMin, int nMffcMax, int fVeryVerbose, Vec_Int_t * vMffc, Vec_Int_t * vInMffc, Sfm_Tim_t * pTim, Sfm_Mit_t * pMit )
func SfmDecMarkMffc(pPivot *ABCNetworkObject, nLevelMin int, nMffcMax int, fVeryVerbose int, vMffc *IntVector, vInMffc *IntVector, pTim *SfmTim, pMit *SfmMit) {
	panic("TODO, stubbed")
}

// Sfm_DecExtract
// C: int Sfm_DecExtract( Abc_Ntk_t * pNtk, Sfm_Par_t * pPars, Abc_Obj_t * pPivot, Vec_Int_t * vRoots, Vec_Int_t * vGates, Vec_Wec_t * vFanins, Vec_Int_t * vMap, Vec_Int_t * vTfi, Vec_Int_t * vTfo, Vec_Int_t * vMffc, Vec_Int_t * vInMffc, Sfm_Tim_t * pTim, Sfm_Mit_t * pMit )
func SfmDecExtract(pNtk *ABCNetwork, pPars *SfmPar, pPivot *ABCNetworkObject, vRoots *IntVector, vGates *IntVector, vFanins *IntLevelVector, vMap *IntVector, vTfi *IntVector, vTfo *IntVector, vMffc *IntVector, vInMffc *IntVector, pTim *SfmTim, pMit *SfmMit) int {
	panic("TODO, stubbed")
}

// Sfm_DecInsert
// C: Abc_Obj_t * Sfm_DecInsert( Abc_Ntk_t * pNtk, Abc_Obj_t * pPivot, int Limit, Vec_Int_t * vGates, Vec_Wec_t * vFanins, Vec_Int_t * vMap, Vec_Ptr_t * vGateHandles, int GateBuf, int GateInv, Vec_Wrd_t * vFuncs, Vec_Int_t * vNewNodes, Sfm_Mit_t * pMit )
func SfmDecInsert(pNtk *ABCNetwork, pPivot *ABCNetworkObject, Limit int, vGates *IntVector, vFanins *IntLevelVector, vMap *IntVector, vGateHandles *PointerVector[MioGate], GateBuf int, GateInv int, vFuncs *WordVector, vNewNodes *IntVector, pMit *SfmMit) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Sfm_DecPrintStats
// C: void Sfm_DecPrintStats( Sfm_Dec_t * p )
func (p *SfmDec) SfmDecPrintStats() {
	panic("TODO, stubbed")
}

// Abc_NtkCountStats
// C: void Abc_NtkCountStats( Sfm_Dec_t * p, int Limit )
func (p *SfmDec) AbcNtkCountStats(Limit int) {
	panic("TODO, stubbed")
}

// Abc_NtkAreaOptOne
// C: Abc_Obj_t * Abc_NtkAreaOptOne( Sfm_Dec_t * p, int i )
func (p *SfmDec) AbcNtkAreaOptOne(i int) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_NtkAreaOpt
// C: void Abc_NtkAreaOpt( Sfm_Dec_t * p )
func (p *SfmDec) AbcNtkAreaOpt() {
	panic("TODO, stubbed")
}

// Abc_NtkAreaOpt2
// C: void Abc_NtkAreaOpt2( Sfm_Dec_t * p )
func (p *SfmDec) AbcNtkAreaOpt2() {
	panic("TODO, stubbed")
}

// Abc_NtkDelayOpt
// C: void Abc_NtkDelayOpt( Sfm_Dec_t * p )
func (p *SfmDec) AbcNtkDelayOpt() {
	panic("TODO, stubbed")
}

// Abc_NtkPerformMfs3
// C: void Abc_NtkPerformMfs3( Abc_Ntk_t * pNtk, Sfm_Par_t * pPars )
func AbcNtkPerformMfs3(pNtk *ABCNetwork, pPars *SfmPar) {
	panic("TODO, stubbed")
}
