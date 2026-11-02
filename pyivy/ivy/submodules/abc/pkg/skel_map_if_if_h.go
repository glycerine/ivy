package abc

import "strings"

// Hand-checked skeleton of src/map/if/if.h.

const (
	IF_MAX_LUTSIZE      = 32
	IF_MAX_FUNC_LUTSIZE = 15
	IF_INFINITY         = 100000000
	IF_COST_MAX         = 4095 // ((1<<12)-1)
	IF_BIG_CHAR         = 120  // ((char)120)
)

// If_Type_t: object types.
type IfType int

const (
	IF_NONE   IfType = iota // 0: non-existent object
	IF_CONST1               // 1: constant 1
	IF_CI                   // 2: combinational input
	IF_CO                   // 3: combinational output
	IF_AND                  // 4: AND node
	IF_VOID                 // 5: unused object
)

// IfifPar is Ifif_Par_t.
type IfifPar struct {
	nLutSize     int                     // the LUT size
	pLutLib      *IfLibLut               // the LUT library
	pLutDelays   [IF_MAX_LUTSIZE]float32 // pin-to-pin delays of the max LUT
	DelayWire    float32                 // wire delay
	nDegree      int                     // structure degree
	fCascade     int                     // cascade
	fVerbose     int                     // verbose
	fVeryVerbose int                     // verbose
}

// IfPar is If_Par_t: mapping parameters.
type IfPar struct {
	// user-controlable parameters
	nLutSize        int     // the LUT size
	nCutsMax        int     // the max number of cuts
	nFlowIters      int     // the number of iterations of area recovery
	nAreaIters      int     // the number of iterations of area recovery
	nGateSize       int     // the max size of the AND/OR gate to map into
	nNonDecLimit    int     // the max size of non-dec nodes
	DelayTarget     float32 // delay target
	Epsilon         float32 // value used in comparison floating point numbers
	nRelaxRatio     int     // delay relaxation ratio
	nStructType     int     // type of the structure
	nAndDelay       int     // delay of AND-gate in LUT library units
	nAndArea        int     // area of AND-gate in LUT library units
	fPreprocess     int     // preprossing
	fArea           int     // area-oriented mapping
	fFancy          int     // a fancy feature
	fExpRed         int     // expand/reduce of the best cuts
	fLatchPaths     int     // reset timing on latch paths
	fLut6Filter     int     // uses filtering of 6-LUT functions
	fEdge           int     // uses edge-based cut selection heuristics
	fPower          int     // uses power-aware cut selection heuristics
	fCutMin         int     // performs cut minimization by removing functionally reducdant variables
	fDelayOpt       int     // special delay optimization
	fDelayOptLut    int     // delay optimization for LUTs
	fDsdBalance     int     // special delay optimization
	fUserRecLib     int     // use recorded library
	fUserSesLib     int     // use SAT-based synthesis
	fBidec          int     // use bi-decomposition
	fUse34Spec      int     // use specialized matching
	fUseBat         int     // use one specialized feature
	fUseBuffs       int     // use buffers to decouple outputs
	fEnableCheck07  int     // enable additional checking
	fEnableCheck08  int     // enable additional checking
	fEnableCheck10  int     // enable additional checking
	fEnableCheck75  int     // enable additional checking
	fEnableCheck75u int     // enable additional checking
	fUseDsd         int     // compute DSD of the cut functions
	fUseDsdTune     int     // use matching based on precomputed manager
	fUseCofVars     int     // use cofactoring variables
	fUseAndVars     int     // use bi-decomposition
	fUseTtPerm      int     // compute truth tables of the cut functions
	fDeriveLuts     int     // enables deriving LUT structures
	fDoAverage      int     // optimize average rather than maximum level
	fHashMapping    int     // perform AIG hashing after mapping
	fVerbose        int     // the verbosity flag
	fVerboseTrace   int     // the verbosity flag
	pLutStruct      string  // char * : LUT structure
	WireDelay       float32 // wire delay
	// internal parameters
	fSkipCutFilter int                                                     // skip cut filter
	fAreaOnly      int                                                     // area only mode
	fTruth         int                                                     // truth table computation enabled
	fUsePerm       int                                                     // use permutation (delay info)
	fUseBdds       int                                                     // use local BDDs as a cost function
	fUseSops       int                                                     // use local SOPs as a cost function
	fUseCnfs       int                                                     // use local CNFs as a cost function
	fUseMv         int                                                     // use local MV-SOPs as a cost function
	nLatchesCi     int                                                     // the number of latches among the CIs
	nLatchesCo     int                                                     // the number of latches among the COs
	nLatchesCiBox  int                                                     // the number of white box outputs among the CIs
	nLatchesCoBox  int                                                     // the number of white box inputs among the COs
	fLiftLeaves    int                                                     // shift the leaves for seq mapping
	fUseCoAttrs    int                                                     // use CO attributes
	DelayTargetNew float32                                                 // new delay target
	FinalDelay     float32                                                 // final delay after mapping
	FinalArea      float32                                                 // final area after mapping
	pLutLib        *IfLibLut                                               // the LUT library
	pTimesArr      []float32                                               // float * : arrival times
	pTimesReq      []float32                                               // float * : required times
	pFuncCost      func(*IfMan, *IfCut) int                                // procedure to compute the user's cost of a cut
	pFuncUser      func(*IfMan, *IfObj, *IfCut) int                        // procedure called for each cut when cut computation is finished
	pFuncCell      func(*IfMan, []uint32, int, int, *strings.Builder) int  // procedure called for cut functions (unsigned *, int, int, char *)
	pFuncCell2     func(*IfMan, []uint64, int, *ByteVector, *[]string) int // procedure called for cut functions (word *, int, Vec_Str_t *, char **)
	pFuncWrite     func(*IfMan) int                                        // procedure called for cut functions
	pReoMan        *ReoMan                                                 // void * : reordering manager
}

