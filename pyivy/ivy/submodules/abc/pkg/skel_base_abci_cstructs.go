package abc

// Hand-written skeleton of C structs that src/base/abci passes around but that
// had no Go declaration: reo_man (src/bdd/reo/reo.h), Fraig_Params_t and
// Prove_Params_t (src/proof/fraig/fraig.h), Amap_Out_t (src/map/amap/amap.h)
// and Abc_Aig_t (src/base/abc/abcAig.c). Types only; no functions.

const REO_CONST_LEVEL = 30000 // the number of the constant level

// ReoUnit is reo_unit: the unit representing one DD node during reordering.
type ReoUnit struct {
	lev       int16    // the level of this node at the beginning
	TopRef    int16    // the top level from which this node is refed (used to update BDD width)
	TopRefNew int16    // the new top level from which this node is refed (used to update BDD width)
	n         int16    // the number of incoming edges (similar to ref count in the BDD)
	Sign      int      // the signature
	pE        *ReoUnit // the pointer to the "else" branch
	pT        *ReoUnit // the pointer to the "then" branch
	Next      *ReoUnit // the link to the next one in the list
	Weight    float64  // the probability of traversing this node
}

// ReoPlane is reo_plane: the set of nodes on one level.
type ReoPlane struct {
	fSifted        int      // to mark the sifted variables
	statsNodes     int      // the number of nodes in the current level
	statsWidth     int      // the width on the current level
	statsApl       float64  // the sum of node probabilities on this level
	statsCost      float64  // the current cost is stored here
	statsCostAbove float64  // the current cost is stored here
	statsCostBelow float64  // the current cost is stored here
	pHead          *ReoUnit // the pointer to the beginning of the unit list
}

// ReoHash is reo_hash: the entry in the hash table.
type ReoHash struct {
	Sign int      // signature of the current cache operation
	Arg1 *ReoUnit // the first argument
	Arg2 *ReoUnit // the second argument
	Arg3 *ReoUnit // the third argument
}

// ReoMan is reo_man: the reordering manager.
type ReoMan struct {
	// these paramaters can be set by the API functions
	fMinWidth int // the flag to enable reordering for minimum width
	fMinApl   int // the flag to enable reordering for minimum APL
	fVerbose  int // the verbosity level
	fVerify   int // the flag toggling verification
	fRemapUp  int // the flag to enable remapping
	nIters    int // the number of iterations of sifting to perform
	// parameters given by the user when reordering is called
	dd     *CuddFoundation // the CUDD BDD manager
	pOrder []int           // int * : the resulting variable order will be returned here
	// derived parameters
	fThisIsAdd  int        // this flag is one if the function is the ADD
	pSupp       []int      // int * : the support of the given function
	nSuppAlloc  int        // the max allowed number of support variables
	nSupp       int        // the number of support variables
	pOrderInt   []int      // int * : the array storing the internal variable permutation
	pVarCosts   []float64  // double * : other arrays
	pLevelOrder []int      // int * : other arrays
	pWidthCofs  []*ReoUnit // reo_unit ** : temporary storage for cofactors used during reordering for width
	// parameters related to cost
	nNodesBeg int
	nNodesCur int
	nNodesEnd int
	nWidthCur int
	nWidthBeg int
	nWidthEnd int
	nAplCur   float64
	nAplBeg   float64
	nAplEnd   float64
	// mapping of the function into planes and back
	pMapToPlanes      []int // int * : the mapping of var indexes into plane levels
	pMapToDdVarsOrig  []int // int * : the mapping of plane levels into the original indexes
	pMapToDdVarsFinal []int // int * : the mapping of plane levels into the final indexes
	// the planes table
	pPlanes    []ReoPlane // reo_plane *
	nPlanes    int
	pTops      []*ReoUnit // reo_unit **
	nTops      int
	nTopsAlloc int
	// the hash table
	HTable     []ReoHash // reo_hash * : the table itself
	nTableSize int       // the size of the hash table
	Signature  int       // the signature counter
	// the referenced node list
	nNodesMaxAlloc int       // this parameters determins how much memory is allocated
	pRefNodes      []*DdNode // DdNode **
	nRefNodes      int
	nRefNodesAlloc int
	// unit memory management
	pUnitFreeList   *ReoUnit
	pMemChunks      [][]ReoUnit // reo_unit ** : memory chunks
	nMemChunks      int
	nMemChunksAlloc int
	nUnitsUsed      int
	// statistic variables
	HashSuccess int
	HashFailure int
	nSwaps      int // the number of swaps
	nNISwaps    int // the number of swaps without interaction
}

// FraigPar is Fraig_Params_t (struct Fraig_ParamsStruct_t_).
type FraigPar struct {
	nPatsRand  int   // the number of words of random simulation info
	nPatsDyna  int   // the number of words of dynamic simulation info
	nBTLimit   int   // the max number of backtracks to perform
	nSeconds   int   // the timeout for the final proof
	fFuncRed   int   // performs only one level hashing
	fFeedBack  int   // enables solver feedback
	fDist1Pats int   // enables distance-1 patterns
	fDoSparse  int   // performs equiv tests for sparse functions
	fChoicing  int   // enables recording structural choices
	fTryProve  int   // tries to solve the final miter
	fVerbose   int   // the verbosiness flag
	fVerboseP  int   // the verbosiness flag (for proof reporting)
	fInternal  int   // is set to 1 for internal fraig calls
	nConfLimit int   // the limit on the number of conflicts
	nInspLimit int64 // ABC_INT64_T : the limit on the number of inspections
}

