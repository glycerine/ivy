package abc

// Hand-checked skeleton of src/opt/lpk/lpkInt.h.

// Hand-written skeleton of the types and static inline functions in
// src/opt/lpk/lpkInt.h. LPK_SIZE_MAX is 100 and LPK_CUTS_MAX is 10000. Go has
// no bitfields; each is a uint32 whose C width is in the comment. A Vec_Ptr_t
// that mixes Abc_Obj_t * leaves with Lpk_Fun_t * decomposition nodes is
// written PointerVector[any].

// LpkCut is Lpk_Cut_t.
type LpkCut struct {
	nLeaves   uint32    // unsigned : 6, (L) the number of leaves
	nNodes    uint32    // unsigned : 6, (M) the number of nodes
	nNodesDup uint32    // unsigned : 6, (Q) nodes outside of MFFC
	nLuts     uint32    // unsigned : 6, (N) the number of LUTs to try
	unused    uint32    // unsigned : 6, unused
	fHasDsd   uint32    // unsigned : 1, set to 1 if the cut has structural DSD (and so cannot be used)
	fMark     uint32    // unsigned : 1, multipurpose mark
	uSign     [2]uint32 // the signature
	Weight    float32   // the weight of the cut: (M - Q)/N(V)   (the larger the better)
	Gain      int       // the gain achieved using this cut
	pLeaves   [100]int  // the leaves of the cut
	pNodes    [100]int  // the nodes of the cut
}

// LpkMan is Lpk_Man_t.
type LpkMan struct {
	// parameters
	pPars *LpkPar // the set of parameters
	// current representation
	pNtk *ABCNetwork       // the network
	pObj *ABCNetworkObject // the node to resynthesize
	// cut representation
	nMffc    int           // the size of MFFC of the node
	nCuts    int           // the total number of cuts
	nCutsMax int           // the largest possible number of cuts
	nEvals   int           // the number of good cuts
	pCuts    [10000]LpkCut // the storage for cuts
	pEvals   [10000]int    // the good cuts
	// visited nodes
	vVisited *NestedVector[*ABCNetworkObject]
	// mapping manager
	pIfMan  *IfMan
	vCover  *IntVector
	vLevels *NestedVector[*ABCNetworkObject]
	// temporary variables
	fCofactoring int                 // working in the cofactoring mode
	fCalledOnce  int                 // limits the depth of MUX cofactoring
	nCalledSRed  int                 // the number of called to SRed
	pRefs        [100]int            // fanin reference counters
	pCands       [100]int            // internal nodes pointing only to the leaves
	vLeaves      *PointerVector[any] // Abc_Obj_t * and Lpk_Fun_t * entries
	vTemp        *PointerVector[ABCNetworkObject]
	// truth table representation
	vTtElems *PointerVector[uint32] // elementary truth tables
	vTtNodes *PointerVector[uint32] // storage for temporary truth tables of the nodes
	vMemory  *IntVector
	vBddDir  *IntVector
	vBddInv  *IntVector
	puSupps  [32]uint32      // the supports of the cofactors
	ppTruths [5][16][]uint32 // unsigned * ppTruths[5][16]
	// variable sets
	vSets   [8]*IntVector
	pDsdMan *KitDsdMan
	// statistics
	nNodesTotal  int // total number of nodes
	nNodesOver   int // nodes with cuts over the limit
	nCutsTotal   int // total number of cuts
	nCutsUseful  int // useful cuts
	nGainTotal   int // the gain in LUTs
	nChanges     int // the number of changed nodes
	nBenefited   int // the number of gainful that benefited from decomposition
	nMuxes       int
	nDsds        int
	nTotalNets   int
	nTotalNets2  int
	nTotalNodes  int
	nTotalNodes2 int
	// counter of non-DSD blocks
	nBlocks [17]int
	// runtime
	timeCuts   int64 // abctime
	timeTruth  int64
	timeSupps  int64
	timeTruth2 int64
	timeTruth3 int64
	timeEval   int64
	timeMap    int64
	timeOther  int64
	timeTotal  int64
	// runtime of eval
	timeEvalMuxAn int64
	timeEvalMuxSp int64
	timeEvalDsdAn int64
	timeEvalDsdSp int64
}

// LpkFun is Lpk_Fun_t.
type LpkFun struct {
	vNodes    *PointerVector[any] // the array of leaves and decomposition nodes
	Id        uint32              // unsigned : 7, the ID of this node
	nVars     uint32              // unsigned : 5, the number of variables
	nLutK     uint32              // unsigned : 4, the number of LUT inputs
	nAreaLim  uint32              // unsigned : 14, the area limit (the largest allowed)
	fSupports uint32              // unsigned : 1, supports of cofactors were precomputed
	fMark     uint32              // unsigned : 1, marks the MUX-based dec
	uSupp     uint32              // the support of this component
	puSupps   [32]uint32          // the supports of the cofactors
	nDelayLim uint32              // the delay limit (the largest allowed)
	pDelays   [16]int             // the delays of the inputs
	pFanins   [16]int8            // char pFanins[16] : the fanins of this function
	pTruth    []uint32            // unsigned pTruth[0] : flexible array member, the truth table (contains room for three truth tables)
}

// LpkRes is Lpk_Res_t.
type LpkRes struct {
	nBSVars    int     // the number of bound set variables
	BSVars     uint32  // the bound set
	nCofVars   int     // the number of cofactoring variables
	pCofVars   [4]int8 // char pCofVars[4] : the cofactoring variables
	nSuppSizeS int     // support size of the smaller (decomposed) function
	nSuppSizeL int     // support size of the larger (composition) function
	DelayEst   int     // estimated delay of the decomposition
	AreaEst    int     // estimated area of the decomposition
	Variable   int     // variable in MUX decomposition
	Polarity   int     // polarity in MUX decomposition
}

// Lpk_LutNumVars (static inline)
// C: int Lpk_LutNumVars( int nLutsLim, int nLutK )
func lpkLutNumVars(nLutsLim int, nLutK int) int {
	panic("TODO, stubbed")
}

// Lpk_LutNumLuts (static inline)
// C: int Lpk_LutNumLuts( int nVarsMax, int nLutK )
func lpkLutNumLuts(nVarsMax int, nLutK int) int {
	panic("TODO, stubbed")
}

// Lpk_FunTruth (static inline; a pointer into the truth-table storage)
// C: unsigned * Lpk_FunTruth( Lpk_Fun_t * p, int Num )
func (p *LpkFun) lpkFunTruth(Num int) []uint32 {
	panic("TODO, stubbed")
}