// IfLibLut is If_LibLut_t: the LUT library.
type IfLibLut struct {
	pName         string                                          // the name of the LUT library
	LutMax        int                                             // the maximum LUT size
	fVarPinDelays int                                             // set to 1 if variable pin delays are specified
	pLutAreas     [IF_MAX_LUTSIZE + 1]float32                     // the areas of LUTs
	pLutDelays    [IF_MAX_LUTSIZE + 1][IF_MAX_LUTSIZE + 1]float32 // the delays of LUTs
}

// IfMan is If_Man_t: the mapping manager.
type IfMan struct {
	pName string
	// mapping parameters
	pPars *IfPar
	// mapping nodes
	pConst1  *IfObj                // the constant 1 node
	vCis     *PointerVector[IfObj] // the primary inputs
	vCos     *PointerVector[IfObj] // the primary outputs
	vObjs    *PointerVector[IfObj] // all objects
	vObjsRev *PointerVector[IfObj] // reverse topological order of objects
	vTemp    *PointerVector[any]   // temporary array
	nObjs    [IF_VOID]int          // the number of objects by type
	// various data
	nLevelMax    int                    // the max number of AIG levels
	fEpsilon     float32                // epsilon used for comparison
	RequiredGlo  float32                // global required times
	RequiredGlo2 float32                // global required times
	AreaGlo      float32                // global area
	nNets        int                    // the sum total of fanins of all LUTs in the mapping
	dPower       float32                // the sum total of switching activities of all LUTs in the mapping
	nCutsUsed    int                    // the number of cuts currently used
	nCutsMerged  int                    // the total number of cuts merged
	puTemp       [4][]uint32            // unsigned * : used for the truth table computation
	puTempW      []uint64               // word * : used for the truth table computation
	SortMode     int                    // one of the three sorting modes
	fNextRound   int                    // set to 1 after the first round
	nChoices     int                    // the number of choice nodes
	vSwitching   *IntVector             // switching activity of each node
	pPerm        [3][IF_MAX_LUTSIZE]int // permutations
	uSharedMask  uint32                 // mask of shared variables
	nShared      int                    // the number of shared variables
	fReqTimeWarn int                    // warning about exceeding required times was printed
	// SOP balancing
	vCover   *IntVector  // used to compute ISOP
	vArray   *IntVector  // intermediate storage
	vAnds    *WordVector // intermediate storage
	vOrGate  *WordVector // intermediate storage
	vAndGate *WordVector // intermediate storage
	// sequential mapping
	vLatchOrder *PointerVector[IfObj] // topological ordering of latches
	vLags       *IntVector            // sequentail lags of all nodes
	nAttempts   int                   // the number of attempts in binary search
	nMaxIters   int                   // the maximum number of iterations
	Period      int                   // the current value of the clock period (for seq mapping)
	// memory management
	nTruth6Words    [IF_MAX_FUNC_LUTSIZE + 1]int // the size of the truth table if allocated
	nPermWords      int                          // the size of the permutation array (in words)
	nObjBytes       int                          // the size of the object
	nCutBytes       int                          // the size of the cut
	nSetBytes       int                          // the size of the cut set
	pMemObj         *MmrFixedArena               // Mem_Fixed_t * : memory manager for objects (entrysize = nEntrySize)
	pMemSet         *MmrFixedArena               // Mem_Fixed_t * : memory manager for sets of cuts (entrysize = nCutSize*(nCutsMax+1))
	pMemCi          *IfSet                       // memory for CI cutsets
	pMemAnd         *IfSet                       // memory for AND cutsets
	pFreeList       *IfSet                       // the list of free cutsets
	nSmallSupp      int                          // the small support
	nCutsTotal      int
	nCutsUseless    [32]int
	nCutsCount      [32]int
	nCutsCountAll   int
	nCutsUselessAll int
	nCuts5, nCuts5a int
	pIfDsdMan       *IfDsdMan                                // DSD manager
	vTtMem          [IF_MAX_FUNC_LUTSIZE + 1]*WordMemory     // truth table memory and hash table
	vTtIsops        [IF_MAX_FUNC_LUTSIZE + 1]*IntLevelVector // mapping of truth table into DSD
	vTtDsds         [IF_MAX_FUNC_LUTSIZE + 1]*IntVector      // mapping of truth table into DSD
	vTtPerms        [IF_MAX_FUNC_LUTSIZE + 1]*ByteVector     // mapping of truth table into permutations
	vTtVars         [IF_MAX_FUNC_LUTSIZE + 1]*ByteVector     // mapping of truth table into selected vars
	vTtDecs         [IF_MAX_FUNC_LUTSIZE + 1]*IntVector      // mapping of truth table into decomposition pattern
	vTtOccurs       [IF_MAX_FUNC_LUTSIZE + 1]*IntVector      // truth table occurange counters
	vPairHash       *IntTupleHash                            // Hash_IntMan_t * : hashing pairs of truth tables
	vPairRes        *IntVector                               // resulting truth table
	vPairPerms      *ByteVector                              // resulting permutation
	vTtMem6         *WordMemory
	pCanonPerm      [IF_MAX_LUTSIZE]int8
	uCanonPhase     uint32
	nCacheHits      int
	nCacheMisses    int
	timeCache       [6]int64 // abctime
	nBestCutSmall   [2]int
	nCountNonDec    [2]int
	vCutData        *IntVector // cut data storage
	pArrTimeProfile [IF_MAX_FUNC_LUTSIZE]int
	vVisited        *PointerVector[IfObj]
	pUserMan        any // void *
	vDump           *IntVector
	pDumpIns        [16]int
	// timing manager
	pManTim  *TimManager // Tim_Man_t *
	vCoAttrs *IntVector  // CO attributes   0=optimize; 1=keep; 2=relax
	// hash table for functions
	nTableSize    [2]int         // hash table size
	nTableEntries [2]int         // hash table entries
	pHashTable    [2][]any       // void ** : hash table bins
	pMemEntries   *MmrFixedArena // Mem_Fixed_t * : memory manager for hash table entries
}