// ProveParams is Prove_Params_t (struct Prove_ParamsStruct_t_).
type ProveParams struct {
	// general parameters
	fUseFraiging  int // enables fraiging
	fUseRewriting int // enables rewriting
	fUseBdds      int // enables BDD construction when other methods fail
	fVerbose      int // prints verbose stats
	// iterations
	nItersMax int // the number of iterations
	// mitering
	nMiteringLimitStart int     // starting mitering limit
	nMiteringLimitMulti float32 // multiplicative coefficient to increase the limit in each iteration
	// rewriting
	nRewritingLimitStart int     // the number of rewriting iterations
	nRewritingLimitMulti float32 // multiplicative coefficient to increase the limit in each iteration
	// fraiging
	nFraigingLimitStart int     // starting backtrack(conflict) limit
	nFraigingLimitMulti float32 // multiplicative coefficient to increase the limit in each iteration
	// last-gasp BDD construction
	nBddSizeLimit int // the number of BDD nodes when construction is aborted
	fBddReorder   int // enables dynamic BDD variable reordering
	// last-gasp mitering
	nMiteringLimitLast int // final mitering limit
	// global SAT solver limits
	nTotalBacktrackLimit int64 // ABC_INT64_T : global limit on the number of backtracks
	nTotalInspectLimit   int64 // ABC_INT64_T : global limit on the number of clause inspects
	// global resources applied
	nTotalBacktracksMade int64 // ABC_INT64_T : the total number of backtracks made
	nTotalInspectsMade   int64 // ABC_INT64_T : the total number of inspects made
}

// AmapOut is Amap_Out_t.
type AmapOut struct {
	pName string // char * : gate name
	Type  int16  // short : node type (-1=input; 0=internal; 1=output)
	nFans int16  // short : number of fanins
	pFans []int  // int pFans[0] : fanin (flexible array member)
}

// AbcAig is Abc_Aig_t (src/base/abc/abcAig.c).
type AbcAig struct {
	pNtkAig          *ABCNetwork                      // the AIG network
	pConst1          *ABCNetworkObject                // the constant 1 object (not a node!)
	pBins            []*ABCNetworkObject              // Abc_Obj_t ** : the table bins
	nBins            int                              // the size of the table
	nEntries         int                              // the total number of entries in the table
	vNodes           *PointerVector[ABCNetworkObject] // the temporary array of nodes
	vStackReplaceOld *PointerVector[ABCNetworkObject] // the nodes to be replaced
	vStackReplaceNew *PointerVector[ABCNetworkObject] // the nodes to be used for replacement
	vLevels          *NestedVector[*ABCNetworkObject] // the nodes to be updated
	vLevelsR         *NestedVector[*ABCNetworkObject] // the nodes to be updated
	vAddedCells      *PointerVector[ABCNetworkObject] // the added nodes
	vUpdatedNets     *PointerVector[ABCNetworkObject] // the nodes whose fanouts have changed
	nStrash0         int
	nStrash1         int
	nStrash5         int
	nStrash2         int
}

// CecParFra is Cec_ParFra_t (src/proof/cec/cec.h): SAT sweeping parameters.
type CecParFra struct {
	jType         int // solver type
	nWords        int // the number of simulation words
	nRounds       int // the number of simulation rounds
	nItersMax     int // the maximum number of iterations of SAT sweeping
	nBTLimit      int // conflict limit at a node
	TimeLimit     int // the runtime limit in seconds
	nLevelMax     int // restriction on the level nodes to be swept
	nDepthMax     int // the depth in terms of steps of speculative reduction
	nCallsRecycle int // calls to perform before recycling SAT solver
	nSatVarMax    int // the max number of SAT variables
	nGenIters     int // pattern generation iterations
	fRewriting    int // enables AIG rewriting
	fCheckMiter   int // the circuit is the miter
	fDualOut      int // miter with separate outputs
	fColorDiff    int // miter with separate outputs
	fSatSweeping  int // enable SAT sweeping
	fRunCSat      int // enable another solver
	fUseCones     int // use cones
	fUseOrigIds   int // enable recording of original IDs
	fVeryVerbose  int // verbose stats
	fVerbose      int // verbose stats
	iOutFail      int // the failed output
}

// CecParCec is Cec_ParCec_t (src/proof/cec/cec.h): combinational equivalence checking parameters.
type CecParCec struct {
	nBTLimit     int // conflict limit at a node
	TimeLimit    int // the runtime limit in seconds
	fUseSmartCnf int // use smart CNF computation
	fRewriting   int // enables AIG rewriting
	fNaive       int // performs naive SAT-based checking
	fSilent      int // print no messages
	fVeryVerbose int // verbose stats
	fVerbose     int // verbose stats
	iOutFail     int // the number of failed output
}
