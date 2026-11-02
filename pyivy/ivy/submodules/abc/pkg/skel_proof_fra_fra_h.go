package abc

// Hand-written skeleton of the types in src/proof/fra/fra.h.
// Field names, order and meaning follow the C. Pointer-vector element types
// are written PointerVector[X] for a C Vec_Ptr_t holding X*.

// FraPar is Fra_Par_t: FRAIG parameters.
type FraPar struct {
	nSimWords       int     // the number of words in the simulation info
	dSimSatur       float64 // the ratio of refined classes when saturation is reached
	fPatScores      int     // enables simulation pattern scoring
	MaxScore        int     // max score after which resimulation is used
	dActConeRatio   float64 // the ratio of cone to be bumped
	dActConeBumpMax float64 // the largest bump in activity
	fChoicing       int     // enables choicing
	fSpeculate      int     // use speculative reduction
	fProve          int     // prove the miter outputs
	fVerbose        int     // verbose output
	fDoSparse       int     // skip sparse functions
	fConeBias       int     // bias variables in the cone (good for unsat runs)
	nBTLimitNode    int     // conflict limit at a node
	nBTLimitMiter   int     // conflict limit at an output
	nLevelMax       int     // the max level to consider seriously
	nFramesP        int     // the number of timeframes to in the prefix
	nFramesK        int     // the number of timeframes to unroll
	nMaxImps        int     // the maximum number of implications to consider
	nMaxLevs        int     // the maximum number of levels to consider
	fRewrite        int     // use rewriting for constraint reduction
	fLatchCorr      int     // computes latch correspondence only
	fUseImps        int     // use implications
	fUse1Hot        int     // use one-hotness conditions
	fWriteImps      int     // record implications
	fDontShowBar    int     // does not show progressbar during fraiging
}

// FraSsw is Fra_Ssw_t: seq SAT sweeping parameters.
type FraSsw struct {
	nPartSize   int     // size of the partition
	nOverSize   int     // size of the overlap between partitions
	nFramesP    int     // number of frames in the prefix
	nFramesK    int     // number of frames for induction (1=simple)
	nMaxImps    int     // max implications to consider
	nMaxLevs    int     // max levels to consider
	nMinDomSize int     // min clock domain considered for optimization
	fUseImps    int     // use implications
	fRewrite    int     // enable rewriting of the specualatively reduced model
	fFraiging   int     // enable comb SAT sweeping as preprocessing
	fLatchCorr  int     // perform register correspondence
	fWriteImps  int     // write implications into a file
	fUse1Hot    int     // use one-hotness constraints
	fVerbose    int     // enable verbose output
	fSilent     int     // disable any output
	nIters      int     // the number of iterations performed
	TimeLimit   float32 // the runtime budget for this call
}

// FraSec is Fra_Sec_t: SEC parameters.
type FraSec struct {
	fTryComb         int // try CEC call as a preprocessing step
	fTryBmc          int // try BMC call as a preprocessing step
	nFramesMax       int // the max number of frames used for induction
	nBTLimit         int // the conflict limit at a node
	nBTLimitGlobal   int // the global conflict limit
	nBTLimitInter    int // the conflict limit for interpolation
	nBddVarsMax      int // the state space limit for BDD reachability
	nBddMax          int // the max number of BDD nodes
	nBddIterMax      int // the limit on the number of BDD iterations
	nPdrTimeout      int // the timeout for PDR in the end
	fPhaseAbstract   int // enables phase abstraction
	fRetimeFirst     int // enables most-forward retiming at the beginning
	fRetimeRegs      int // enables min-register retiming at the beginning
	fFraiging        int // enables fraiging at the beginning
	fInduction       int // enable the use of induction
	fInterpolation   int // enables interpolation
	fInterSeparate   int // enables interpolation for each outputs separately
	fReachability    int // enables BDD based reachability
	fReorderImage    int // enables BDD reordering during image computation
	fStopOnFirstFail int // enables stopping after first output of a miter has failed to prove
	fUseNewProver    int // the new prover
	fUsePdr          int // the PDR
	fSilent          int // disables all output
	fVerbose         int // enables verbose reporting of statistics
	fVeryVerbose     int // enables very verbose reporting
	TimeLimit        int // enables the timeout
	fReadUnsolved    int // inserts the unsolved model back
	nSMnumber        int // the number of model written
	// internal parameters
	fRecursive      int // set to 1 when SEC is called recursively
	fReportSolution int // enables report solution in a special form
}