// IfCut is If_Cut_t: a priority cut.
type IfCut struct {
	Area      float32 // area (or area-flow) of the cut
	Edge      float32 // the edge flow
	Power     float32 // the power flow
	Delay     float32 // delay of the cut
	iCutFunc  int     // TT ID of the cut
	uMaskFunc int     // polarity bitmask
	uSign     uint32  // cut signature
	Cost      uint32  // unsigned : 12 : the user's cost of the cut (related to IF_COST_MAX)
	fCompl    uint32  // unsigned : 1 : the complemented attribute
	fUser     uint32  // unsigned : 1 : using the user's area and delay
	fUseless  uint32  // unsigned : 1 : cannot be used in the mapping
	fAndCut   uint32  // unsigned : 1 : matched with AND gate
	nLimit    uint32  // unsigned : 8 : the maximum number of leaves
	nLeaves   uint32  // unsigned : 8 : the number of leaves
	pLeaves   []int   // int pLeaves[0] : flexible array member
}

// IfSet is If_Set_t: a set of priority cuts.
type IfSet struct {
	nCutsMax int16    // short : the max number of cuts
	nCuts    int16    // short : the current number of cuts
	pNext    *IfSet   // next cutset in the free list
	ppCuts   []*IfCut // If_Cut_t ** : the array of pointers to the cuts
}

// IfObj is If_Obj_t: a node extension.
type IfObj struct {
	Type        uint32  // unsigned : 4 : object
	fCompl0     uint32  // unsigned : 1 : complemented attribute
	fCompl1     uint32  // unsigned : 1 : complemented attribute
	fPhase      uint32  // unsigned : 1 : phase of the node
	fRepr       uint32  // unsigned : 1 : representative of the equivalence class
	fMark       uint32  // unsigned : 1 : multipurpose mark
	fVisit      uint32  // unsigned : 1 : multipurpose mark
	fSpec       uint32  // unsigned : 1 : multipurpose mark
	fDriver     uint32  // unsigned : 1 : multipurpose mark
	fSkipCut    uint32  // unsigned : 1 : multipurpose mark
	Level       uint32  // unsigned : 19 : logic level of the node
	Id          int     // integer ID
	IdPio       int     // integer ID of PIs/POs
	nRefs       int     // the number of references
	nVisits     int     // the number of visits to this node
	nVisitsCopy int     // the number of visits to this node
	pFanin0     *IfObj  // the first fanin
	pFanin1     *IfObj  // the second fanin
	pEquiv      *IfObj  // the choice node
	EstRefs     float32 // estimated reference counter
	Required    float32 // required time of the onde
	LValue      float32 // sequential arrival time of the node
	// union { void * pCopy; int iCopy; }
	pCopy   any // used for object duplication
	iCopy   int
	pCutSet *IfSet // the pointer to the cutset
	CutBest IfCut  // the best cut selected
}

// IfBox is If_Box_t.
type IfBox struct {
	pName   string
	fSeq    int8
	fBlack  int8
	fOuter  int8
	fUnused int8
	Id      int
	nPis    int
	nPos    int
	pDelays []int // int *
}

// IfLibBox is If_LibBox_t.
type IfLibBox struct {
	nBoxes int
	vBoxes *PointerVector[IfBox]
}

// If_Regular (static inline)
// C: If_Obj_t * If_Regular( If_Obj_t * p )
func ifRegular(p *IfObj) *IfObj {
	panic("TODO, stubbed")
}

// If_Not (static inline)
// C: If_Obj_t * If_Not( If_Obj_t * p )
func ifNot(p *IfObj) *IfObj {
	panic("TODO, stubbed")
}

// If_NotCond (static inline)
// C: If_Obj_t * If_NotCond( If_Obj_t * p, int c )
func ifNotCond(p *IfObj, c int) *IfObj {
	panic("TODO, stubbed")
}

// If_IsComplement (static inline)
// C: int If_IsComplement( If_Obj_t * p )
func ifIsComplement(p *IfObj) int {
	panic("TODO, stubbed")
}

// If_ManCiNum (static inline)
// C: int If_ManCiNum( If_Man_t * p )
func (p *IfMan) ifManCiNum() int {
	panic("TODO, stubbed")
}

// If_ManCoNum (static inline)
// C: int If_ManCoNum( If_Man_t * p )
func (p *IfMan) ifManCoNum() int {
	panic("TODO, stubbed")
}

// If_ManAndNum (static inline)
// C: int If_ManAndNum( If_Man_t * p )
func (p *IfMan) ifManAndNum() int {
	panic("TODO, stubbed")
}