// FraCla is Fra_Cla_t: FRAIG equivalence classes.
type FraCla struct {
	pAig            *AigManager                // the original AIG manager
	pMemRepr        []*AigObject               // Aig_Obj_t ** : pointers to representatives of each node
	vClasses        *PointerVector[*AigObject] // equivalence classes (each an Aig_Obj_t **)
	vClasses1       *PointerVector[AigObject]  // equivalence class of Const1 node
	vClassesTemp    *PointerVector[*AigObject] // temporary storage for new classes
	pMemClasses     []*AigObject               // Aig_Obj_t ** : memory allocated for equivalence classes
	pMemClassesFree []*AigObject               // Aig_Obj_t ** : memory allocated for equivalence classes to be used
	vClassOld       *PointerVector[AigObject]  // old equivalence class after splitting
	vClassNew       *PointerVector[AigObject]  // new equivalence class(es) after splitting
	nPairs          int                        // the number of pairs of nodes
	fRefinement     int                        // set to 1 when refinement has happened
	vImps           *IntVector                 // implications
	// procedures used for class refinement
	pFuncNodeHash      func(*AigObject, int) int        // returns has key of the node
	pFuncNodeIsConst   func(*AigObject) int             // returns 1 if the node is a constant
	pFuncNodesAreEqual func(*AigObject, *AigObject) int // returns 1 if nodes are equal up to a complement
}

// FraSml is Fra_Sml_t: simulation manager.
type FraSml struct {
	pAig         *AigManager // the original AIG manager
	nPref        int         // the number of times frames in the prefix
	nFrames      int         // the number of times frames
	nWordsFrame  int         // the number of words in each time frame
	nWordsTotal  int         // the total number of words at a node
	nWordsPref   int         // the number of word in the prefix
	fNonConstOut int         // have seen a non-const-0 output during simulation
	nSimRounds   int         // statistics
	timeSim      int         // statistics
	pData        []uint32    // unsigned pData[0] : flexible array member, simulation data for the nodes
}

// FraBmc is Fra_Bmc_t, defined in fraBmc.c.
type FraBmc struct {
	// parameters
	nPref      int // the size of the prefix
	nDepth     int // the depth of the frames
	nFramesAll int // the total number of timeframes
	// implications to be filtered
	vImps *IntVector
	// AIG managers
	pAig       *AigManager // the original AIG manager
	pAigFrames *AigManager // initialized timeframes
	pAigFraig  *AigManager // the fraiged initialized timeframes
	// mapping of nodes
	pObjToFrames []*AigObject // Aig_Obj_t ** : mapping of the original node into frames
	pObjToFraig  []*AigObject // Aig_Obj_t ** : mapping of the frames node into fraig
}

// FraMan is Fra_Man_t: FRAIG manager.
type FraMan struct {
	// high-level data
	pPars *FraPar // parameters governing fraiging
	// AIG managers
	pManAig   *AigManager // the starting AIG manager
	pManFraig *AigManager // the final AIG manager
	// mapping AIG into FRAIG
	nFramesAll int          // the number of timeframes used
	pMemFraig  []*AigObject // Aig_Obj_t ** : memory allocated for points to the fraig nodes
	nSizeAlloc int          // allocated size of the arrays for timeframe nodes
	// equivalence classes
	pCla *FraCla // representation of (candidate) equivalent nodes
	// simulation info
	pSml *FraSml // simulation manager
	// bounded model checking manager
	pBmc *FraBmc
	// counter example storage
	nPatWords int      // the number of words in the counter example
	pPatWords []uint32 // unsigned * : the counter example
	vCex      *IntVector
	// one-hotness conditions
	vOneHots *IntVector
	// satisfiability solving
	pSat            *SatSolver                  // SAT solver
	nSatVars        int                         // the number of variables currently used
	vPiVars         *PointerVector[AigObject]   // the PIs of the cone used
	nBTLimitGlobal  int64                       // resource limit
	nInsLimitGlobal int64                       // resource limit
	pMemFanins      []*PointerVector[AigObject] // Vec_Ptr_t ** : the arrays of fanins for some FRAIG nodes
	pMemSatNums     []int                       // int * : the array of SAT numbers for some FRAIG nodes
	nMemAlloc       int                         // allocated size of the arrays for FRAIG varnums and fanins
	vTimeouts       *PointerVector[AigObject]   // the nodes, for which equivalence checking timed out
	// statistics
	nSimRounds       int
	nNodesMiter      int
	nLitsBeg         int
	nLitsEnd         int
	nNodesBeg        int
	nNodesEnd        int
	nRegsBeg         int
	nRegsEnd         int
	nSatCalls        int
	nSatCallsSat     int
	nSatCallsUnsat   int
	nSatProof        int
	nSatFails        int
	nSatFailsReal    int
	nSpeculs         int
	nChoices         int
	nChoicesFake     int
	nSatCallsRecent  int
	nSatCallsSkipped int
	// runtime
	timeSim      int64
	timeTrav     int64
	timeRwr      int64
	timeSat      int64
	timeSatUnsat int64
	timeSatSat   int64
	timeSatFail  int64
	timeRef      int64
	timeTotal    int64
	time1        int64
	time2        int64
}