// If_ManObjNum (static inline)
// C: int If_ManObjNum( If_Man_t * p )
func (p *IfMan) ifManObjNum() int {
	panic("TODO, stubbed")
}

// If_ManConst1 (static inline)
// C: If_Obj_t * If_ManConst1( If_Man_t * p )
func (p *IfMan) ifManConst1() *IfObj {
	panic("TODO, stubbed")
}

// If_ManCi (static inline)
// C: If_Obj_t * If_ManCi( If_Man_t * p, int i )
func (p *IfMan) ifManCi(i int) *IfObj {
	panic("TODO, stubbed")
}

// If_ManCo (static inline)
// C: If_Obj_t * If_ManCo( If_Man_t * p, int i )
func (p *IfMan) ifManCo(i int) *IfObj {
	panic("TODO, stubbed")
}

// If_ManLi (static inline)
// C: If_Obj_t * If_ManLi( If_Man_t * p, int i )
func (p *IfMan) ifManLi(i int) *IfObj {
	panic("TODO, stubbed")
}

// If_ManLo (static inline)
// C: If_Obj_t * If_ManLo( If_Man_t * p, int i )
func (p *IfMan) ifManLo(i int) *IfObj {
	panic("TODO, stubbed")
}

// If_ManObj (static inline)
// C: If_Obj_t * If_ManObj( If_Man_t * p, int i )
func (p *IfMan) ifManObj(i int) *IfObj {
	panic("TODO, stubbed")
}

// If_ObjIsConst1 (static inline)
// C: int If_ObjIsConst1( If_Obj_t * pObj )
func ifObjIsConst1(pObj *IfObj) int {
	panic("TODO, stubbed")
}

// If_ObjIsCi (static inline)
// C: int If_ObjIsCi( If_Obj_t * pObj )
func ifObjIsCi(pObj *IfObj) int {
	panic("TODO, stubbed")
}

// If_ObjIsCo (static inline)
// C: int If_ObjIsCo( If_Obj_t * pObj )
func ifObjIsCo(pObj *IfObj) int {
	panic("TODO, stubbed")
}

// If_ObjIsTerm (static inline)
// C: int If_ObjIsTerm( If_Obj_t * pObj )
func ifObjIsTerm(pObj *IfObj) int {
	panic("TODO, stubbed")
}

// If_ObjIsLatch (static inline)
// C: int If_ObjIsLatch( If_Obj_t * pObj )
func ifObjIsLatch(pObj *IfObj) int {
	panic("TODO, stubbed")
}

// If_ObjIsAnd (static inline)
// C: int If_ObjIsAnd( If_Obj_t * pObj )
func ifObjIsAnd(pObj *IfObj) int {
	panic("TODO, stubbed")
}

// If_ObjId (static inline)
// C: int If_ObjId( If_Obj_t * pObj )
func ifObjId(pObj *IfObj) int {
	panic("TODO, stubbed")
}

// If_ObjFanin0 (static inline)
// C: If_Obj_t * If_ObjFanin0( If_Obj_t * pObj )
func ifObjFanin0(pObj *IfObj) *IfObj {
	panic("TODO, stubbed")
}

// If_ObjFanin1 (static inline)
// C: If_Obj_t * If_ObjFanin1( If_Obj_t * pObj )
func ifObjFanin1(pObj *IfObj) *IfObj {
	panic("TODO, stubbed")
}

// If_ObjFaninC0 (static inline)
// C: int If_ObjFaninC0( If_Obj_t * pObj )
func ifObjFaninC0(pObj *IfObj) int {
	panic("TODO, stubbed")
}

// If_ObjFaninC1 (static inline)
// C: int If_ObjFaninC1( If_Obj_t * pObj )
func ifObjFaninC1(pObj *IfObj) int {
	panic("TODO, stubbed")
}

// If_ObjCopy (static inline; returns a void *)
// C: void * If_ObjCopy( If_Obj_t * pObj )
func ifObjCopy(pObj *IfObj) any {
	panic("TODO, stubbed")
}

// If_ObjLevel (static inline)
// C: int If_ObjLevel( If_Obj_t * pObj )
func ifObjLevel(pObj *IfObj) int {
	panic("TODO, stubbed")
}

// If_ObjSetLevel (static inline)
// C: void If_ObjSetLevel( If_Obj_t * pObj, int Level )
func ifObjSetLevel(pObj *IfObj, Level int) {
	panic("TODO, stubbed")
}

// If_ObjSetCopy (static inline)
// C: void If_ObjSetCopy( If_Obj_t * pObj, void * pCopy )
func ifObjSetCopy(pObj *IfObj, pCopy any) {
	panic("TODO, stubbed")
}

// If_ObjSetChoice (static inline)
// C: void If_ObjSetChoice( If_Obj_t * pObj, If_Obj_t * pEqu )
func ifObjSetChoice(pObj *IfObj, pEqu *IfObj) {
	panic("TODO, stubbed")
}

// If_CutLeaveNum (static inline)
// C: int If_CutLeaveNum( If_Cut_t * pCut )
func ifCutLeaveNum(pCut *IfCut) int {
	panic("TODO, stubbed")
}

// If_CutLeaves (static inline; int * into the flexible leaf array)
// C: int * If_CutLeaves( If_Cut_t * pCut )
func ifCutLeaves(pCut *IfCut) []int {
	panic("TODO, stubbed")
}

// If_CutLeaf (static inline)
// C: If_Obj_t * If_CutLeaf( If_Man_t * p, If_Cut_t * pCut, int i )
func (p *IfMan) ifCutLeaf(pCut *IfCut, i int) *IfObj {
	panic("TODO, stubbed")
}

// If_CutSuppMask (static inline)
// C: unsigned If_CutSuppMask( If_Cut_t * pCut )
func ifCutSuppMask(pCut *IfCut) uint32 {
	panic("TODO, stubbed")
}

// If_CutTruthWords (static inline)
// C: int If_CutTruthWords( int nVarsMax )
func ifCutTruthWords(nVarsMax int) int {
	panic("TODO, stubbed")
}

// If_CutPermWords (static inline)
// C: int If_CutPermWords( int nVarsMax )
func ifCutPermWords(nVarsMax int) int {
	panic("TODO, stubbed")
}

// If_CutLeafBit (static inline)
// C: int If_CutLeafBit( If_Cut_t * pCut, int i )
func ifCutLeafBit(pCut *IfCut, i int) int {
	panic("TODO, stubbed")
}

// If_CutPerm (static inline; char * into the cut)
// C: char * If_CutPerm( If_Cut_t * pCut )
func ifCutPerm(pCut *IfCut) []int8 {
	panic("TODO, stubbed")
}

// If_CutCopy (static inline)
// C: void If_CutCopy( If_Man_t * p, If_Cut_t * pDst, If_Cut_t * pSrc )
func (p *IfMan) ifCutCopy(pDst *IfCut, pSrc *IfCut) {
	panic("TODO, stubbed")
}

// If_CutSetup (static inline)
// C: void If_CutSetup( If_Man_t * p, If_Cut_t * pCut )
func (p *IfMan) ifCutSetup(pCut *IfCut) {
	panic("TODO, stubbed")
}

// If_ObjCutBest (static inline)
// C: If_Cut_t * If_ObjCutBest( If_Obj_t * pObj )
func ifObjCutBest(pObj *IfObj) *IfCut {
	panic("TODO, stubbed")
}

// If_ObjCutSign (static inline)
// C: unsigned If_ObjCutSign( unsigned ObjId )
func ifObjCutSign(ObjId uint32) uint32 {
	panic("TODO, stubbed")
}

// If_ObjCutSignCompute (static inline)
// C: unsigned If_ObjCutSignCompute( If_Cut_t * p )
func ifObjCutSignCompute(p *IfCut) uint32 {
	panic("TODO, stubbed")
}

// If_ObjArrTime (static inline)
// C: float If_ObjArrTime( If_Obj_t * pObj )
func ifObjArrTime(pObj *IfObj) float32 {
	panic("TODO, stubbed")
}

// If_ObjSetArrTime (static inline)
// C: void If_ObjSetArrTime( If_Obj_t * pObj, float ArrTime )
func ifObjSetArrTime(pObj *IfObj, ArrTime float32) {
	panic("TODO, stubbed")
}

// If_ObjLValue (static inline)
// C: float If_ObjLValue( If_Obj_t * pObj )
func ifObjLValue(pObj *IfObj) float32 {
	panic("TODO, stubbed")
}

// If_ObjSetLValue (static inline)
// C: void If_ObjSetLValue( If_Obj_t * pObj, float LValue )
func ifObjSetLValue(pObj *IfObj, LValue float32) {
	panic("TODO, stubbed")
}

// If_CutData (static inline; void *)
// C: void * If_CutData( If_Cut_t * pCut )
func ifCutData(pCut *IfCut) any {
	panic("TODO, stubbed")
}

// If_CutSetData (static inline)
// C: void If_CutSetData( If_Cut_t * pCut, void * pData )
func ifCutSetData(pCut *IfCut, pData any) {
	panic("TODO, stubbed")
}

// If_CutDataInt (static inline)
// C: int If_CutDataInt( If_Cut_t * pCut )
func ifCutDataInt(pCut *IfCut) int {
	panic("TODO, stubbed")
}

// If_CutSetDataInt (static inline)
// C: void If_CutSetDataInt( If_Cut_t * pCut, int Data )
func ifCutSetDataInt(pCut *IfCut, Data int) {
	panic("TODO, stubbed")
}

// If_CutTruthLit (static inline)
// C: int If_CutTruthLit( If_Cut_t * pCut )
func ifCutTruthLit(pCut *IfCut) int {
	panic("TODO, stubbed")
}

// If_CutTruthIsCompl (static inline)
// C: int If_CutTruthIsCompl( If_Cut_t * pCut )
func ifCutTruthIsCompl(pCut *IfCut) int {
	panic("TODO, stubbed")
}

// If_CutTruthWR (static inline)
// C: word * If_CutTruthWR( If_Man_t * p, If_Cut_t * pCut )
func (p *IfMan) ifCutTruthWR(pCut *IfCut) []uint64 {
	panic("TODO, stubbed")
}

// If_CutTruthUR (static inline)
// C: unsigned * If_CutTruthUR( If_Man_t * p, If_Cut_t * pCut)
func (p *IfMan) ifCutTruthUR(pCut *IfCut) []uint32 {
	panic("TODO, stubbed")
}

// If_CutTruthW (static inline)
// C: word * If_CutTruthW( If_Man_t * p, If_Cut_t * pCut )
func (p *IfMan) ifCutTruthW(pCut *IfCut) []uint64 {
	panic("TODO, stubbed")
}

// If_CutTruth (static inline)
// C: unsigned * If_CutTruth( If_Man_t * p, If_Cut_t * pCut )
func (p *IfMan) ifCutTruth(pCut *IfCut) []uint32 {
	panic("TODO, stubbed")
}

// If_CutDsdLit (static inline)
// C: int If_CutDsdLit( If_Man_t * p, If_Cut_t * pCut )
func (p *IfMan) ifCutDsdLit(pCut *IfCut) int {
	panic("TODO, stubbed")
}

// If_CutDsdIsCompl (static inline)
// C: int If_CutDsdIsCompl( If_Man_t * p, If_Cut_t * pCut )
func (p *IfMan) ifCutDsdIsCompl(pCut *IfCut) int {
	panic("TODO, stubbed")
}

// If_CutDsdPerm (static inline; char *)
// C: char * If_CutDsdPerm( If_Man_t * p, If_Cut_t * pCut )
func (p *IfMan) ifCutDsdPerm(pCut *IfCut) []int8 {
	panic("TODO, stubbed")
}

// If_CutLutArea (static inline)
// C: float If_CutLutArea( If_Man_t * p, If_Cut_t * pCut )
func (p *IfMan) ifCutLutArea(pCut *IfCut) float32 {
	panic("TODO, stubbed")
}

// If_CutLutDelay (static inline)
// C: float If_CutLutDelay( If_LibLut_t * p, int Size, int iPin )
func ifCutLutDelay(p *IfLibLut, Size int, iPin int) float32 {
	panic("TODO, stubbed")
}
